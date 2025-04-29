package consts

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const (
	ProviderNameOpenAI spec.ProviderName = "openai"
)

const (
	GPTO4Mini  spec.ModelName = "o4-mini"
	GPTO3      spec.ModelName = "o3"
	GPTO3Mini  spec.ModelName = "o3-mini"
	GPTO1      spec.ModelName = "o1"
	GPT41      spec.ModelName = "gpt-4.1"
	GPT41Mini  spec.ModelName = "gpt-4.1-mini"
	GPT4OMini  spec.ModelName = "gpt-4o-mini"
	GPT4O      spec.ModelName = "gpt-4o"
	GPT4       spec.ModelName = "gpt-4"
	GPT35Turbo spec.ModelName = "gpt-3.5-turbo"
)

const (
	DisplayNameGPTO4Mini  = "OpenAI o4 mini"
	DisplayNameGPTO3      = "OpenAI o3"
	DisplayNameGPTO3Mini  = "OpenAI o3 mini"
	DisplayNameGPTO1      = "OpenAI o1"
	DisplayNameGPTO1Mini  = "OpenAI o1 mini"
	DisplayNameGPT41      = "OpenAI GPT 4.1"
	DisplayNameGPT41Mini  = "OpenAI GPT 4.1 mini"
	DisplayNameGPT4OMini  = "OpenAI GPT 4o mini"
	DisplayNameGPT4O      = "OpenAI GPT 4o"
	DisplayNameGPT4       = "OpenAI GPT 4"
	DisplayNameGPT35Turbo = "OpenAI GPT 3.5 turbo"
)

var OpenAIModels = map[spec.ModelName]spec.ModelParams{
	GPTO4Mini: {
		Name:            GPTO4Mini,
		Stream:          true,
		MaxPromptLength: 32768,
		MaxOutputLength: 32768,
		Temperature:     spec.Float64Ptr(1.0),
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: "Formatting re-enabled",
		Timeout:      120,
	},
	GPTO3: {
		Name:            GPTO3,
		MaxPromptLength: 32768,
		MaxOutputLength: 32768,
		Temperature:     spec.Float64Ptr(1.0),
		Stream:          true,
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: "Formatting re-enabled",
		Timeout:      120,
	},
	GPTO3Mini: {
		Name:            GPTO3Mini,
		Stream:          true,
		MaxPromptLength: 16384,
		MaxOutputLength: 16384,
		Temperature:     spec.Float64Ptr(1.0),
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: "Formatting re-enabled",
		Timeout:      120,
	},
	GPTO1: {
		Name:            GPTO1,
		MaxPromptLength: 16384,
		MaxOutputLength: 16384,
		Temperature:     spec.Float64Ptr(1.0),
		Stream:          true,
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: "Formatting re-enabled",
		Timeout:      120,
	},
	GPT41: {
		Name:            GPT41,
		MaxPromptLength: 32768,
		MaxOutputLength: 32768,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	GPT41Mini: {
		Name:            GPT41Mini,
		MaxPromptLength: 32768,
		MaxOutputLength: 32768,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	GPT4O: {
		Name:            GPT4O,
		MaxPromptLength: 16384,
		MaxOutputLength: 16384,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	GPT4: {
		Name:            GPT4,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	GPT35Turbo: {
		Name:            GPT35Turbo,
		MaxPromptLength: 2400,
		MaxOutputLength: 2400,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	GPT4OMini: {
		Name:            GPT4OMini,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
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
