package spec

import "context"

// ModelName is an enumeration of model names.
type ModelName string

// ModelInfo represents information about a model.
type ModelInfo struct {
	Name        ModelName    `json:"name"`
	DisplayName string       `json:"displayName"`
	Provider    ProviderName `json:"provider"`
	// MaxPromptLength is set as the one in ModelInfo. The value from input is used if it is less than the ModelInfo
	MaxPromptLength int `json:"maxPromptLength"`
	// MaxOutputLength is set only if it is in input and less than the ModelInfo
	MaxOutputLength int `json:"maxOutputLength"`
	// Add this only if you want to override the one set in provider.
	// Temperature resolution is: ModelInfo > inputParams > ProviderInfo
	DefaultTemperature *float64 `json:"defaultTemperature"`
	// Add this only if you want to override the one set in provider.
	// Streaming resolution is: ModelInfo > ProviderInfo
	StreamingSupport *bool `json:"streamingSupport"`
}

// ProviderName is an enumeration of provider names.
type ProviderName string

// ProviderInfo represents information about a provider.
type ProviderInfo struct {
	Name                     ProviderName      `json:"name"`
	APIKey                   string            `json:"apiKey"`
	Engine                   string            `json:"engine"`
	DefaultOrigin            string            `json:"defaultOrigin"`
	DefaultModel             ModelName         `json:"defaultModel"`
	AdditionalSettings       map[string]any    `json:"additionalSettings"`
	Timeout                  int               `json:"timeout"`
	APIKeyHeaderKey          string            `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string `json:"defaultHeaders"`
	ChatCompletionPathPrefix string            `json:"chatCompletionPathPrefix"`
	// Temperature resolution is: inputParams > ModelInfo > ProviderInfo
	DefaultTemperature float64  `json:"defaultTemperature"`
	ModelPrefixes      []string `json:"modelPrefixes"`
	// Streaming resolution is: ModelInfo > ProviderInfo
	StreamingSupport bool                    `json:"streamingSupport"`
	Models           map[ModelName]ModelInfo `json:"models"`
}

func (p *ProviderInfo) IsConfigured() bool {
	return p.APIKey != ""
}

type CompletionProvider interface {
	MakeCompletion(
		ctx context.Context,
		modelInfo ModelInfo,
		prompt string,
		prevMessages []ChatCompletionRequestMessage,
		inputParams map[string]any,
	) (*CompletionRequest, error)
	FetchCompletion(
		ctx context.Context,
		input CompletionRequest,
		onStreamData func(data string) error,
	) (*CompletionResponse, error)
	SetProviderAttribute(
		ctx context.Context,
		apiKey *string,
		defaultModel *string,
		defaultTemperature *float64,
		defaultOrigin *string,
	) error
	IsConfigured(ctx context.Context) bool
}
