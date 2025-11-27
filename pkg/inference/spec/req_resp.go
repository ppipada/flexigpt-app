package spec

import (
	"github.com/ppipada/flexigpt-app/pkg/attachment"
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
	ResponseContentTypeText            ResponseContentType = "text"
	ResponseContentTypeThinking        ResponseContentType = "thinking"
	ResponseContentTypeThinkingSummary ResponseContentType = "thinkingSummary"
	// ResponseContentTypeImage represents image content (for example, a data URL or
	// some other image identifier/handle). Callers are free to interpret the
	// Content field appropriately for their UI.
	ResponseContentTypeImage ResponseContentType = "image"
)

type ResponseContent struct {
	Type    ResponseContentType `json:"type"`
	Content string              `json:"content"`
}

// ResponseToolCall captures model-requested (or model-returned) tool invocations
// so the frontend can decide how to proceed (execute locally, display results,
// etc.).
type ResponseToolCall struct {
	ID        string `json:"id,omitempty"`
	CallID    string `json:"callID,omitempty"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
}

type BuildCompletionDataRequestBody struct {
	ModelParams    ModelParams                 `json:"modelParams"            required:"true"`
	CurrentMessage ChatCompletionDataMessage   `json:"currentMessage"         required:"true"`
	PrevMessages   []ChatCompletionDataMessage `json:"prevMessages,omitempty"`
	ToolChoices    []ChatCompletionToolChoice  `json:"toolChoices,omitempty"`
	Attachments    []attachment.Attachment     `json:"attachments,omitempty"`
}

type BuildCompletionDataRequest struct {
	Provider modelpresetSpec.ProviderName `path:"provider" required:"true"`
	Body     *BuildCompletionDataRequestBody
}

type FetchCompletionData struct {
	ModelParams ModelParams                 `json:"modelParams"`
	Messages    []ChatCompletionDataMessage `json:"messages,omitempty"`
	ToolChoices []FetchCompletionToolChoice `json:"toolChoices,omitempty"`
	Attachments []attachment.Attachment     `json:"attachments,omitempty"`
}

type BuildCompletionDataResponse struct {
	Body *FetchCompletionData
}

type FetchCompletionRequestBody struct {
	FetchCompletionData  *FetchCompletionData            `json:"fetchCompletionData" required:"true"`
	OnStreamTextData     func(textData string) error     `json:"-"`
	OnStreamThinkingData func(thinkingData string) error `json:"-"`
}

type FetchCompletionRequest struct {
	Provider modelpresetSpec.ProviderName `path:"provider" required:"true"`
	Body     *FetchCompletionRequestBody
}

type FetchCompletionResponseBody struct {
	RequestDetails  *APIRequestDetails  `json:"requestDetails,omitempty"`
	ResponseDetails *APIResponseDetails `json:"responseDetails,omitempty"`
	ErrorDetails    *APIErrorDetails    `json:"errorDetails,omitempty"`
	ResponseContent []ResponseContent   `json:"responseContent,omitempty"`
	ToolCalls       []ResponseToolCall  `json:"toolCalls,omitempty"`
}

type FetchCompletionResponse struct {
	Body *FetchCompletionResponseBody
}
