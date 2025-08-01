package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

const (
	DeepseekChat     spec.ModelName = "deepseek-chat"
	DeepseekReasoner spec.ModelName = "deepseek-reasoner"
)

const (
	DisplayNameDeepseekChat     spec.ModelDisplayName = "Deepseek Chat"
	DisplayNameDeepseekReasoner spec.ModelDisplayName = "Deepseek Reasoner"
)

const (
	SlugDeepseekChat     spec.ModelSlug = "dsChat"
	SlugDeepseekReasoner spec.ModelSlug = "dsReason"
)

const (
	ModelPresetIDDeepseekChat     spec.ModelPresetID = "dsChat"
	ModelPresetIDDeepseekReasoner spec.ModelPresetID = "dsReason"
)

var DeepseekModelPresets = map[spec.ModelPresetID]spec.ModelPreset{
	ModelPresetIDDeepseekChat: {
		ID:          ModelPresetIDDeepseekChat,
		Name:        DeepseekChat,
		DisplayName: DisplayNameDeepseekChat,
		Slug:        SlugDeepseekChat,
		IsEnabled:   true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(8192),
		MaxOutputLength: IntPtr(8192),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	ModelPresetIDDeepseekReasoner: {
		ID:          ModelPresetIDDeepseekReasoner,
		Name:        DeepseekChat,
		DisplayName: DisplayNameDeepseekReasoner,
		Slug:        SlugDeepseekReasoner,
		IsEnabled:   true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(8192),
		MaxOutputLength: IntPtr(8192),
		Temperature:     Float64Ptr(1.0),
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: StringPtr(""),
		Timeout:      IntPtr(120),
	},
}

var DeepseekProviderPreset = spec.ProviderPreset{
	DefaultModelPresetID: ModelPresetIDDeepseekChat,
	ModelPresets:         DeepseekModelPresets,
}
