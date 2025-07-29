package settingstore

import (
	inferenceSpec "github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelConsts "github.com/ppipada/flexigpt-app/pkg/model/consts"
	modelSpec "github.com/ppipada/flexigpt-app/pkg/model/spec"
)

// Define the default AI settings.
var DefaultAISettings = map[modelSpec.ProviderName]AISetting{
	modelConsts.ProviderNameAnthropic: {
		IsEnabled:                true,
		APIKey:                   inferenceSpec.AnthropicProviderInfo.APIKey,
		Origin:                   inferenceSpec.AnthropicProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.AnthropicProviderInfo.ChatCompletionPathPrefix,
	},

	modelConsts.ProviderNameDeepseek: {
		IsEnabled:                false,
		APIKey:                   inferenceSpec.DeepseekProviderInfo.APIKey,
		Origin:                   inferenceSpec.DeepseekProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.DeepseekProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameGoogle: {
		IsEnabled:                false,
		APIKey:                   inferenceSpec.GoogleProviderInfo.APIKey,
		Origin:                   inferenceSpec.GoogleProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.GoogleProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameHuggingFace: {
		IsEnabled:                false,
		APIKey:                   inferenceSpec.HuggingfaceProviderInfo.APIKey,
		Origin:                   inferenceSpec.HuggingfaceProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.HuggingfaceProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameLlamaCPP: {
		IsEnabled:                false,
		APIKey:                   inferenceSpec.LlamacppProviderInfo.APIKey,
		Origin:                   inferenceSpec.LlamacppProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.LlamacppProviderInfo.ChatCompletionPathPrefix,
	},
	modelConsts.ProviderNameOpenAI: {
		IsEnabled:                true,
		APIKey:                   inferenceSpec.OpenAIProviderInfo.APIKey,
		Origin:                   inferenceSpec.OpenAIProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.LlamacppProviderInfo.ChatCompletionPathPrefix,
	},
}

var DefaultSettingsData = SettingsSchema{
	Version: "1.0",
	App: AppSettings{
		DefaultProvider: modelConsts.ProviderNameOpenAI,
	},
	AISettings: DefaultAISettings,
}
