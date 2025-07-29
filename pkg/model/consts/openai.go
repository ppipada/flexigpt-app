package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

const (
	ProviderNameOpenAI spec.ProviderName = "openai"
)

const (
	O4Mini    spec.ModelName = "o4-mini"
	O3Pro     spec.ModelName = "o3-pro"
	O3        spec.ModelName = "o3"
	O3Mini    spec.ModelName = "o3-mini"
	GPT41     spec.ModelName = "gpt-4.1"
	GPT41Mini spec.ModelName = "gpt-4.1-mini"
	GPT4O     spec.ModelName = "gpt-4o"
	GPT4OMini spec.ModelName = "gpt-4o-mini"
)

const (
	DisplayNameO4Mini    spec.ModelDisplayName = "OpenAI o4 Mini"
	DisplayNameO3Pro     spec.ModelDisplayName = "OpenAI o3 Pro"
	DisplayNameO3        spec.ModelDisplayName = "OpenAI o3"
	DisplayNameO3Mini    spec.ModelDisplayName = "OpenAI o3 Mini"
	DisplayNameGPT41     spec.ModelDisplayName = "OpenAI GPT 4.1"
	DisplayNameGPT41Mini spec.ModelDisplayName = "OpenAI GPT 4.1 Mini"
	DisplayNameGPT4O     spec.ModelDisplayName = "OpenAI GPT 4o"
	DisplayNameGPT4OMini spec.ModelDisplayName = "OpenAI GPT 4o Mini"
)

const (
	ShortCommandO4Mini    spec.ModelShortCommand = "o4Mini"
	ShortCommandO3Pro     spec.ModelShortCommand = "o3Pro"
	ShortCommandO3        spec.ModelShortCommand = "o3"
	ShortCommandO3Mini    spec.ModelShortCommand = "o3Mini"
	ShortCommandGPT41     spec.ModelShortCommand = "gpt41"
	ShortCommandGPT41Mini spec.ModelShortCommand = "gpt41Mini"
	ShortCommandGPT4O     spec.ModelShortCommand = "gpt4o"
	ShortCommandGPT4OMini spec.ModelShortCommand = "gpt4oMini"
)

const (
	ModelPresetIDO4Mini    spec.ModelPresetID = "o4Mini"
	ModelPresetIDO3Pro     spec.ModelPresetID = "o3Pro"
	ModelPresetIDO3        spec.ModelPresetID = "o3"
	ModelPresetIDO3Mini    spec.ModelPresetID = "o3Mini"
	ModelPresetIDGPT41     spec.ModelPresetID = "gpt41"
	ModelPresetIDGPT41Mini spec.ModelPresetID = "gpt41Mini"
	ModelPresetIDGPT4O     spec.ModelPresetID = "gpt4o"
	ModelPresetIDGPT4OMini spec.ModelPresetID = "gpt4oMini"
)

var OpenAIModelPresets = map[spec.ModelPresetID]spec.ModelPreset{
	ModelPresetIDO4Mini: {
		ID:           ModelPresetIDO4Mini,
		Name:         O4Mini,
		DisplayName:  DisplayNameO4Mini,
		ShortCommand: ShortCommandO4Mini,
		IsEnabled:    false,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(32768),
		MaxOutputLength: IntPtr(32768),
		Temperature:     Float64Ptr(1.0),
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: StringPtr("Formatting re-enabled.\nAlways output in Markdown format."),
		Timeout:      IntPtr(120),
	},
	ModelPresetIDO3Pro: {
		ID:           ModelPresetIDO3Pro,
		Name:         O3Pro,
		DisplayName:  DisplayNameO3Pro,
		ShortCommand: ShortCommandO3Pro,
		IsEnabled:    false,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(32768),
		MaxOutputLength: IntPtr(32768),
		Temperature:     Float64Ptr(1.0),
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: StringPtr("Formatting re-enabled.\nAlways output in Markdown format."),
		Timeout:      IntPtr(120),
	},
	ModelPresetIDO3: {
		ID:           ModelPresetIDO3,
		Name:         O3,
		DisplayName:  DisplayNameO3,
		ShortCommand: ShortCommandO3,
		IsEnabled:    true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(32768),
		MaxOutputLength: IntPtr(32768),
		Temperature:     Float64Ptr(1.0),
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: StringPtr("Formatting re-enabled.\nAlways output in Markdown format."),
		Timeout:      IntPtr(120),
	},
	ModelPresetIDO3Mini: {
		ID:           ModelPresetIDO3Mini,
		Name:         O3Mini,
		DisplayName:  DisplayNameO3Mini,
		ShortCommand: ShortCommandO3Mini,
		IsEnabled:    false,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(16384),
		MaxOutputLength: IntPtr(16384),
		Temperature:     Float64Ptr(1.0),
		Reasoning: &spec.ReasoningParams{
			Type:  spec.ReasoningTypeSingleWithLevels,
			Level: spec.ReasoningLevelMedium,
		},
		SystemPrompt: StringPtr("Formatting re-enabled.\nAlways output in Markdown format."),
		Timeout:      IntPtr(120),
	},
	ModelPresetIDGPT41: {
		ID:           ModelPresetIDGPT41,
		Name:         GPT41,
		DisplayName:  DisplayNameGPT41,
		ShortCommand: ShortCommandGPT41,
		IsEnabled:    true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(16384),
		MaxOutputLength: IntPtr(16384),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	ModelPresetIDGPT41Mini: {
		ID:           ModelPresetIDGPT41Mini,
		Name:         GPT41Mini,
		DisplayName:  DisplayNameGPT41Mini,
		ShortCommand: ShortCommandGPT41Mini,
		IsEnabled:    false,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(32768),
		MaxOutputLength: IntPtr(32768),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	ModelPresetIDGPT4O: {
		ID:           ModelPresetIDGPT4O,
		Name:         GPT4O,
		DisplayName:  DisplayNameGPT4O,
		ShortCommand: ShortCommandGPT4O,
		IsEnabled:    false,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(16384),
		MaxOutputLength: IntPtr(16384),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	ModelPresetIDGPT4OMini: {
		ID:           ModelPresetIDGPT4OMini,
		Name:         GPT4OMini,
		DisplayName:  DisplayNameGPT4OMini,
		ShortCommand: ShortCommandGPT4OMini,
		IsEnabled:    false,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(4096),
		MaxOutputLength: IntPtr(4096),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
}

var OpenAIProviderPreset = spec.ProviderPreset{
	DefaultModelPresetID: ModelPresetIDGPT41,
	ModelPresets:         OpenAIModelPresets,
}
