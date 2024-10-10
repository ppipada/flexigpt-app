package spec

import (
	aiproviderSpec "github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

// AISetting represents the settings for an AI provider.
type AISetting struct {
	ApiKey             string                   `json:"apiKey"`
	DefaultModel       aiproviderSpec.ModelName `json:"defaultModel"`
	DefaultTemperature float64                  `json:"defaultTemperature"`
	DefaultOrigin      string                   `json:"defaultOrigin"`
	AdditionalSettings map[string]interface{}   `json:"additionalSettings"`
}

// AISettingsSchema represents the schema for AI settings for different providers.
type AISettingsSchema map[aiproviderSpec.ProviderName]AISetting

// AppSettings app settings.
type AppSettings struct {
	DefaultProvider aiproviderSpec.ProviderName `json:"defaultProvider"`
}

// SettingsSchema represents the complete settings schema including app settings.
type SettingsSchema struct {
	AISettings AISettingsSchema `json:"aiSettings"`
	App        AppSettings      `json:"app"`
}

// ISettingStoreAPI defines the interface for settings API.
type ISettingStoreAPI interface {
	GetAllSettings(forceFetch bool) (*SettingsSchema, error)
	SetSetting(key string, value interface{}) error
}
