package consts

import "github.com/ppipada/flexigpt-app/pkg/model/spec"

const (
	OpenAICompatibleAPIKeyHeaderKey          = "Authorization"
	OpenAICompatibleChatCompletionPathPrefix = "/v1/chat/completions"
)

var OpenAICompatibleDefaultHeaders = map[string]string{"content-type": "application/json"}

var InbuiltProviderModels = map[spec.ProviderName]spec.ProviderPreset{
	ProviderNameAnthropic:   AnthropicProviderPreset,
	ProviderNameDeepseek:    DeepseekProviderPreset,
	ProviderNameGoogle:      GoogleProviderPreset,
	ProviderNameHuggingFace: HuggingfaceProviderPreset,
	ProviderNameLlamaCPP:    LlamacppProviderPreset,
	ProviderNameOpenAI:      OpenAIProviderPreset,
}
