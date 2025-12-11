package inference

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
	openaiSharedConstant "github.com/openai/openai-go/v3/shared/constant"

	"github.com/ppipada/flexigpt-app/pkg/attachment"
	"github.com/ppipada/flexigpt-app/pkg/fileutil"
	"github.com/ppipada/flexigpt-app/pkg/inference/debugclient"
	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"
)

// OpenAIResponsesAPI struct that implements the CompletionProvider interface.
type OpenAIResponsesAPI struct {
	ProviderParams *spec.ProviderParams
	Debug          bool
	client         *openai.Client
}

func NewOpenAIResponsesAPI(
	pi spec.ProviderParams,
	debug bool,
) (*OpenAIResponsesAPI, error) {
	if pi.Name == "" || pi.Origin == "" {
		return nil, errors.New("openai responses api LLM: invalid args")
	}
	return &OpenAIResponsesAPI{
		ProviderParams: &pi,
		Debug:          debug,
	}, nil
}

func (api *OpenAIResponsesAPI) InitLLM(ctx context.Context) error {
	if !api.IsConfigured(ctx) {
		slog.Debug(
			string(
				api.ProviderParams.Name,
			) + ": No API key given. Not initializing OpenAIResponsesAPI LLM object",
		)
		return nil
	}

	opts := []option.RequestOption{
		option.WithAPIKey(api.ProviderParams.APIKey),
	}

	providerURL := modelpresetSpec.DefaultOpenAIOrigin
	if api.ProviderParams.Origin != "" {
		baseURL := api.ProviderParams.Origin
		// Remove trailing slash from baseURL if present.
		baseURL = strings.TrimSuffix(baseURL, "/")

		pathPrefix := api.ProviderParams.ChatCompletionPathPrefix
		// Remove 'responses' from pathPrefix if present,
		// This is because openai sdk adds 'responses' internally.
		pathPrefix = strings.TrimSuffix(
			pathPrefix,
			"responses",
		)
		providerURL = baseURL + pathPrefix
		opts = append(opts, option.WithBaseURL(strings.TrimSuffix(providerURL, "/")))
	}

	for k, v := range api.ProviderParams.DefaultHeaders {
		opts = append(opts, option.WithHeader(strings.TrimSpace(k), strings.TrimSpace(v)))
	}

	if api.ProviderParams.APIKeyHeaderKey != "" &&
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

	c := openai.NewClient(opts...)
	api.client = &c
	slog.Info(
		"openai responses api LLM provider initialized",
		"name",
		string(api.ProviderParams.Name),
		"URL",
		providerURL,
	)
	return nil
}

func (api *OpenAIResponsesAPI) DeInitLLM(ctx context.Context) error {
	api.client = nil
	slog.Info(
		"openai responses api LLM: provider de initialized",
		"name",
		string(api.ProviderParams.Name),
	)
	return nil
}

func (api *OpenAIResponsesAPI) GetProviderInfo(ctx context.Context) *spec.ProviderParams {
	return api.ProviderParams
}

func (api *OpenAIResponsesAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderParams.APIKey != ""
}

// SetProviderAPIKey sets the key for a provider.
func (api *OpenAIResponsesAPI) SetProviderAPIKey(
	ctx context.Context,
	apiKey string,
) error {
	if apiKey == "" {
		return errors.New("openai responses api LLM: invalid apikey provided")
	}
	if api.ProviderParams == nil {
		return errors.New("openai responses api LLM: no ProviderParams found")
	}

	api.ProviderParams.APIKey = apiKey

	return nil
}

func (api *OpenAIResponsesAPI) BuildCompletionData(
	ctx context.Context,
	modelParams spec.ModelParams,
	currentMessage spec.ChatCompletionDataMessage,
	prevMessages []spec.ChatCompletionDataMessage,
) (*spec.FetchCompletionData, error) {
	return getCompletionData(modelParams, currentMessage, prevMessages), nil
}

