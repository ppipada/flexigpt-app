package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"
)

const ProviderNameAnthropic spec.ProviderName = "anthropic"

const (
	Claude4Opus    spec.ModelName = "claude-opus-4-20250514"
	Claude4Sonnet  spec.ModelName = "claude-sonnet-4-20250514"
	Claude37Sonnet spec.ModelName = "claude-3-7-sonnet-20250219"
	Claude35Sonnet spec.ModelName = "claude-3-5-sonnet-20241022"
	Claude35Haiku  spec.ModelName = "claude-3-5-haiku-20241022"
	Claude3Opus    spec.ModelName = "claude-3-opus-20240229"
	Claude3Sonnet  spec.ModelName = "claude-3-sonnet-20240229"
	Claude3Haiku   spec.ModelName = "claude-3-haiku-20240307"
)

const (
	DisplayNameClaude4Opus    = "Claude 4 Opus"
	DisplayNameClaude4Sonnet  = "Claude 4 Sonnet"
	DisplayNameClaude37Sonnet = "Claude 3.7 Sonnet"
	DisplayNameClaude35Sonnet = "Claude 3.5 Sonnet"
	DisplayNameClaude35Haiku  = "Claude 3.5 Haiku"
	DisplayNameClaude3Opus    = "Claude 3 Opus"
	DisplayNameClaude3Sonnet  = "Claude 3 Sonnet"
	DisplayNameClaude3Haiku   = "Claude 3 Haiku"
)

var AnthropicModelDefaults = map[spec.ModelName]spec.ModelDefaults{
	Claude4Opus: {
		DisplayName: DisplayNameClaude4Opus,
		IsEnabled:   true,
	},
	Claude4Sonnet: {
		DisplayName: DisplayNameClaude4Sonnet,
		IsEnabled:   true,
	},
	Claude37Sonnet: {
		DisplayName: DisplayNameClaude37Sonnet,
		IsEnabled:   false,
	},
	Claude35Sonnet: {
		DisplayName: DisplayNameClaude35Sonnet,
		IsEnabled:   false,
	},
	Claude35Haiku: {
		DisplayName: DisplayNameClaude35Haiku,
		IsEnabled:   false,
	},
	Claude3Opus: {
		DisplayName: DisplayNameClaude3Opus,
		IsEnabled:   false,
	},
	Claude3Sonnet: {
		DisplayName: DisplayNameClaude3Sonnet,
		IsEnabled:   false,
	},
	Claude3Haiku: {
		DisplayName: DisplayNameClaude3Haiku,
		IsEnabled:   false,
	},
}

var AnthropicModels = map[spec.ModelName]spec.ModelParams{
	Claude4Opus: {
		Name:            Claude4Opus,
		MaxPromptLength: 65536,
		MaxOutputLength: 16384,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		Reasoning: &spec.ReasoningParams{
			Type:   spec.ReasoningTypeHybridWithTokens,
			Tokens: 1024,
		},
		SystemPrompt: "",
		Timeout:      120,
	},
	Claude4Sonnet: {
		Name:            Claude4Sonnet,
		MaxPromptLength: 65536,
		MaxOutputLength: 16384,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		Reasoning: &spec.ReasoningParams{
			Type:   spec.ReasoningTypeHybridWithTokens,
			Tokens: 1024,
		},
		SystemPrompt: "",
		Timeout:      120,
	},
	Claude37Sonnet: {
		Name:            Claude37Sonnet,
		MaxPromptLength: 65536,
		MaxOutputLength: 16384,
		Temperature:     Float64Ptr(0.1),
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
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Claude35Haiku: {
		Name:            Claude35Haiku,
		MaxPromptLength: 16384,
		MaxOutputLength: 8192,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Claude3Opus: {
		Name:            Claude3Opus,
		MaxPromptLength: 16384,
		MaxOutputLength: 4096,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Claude3Sonnet: {
		Name:            Claude3Sonnet,
		MaxPromptLength: 8192,
		MaxOutputLength: 4096,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Claude3Haiku: {
		Name:            Claude3Haiku,
		MaxPromptLength: 8192,
		MaxOutputLength: 4096,
		Temperature:     Float64Ptr(0.1),
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
