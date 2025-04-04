package spec

// ModelName is an enumeration of model names.
type ModelName string

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

// ProviderName is an enumeration of provider names.
type ProviderName string

// ProviderInfo represents information about a provider.
type ProviderInfo struct {
	Name         ProviderName `json:"name"`
	APIKey       string       `json:"apiKey"`
	DefaultModel ModelName    `json:"defaultModel"`
	Engine       string       `json:"engine"`
	Origin       string       `json:"origin"`

	APIKeyHeaderKey          string                  `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string       `json:"defaultHeaders"`
	ChatCompletionPathPrefix string                  `json:"chatCompletionPathPrefix"`
	ModelPrefixes            []string                `json:"modelPrefixes"`
	Models                   map[ModelName]ModelInfo `json:"models"`
}

func (p *ProviderInfo) IsConfigured() bool {
	return p.APIKey != ""
}
