package spec

import (
	modelSpec "github.com/ppipada/flexigpt-app/pkg/model/spec"
)

type ModelSetting struct {
	DisplayName          string                     `json:"displayName"                    required:"true"`
	IsEnabled            bool                       `json:"isEnabled"                      required:"true"`
	Stream               *bool                      `json:"stream,omitempty"`
	MaxPromptLength      *int                       `json:"maxPromptLength,omitempty"`
	MaxOutputLength      *int                       `json:"maxOutputLength,omitempty"`
	Temperature          *float64                   `json:"temperature,omitempty"`
	Reasoning            *modelSpec.ReasoningParams `json:"reasoning,omitempty"`
	SystemPrompt         *string                    `json:"systemPrompt,omitempty"`
	Timeout              *int                       `json:"timeout,omitempty"`
	AdditionalParameters *map[string]any            `json:"additionalParameters,omitempty"`
}

// AISetting represents the settings for an AI provider.
type AISetting struct {
	IsEnabled                bool                                 `json:"isEnabled"`
	APIKey                   string                               `json:"apiKey"`
	DefaultModel             modelSpec.ModelName                  `json:"defaultModel"`
	Origin                   string                               `json:"origin"`
	ChatCompletionPathPrefix string                               `json:"chatCompletionPathPrefix"`
	ModelSettings            map[modelSpec.ModelName]ModelSetting `json:"modelSettings"`
}

// AISettingsSchema represents the schema for AI settings for different providers.
type AISettingsSchema map[modelSpec.ProviderName]AISetting

// AppSettings app settings.
type AppSettings struct {
	DefaultProvider modelSpec.ProviderName `json:"defaultProvider"`
}

// SettingsSchema represents the complete settings schema including app settings.
type SettingsSchema struct {
	Version    string           `json:"version"`
	AISettings AISettingsSchema `json:"aiSettings"`
	App        AppSettings      `json:"app"`
}
