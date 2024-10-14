package huggingface

import "github.com/flexigpt/flexiui/pkg/aiprovider/spec"

const ProviderNameHuggingFace spec.ProviderName = "huggingface"

const (
	DEEPSEEK_CODER_1_3B_INSTRUCT spec.ModelName = "deepseek-ai/deepseek-coder-1.3b-instruct"
)

var HuggingfaceModels = map[spec.ModelName]spec.ModelInfo{
	DEEPSEEK_CODER_1_3B_INSTRUCT: {
		Name:               DEEPSEEK_CODER_1_3B_INSTRUCT,
		DisplayName:        "HF Deepseek Coder 1.3b",
		Provider:           ProviderNameHuggingFace,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
}

var HuggingfaceProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameHuggingFace,
	ApiKey:                   "",
	Engine:                   "",
	DefaultOrigin:            "https://api-inference.huggingface.co",
	DefaultModel:             DEEPSEEK_CODER_1_3B_INSTRUCT,
	AdditionalSettings:       map[string]interface{}{},
	Timeout:                  120,
	ApiKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/models",
	DefaultTemperature:       0.1,
	ModelPrefixes: []string{
		"microsoft/",
		"replit/",
		"Salesforce/",
		"bigcode/",
		"deepseek-ai/",
	},
	StreamingSupport: false,
	Descriptions: map[string]string{
		"apiKey":                   "Your huggingface API key.",
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
