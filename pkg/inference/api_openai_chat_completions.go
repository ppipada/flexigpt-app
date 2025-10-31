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
	"github.com/openai/openai-go/v2/shared"

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

	newClient := NewDebugHTTPClient(api.Debug, false)
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
) (*spec.CompletionData, error) {
	return getCompletionData(modelParams, currentMessage, prevMessages), nil
}

func (api *OpenAIChatCompletionsAPI) FetchCompletion(
	ctx context.Context,
	completionData *spec.CompletionData,
	onStreamTextData, onStreamThinkingData func(string) error,
) (*spec.CompletionResponse, error) {
	if api.client == nil {
		return nil, errors.New("openai chat completions api LLM: client not initialized")
	}
	if completionData == nil || len(completionData.Messages) == 0 {
		return nil, errors.New("openai chat completions api LLM: empty completion data")
	}

	// Build OpenAI chat messages.
	msgs, err := toOpenAIChatMessages(
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
		case modelpresetSpec.ReasoningLevelLow,
			modelpresetSpec.ReasoningLevelMedium,
			modelpresetSpec.ReasoningLevelHigh,
			modelpresetSpec.ReasoningLevelMinimal:
			params.ReasoningEffort = shared.ReasoningEffort(string(rp.Level))
		default:
			return nil, fmt.Errorf("invalid level %q for singleWithLevels", rp.Level)

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
) (*spec.CompletionResponse, error) {
	completionResp := &spec.CompletionResponse{}
	ctx = AddDebugResponseToCtx(ctx)
	resp, err := api.client.Chat.Completions.New(ctx, params, option.WithRequestTimeout(timeout))
	isNilResp := resp == nil || len(resp.Choices) == 0
	attachDebugResp(ctx, completionResp, err, isNilResp)
	if isNilResp {
		return completionResp, nil
	}
	full := resp.Choices[0].Message.Content
	completionResp.ResponseContent = []spec.ResponseContent{
		{Type: spec.ResponseContentTypeText, Content: full},
	}
	return completionResp, nil
}

func (api *OpenAIChatCompletionsAPI) doStreaming(
	ctx context.Context,
	params openai.ChatCompletionNewParams,
	onStreamTextData, onStreamThinkingData func(string) error,
	timeout time.Duration,
) (*spec.CompletionResponse, error) {
	// No thinking data available in openai chat completions API, hence no thinking writer.
	write, flush := NewBufferedStreamer(onStreamTextData, FlushInterval, FlushChunkSize)

	completionResp := &spec.CompletionResponse{}
	ctx = AddDebugResponseToCtx(ctx)
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
	attachDebugResp(ctx, completionResp, streamErr, isNilResp)
	if isNilResp {
		return completionResp, nil
	}

	fullResp := acc.Choices[0].Message.Content
	completionResp.ResponseContent = []spec.ResponseContent{
		{Type: spec.ResponseContentTypeText, Content: fullResp},
	}
	return completionResp, streamErr
}

func toOpenAIChatMessages(
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

	for _, m := range messages {
		if m.Content == nil {
			continue
		}
		switch m.Role {
		case spec.System:
			out = append(out, openai.SystemMessage(*m.Content))
		case spec.Developer:
			out = append(out, openai.DeveloperMessage(*m.Content))
		case spec.User:
			out = append(out, openai.UserMessage(*m.Content))
		case spec.Assistant:
			out = append(out, openai.AssistantMessage(*m.Content))
		case spec.Function, spec.Tool:
			toolCallID := "1"
			out = append(out, openai.ToolMessage(*m.Content, toolCallID))
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
