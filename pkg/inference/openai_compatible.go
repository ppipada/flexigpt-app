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

	"github.com/ppipada/flexigpt-app/pkg/builtin"
	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

// OpenAICompatibleAPI struct that implements the CompletionProvider interface.
type OpenAICompatibleAPI struct {
	ProviderParams *spec.ProviderParams
	Debug          bool
	client         *openai.Client
}

// NewOpenAICompatibleAPI creates a new instance of OpenAICompatibleProvider with the provided ProviderParams.
func NewOpenAICompatibleAPI(pi spec.ProviderParams, debug bool) (*OpenAICompatibleAPI, error) {
	if pi.Name == "" || pi.Origin == "" {
		return nil, errors.New("openai compatible LLM: invalid args")
	}
	return &OpenAICompatibleAPI{
		ProviderParams: &pi,
		Debug:          debug,
	}, nil
}

func (api *OpenAICompatibleAPI) InitLLM(ctx context.Context) error {
	if !api.IsConfigured(ctx) {
		slog.Debug(
			string(
				api.ProviderParams.Name,
			) + ": No API key given. Not initializing OpenAICompatibleAPI LLM object",
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
		"openai compatible LLM provider initialized",
		"name",
		string(api.ProviderParams.Name),
		"URL",
		providerURL,
	)
	return nil
}

func (api *OpenAICompatibleAPI) DeInitLLM(ctx context.Context) error {
	api.client = nil
	slog.Info(
		"openai compatible LLM: provider de initialized",
		"name",
		string(api.ProviderParams.Name),
	)
	return nil
}

func (api *OpenAICompatibleAPI) GetProviderInfo(ctx context.Context) *spec.ProviderParams {
	return api.ProviderParams
}

func (api *OpenAICompatibleAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderParams.APIKey != ""
}

// SetProviderAPIKey sets the key for a provider.
func (api *OpenAICompatibleAPI) SetProviderAPIKey(
	ctx context.Context,
	apiKey string,
) error {
	if apiKey == "" {
		return errors.New("openai compatible LLM: invalid apikey provided")
	}
	if api.ProviderParams == nil {
		return errors.New("openai compatible LLM: no ProviderParams found")
	}

	api.ProviderParams.APIKey = apiKey

	return nil
}

func (api *OpenAICompatibleAPI) FetchCompletion(
	ctx context.Context,
	prompt string,
	modelParams spec.ModelParams,
	prevMessages []spec.ChatCompletionRequestMessage,
	onStreamTextData, onStreamThinkingData func(string) error,
) (*spec.CompletionResponse, error) {
	if api.client == nil {
		return nil, errors.New("openai compatible LLM: client not initialized")
	}

	input := getCompletionRequest(prompt, modelParams, prevMessages)
	if len(input.Messages) == 0 && strings.TrimSpace(input.ModelParams.SystemPrompt) == "" {
		return nil, errors.New("openai compatible LLM: empty input messages")
	}

	// Build OpenAI chat messages.
	msgs, err := toOpenAIChatMessages(
		input.ModelParams.SystemPrompt,
		input.Messages,
		string(input.ModelParams.Name),
		api.ProviderParams.Name,
	)
	if err != nil {
		return nil, err
	}

	params := openai.ChatCompletionNewParams{
		Model:               shared.ChatModel(input.ModelParams.Name),
		MaxCompletionTokens: openai.Int(int64(input.ModelParams.MaxOutputLength)),
		Messages:            msgs,
	}
	if input.ModelParams.Temperature != nil {
		params.Temperature = openai.Float(*input.ModelParams.Temperature)
	}

	if rp := input.ModelParams.Reasoning; rp != nil &&
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
	if input.ModelParams.Timeout > 0 {
		timeout = time.Duration(input.ModelParams.Timeout) * time.Second
	}

	if input.ModelParams.Stream && onStreamTextData != nil && onStreamThinkingData != nil {
		return api.doStreaming(ctx, params, onStreamTextData, onStreamThinkingData, timeout)
	}
	return api.doNonStreaming(ctx, params, timeout)
}

func (api *OpenAICompatibleAPI) doNonStreaming(
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
	completionResp.RespContent = &full
	return completionResp, nil
}

func (api *OpenAICompatibleAPI) doStreaming(
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
	completionResp.RespContent = &fullResp
	return completionResp, streamErr
}

func toOpenAIChatMessages(
	systemPrompt string,
	messages []spec.ChatCompletionRequestMessage,
	modelName string,
	providerName modelpresetSpec.ProviderName,
) ([]openai.ChatCompletionMessageParamUnion, error) {
	var out []openai.ChatCompletionMessageParamUnion

	// System/developer prompt.
	if sp := strings.TrimSpace(systemPrompt); sp != "" {
		msg := openai.SystemMessage(sp)
		if providerName == builtin.ProviderNameOpenAI &&
			(strings.HasPrefix(modelName, "o") || (strings.HasPrefix(modelName, "gpt-5"))) {
			// If the SDK exposes an enum for this, use it; otherwise the raw string works.
			msg = openai.DeveloperMessage(sp)
		}
		out = append(out, msg)
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
