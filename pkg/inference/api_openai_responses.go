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
		// This is because openai sdk adds 'chat/completions' internally.
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

	// Build OpenAI chat messages.
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
	var toolNameMap map[string]spec.FetchCompletionToolChoice
	if len(completionData.ToolChoices) > 0 {
		toolDefs, m, err := toOpenAIResponseTools(completionData.ToolChoices)
		if err != nil {
			return nil, err
		}
		if len(toolDefs) > 0 {
			params.Tools = toolDefs
			toolNameMap = m
		}
	}

	timeout := modelpresetSpec.DefaultAPITimeout
	if completionData.ModelParams.Timeout > 0 {
		timeout = time.Duration(completionData.ModelParams.Timeout) * time.Second
	}

	if completionData.ModelParams.Stream && onStreamTextData != nil && onStreamThinkingData != nil {
		return api.doStreaming(ctx, params, onStreamTextData, onStreamThinkingData, timeout, toolNameMap)
	}
	return api.doNonStreaming(ctx, params, timeout, toolNameMap)
}

func (api *OpenAIResponsesAPI) doNonStreaming(
	ctx context.Context,
	params responses.ResponseNewParams,
	timeout time.Duration,
	toolNameMap map[string]spec.FetchCompletionToolChoice,
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

	if toolCalls := extractOpenAIResponseToolCalls(resp, toolNameMap); len(toolCalls) > 0 {
		completionResp.Body.ToolCalls = toolCalls
	}

	return completionResp, nil
}

func (api *OpenAIResponsesAPI) doStreaming(
	ctx context.Context,
	params responses.ResponseNewParams,
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

	if toolCalls := extractOpenAIResponseToolCalls(&respFull, toolNameMap); len(toolCalls) > 0 {
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
		if m.Content == nil && len(m.Attachments) == 0 {
			continue
		}
		switch m.Role {
		case modelpresetSpec.RoleSystem:
			out = append(out, responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(*m.Content),
					},
					Role: responses.EasyInputMessageRoleSystem,
					Type: responses.EasyInputMessageTypeMessage,
				},
			})

		case modelpresetSpec.RoleDeveloper:
			out = append(out, responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(*m.Content),
					},
					Role: responses.EasyInputMessageRoleDeveloper,
					Type: responses.EasyInputMessageTypeMessage,
				},
			})

		case modelpresetSpec.RoleUser:
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

			// No attachments: plain text user message.
			if len(attachmentContent) == 0 {
				text := ""
				if m.Content != nil {
					text = *m.Content
				}
				if strings.TrimSpace(text) == "" {
					// Nothing to send for this turn.
					continue
				}
				out = append(out, responses.ResponseInputItemUnionParam{
					OfMessage: &responses.EasyInputMessageParam{
						Content: responses.EasyInputMessageContentUnionParam{
							OfString: openai.String(text),
						},
						Role: responses.EasyInputMessageRoleUser,
						Type: responses.EasyInputMessageTypeMessage,
					},
				})
				continue
			}

			// Text + attachments for this user turn.
			var contentList responses.ResponseInputMessageContentListParam
			if m.Content != nil {
				if c := strings.TrimSpace(*m.Content); c != "" {
					contentList = append(
						contentList,
						responses.ResponseInputContentParamOfInputText(c),
					)
				}
			}
			contentList = append(contentList, attachmentContent...)

			out = append(out, responses.ResponseInputItemParamOfMessage(
				contentList,
				responses.EasyInputMessageRoleUser,
			))

		case modelpresetSpec.RoleAssistant:
			out = append(out, responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(*m.Content),
					},
					Role: responses.EasyInputMessageRoleAssistant,
					Type: responses.EasyInputMessageTypeMessage,
				},
			})

		case modelpresetSpec.RoleFunction, modelpresetSpec.RoleTool:
			// Not used here.
		}
	}
	return out, nil
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
) []spec.ResponseContent {
	if inputResp == nil || len(inputResp.Output) == 0 {
		return []spec.ResponseContent{}
	}
	resp := make([]spec.ResponseContent, 0, 4)
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
		if item.Type == string(openaiSharedConstant.ImageGenerationCall("").Default()) {
			// Image generation outputs return a base64-encoded image. Wrap it in a
			// data URL and expose it as an image content block. The frontend can
			// either render it directly or treat it as markdown.
			data := strings.TrimSpace(item.Result)
			if data != "" {
				dataURL := "data:image/png;base64," + data
				md := fmt.Sprintf("![Generated image](%s)", dataURL)
				resp = append(
					resp,
					spec.ResponseContent{
						Type:    spec.ResponseContentTypeImage,
						Content: md,
					},
				)
			}
		}
	}

	thinkingSummaryStr := thinkingSummaryText.String()
	if thinkingSummaryStr != "" {
		resp = append(
			resp,
			spec.ResponseContent{
				Type:    spec.ResponseContentTypeThinkingSummary,
				Content: thinkingSummaryStr,
			},
		)
	}

	thinkingStr := thinkingText.String()
	if thinkingStr != "" {
		resp = append(
			resp,
			spec.ResponseContent{Type: spec.ResponseContentTypeThinking, Content: thinkingStr},
		)
	}
	outStr := outputText.String()

	resp = append(
		resp,
		spec.ResponseContent{Type: spec.ResponseContentTypeText, Content: outStr},
	)

	return resp
}

func toOpenAIResponseTools(
	tools []spec.FetchCompletionToolChoice,
) ([]responses.ToolUnionParam, map[string]spec.FetchCompletionToolChoice, error) {
	if len(tools) == 0 {
		return nil, nil, nil
	}

	ordered, nameMap := buildToolNameMapping(tools)
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
	toolNameMap map[string]spec.FetchCompletionToolChoice,
) []toolSpec.ToolCall {
	if resp == nil {
		return nil
	}

	out := make([]toolSpec.ToolCall, 0)

	for _, item := range resp.Output {
		switch item.Type {
		case string(openaiSharedConstant.FunctionCall("").Default()):
			fn := item.AsFunctionCall()

			name := fn.Name
			var toolChoice *toolSpec.ToolChoice
			if toolNameMap != nil {
				if ct, ok := toolNameMap[name]; ok {
					name = toolIdentityFromChoice(ct)
					// Add the actual choice to response.
					toolChoice = &ct.ToolChoice
				}
			}

			out = append(
				out,
				toolSpec.ToolCall{
					ID:         fn.ID,
					CallID:     fn.CallID,
					Name:       name,
					Arguments:  fn.Arguments,
					Type:       item.Type,
					Status:     string(fn.Status),
					ToolChoice: toolChoice,
				},
			)

		case string(openaiSharedConstant.CustomToolCall("").Default()):
			fn := item.AsCustomToolCall()

			name := fn.Name
			var toolChoice *toolSpec.ToolChoice
			if toolNameMap != nil {
				if ct, ok := toolNameMap[name]; ok {
					name = toolIdentityFromChoice(ct)
					toolChoice = &ct.ToolChoice
				}
			}

			out = append(
				out,
				toolSpec.ToolCall{
					ID:         fn.ID,
					CallID:     fn.CallID,
					Name:       name,
					Arguments:  fn.Input,
					Type:       item.Type,
					ToolChoice: toolChoice,
				},
			)
		}
	}

	if len(out) == 0 {
		return nil
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
