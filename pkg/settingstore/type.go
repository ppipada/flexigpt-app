package settingstore

import (
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

// AISetting represents the settings for an AI provider.
type AISetting struct {
	IsEnabled                bool   `json:"isEnabled"`
	APIKey                   string `json:"apiKey"`
	Origin                   string `json:"origin"`
	ChatCompletionPathPrefix string `json:"chatCompletionPathPrefix"`
}

// AISettingsSchema represents the schema for AI settings for different providers.
type AISettingsSchema map[modelpresetSpec.ProviderName]AISetting

// AppSettings app settings.
type AppSettings struct {
	DefaultProvider modelpresetSpec.ProviderName `json:"defaultProvider"`
}

// SettingsSchema represents the complete settings schema including app settings.
type SettingsSchema struct {
	Version    string           `json:"version"`
	AISettings AISettingsSchema `json:"aiSettings"`
	App        AppSettings      `json:"app"`
}
