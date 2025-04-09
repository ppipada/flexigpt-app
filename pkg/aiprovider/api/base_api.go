package api

import (
	"context"
	"errors"

	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
	"github.com/tmc/langchaingo/llms"
)

type BaseAIAPI struct {
	ProviderInfo *spec.ProviderInfo
	Debug        bool
}

// NewOpenAIAPI creates a new instance of BaseAIAPI with input ProviderInfo.
func NewBaseAIAPI(p *spec.ProviderInfo, debug bool) *BaseAIAPI {
	return &BaseAIAPI{
		ProviderInfo: p,
		Debug:        debug,
	}
}

// IsConfigured checks if the API is configured.
func (api *BaseAIAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderInfo.IsConfigured()
}

func (api *BaseAIAPI) GetProviderInfo(ctx context.Context) *spec.ProviderInfo {
	return api.ProviderInfo
}

// SetProviderAPIKey sets the key for a provider.
func (api *BaseAIAPI) SetProviderAPIKey(
	ctx context.Context,
	apiKey string,
) error {
	if apiKey == "" {
		return errors.New("invalid apikey provided")
	}
	if api.ProviderInfo == nil {
		return errors.New("no ProviderInfo found")
	}

	api.ProviderInfo.APIKey = apiKey

	return nil
}

// SetProviderAttribute sets the attributes of a provider.
func (api *BaseAIAPI) SetProviderAttribute(
	ctx context.Context,
	origin *string,
	chatCompletionPathPrefix *string,
) error {
	if origin == nil && chatCompletionPathPrefix == nil {
		return errors.New("no attribute provided for set")
	}
	if api.ProviderInfo == nil {
		return errors.New("no ProviderInfo found")
	}
	if origin != nil && *origin != "" {
		api.ProviderInfo.Origin = *origin
	}
	if chatCompletionPathPrefix != nil {
		api.ProviderInfo.ChatCompletionPathPrefix = *chatCompletionPathPrefix
	}

	return nil
}

func (api *BaseAIAPI) getCompletionRequest(
	prompt string,
	modelParams spec.ModelParams,
	inbuiltModelParams *spec.ModelParams,
	prevMessages []spec.ChatCompletionRequestMessage,
) *spec.CompletionRequest {
	completionRequest := spec.CompletionRequest{
		ModelParams: spec.ModelParams{
			Name:                 modelParams.Name,
			AdditionalParameters: modelParams.AdditionalParameters,
			Temperature:          modelParams.Temperature,
		},
	}

	// Cannot turn on streaming of it is set false in model
	if inbuiltModelParams != nil && modelParams.Stream {
		completionRequest.ModelParams.Stream = inbuiltModelParams.Stream
	} else {
		completionRequest.ModelParams.Stream = modelParams.Stream
	}

	if inbuiltModelParams != nil &&
		modelParams.MaxOutputLength > inbuiltModelParams.MaxOutputLength {
		completionRequest.ModelParams.MaxOutputLength = inbuiltModelParams.MaxOutputLength
	} else {
		completionRequest.ModelParams.MaxOutputLength = modelParams.MaxOutputLength
	}

	if inbuiltModelParams != nil &&
		modelParams.MaxPromptLength > inbuiltModelParams.MaxPromptLength {
		completionRequest.ModelParams.MaxPromptLength = inbuiltModelParams.MaxPromptLength
	} else {
		completionRequest.ModelParams.MaxPromptLength = modelParams.MaxPromptLength
	}

	reqSystemPrompt := modelParams.SystemPrompt
	if inbuiltModelParams != nil && inbuiltModelParams.SystemPrompt != "" {
		reqSystemPrompt = inbuiltModelParams.SystemPrompt + "\n" + reqSystemPrompt
	}
	completionRequest.ModelParams.SystemPrompt = reqSystemPrompt

	// Handle messages
	messages := append([]spec.ChatCompletionRequestMessage{}, prevMessages...)
	if prompt != "" {
		message := spec.ChatCompletionRequestMessage{
			Role:    "user",
			Content: &prompt,
		}
		messages = append(messages, message)
	}
	completionRequest.Messages = messages

	// Assuming filterMessagesByTokenCount is implemented elsewhere
	completionRequest.Messages = FilterMessagesByTokenCount(
		completionRequest.Messages,
		completionRequest.ModelParams.MaxPromptLength,
	)

	return &completionRequest
}

// FetchCompletion processes the completion request.
func (api *BaseAIAPI) FetchCompletion(
	ctx context.Context,
	llm llms.Model,
	prompt string,
	modelParams spec.ModelParams,
	inbuiltModelParams *spec.ModelParams,
	prevMessages []spec.ChatCompletionRequestMessage,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	input := api.getCompletionRequest(prompt, modelParams, inbuiltModelParams, prevMessages)
	if len(input.Messages) == 0 {
		return nil, errors.New("empty input messages")
	}
	if llm == nil {
		return nil, errors.New("llm not initialized")
	}
	options := []llms.CallOption{
		llms.WithModel(string(input.ModelParams.Name)),
	}

	if input.ModelParams.Temperature != nil {
		options = append(options, llms.WithTemperature(*input.ModelParams.Temperature))
	}

	options = append(options, llms.WithMaxTokens(input.ModelParams.MaxOutputLength))

	count := 0
	if input.ModelParams.Stream && onStreamData != nil {
		streamingFunc := func(ctx context.Context, chunk []byte) error {
			if count < 5 {
				// slog.Info("stream", "got chunk", string(chunk))
				count += 1
			}

			err := onStreamData(string(chunk))
			return err
		}
		options = append(options, llms.WithStreamingFunc(streamingFunc))
	}

	content := []llms.MessageContent{}
	if input.ModelParams.SystemPrompt != "" {
		content = append(
			content,
			llms.TextParts(llms.ChatMessageTypeSystem, input.ModelParams.SystemPrompt),
		)
	}
	for _, msg := range input.Messages {
		role := LangchainRoleMap[msg.Role]
		if msg.Content != nil {
			content = append(content, llms.TextParts(role, *msg.Content))
		}
	}
	if len(content) == 0 {
		return nil, errors.New("empty input content messages")
	}

	completionResp := &spec.CompletionResponse{}

	ctx = AddDebugResponseToCtx(ctx)
	resp, err := llm.GenerateContent(ctx, content, options...)
	debugResp, ok := GetDebugHTTPResponse(ctx)
	if err != nil {
		if ok && debugResp != nil && debugResp.ErrorDetails != nil {
			completionResp.ErrorDetails = debugResp.ErrorDetails
			return completionResp, nil
		}
		return nil, err
	}
	if ok && debugResp != nil {
		completionResp.RequestDetails = debugResp.RequestDetails
		completionResp.ErrorDetails = debugResp.ErrorDetails
		completionResp.ResponseDetails = debugResp.ResponseDetails
	}

	if resp == nil || len(resp.Choices) == 0 {
		if ok && debugResp != nil {
			if completionResp.ErrorDetails == nil {
				completionResp.ErrorDetails = &spec.APIErrorDetails{
					Message: "Got nil response from LLM api",
				}
			} else {
				completionResp.ErrorDetails.Message += "Got nil response from LLM api"
			}
			return completionResp, nil
		}
		return nil, errors.New("got nil response from LLM api")
	}
	completionResp.RespContent = &resp.Choices[0].Content
	return completionResp, nil
}
