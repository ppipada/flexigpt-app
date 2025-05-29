package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"
)

const ProviderNameHuggingFace spec.ProviderName = "huggingface"

const (
	DeepseekCoder13BInstruct spec.ModelName = "deepseek-ai/deepseek-coder-1.3b-instruct"
)

const DisplayNameDeepseekCoder13BInstruct = "HF Deepseek Coder 1.3b"

var HuggingfaceModelDefaults = map[spec.ModelName]spec.ModelDefaults{
	DeepseekCoder13BInstruct: {
		DisplayName: DisplayNameDeepseekCoder13BInstruct,
		IsEnabled:   true,
	},
}

var HuggingfaceModels = map[spec.ModelName]spec.ModelParams{
	DeepseekCoder13BInstruct: {
		Name:            DeepseekCoder13BInstruct,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
		Temperature:     spec.Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
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
