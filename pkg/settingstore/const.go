package settingstore

import (
	inferenceSpec "github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetConsts "github.com/ppipada/flexigpt-app/pkg/modelpreset/consts"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

// Define the default AI settings.
var DefaultAISettings = map[modelpresetSpec.ProviderName]AISetting{
	modelpresetConsts.ProviderNameAnthropic: {
		IsEnabled:                true,
		APIKey:                   inferenceSpec.AnthropicProviderInfo.APIKey,
		Origin:                   inferenceSpec.AnthropicProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.AnthropicProviderInfo.ChatCompletionPathPrefix,
	},

	modelpresetConsts.ProviderNameDeepseek: {
		IsEnabled:                false,
		APIKey:                   inferenceSpec.DeepseekProviderInfo.APIKey,
		Origin:                   inferenceSpec.DeepseekProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.DeepseekProviderInfo.ChatCompletionPathPrefix,
	},
	modelpresetConsts.ProviderNameGoogle: {
		IsEnabled:                false,
		APIKey:                   inferenceSpec.GoogleProviderInfo.APIKey,
		Origin:                   inferenceSpec.GoogleProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.GoogleProviderInfo.ChatCompletionPathPrefix,
	},
	modelpresetConsts.ProviderNameHuggingFace: {
		IsEnabled:                false,
		APIKey:                   inferenceSpec.HuggingfaceProviderInfo.APIKey,
		Origin:                   inferenceSpec.HuggingfaceProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.HuggingfaceProviderInfo.ChatCompletionPathPrefix,
	},
	modelpresetConsts.ProviderNameLlamaCPP: {
		IsEnabled:                false,
		APIKey:                   inferenceSpec.LlamacppProviderInfo.APIKey,
		Origin:                   inferenceSpec.LlamacppProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.LlamacppProviderInfo.ChatCompletionPathPrefix,
	},
	modelpresetConsts.ProviderNameOpenAI: {
		IsEnabled:                true,
		APIKey:                   inferenceSpec.OpenAIProviderInfo.APIKey,
		Origin:                   inferenceSpec.OpenAIProviderInfo.Origin,
		ChatCompletionPathPrefix: inferenceSpec.LlamacppProviderInfo.ChatCompletionPathPrefix,
	},
}

var DefaultSettingsData = SettingsSchema{
	Version: "1.0",
	App: AppSettings{
		DefaultProvider: modelpresetConsts.ProviderNameOpenAI,
	},
	AISettings: DefaultAISettings,
}
