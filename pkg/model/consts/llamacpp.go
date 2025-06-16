package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

const ProviderNameLlamaCPP spec.ProviderName = "llamacpp"

const (
	Llama3  spec.ModelName = "llama3"
	Llama31 spec.ModelName = "llama3.1"
)

const (
	DisplayNameLlama3  = "Llama 3"
	DisplayNameLlama31 = "Llama 3.1"
)

var LlamacppModelDefaults = map[spec.ModelName]spec.ModelDefaults{
	Llama31: {
		DisplayName: DisplayNameLlama31,
		IsEnabled:   true,
	},
	Llama3: {
		DisplayName: DisplayNameLlama3,
		IsEnabled:   true,
	},
}

var LlamacppModels = map[spec.ModelName]spec.ModelParams{
	Llama3: {
		Name:            Llama3,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Llama31: {
		Name:            Llama31,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
}

var LlamacppProviderInfo = spec.ProviderInfo{
	Name:   ProviderNameLlamaCPP,
	APIKey: "",
	Origin: "http://127.0.0.1:8080",
	Type:   spec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: OpenAICompatibleChatCompletionPathPrefix,
}
