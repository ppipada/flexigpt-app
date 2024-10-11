package spec

// ProviderName is an enumeration of provider names.
type ProviderName string

const (
	OPENAI      ProviderName = "openai"
	ANTHROPIC   ProviderName = "anthropic"
	GOOGLE      ProviderName = "google"
	HUGGINGFACE ProviderName = "huggingface"
	LLAMACPP    ProviderName = "llamacpp"
)

// ProviderInfo represents information about a provider.
type ProviderInfo struct {
	Name                     ProviderName           `json:"name"`
	ApiKey                   string                 `json:"apiKey"`
	Engine                   string                 `json:"engine"`
	DefaultOrigin            string                 `json:"defaultOrigin"`
	DefaultModel             ModelName              `json:"defaultModel"`
	AdditionalSettings       map[string]interface{} `json:"additionalSettings"`
	Timeout                  int                    `json:"timeout"`
	ApiKeyHeaderKey          string                 `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string      `json:"defaultHeaders"`
	ChatCompletionPathPrefix string                 `json:"chatCompletionPathPrefix"`
	DefaultTemperature       float64                `json:"defaultTemperature"`
	ModelPrefixes            []string               `json:"modelPrefixes"`
	Descriptions             map[string]string      `json:"descriptions"`
	StreamingSupport         bool                   `json:"streamingSupport"`
}

// GetDescription returns the description for a given key.
func (p *ProviderInfo) GetDescription(key string) string {
	if description, exists := p.Descriptions[key]; exists {
		return description
	}
	return ""
}

func (p *ProviderInfo) IsConfigured() bool {
	return p.ApiKey != ""
}

type CompletionProvider interface {
	GetProviderInfo() (*ProviderInfo, error)
	GetCompletionRequest(
		prompt string,
		prevMessages []ChatCompletionRequestMessage,
		inputParams map[string]interface{},
		stream bool,
	) (*CompletionRequest, error)
	Completion(
		input CompletionRequest,
		onStreamData func(data string) error,
	) (*CompletionResponse, error)
	SetAttribute(
		apiKey *string,
		defaultModel *string,
		defaultTemperature *float64,
		defaultOrigin *string,
	) error
	IsConfigured() bool
}
