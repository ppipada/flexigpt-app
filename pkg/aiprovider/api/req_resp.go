package api

import "github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"

type AddProviderRequestBody struct {
	APIKey                   string `json:"apiKey"`
	Origin                   string `json:"origin"`
	ChatCompletionPathPrefix string `json:"chatCompletionPathPrefix"`
}

type AddProviderRequest struct {
	Provider spec.ProviderName `path:"provider" required:"true"`
	Body     *AddProviderRequestBody
}

type AddProviderResponse struct{}

type DeleteProviderRequest struct {
	Provider spec.ProviderName `path:"provider" required:"true"`
}

type DeleteProviderResponse struct{}

type SetDefaultProviderRequestBody struct {
	Provider spec.ProviderName `json:"provider" required:"true"`
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
	DefaultProvider              spec.ProviderName                                           `json:"defaultProvider"`
	ConfiguredProviders          []spec.ProviderInfo                                         `json:"configuredProviders"`
	InbuiltProviderModels        map[spec.ProviderName]map[spec.ModelName]spec.ModelParams   `json:"inbuiltProviderModels"`
	InbuiltProviderModelDefaults map[spec.ProviderName]map[spec.ModelName]spec.ModelDefaults `json:"inbuiltProviderModelDefaults"`
}

type SetProviderAPIKeyRequestBody struct {
	APIKey string `json:"apiKey" required:"true"`
}

type SetProviderAPIKeyRequest struct {
	Provider spec.ProviderName `path:"provider" required:"true"`
	Body     *SetProviderAPIKeyRequestBody
}

type SetProviderAPIKeyResponse struct{}

type SetProviderAttributeRequestBody struct {
	Origin                   *string `json:"origin,omitempty"`
	ChatCompletionPathPrefix *string `json:"chatCompletionPathPrefix,omitempty"`
}

type SetProviderAttributeRequest struct {
	Provider spec.ProviderName `path:"provider" required:"true"`
	Body     *SetProviderAttributeRequestBody
}

type SetProviderAttributeResponse struct{}

type FetchCompletionRequestBody struct {
	Provider     spec.ProviderName                   `json:"provider"         required:"true"`
	Prompt       string                              `json:"prompt"           required:"true"`
	ModelParams  spec.ModelParams                    `json:"spec.ModelParams" required:"true"`
	PrevMessages []spec.ChatCompletionRequestMessage `json:"prevMessages"`
	OnStreamData func(data string) error             `json:"-"`
}

type FetchCompletionRequest struct {
	Body *FetchCompletionRequestBody
}

type FetchCompletionResponse struct {
	Body *CompletionResponse
}
