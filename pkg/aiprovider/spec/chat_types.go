package spec

import "context"

type ChatCompletionRoleEnum string

const (
	System    ChatCompletionRoleEnum = "system"
	Developer ChatCompletionRoleEnum = "developer"
	User      ChatCompletionRoleEnum = "user"
	Assistant ChatCompletionRoleEnum = "assistant"
	Function  ChatCompletionRoleEnum = "function"
)

type ChatCompletionFunctions struct {
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type ChatCompletionRequestMessageFunctionCall struct {
	Name      *string `json:"name,omitempty"`
	Arguments *string `json:"arguments,omitempty"`
}

type ChatCompletionRequestMessage struct {
	Role         ChatCompletionRoleEnum                    `json:"role"`
	Content      *string                                   `json:"content,omitempty"`
	Name         *string                                   `json:"name,omitempty"`
	FunctionCall *ChatCompletionRequestMessageFunctionCall `json:"functionCall,omitempty"`
}

type ChatCompletionResponseMessage struct {
	Role         ChatCompletionRoleEnum                    `json:"role"`
	Content      *string                                   `json:"content,omitempty"`
	FunctionCall *ChatCompletionRequestMessageFunctionCall `json:"functionCall,omitempty"`
}

type CreateChatCompletionRequestFunctionCall any

type CreateChatCompletionRequestFunctionCallOneOf struct {
	Name string `json:"name"`
}

// ModelParams represents input information about a model to a completion.
type ModelParams struct {
	Name         ModelName `json:"name"`
	Stream       *bool     `json:"stream"`
	PromptLength *int      `json:"promptLength,omitempty"`
	OutputLength *int      `json:"outputLength,omitempty"`
	Temperature  *float64  `json:"temperature,omitempty"`

	ReasoningSupport     *bool           `json:"reasoningSupport,omitempty"`
	SystemPrompt         *string         `json:"systemPrompt,omitempty"`
	Timeout              *int            `json:"timeout,omitempty"`
	AdditionalParameters *map[string]any `json:"additionalParameters,omitempty"`
}

type CompletionRequest struct {
	ModelParams  ModelParams                             `json:"modelParams"`
	Messages     []ChatCompletionRequestMessage          `json:"messages,omitempty"`
	Functions    []ChatCompletionFunctions               `json:"functions,omitempty"`
	FunctionCall CreateChatCompletionRequestFunctionCall `json:"functionCall,omitempty"`
}

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

type CompletionProvider interface {
	CreateCompletionRequest(
		ctx context.Context,
		prompt string,
		modelParams ModelParams,
		prevMessages []ChatCompletionRequestMessage,
	) (*CompletionRequest, error)
	FetchCompletion(
		ctx context.Context,
		input CompletionRequest,
		onStreamData func(data string) error,
	) (*CompletionResponse, error)
	SetProviderAttribute(
		ctx context.Context,
		apiKey *string,
		defaultModel *string,
		origin *string,
	) error
	IsConfigured(ctx context.Context) bool
}
