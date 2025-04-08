package spec

import "context"

type (
	ModelName    string
	ProviderName string
	ProviderType string
)

const (
	InbuiltSpecific         ProviderType = "inbuiltSpecific"
	InbuiltOpenAICompatible ProviderType = "inbuiltOpenAICompatible"
	CustomOpenAICompatible  ProviderType = "customOpenAICompatible"
)

// ModelInfo represents information about a model.
type ModelInfo struct {
	Name                ModelName    `json:"name"`
	DisplayName         string       `json:"displayName"`
	Provider            ProviderName `json:"provider"`
	MaxPromptLength     int          `json:"maxPromptLength"`
	MaxOutputLength     int          `json:"maxOutputLength"`
	DefaultTemperature  float64      `json:"defaultTemperature"`
	StreamingSupport    bool         `json:"streamingSupport"`
	ReasoningSupport    bool         `json:"reasoningSupport"`
	DefaultSystemPrompt string       `json:"defaultSystemPrompt"` // Any default system message to add. E.g: for o1 models
	Timeout             int          `json:"timeout"`
}

// ProviderInfo represents information about a provider.
type ProviderInfo struct {
	Name                     ProviderName            `json:"name"`
	APIKey                   string                  `json:"apiKey"`
	DefaultModel             ModelName               `json:"defaultModel"`
	Engine                   string                  `json:"engine"`
	Origin                   string                  `json:"origin"`
	Type                     ProviderType            `json:"type"`
	APIKeyHeaderKey          string                  `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string       `json:"defaultHeaders"`
	ChatCompletionPathPrefix string                  `json:"chatCompletionPathPrefix"`
	Models                   map[ModelName]ModelInfo `json:"models"`
}

func (p *ProviderInfo) IsConfigured() bool {
	return p.APIKey != ""
}

type CompletionProvider interface {
	GetProviderInfo(
		ctx context.Context,
	) *ProviderInfo
	IsConfigured(ctx context.Context) bool
	SetProviderAttribute(
		ctx context.Context,
		apiKey *string,
		origin *string,
		chatCompletionPathPrefix *string,
	) error
	FetchCompletion(
		ctx context.Context,
		prompt string,
		modelParams ModelParams,
		prevMessages []ChatCompletionRequestMessage,
		onStreamData func(data string) error,
	) (*CompletionResponse, error)
}
