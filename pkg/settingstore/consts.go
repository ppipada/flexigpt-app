package settingstore

import (
	modelConsts "github.com/ppipada/flexigpt-app/pkg/model/consts"
	modelSpec "github.com/ppipada/flexigpt-app/pkg/model/spec"
)

// Define the default AI settings.
var DefaultAISettings = map[modelSpec.ProviderName]AISetting{
	modelConsts.ProviderNameAnthropic: {
		IsEnabled:                true,
		APIKey:                   modelConsts.AnthropicProviderInfo.APIKey,
		Origin:                   modelConsts.AnthropicProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.AnthropicProviderInfo.ChatCompletionPathPrefix,
	},

	modelConsts.ProviderNameDeepseek: {
		IsEnabled:                false,
		APIKey:                   modelConsts.DeepseekProviderInfo.APIKey,
		Origin:                   modelConsts.DeepseekProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.DeepseekProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameGoogle: {
		IsEnabled:                false,
		APIKey:                   modelConsts.GoogleProviderInfo.APIKey,
		Origin:                   modelConsts.GoogleProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.GoogleProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameHuggingFace: {
		IsEnabled:                false,
		APIKey:                   modelConsts.HuggingfaceProviderInfo.APIKey,
		Origin:                   modelConsts.HuggingfaceProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.HuggingfaceProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameLlamaCPP: {
		IsEnabled:                false,
		APIKey:                   modelConsts.LlamacppProviderInfo.APIKey,
		Origin:                   modelConsts.LlamacppProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.LlamacppProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameOpenAI: {
		IsEnabled:                true,
		APIKey:                   modelConsts.OpenAIProviderInfo.APIKey,
		Origin:                   modelConsts.OpenAIProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.LlamacppProviderInfo.ChatCompletionPathPrefix,
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
