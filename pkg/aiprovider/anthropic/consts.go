package anthropic

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const ProviderNameAnthropic spec.ProviderName = "anthropic"

const (
	Claude37Sonnet spec.ModelName = "claude-3-7-sonnet-20250219"
	Claude35Sonnet spec.ModelName = "claude-3-5-sonnet-20241022"
	Claude35Haiku  spec.ModelName = "claude-3-5-haiku-20241022"
	Claude3Opus    spec.ModelName = "claude-3-opus-20240229"
	Claude3Sonnet  spec.ModelName = "claude-3-sonnet-20240229"
	Claude3Haiku   spec.ModelName = "claude-3-haiku-20240307"
)

var AnthropicModels = map[spec.ModelName]spec.ModelInfo{
	Claude37Sonnet: {
		Name:                Claude37Sonnet,
		DisplayName:         "Claude 3.7 Sonnet",
		Provider:            ProviderNameAnthropic,
		MaxPromptLength:     8192,
		MaxOutputLength:     8192,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    true,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
	Claude35Sonnet: {
		Name:                Claude35Sonnet,
		DisplayName:         "Claude 3.5 Sonnet",
		Provider:            ProviderNameAnthropic,
		MaxPromptLength:     8192,
		MaxOutputLength:     8192,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    false,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
	Claude35Haiku: {
		Name:                Claude35Haiku,
		DisplayName:         "Claude 3.5 Haiku",
		Provider:            ProviderNameAnthropic,
		MaxPromptLength:     8192,
		MaxOutputLength:     8192,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    false,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
	Claude3Opus: {
		Name:                Claude3Opus,
		DisplayName:         "Claude 3 Opus",
		Provider:            ProviderNameAnthropic,
		MaxPromptLength:     8192,
		MaxOutputLength:     8192,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    false,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
	Claude3Sonnet: {
		Name:                Claude3Sonnet,
		DisplayName:         "Claude 3 Sonnet",
		Provider:            ProviderNameAnthropic,
		MaxPromptLength:     8192,
		MaxOutputLength:     8192,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    false,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
	Claude3Haiku: {
		Name:                Claude3Haiku,
		DisplayName:         "Claude 3 Haiku",
		Provider:            ProviderNameAnthropic,
		MaxPromptLength:     8192,
		MaxOutputLength:     8192,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    false,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
}

var AnthropicProviderInfo = spec.ProviderInfo{
	Name:         ProviderNameAnthropic,
	APIKey:       "",
	Engine:       "",
	Origin:       "https://api.anthropic.com/v1",
	DefaultModel: Claude35Sonnet,

	APIKeyHeaderKey: "x-api-key",
	DefaultHeaders: map[string]string{
		"content-type":      "application/json",
		"accept":            "application/json",
		"anthropic-version": "2023-06-01",
	},
	ChatCompletionPathPrefix: "/messages",
	Models:                   AnthropicModels,
}
