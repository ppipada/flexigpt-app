package consts

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const ProviderNameAnthropic spec.ProviderName = "anthropic"

const (
	Claude37Sonnet spec.ModelName = "claude-3.7-sonnet-20250219"
	Claude35Sonnet spec.ModelName = "claude-3.5-sonnet-20241022"
	Claude35Haiku  spec.ModelName = "claude-3.5-haiku-20241022"
	Claude3Opus    spec.ModelName = "claude-3-opus-20240229"
	Claude3Sonnet  spec.ModelName = "claude-3-sonnet-20240229"
	Claude3Haiku   spec.ModelName = "claude-3-haiku-20240307"
)

const (
	DisplayNameClaude37Sonnet = "Claude 3.7 Sonnet"
	DisplayNameClaude35Sonnet = "Claude 3.5 Sonnet"
	DisplayNameClaude35Haiku  = "Claude 3.5 Haiku"
	DisplayNameClaude3Opus    = "Claude 3 Opus"
	DisplayNameClaude3Sonnet  = "Claude 3 Sonnet"
	DisplayNameClaude3Haiku   = "Claude 3 Haiku"
)

var AnthropicModels = map[spec.ModelName]spec.ModelParams{
	Claude37Sonnet: {
		Name:            Claude37Sonnet,
		MaxPromptLength: 16384,
		MaxOutputLength: 8192,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		Reasoning: &spec.ReasoningParams{
			Type:   spec.ReasoningTypeHybridWithTokens,
			Tokens: 1024,
		},
		SystemPrompt: "",
		Timeout:      120,
	},
	Claude35Sonnet: {
		Name:            Claude35Sonnet,
		MaxPromptLength: 16384,
		MaxOutputLength: 8192,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Claude35Haiku: {
		Name:            Claude35Haiku,
		MaxPromptLength: 16384,
		MaxOutputLength: 8192,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Claude3Opus: {
		Name:            Claude3Opus,
		MaxPromptLength: 16384,
		MaxOutputLength: 4096,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Claude3Sonnet: {
		Name:            Claude3Sonnet,
		MaxPromptLength: 8192,
		MaxOutputLength: 4096,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Claude3Haiku: {
		Name:            Claude3Haiku,
		MaxPromptLength: 8192,
		MaxOutputLength: 4096,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
}

var AnthropicProviderInfo = spec.ProviderInfo{
	Name:   ProviderNameAnthropic,
	APIKey: "",
	Origin: "https://api.anthropic.com",
	Type:   spec.InbuiltSpecific,

	APIKeyHeaderKey: "x-api-key",
	DefaultHeaders: map[string]string{
		"content-type":      "application/json",
		"accept":            "application/json",
		"anthropic-version": "2023-06-01",
	},
	ChatCompletionPathPrefix: "/v1/messages",
}
