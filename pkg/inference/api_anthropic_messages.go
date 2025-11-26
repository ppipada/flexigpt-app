package inference

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"

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
		// This is because openai sdk adds 'v1/messages' internally.
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

	newClient := NewDebugHTTPClient(api.Debug, false)
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
		completionData.ModelParams.SystemPrompt,
		completionData.Messages,
		completionData.Attachments,
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

	if len(completionData.ToolChoices) > 0 {
		toolDefs, err := toAnthropicTools(completionData.ToolChoices)
		if err != nil {
			return nil, err
		}
		if len(toolDefs) > 0 {
			params.Tools = toolDefs
		}
	}

	if completionData.ModelParams.Stream && onStreamTextData != nil && onStreamThinkingData != nil {
		return api.doStreaming(ctx, params, onStreamTextData, onStreamThinkingData, timeout)
	}
	return api.doNonStreaming(ctx, params, timeout)
}

func (api *AnthropicMessagesAPI) doNonStreaming(
	ctx context.Context,
	params anthropic.MessageNewParams,
	timeout time.Duration,
) (*spec.FetchCompletionResponse, error) {
	completionResp := &spec.FetchCompletionResponse{Body: &spec.FetchCompletionResponseBody{}}
	ctx = AddDebugResponseToCtx(ctx)

	resp, err := api.client.Messages.New(ctx, params, option.WithRequestTimeout(timeout))
	isNilResp := resp == nil || len(resp.Content) == 0
	attachDebugResp(ctx, completionResp.Body, err, isNilResp)
	if isNilResp {
		return completionResp, nil
	}

	respContent := getResponseContentFromAnthropicMessage(resp)
	completionResp.Body.ResponseContent = respContent
	if toolCalls := extractAnthropicToolCalls(resp); len(toolCalls) > 0 {
		completionResp.Body.ToolCalls = toolCalls
	}
	return completionResp, nil
}

