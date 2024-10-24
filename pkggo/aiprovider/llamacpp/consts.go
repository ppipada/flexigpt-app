package llamacpp

import (
	"github.com/flexigpt/flexiui/pkggo/aiprovider/spec"
)

const ProviderNameLlamaCPP spec.ProviderName = "llamacpp"

const (
	LLAMA_3   spec.ModelName = "llama3"
	LLAMA_3_1 spec.ModelName = "llama3.1"
)

var LlamacppModels = map[spec.ModelName]spec.ModelInfo{
	LLAMA_3: {
		Name:            LLAMA_3,
		DisplayName:     "Llama 3",
		Provider:        ProviderNameLlamaCPP,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
	LLAMA_3_1: {
		Name:            LLAMA_3_1,
		DisplayName:     "Llama 3.1",
		Provider:        ProviderNameLlamaCPP,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
}

var LlamacppProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameLlamaCPP,
	ApiKey:                   "",
	Engine:                   "",
	DefaultOrigin:            "http://127.0.0.1:8080",
	DefaultModel:             LLAMA_3,
	AdditionalSettings:       map[string]interface{}{},
	Timeout:                  120,
	ApiKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/completion",
	DefaultTemperature:       0.1,
	StreamingSupport:         false,
}
