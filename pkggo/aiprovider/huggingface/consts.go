package huggingface

import (
	"github.com/flexigpt/flexiui/pkggo/aiprovider/spec"
)

const ProviderNameHuggingFace spec.ProviderName = "huggingface"

const (
	DEEPSEEK_CODER_1_3B_INSTRUCT spec.ModelName = "deepseek-ai/deepseek-coder-1.3b-instruct"
)

var HuggingfaceModels = map[spec.ModelName]spec.ModelInfo{
	DEEPSEEK_CODER_1_3B_INSTRUCT: {
		Name:            DEEPSEEK_CODER_1_3B_INSTRUCT,
		DisplayName:     "HF Deepseek Coder 1.3b",
		Provider:        ProviderNameHuggingFace,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
}

var HuggingfaceProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameHuggingFace,
	ApiKey:                   "",
	Engine:                   "",
	DefaultOrigin:            "https://api-inference.huggingface.co",
	DefaultModel:             DEEPSEEK_CODER_1_3B_INSTRUCT,
	AdditionalSettings:       map[string]interface{}{},
	Timeout:                  120,
	ApiKeyHeaderKey:          "Authorization",
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
