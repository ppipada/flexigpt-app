package anthropic

import "github.com/flexigpt/flexiui/pkg/aiprovider/spec"

const ProviderNameAnthropic spec.ProviderName = "anthropic"

const (
	CLAUDE_3_5_SONNET spec.ModelName = "claude-3-5-sonnet-20240620"
	CLAUDE_3_OPUS     spec.ModelName = "claude-3-opus-20240229"
	CLAUDE_3_SONNET   spec.ModelName = "claude-3-sonnet-20240229"
	CLAUDE_3_HAIKU    spec.ModelName = "claude-3-haiku-20240307"
)

var AnthropicModels = map[spec.ModelName]spec.ModelInfo{
	CLAUDE_3_5_SONNET: {
		Name:               CLAUDE_3_5_SONNET,
		DisplayName:        "Claude 3.5 Sonnet",
		Provider:           ProviderNameAnthropic,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	CLAUDE_3_OPUS: {
		Name:               CLAUDE_3_OPUS,
		DisplayName:        "Claude 3 Opus",
		Provider:           ProviderNameAnthropic,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	CLAUDE_3_SONNET: {
		Name:               CLAUDE_3_SONNET,
		DisplayName:        "Claude 3 Sonnet",
		Provider:           ProviderNameAnthropic,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
	},
	CLAUDE_3_HAIKU: {
		Name:               CLAUDE_3_HAIKU,
		DisplayName:        "Claude 3 Haiku",
		Provider:           ProviderNameAnthropic,
		MaxPromptLength:    4096,
		MaxOutputLength:    4096,
		DefaultTemperature: 0.1,
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
	Descriptions: map[string]string{
		"apiKey":                   "Your anthropic API key.",
		"engine":                   "The engine to be used for processing.",
		"defaultOrigin":            "Default origin to use for requests. This can be used to talk to any server that serves a compatible API",
		"defaultModel":             "Default model to use for chat requests",
		"additionalSettings":       "Any additional settings to pass to the model. Input as a JSON object",
		"timeout":                  "The timeout duration in milliseconds.",
		"apiKeyHeaderKey":          "The header key for the API key.",
		"defaultHeaders":           "The default headers to be included in requests.",
		"chatCompletionPathPrefix": "The path prefix for chat completions.",
		"defaultTemperature":       "Default temperature setting for chat requests",
		"modelPrefixes":            "Optional prefixes for models.",
	},
}
