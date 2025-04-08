package llamacpp

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/openaicompat"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const ProviderNameLlamaCPP spec.ProviderName = "llamacpp"

const (
	Llama3  spec.ModelName = "llama3"
	Llama31 spec.ModelName = "llama3.1"
)

var LlamacppModels = map[spec.ModelName]spec.ModelInfo{
	Llama3: {
		Name:                Llama3,
		DisplayName:         "Llama 3",
		Provider:            ProviderNameLlamaCPP,
		MaxPromptLength:     4096,
		MaxOutputLength:     4096,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    false,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
	Llama31: {
		Name:                Llama31,
		DisplayName:         "Llama 3.1",
		Provider:            ProviderNameLlamaCPP,
		MaxPromptLength:     4096,
		MaxOutputLength:     4096,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    false,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
}

var LlamacppProviderInfo = spec.ProviderInfo{
	Name:         ProviderNameLlamaCPP,
	APIKey:       "",
	Engine:       "",
	Origin:       "http://127.0.0.1:8080",
	DefaultModel: Llama3,
	Type:         spec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          openaicompat.APIKeyHeaderKey,
	DefaultHeaders:           openaicompat.DefaultHeaders,
	ChatCompletionPathPrefix: openaicompat.ChatCompletionPathPrefix,
	Models:                   LlamacppModels,
}
