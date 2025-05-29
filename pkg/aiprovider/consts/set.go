package consts

import "github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"

const (
	OpenAICompatibleAPIKeyHeaderKey          = "Authorization"
	OpenAICompatibleChatCompletionPathPrefix = "/v1/chat/completions"
)

var OpenAICompatibleDefaultHeaders = map[string]string{"content-type": "application/json"}

var InbuiltProviders = map[spec.ProviderName]spec.ProviderInfo{
	ProviderNameAnthropic:   AnthropicProviderInfo,
	ProviderNameDeepseek:    DeepseekProviderInfo,
	ProviderNameGoogle:      GoogleProviderInfo,
	ProviderNameHuggingFace: HuggingfaceProviderInfo,
	ProviderNameLlamaCPP:    LlamacppProviderInfo,
	ProviderNameOpenAI:      OpenAIProviderInfo,
}

var InbuiltProviderModels = map[spec.ProviderName]map[spec.ModelName]spec.ModelParams{
	ProviderNameAnthropic:   AnthropicModels,
	ProviderNameDeepseek:    DeepseekModels,
	ProviderNameGoogle:      GoogleModels,
	ProviderNameHuggingFace: HuggingfaceModels,
	ProviderNameLlamaCPP:    LlamacppModels,
	ProviderNameOpenAI:      OpenAIModels,
}

var InbuiltProviderModelDefaults = map[spec.ProviderName]map[spec.ModelName]spec.ModelDefaults{
	ProviderNameAnthropic:   AnthropicModelDefaults,
	ProviderNameDeepseek:    DeepseekModelDefaults,
	ProviderNameGoogle:      GoogleModelDefaults,
	ProviderNameHuggingFace: HuggingfaceModelDefaults,
	ProviderNameLlamaCPP:    LlamacppModelDefaults,
	ProviderNameOpenAI:      OpenAIModelDefaults,
}
