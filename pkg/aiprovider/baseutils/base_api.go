package baseutils

import (
	"context"
	"fmt"

	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
	"github.com/tmc/langchaingo/llms"
)

type BaseAIAPI struct {
	ProviderInfo *spec.ProviderInfo
	Debug        bool
}

// NewOpenAIAPI creates a new instance of BaseAIAPI with input ProviderInfo
func NewBaseAIAPI(p *spec.ProviderInfo, debug bool) *BaseAIAPI {
	return &BaseAIAPI{
		ProviderInfo: p,
		Debug:        debug,
	}
}

// IsConfigured checks if the API is configured
func (api *BaseAIAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderInfo.IsConfigured()
}

// GetProviderInfo returns the provider information
func (api *BaseAIAPI) GetProviderInfo(ctx context.Context) (*spec.ProviderInfo, error) {
	if api.ProviderInfo == nil {
		return nil, fmt.Errorf("provider info is not set")
	}
	return api.ProviderInfo, nil
}

// GetCompletionRequest creates a new completion request
func (api *BaseAIAPI) GetCompletionRequest(
	ctx context.Context,
	modelInfo spec.ModelInfo,
	prompt string,
	prevMessages []spec.ChatCompletionRequestMessage,
	inputParams map[string]interface{},
	stream bool,
) (*spec.CompletionRequest, error) {
	defaultModel := modelInfo.Name
	defaultTemperature := api.ProviderInfo.DefaultTemperature
	maxPromptLength := modelInfo.MaxPromptLength

	if inputParams == nil {
		inputParams = make(map[string]interface{})
	}

	messages := append([]spec.ChatCompletionRequestMessage{}, prevMessages...)
	if prompt != "" {
		message := spec.ChatCompletionRequestMessage{
			Role:    "user",
			Content: &prompt,
		}
		messages = append(messages, message)
	}

	completionRequest := spec.CompletionRequest{
		Model:           string(defaultModel),
		Messages:        messages,
		Temperature:     defaultTemperature,
		Stream:          stream,
		MaxPromptLength: maxPromptLength,
	}

	for key, value := range inputParams {
		switch key {
		case "model":
			if model, ok := value.(string); ok {
				completionRequest.Model = model
			}
		case "maxOutputLength":
			if maxOutputLength, ok := value.(int); ok {
				completionRequest.MaxOutputLength = &maxOutputLength
			}
		case "temperature":
			if temperature, ok := value.(float64); ok {
				completionRequest.Temperature = temperature
			}
		case "maxPromptLength":
			if maxPromptLength, ok := value.(int); ok {
				completionRequest.MaxPromptLength = maxPromptLength
			}
		case "systemPrompt":
			if systemPrompt, ok := value.(string); ok {
				completionRequest.SystemPrompt = &systemPrompt
			}
		default:
			if key != "provider" {
				if completionRequest.AdditionalParameters == nil {
					completionRequest.AdditionalParameters = make(map[string]interface{})
				}
				completionRequest.AdditionalParameters[key] = value
			}
		}
	}

	// Assuming filterMessagesByTokenCount is implemented elsewhere
	completionRequest.Messages = FilterMessagesByTokenCount(
		completionRequest.Messages,
		completionRequest.MaxPromptLength,
	)

	return &completionRequest, nil

}

// SetProviderAttribute sets the attributes for the OpenAIAPI
func (api *BaseAIAPI) SetProviderAttribute(
	ctx context.Context,
	apiKey *string,
	defaultModel *string,
	defaultTemperature *float64,
	defaultOrigin *string,
) error {
	if apiKey == nil && defaultModel == nil && defaultTemperature == nil && defaultOrigin == nil {
		return fmt.Errorf("No attribute provided for set")
	}
	if api.ProviderInfo == nil {
		return fmt.Errorf("No ProviderInfo found")
	}
	if apiKey != nil {
		api.ProviderInfo.ApiKey = *apiKey
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

// FetchCompletion processes the completion request
func (api *BaseAIAPI) FetchCompletion(
	ctx context.Context,
	llm llms.Model,
	input spec.CompletionRequest,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	if len(input.Messages) == 0 {
		return nil, fmt.Errorf("Empty input messages")
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
		return nil, fmt.Errorf("Empty input content messages")
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
		completionResp.ResponseDetails = debugResp.ResponseDetails
		completionResp.ErrorDetails = debugResp.ErrorDetails
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
		return nil, fmt.Errorf("Got nil response from LLM api")
	}
	completionResp.RespContent = &resp.Choices[0].Content
	return completionResp, nil
}