func (api *OpenAIResponsesAPI) FetchCompletion(
	ctx context.Context,
	completionData *spec.FetchCompletionData,
	onStreamTextData, onStreamThinkingData func(string) error,
) (*spec.FetchCompletionResponse, error) {
	if api.client == nil {
		return nil, errors.New("openai responses api LLM: client not initialized")
	}
	if completionData == nil || len(completionData.Messages) == 0 {
		return nil, errors.New("openai responses api LLM: empty completion data")
	}

	// Build OpenAI Responses input messages.
	msgs, err := toOpenAIResponsesMessages(
		ctx,
		completionData.Messages,
		completionData.ModelParams.Name,
		api.ProviderParams.Name,
	)
	if err != nil {
		return nil, err
	}

	params := responses.ResponseNewParams{
		Model:           shared.ChatModel(completionData.ModelParams.Name),
		MaxOutputTokens: openai.Int(int64(completionData.ModelParams.MaxOutputLength)),
		Input:           responses.ResponseNewParamsInputUnion{OfInputItemList: msgs},
		Store:           openai.Bool(false),
	}

	sysPrompt := strings.TrimSpace(completionData.ModelParams.SystemPrompt)
	if sysPrompt != "" {
		params.Instructions = openai.String(sysPrompt)
	}
	if completionData.ModelParams.Temperature != nil {
		params.Temperature = openai.Float(*completionData.ModelParams.Temperature)
	}

	if rp := completionData.ModelParams.Reasoning; rp != nil &&
		rp.Type == modelpresetSpec.ReasoningTypeSingleWithLevels {
		switch rp.Level {
		case
			modelpresetSpec.ReasoningLevelNone,
			modelpresetSpec.ReasoningLevelMinimal,
			modelpresetSpec.ReasoningLevelLow,
			modelpresetSpec.ReasoningLevelMedium,
			modelpresetSpec.ReasoningLevelHigh:
			params.Reasoning = shared.ReasoningParam{
				Effort:  shared.ReasoningEffort(string(rp.Level)),
				Summary: shared.ReasoningSummaryAuto,
			}
		default:
			return nil, fmt.Errorf("invalid level %q for singleWithLevels", rp.Level)
		}
	}

	// Build tools + tool name map.
	var toolChoiceNameMap map[string]spec.FetchCompletionToolChoice
	if len(completionData.ToolChoices) > 0 {
		toolDefs, m, err := toOpenAIResponseTools(completionData.ToolChoices)
		if err != nil {
			return nil, err
		}
		if len(toolDefs) > 0 {
			params.Tools = toolDefs
			toolChoiceNameMap = m
		}
	}

	timeout := modelpresetSpec.DefaultAPITimeout
	if completionData.ModelParams.Timeout > 0 {
		timeout = time.Duration(completionData.ModelParams.Timeout) * time.Second
	}

	if completionData.ModelParams.Stream && onStreamTextData != nil && onStreamThinkingData != nil {
		return api.doStreaming(ctx, params, onStreamTextData, onStreamThinkingData, timeout, toolChoiceNameMap)
	}
	return api.doNonStreaming(ctx, params, timeout, toolChoiceNameMap)
}

func (api *OpenAIResponsesAPI) doNonStreaming(
	ctx context.Context,
	params responses.ResponseNewParams,
	timeout time.Duration,
	toolChoiceNameMap map[string]spec.FetchCompletionToolChoice,
) (*spec.FetchCompletionResponse, error) {
	completionResp := &spec.FetchCompletionResponse{Body: &spec.FetchCompletionResponseBody{}}
	ctx = debugclient.AddDebugResponseToCtx(ctx)

	resp, err := api.client.Responses.New(ctx, params, option.WithRequestTimeout(timeout))
	isNilResp := resp == nil || len(resp.Output) == 0

	attachDebugResp(ctx, completionResp.Body, err, isNilResp, "", resp)
	completionResp.Body.Usage = usageFromOpenAIResponse(resp)

	if isNilResp {
		return completionResp, nil
	}

	completionResp.Body.ResponseContent = getResponseContentFromOpenAIOutput(resp)

	if toolCalls := extractOpenAIResponseToolCalls(resp, toolChoiceNameMap); len(toolCalls) > 0 {
		completionResp.Body.ToolCalls = toolCalls
	}

	return completionResp, nil
}

