package inference

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/attachment"
	"github.com/ppipada/flexigpt-app/pkg/fileutil"
	"github.com/ppipada/flexigpt-app/pkg/inference/debugclient"
	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	anthropicSharedConstant "github.com/anthropics/anthropic-sdk-go/shared/constant"
)

// AnthropicMessagesAPI implements CompletionProvider for Anthropics' Messages API.
type AnthropicMessagesAPI struct {
	ProviderParams *spec.ProviderParams
	Debug          bool
	client         *anthropic.Client
}

// NewAnthropicMessagesAPI creates a new instance of Anthropics provider.
func NewAnthropicMessagesAPI(
	pi spec.ProviderParams,
	debug bool,
) (*AnthropicMessagesAPI, error) {
	if pi.Name == "" || pi.Origin == "" {
		return nil, errors.New("anthropic messages api LLM: invalid args")
	}
	return &AnthropicMessagesAPI{
		ProviderParams: &pi,
		Debug:          debug,
	}, nil
}

func (api *AnthropicMessagesAPI) InitLLM(ctx context.Context) error {
	if !api.IsConfigured(ctx) {
		slog.Debug(
			string(
				api.ProviderParams.Name,
			) + ": No API key given. Not initializing Anthropics client",
		)
		return nil
	}

	opts := []option.RequestOption{
		// Sets x-api-key.
		option.WithAPIKey(api.ProviderParams.APIKey),
	}

	providerURL := modelpresetSpec.DefaultAnthropicOrigin
	if api.ProviderParams.Origin != "" {
		baseURL := strings.TrimSuffix(api.ProviderParams.Origin, "/")
		// Remove 'v1/messages' from pathPrefix if present,
		// This is because anthropic sdk adds 'v1/messages' internally.
		pathPrefix := strings.TrimSuffix(
			api.ProviderParams.ChatCompletionPathPrefix,
			"v1/messages",
		)
		providerURL = baseURL + pathPrefix
		opts = append(opts, option.WithBaseURL(strings.TrimSuffix(providerURL, "/")))
	}

	// Add default headers.
	for k, v := range api.ProviderParams.DefaultHeaders {
		opts = append(opts, option.WithHeader(strings.TrimSpace(k), strings.TrimSpace(v)))
	}

	// If the caller provided a non-standard API key header, attach it.
	if api.ProviderParams.APIKeyHeaderKey != "" &&
		!strings.EqualFold(
			api.ProviderParams.APIKeyHeaderKey,
			modelpresetSpec.DefaultAnthropicAuthorizationHeaderKey,
		) &&
		!strings.EqualFold(
			api.ProviderParams.APIKeyHeaderKey,
			modelpresetSpec.DefaultAuthorizationHeaderKey,
		) {
		opts = append(
			opts,
			option.WithHeader(api.ProviderParams.APIKeyHeaderKey, api.ProviderParams.APIKey),
		)
	}

	dbgCfg := debugclient.DefaultDebugConfig
	dbgCfg.LogToSlog = api.Debug

	newClient := debugclient.NewDebugHTTPClient(dbgCfg)

	opts = append(opts, option.WithHTTPClient(newClient))

	c := anthropic.NewClient(opts...)
	api.client = &c
	slog.Info(
		"anthropic messages api LLM provider initialized",
		"name", string(api.ProviderParams.Name),
		"URL", providerURL,
	)
	return nil
}

func (api *AnthropicMessagesAPI) DeInitLLM(ctx context.Context) error {
	api.client = nil
	slog.Info(
		"anthropic messages api LLM: provider de initialized",
		"name",
		string(api.ProviderParams.Name),
	)
	return nil
}

func (api *AnthropicMessagesAPI) GetProviderInfo(ctx context.Context) *spec.ProviderParams {
	return api.ProviderParams
}

func (api *AnthropicMessagesAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderParams.APIKey != ""
}

func (api *AnthropicMessagesAPI) SetProviderAPIKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return errors.New("anthropic messages api LLM: invalid apikey provided")
	}
	if api.ProviderParams == nil {
		return errors.New("anthropic messages api LLM: no ProviderParams found")
	}
	api.ProviderParams.APIKey = apiKey
	return nil
}

