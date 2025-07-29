package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

const ProviderNameGoogle spec.ProviderName = "google"

const (
	Gemini25Pro       spec.ModelName = "gemini-2.5-pro"
	Gemini25Flash     spec.ModelName = "gemini-2.5-flash"
	Gemini25FlashLite spec.ModelName = "gemini-2.5-flash-lite-preview-06-17"
)

const (
	DisplayNameGemini25Pro       spec.ModelDisplayName = "Google Gemini 2.5 Pro"
	DisplayNameGemini25Flash     spec.ModelDisplayName = "Google Gemini 2.5 Flash"
	DisplayNameGemini25FlashLite spec.ModelDisplayName = "Google Gemini 2.5 Flash Lite"
)

const (
	ShortCommandGemini25Pro       spec.ModelShortCommand = "gemini25Pro"
	ShortCommandGemini25Flash     spec.ModelShortCommand = "gemini25Flash"
	ShortCommandGemini25FlashLite spec.ModelShortCommand = "gemini25FlashLite"
)

const (
	ModelPresetIDGemini25Pro       spec.ModelPresetID = "gemini25Pro"
	ModelPresetIDGemini25Flash     spec.ModelPresetID = "gemini25Flash"
	ModelPresetIDGemini25FlashLite spec.ModelPresetID = "gemini25FlashLite"
)

var GoogleModelPresets = map[spec.ModelPresetID]spec.ModelPreset{
	ModelPresetIDGemini25Pro: {
		ID:           ModelPresetIDGemini25Pro,
		Name:         Gemini25Pro,
		DisplayName:  DisplayNameGemini25Pro,
		ShortCommand: ShortCommandGemini25Pro,
		IsEnabled:    true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(32768),
		MaxOutputLength: IntPtr(32768),
		Temperature:     Float64Ptr(1.0),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	ModelPresetIDGemini25Flash: {
		ID:           ModelPresetIDGemini25Flash,
		Name:         Gemini25Flash,
		DisplayName:  DisplayNameGemini25Flash,
		ShortCommand: ShortCommandGemini25Flash,
		IsEnabled:    true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(32768),
		MaxOutputLength: IntPtr(32768),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	ModelPresetIDGemini25FlashLite: {
		ID:           ModelPresetIDGemini25FlashLite,
		Name:         Gemini25FlashLite,
		DisplayName:  DisplayNameGemini25FlashLite,
		ShortCommand: ShortCommandGemini25FlashLite,
		IsEnabled:    true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(32768),
		MaxOutputLength: IntPtr(32768),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
}

var GoogleProviderPreset = spec.ProviderPreset{
	DefaultModelPresetID: ModelPresetIDGemini25Flash,
	ModelPresets:         GoogleModelPresets,
}
