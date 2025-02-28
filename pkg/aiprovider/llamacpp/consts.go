package llamacpp

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const ProviderNameLlamaCPP spec.ProviderName = "llamacpp"

const (
	Llama3  spec.ModelName = "llama3"
	Llama31 spec.ModelName = "llama3.1"
)

var LlamacppModels = map[spec.ModelName]spec.ModelInfo{
	Llama3: {
		Name:            Llama3,
		DisplayName:     "Llama 3",
		Provider:        ProviderNameLlamaCPP,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
	Llama31: {
		Name:            Llama31,
		DisplayName:     "Llama 3.1",
		Provider:        ProviderNameLlamaCPP,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
}

var LlamacppProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameLlamaCPP,
	APIKey:                   "",
	Engine:                   "",
	DefaultOrigin:            "http://127.0.0.1:8080",
	DefaultModel:             Llama3,
	AdditionalSettings:       map[string]interface{}{},
	Timeout:                  120,
	APIKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/completion",
	DefaultTemperature:       0.1,
	StreamingSupport:         false,
	Models:                   LlamacppModels,
}
