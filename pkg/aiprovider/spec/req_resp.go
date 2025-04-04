package spec

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
	APIKey       *string `json:"apiKey,omitempty"`
	DefaultModel *string `json:"defaultModel,omitempty"`
	Origin       *string `json:"origin,omitempty"`
}

type SetProviderAttributeRequest struct {
	Provider ProviderName `path:"provider" required:"true"`
	Body     *SetProviderAttributeRequestBody
}

type SetProviderAttributeResponse struct{}

type MakeCompletionRequestBody struct {
	Prompt       string                         `json:"prompt"       required:"true"`
	ModelParams  ModelParams                    `json:"modelParams"  required:"true"`
	PrevMessages []ChatCompletionRequestMessage `json:"prevMessages"`
}

type MakeCompletionRequest struct {
	Provider ProviderName `path:"provider" required:"true"`
	Body     *MakeCompletionRequestBody
}

type MakeCompletionResponse struct {
	Body *CompletionRequest
}

type FetchCompletionRequestBody struct {
	Provider     ProviderName            `json:"provider" required:"true"`
	Input        *CompletionRequest      `json:"input"    required:"true"`
	OnStreamData func(data string) error `json:"-"`
}

type FetchCompletionRequest struct {
	Body *FetchCompletionRequestBody
}

type FetchCompletionResponse struct {
	Body *CompletionResponse
}
