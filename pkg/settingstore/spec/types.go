package spec

import (
	aiproviderSpec "github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

type ModelSetting struct {
	DisplayName          string                          `json:"displayName"                    required:"true"`
	IsEnabled            bool                            `json:"isEnabled"                      required:"true"`
	Stream               *bool                           `json:"stream,omitempty"`
	MaxPromptLength      *int                            `json:"maxPromptLength,omitempty"`
	MaxOutputLength      *int                            `json:"maxOutputLength,omitempty"`
	Temperature          *float64                        `json:"temperature,omitempty"`
	Reasoning            *aiproviderSpec.ReasoningParams `json:"reasoning,omitempty"`
	SystemPrompt         *string                         `json:"systemPrompt,omitempty"`
	Timeout              *int                            `json:"timeout,omitempty"`
	AdditionalParameters *map[string]any                 `json:"additionalParameters,omitempty"`
}

// AISetting represents the settings for an AI provider.
type AISetting struct {
	IsEnabled                bool                                      `json:"isEnabled"`
	APIKey                   string                                    `json:"apiKey"`
	DefaultModel             aiproviderSpec.ModelName                  `json:"defaultModel"`
	Origin                   string                                    `json:"origin"`
	ChatCompletionPathPrefix string                                    `json:"chatCompletionPathPrefix"`
	ModelSettings            map[aiproviderSpec.ModelName]ModelSetting `json:"modelSettings"`
}

// AISettingsSchema represents the schema for AI settings for different providers.
type AISettingsSchema map[aiproviderSpec.ProviderName]AISetting

// AppSettings app settings.
type AppSettings struct {
	DefaultProvider aiproviderSpec.ProviderName `json:"defaultProvider"`
}

// SettingsSchema represents the complete settings schema including app settings.
type SettingsSchema struct {
	Version    string           `json:"version"`
	AISettings AISettingsSchema `json:"aiSettings"`
	App        AppSettings      `json:"app"`
}
