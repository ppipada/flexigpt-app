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
		Name:            DeepseekCoder13BInstruct,
		DisplayName:     "HF Deepseek Coder 1.3b",
		Provider:        ProviderNameHuggingFace,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
}

var HuggingfaceProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameHuggingFace,
	APIKey:                   "",
	Engine:                   "",
	DefaultOrigin:            "https://api-inference.huggingface.co",
	DefaultModel:             DeepseekCoder13BInstruct,
	AdditionalSettings:       map[string]interface{}{},
	Timeout:                  120,
	APIKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/models",
	DefaultTemperature:       0.1,
	ModelPrefixes: []string{
		"microsoft/",
		"replit/",
		"Salesforce/",
		"bigcode/",
		"deepseek-ai/",
	},
	StreamingSupport: false,
}
