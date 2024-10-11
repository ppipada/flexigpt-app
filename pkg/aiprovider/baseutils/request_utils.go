package baseutils

import "github.com/flexigpt/flexiui/pkg/aiprovider/spec"

// GetCompletionRequest creates a completion request with given parameters.
func GetCompletionRequest(
	prompt string,
	prevMessages []spec.ChatCompletionRequestMessage,
	inputParams map[string]interface{},
	stream bool,
	providerInfo *spec.ProviderInfo,
) spec.CompletionRequest {
	defaultModel := providerInfo.DefaultModel
	defaultTemperature := providerInfo.DefaultTemperature
	maxPromptLength := spec.AllModelInfo[defaultModel].MaxPromptLength

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

	return completionRequest
}
