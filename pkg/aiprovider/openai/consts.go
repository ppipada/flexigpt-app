package openai

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/baseutils"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const (
	ProviderNameOpenAI spec.ProviderName = "openai"
)

const (
	GPTO3Mini    spec.ModelName = "o3-mini"
	GPTO1        spec.ModelName = "o1"
	GPTO1Preview spec.ModelName = "o1-preview"
	GPTO1Mini    spec.ModelName = "o1-mini"
	GPT45Preview spec.ModelName = "gpt-4.5-preview"
	GPT4OMini    spec.ModelName = "gpt-4o-mini"
	GPT4O        spec.ModelName = "gpt-4o"
	GPT4         spec.ModelName = "gpt-4"
	GPT35Turbo   spec.ModelName = "gpt-3.5-turbo"
)

var OpenAIModels = map[spec.ModelName]spec.ModelInfo{
	GPTO3Mini: {
		Name:                GPTO3Mini,
		DisplayName:         "OpenAI o3 mini",
		Provider:            ProviderNameOpenAI,
		MaxPromptLength:     8192,
		MaxOutputLength:     8192,
		DefaultTemperature:  baseutils.Float64Ptr(1),
		StreamingSupport:    baseutils.BoolPtr(false),
		DefaultSystemPrompt: "Formatting re-enabled",
	},
	GPTO1: {
		Name:                GPTO1,
		DisplayName:         "OpenAI o1",
		Provider:            ProviderNameOpenAI,
		MaxPromptLength:     8192,
		MaxOutputLength:     8192,
		DefaultTemperature:  baseutils.Float64Ptr(1),
		StreamingSupport:    baseutils.BoolPtr(false),
		DefaultSystemPrompt: "Formatting re-enabled",
	},
	GPTO1Preview: {
		Name:               GPTO1Preview,
		DisplayName:        "OpenAI o1 preview",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    8192,
		MaxOutputLength:    8192,
		DefaultTemperature: baseutils.Float64Ptr(1),
		StreamingSupport:   baseutils.BoolPtr(false),
	},
	GPTO1Mini: {
		Name:               GPTO1Mini,
		DisplayName:        "OpenAI o1 mini",
		Provider:           ProviderNameOpenAI,
		MaxPromptLength:    8192,
		MaxOutputLength:    8192,
		DefaultTemperature: baseutils.Float64Ptr(1),
		StreamingSupport:   baseutils.BoolPtr(false),
	},
	GPT45Preview: {
		Name:            GPT45Preview,
		DisplayName:     "OpenAI GPT 4.5 Preview",
		Provider:        ProviderNameOpenAI,
		MaxPromptLength: 8192,
		MaxOutputLength: 8192,
	},
	GPT4O: {
		Name:            GPT4O,
		DisplayName:     "OpenAI GPT 4o",
		Provider:        ProviderNameOpenAI,
		MaxPromptLength: 8192,
		MaxOutputLength: 8192,
	},
	GPT4: {
		Name:            GPT4,
		DisplayName:     "OpenAI GPT 4",
		Provider:        ProviderNameOpenAI,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
	GPT35Turbo: {
		Name:            GPT35Turbo,
		DisplayName:     "OpenAI GPT 3.5 turbo",
		Provider:        ProviderNameOpenAI,
		MaxPromptLength: 2400,
		MaxOutputLength: 2400,
	},
	GPT4OMini: {
		Name:            GPT4OMini,
		DisplayName:     "OpenAI GPT 4o mini",
		Provider:        ProviderNameOpenAI,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
}

var OpenAIProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameOpenAI,
	APIKey:                   "",
	Engine:                   "",
	Origin:                   "https://api.openai.com/v1",
	DefaultModel:             GPTO3Mini,
	AdditionalSettings:       map[string]any{},
	Timeout:                  120,
	APIKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/chat/completions",
	DefaultTemperature:       0.0,
	StreamingSupport:         true,
	Models:                   OpenAIModels,
}
