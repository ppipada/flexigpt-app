package spec

import (
	aiproviderSpec "github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

// AISetting represents the settings for an AI provider.
type AISetting struct {
	IsEnabled          bool                     `json:"isEnabled"`
	APIKey             string                   `json:"apiKey"`
	DefaultModel       aiproviderSpec.ModelName `json:"defaultModel"`
	DefaultTemperature float64                  `json:"defaultTemperature"`
	DefaultOrigin      string                   `json:"defaultOrigin"`
	AdditionalSettings map[string]any           `json:"additionalSettings"`
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
