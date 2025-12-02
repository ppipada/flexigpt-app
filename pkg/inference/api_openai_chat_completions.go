package inference

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/packages/param"
	"github.com/openai/openai-go/v2/shared"

	"github.com/ppipada/flexigpt-app/pkg/attachment"
	"github.com/ppipada/flexigpt-app/pkg/fileutil"
	"github.com/ppipada/flexigpt-app/pkg/inference/debugclient"
	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

// OpenAIChatCompletionsAPI struct that implements the CompletionProvider interface.
type OpenAIChatCompletionsAPI struct {
	ProviderParams *spec.ProviderParams
	Debug          bool
	client         *openai.Client
}

func NewOpenAIChatCompletionsAPI(
	pi spec.ProviderParams,
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

func (api *OpenAIChatCompletionsAPI) GetProviderInfo(ctx context.Context) *spec.ProviderParams {
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
	modelParams spec.ModelParams,
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
		rp.Type == modelpresetSpec.ReasoningTypeSingleWithLevels {
		switch rp.Level {
		case
			modelpresetSpec.ReasoningLevelNone,
			modelpresetSpec.ReasoningLevelMinimal,
			modelpresetSpec.ReasoningLevelLow,
			modelpresetSpec.ReasoningLevelMedium,
			modelpresetSpec.ReasoningLevelHigh:
			params.ReasoningEffort = shared.ReasoningEffort(string(rp.Level))
		default:
			return nil, fmt.Errorf("invalid level %q for singleWithLevels", rp.Level)

		}
	}
	if len(completionData.ToolChoices) > 0 {
		toolDefs, err := toOpenAIChatTools(completionData.ToolChoices)
		if err != nil {
			return nil, err
		}
		if len(toolDefs) > 0 {
			params.Tools = toolDefs
		}
	}
	timeout := modelpresetSpec.DefaultAPITimeout
	if completionData.ModelParams.Timeout > 0 {
		timeout = time.Duration(completionData.ModelParams.Timeout) * time.Second
	}

	if completionData.ModelParams.Stream && onStreamTextData != nil && onStreamThinkingData != nil {
		return api.doStreaming(ctx, params, onStreamTextData, onStreamThinkingData, timeout)
	}
	return api.doNonStreaming(ctx, params, timeout)
}

func (api *OpenAIChatCompletionsAPI) doNonStreaming(
	ctx context.Context,
	params openai.ChatCompletionNewParams,
	timeout time.Duration,
) (*spec.FetchCompletionResponse, error) {
	completionResp := &spec.FetchCompletionResponse{Body: &spec.FetchCompletionResponseBody{}}
	ctx = debugclient.AddDebugResponseToCtx(ctx)
	resp, err := api.client.Chat.Completions.New(ctx, params, option.WithRequestTimeout(timeout))
	isNilResp := resp == nil || len(resp.Choices) == 0
	attachDebugResp(ctx, completionResp.Body, err, isNilResp, "", resp)
	if isNilResp {
		return completionResp, nil
	}
	full := resp.Choices[0].Message.Content
	completionResp.Body.ResponseContent = []spec.ResponseContent{
		{Type: spec.ResponseContentTypeText, Content: full},
	}
	if toolCalls := extractOpenAIChatToolCalls(resp.Choices); len(toolCalls) > 0 {
		completionResp.Body.ToolCalls = toolCalls
	}
	return completionResp, nil
}

func (api *OpenAIChatCompletionsAPI) doStreaming(
	ctx context.Context,
	params openai.ChatCompletionNewParams,
	onStreamTextData, onStreamThinkingData func(string) error,
	timeout time.Duration,
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

	if isNilResp {
		return completionResp, nil
	}

	fullResp := acc.Choices[0].Message.Content
	completionResp.Body.ResponseContent = []spec.ResponseContent{
		{Type: spec.ResponseContentTypeText, Content: fullResp},
	}
	if toolCalls := extractOpenAIChatToolCalls(acc.Choices); len(toolCalls) > 0 {
		completionResp.Body.ToolCalls = toolCalls
	}
	return completionResp, streamErr
}

func toOpenAIChatMessages(
	ctx context.Context,
	systemPrompt string,
	messages []spec.ChatCompletionDataMessage,
	modelName modelpresetSpec.ModelName,
	providerName modelpresetSpec.ProviderName,
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
		if m.Role == spec.User {
			lastUserIdx = i
		}
	}

	for idx, m := range messages {
		if m.Content == nil && len(m.Attachments) == 0 {
			continue
		}
		switch m.Role {
		case spec.System:
			out = append(out, openai.SystemMessage(*m.Content))
		case spec.Developer:
			out = append(out, openai.DeveloperMessage(*m.Content))
		case spec.User:
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
				blocks, err := attachment.BuildContentBlocks(ctx, m.Attachments, overrideOriginalAttachment)
				if err != nil {
					return nil, err
				}
				attachmentContent, err = contentBlocksToOpenAIChat(blocks)
				if err != nil {
					return nil, err
				}
			}

			if len(attachmentContent) == 0 {
				text := ""
				if m.Content != nil {
					text = *m.Content
				}
				if strings.TrimSpace(text) == "" {
					// Nothing to send for this turn.
					continue
				}
				out = append(out, openai.UserMessage(*m.Content))
				continue
			}

			var parts []openai.ChatCompletionContentPartUnionParam
			if c := strings.TrimSpace(*m.Content); c != "" {
				parts = append(parts, openai.TextContentPart(c))
			}
			parts = append(parts, attachmentContent...)
			out = append(out, openai.UserMessage(parts))

		case spec.Assistant:
			out = append(out, openai.AssistantMessage(*m.Content))
		case spec.Function, spec.Tool:
			toolCallID := "1"
			out = append(out, openai.ToolMessage(*m.Content, toolCallID))
		}
	}
	return out, nil
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
		switch b.Kind {
		case attachment.ContentBlockText:
			if strings.TrimSpace(b.Text) == "" {
				continue
			}
			out = append(out, openai.TextContentPart(b.Text))

		case attachment.ContentBlockImage:
			if b.Base64Data == "" {
				continue
			}
			mime := b.MIMEType
			if mime == "" {
				mime = string(fileutil.DefaultImageMIME)
			}
			dataURL := fmt.Sprintf("data:%s;base64,%s", mime, b.Base64Data)
			img := openai.ChatCompletionContentPartImageImageURLParam{
				URL:    dataURL,
				Detail: "auto",
			}
			out = append(out, openai.ImageContentPart(img))

		case attachment.ContentBlockFile:
			if b.Base64Data == "" {
				continue
			}
			mime := b.MIMEType
			if mime == "" {
				mime = "application/octet-stream"
			}
			dataURL := fmt.Sprintf("data:%s;base64,%s", mime, b.Base64Data)
			var fileParam openai.ChatCompletionContentPartFileFileParam
			fileParam.FileData = param.NewOpt(dataURL)
			if b.FileName != "" {
				fileParam.Filename = param.NewOpt(b.FileName)
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
) ([]openai.ChatCompletionToolUnionParam, error) {
	if len(tools) == 0 {
		return nil, nil
	}
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(tools))
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
		fn := shared.FunctionDefinitionParam{
			Name:       toolFunctionName(ct),
			Parameters: schema,
		}
		if desc := toolDescription(ct); desc != "" {
			fn.Description = openai.String(desc)
		}
		out = append(out, openai.ChatCompletionFunctionTool(fn))
	}
	return out, nil
}

