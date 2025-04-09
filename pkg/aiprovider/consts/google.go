package consts

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const ProviderNameGoogle spec.ProviderName = "google"

const (
	Gemini2FlashExp spec.ModelName = "gemini-2-flash-exp"
	Gemini15Flash   spec.ModelName = "gemini-1.5-flash"
	Gemini15Pro     spec.ModelName = "gemini-1.5-pro"
)

const (
	DisplayNameGemini2FlashExp = "Google Gemini 2 Flash exp"
	DisplayNameGemini15Flash   = "Google Gemini 1.5 Flash"
	DisplayNameGemini15Pro     = "Google Gemini 1.5 Pro"
)

var GoogleModels = map[spec.ModelName]spec.ModelParams{
	Gemini2FlashExp: {
		Name:             Gemini2FlashExp,
		MaxPromptLength:  4096,
		MaxOutputLength:  8192,
		Temperature:      spec.Float64Ptr(0.1),
		Stream:           true,
		ReasoningSupport: false,
		SystemPrompt:     "",
		Timeout:          120,
	},
	Gemini15Flash: {
		Name:             Gemini15Flash,
		MaxPromptLength:  4096,
		MaxOutputLength:  8192,
		Temperature:      spec.Float64Ptr(0.1),
		Stream:           true,
		ReasoningSupport: false,
		SystemPrompt:     "",
		Timeout:          120,
	},
	Gemini15Pro: {
		Name:             Gemini15Pro,
		MaxPromptLength:  4096,
		MaxOutputLength:  8192,
		Temperature:      spec.Float64Ptr(0.1),
		Stream:           true,
		ReasoningSupport: false,
		SystemPrompt:     "",
		Timeout:          120,
	},
}

var GoogleProviderInfo = spec.ProviderInfo{
	Name:   ProviderNameGoogle,
	APIKey: "",
	Origin: "https://generativelanguage.googleapis.com",
	Type:   spec.InbuiltOpenAICompatible,

	APIKeyHeaderKey:          OpenAICompatibleAPIKeyHeaderKey,
	DefaultHeaders:           OpenAICompatibleDefaultHeaders,
	ChatCompletionPathPrefix: "/v1beta/openai/chat/completions",
}
