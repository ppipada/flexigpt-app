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
	"github.com/openai/openai-go/v3/shared"
	openaiSharedConstant "github.com/openai/openai-go/v3/shared/constant"

	"github.com/ppipada/flexigpt-app/pkg/attachment"
	"github.com/ppipada/flexigpt-app/pkg/fileutil"
	"github.com/ppipada/flexigpt-app/pkg/inference/debugclient"
	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"

	inferencegoSpec "github.com/ppipada/inference-go/spec"
)

// OpenAIChatCompletionsAPI struct that implements the CompletionProvider interface.
type OpenAIChatCompletionsAPI struct {
	ProviderParams *inferencegoSpec.ProviderParam
	Debug          bool
	client         *openai.Client
}

func NewOpenAIChatCompletionsAPI(
	pi inferencegoSpec.ProviderParam,
	debug bool,
) (*OpenAIChatCompletionsAPI, error) {
	if pi.Name == "" || pi.Origin == "" {
		return nil, errors.New("openai chat completions api LLM: invalid args")
	}
	return &OpenAIChatCompletionsAPI{
		ProviderParams: &pi,
		Debug:          debug,
	}, nil
}

func (api *OpenAIChatCompletionsAPI) InitLLM(ctx context.Context) error {
	if !api.IsConfigured(ctx) {
		slog.Debug(
			string(
				api.ProviderParams.Name,
			) + ": No API key given. Not initializing OpenAIChatCompletionsAPI LLM object",
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
		// Remove 'chat/completions' from pathPrefix if present,
		// This is because openai sdk adds 'chat/completions' internally.
		pathPrefix = strings.TrimSuffix(
			pathPrefix,
			"chat/completions",
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
		"openai chat completions api LLM provider initialized",
		"name",
		string(api.ProviderParams.Name),
		"URL",
		providerURL,
	)
	return nil
}

func (api *OpenAIChatCompletionsAPI) DeInitLLM(ctx context.Context) error {
	api.client = nil
	slog.Info(
		"openai chat completions api LLM: provider de initialized",
		"name",
		string(api.ProviderParams.Name),
	)
	return nil
}

func (api *OpenAIChatCompletionsAPI) GetProviderInfo(ctx context.Context) *inferencegoSpec.ProviderParam {
	return api.ProviderParams
}

func (api *OpenAIChatCompletionsAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderParams.APIKey != ""
}

// SetProviderAPIKey sets the key for a provider.
func (api *OpenAIChatCompletionsAPI) SetProviderAPIKey(
	ctx context.Context,
	apiKey string,
) error {
	if apiKey == "" {
		return errors.New("openai chat completions api LLM: invalid apikey provided")
	}
	if api.ProviderParams == nil {
		return errors.New("openai chat completions api LLM: no ProviderParams found")
	}

	api.ProviderParams.APIKey = apiKey

	return nil
}

func (api *OpenAIChatCompletionsAPI) BuildCompletionData(
	ctx context.Context,
	modelParams inferencegoSpec.ModelParam,
	currentMessage spec.ChatCompletionDataMessage,
	prevMessages []spec.ChatCompletionDataMessage,
) (*spec.FetchCompletionData, error) {
	return getCompletionData(modelParams, currentMessage, prevMessages), nil
}

func (api *OpenAIChatCompletionsAPI) FetchCompletion(
	ctx context.Context,
	completionData *spec.FetchCompletionData,
	onStreamTextData, onStreamThinkingData func(string) error,
) (*spec.FetchCompletionResponse, error) {
	if api.client == nil {
		return nil, errors.New("openai chat completions api LLM: client not initialized")
	}
	if completionData == nil || len(completionData.Messages) == 0 {
		return nil, errors.New("openai chat completions api LLM: empty completion data")
	}

	// Build OpenAI chat messages.
	msgs, err := toOpenAIChatMessages(
		ctx,
		completionData.ModelParams.SystemPrompt,
		completionData.Messages,
		completionData.ModelParams.Name,
		api.ProviderParams.Name,
	)
	if err != nil {
		return nil, err
	}

	params := openai.ChatCompletionNewParams{
		Model:               shared.ChatModel(completionData.ModelParams.Name),
		MaxCompletionTokens: openai.Int(int64(completionData.ModelParams.MaxOutputLength)),
		Messages:            msgs,
	}
	if completionData.ModelParams.Temperature != nil {
		params.Temperature = openai.Float(*completionData.ModelParams.Temperature)
	}

	if rp := completionData.ModelParams.Reasoning; rp != nil &&
		rp.Type == inferencegoSpec.ReasoningTypeSingleWithLevels {
		switch rp.Level {
		case
			inferencegoSpec.ReasoningLevelNone,
			inferencegoSpec.ReasoningLevelMinimal,
			inferencegoSpec.ReasoningLevelLow,
			inferencegoSpec.ReasoningLevelMedium,
			inferencegoSpec.ReasoningLevelHigh,
			inferencegoSpec.ReasoningLevelXHigh:
			params.ReasoningEffort = shared.ReasoningEffort(string(rp.Level))
		default:
			return nil, fmt.Errorf("invalid level %q for singleWithLevels", rp.Level)

		}
	}
	var toolChoiceNameMap map[string]spec.FetchCompletionToolChoice
	if len(completionData.ToolChoices) > 0 {
		toolDefs, m, err := toOpenAIChatTools(completionData.ToolChoices)
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

func (api *OpenAIChatCompletionsAPI) doNonStreaming(
	ctx context.Context,
	params openai.ChatCompletionNewParams,
	timeout time.Duration,
	toolChoiceNameMap map[string]spec.FetchCompletionToolChoice,
) (*spec.FetchCompletionResponse, error) {
	completionResp := &spec.FetchCompletionResponse{Body: &spec.FetchCompletionResponseBody{}}
	ctx = debugclient.AddDebugResponseToCtx(ctx)
	resp, err := api.client.Chat.Completions.New(ctx, params, option.WithRequestTimeout(timeout))
	isNilResp := resp == nil || len(resp.Choices) == 0
	attachDebugResp(ctx, completionResp.Body, err, isNilResp, "", resp)
	completionResp.Body.Usage = usageFromOpenAIChatCompletion(resp)
	if isNilResp {
		return completionResp, nil
	}
	full := resp.Choices[0].Message.Content
	completionResp.Body.Content = &full
	if toolCalls := extractOpenAIChatToolCalls(resp.Choices, toolChoiceNameMap); len(toolCalls) > 0 {
		completionResp.Body.ToolCalls = toolCalls
	}
	return completionResp, nil
}

func (api *OpenAIChatCompletionsAPI) doStreaming(
	ctx context.Context,
	params openai.ChatCompletionNewParams,
	onStreamTextData, onStreamThinkingData func(string) error,
	timeout time.Duration,
	toolChoiceNameMap map[string]spec.FetchCompletionToolChoice,
) (*spec.FetchCompletionResponse, error) {
	// No thinking data available in openai chat completions API, hence no thinking writer.
	write, flush := NewBufferedStreamer(onStreamTextData, FlushInterval, FlushChunkSize)

	completionResp := &spec.FetchCompletionResponse{Body: &spec.FetchCompletionResponseBody{}}
	ctx = debugclient.AddDebugResponseToCtx(ctx)
	stream := api.client.Chat.Completions.NewStreaming(
		ctx,
		params,
		option.WithRequestTimeout(timeout),
	)
	acc := openai.ChatCompletionAccumulator{}
	var streamWriteErr error
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		// When this fires, the current chunk value will not contain content data.
		if _, ok := acc.JustFinishedContent(); ok {
			continue
		}

		if _, ok := acc.JustFinishedRefusal(); ok {
			continue
		}

		if _, ok := acc.JustFinishedToolCall(); ok {
			continue
		}

		// It's best to use chunks after handling JustFinished events.
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			streamWriteErr = write(chunk.Choices[0].Delta.Content)
			if streamWriteErr != nil {
				break
			}
		}
	}
	if flush != nil {
		flush()
	}

	streamErr := errors.Join(stream.Err(), streamWriteErr)
	isNilResp := len(acc.Choices) == 0

	attachDebugResp(ctx, completionResp.Body, streamErr, isNilResp, "", acc.ChatCompletion)
	completionResp.Body.Usage = usageFromOpenAIChatCompletion(&acc.ChatCompletion)

	if isNilResp {
		return completionResp, nil
	}

	fullResp := acc.Choices[0].Message.Content
	completionResp.Body.Content = &fullResp
	if toolCalls := extractOpenAIChatToolCalls(acc.Choices, toolChoiceNameMap); len(toolCalls) > 0 {
		completionResp.Body.ToolCalls = toolCalls
	}
	return completionResp, streamErr
}

func toOpenAIChatMessages(
	ctx context.Context,
	systemPrompt string,
	messages []spec.ChatCompletionDataMessage,
	modelName inferencegoSpec.ModelName,
	providerName inferencegoSpec.ProviderName,
) ([]openai.ChatCompletionMessageParamUnion, error) {
	var out []openai.ChatCompletionMessageParamUnion

	// System/developer prompt.
	if msg := getOpenAIMessageFromSystemPrompt(string(providerName), string(modelName), systemPrompt); msg != nil {
		out = append(out, *msg)
	}

	// Identify the last user message; attachments are associated with the
	// "current" user turn in the frontend, which is always the last user role.
	lastUserIdx := -1
	for i, m := range messages {
		if m.Role == modelpresetSpec.RoleUser {
			lastUserIdx = i
		}
	}

	for idx, m := range messages {
		if m.Content == nil && len(m.Attachments) == 0 && len(m.ToolOutputs) == 0 && len(m.ToolCalls) == 0 {
			continue
		}
		// Normalized trimmed content string (may be empty).
		content := ""
		if m.Content != nil {
			content = strings.TrimSpace(*m.Content)
		}

		switch m.Role {

		case modelpresetSpec.RoleSystem:
			if content != "" {
				out = append(out, openai.SystemMessage(content))
			}

		case modelpresetSpec.RoleDeveloper:
			if content != "" {
				out = append(out, openai.DeveloperMessage(content))
			}
		case modelpresetSpec.RoleUser:
			if m.Content == nil && len(m.Attachments) == 0 && len(m.ToolOutputs) == 0 {
				continue
			}
			var attachmentContent []openai.ChatCompletionContentPartUnionParam
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
				attachmentContent, err = contentBlocksToOpenAIChat(blocks)
				if err != nil {
					return nil, err
				}
			}

			var parts []openai.ChatCompletionContentPartUnionParam
			if content != "" {
				parts = append(parts, openai.TextContentPart(content))
			}

			if len(attachmentContent) != 0 {
				parts = append(parts, attachmentContent...)
			}

			if len(parts) > 0 {
				out = append(out, openai.UserMessage(parts))
			}

			// Map structured tool outputs, if present, to OpenAI tool messages.
			// Input tools may be in user or tool role, so process both.
			if len(m.ToolOutputs) > 0 {
				o := toolOutputsToOpenAIChat(m.ToolOutputs)
				out = append(out, o...)
			}

		case modelpresetSpec.RoleAssistant:
			if content == "" && len(m.ToolCalls) == 0 {
				// Nothing to send.
				continue
			}

			var toolCalls []openai.ChatCompletionMessageToolCallUnionParam
			for _, toolCall := range m.ToolCalls {
				switch toolCall.Type {
				case string(openaiSharedConstant.Function("").Default()):
					f := openai.ChatCompletionMessageFunctionToolCallParam{
						ID: toolCall.ID,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Arguments: toolCall.Arguments,
							Name:      toolCall.Name,
						},
						Type: openaiSharedConstant.Function("").Default(),
					}
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &f,
					})

				case string(openaiSharedConstant.Custom("").Default()):
					f := openai.ChatCompletionMessageCustomToolCallParam{
						ID: toolCall.ID,
						Custom: openai.ChatCompletionMessageCustomToolCallCustomParam{
							Input: toolCall.Arguments,
							Name:  toolCall.Name,
						},
						Type: openaiSharedConstant.Custom("").Default(),
					}
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
						OfCustom: &f,
					})
				}
			}

			if len(toolCalls) == 0 {
				// No tool calls: plain assistant message.
				out = append(out, openai.AssistantMessage(content))
				continue
			}

			assistant := openai.ChatCompletionAssistantMessageParam{
				Role:      openaiSharedConstant.Assistant("").Default(),
				ToolCalls: toolCalls,
			}
			if content != "" {
				assistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: param.NewOpt(content),
				}
			}

			out = append(out, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistant})

		case modelpresetSpec.RoleFunction, modelpresetSpec.RoleTool:
			// Map structured tool outputs, if present, to OpenAI tool messages.
			// Input tools may be in user or tool role, so process both.
			if len(m.ToolOutputs) > 0 {
				o := toolOutputsToOpenAIChat(m.ToolOutputs)
				out = append(out, o...)
			}
			// If not tool output and we still have content in this role, attach it.
			if content != "" {
				out = append(out, openai.UserMessage(content))
			}
		}
	}
	return out, nil
}

