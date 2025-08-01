package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

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
	SlugClaude4Opus    spec.ModelSlug = "opus4"
	SlugClaude4Sonnet  spec.ModelSlug = "sonnet4"
	SlugClaude37Sonnet spec.ModelSlug = "sonnet37"
	SlugClaude35Sonnet spec.ModelSlug = "sonnet35"
	SlugClaude35Haiku  spec.ModelSlug = "haiku35"
)

const (
	ModelPresetIDClaude4Opus    spec.ModelPresetID = "opus4"
	ModelPresetIDClaude4Sonnet  spec.ModelPresetID = "sonnet4"
	ModelPresetIDClaude37Sonnet spec.ModelPresetID = "sonnet37"
	ModelPresetIDClaude35Sonnet spec.ModelPresetID = "sonnet35"
	ModelPresetIDClaude35Haiku  spec.ModelPresetID = "haiku35"
)

var AnthropicModelPresets = map[spec.ModelPresetID]spec.ModelPreset{
	ModelPresetIDClaude4Opus: {
		ID:          ModelPresetIDClaude4Opus,
		Name:        Claude4Opus,
		DisplayName: DisplayNameClaude4Opus,
		Slug:        SlugClaude4Opus,
		IsEnabled:   true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(65536),
		MaxOutputLength: IntPtr(16384),
		Temperature:     Float64Ptr(0.1),
		Reasoning: &spec.ReasoningParams{
			Type:   spec.ReasoningTypeHybridWithTokens,
			Tokens: 1024,
		},
		SystemPrompt: StringPtr(""),
		Timeout:      IntPtr(120),
	},
	ModelPresetIDClaude4Sonnet: {
		ID:          ModelPresetIDClaude4Sonnet,
		Name:        Claude4Sonnet,
		DisplayName: DisplayNameClaude4Sonnet,
		Slug:        SlugClaude4Sonnet,
		IsEnabled:   true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(65536),
		MaxOutputLength: IntPtr(16384),
		Temperature:     Float64Ptr(0.1),
		Reasoning: &spec.ReasoningParams{
			Type:   spec.ReasoningTypeHybridWithTokens,
			Tokens: 1024,
		},
		SystemPrompt: StringPtr(""),
		Timeout:      IntPtr(120),
	},
	ModelPresetIDClaude37Sonnet: {
		ID:          ModelPresetIDClaude37Sonnet,
		Name:        Claude37Sonnet,
		DisplayName: DisplayNameClaude37Sonnet,
		Slug:        SlugClaude37Sonnet,
		IsEnabled:   false,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(65536),
		MaxOutputLength: IntPtr(16384),
		Temperature:     Float64Ptr(0.1),
		Reasoning: &spec.ReasoningParams{
			Type:   spec.ReasoningTypeHybridWithTokens,
			Tokens: 1024,
		},
		SystemPrompt: StringPtr(""),
		Timeout:      IntPtr(120),
	},
	ModelPresetIDClaude35Sonnet: {
		ID:          ModelPresetIDClaude35Sonnet,
		Name:        Claude35Sonnet,
		DisplayName: DisplayNameClaude35Sonnet,
		Slug:        SlugClaude35Sonnet,
		IsEnabled:   false,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(16384),
		MaxOutputLength: IntPtr(8192),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	ModelPresetIDClaude35Haiku: {
		ID:          ModelPresetIDClaude35Haiku,
		Name:        Claude35Haiku,
		DisplayName: DisplayNameClaude35Haiku,
		Slug:        SlugClaude35Haiku,
		IsEnabled:   false,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(16384),
		MaxOutputLength: IntPtr(8192),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
}

var AnthropicProviderPreset = spec.ProviderPreset{
	DefaultModelPresetID: ModelPresetIDClaude4Sonnet,
	ModelPresets:         AnthropicModelPresets,
}
