package anthropic

import (
	"github.com/flexigpt/flexiui/pkggo/aiprovider/spec"
)

const ProviderNameAnthropic spec.ProviderName = "anthropic"

const (
	CLAUDE_3_5_SONNET spec.ModelName = "claude-3-5-sonnet-20240620"
	CLAUDE_3_OPUS     spec.ModelName = "claude-3-opus-20240229"
	CLAUDE_3_SONNET   spec.ModelName = "claude-3-sonnet-20240229"
	CLAUDE_3_HAIKU    spec.ModelName = "claude-3-haiku-20240307"
)

var AnthropicModels = map[spec.ModelName]spec.ModelInfo{
	CLAUDE_3_5_SONNET: {
		Name:            CLAUDE_3_5_SONNET,
		DisplayName:     "Claude 3.5 Sonnet",
		Provider:        ProviderNameAnthropic,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
	CLAUDE_3_OPUS: {
		Name:            CLAUDE_3_OPUS,
		DisplayName:     "Claude 3 Opus",
		Provider:        ProviderNameAnthropic,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
	CLAUDE_3_SONNET: {
		Name:            CLAUDE_3_SONNET,
		DisplayName:     "Claude 3 Sonnet",
		Provider:        ProviderNameAnthropic,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
	CLAUDE_3_HAIKU: {
		Name:            CLAUDE_3_HAIKU,
		DisplayName:     "Claude 3 Haiku",
		Provider:        ProviderNameAnthropic,
		MaxPromptLength: 4096,
		MaxOutputLength: 4096,
	},
}

var AnthropicProviderInfo = spec.ProviderInfo{
	Name:               ProviderNameAnthropic,
	ApiKey:             "",
	Engine:             "",
	DefaultOrigin:      "https://api.anthropic.com/v1",
	DefaultModel:       CLAUDE_3_HAIKU,
	AdditionalSettings: map[string]interface{}{},
	Timeout:            120,
	ApiKeyHeaderKey:    "x-api-key",
	DefaultHeaders: map[string]string{
		"content-type":      "application/json",
		"accept":            "application/json",
		"anthropic-version": "2023-06-01",
	},
	ChatCompletionPathPrefix: "/messages",
	DefaultTemperature:       0.1,
	StreamingSupport:         true,
}
