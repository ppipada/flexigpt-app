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
		DefaultModel:             modelConsts.Claude4Sonnet,
		Origin:                   modelConsts.AnthropicProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.AnthropicProviderInfo.ChatCompletionPathPrefix,
	},

	modelConsts.ProviderNameDeepseek: {
		IsEnabled:                false,
		APIKey:                   modelConsts.DeepseekProviderInfo.APIKey,
		DefaultModel:             modelConsts.DeepseekChat,
		Origin:                   modelConsts.DeepseekProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.DeepseekProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameGoogle: {
		IsEnabled:                false,
		APIKey:                   modelConsts.GoogleProviderInfo.APIKey,
		DefaultModel:             modelConsts.Gemini25Flash,
		Origin:                   modelConsts.GoogleProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.GoogleProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameHuggingFace: {
		IsEnabled:                false,
		APIKey:                   modelConsts.HuggingfaceProviderInfo.APIKey,
		DefaultModel:             modelConsts.DeepseekCoder13BInstruct,
		Origin:                   modelConsts.HuggingfaceProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.HuggingfaceProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameLlamaCPP: {
		IsEnabled:                false,
		APIKey:                   modelConsts.LlamacppProviderInfo.APIKey,
		DefaultModel:             modelConsts.Llama4Scout,
		Origin:                   modelConsts.LlamacppProviderInfo.Origin,
		ChatCompletionPathPrefix: modelConsts.LlamacppProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameOpenAI: {
		IsEnabled:                true,
		APIKey:                   modelConsts.OpenAIProviderInfo.APIKey,
		DefaultModel:             modelConsts.GPT41,
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
