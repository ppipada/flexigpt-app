package api

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"
	"github.com/tmc/langchaingo/llms"
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
	ModelParams  spec.ModelParams                             `json:"modelParams"`
	Messages     []spec.ChatCompletionRequestMessage          `json:"messages,omitempty"`
	Functions    []spec.ChatCompletionFunctions               `json:"functions,omitempty"`
	FunctionCall spec.CreateChatCompletionRequestFunctionCall `json:"functionCall,omitempty"`
}

type CompletionProvider interface {
	GetProviderInfo(
		ctx context.Context,
	) *spec.ProviderInfo
	IsConfigured(ctx context.Context) bool
	GetLLMsModel(ctx context.Context) llms.Model
	InitLLM(ctx context.Context) error
	SetProviderAPIKey(
		ctx context.Context,
		apiKey string,
	) error
	SetProviderAttribute(
		ctx context.Context,
		origin *string,
		chatCompletionPathPrefix *string,
	) error
	FetchCompletion(
		ctx context.Context,
		llm llms.Model,
		prompt string,
		modelParams spec.ModelParams,
		inbuiltModelParams *spec.ModelParams,
		prevMessages []spec.ChatCompletionRequestMessage,
		onStreamData func(data string) error,
	) (*CompletionResponse, error)
}
