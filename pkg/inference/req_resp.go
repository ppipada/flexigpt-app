package inference

import (
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

type APIRequestDetails struct {
	URL         *string        `json:"url,omitempty"`
	Method      *string        `json:"method,omitempty"`
	Headers     map[string]any `json:"headers,omitempty"`
	Params      map[string]any `json:"params,omitempty"`
	Data        any            `json:"data,omitempty"`
	Timeout     *int           `json:"timeout,omitempty"`
	CurlCommand *string        `json:"curlCommand,omitempty"`
}

type APIResponseDetails struct {
	Data           any                `json:"data"`
	Status         int                `json:"status"`
	Headers        map[string]any     `json:"headers"`
	RequestDetails *APIRequestDetails `json:"requestDetails,omitempty"`
}

type APIErrorDetails struct {
	Message         string              `json:"message"`
	RequestDetails  *APIRequestDetails  `json:"requestDetails,omitempty"`
	ResponseDetails *APIResponseDetails `json:"responseDetails,omitempty"`
}

type APIFetchResponse[T any] struct {
	Data            *T                  `json:"data,omitempty"`
	ResponseDetails *APIResponseDetails `json:"responseDetails,omitempty"`
	RequestDetails  *APIRequestDetails  `json:"requestDetails,omitempty"`
	ErrorDetails    *APIErrorDetails    `json:"errorDetails,omitempty"`
}

type CompletionResponse struct {
	RequestDetails  *APIRequestDetails  `json:"requestDetails,omitempty"`
	ResponseDetails *APIResponseDetails `json:"responseDetails,omitempty"`
	ErrorDetails    *APIErrorDetails    `json:"errorDetails,omitempty"`
	RespContent     *string             `json:"respContent,omitempty"`
	FunctionName    *string             `json:"functionName,omitempty"`
	FunctionArgs    any                 `json:"functionArgs,omitempty"`
}
type CompletionRequest struct {
	ModelParams  spec.ModelParams                        `json:"modelParams"`
	Messages     []ChatCompletionRequestMessage          `json:"messages,omitempty"`
	Functions    []ChatCompletionFunctions               `json:"functions,omitempty"`
	FunctionCall CreateChatCompletionRequestFunctionCall `json:"functionCall,omitempty"`
}
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

type GetConfigurationInfoResponseBody struct {
	DefaultProvider       spec.ProviderName                         `json:"defaultProvider"`
	ConfiguredProviders   []spec.ProviderInfo                       `json:"configuredProviders"`
	InbuiltProviderModels map[spec.ProviderName]spec.ProviderPreset `json:"inbuiltProviderModels"`
}

type GetConfigurationInfoResponse struct {
	Body *GetConfigurationInfoResponseBody
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
	Provider     spec.ProviderName              `json:"provider"         required:"true"`
	Prompt       string                         `json:"prompt"           required:"true"`
	ModelParams  spec.ModelParams               `json:"spec.ModelParams" required:"true"`
	PrevMessages []ChatCompletionRequestMessage `json:"prevMessages"`
	OnStreamData func(data string) error        `json:"-"`
}

type FetchCompletionRequest struct {
	Body *FetchCompletionRequestBody
}

type FetchCompletionResponse struct {
	Body *CompletionResponse
}
