package llamacpp

import "github.com/flexigpt/flexiui/pkg/aiprovider/spec"

const ProviderNameLlamaCPP spec.ProviderName = "llamacpp"

const (
	LLAMA_3   spec.ModelName = "llama3"
	LLAMA_3_1 spec.ModelName = "llama3.1"
)

var LlamacppModels = map[spec.ModelName]spec.ModelInfo{
	LLAMA_3: {
		Name:               LLAMA_3,
		DisplayName:        "Llama 3",
		Provider:           ProviderNameLlamaCPP,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	LLAMA_3_1: {
		Name:               LLAMA_3_1,
		DisplayName:        "Llama 3.1",
		Provider:           ProviderNameLlamaCPP,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
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
	Descriptions: map[string]string{
		"apiKey":                   "Your llamacpp API key.",
		"engine":                   "The engine to be used for processing.",
		"defaultOrigin":            "Default origin to use for requests. This can be used to talk to any server that serves a compatible API",
		"defaultModel":             "Default model to use for chat requests",
		"additionalSettings":       "Any additional settings to pass to the model. Input as a JSON object",
		"timeout":                  "The timeout duration in milliseconds.",
		"apiKeyHeaderKey":          "The header key for the API key.",
		"defaultHeaders":           "The default headers to be included in requests.",
		"chatCompletionPathPrefix": "The path prefix for chat completions.",
		"defaultTemperature":       "Default temperature setting for chat requests",
		"modelPrefixes":            "Optional prefixes for models.",
	},
}