func toolOutputsToOpenAIChat(
	toolOutputs []toolSpec.ToolOutput,
) []openai.ChatCompletionMessageParamUnion {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(toolOutputs))

	// Tool outputs with a callID can be represented as proper tool messages.
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
		out = append(out, openai.ToolMessage(toolContent, o.ID))
	}

	// Any tool outputs without a callID are rendered as plain user text
	// so the model still sees the information.
	var orphanOutputs []toolSpec.ToolOutput
	for _, o := range toolOutputs {
		if strings.TrimSpace(o.ID) == "" {
			orphanOutputs = append(orphanOutputs, o)
		}
	}
	if len(orphanOutputs) > 0 {
		if text := renderToolOutputsAsText(orphanOutputs); text != "" {
			out = append(out, openai.UserMessage(text))
		}
	}
	return out
}

// contentBlocksToOpenAIChat converts generic content blocks into OpenAI
// ChatCompletion content parts (text/image/file) so that attachments can be
// sent to the multimodal chat endpoint.
func contentBlocksToOpenAIChat(
	blocks []attachment.ContentBlock,
) ([]openai.ChatCompletionContentPartUnionParam, error) {
	if len(blocks) == 0 {
		return nil, nil
	}

	out := make([]openai.ChatCompletionContentPartUnionParam, 0, len(blocks))
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
			out = append(out, openai.TextContentPart(txt))

		case attachment.ContentBlockImage:
			if bData == "" {
				continue
			}

			if mime == "" {
				mime = string(fileutil.DefaultImageMIME)
			}
			dataURL := fmt.Sprintf("data:%s;base64,%s", mime, bData)
			img := openai.ChatCompletionContentPartImageImageURLParam{
				URL:    dataURL,
				Detail: "auto",
			}
			out = append(out, openai.ImageContentPart(img))

		case attachment.ContentBlockFile:
			if bData == "" {
				continue
			}
			if mime == "" {
				mime = string(fileutil.MIMEApplicationOctetStream)
			}
			dataURL := fmt.Sprintf("data:%s;base64,%s", mime, bData)
			var fileParam openai.ChatCompletionContentPartFileFileParam
			fileParam.FileData = param.NewOpt(dataURL)
			if fname != "" {
				fileParam.Filename = param.NewOpt(fname)
			}
			out = append(out, openai.FileContentPart(fileParam))

		default:
			continue
		}
	}
	return out, nil
}

