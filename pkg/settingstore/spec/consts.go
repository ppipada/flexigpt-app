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
		ModelPresets: map[modelSpec.ModelName]modelSpec.ModelPreset{
			modelConsts.Claude4Sonnet: {
				Name:         modelConsts.Claude4Sonnet,
				DisplayName:  modelConsts.DisplayNameClaude4Sonnet,
				IsEnabled:    true,
				ShortCommand: modelConsts.ShortCommandClaude4Sonnet,
			},
		},
	},

	modelConsts.ProviderNameDeepseek: {
		IsEnabled:                false,
		APIKey:                   modelConsts.DeepseekProviderInfo.APIKey,
		DefaultModel:             modelConsts.DeepseekChat,
		Origin:                   modelConsts.DeepseekProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.DeepseekProviderInfo.ChatCompletionPathPrefix,
		ModelPresets: map[modelSpec.ModelName]modelSpec.ModelPreset{
			modelConsts.DeepseekChat: {
				Name:         modelConsts.DeepseekChat,
				DisplayName:  modelConsts.DisplayNameDeepseekChat,
				IsEnabled:    true,
				ShortCommand: modelConsts.ShortCommandDeepseekChat,
			},
		},
	},
	modelConsts.ProviderNameGoogle: {
		IsEnabled:                false,
		APIKey:                   modelConsts.GoogleProviderInfo.APIKey,
		DefaultModel:             modelConsts.Gemini25Flash,
		Origin:                   modelConsts.GoogleProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.GoogleProviderInfo.ChatCompletionPathPrefix,
		ModelPresets: map[modelSpec.ModelName]modelSpec.ModelPreset{
			modelConsts.Gemini25Flash: {
				Name:         modelConsts.Gemini25Flash,
				DisplayName:  modelConsts.DisplayNameGemini25Flash,
				IsEnabled:    true,
				ShortCommand: modelConsts.ShortCommandGemini25Flash,
			},
		},
	},
	modelConsts.ProviderNameHuggingFace: {
		IsEnabled:                false,
		APIKey:                   modelConsts.HuggingfaceProviderInfo.APIKey,
		DefaultModel:             modelConsts.DeepseekCoder13BInstruct,
		Origin:                   modelConsts.HuggingfaceProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.HuggingfaceProviderInfo.ChatCompletionPathPrefix,
		ModelPresets: map[modelSpec.ModelName]modelSpec.ModelPreset{
			modelConsts.DeepseekCoder13BInstruct: {
				Name:         modelConsts.DeepseekCoder13BInstruct,
				DisplayName:  modelConsts.DisplayNameDeepseekCoder13BInstruct,
				IsEnabled:    true,
				ShortCommand: modelConsts.ShortCommandDeepseekCoder13BInstruct,
			},
		},
	},
	modelConsts.ProviderNameLlamaCPP: {
		IsEnabled:                false,
		APIKey:                   modelConsts.LlamacppProviderInfo.APIKey,
		DefaultModel:             modelConsts.Llama4Scout,
		Origin:                   modelConsts.LlamacppProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.LlamacppProviderInfo.ChatCompletionPathPrefix,
		ModelPresets: map[modelSpec.ModelName]modelSpec.ModelPreset{
			modelConsts.Llama4Scout: {
				Name:         modelConsts.Llama4Scout,
				DisplayName:  modelConsts.DisplayNameLlama4Scout,
				IsEnabled:    true,
				ShortCommand: modelConsts.ShortCommandLlama4Scout,
			},
		},
	},
	modelConsts.ProviderNameOpenAI: {
		IsEnabled:                true,
		APIKey:                   modelConsts.OpenAIProviderInfo.APIKey,
		DefaultModel:             modelConsts.GPT41,
		Origin:                   modelConsts.OpenAIProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.LlamacppProviderInfo.ChatCompletionPathPrefix,
		ModelPresets: map[modelSpec.ModelName]modelSpec.ModelPreset{
			modelConsts.GPT41: {
				Name:         modelConsts.GPT41,
				DisplayName:  modelConsts.DisplayNameGPT41,
				IsEnabled:    true,
				ShortCommand: modelConsts.ShortCommandGPT41,
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
