package spec

import (
	aiproviderConsts "github.com/flexigpt/flexiui/pkg/aiprovider/consts"
	aiproviderSpec "github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

// Define the default AI settings.
var DefaultAISettings = map[aiproviderSpec.ProviderName]AISetting{
	aiproviderConsts.ProviderNameAnthropic: {
		IsEnabled:                true,
		APIKey:                   aiproviderConsts.AnthropicProviderInfo.APIKey,
		DefaultModel:             aiproviderConsts.Claude37Sonnet,
		Origin:                   aiproviderConsts.AnthropicProviderInfo.Origin,
		ChatCompletionPathPrefix: aiproviderConsts.AnthropicProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			aiproviderConsts.Claude37Sonnet: {
				DisplayName: aiproviderConsts.DisplayNameClaude37Sonnet,
				IsEnabled:   true,
			},
			aiproviderConsts.Claude35Sonnet: {
				DisplayName: aiproviderConsts.DisplayNameClaude35Sonnet,
				IsEnabled:   true,
			},
			aiproviderConsts.Claude35Haiku: {
				DisplayName: aiproviderConsts.DisplayNameClaude35Haiku,
				IsEnabled:   true,
			},
			aiproviderConsts.Claude3Opus: {
				DisplayName: aiproviderConsts.DisplayNameClaude3Opus,
				IsEnabled:   false,
			},
			aiproviderConsts.Claude3Sonnet: {
				DisplayName: aiproviderConsts.DisplayNameClaude3Sonnet,
				IsEnabled:   false,
			},
			aiproviderConsts.Claude3Haiku: {
				DisplayName: aiproviderConsts.DisplayNameClaude3Haiku,
				IsEnabled:   false,
			},
		},
	},

	aiproviderConsts.ProviderNameDeepseek: {
		IsEnabled:                false,
		APIKey:                   aiproviderConsts.DeepseekProviderInfo.APIKey,
		DefaultModel:             aiproviderConsts.DeepseekChat,
		Origin:                   aiproviderConsts.DeepseekProviderInfo.Origin,
		ChatCompletionPathPrefix: aiproviderConsts.DeepseekProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			aiproviderConsts.DeepseekChat: {
				DisplayName: aiproviderConsts.DisplayNameDeepseekChat,
				IsEnabled:   true,
			},
			aiproviderConsts.DeepseekReasoner: {
				DisplayName: aiproviderConsts.DisplayNameDeepseekReasoner,
				IsEnabled:   true,
			},
		},
	},
	aiproviderConsts.ProviderNameGoogle: {
		IsEnabled:                false,
		APIKey:                   aiproviderConsts.GoogleProviderInfo.APIKey,
		DefaultModel:             aiproviderConsts.Gemini2Flash,
		Origin:                   aiproviderConsts.GoogleProviderInfo.Origin,
		ChatCompletionPathPrefix: aiproviderConsts.GoogleProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			aiproviderConsts.Gemini25Pro: {
				DisplayName: aiproviderConsts.DisplayNameGemini25Pro,
				IsEnabled:   true,
			},
			aiproviderConsts.Gemini2Flash: {
				DisplayName: aiproviderConsts.DisplayNameGemini2Flash,
				IsEnabled:   true,
			},
			aiproviderConsts.Gemini2FlashLite: {
				DisplayName: aiproviderConsts.DisplayNameGemini2FlashLite,
				IsEnabled:   false,
			},
			aiproviderConsts.Gemini15Pro: {
				DisplayName: aiproviderConsts.DisplayNameGemini15Pro,
				IsEnabled:   false,
			},
		},
	},
	aiproviderConsts.ProviderNameHuggingFace: {
		IsEnabled:                false,
		APIKey:                   aiproviderConsts.HuggingfaceProviderInfo.APIKey,
		DefaultModel:             aiproviderConsts.DeepseekCoder13BInstruct,
		Origin:                   aiproviderConsts.HuggingfaceProviderInfo.Origin,
		ChatCompletionPathPrefix: aiproviderConsts.HuggingfaceProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			aiproviderConsts.DeepseekCoder13BInstruct: {
				DisplayName: aiproviderConsts.DisplayNameDeepseekCoder13BInstruct,
				IsEnabled:   true,
			},
		},
	},
	aiproviderConsts.ProviderNameLlamaCPP: {
		IsEnabled:                false,
		APIKey:                   aiproviderConsts.LlamacppProviderInfo.APIKey,
		DefaultModel:             aiproviderConsts.Llama31,
		Origin:                   aiproviderConsts.LlamacppProviderInfo.Origin,
		ChatCompletionPathPrefix: aiproviderConsts.LlamacppProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			aiproviderConsts.Llama31: {
				DisplayName: aiproviderConsts.DisplayNameLlama31,
				IsEnabled:   true,
			},
			aiproviderConsts.Llama3: {
				DisplayName: aiproviderConsts.DisplayNameLlama3,
				IsEnabled:   true,
			},
		},
	},
	aiproviderConsts.ProviderNameOpenAI: {
		IsEnabled:                true,
		APIKey:                   aiproviderConsts.OpenAIProviderInfo.APIKey,
		DefaultModel:             aiproviderConsts.GPTO3Mini,
		Origin:                   aiproviderConsts.OpenAIProviderInfo.Origin,
		ChatCompletionPathPrefix: aiproviderConsts.LlamacppProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			aiproviderConsts.GPTO3Mini: {
				DisplayName: aiproviderConsts.DisplayNameGPTO3Mini,
				IsEnabled:   true,
			},
			aiproviderConsts.GPTO1: {
				DisplayName: aiproviderConsts.DisplayNameGPTO1,
				IsEnabled:   true,
			},
			aiproviderConsts.GPTO1Mini: {
				DisplayName: aiproviderConsts.DisplayNameGPTO1Mini,
				IsEnabled:   false,
			},
			aiproviderConsts.GPT45Preview: {
				DisplayName: aiproviderConsts.DisplayNameGPT45Preview,
				IsEnabled:   false,
			},
			aiproviderConsts.GPT41: {
				DisplayName: aiproviderConsts.DisplayNameGPT41,
				IsEnabled:   true,
			},
			aiproviderConsts.GPT41Mini: {
				DisplayName: aiproviderConsts.DisplayNameGPT41Mini,
				IsEnabled:   false,
			},
			aiproviderConsts.GPT4O: {
				DisplayName: aiproviderConsts.DisplayNameGPT4O,
				IsEnabled:   false,
			},
			aiproviderConsts.GPT4OMini: {
				DisplayName: aiproviderConsts.DisplayNameGPT4OMini,
				IsEnabled:   false,
			},
			aiproviderConsts.GPT4: {
				DisplayName: aiproviderConsts.DisplayNameGPT4,
				IsEnabled:   false,
			},
			aiproviderConsts.GPT35Turbo: {
				DisplayName: aiproviderConsts.DisplayNameGPT35Turbo,
				IsEnabled:   false,
			},
		},
	},
}

// Define the default settings data.
var DefaultSettingsData = SettingsSchema{
	Version: "1.0",
	App: AppSettings{
		DefaultProvider: aiproviderConsts.ProviderNameOpenAI,
	},
	AISettings: DefaultAISettings,
}
