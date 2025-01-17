package google

import (
	"github.com/flexigpt/flexiui/pkggo/aiprovider/spec"
)

const ProviderNameGoogle spec.ProviderName = "google"

const (
	Gemini15Flash spec.ModelName = "gemini-1.5-flash"
	Gemini15Pro   spec.ModelName = "gemini-1.5-pro"
)

var GoogleModels = map[spec.ModelName]spec.ModelInfo{
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
	AdditionalSettings:       map[string]interface{}{},
	Timeout:                  120,
	APIKeyHeaderKey:          "x-goog-api-key",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/v1beta/models",
	DefaultTemperature:       0.1,
	StreamingSupport:         true,
}
