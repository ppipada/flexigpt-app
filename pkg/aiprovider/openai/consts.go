package openai

import "github.com/flexigpt/flexiui/pkg/aiprovider/spec"

const (
	ProviderNameOpenAI spec.ProviderName = "openai"
)

const (
	GPT_O1_PREVIEW spec.ModelName = "o1-preview"
	GPT_O1_MINI    spec.ModelName = "o1-mini"
	GPT_4O_MINI    spec.ModelName = "gpt-4o-mini"
	GPT_4O         spec.ModelName = "gpt-4o"
	GPT_4          spec.ModelName = "gpt-4"
	GPT_3_5_TURBO  spec.ModelName = "gpt-3.5-turbo"
)

var OpenAIModels = map[spec.ModelName]spec.ModelInfo{
	GPT_O1_PREVIEW: {
		Name:               GPT_O1_PREVIEW,
		DisplayName:        "OpenAI GPT o1 preview",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 1,
	},
	GPT_O1_MINI: {
		Name:               GPT_O1_MINI,
		DisplayName:        "OpenAI GPT o1 mini",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 1,
	},
	GPT_4O: {
		Name:               GPT_4O,
		DisplayName:        "OpenAI GPT 4o",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	GPT_4: {
		Name:               GPT_4,
		DisplayName:        "OpenAI GPT 4",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	GPT_3_5_TURBO: {
		Name:               GPT_3_5_TURBO,
		DisplayName:        "OpenAI GPT 3.5 turbo",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    2400,
		MaxOutputLength:    2400,
		DefaultTemperature: 0.1,
	},
	GPT_4O_MINI: {
		Name:               GPT_4O_MINI,
		DisplayName:        "OpenAI GPT 4o mini",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
}

var OpenAIProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameOpenAI,
	ApiKey:                   "",
	Engine:                   "",
	DefaultOrigin:            "https://api.openai.com/v1",
	DefaultModel:             GPT_4O_MINI,
	AdditionalSettings:       map[string]interface{}{},
	Timeout:                  120,
	ApiKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/chat/completions",
	DefaultTemperature:       0.1,
	StreamingSupport:         true,
}