func (api *OpenAIResponsesAPI) doStreaming(
	ctx context.Context,
	params responses.ResponseNewParams,
	onStreamTextData, onStreamThinkingData func(string) error,
	timeout time.Duration,
	toolChoiceNameMap map[string]spec.FetchCompletionToolChoice,
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

	var respFull responses.Response

	completionResp := &spec.FetchCompletionResponse{Body: &spec.FetchCompletionResponseBody{}}
	ctx = debugclient.AddDebugResponseToCtx(ctx)
	stream := api.client.Responses.NewStreaming(
		ctx,
		params,
		option.WithRequestTimeout(timeout),
	)

	var streamWriteErr error
	for stream.Next() {
		chunk := stream.Current()
		if chunk.Type == "response.output_text.delta" {
			streamWriteErr = writeTextData(chunk.Delta)
			if streamWriteErr != nil {
				break
			}
		}
		if chunk.Type == "response.reasoning_summary_text.delta" {
			streamWriteErr = writeThinkingData(chunk.Delta)
			if streamWriteErr != nil {
				break
			}
		}
		if chunk.Type == "response.reasoning_text.delta" {
			streamWriteErr = writeThinkingData(chunk.Delta)
			if streamWriteErr != nil {
				break
			}
		}

		if chunk.Type == "response.completed" {
			respFull = chunk.Response
			break
		}
		if chunk.Type == "response.failed" {
			respFull = chunk.Response
			streamWriteErr = fmt.Errorf("API failed, %s", respFull.Error.RawJSON())
			break
		}
		if chunk.Type == "response.incomplete" {
			respFull = chunk.Response
			streamWriteErr = fmt.Errorf(
				"API finished as incomplete, %s",
				respFull.IncompleteDetails.Reason,
			)
			break
		}
	}
	if flushTextData != nil {
		flushTextData()
	}
	if flushThinkingData != nil {
		flushThinkingData()
	}

	streamErr := errors.Join(stream.Err(), streamWriteErr)
	isNilResp := len(respFull.Output) == 0

	attachDebugResp(ctx, completionResp.Body, streamErr, isNilResp, "", respFull)
	completionResp.Body.Usage = usageFromOpenAIResponse(&respFull)

	if isNilResp {
		return completionResp, nil
	}

	respContent := getResponseContentFromOpenAIOutput(&respFull)
	completionResp.Body.ResponseContent = respContent

	if toolCalls := extractOpenAIResponseToolCalls(&respFull, toolChoiceNameMap); len(toolCalls) > 0 {
		completionResp.Body.ToolCalls = toolCalls
	}

	return completionResp, streamErr
}

