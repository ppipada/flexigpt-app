package spec

import (
	"github.com/ppipada/flexigpt-app/pkg/attachment"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"

	inferencegoSpec "github.com/ppipada/inference-go/spec"
)

type AddProviderRequestBody struct {
	SDKType                  inferencegoSpec.ProviderSDKType `json:"sdkType"`
	Origin                   string                          `json:"origin"`
	ChatCompletionPathPrefix string                          `json:"chatCompletionPathPrefix"`
	APIKeyHeaderKey          string                          `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string               `json:"defaultHeaders"`
}

type AddProviderRequest struct {
	Provider inferencegoSpec.ProviderName `path:"provider" required:"true"`
	Body     *AddProviderRequestBody
}

type AddProviderResponse struct{}

type DeleteProviderRequest struct {
	Provider inferencegoSpec.ProviderName `path:"provider" required:"true"`
}

type DeleteProviderResponse struct{}

type SetProviderAPIKeyRequestBody struct {
	APIKey string `json:"apiKey" required:"true"`
}

type SetProviderAPIKeyRequest struct {
	Provider inferencegoSpec.ProviderName `path:"provider" required:"true"`
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

type ChatCompletionDataMessage struct {
	Role    modelpresetSpec.RoleEnum `json:"role"`
	Name    *string                  `json:"name,omitempty"`
	Content *string                  `json:"content,omitempty"`

	ReasoningContents []modelpresetSpec.ReasoningContent `json:"reasoningContents,omitempty"`
	Citations         []modelpresetSpec.Citation         `json:"citations,omitempty"`
	Attachments       []attachment.Attachment            `json:"attachments,omitempty"`
	ToolCalls         []toolSpec.ToolCall                `json:"toolCalls,omitempty"`
	ToolOutputs       []toolSpec.ToolOutput              `json:"toolOutputs,omitempty"`
}

type FetchCompletionToolChoice struct {
	toolSpec.ToolChoice

	Tool *toolSpec.Tool `json:"tool"`
}

type BuildCompletionDataRequestBody struct {
	ModelParams    inferencegoSpec.ModelParam  `json:"modelParams"            required:"true"`
	CurrentMessage ChatCompletionDataMessage   `json:"currentMessage"         required:"true"`
	PrevMessages   []ChatCompletionDataMessage `json:"prevMessages,omitempty"`
	ToolChoices    []toolSpec.ToolChoice       `json:"toolChoices,omitempty"`
}

type BuildCompletionDataRequest struct {
	Provider inferencegoSpec.ProviderName `path:"provider" required:"true"`
	Body     *BuildCompletionDataRequestBody
}

type FetchCompletionData struct {
	ModelParams inferencegoSpec.ModelParam  `json:"modelParams"`
	Messages    []ChatCompletionDataMessage `json:"messages,omitempty"`
	ToolChoices []FetchCompletionToolChoice `json:"toolChoices,omitempty"`
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
	Provider inferencegoSpec.ProviderName `path:"provider" required:"true"`
	Body     *FetchCompletionRequestBody
}

type FetchCompletionResponseBody struct {
	RequestDetails  *APIRequestDetails  `json:"requestDetails,omitempty"`
	ResponseDetails *APIResponseDetails `json:"responseDetails,omitempty"`
	ErrorDetails    *APIErrorDetails    `json:"errorDetails,omitempty"`

	Content           *string                            `json:"content,omitempty"`
	ReasoningContents []modelpresetSpec.ReasoningContent `json:"reasoningContents,omitempty"`
	Citations         []modelpresetSpec.Citation         `json:"citations,omitempty"`
	ToolCalls         []toolSpec.ToolCall                `json:"toolCalls,omitempty"`
	Usage             *modelpresetSpec.Usage             `json:"usage,omitempty"`
}

type FetchCompletionResponse struct {
	Body *FetchCompletionResponseBody
}
