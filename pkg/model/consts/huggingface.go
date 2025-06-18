package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

const ProviderNameHuggingFace spec.ProviderName = "huggingface"

const (
	DeepseekCoder13BInstruct              spec.ModelName         = "deepseek-ai/deepseek-coder-1.3b-instruct"
	DisplayNameDeepseekCoder13BInstruct   spec.ModelDisplayName  = "HF Deepseek Coder 1.3b"
	ShortCommandDeepseekCoder13BInstruct  spec.ModelShortCommand = "hfDSCoder13"
	ModelPresetIDDeepseekCoder13BInstruct spec.ModelPresetID     = "hfDSCoder13"
)

var HuggingfaceModelPresets = map[spec.ModelPresetID]spec.ModelPreset{
	ModelPresetIDDeepseekCoder13BInstruct: {
		ID:           ModelPresetIDDeepseekCoder13BInstruct,
		Name:         DeepseekCoder13BInstruct,
		DisplayName:  DisplayNameDeepseekCoder13BInstruct,
		ShortCommand: ShortCommandDeepseekCoder13BInstruct,
		IsEnabled:    true,

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

var HuggingfaceProviderInfo = spec.ProviderInfo{
	Name:   ProviderNameHuggingFace,
	APIKey: "",
	Origin: "https://api-inference.huggingface.co",
	Type:   spec.InbuiltSpecific,

	APIKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/models",
}
