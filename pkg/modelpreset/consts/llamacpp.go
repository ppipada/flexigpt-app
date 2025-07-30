package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

const ProviderNameLlamaCPP spec.ProviderName = "llamacpp"

const (
	Llama4Behemoth spec.ModelName = "llama4-behemoth"
	Llama4Maverick spec.ModelName = "llama4-maverick"
	Llama4Scout    spec.ModelName = "llama4-scout"
)

const (
	DisplayNameLlama4Behemoth spec.ModelDisplayName = "LLama 4 Behemoth"
	DisplayNameLlama4Maverick spec.ModelDisplayName = "LLama 4 Maverick"
	DisplayNameLlama4Scout    spec.ModelDisplayName = "LLama 4 Scout"
)

const (
	SlugLlama4Behemoth spec.ModelSlug = "llama4Behemoth"
	SlugLlama4Maverick spec.ModelSlug = "llama4Maverick"
	SlugLlama4Scout    spec.ModelSlug = "llama4Scout"
)

const (
	ModelPresetIDLlama4Behemoth spec.ModelPresetID = "llama4Behemoth"
	ModelPresetIDLlama4Maverick spec.ModelPresetID = "llama4Maverick"
	ModelPresetIDLlama4Scout    spec.ModelPresetID = "llama4Scout"
)

var LlamacppModelPresets = map[spec.ModelPresetID]spec.ModelPreset{
	ModelPresetIDLlama4Behemoth: {
		ID:          ModelPresetIDLlama4Behemoth,
		Name:        Llama4Behemoth,
		DisplayName: DisplayNameLlama4Behemoth,
		Slug:        SlugLlama4Behemoth,
		IsEnabled:   true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(4096),
		MaxOutputLength: IntPtr(4096),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	ModelPresetIDLlama4Maverick: {
		ID:          ModelPresetIDLlama4Maverick,
		Name:        Llama4Maverick,
		DisplayName: DisplayNameLlama4Maverick,
		Slug:        SlugLlama4Maverick,
		IsEnabled:   true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(4096),
		MaxOutputLength: IntPtr(4096),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	ModelPresetIDLlama4Scout: {
		ID:          ModelPresetIDLlama4Scout,
		Name:        Llama4Scout,
		DisplayName: DisplayNameLlama4Scout,
		Slug:        SlugLlama4Scout,
		IsEnabled:   true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(4096),
		MaxOutputLength: IntPtr(4096),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
}

var LlamacppProviderPreset = spec.ProviderPreset{
	DefaultModelPresetID: ModelPresetIDLlama4Scout,
	ModelPresets:         LlamacppModelPresets,
}
