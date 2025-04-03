package spec

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/anthropic"
	"github.com/flexigpt/flexiui/pkg/aiprovider/deepseek"
	"github.com/flexigpt/flexiui/pkg/aiprovider/google"
	"github.com/flexigpt/flexiui/pkg/aiprovider/huggingface"
	"github.com/flexigpt/flexiui/pkg/aiprovider/llamacpp"
	"github.com/flexigpt/flexiui/pkg/aiprovider/openai"
	aiproviderSpec "github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

// Define the default AI settings.
var DefaultAISettings = map[aiproviderSpec.ProviderName]AISetting{
	anthropic.ProviderNameAnthropic: {
		IsEnabled:          true,
		APIKey:             anthropic.AnthropicProviderInfo.APIKey,
		DefaultModel:       anthropic.AnthropicProviderInfo.DefaultModel,
		Origin:      anthropic.AnthropicProviderInfo.Origin,
		DefaultTemperature: anthropic.AnthropicProviderInfo.DefaultTemperature,
		AdditionalSettings: anthropic.AnthropicProviderInfo.AdditionalSettings,
	},
	deepseek.ProviderNameDeepseek: {
		IsEnabled:          false,
		APIKey:             deepseek.DeepseekProviderInfo.APIKey,
		DefaultModel:       deepseek.DeepseekProviderInfo.DefaultModel,
		Origin:      deepseek.DeepseekProviderInfo.Origin,
		DefaultTemperature: deepseek.DeepseekProviderInfo.DefaultTemperature,
		AdditionalSettings: deepseek.DeepseekProviderInfo.AdditionalSettings,
	},
	google.ProviderNameGoogle: {
		IsEnabled:          false,
		APIKey:             google.GoogleProviderInfo.APIKey,
		DefaultModel:       google.GoogleProviderInfo.DefaultModel,
		Origin:      google.GoogleProviderInfo.Origin,
		DefaultTemperature: google.GoogleProviderInfo.DefaultTemperature,
		AdditionalSettings: google.GoogleProviderInfo.AdditionalSettings,
	},
	huggingface.ProviderNameHuggingFace: {
		IsEnabled:          false,
		APIKey:             huggingface.HuggingfaceProviderInfo.APIKey,
		DefaultModel:       huggingface.HuggingfaceProviderInfo.DefaultModel,
		Origin:      huggingface.HuggingfaceProviderInfo.Origin,
		DefaultTemperature: huggingface.HuggingfaceProviderInfo.DefaultTemperature,
		AdditionalSettings: huggingface.HuggingfaceProviderInfo.AdditionalSettings,
	},
	llamacpp.ProviderNameLlamaCPP: {
		IsEnabled:          false,
		APIKey:             llamacpp.LlamacppProviderInfo.APIKey,
		DefaultModel:       llamacpp.LlamacppProviderInfo.DefaultModel,
		Origin:      llamacpp.LlamacppProviderInfo.Origin,
		DefaultTemperature: llamacpp.LlamacppProviderInfo.DefaultTemperature,
		AdditionalSettings: llamacpp.LlamacppProviderInfo.AdditionalSettings,
	},
	openai.ProviderNameOpenAI: {
		IsEnabled:          true,
		APIKey:             openai.OpenAIProviderInfo.APIKey,
		DefaultModel:       openai.OpenAIProviderInfo.DefaultModel,
		Origin:      openai.OpenAIProviderInfo.Origin,
		DefaultTemperature: openai.OpenAIProviderInfo.DefaultTemperature,
		AdditionalSettings: openai.OpenAIProviderInfo.AdditionalSettings,
	},
}

// Define the default settings data.
var DefaultSettingsData = SettingsSchema{
	Version: "1.0",
	App: AppSettings{
		DefaultProvider: openai.ProviderNameOpenAI,
	},
	AISettings: DefaultAISettings,
}

// Define the sensitive keys.
var SensitiveKeys = []string{
	"aiSettings.openai.apiKey",
	"aiSettings.anthropic.apiKey",
	"aiSettings.deepseek.apiKey",
	"aiSettings.huggingface.apiKey",
	"aiSettings.google.apiKey",
	"aiSettings.llamacpp.apiKey",
}
