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
		IsEnabled:    true,
		APIKey:       anthropic.AnthropicProviderInfo.APIKey,
		DefaultModel: anthropic.AnthropicProviderInfo.DefaultModel,
		Origin:       anthropic.AnthropicProviderInfo.Origin,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			anthropic.Claude37Sonnet: {
				DisplayName: anthropic.AnthropicModels[anthropic.Claude37Sonnet].DisplayName,
				IsEnabled:   true,
			},
			anthropic.Claude35Sonnet: {
				DisplayName: anthropic.AnthropicModels[anthropic.Claude35Sonnet].DisplayName,
				IsEnabled:   true,
			},
			anthropic.Claude35Haiku: {
				DisplayName: anthropic.AnthropicModels[anthropic.Claude35Haiku].DisplayName,
				IsEnabled:   true,
			},
			anthropic.Claude3Opus: {
				DisplayName: anthropic.AnthropicModels[anthropic.Claude3Opus].DisplayName,
				IsEnabled:   false,
			},
			anthropic.Claude3Sonnet: {
				DisplayName: anthropic.AnthropicModels[anthropic.Claude3Sonnet].DisplayName,
				IsEnabled:   false,
			},
			anthropic.Claude3Haiku: {
				DisplayName: anthropic.AnthropicModels[anthropic.Claude3Haiku].DisplayName,
				IsEnabled:   false,
			},
		},
	},

	deepseek.ProviderNameDeepseek: {
		IsEnabled:    false,
		APIKey:       deepseek.DeepseekProviderInfo.APIKey,
		DefaultModel: deepseek.DeepseekProviderInfo.DefaultModel,
		Origin:       deepseek.DeepseekProviderInfo.Origin,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			deepseek.DeepseekChat: {
				DisplayName: deepseek.DeepseekModels[deepseek.DeepseekChat].DisplayName,
				IsEnabled:   true,
			},
			deepseek.DeepseekReasoner: {
				DisplayName: deepseek.DeepseekModels[deepseek.DeepseekReasoner].DisplayName,
				IsEnabled:   true,
			},
		},
	},
	google.ProviderNameGoogle: {
		IsEnabled:    false,
		APIKey:       google.GoogleProviderInfo.APIKey,
		DefaultModel: google.GoogleProviderInfo.DefaultModel,
		Origin:       google.GoogleProviderInfo.Origin,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			google.Gemini2FlashExp: {
				DisplayName: google.GoogleModels[google.Gemini2FlashExp].DisplayName,
				IsEnabled:   true,
			},
			google.Gemini15Flash: {
				DisplayName: google.GoogleModels[google.Gemini15Flash].DisplayName,
				IsEnabled:   true,
			},
			google.Gemini15Pro: {
				DisplayName: google.GoogleModels[google.Gemini15Pro].DisplayName,
				IsEnabled:   true,
			},
		},
	},
	huggingface.ProviderNameHuggingFace: {
		IsEnabled:    false,
		APIKey:       huggingface.HuggingfaceProviderInfo.APIKey,
		DefaultModel: huggingface.HuggingfaceProviderInfo.DefaultModel,
		Origin:       huggingface.HuggingfaceProviderInfo.Origin,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			huggingface.DeepseekCoder13BInstruct: {
				DisplayName: huggingface.HuggingfaceModels[huggingface.DeepseekCoder13BInstruct].DisplayName,
				IsEnabled:   true,
			},
		},
	},
	llamacpp.ProviderNameLlamaCPP: {
		IsEnabled:    false,
		APIKey:       llamacpp.LlamacppProviderInfo.APIKey,
		DefaultModel: llamacpp.LlamacppProviderInfo.DefaultModel,
		Origin:       llamacpp.LlamacppProviderInfo.Origin,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			llamacpp.Llama31: {
				DisplayName: llamacpp.LlamacppModels[llamacpp.Llama31].DisplayName,
				IsEnabled:   true,
			},
			llamacpp.Llama3: {
				DisplayName: llamacpp.LlamacppModels[llamacpp.Llama3].DisplayName,
				IsEnabled:   true,
			},
		},
	},
	openai.ProviderNameOpenAI: {
		IsEnabled:    true,
		APIKey:       openai.OpenAIProviderInfo.APIKey,
		DefaultModel: openai.OpenAIProviderInfo.DefaultModel,
		Origin:       openai.OpenAIProviderInfo.Origin,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			openai.GPTO3Mini: {
				DisplayName: openai.OpenAIModels[openai.GPTO3Mini].DisplayName,
				IsEnabled:   true,
			},
			openai.GPTO1: {
				DisplayName: openai.OpenAIModels[openai.GPTO1].DisplayName,
				IsEnabled:   true,
			},
			openai.GPTO1Preview: {
				DisplayName: openai.OpenAIModels[openai.GPTO1Preview].DisplayName,
				IsEnabled:   false,
			},
			openai.GPTO1Mini: {
				DisplayName: openai.OpenAIModels[openai.GPTO1Mini].DisplayName,
				IsEnabled:   false,
			},
			openai.GPT45Preview: {
				DisplayName: openai.OpenAIModels[openai.GPT45Preview].DisplayName,
				IsEnabled:   true,
			},
			openai.GPT4OMini: {
				DisplayName: openai.OpenAIModels[openai.GPT4OMini].DisplayName,
				IsEnabled:   false,
			},
			openai.GPT4O: {
				DisplayName: openai.OpenAIModels[openai.GPT4O].DisplayName,
				IsEnabled:   true,
			},
			openai.GPT4: {
				DisplayName: openai.OpenAIModels[openai.GPT4].DisplayName,
				IsEnabled:   false,
			},
			openai.GPT35Turbo: {
				DisplayName: openai.OpenAIModels[openai.GPT35Turbo].DisplayName,
				IsEnabled:   false,
			},
		},
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