func (api *AnthropicMessagesAPI) BuildCompletionData(
	ctx context.Context,
	modelParams spec.ModelParams,
	currentMessage spec.ChatCompletionDataMessage,
	prevMessages []spec.ChatCompletionDataMessage,
) (*spec.FetchCompletionData, error) {
	return getCompletionData(modelParams, currentMessage, prevMessages), nil
}

func (api *AnthropicMessagesAPI) FetchCompletion(
	ctx context.Context,
	completionData *spec.FetchCompletionData,
	onStreamTextData, onStreamThinkingData func(string) error,
) (*spec.FetchCompletionResponse, error) {
	if api.client == nil {
		return nil, errors.New("anthropic messages api LLM: client not initialized")
	}
	if completionData == nil || len(completionData.Messages) == 0 {
		return nil, errors.New("anthropic messages api LLM: empty completion data")
	}

	msgs, sysParams, err := toAnthropicMessages(
		ctx,
		completionData.ModelParams.SystemPrompt,
		completionData.Messages,
	)
	if err != nil {
		return nil, err
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(completionData.ModelParams.Name),
		MaxTokens: int64(completionData.ModelParams.MaxOutputLength),
		Messages:  msgs,
	}
	if len(sysParams) > 0 {
		params.System = sysParams
	}

	if rp := completionData.ModelParams.Reasoning; rp != nil &&
		rp.Type == modelpresetSpec.ReasoningTypeHybridWithTokens &&
		rp.Tokens > 0 {
		tokens := max(rp.Tokens, 1024)
		params.Thinking = anthropic.ThinkingConfigParamOfEnabled(int64(tokens))
	} else if t := completionData.ModelParams.Temperature; t != nil {
		params.Temperature = anthropic.Float(*t)
	}

	timeout := modelpresetSpec.DefaultAPITimeout
	if completionData.ModelParams.Timeout > 0 {
		timeout = time.Duration(completionData.ModelParams.Timeout) * time.Second
	}

	var toolNameMap map[string]spec.FetchCompletionToolChoice

	if len(completionData.ToolChoices) > 0 {
		toolDefs, m, err := toAnthropicTools(completionData.ToolChoices)
		if err != nil {
			return nil, err
		}
		if len(toolDefs) > 0 {
			params.Tools = toolDefs
			toolNameMap = m
		}
	}

	if completionData.ModelParams.Stream && onStreamTextData != nil && onStreamThinkingData != nil {
		return api.doStreaming(ctx, params, onStreamTextData, onStreamThinkingData, timeout, toolNameMap)
	}
	return api.doNonStreaming(ctx, params, timeout, toolNameMap)
}

func (api *AnthropicMessagesAPI) doNonStreaming(
	ctx context.Context,
	params anthropic.MessageNewParams,
	timeout time.Duration,
	toolNameMap map[string]spec.FetchCompletionToolChoice,
) (*spec.FetchCompletionResponse, error) {
	completionResp := &spec.FetchCompletionResponse{Body: &spec.FetchCompletionResponseBody{}}
	ctx = debugclient.AddDebugResponseToCtx(ctx)

	resp, err := api.client.Messages.New(ctx, params, option.WithRequestTimeout(timeout))
	isNilResp := resp == nil || len(resp.Content) == 0
	attachDebugResp(ctx, completionResp.Body, err, isNilResp, "", resp)
	completionResp.Body.Usage = usageFromAnthropicMessage(resp)

	if isNilResp {
		return completionResp, nil
	}

	respContent := getResponseContentFromAnthropicMessage(resp)
	completionResp.Body.ResponseContent = respContent
	if toolCalls := extractAnthropicToolCalls(resp, toolNameMap); len(toolCalls) > 0 {
		completionResp.Body.ToolCalls = toolCalls
	}
	return completionResp, nil
}

