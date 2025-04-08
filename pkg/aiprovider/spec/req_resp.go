package spec

type AddProviderRequestBody struct {
	APIKey                   string `json:"apiKey"`
	Origin                   string `json:"origin"`
	ChatCompletionPathPrefix string `json:"chatCompletionPathPrefix"`
}

type AddProviderRequest struct {
	Provider ProviderName `path:"provider" required:"true"`
	Body     *AddProviderRequestBody
}

type AddProviderResponse struct{}

type DeleteProviderRequest struct {
	Provider ProviderName `path:"provider" required:"true"`
}

type DeleteProviderResponse struct{}

type SetDefaultProviderRequestBody struct {
	Provider ProviderName `json:"provider" required:"true"`
}

type SetDefaultProviderRequest struct {
	Body *SetDefaultProviderRequestBody
}

type SetDefaultProviderResponse struct{}

type GetConfigurationInfoRequest struct{}

type GetConfigurationInfoResponse struct {
	Body *GetConfigurationInfoResponseBody
}

type GetConfigurationInfoResponseBody struct {
	DefaultProvider     ProviderName   `json:"defaultProvider"`
	ConfiguredProviders []ProviderInfo `json:"configuredProviders"`
}

type SetProviderAttributeRequestBody struct {
	APIKey                   *string `json:"apiKey,omitempty"`
	Origin                   *string `json:"origin,omitempty"`
	ChatCompletionPathPrefix *string `json:"chatCompletionPathPrefix,omitempty"`
}

type SetProviderAttributeRequest struct {
	Provider ProviderName `path:"provider" required:"true"`
	Body     *SetProviderAttributeRequestBody
}

type SetProviderAttributeResponse struct{}

type FetchCompletionRequestBody struct {
	Provider     ProviderName                   `json:"provider"     required:"true"`
	Prompt       string                         `json:"prompt"       required:"true"`
	ModelParams  ModelParams                    `json:"modelParams"  required:"true"`
	PrevMessages []ChatCompletionRequestMessage `json:"prevMessages"`
	OnStreamData func(data string) error        `json:"-"`
}

type FetchCompletionRequest struct {
	Body *FetchCompletionRequestBody
}

type FetchCompletionResponse struct {
	Body *CompletionResponse
}
