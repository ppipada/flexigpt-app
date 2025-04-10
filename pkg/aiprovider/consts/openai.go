package consts

import (
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
	GPT45Preview spec.ModelName = "gpt-4-5-preview"
	GPT4OMini    spec.ModelName = "gpt-4o-mini"
	GPT4O        spec.ModelName = "gpt-4o"
	GPT4         spec.ModelName = "gpt-4"
	GPT35Turbo   spec.ModelName = "gpt-3-5-turbo"
)

const (
	DisplayNameGPTO3Mini    = "OpenAI o3 mini"
	DisplayNameGPTO1        = "OpenAI o1"
	DisplayNameGPTO1Preview = "OpenAI o1 preview"
	DisplayNameGPTO1Mini    = "OpenAI o1 mini"
	DisplayNameGPT45Preview = "OpenAI GPT 4.5 Preview"
	DisplayNameGPT4OMini    = "OpenAI GPT 4o mini"
	DisplayNameGPT4O        = "OpenAI GPT 4o"
	DisplayNameGPT4         = "OpenAI GPT 4"
	DisplayNameGPT35Turbo   = "OpenAI GPT 3.5 turbo"
)

var OpenAIModels = map[spec.ModelName]spec.ModelParams{
	GPTO3Mini: {
		Name:             GPTO3Mini,
		Stream:           true,
		MaxPromptLength:  16384,
		MaxOutputLength:  16384,
		Temperature:      spec.Float64Ptr(1.0),
		ReasoningSupport: true,
		SystemPrompt:     "Formatting re-enabled",
		Timeout:          120,
	},
	GPTO1: {
		Name:             GPTO1,
		MaxPromptLength:  16384,
		MaxOutputLength:  16384,
		Temperature:      spec.Float64Ptr(1.0),
		Stream:           true,
		ReasoningSupport: true,
		SystemPrompt:     "Formatting re-enabled",
		Timeout:          120,
	},
	GPTO1Preview: {
		Name:             GPTO1Preview,
		MaxPromptLength:  16384,
		MaxOutputLength:  16384,
		Temperature:      spec.Float64Ptr(1.0),
		Stream:           true,
		ReasoningSupport: true,
		SystemPrompt:     "",
		Timeout:          120,
	},
	GPTO1Mini: {
		Name:             GPTO1Mini,
		MaxPromptLength:  16384,
		MaxOutputLength:  16384,
		Temperature:      spec.Float64Ptr(1.0),
		Stream:           true,
		ReasoningSupport: true,
		SystemPrompt:     "",
		Timeout:          120,
	},
	GPT45Preview: {
		Name:             GPT45Preview,
		MaxPromptLength:  16384,
		MaxOutputLength:  16384,
		Temperature:      spec.Float64Ptr(0.1),
		Stream:           true,
		ReasoningSupport: false,
		SystemPrompt:     "",
		Timeout:          120,
	},
	GPT4O: {
		Name:             GPT4O,
		MaxPromptLength:  16384,
		MaxOutputLength:  16384,
		Temperature:      spec.Float64Ptr(0.1),
		Stream:           true,
		ReasoningSupport: false,
		SystemPrompt:     "",
		Timeout:          120,
	},
	GPT4: {
		Name:             GPT4,
		MaxPromptLength:  4096,
		MaxOutputLength:  4096,
		Temperature:      spec.Float64Ptr(0.1),
		Stream:           true,
		ReasoningSupport: false,
		SystemPrompt:     "",
		Timeout:          120,
	},
	GPT35Turbo: {
		Name:             GPT35Turbo,
		MaxPromptLength:  2400,
		MaxOutputLength:  2400,
		Temperature:      spec.Float64Ptr(0.1),
		Stream:           true,
		ReasoningSupport: false,
		SystemPrompt:     "",
		Timeout:          120,
	},
	GPT4OMini: {
		Name:             GPT4OMini,
		MaxPromptLength:  4096,
		MaxOutputLength:  4096,
		Temperature:      spec.Float64Ptr(0.1),
		Stream:           true,
		ReasoningSupport: false,
		SystemPrompt:     "",
		Timeout:          120,
	},
}

var OpenAIProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameOpenAI,
	APIKey:                   "",
	Origin:                   "https://api.openai.com",
	Type:                     spec.InbuiltOpenAICompatible,
	APIKeyHeaderKey:          OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: OpenAICompatibleChatCompletionPathPrefix,
}