func (api *AnthropicMessagesAPI) doStreaming(
	ctx context.Context,
	params anthropic.MessageNewParams,
	onStreamTextData, onStreamThinkingData func(string) error,
	timeout time.Duration,
	toolNameMap map[string]spec.FetchCompletionToolChoice,
) (*spec.FetchCompletionResponse, error) {
	writeTextData, flushTextData := NewBufferedStreamer(
		onStreamTextData,
		FlushInterval,
		FlushChunkSize,
	)
	writeThinkingData, flushThinkingData := NewBufferedStreamer(
		onStreamThinkingData,
		FlushInterval,
		FlushChunkSize,
	)

	completionResp := &spec.FetchCompletionResponse{Body: &spec.FetchCompletionResponseBody{}}
	ctx = debugclient.AddDebugResponseToCtx(ctx)

	stream := api.client.Messages.NewStreaming(
		ctx,
		params,
		option.WithRequestTimeout(timeout),
	)

	respFull := anthropic.Message{}
	var (
		streamWriteErr      error
		streamAccumulateErr error
	)

	for stream.Next() {
		event := stream.Current()
		err := respFull.Accumulate(event)
		if err != nil {
			streamAccumulateErr = err
			break
		}

		switch eventVariant := event.AsAny().(type) {
		case anthropic.MessageStartEvent:
			// Contains a Message object with empty content i.e message metadata.
		case anthropic.MessageDeltaEvent:
			// Indicating top-level changes to the final Message object metadata.
		case anthropic.MessageStopEvent:
			// Done.
			break
		case anthropic.ContentBlockStopEvent:
			// Content block done.
		case anthropic.ContentBlockStartEvent:
			streamWriteErr = handleContentBlockStartEvent(eventVariant, writeTextData, writeThinkingData)
			if streamWriteErr != nil {
				break
			}
		case anthropic.ContentBlockDeltaEvent:
			streamWriteErr = handleContentBlockDeltaEvent(eventVariant, writeTextData, writeThinkingData)
			if streamWriteErr != nil {
				break
			}
		default:
			// No valid variant.
		}
	}

	if flushTextData != nil {
		flushTextData()
	}

	if flushThinkingData != nil {
		flushThinkingData()
	}

	streamErr := errors.Join(stream.Err(), streamAccumulateErr, streamWriteErr)
	isNilResp := len(respFull.Content) == 0
	attachDebugResp(ctx, completionResp.Body, streamErr, isNilResp, "", respFull)
	completionResp.Body.Usage = usageFromAnthropicMessage(&respFull)

	if isNilResp {
		return completionResp, nil
	}

	respContent := getResponseContentFromAnthropicMessage(&respFull)
	completionResp.Body.ResponseContent = respContent
	if toolCalls := extractAnthropicToolCalls(&respFull, toolNameMap); len(toolCalls) > 0 {
		completionResp.Body.ToolCalls = toolCalls
	}
	return completionResp, streamErr
}

func handleContentBlockStartEvent(
	eventVariant anthropic.ContentBlockStartEvent,
	writeTextData, writeThinkingData func(string) error,
) error {
	switch cbStartvariant := eventVariant.ContentBlock.AsAny().(type) {
	case anthropic.TextBlock:
		return writeTextData(cbStartvariant.Text)

	case anthropic.ThinkingBlock:
		return writeThinkingData(cbStartvariant.Thinking)

	case anthropic.RedactedThinkingBlock:
	case anthropic.ToolUseBlock:
	case anthropic.ServerToolUseBlock:
	case anthropic.WebSearchToolResultBlock:
	default:
		// No valid variant.
	}
	return nil
}

func handleContentBlockDeltaEvent(
	eventVariant anthropic.ContentBlockDeltaEvent,
	writeTextData, writeThinkingData func(string) error,
) error {
	switch deltaVariant := eventVariant.Delta.AsAny().(type) {
	case anthropic.TextDelta:
		return writeTextData(deltaVariant.Text)

	case anthropic.ThinkingDelta:
		return writeThinkingData(deltaVariant.Thinking)

	case anthropic.InputJSONDelta:
	case anthropic.CitationsDelta:
	case anthropic.SignatureDelta:
	default:
		// No valid variant.
	}
	return nil
}