func (api *AnthropicMessagesAPI) doStreaming(
	ctx context.Context,
	params anthropic.MessageNewParams,
	onStreamTextData, onStreamThinkingData func(string) error,
	timeout time.Duration,
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
	ctx = AddDebugResponseToCtx(ctx)

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
	attachDebugResp(ctx, completionResp.Body, streamErr, isNilResp)
	if isNilResp {
		return completionResp, nil
	}

	respContent := getResponseContentFromAnthropicMessage(&respFull)
	completionResp.Body.ResponseContent = respContent
	if toolCalls := extractAnthropicToolCalls(&respFull); len(toolCalls) > 0 {
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
	systemPrompt string,
	messages []spec.ChatCompletionDataMessage,
	attachments []spec.ChatCompletionAttachment,
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
		if m.Role == spec.User {
			lastUserIdx = i
		}
	}

	attachmentBlocks, err := buildAnthropicAttachmentBlocks(attachments)
	if err != nil {
		return nil, nil, err
	}

	for idx, m := range messages {
		if m.Content == nil {
			continue
		}
		content := strings.TrimSpace(*m.Content)

		switch m.Role {
		case spec.System, spec.Developer:
			if content != "" {
				sysParts = append(sysParts, content)
			}
		case spec.User:
			// Attach any resolved image/file blocks to the final user message.
			if idx == lastUserIdx && len(attachmentBlocks) > 0 {
				blocks := make([]anthropic.ContentBlockParamUnion, 0, 1+len(attachmentBlocks))
				if content != "" {
					blocks = append(blocks, anthropic.NewTextBlock(content))
				}
				blocks = append(blocks, attachmentBlocks...)
				out = append(out, anthropic.NewUserMessage(blocks...))
			} else {
				out = append(out, anthropic.NewUserMessage(anthropic.NewTextBlock(content)))
			}
		case spec.Assistant:
			out = append(
				out,
				anthropic.NewAssistantMessage(anthropic.NewTextBlock(content)),
			)
		case spec.Function, spec.Tool:
			// Anthropic tools need to be processed independently of messages, skip for now.
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

// buildAnthropicAttachmentBlocks converts generic attachments into Anthropic
// content blocks. Images are sent as base64 image blocks; other attachment
// kinds fall back to short text descriptions.
func buildAnthropicAttachmentBlocks(
	attachments []spec.ChatCompletionAttachment,
) ([]anthropic.ContentBlockParamUnion, error) {
	if len(attachments) == 0 {
		return nil, nil
	}

	out := make([]anthropic.ContentBlockParamUnion, 0, len(attachments))
	for _, att := range attachments {
		switch att.Kind {
		case spec.AttachmentImage:
			if att.ImageRef == nil {
				continue
			}
			encoded, mimeType, err := imageEncodingFromRef(att.ImageRef)
			if err != nil {
				return nil, err
			}
			out = append(out, anthropic.NewImageBlockBase64(mimeType, encoded))

		case spec.AttachmentFile:
			// Anthropics' Messages API doesn't support arbitrary files directly in
			// the same way as OpenAI's multimodal inputs. Represent the file as a
			// short textual handle instead so future resolvers (doc search, PDF
			// extractors, etc.) can plug in without changing the provider wiring.
			if text := formatAttachmentAsText(att); text != "" {
				out = append(out, anthropic.NewTextBlock(text))
			}

		case spec.AttachmentDocIndex, spec.AttachmentPR, spec.AttachmentCommit, spec.AttachmentSnapshot:
			if text := formatAttachmentAsText(att); text != "" {
				out = append(out, anthropic.NewTextBlock(text))
			}
		default:
			continue
		}
	}
	return out, nil
}

func getResponseContentFromAnthropicMessage(msg *anthropic.Message) []spec.ResponseContent {
	if msg == nil {
		return []spec.ResponseContent{}
	}
	resp := make([]spec.ResponseContent, 0, len(msg.Content))
	for _, content := range msg.Content {
		switch variant := content.AsAny().(type) {
		case anthropic.TextBlock:
			resp = append(resp, spec.ResponseContent{Type: spec.ResponseContentTypeText, Content: variant.Text})
		case anthropic.ThinkingBlock:
			resp = append(resp, spec.ResponseContent{Type: spec.ResponseContentTypeThinking, Content: variant.Thinking})
		case anthropic.RedactedThinkingBlock:
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
) ([]anthropic.ToolUnionParam, error) {
	if len(tools) == 0 {
		return nil, nil
	}
	out := make([]anthropic.ToolUnionParam, 0, len(tools))
	for _, ct := range tools {
		schema, err := decodeToolArgSchema(ct.Tool.ArgSchema)
		if err != nil {
			return nil, fmt.Errorf(
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

		toolUnion := anthropic.ToolUnionParamOfTool(inputSchema, toolFunctionName(ct))
		if variant := toolUnion.OfTool; variant != nil {
			variant.Type = anthropic.ToolTypeCustom
			if desc := toolDescription(ct); desc != "" {
				variant.Description = anthropic.String(desc)
			}
		}
		out = append(out, toolUnion)
	}
	return out, nil
}

func extractAnthropicToolCalls(msg *anthropic.Message) []spec.ResponseToolCall {
	if msg == nil {
		return nil
	}
	out := make([]spec.ResponseToolCall, 0, len(msg.Content))
	for _, content := range msg.Content {
		variant, ok := content.AsAny().(anthropic.ToolUseBlock)
		if !ok {
			continue
		}
		call := spec.ResponseToolCall{
			ID:        strings.TrimSpace(variant.ID),
			CallID:    strings.TrimSpace(variant.ID),
			Name:      strings.TrimSpace(variant.Name),
			Arguments: strings.TrimSpace(string(variant.Input)),
			Type:      string(variant.Type),
		}
		out = append(out, call)

	}
	if len(out) == 0 {
		return nil
	}
	return out
}
