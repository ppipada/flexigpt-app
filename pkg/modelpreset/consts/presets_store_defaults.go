package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

var DefaultProviderPresets = map[spec.ProviderName]spec.ProviderPreset{
	ProviderNameAnthropic: {
		DefaultModelPresetID: ModelPresetIDClaude4Sonnet,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			ModelPresetIDClaude4Sonnet: {
				ID:          ModelPresetIDClaude4Sonnet,
				Name:        Claude4Sonnet,
				DisplayName: DisplayNameClaude4Sonnet,
				IsEnabled:   true,
				Slug:        SlugClaude4Sonnet,
			},
		},
	},

	ProviderNameDeepseek: {
		DefaultModelPresetID: ModelPresetIDDeepseekChat,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			ModelPresetIDDeepseekChat: {
				ID:          ModelPresetIDDeepseekChat,
				Name:        DeepseekChat,
				DisplayName: DisplayNameDeepseekChat,
				IsEnabled:   true,
				Slug:        SlugDeepseekChat,
			},
		},
	},

	ProviderNameGoogle: {
		DefaultModelPresetID: ModelPresetIDGemini25Flash,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			ModelPresetIDGemini25Flash: {
				ID:          ModelPresetIDGemini25Flash,
				Name:        Gemini25Flash,
				DisplayName: DisplayNameGemini25Flash,
				IsEnabled:   true,
				Slug:        SlugGemini25Flash,
			},
		},
	},

	ProviderNameHuggingFace: {
		DefaultModelPresetID: ModelPresetIDDeepseekCoder13BInstruct,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			ModelPresetIDDeepseekCoder13BInstruct: {
				ID:          ModelPresetIDDeepseekCoder13BInstruct,
				Name:        DeepseekCoder13BInstruct,
				DisplayName: DisplayNameDeepseekCoder13BInstruct,
				IsEnabled:   true,
				Slug:        SlugDeepseekCoder13BInstruct,
			},
		},
	},

	ProviderNameLlamaCPP: {
		DefaultModelPresetID: ModelPresetIDLlama4Scout,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			ModelPresetIDLlama4Scout: {
				ID:          ModelPresetIDLlama4Scout,
				Name:        Llama4Scout,
				DisplayName: DisplayNameLlama4Scout,
				IsEnabled:   true,
				Slug:        SlugLlama4Scout,
			},
		},
	},

	ProviderNameOpenAI: {
		DefaultModelPresetID: ModelPresetIDGPT41,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			ModelPresetIDGPT41: {
				ID:          ModelPresetIDGPT41,
				Name:        GPT41,
				DisplayName: DisplayNameGPT41,
				IsEnabled:   true,
				Slug:        SlugGPT41,
			},
		},
	},
}

var DefaultPresetsSchema = spec.PresetsSchema{
	Version:         "1.0",
	ProviderPresets: DefaultProviderPresets,
}