func toAnthropicMessages(
	ctx context.Context,
	systemPrompt string,
	messages []spec.ChatCompletionDataMessage,
) (msgs []anthropic.MessageParam, sysPrompts []anthropic.TextBlockParam, err error) {
	var out []anthropic.MessageParam
	var sysParts []string

	if s := strings.TrimSpace(systemPrompt); s != "" {
		sysParts = append(sysParts, s)
	}

	// Identify the last user message; attachments are associated with the current
	// user turn.
	lastUserIdx := -1
	for i, m := range messages {
		if m.Role == modelpresetSpec.RoleUser {
			lastUserIdx = i
		}
	}

	for idx, m := range messages {
		// Skip messages with no text, no attachments, no tool outputs, and no tool calls.
		if m.Content == nil &&
			len(m.Attachments) == 0 &&
			len(m.ToolOutputs) == 0 &&
			len(m.ToolCalls) == 0 {
			continue
		}
		content := ""
		if m.Content != nil {
			content = strings.TrimSpace(*m.Content)
		}

		switch m.Role {
		case modelpresetSpec.RoleSystem, modelpresetSpec.RoleDeveloper:
			if content != "" {
				sysParts = append(sysParts, content)
			}

		case modelpresetSpec.RoleUser:
			if m.Content == nil && len(m.Attachments) == 0 && len(m.ToolOutputs) == 0 {
				continue
			}
			var attachmentContent []anthropic.ContentBlockParamUnion
			if len(m.Attachments) > 0 {
				// Dont override original attachments in rehydration.
				overrideOriginalAttachment := false
				if idx == lastUserIdx {
					// For the current message, we may attach even if things changed between build and fetch.
					// There is a race here: if we override here, how to propagate to the caller that blocks have
					// changed now, so update the OrigRef fields in attachments or this message too.
					overrideOriginalAttachment = true
				}
				blocks, err := attachment.BuildContentBlocks(
					ctx,
					m.Attachments,
					attachment.WithOverrideOriginalContentBlock(overrideOriginalAttachment),
					attachment.WithOnlyTextKindContentBlock(false),
					attachment.WithForceFetchContentBlock(false),
				)
				if err != nil {
					return nil, nil, err
				}
				attachmentContent, err = contentBlocksToAnthropic(blocks)
				if err != nil {
					return nil, nil, err
				}
			}

			toolResultBlocks := toolOutputsToAnthropicBlocks(m.ToolOutputs)

			parts := make(
				[]anthropic.ContentBlockParamUnion,
				0,
				1+len(attachmentContent)+len(toolResultBlocks),
			)
			if content != "" {
				parts = append(parts, anthropic.NewTextBlock(content))
			}
			if len(attachmentContent) != 0 {
				parts = append(parts, attachmentContent...)
			}
			if len(toolResultBlocks) != 0 {
				parts = append(parts, toolResultBlocks...)
			}

			if len(parts) > 0 {
				out = append(out, anthropic.NewUserMessage(parts...))
			}

		case modelpresetSpec.RoleAssistant:
			// Combine assistant text and tool calls into a single assistant message
			// containing text and tool_use blocks.
			var parts []anthropic.ContentBlockParamUnion
			if content != "" {
				parts = append(parts, anthropic.NewTextBlock(content))
			}
			if len(m.ToolCalls) > 0 {
				toolUseBlocks := toolCallsToAnthropicBlocks(m.ToolCalls)
				if len(toolUseBlocks) != 0 {
					parts = append(parts, toolUseBlocks...)
				}
			}
			if len(parts) > 0 {
				out = append(out, anthropic.NewAssistantMessage(parts...))
			}

		case modelpresetSpec.RoleFunction, modelpresetSpec.RoleTool:
			// Tool results and any explanatory text are represented as a user message
			// with tool_result blocks (and optional text).
			var parts []anthropic.ContentBlockParamUnion
			if content != "" {
				parts = append(parts, anthropic.NewTextBlock(content))
			}
			toolResultBlocks := toolOutputsToAnthropicBlocks(m.ToolOutputs)
			if len(toolResultBlocks) != 0 {
				parts = append(parts, toolResultBlocks...)
			}

			if len(parts) > 0 {
				out = append(out, anthropic.NewUserMessage(parts...))
			}

		default:
			// Ignore unknown roles.
		}
	}

	sysParams := []anthropic.TextBlockParam{}
	if len(sysParts) > 0 {
		// Anthropic seems to support array of prompts too, but lets keeps sys prompt as a single string for now.
		sysStr := strings.Join(sysParts, "\n\n")
		sysParams = append(sysParams, anthropic.TextBlockParam{Text: sysStr})
	}

	return out, sysParams, nil
}

func toolCallsToAnthropicBlocks(
	toolCalls []toolSpec.ToolCall,
) []anthropic.ContentBlockParamUnion {
	if len(toolCalls) == 0 {
		return []anthropic.ContentBlockParamUnion{}
	}

	out := make([]anthropic.ContentBlockParamUnion, 0, len(toolCalls))
	for _, tc := range toolCalls {
		if tc.ID == "" || tc.Name == "" {
			continue
		}

		args := strings.TrimSpace(tc.Arguments)
		if args == "" {
			args = "{}"
		}
		jArgs := json.RawMessage(args)

		out = append(out, anthropic.NewToolUseBlock(tc.ID, jArgs, tc.Name))
	}

	if len(out) == 0 {
		return []anthropic.ContentBlockParamUnion{}
	}
	return out
}

