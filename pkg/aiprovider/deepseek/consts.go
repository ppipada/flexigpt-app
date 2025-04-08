package deepseek

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/openaicompat"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const ProviderNameDeepseek spec.ProviderName = "deepseek"

const (
	DeepseekChat     spec.ModelName = "deepseek-chat"
	DeepseekReasoner spec.ModelName = "deepseek-reasoner"
)

var DeepseekModels = map[spec.ModelName]spec.ModelInfo{
	DeepseekChat: {
		Name:                DeepseekChat,
		DisplayName:         "Deepseek Chat",
		Provider:            ProviderNameDeepseek,
		MaxPromptLength:     8192,
		MaxOutputLength:     8192,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    false,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
	DeepseekReasoner: {
		Name:                DeepseekChat,
		DisplayName:         "Deepseek Reasoner",
		Provider:            ProviderNameDeepseek,
		MaxPromptLength:     8192,
		MaxOutputLength:     8192,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    true,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
}

var DeepseekProviderInfo = spec.ProviderInfo{
	Name:         ProviderNameDeepseek,
	APIKey:       "",
	Origin:       "https://api.deepseek.com",
	DefaultModel: DeepseekChat,
	Type:         spec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          openaicompat.APIKeyHeaderKey,
	DefaultHeaders:           openaicompat.DefaultHeaders,
	ChatCompletionPathPrefix: openaicompat.ChatCompletionPathPrefix,
	Models:                   DeepseekModels,
}