func extractOpenAIChatToolCalls(choices []openai.ChatCompletionChoice) []spec.ResponseToolCall {
	if len(choices) == 0 {
		return nil
	}
	return convertOpenAIChatMessageToolCalls(&choices[0].Message)
}

func convertOpenAIChatMessageToolCalls(
	msg *openai.ChatCompletionMessage,
) []spec.ResponseToolCall {
	if msg == nil || len(msg.ToolCalls) == 0 {
		return nil
	}

	out := make([]spec.ResponseToolCall, 0, len(msg.ToolCalls))
	for _, tc := range msg.ToolCalls {
		switch variant := tc.AsAny().(type) {
		case openai.ChatCompletionMessageFunctionToolCall:
			out = append(
				out,
				spec.ResponseToolCall{
					ID:        variant.ID,
					CallID:    variant.ID,
					Name:      variant.Function.Name,
					Arguments: variant.Function.Arguments,
					Type:      string(variant.Type),
				},
			)
		case openai.ChatCompletionMessageCustomToolCall:
			out = append(
				out,
				spec.ResponseToolCall{
					ID:        variant.ID,
					CallID:    variant.ID,
					Name:      variant.Custom.Name,
					Arguments: variant.Custom.Input,
					Type:      string(variant.Type),
				},
			)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