func toolOutputsToAnthropicBlocks(
	toolOutputs []toolSpec.ToolOutput,
) []anthropic.ContentBlockParamUnion {
	if len(toolOutputs) == 0 {
		return []anthropic.ContentBlockParamUnion{}
	}

	out := make([]anthropic.ContentBlockParamUnion, 0, len(toolOutputs)+1)

	// Tool outputs with a callID => proper tool_result blocks.
	for _, o := range toolOutputs {
		if o.ID == "" {
			continue
		}
		toolContent := strings.TrimSpace(o.RawOutput)
		if toolContent == "" {
			toolContent = strings.TrimSpace(o.Summary)
		}
		if toolContent == "" {
			continue
		}

		tr := anthropic.NewToolResultBlock(
			o.ID,
			toolContent,
			false, // May want to see how to propagate tool errors to API.
		)
		out = append(out, tr)
	}

	// Orphan outputs (no callID) => rendered as a single text block.
	var orphanOutputs []toolSpec.ToolOutput
	for _, o := range toolOutputs {
		if o.ID == "" {
			orphanOutputs = append(orphanOutputs, o)
		}
	}
	if len(orphanOutputs) > 0 {
		if text := strings.TrimSpace(renderToolOutputsAsText(orphanOutputs)); text != "" {
			out = append(out, anthropic.NewTextBlock(text))
		}
	}

	if len(out) == 0 {
		return []anthropic.ContentBlockParamUnion{}
	}
	return out
}

// contentBlocksToAnthropic converts generic content blocks into Anthropic
// content blocks. Images are sent as base64 image blocks; files as document
// blocks (currently assuming PDF base64); text as text blocks.
func contentBlocksToAnthropic(
	blocks []attachment.ContentBlock,
) ([]anthropic.ContentBlockParamUnion, error) {
	if len(blocks) == 0 {
		return nil, nil
	}

	out := make([]anthropic.ContentBlockParamUnion, 0, len(blocks))
	for _, b := range blocks {
		txt := ""
		if b.Text != nil {
			txt = strings.TrimSpace(*b.Text)
		}
		bData := ""
		if b.Base64Data != nil {
			bData = strings.TrimSpace(*b.Base64Data)
		}
		mime := ""
		if b.MIMEType != nil {
			mime = strings.TrimSpace(*b.MIMEType)
		}

		switch b.Kind {
		case attachment.ContentBlockText:
			if txt == "" {
				continue
			}
			out = append(out, anthropic.NewTextBlock(txt))

		case attachment.ContentBlockImage:
			if bData == "" {
				continue
			}
			if mime == "" {
				mime = string(fileutil.DefaultImageMIME)
			}
			out = append(out, anthropic.NewImageBlockBase64(mime, bData))

		case attachment.ContentBlockFile:
			if bData == "" {
				continue
			}
			// For now, treat generic file blocks as base64 PDF sources.
			out = append(
				out,
				anthropic.NewDocumentBlock(
					anthropic.Base64PDFSourceParam{Data: bData},
				),
			)

		default:
			continue
		}
	}
	return out, nil
}

func getResponseContentFromAnthropicMessage(msg *anthropic.Message) []modelpresetSpec.MessageContent {
	if msg == nil {
		return []modelpresetSpec.MessageContent{}
	}
	resp := make([]modelpresetSpec.MessageContent, 0, len(msg.Content))
	for _, content := range msg.Content {
		switch variant := content.AsAny().(type) {
		case anthropic.TextBlock:
			resp = append(resp, modelpresetSpec.MessageContent{Type: modelpresetSpec.MessageContentTypeText, Content: variant.Text})
		case anthropic.ThinkingBlock:
			resp = append(resp, modelpresetSpec.MessageContent{Type: modelpresetSpec.MessageContentTypeThinking, Content: variant.Thinking})
		case anthropic.RedactedThinkingBlock:
			resp = append(resp, modelpresetSpec.MessageContent{Type: modelpresetSpec.MessageContentTypeRedactedThinking, Content: variant.Data})
		case anthropic.ToolUseBlock:
		case anthropic.ServerToolUseBlock:
		case anthropic.WebSearchToolResultBlock:
		default:
			// Invalid variant, dont handle as of now.
		}
	}
	return resp
}