func toOpenAIResponsesMessages(
	ctx context.Context,
	messages []spec.ChatCompletionDataMessage,
	modelName modelpresetSpec.ModelName,
	providerName modelpresetSpec.ProviderName,
) (responses.ResponseInputParam, error) {
	var out responses.ResponseInputParam

	// Identify last user message index; attachments belong to the current user turn.
	lastUserIdx := -1
	for i, m := range messages {
		if m.Role == modelpresetSpec.RoleUser {
			lastUserIdx = i
		}
	}

	for idx, m := range messages {
		// Skip only if there is no text, no attachments, no tool outputs, and no tool calls.
		if m.Content == nil &&
			len(m.Attachments) == 0 &&
			len(m.ToolOutputs) == 0 &&
			len(m.ToolCalls) == 0 {
			continue
		}

		// Normalized trimmed content string (may be empty).
		content := ""
		if m.Content != nil {
			content = strings.TrimSpace(*m.Content)
		}

		switch m.Role {
		case modelpresetSpec.RoleSystem:
			if content == "" {
				continue
			}
			out = append(out, responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(content),
					},
					Role: responses.EasyInputMessageRoleSystem,
					Type: responses.EasyInputMessageTypeMessage,
				},
			})

		case modelpresetSpec.RoleDeveloper:
			if content == "" {
				continue
			}
			out = append(out, responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(content),
					},
					Role: responses.EasyInputMessageRoleDeveloper,
					Type: responses.EasyInputMessageTypeMessage,
				},
			})

		case modelpresetSpec.RoleUser:
			if m.Content == nil && len(m.Attachments) == 0 && len(m.ToolOutputs) == 0 {
				continue
			}
			// Build attachments for this specific user turn (if any).
			var attachmentContent []responses.ResponseInputContentUnionParam
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
					return nil, err
				}
				attachmentContent, err = contentBlocksToOpenAI(blocks)
				if err != nil {
					return nil, err
				}
			}

			var parts responses.ResponseInputMessageContentListParam
			if content != "" {
				parts = append(
					parts,
					responses.ResponseInputContentParamOfInputText(content),
				)
			}

			// No attachments: plain text user message.
			if len(attachmentContent) != 0 {
				parts = append(parts, attachmentContent...)
			}

			if len(parts) > 0 {
				out = append(out, responses.ResponseInputItemParamOfMessage(
					parts,
					responses.EasyInputMessageRoleUser,
				))
			}
			if len(m.ToolOutputs) > 0 {
				o := toolOutputsToOpenAIResponses(m.ToolOutputs)
				out = append(out, o...)
			}

		case modelpresetSpec.RoleAssistant:
			// Integrate assistant text + tool calls into a single assistant message.
			if content != "" {
				out = append(out, responses.ResponseInputItemUnionParam{
					OfMessage: &responses.EasyInputMessageParam{
						Content: responses.EasyInputMessageContentUnionParam{
							OfString: openai.String(content),
						},
						Role: responses.EasyInputMessageRoleAssistant,
						Type: responses.EasyInputMessageTypeMessage,
					},
				})
			}
			if len(m.ToolCalls) == 0 {
				continue
			}

			var parts []responses.ResponseInputItemUnionParam

			for _, toolCall := range m.ToolCalls {
				switch toolCall.Type {
				case string(openaiSharedConstant.FunctionCall("").Default()):
					fc := responses.ResponseFunctionToolCallParam{
						ID:        param.NewOpt(toolCall.ID),
						CallID:    toolCall.CallID,
						Name:      toolCall.Name,
						Arguments: toolCall.Arguments,
						Type:      openaiSharedConstant.FunctionCall("").Default(),
					}

					parts = append(parts,
						responses.ResponseInputItemUnionParam{
							OfFunctionCall: &fc,
						},
					)

				case string(openaiSharedConstant.CustomToolCall("").Default()):
					cc := responses.ResponseCustomToolCallParam{
						ID:     param.NewOpt(toolCall.ID),
						CallID: toolCall.CallID,
						Name:   toolCall.Name,
						Input:  toolCall.Arguments,
						Type:   openaiSharedConstant.CustomToolCall("").Default(),
					}

					parts = append(parts,
						responses.ResponseInputItemUnionParam{
							OfCustomToolCall: &cc,
						},
					)
				}
			}

			if len(parts) != 0 {
				out = append(out, parts...)
			}

		case modelpresetSpec.RoleFunction, modelpresetSpec.RoleTool:
			if len(m.ToolOutputs) > 0 {
				o := toolOutputsToOpenAIResponses(m.ToolOutputs)
				out = append(out, o...)
			}
			if strings.TrimSpace(content) != "" {
				out = append(out, responses.ResponseInputItemUnionParam{
					OfMessage: &responses.EasyInputMessageParam{
						Content: responses.EasyInputMessageContentUnionParam{
							OfString: openai.String(content),
						},
						Role: responses.EasyInputMessageRoleUser,
						Type: responses.EasyInputMessageTypeMessage,
					},
				})
			}
		}
	}
	return out, nil
}

