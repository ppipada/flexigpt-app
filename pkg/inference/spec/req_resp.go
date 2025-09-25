package spec

import (
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

type AddProviderRequestBody struct {
	SDKType                  modelpresetSpec.ProviderSDKType `json:"sdkType"`
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

type ResponseContentType string

const (
	ResponseContentTypeText            = "text"
	ResponseContentTypeThinking        = "thinking"
	ResponseContentTypeThinkingSummary = "thinkingSummary"
)

type ResponseContent struct {
	Type    ResponseContentType `json:"type"`
	Content string              `json:"content"`
}

type CompletionResponse struct {
	RequestDetails  *APIRequestDetails  `json:"requestDetails,omitempty"`
	ResponseDetails *APIResponseDetails `json:"responseDetails,omitempty"`
	ErrorDetails    *APIErrorDetails    `json:"errorDetails,omitempty"`
	ResponseContent []ResponseContent   `json:"responseContent,omitempty"`
}

type CompletionData struct {
	ModelParams  ModelParams                          `json:"modelParams"`
	Messages     []ChatCompletionDataMessage          `json:"messages,omitempty"`
	Functions    []ChatCompletionFunctions            `json:"functions,omitempty"`
	FunctionCall CreateChatCompletionDataFunctionCall `json:"functionCall,omitempty"`
}

type BuildCompletionDataRequestBody struct {
	Prompt       string                      `json:"prompt"       required:"true"`
	ModelParams  ModelParams                 `json:"modelParams"  required:"true"`
	PrevMessages []ChatCompletionDataMessage `json:"prevMessages"`
}

type BuildCompletionDataRequest struct {
	Provider modelpresetSpec.ProviderName `path:"provider" required:"true"`
	Body     *BuildCompletionDataRequestBody
}

type BuildCompletionDataResponse struct {
	Body *CompletionData
}

type FetchCompletionRequestBody struct {
	CompletionData       *CompletionData                 `json:"completionData" required:"true"`
	OnStreamTextData     func(textData string) error     `json:"-"`
	OnStreamThinkingData func(thinkingData string) error `json:"-"`
}

type FetchCompletionRequest struct {
	Provider modelpresetSpec.ProviderName `path:"provider" required:"true"`
	Body     *FetchCompletionRequestBody
}

type FetchCompletionResponse struct {
	Body *CompletionResponse
}