func toAnthropicTools(
	tools []spec.FetchCompletionToolChoice,
) ([]anthropic.ToolUnionParam, map[string]spec.FetchCompletionToolChoice, error) {
	if len(tools) == 0 {
		return nil, nil, nil
	}

	ordered, nameMap := buildToolChoiceNameMapping(tools)

	out := make([]anthropic.ToolUnionParam, 0, len(tools))
	for _, tw := range ordered {
		ct := tw.Choice
		schema, err := decodeToolArgSchema(ct.Tool.ArgSchema)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"invalid tool schema for %s/%s@%s: %w",
				ct.BundleID,
				ct.Tool.Slug,
				ct.Tool.Version,
				err,
			)
		}

		inputSchema := anthropic.ToolInputSchemaParam{
			Type: anthropicSharedConstant.Object("object"),
		}
		if tVal, ok := schema["type"].(string); ok && strings.TrimSpace(tVal) != "" {
			inputSchema.Type = anthropicSharedConstant.Object(
				strings.ToLower(strings.TrimSpace(tVal)),
			)
			delete(schema, "type")
		}
		if props, ok := schema["properties"]; ok {
			inputSchema.Properties = props
			delete(schema, "properties")
		}
		if req, ok := schema["required"]; ok {
			switch v := req.(type) {
			case []any:
				required := make([]string, 0, len(v))
				for _, item := range v {
					if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
						required = append(required, strings.TrimSpace(s))
					}
				}
				if len(required) > 0 {
					inputSchema.Required = required
				}
			case []string:
				required := make([]string, 0, len(v))
				for _, item := range v {
					if strings.TrimSpace(item) != "" {
						required = append(required, strings.TrimSpace(item))
					}
				}
				if len(required) > 0 {
					inputSchema.Required = required
				}
			}
			delete(schema, "required")
		}
		if len(schema) > 0 {
			inputSchema.ExtraFields = schema
		}

		toolUnion := anthropic.ToolUnionParamOfTool(inputSchema, tw.Name)
		if variant := toolUnion.OfTool; variant != nil {
			variant.Type = anthropic.ToolTypeCustom
			if desc := toolDescription(ct); desc != "" {
				variant.Description = anthropic.String(desc)
			}
		}
		out = append(out, toolUnion)
	}
	return out, nameMap, nil
}

func extractAnthropicToolCalls(
	msg *anthropic.Message,
	toolNameMap map[string]spec.FetchCompletionToolChoice,
) []toolSpec.ToolCall {
	if msg == nil {
		return nil
	}
	out := make([]toolSpec.ToolCall, 0, len(msg.Content))
	for _, content := range msg.Content {
		variant, ok := content.AsAny().(anthropic.ToolUseBlock)
		if !ok {
			continue
		}
		name := strings.TrimSpace(variant.Name)
		vID := strings.TrimSpace(variant.ID)
		if vID == "" || name == "" {
			continue
		}
		var toolChoice *toolSpec.ToolChoice
		if toolNameMap != nil {
			if ct, ok := toolNameMap[name]; ok {
				// Add the actual choice to response.
				toolChoice = &ct.ToolChoice
			}
		}

		call := toolSpec.ToolCall{
			ID:         vID,
			CallID:     vID,
			Name:       variant.Name,
			Arguments:  strings.TrimSpace(string(variant.Input)),
			Type:       string(variant.Type),
			ToolChoice: toolChoice,
		}
		out = append(out, call)

	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// usageFromAnthropicMessage normalizes Anthropic usage into spec.Usage.
func usageFromAnthropicMessage(msg *anthropic.Message) *modelpresetSpec.Usage {
	uOut := &modelpresetSpec.Usage{}
	if msg == nil {
		return uOut
	}

	u := msg.Usage

	uOut.InputTokensCached = u.CacheReadInputTokens
	uOut.InputTokensUncached = u.InputTokens
	uOut.InputTokensTotal = u.CacheReadInputTokens + u.InputTokens
	uOut.OutputTokens = u.OutputTokens
	// Anthropic does not currently expose explicit reasoning token counts.
	uOut.ReasoningTokens = 0

	return uOut
}
