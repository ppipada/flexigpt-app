package spec

import (
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
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
	ModelParams  ModelParams                             `json:"modelParams"`
	Messages     []ChatCompletionRequestMessage          `json:"messages,omitempty"`
	Functions    []ChatCompletionFunctions               `json:"functions,omitempty"`
	FunctionCall CreateChatCompletionRequestFunctionCall `json:"functionCall,omitempty"`
}

type AddProviderRequestBody struct {
	APIType                  modelpresetSpec.ProviderAPIType `json:"apiType"`
	Origin                   string                          `json:"origin"`
	ChatCompletionPathPrefix string                          `json:"chatCompletionPathPrefix"`
	APIKeyHeaderKey          string                          `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string               `json:"defaultHeaders"`
}

type AddProviderRequest struct {
	Provider modelpresetSpec.ProviderName `path:"provider" required:"true"`
	Body     *AddProviderRequestBody
}

type AddProviderResponse struct{}

type DeleteProviderRequest struct {
	Provider modelpresetSpec.ProviderName `path:"provider" required:"true"`
}

type DeleteProviderResponse struct{}

type SetProviderAPIKeyRequestBody struct {
	APIKey string `json:"apiKey" required:"true"`
}

type SetProviderAPIKeyRequest struct {
	Provider modelpresetSpec.ProviderName `path:"provider" required:"true"`
	Body     *SetProviderAPIKeyRequestBody
}

type SetProviderAPIKeyResponse struct{}

type FetchCompletionRequestBody struct {
	Provider     modelpresetSpec.ProviderName   `json:"provider"         required:"true"`
	Prompt       string                         `json:"prompt"           required:"true"`
	ModelParams  ModelParams                    `json:"spec.ModelParams" required:"true"`
	PrevMessages []ChatCompletionRequestMessage `json:"prevMessages"`
	OnStreamData func(data string) error        `json:"-"`
}

type FetchCompletionRequest struct {
	Body *FetchCompletionRequestBody
}

type FetchCompletionResponse struct {
	Body *CompletionResponse
}
