package google

import (
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

const ProviderNameGoogle spec.ProviderName = "google"

const (
	GEMINI_1_5_FLASH spec.ModelName = "gemini-1.5-flash"
	GEMINI_1_5_PRO   spec.ModelName = "gemini-1.5-pro"
)

var GoogleModels = map[spec.ModelName]spec.ModelInfo{
	GEMINI_1_5_FLASH: {
		Name:            GEMINI_1_5_FLASH,
		DisplayName:     "Google Gemini 1.5 Flash",
		Provider:        ProviderNameGoogle,
		MaxPromptLength: 4096,
		MaxOutputLength: 8192,
	},
	GEMINI_1_5_PRO: {
		Name:            GEMINI_1_5_PRO,
		DisplayName:     "Google Gemini 1.5 Pro",
		Provider:        ProviderNameGoogle,
		MaxPromptLength: 4096,
		MaxOutputLength: 8192,
	},
}

var GoogleProviderInfo = spec.ProviderInfo{
	Name:                     ProviderNameGoogle,
	ApiKey:                   "",
	Engine:                   "",
	DefaultOrigin:            "https://generativelanguage.googleapis.com",
	DefaultModel:             GEMINI_1_5_FLASH,
	AdditionalSettings:       map[string]interface{}{},
	Timeout:                  120,
	ApiKeyHeaderKey:          "x-goog-api-key",
	DefaultHeaders:           map[string]string{"content-type": "application/json"},
	ChatCompletionPathPrefix: "/v1beta/models",
	DefaultTemperature:       0.1,
	StreamingSupport:         true,
}