func toolOutputsToOpenAIResponses(
	toolOutputs []toolSpec.ToolOutput,
) []responses.ResponseInputItemUnionParam {
	out := make([]responses.ResponseInputItemUnionParam, 0, len(toolOutputs))

	// Tool outputs with a callID can be represented as proper tool messages.
	for _, o := range toolOutputs {
		callID := strings.TrimSpace(o.CallID)
		if callID == "" {
			continue
		}
		toolContent := strings.TrimSpace(o.RawOutput)
		if toolContent == "" {
			toolContent = strings.TrimSpace(o.Summary)
		}
		if toolContent == "" {
			continue
		}
		out = append(out, responses.ResponseInputItemParamOfFunctionCallOutput(callID, toolContent))
	}

	// Any tool outputs without a callID are rendered as plain user text
	// so the model still sees the information.
	var orphanOutputs []toolSpec.ToolOutput
	for _, o := range toolOutputs {
		if strings.TrimSpace(o.CallID) == "" {
			orphanOutputs = append(orphanOutputs, o)
		}
	}
	if len(orphanOutputs) > 0 {
		if text := renderToolOutputsAsText(orphanOutputs); text != "" {
			o := responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(text),
					},
					Role: responses.EasyInputMessageRoleUser,
					Type: responses.EasyInputMessageTypeMessage,
				},
			}

			out = append(out, o)
		}
	}
	return out
}

func contentBlocksToOpenAI(
	blocks []attachment.ContentBlock,
) ([]responses.ResponseInputContentUnionParam, error) {
	if len(blocks) == 0 {
		return nil, nil
	}
	out := make([]responses.ResponseInputContentUnionParam, 0, len(blocks))

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
		fname := ""
		if b.FileName != nil {
			fname = strings.TrimSpace(*b.FileName)
		}
		switch b.Kind {
		case attachment.ContentBlockText:
			if txt == "" {
				continue
			}
			out = append(
				out,
				responses.ResponseInputContentParamOfInputText(txt),
			)

		case attachment.ContentBlockImage:
			if bData == "" {
				continue
			}
			if mime == "" {
				mime = string(fileutil.DefaultImageMIME)
			}
			dataURL := fmt.Sprintf("data:%s;base64,%s", mime, bData)
			img := responses.ResponseInputImageParam{
				Detail:   responses.ResponseInputImageDetailAuto,
				ImageURL: param.NewOpt(dataURL),
			}
			out = append(
				out,
				responses.ResponseInputContentUnionParam{OfInputImage: &img},
			)

		case attachment.ContentBlockFile:
			if bData == "" {
				continue
			}

			if mime == "" {
				mime = string(fileutil.MIMEApplicationOctetStream)
			}
			dataURL := fmt.Sprintf("data:%s;base64,%s", mime, bData)
			fileParam := responses.ResponseInputFileParam{
				FileData: param.NewOpt(dataURL),
				Filename: param.NewOpt(fname),
				Type:     openaiSharedConstant.InputFile("").Default(),
			}
			out = append(
				out,
				responses.ResponseInputContentUnionParam{OfInputFile: &fileParam},
			)

		default:
			continue
		}
	}
	return out, nil
}

func getResponseContentFromOpenAIOutput(
	inputResp *responses.Response,
) []modelpresetSpec.MessageContent {
	if inputResp == nil || len(inputResp.Output) == 0 {
		return []modelpresetSpec.MessageContent{}
	}
	resp := make([]modelpresetSpec.MessageContent, 0, 4)
	var outputText strings.Builder
	var thinkingSummaryText strings.Builder
	var thinkingText strings.Builder

	for _, item := range inputResp.Output {
		if item.Type == string(openaiSharedConstant.Message("").Default()) {
			for _, content := range item.Content {
				if content.Type == string(openaiSharedConstant.OutputText("").Default()) {
					outputText.WriteString(content.Text)
				}
			}
		}
		if item.Type == string(openaiSharedConstant.Reasoning("").Default()) {
			ti := item.AsReasoning()
			for _, content := range ti.Content {
				thinkingText.WriteString(content.Text)
			}
			for _, content := range ti.Summary {
				thinkingSummaryText.WriteString(content.Text)
			}
		}
	}

	thinkingSummaryStr := thinkingSummaryText.String()
	if thinkingSummaryStr != "" {
		resp = append(
			resp,
			modelpresetSpec.MessageContent{
				Type:    modelpresetSpec.MessageContentTypeThinkingSummary,
				Content: thinkingSummaryStr,
			},
		)
	}

	thinkingStr := thinkingText.String()
	if thinkingStr != "" {
		resp = append(
			resp,
			modelpresetSpec.MessageContent{Type: modelpresetSpec.MessageContentTypeThinking, Content: thinkingStr},
		)
	}
	outStr := outputText.String()

	resp = append(
		resp,
		modelpresetSpec.MessageContent{Type: modelpresetSpec.MessageContentTypeText, Content: outStr},
	)

	return resp
}

