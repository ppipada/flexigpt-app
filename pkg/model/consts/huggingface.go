package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

const ProviderNameHuggingFace spec.ProviderName = "huggingface"

const (
	DeepseekCoder13BInstruct             spec.ModelName         = "deepseek-ai/deepseek-coder-1.3b-instruct"
	DisplayNameDeepseekCoder13BInstruct  spec.ModelDisplayName  = "HF Deepseek Coder 1.3b"
	ShortCommandDeepseekCoder13BInstruct spec.ModelShortCommand = "hfDSCoder13"
)

var HuggingfaceModels = map[spec.ModelName]spec.ModelPreset{
	DeepseekCoder13BInstruct: {
		Name:            DeepseekCoder13BInstruct,
		DisplayName:     DisplayNameDeepseekCoder13BInstruct,
		IsEnabled:       true,
		ShortCommand:    ShortCommandDeepseekCoder13BInstruct,
		MaxPromptLength: IntPtr(4096),
		MaxOutputLength: IntPtr(4096),
		Temperature:     Float64Ptr(0.1),
		Stream:          BoolPtr(true),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
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
