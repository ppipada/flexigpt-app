package spec

import (
	"github.com/flexigpt/flexiui/pkggo/aiprovider/anthropic"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/google"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/huggingface"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/llamacpp"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/openai"
	aiproviderSpec "github.com/flexigpt/flexiui/pkggo/aiprovider/spec"
)

// Define the default AI settings.
var DefaultAISettings = map[aiproviderSpec.ProviderName]AISetting{
	anthropic.ProviderNameAnthropic: {
		IsEnabled:          true,
		APIKey:             anthropic.AnthropicProviderInfo.APIKey,
		DefaultModel:       anthropic.AnthropicProviderInfo.DefaultModel,
		DefaultOrigin:      anthropic.AnthropicProviderInfo.DefaultOrigin,
		DefaultTemperature: anthropic.AnthropicProviderInfo.DefaultTemperature,
		AdditionalSettings: anthropic.AnthropicProviderInfo.AdditionalSettings,
	},
	google.ProviderNameGoogle: {
		IsEnabled:          false,
		APIKey:             google.GoogleProviderInfo.APIKey,
		DefaultModel:       google.GoogleProviderInfo.DefaultModel,
		DefaultOrigin:      google.GoogleProviderInfo.DefaultOrigin,
		DefaultTemperature: google.GoogleProviderInfo.DefaultTemperature,
		AdditionalSettings: google.GoogleProviderInfo.AdditionalSettings,
	},
	huggingface.ProviderNameHuggingFace: {
		IsEnabled:          false,
		APIKey:             huggingface.HuggingfaceProviderInfo.APIKey,
		DefaultModel:       huggingface.HuggingfaceProviderInfo.DefaultModel,
		DefaultOrigin:      huggingface.HuggingfaceProviderInfo.DefaultOrigin,
		DefaultTemperature: huggingface.HuggingfaceProviderInfo.DefaultTemperature,
		AdditionalSettings: huggingface.HuggingfaceProviderInfo.AdditionalSettings,
	},
	llamacpp.ProviderNameLlamaCPP: {
		IsEnabled:          false,
		APIKey:             llamacpp.LlamacppProviderInfo.APIKey,
		DefaultModel:       llamacpp.LlamacppProviderInfo.DefaultModel,
		DefaultOrigin:      llamacpp.LlamacppProviderInfo.DefaultOrigin,
		DefaultTemperature: llamacpp.LlamacppProviderInfo.DefaultTemperature,
		AdditionalSettings: llamacpp.LlamacppProviderInfo.AdditionalSettings,
	},
	openai.ProviderNameOpenAI: {
		IsEnabled:          true,
		APIKey:             openai.OpenAIProviderInfo.APIKey,
		DefaultModel:       openai.OpenAIProviderInfo.DefaultModel,
		DefaultOrigin:      openai.OpenAIProviderInfo.DefaultOrigin,
		DefaultTemperature: openai.OpenAIProviderInfo.DefaultTemperature,
		AdditionalSettings: openai.OpenAIProviderInfo.AdditionalSettings,
	},
}

// Define the default settings data.
var DefaultSettingsData = SettingsSchema{
	App: AppSettings{
		DefaultProvider: openai.ProviderNameOpenAI,
	},
	AISettings: DefaultAISettings,
}

// Define the sensitive keys.
var SensitiveKeys = []string{
	"aiSettings.openai.apiKey",
	"aiSettings.anthropic.apiKey",
	"aiSettings.huggingface.apiKey",
	"aiSettings.google.apiKey",
	"aiSettings.llamacpp.apiKey",
}