func toOpenAIResponseTools(
	tools []spec.FetchCompletionToolChoice,
) ([]responses.ToolUnionParam, map[string]spec.FetchCompletionToolChoice, error) {
	if len(tools) == 0 {
		return nil, nil, nil
	}

	ordered, nameMap := buildToolChoiceNameMapping(tools)
	out := make([]responses.ToolUnionParam, 0, len(ordered))

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

		tool := responses.FunctionToolParam{
			// Use the derived name based only on slug + disambiguation.
			Name:       tw.Name,
			Parameters: schema,
			Type:       openaiSharedConstant.Function("function"),
		}
		if desc := toolDescription(ct); desc != "" {
			tool.Description = openai.String(desc)
		}

		out = append(out, responses.ToolUnionParam{OfFunction: &tool})
	}

	return out, nameMap, nil
}

func extractOpenAIResponseToolCalls(
	resp *responses.Response,
	toolChoiceNameMap map[string]spec.FetchCompletionToolChoice,
) []toolSpec.ToolCall {
	if resp == nil {
		return []toolSpec.ToolCall{}
	}

	out := make([]toolSpec.ToolCall, 0)

	for _, item := range resp.Output {
		switch item.Type {
		case string(openaiSharedConstant.FunctionCall("").Default()):
			fn := item.AsFunctionCall()
			if fn.CallID == "" || strings.TrimSpace(fn.Name) == "" {
				continue
			}
			name := fn.Name
			var toolChoice *toolSpec.ToolChoice
			if toolChoiceNameMap != nil {
				if ct, ok := toolChoiceNameMap[name]; ok {
					// Add the actual choice to response.
					toolChoice = &ct.ToolChoice
				}
			}

			out = append(
				out,
				toolSpec.ToolCall{
					ID:         fn.ID,
					CallID:     fn.CallID,
					Name:       fn.Name,
					Arguments:  fn.Arguments,
					Type:       item.Type,
					Status:     string(fn.Status),
					ToolChoice: toolChoice,
				},
			)

		case string(openaiSharedConstant.CustomToolCall("").Default()):
			fn := item.AsCustomToolCall()
			if fn.CallID == "" || strings.TrimSpace(fn.Name) == "" {
				continue
			}
			name := fn.Name
			var toolChoice *toolSpec.ToolChoice
			if toolChoiceNameMap != nil {
				if ct, ok := toolChoiceNameMap[name]; ok {
					toolChoice = &ct.ToolChoice
				}
			}

			out = append(
				out,
				toolSpec.ToolCall{
					ID:         fn.ID,
					CallID:     fn.CallID,
					Name:       fn.Name,
					Arguments:  fn.Input,
					Type:       item.Type,
					ToolChoice: toolChoice,
				},
			)
		}
	}

	if len(out) == 0 {
		return []toolSpec.ToolCall{}
	}
	return out
}

// usageFromOpenAIResponse normalizes OpenAI Responses API usage into spec.Usage.
func usageFromOpenAIResponse(resp *responses.Response) *modelpresetSpec.Usage {
	uOut := &modelpresetSpec.Usage{}
	if resp == nil {
		return uOut
	}

	u := resp.Usage

	uOut.InputTokensTotal = u.InputTokens
	uOut.InputTokensCached = u.InputTokensDetails.CachedTokens
	uOut.InputTokensUncached = max(u.InputTokens-u.InputTokensDetails.CachedTokens, 0)
	uOut.OutputTokens = u.OutputTokens
	uOut.ReasoningTokens = u.OutputTokensDetails.ReasoningTokens

	return uOut
}
