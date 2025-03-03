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

// MakeCompletion creates a new completion request.
func (api *BaseAIAPI) MakeCompletion(
	ctx context.Context,
	modelInfo spec.ModelInfo,
	prompt string,
	prevMessages []spec.ChatCompletionRequestMessage,
	inputParams map[string]any,
) (*spec.CompletionRequest, error) {
	if inputParams == nil {
		inputParams = make(map[string]any)
	}

	completionRequest := spec.CompletionRequest{
		Model:           string(modelInfo.Name),
		Temperature:     api.ProviderInfo.DefaultTemperature,
		Stream:          api.ProviderInfo.StreamingSupport,
		MaxPromptLength: modelInfo.MaxPromptLength,
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

	// Handle StreamingSupport. Streaming resolution is: ModelInfo > ProviderInfo
	if modelInfo.StreamingSupport != nil {
		completionRequest.Stream = *modelInfo.StreamingSupport
	}
	// Handle temperature. Temperature resolution is: ModelInfo > inputParams > ProviderInfo
	if modelInfo.DefaultTemperature != nil {
		completionRequest.Temperature = *modelInfo.DefaultTemperature
	} else if inTemp, ok := inputParams["temperature"].(float64); ok {
		completionRequest.Temperature = inTemp
	}

	if inMaxOutputLength, ok := inputParams["maxOutputLength"].(int); ok {
		if inMaxOutputLength <= modelInfo.MaxOutputLength {
			completionRequest.MaxOutputLength = &inMaxOutputLength
		}
	}

	if inMaxPromptLength, ok := inputParams["maxPromptLength"].(int); ok {
		if inMaxPromptLength <= modelInfo.MaxPromptLength {
			completionRequest.MaxPromptLength = inMaxPromptLength
		}
	}

	if inSystemPrompt, ok := inputParams["systemPrompt"].(string); ok {
		completionRequest.SystemPrompt = &inSystemPrompt
	}

	for key, value := range inputParams {
		switch key {
		case "systemPrompt",
			"maxPromptLength",
			"maxOutputLength",
			"temperature",
			"model",
			"provider":
			// Do nothing for these keys
		default:
			if completionRequest.AdditionalParameters == nil {
				completionRequest.AdditionalParameters = make(map[string]any)
			}
			completionRequest.AdditionalParameters[key] = value
		}
	}

	// Assuming filterMessagesByTokenCount is implemented elsewhere
	completionRequest.Messages = FilterMessagesByTokenCount(
		completionRequest.Messages,
		completionRequest.MaxPromptLength,
	)

	return &completionRequest, nil
}

// SetProviderAttribute sets the attributes for the OpenAIAPI.
func (api *BaseAIAPI) SetProviderAttribute(
	ctx context.Context,
	apiKey *string,
	defaultModel *string,
	defaultTemperature *float64,
	defaultOrigin *string,
) error {
	if apiKey == nil && defaultModel == nil && defaultTemperature == nil && defaultOrigin == nil {
		return errors.New("no attribute provided for set")
	}
	if api.ProviderInfo == nil {
		return errors.New("no ProviderInfo found")
	}
	if apiKey != nil {
		api.ProviderInfo.APIKey = *apiKey
	}
	if defaultOrigin != nil {
		api.ProviderInfo.DefaultOrigin = *defaultOrigin
	}
	if defaultModel != nil {
		api.ProviderInfo.DefaultModel = spec.ModelName(*defaultModel)
	}
	if defaultTemperature != nil {
		api.ProviderInfo.DefaultTemperature = *defaultTemperature
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
		llms.WithModel(input.Model),
		llms.WithTemperature(input.Temperature),
	}
	if input.MaxOutputLength != nil {
		options = append(options, llms.WithMaxTokens(*input.MaxOutputLength))
	}
	if input.Stream && onStreamData != nil {
		streamingFunc := func(ctx context.Context, chunk []byte) error {
			err := onStreamData(string(chunk))
			return err
		}
		options = append(options, llms.WithStreamingFunc(streamingFunc))
	}

	content := []llms.MessageContent{}
	if input.SystemPrompt != nil {
		content = append(content, llms.TextParts(llms.ChatMessageTypeSystem, *input.SystemPrompt))
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
