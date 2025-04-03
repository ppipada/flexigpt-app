package deepseek

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const ProviderNameDeepseek spec.ProviderName = "deepseek"

const (
	DeepseekChat     spec.ModelName = "deepseek-chat"
	DeepseekReasoner spec.ModelName = "deepseek-reasoner"
)

var DeepseekModels = map[spec.ModelName]spec.ModelInfo{
	DeepseekChat: {
		Name:            DeepseekChat,
		DisplayName:     "Deepseek Chat",
		Provider:        ProviderNameDeepseek,
		MaxPromptLength: 8192,
		MaxOutputLength: 8192,
	},
	DeepseekReasoner: {
		Name:            DeepseekChat,
		DisplayName:     "Deepseek Reasoner",
		Provider:        ProviderNameDeepseek,
		MaxPromptLength: 8192,
		MaxOutputLength: 8192,
	},
}

var DeepseekProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameDeepseek,
	APIKey:                   "",
	Engine:                   "",
	Origin:                   "https://api.deepseek.com",
	DefaultModel:             DeepseekChat,
	AdditionalSettings:       map[string]any{},
	Timeout:                  120,
	APIKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "",
	DefaultTemperature:       0.1,
	StreamingSupport:         true,
	Models:                   DeepseekModels,
}
