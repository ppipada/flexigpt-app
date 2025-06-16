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
	DisplayNameDeepseekChat     spec.ModelDisplayName = "Deepseek Chat"
	DisplayNameDeepseekReasoner spec.ModelDisplayName = "Deepseek Reasoner"
)

const (
	ShortCommandDeepseekChat     spec.ModelShortCommand = "dsChat"
	ShortCommandDeepseekReasoner spec.ModelShortCommand = "dsReason"
)

var DeepseekModels = map[spec.ModelName]spec.ModelPreset{
	DeepseekChat: {
		Name:            DeepseekChat,
		DisplayName:     DisplayNameDeepseekChat,
		IsEnabled:       true,
		ShortCommand:    ShortCommandDeepseekChat,
		MaxPromptLength: IntPtr(8192),
		MaxOutputLength: IntPtr(8192),
		Temperature:     Float64Ptr(0.1),
		Stream:          BoolPtr(true),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	DeepseekReasoner: {
		Name:            DeepseekChat,
		DisplayName:     DisplayNameDeepseekReasoner,
		IsEnabled:       true,
		ShortCommand:    ShortCommandDeepseekReasoner,
		MaxPromptLength: IntPtr(8192),
		MaxOutputLength: IntPtr(8192),
		Temperature:     Float64Ptr(1.0),
		Stream:          BoolPtr(true),
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: StringPtr(""),
		Timeout:      IntPtr(120),
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
