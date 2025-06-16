package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

const ProviderNameAnthropic spec.ProviderName = "anthropic"

const (
	Claude4Opus    spec.ModelName = "claude-opus-4-20250514"
	Claude4Sonnet  spec.ModelName = "claude-sonnet-4-20250514"
	Claude37Sonnet spec.ModelName = "claude-3-7-sonnet-20250219"
	Claude35Sonnet spec.ModelName = "claude-3-5-sonnet-20241022"
	Claude35Haiku  spec.ModelName = "claude-3-5-haiku-20241022"
)

const (
	DisplayNameClaude4Opus    spec.ModelDisplayName = "Claude 4 Opus"
	DisplayNameClaude4Sonnet  spec.ModelDisplayName = "Claude 4 Sonnet"
	DisplayNameClaude37Sonnet spec.ModelDisplayName = "Claude 3.7 Sonnet"
	DisplayNameClaude35Sonnet spec.ModelDisplayName = "Claude 3.5 Sonnet"
	DisplayNameClaude35Haiku  spec.ModelDisplayName = "Claude 3.5 Haiku"
)

const (
	ShortCommandClaude4Opus    spec.ModelShortCommand = "opus4"
	ShortCommandClaude4Sonnet  spec.ModelShortCommand = "sonnet4"
	ShortCommandClaude37Sonnet spec.ModelShortCommand = "sonnet37"
	ShortCommandClaude35Sonnet spec.ModelShortCommand = "sonnet35"
	ShortCommandClaude35Haiku  spec.ModelShortCommand = "haiku35"
)

var AnthropicModels = map[spec.ModelName]spec.ModelPreset{
	Claude4Opus: {
		Name:            Claude4Opus,
		DisplayName:     DisplayNameClaude4Opus,
		IsEnabled:       true,
		ShortCommand:    ShortCommandClaude4Opus,
		MaxPromptLength: IntPtr(65536),
		MaxOutputLength: IntPtr(16384),
		Temperature:     Float64Ptr(0.1),
		Stream:          BoolPtr(true),
		Reasoning: &spec.ReasoningParams{
			Type:   spec.ReasoningTypeHybridWithTokens,
			Tokens: 1024,
		},
		SystemPrompt: StringPtr(""),
		Timeout:      IntPtr(120),
	},
	Claude4Sonnet: {
		Name:            Claude4Sonnet,
		DisplayName:     DisplayNameClaude4Sonnet,
		IsEnabled:       true,
		ShortCommand:    ShortCommandClaude4Sonnet,
		MaxPromptLength: IntPtr(65536),
		MaxOutputLength: IntPtr(16384),
		Temperature:     Float64Ptr(0.1),
		Stream:          BoolPtr(true),
		Reasoning: &spec.ReasoningParams{
			Type:   spec.ReasoningTypeHybridWithTokens,
			Tokens: 1024,
		},
		SystemPrompt: StringPtr(""),
		Timeout:      IntPtr(120),
	},
	Claude37Sonnet: {
		Name:            Claude37Sonnet,
		DisplayName:     DisplayNameClaude37Sonnet,
		IsEnabled:       false,
		ShortCommand:    ShortCommandClaude37Sonnet,
		MaxPromptLength: IntPtr(65536),
		MaxOutputLength: IntPtr(16384),
		Temperature:     Float64Ptr(0.1),
		Stream:          BoolPtr(true),
		Reasoning: &spec.ReasoningParams{
			Type:   spec.ReasoningTypeHybridWithTokens,
			Tokens: 1024,
		},
		SystemPrompt: StringPtr(""),
		Timeout:      IntPtr(120),
	},
	Claude35Sonnet: {
		Name:            Claude35Sonnet,
		DisplayName:     DisplayNameClaude35Sonnet,
		IsEnabled:       false,
		ShortCommand:    ShortCommandClaude35Sonnet,
		MaxPromptLength: IntPtr(16384),
		MaxOutputLength: IntPtr(8192),
		Temperature:     Float64Ptr(0.1),
		Stream:          BoolPtr(true),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	Claude35Haiku: {
		Name:            Claude35Haiku,
		DisplayName:     DisplayNameClaude35Haiku,
		IsEnabled:       false,
		ShortCommand:    ShortCommandClaude35Haiku,
		MaxPromptLength: IntPtr(16384),
		MaxOutputLength: IntPtr(8192),
		Temperature:     Float64Ptr(0.1),
		Stream:          BoolPtr(true),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
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
