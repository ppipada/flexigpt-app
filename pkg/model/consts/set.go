package consts

import "github.com/ppipada/flexigpt-app/pkg/model/spec"

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

var InbuiltProviderModels = map[spec.ProviderName]spec.ProviderPreset{
	ProviderNameAnthropic:   AnthropicProviderPreset,
	ProviderNameDeepseek:    DeepseekProviderPreset,
	ProviderNameGoogle:      GoogleProviderPreset,
	ProviderNameHuggingFace: HuggingfaceProviderPreset,
	ProviderNameLlamaCPP:    LlamacppProviderPreset,
	ProviderNameOpenAI:      OpenAIProviderPreset,
}

var GlobalModelParamsDefault = spec.ModelParams{
	Name:                 "name",
	Stream:               true,
	MaxPromptLength:      2048,
	MaxOutputLength:      512,
	Temperature:          Float64Ptr(0.1),
	Reasoning:            nil,
	SystemPrompt:         "",
	Timeout:              60,
	AdditionalParameters: map[string]any{},
}
