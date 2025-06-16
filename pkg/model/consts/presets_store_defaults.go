package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

var DefaultModelPresetsMap = map[spec.ProviderName]spec.ProviderModelPreset{
	ProviderNameAnthropic: {
		Claude4Sonnet: spec.ModelPreset{
			Name:         Claude4Sonnet,
			DisplayName:  DisplayNameClaude4Sonnet,
			IsEnabled:    true,
			ShortCommand: ShortCommandClaude4Sonnet,
		},
	},

	ProviderNameDeepseek: {
		DeepseekChat: {
			Name:         DeepseekChat,
			DisplayName:  DisplayNameDeepseekChat,
			IsEnabled:    true,
			ShortCommand: ShortCommandDeepseekChat,
		},
	},
	ProviderNameGoogle: {
		Gemini25Flash: {
			Name:         Gemini25Flash,
			DisplayName:  DisplayNameGemini25Flash,
			IsEnabled:    true,
			ShortCommand: ShortCommandGemini25Flash,
		},
	},
	ProviderNameHuggingFace: {
		DeepseekCoder13BInstruct: {
			Name:         DeepseekCoder13BInstruct,
			DisplayName:  DisplayNameDeepseekCoder13BInstruct,
			IsEnabled:    true,
			ShortCommand: ShortCommandDeepseekCoder13BInstruct,
		},
	},
	ProviderNameLlamaCPP: {
		Llama4Scout: {
			Name:         Llama4Scout,
			DisplayName:  DisplayNameLlama4Scout,
			IsEnabled:    true,
			ShortCommand: ShortCommandLlama4Scout,
		},
	},
	ProviderNameOpenAI: {
		GPT41: {
			Name:         GPT41,
			DisplayName:  DisplayNameGPT41,
			IsEnabled:    true,
			ShortCommand: ShortCommandGPT41,
		},
	},
}

var DefaultModelPresetsSchema = spec.ModelPresetsSchema{
	Version:      "1.0",
	ModelPresets: DefaultModelPresetsMap,
}
