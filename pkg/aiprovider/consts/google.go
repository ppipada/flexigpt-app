package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"
)

const ProviderNameGoogle spec.ProviderName = "google"

const (
	Gemini25Pro      spec.ModelName = "gemini-2.5-pro-preview-05-06"
	Gemini25Flash    spec.ModelName = "gemini-2.5-flash-preview-04-17"
	Gemini2Flash     spec.ModelName = "gemini-2.0-flash"
	Gemini2FlashLite spec.ModelName = "gemini-2.0-flash-lite"
	Gemini15Pro      spec.ModelName = "gemini-1.5-pro"
)

const (
	DisplayNameGemini25Pro      = "Google Gemini 2.5 Pro"
	DisplayNameGemini25Flash    = "Google Gemini 2.5 Flash"
	DisplayNameGemini2Flash     = "Google Gemini 2.0 Flash"
	DisplayNameGemini2FlashLite = "Google Gemini 2.0 Flash Lite"
	DisplayNameGemini15Pro      = "Google Gemini 1.5 Pro"
)

var GoogleModelDefaults = map[spec.ModelName]spec.ModelDefaults{
	Gemini25Pro: {
		DisplayName: DisplayNameGemini25Pro,
		IsEnabled:   true,
	},
	Gemini25Flash: {
		DisplayName: DisplayNameGemini25Flash,
		IsEnabled:   true,
	},
	Gemini2Flash: {
		DisplayName: DisplayNameGemini2Flash,
		IsEnabled:   false,
	},
	Gemini2FlashLite: {
		DisplayName: DisplayNameGemini2FlashLite,
		IsEnabled:   false,
	},
	Gemini15Pro: {
		DisplayName: DisplayNameGemini15Pro,
		IsEnabled:   false,
	},
}

var GoogleModels = map[spec.ModelName]spec.ModelParams{
	Gemini25Pro: {
		Name:            Gemini25Pro,
		MaxPromptLength: 32768,
		MaxOutputLength: 32768,
		Temperature:     Float64Ptr(1.0),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Gemini25Flash: {
		Name:            Gemini25Flash,
		MaxPromptLength: 32768,
		MaxOutputLength: 32768,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Gemini2Flash: {
		Name:            Gemini2Flash,
		MaxPromptLength: 16384,
		MaxOutputLength: 8192,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Gemini2FlashLite: {
		Name:            Gemini2FlashLite,
		MaxPromptLength: 16384,
		MaxOutputLength: 8192,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
	},
	Gemini15Pro: {
		Name:            Gemini15Pro,
		MaxPromptLength: 16384,
		MaxOutputLength: 8192,
		Temperature:     Float64Ptr(0.1),
		Stream:          true,
		SystemPrompt:    "",
		Timeout:         120,
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
