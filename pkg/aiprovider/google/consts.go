package google

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const ProviderNameGoogle spec.ProviderName = "google"

const (
	Gemini2FlashExp spec.ModelName = "gemini-2-flash-exp"
	Gemini15Flash   spec.ModelName = "gemini-1.5-flash"
	Gemini15Pro     spec.ModelName = "gemini-1.5-pro"
)

var GoogleModels = map[spec.ModelName]spec.ModelInfo{
	Gemini2FlashExp: {
		Name:            Gemini2FlashExp,
		DisplayName:     "Google Gemini 2 Flash exp",
		Provider:        ProviderNameGoogle,
		MaxPromptLength: 4096,
		MaxOutputLength: 8192,
	},
	Gemini15Flash: {
		Name:            Gemini15Flash,
		DisplayName:     "Google Gemini 1.5 Flash",
		Provider:        ProviderNameGoogle,
		MaxPromptLength: 4096,
		MaxOutputLength: 8192,
	},
	Gemini15Pro: {
		Name:            Gemini15Pro,
		DisplayName:     "Google Gemini 1.5 Pro",
		Provider:        ProviderNameGoogle,
		MaxPromptLength: 4096,
		MaxOutputLength: 8192,
	},
}

var GoogleProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameGoogle,
	APIKey:                   "",
	Engine:                   "",
	DefaultOrigin:            "https://generativelanguage.googleapis.com",
	DefaultModel:             Gemini15Flash,
	AdditionalSettings:       map[string]any{},
	Timeout:                  120,
	APIKeyHeaderKey:          "Authorization",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/v1beta/openai/chat/completions",
	DefaultTemperature:       0.1,
	StreamingSupport:         true,
	Models:                   GoogleModels,
}
