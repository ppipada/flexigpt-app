package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

const (
	DeepseekCoder13BInstruct              spec.ModelName        = "deepseek-ai/deepseek-coder-1.3b-instruct"
	DisplayNameDeepseekCoder13BInstruct   spec.ModelDisplayName = "HF Deepseek Coder 1.3b"
	SlugDeepseekCoder13BInstruct          spec.ModelSlug        = "hfDSCoder13"
	ModelPresetIDDeepseekCoder13BInstruct spec.ModelPresetID    = "hfDSCoder13"
)

var HuggingfaceModelPresets = map[spec.ModelPresetID]spec.ModelPreset{
	ModelPresetIDDeepseekCoder13BInstruct: {
		ID:          ModelPresetIDDeepseekCoder13BInstruct,
		Name:        DeepseekCoder13BInstruct,
		DisplayName: DisplayNameDeepseekCoder13BInstruct,
		Slug:        SlugDeepseekCoder13BInstruct,
		IsEnabled:   true,

		Stream:          BoolPtr(true),
		MaxPromptLength: IntPtr(4096),
		MaxOutputLength: IntPtr(4096),
		Temperature:     Float64Ptr(0.1),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
}

var HuggingfaceProviderPreset = spec.ProviderPreset{
	DefaultModelPresetID: ModelPresetIDDeepseekCoder13BInstruct,
	ModelPresets:         HuggingfaceModelPresets,
}
