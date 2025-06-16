package spec

import (
	modelConsts "github.com/ppipada/flexigpt-app/pkg/model/consts"
	modelSpec "github.com/ppipada/flexigpt-app/pkg/model/spec"
)

// Define the default AI settings.
var DefaultAISettings = map[modelSpec.ProviderName]AISetting{
	modelConsts.ProviderNameAnthropic: {
		IsEnabled:                true,
		APIKey:                   modelConsts.AnthropicProviderInfo.APIKey,
		DefaultModel:             modelConsts.Claude4Sonnet,
		Origin:                   modelConsts.AnthropicProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.AnthropicProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[modelSpec.ModelName]ModelSetting{
			modelConsts.Claude4Sonnet: {
				DisplayName: modelConsts.DisplayNameClaude4Sonnet,
				IsEnabled:   true,
			},
		},
	},

	modelConsts.ProviderNameDeepseek: {
		IsEnabled:                false,
		APIKey:                   modelConsts.DeepseekProviderInfo.APIKey,
		DefaultModel:             modelConsts.DeepseekChat,
		Origin:                   modelConsts.DeepseekProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.DeepseekProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[modelSpec.ModelName]ModelSetting{
			modelConsts.DeepseekChat: {
				DisplayName: modelConsts.DisplayNameDeepseekChat,
				IsEnabled:   true,
			},
		},
	},
	modelConsts.ProviderNameGoogle: {
		IsEnabled:                false,
		APIKey:                   modelConsts.GoogleProviderInfo.APIKey,
		DefaultModel:             modelConsts.Gemini25Flash,
		Origin:                   modelConsts.GoogleProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.GoogleProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[modelSpec.ModelName]ModelSetting{
			modelConsts.Gemini25Flash: {
				DisplayName: modelConsts.DisplayNameGemini25Flash,
				IsEnabled:   true,
			},
		},
	},
	modelConsts.ProviderNameHuggingFace: {
		IsEnabled:                false,
		APIKey:                   modelConsts.HuggingfaceProviderInfo.APIKey,
		DefaultModel:             modelConsts.DeepseekCoder13BInstruct,
		Origin:                   modelConsts.HuggingfaceProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.HuggingfaceProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[modelSpec.ModelName]ModelSetting{
			modelConsts.DeepseekCoder13BInstruct: {
				DisplayName: modelConsts.DisplayNameDeepseekCoder13BInstruct,
				IsEnabled:   true,
			},
		},
	},
	modelConsts.ProviderNameLlamaCPP: {
		IsEnabled:                false,
		APIKey:                   modelConsts.LlamacppProviderInfo.APIKey,
		DefaultModel:             modelConsts.Llama31,
		Origin:                   modelConsts.LlamacppProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.LlamacppProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[modelSpec.ModelName]ModelSetting{
			modelConsts.Llama31: {
				DisplayName: modelConsts.DisplayNameLlama31,
				IsEnabled:   true,
			},
		},
	},
	modelConsts.ProviderNameOpenAI: {
		IsEnabled:                true,
		APIKey:                   modelConsts.OpenAIProviderInfo.APIKey,
		DefaultModel:             modelConsts.GPT41,
		Origin:                   modelConsts.OpenAIProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.LlamacppProviderInfo.ChatCompletionPathPrefix,
		ModelSettings: map[modelSpec.ModelName]ModelSetting{
			modelConsts.GPT41: {
				DisplayName: modelConsts.DisplayNameGPT41,
				IsEnabled:   true,
			},
		},
	},
}

// Define the default settings data.
var DefaultSettingsData = SettingsSchema{
	Version: "1.0",
	App: AppSettings{
		DefaultProvider: modelConsts.ProviderNameOpenAI,
	},
	AISettings: DefaultAISettings,
}
