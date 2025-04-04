package huggingface

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const ProviderNameHuggingFace spec.ProviderName = "huggingface"

const (
	DeepseekCoder13BInstruct spec.ModelName = "deepseek-ai/deepseek-coder-1.3b-instruct"
)

var HuggingfaceModels = map[spec.ModelName]spec.ModelInfo{
	DeepseekCoder13BInstruct: {
		Name:                DeepseekCoder13BInstruct,
		DisplayName:         "HF Deepseek Coder 1.3b",
		Provider:            ProviderNameHuggingFace,
		MaxPromptLength:     4096,
		MaxOutputLength:     4096,
		DefaultTemperature:  0.1,
		StreamingSupport:    true,
		ReasoningSupport:    false,
		DefaultSystemPrompt: "",
		Timeout:             120,
	},
}

var HuggingfaceProviderInfo = spec.ProviderInfo{
	Name:         ProviderNameHuggingFace,
	APIKey:       "",
	Engine:       "",
	Origin:       "https://api-inference.huggingface.co",
	DefaultModel: DeepseekCoder13BInstruct,

	APIKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/models",

	Models: HuggingfaceModels,
	ModelPrefixes: []string{
		"microsoft/",
		"replit/",
		"Salesforce/",
		"bigcode/",
		"deepseek-ai/",
	},
}
