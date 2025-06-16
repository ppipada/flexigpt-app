package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

const ProviderNameDeepseek spec.ProviderName = "deepseek"

const (
	DeepseekChat     spec.ModelName = "deepseek-chat"
	DeepseekReasoner spec.ModelName = "deepseek-reasoner"
)

const (
	DisplayNameDeepseekChat     = "Deepseek Chat"
	DisplayNameDeepseekReasoner = "Deepseek Reasoner"
)

var DeepseekModelDefaults = map[spec.ModelName]spec.ModelDefaults{
	DeepseekChat: {
		DisplayName: DisplayNameDeepseekChat,
		IsEnabled:   true,
	},
	DeepseekReasoner: {
		DisplayName: DisplayNameDeepseekReasoner,
		IsEnabled:   true,
	},
}

var DeepseekModels = map[spec.ModelName]spec.ModelParams{
	DeepseekChat: {
		Name:            DeepseekChat,
		MaxPromptLength: 8192,
		MaxOutputLength: 8192,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	DeepseekReasoner: {
		Name:            DeepseekChat,
		MaxPromptLength: 8192,
		MaxOutputLength: 8192,
		Temperature:     Float64Ptr(1.0),
		Stream:          true,
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: "",
		Timeout:      120,
	},
}

var DeepseekProviderInfo = spec.ProviderInfo{
	Name:   ProviderNameDeepseek,
	APIKey: "",
	Origin: "https://api.deepseek.com",
	Type:   spec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: OpenAICompatibleChatCompletionPathPrefix,
}