func getOpenAIMessageFromSystemPrompt(
	providerName, modelName, systemPrompt string,
) *openai.ChatCompletionMessageParamUnion {
	sp := strings.TrimSpace(systemPrompt)
	if sp == "" {
		return nil
	}
	msg := openai.SystemMessage(sp)
	// Convert a system message to a developer message for o series models.
	if providerName == "openai" &&
		(strings.HasPrefix(modelName, "o") || (strings.HasPrefix(modelName, "gpt-5"))) {
		// If the SDK exposes an enum for this, use it; otherwise the raw string works.
		msg = openai.DeveloperMessage(sp)
	}
	return &msg
}

func toOpenAIChatTools(
	tools []spec.FetchCompletionToolChoice,
) ([]openai.ChatCompletionToolUnionParam, map[string]spec.FetchCompletionToolChoice, error) {
	if len(tools) == 0 {
		return nil, nil, nil
	}
	ordered, nameMap := buildToolChoiceNameMapping(tools)
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(tools))

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
		fn := shared.FunctionDefinitionParam{
			Name:       tw.Name,
			Parameters: schema,
		}
		if desc := toolDescription(ct); desc != "" {
			fn.Description = openai.String(desc)
		}
		out = append(out, openai.ChatCompletionFunctionTool(fn))
	}
	return out, nameMap, nil
}

