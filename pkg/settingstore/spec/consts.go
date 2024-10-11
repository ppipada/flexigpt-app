package spec

import (
	aiproviderSpec "github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

// Define the default AI settings
var DefaultAISettings = map[aiproviderSpec.ProviderName]AISetting{
	aiproviderSpec.ProviderNameAnthropic: {
		ApiKey:             aiproviderSpec.AnthropicProviderInfo.ApiKey,
		DefaultModel:       aiproviderSpec.AnthropicProviderInfo.DefaultModel,
		DefaultOrigin:      aiproviderSpec.AnthropicProviderInfo.DefaultOrigin,
		DefaultTemperature: aiproviderSpec.AnthropicProviderInfo.DefaultTemperature,
		AdditionalSettings: aiproviderSpec.AnthropicProviderInfo.AdditionalSettings,
	},
	aiproviderSpec.ProviderNameGoogle: {
		ApiKey:             aiproviderSpec.GoogleProviderInfo.ApiKey,
		DefaultModel:       aiproviderSpec.GoogleProviderInfo.DefaultModel,
		DefaultOrigin:      aiproviderSpec.GoogleProviderInfo.DefaultOrigin,
		DefaultTemperature: aiproviderSpec.GoogleProviderInfo.DefaultTemperature,
		AdditionalSettings: aiproviderSpec.GoogleProviderInfo.AdditionalSettings,
	},
	aiproviderSpec.ProviderNameHuggingFace: {
		ApiKey:             aiproviderSpec.HuggingfaceProviderInfo.ApiKey,
		DefaultModel:       aiproviderSpec.HuggingfaceProviderInfo.DefaultModel,
		DefaultOrigin:      aiproviderSpec.HuggingfaceProviderInfo.DefaultOrigin,
		DefaultTemperature: aiproviderSpec.HuggingfaceProviderInfo.DefaultTemperature,
		AdditionalSettings: aiproviderSpec.HuggingfaceProviderInfo.AdditionalSettings,
	},
	aiproviderSpec.ProviderNameLlamaCPP: {
		ApiKey:             aiproviderSpec.LlamacppProviderInfo.ApiKey,
		DefaultModel:       aiproviderSpec.LlamacppProviderInfo.DefaultModel,
		DefaultOrigin:      aiproviderSpec.LlamacppProviderInfo.DefaultOrigin,
		DefaultTemperature: aiproviderSpec.LlamacppProviderInfo.DefaultTemperature,
		AdditionalSettings: aiproviderSpec.LlamacppProviderInfo.AdditionalSettings,
	},
	aiproviderSpec.ProviderNameOpenAI: {
		ApiKey:             aiproviderSpec.OpenAIProviderInfo.ApiKey,
		DefaultModel:       aiproviderSpec.OpenAIProviderInfo.DefaultModel,
		DefaultOrigin:      aiproviderSpec.OpenAIProviderInfo.DefaultOrigin,
		DefaultTemperature: aiproviderSpec.OpenAIProviderInfo.DefaultTemperature,
		AdditionalSettings: aiproviderSpec.OpenAIProviderInfo.AdditionalSettings,
	},
}

// Define the default settings data
var DefaultSettingsData = SettingsSchema{
	App: AppSettings{
		DefaultProvider: aiproviderSpec.ProviderNameOpenAI,
	},
	AISettings: DefaultAISettings,
}

// Define the sensitive keys
var SensitiveKeys = []string{
	"aiSettings.openai.apiKey",
	"aiSettings.anthropic.apiKey",
	"aiSettings.huggingface.apiKey",
	"aiSettings.google.apiKey",
	"aiSettings.llamacpp.apiKey",
}
