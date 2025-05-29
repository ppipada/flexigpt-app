package spec

import (
	aiproviderConsts "github.com/ppipada/flexigpt-app/pkg/aiprovider/consts"
	aiproviderSpec "github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"
)

// Define the default AI settings.
var DefaultAISettings = map[aiproviderSpec.ProviderName]AISetting{
	aiproviderConsts.ProviderNameAnthropic: {
		IsEnabled:                true,
		APIKey:                   aiproviderConsts.AnthropicProviderInfo.APIKey,
		DefaultModel:             aiproviderConsts.Claude4Sonnet,
		Origin:                   aiproviderConsts.AnthropicProviderInfo.Origin,
		ChatCompletionPathPrefix: aiproviderConsts.AnthropicProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			aiproviderConsts.Claude4Sonnet: {
				DisplayName: aiproviderConsts.DisplayNameClaude4Sonnet,
				IsEnabled:   true,
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
		},
	},
	aiproviderConsts.ProviderNameGoogle: {
		IsEnabled:                false,
		APIKey:                   aiproviderConsts.GoogleProviderInfo.APIKey,
		DefaultModel:             aiproviderConsts.Gemini25Flash,
		Origin:                   aiproviderConsts.GoogleProviderInfo.Origin,
		ChatCompletionPathPrefix: aiproviderConsts.GoogleProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			aiproviderConsts.Gemini25Flash: {
				DisplayName: aiproviderConsts.DisplayNameGemini25Flash,
				IsEnabled:   true,
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
		},
	},
	aiproviderConsts.ProviderNameOpenAI: {
		IsEnabled:                true,
		APIKey:                   aiproviderConsts.OpenAIProviderInfo.APIKey,
		DefaultModel:             aiproviderConsts.GPT41,
		Origin:                   aiproviderConsts.OpenAIProviderInfo.Origin,
		ChatCompletionPathPrefix: aiproviderConsts.LlamacppProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[aiproviderSpec.ModelName]ModelSetting{
			aiproviderConsts.GPT41: {
				DisplayName: aiproviderConsts.DisplayNameGPT41,
				IsEnabled:   true,
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