func extractOpenAIChatToolCalls(
	choices []openai.ChatCompletionChoice,
	toolChoiceNameMap map[string]spec.FetchCompletionToolChoice,
) []toolSpec.ToolCall {
	if len(choices) == 0 {
		return []toolSpec.ToolCall{}
	}
	msg := &choices[0].Message
	if len(msg.ToolCalls) == 0 {
		return []toolSpec.ToolCall{}
	}

	out := make([]toolSpec.ToolCall, 0)
	for _, tc := range msg.ToolCalls {
		switch tc.Type {
		case string(openaiSharedConstant.Function("").Default()):
			if tc.ID == "" || strings.TrimSpace(tc.Function.Name) == "" {
				continue
			}
			name := tc.Function.Name
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
					ID:         tc.ID,
					CallID:     tc.ID,
					Name:       tc.Function.Name,
					Arguments:  tc.Function.Arguments,
					Type:       tc.Type,
					ToolChoice: toolChoice,
				},
			)
		case string(openaiSharedConstant.Custom("").Default()):
			if tc.ID == "" || strings.TrimSpace(tc.Custom.Name) == "" {
				continue
			}
			name := tc.Custom.Name
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
					ID:         tc.ID,
					CallID:     tc.ID,
					Name:       tc.Custom.Name,
					Arguments:  tc.Custom.Input,
					Type:       tc.Type,
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

// usageFromOpenAIChatCompletion normalizes OpenAI ChatCompletion usage into spec.Usage.
func usageFromOpenAIChatCompletion(resp *openai.ChatCompletion) *modelpresetSpec.Usage {
	uOut := &modelpresetSpec.Usage{}
	if resp == nil {
		return uOut
	}

	u := resp.Usage

	uOut.InputTokensTotal = u.PromptTokens
	uOut.InputTokensCached = u.PromptTokensDetails.CachedTokens
	uOut.InputTokensUncached = max(u.PromptTokens-u.PromptTokensDetails.CachedTokens, 0)
	uOut.OutputTokens = u.CompletionTokens
	uOut.ReasoningTokens = u.CompletionTokensDetails.ReasoningTokens

	return uOut
}
