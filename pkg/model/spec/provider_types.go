package spec

type (
	ProviderName string
	ProviderType string
)

const (
	InbuiltSpecific         ProviderType = "inbuiltSpecific"
	InbuiltOpenAICompatible ProviderType = "inbuiltOpenAICompatible"
	CustomOpenAICompatible  ProviderType = "customOpenAICompatible"
)

// ProviderInfo represents information about a provider.
type ProviderInfo struct {
	Name                     ProviderName      `json:"name"`
	APIKey                   string            `json:"apiKey"`
	Origin                   string            `json:"origin"`
	ChatCompletionPathPrefix string            `json:"chatCompletionPathPrefix"`
	APIKeyHeaderKey          string            `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string `json:"defaultHeaders"`
	Type                     ProviderType      `json:"type"`
}
