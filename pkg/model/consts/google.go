package consts

import (
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

const ProviderNameGoogle spec.ProviderName = "google"

const (
	Gemini25Pro   spec.ModelName = "gemini-2.5-pro-preview-05-06"
	Gemini25Flash spec.ModelName = "gemini-2.5-flash-preview-05-20"
)

const (
	DisplayNameGemini25Pro   spec.ModelDisplayName = "Google Gemini 2.5 Pro"
	DisplayNameGemini25Flash spec.ModelDisplayName = "Google Gemini 2.5 Flash"
)

const (
	ShortCommandGemini25Pro   spec.ModelShortCommand = "gemini25Pro"
	ShortCommandGemini25Flash spec.ModelShortCommand = "gemini25Flash"
)

var GoogleModels = map[spec.ModelName]spec.ModelPreset{
	Gemini25Pro: {
		Name:            Gemini25Pro,
		DisplayName:     DisplayNameGemini25Pro,
		IsEnabled:       true,
		ShortCommand:    ShortCommandGemini25Pro,
		MaxPromptLength: IntPtr(32768),
		MaxOutputLength: IntPtr(32768),
		Temperature:     Float64Ptr(1.0),
		Stream:          BoolPtr(true),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
	},
	Gemini25Flash: {
		Name:            Gemini25Flash,
		DisplayName:     DisplayNameGemini25Flash,
		IsEnabled:       true,
		ShortCommand:    ShortCommandGemini25Flash,
		MaxPromptLength: IntPtr(32768),
		MaxOutputLength: IntPtr(32768),
		Temperature:     Float64Ptr(0.1),
		Stream:          BoolPtr(true),
		SystemPrompt:    StringPtr(""),
		Timeout:         IntPtr(120),
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
