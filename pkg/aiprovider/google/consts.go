package google

import "github.com/flexigpt/flexiui/pkg/aiprovider/spec"

const ProviderNameGoogle spec.ProviderName = "google"

const (
	GEMINI_1_5_FLASH spec.ModelName = "gemini-1.5-flash"
	GEMINI_1_5_PRO   spec.ModelName = "gemini-1.5-pro"
)

var GoogleModels = map[spec.ModelName]spec.ModelInfo{
	GEMINI_1_5_FLASH: {
		Name:               GEMINI_1_5_FLASH,
		DisplayName:        "Google Gemini 1.5 Flash",
		Provider:           ProviderNameGoogle,
		MaxPromptLength:    4096,
		MaxOutputLength:    8192,
		DefaultTemperature: 0.1,
	},
	GEMINI_1_5_PRO: {
		Name:               GEMINI_1_5_PRO,
		DisplayName:        "Google Gemini 1.5 Pro",
		Provider:           ProviderNameGoogle,
		MaxPromptLength:    4096,
		MaxOutputLength:    8192,
		DefaultTemperature: 0.1,
	},
}

var GoogleProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameGoogle,
	ApiKey:                   "",
	Engine:                   "",
	DefaultOrigin:            "https://generativelanguage.googleapis.com",
	DefaultModel:             GEMINI_1_5_FLASH,
	AdditionalSettings:       map[string]interface{}{},
	Timeout:                  120,
	ApiKeyHeaderKey:          "x-goog-api-key",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/v1beta/models",
	DefaultTemperature:       0.1,
	StreamingSupport:         true,
	Descriptions: map[string]string{
		"apiKey":                   "Your google generative AI API key.",
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
