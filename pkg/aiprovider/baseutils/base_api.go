package baseutils

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

func (api *BaseAIAPI) CreateCompletionRequest(
	ctx context.Context,
	prompt string,
	modelParams spec.ModelParams,
	prevMessages []spec.ChatCompletionRequestMessage,
) (*spec.CompletionRequest, error) {
	modelInfo, ok := api.ProviderInfo.Models[modelParams.Name]
	if !ok {
		return nil, errors.New(
			"Provider: " + string(
				api.ProviderInfo.Name,
			) + " ModelInfo not found for Model: " + string(
				modelParams.Name,
			),
		)
	}
	completionRequest := spec.CompletionRequest{
		ModelParams: spec.ModelParams{
			Name:                 modelParams.Name,
			AdditionalParameters: modelParams.AdditionalParameters,
		},
	}

	// Cannot turn on streaming of it is set false in model
	if modelParams.Stream != nil && *modelParams.Stream {
		completionRequest.ModelParams.Stream = &modelInfo.StreamingSupport
	} else {
		completionRequest.ModelParams.Stream = BoolPtr(modelInfo.StreamingSupport)
	}

	if modelParams.Temperature != nil {
		completionRequest.ModelParams.Temperature = modelParams.Temperature
	} else {
		completionRequest.ModelParams.Temperature = &modelInfo.DefaultTemperature
	}

	if modelParams.OutputLength != nil && *modelParams.OutputLength <= modelInfo.MaxOutputLength {
		completionRequest.ModelParams.OutputLength = modelParams.OutputLength
	} else {
		completionRequest.ModelParams.OutputLength = &modelInfo.MaxOutputLength
	}

	if modelParams.PromptLength != nil && *modelParams.PromptLength <= modelInfo.MaxPromptLength {
		completionRequest.ModelParams.PromptLength = modelParams.PromptLength
	} else {
		completionRequest.ModelParams.PromptLength = &modelInfo.MaxPromptLength
	}

	reqSystemPrompt := modelInfo.DefaultSystemPrompt
	if modelParams.SystemPrompt != nil {
		reqSystemPrompt += "\n" + *modelParams.SystemPrompt
	}
	if reqSystemPrompt != "" {
		completionRequest.ModelParams.SystemPrompt = &reqSystemPrompt
	}

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
		*completionRequest.ModelParams.PromptLength,
	)

	return &completionRequest, nil
}

// SetProviderAttribute sets the attributes for the OpenAIAPI.
func (api *BaseAIAPI) SetProviderAttribute(
	ctx context.Context,
	apiKey *string,
	defaultModel *string,
	origin *string,
) error {
	if apiKey == nil && defaultModel == nil && origin == nil {
		return errors.New("no attribute provided for set")
	}
	if api.ProviderInfo == nil {
		return errors.New("no ProviderInfo found")
	}
	if apiKey != nil {
		api.ProviderInfo.APIKey = *apiKey
	}
	if origin != nil {
		api.ProviderInfo.Origin = *origin
	}
	if defaultModel != nil {
		api.ProviderInfo.DefaultModel = spec.ModelName(*defaultModel)
	}

	return nil
}

// FetchCompletion processes the completion request.
func (api *BaseAIAPI) FetchCompletion(
	ctx context.Context,
	llm llms.Model,
	input spec.CompletionRequest,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
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

	if input.ModelParams.OutputLength != nil {
		options = append(options, llms.WithMaxTokens(*input.ModelParams.OutputLength))
	}

	count := 0
	if input.ModelParams.Stream != nil && *input.ModelParams.Stream && onStreamData != nil {
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
	if input.ModelParams.SystemPrompt != nil && *input.ModelParams.SystemPrompt != "" {
		content = append(
			content,
			llms.TextParts(llms.ChatMessageTypeSystem, *input.ModelParams.SystemPrompt),
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
