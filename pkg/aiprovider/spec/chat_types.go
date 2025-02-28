package spec

type ChatCompletionRoleEnum string

const (
	System    ChatCompletionRoleEnum = "system"
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

type CompletionRequest struct {
	Model                string                                  `json:"model"`
	Messages             []ChatCompletionRequestMessage          `json:"messages,omitempty"`
	Temperature          float64                                 `json:"temperature"`
	MaxPromptLength      int                                     `json:"maxPromptLength"`
	Stream               bool                                    `json:"stream"`
	SystemPrompt         *string                                 `json:"systemPrompt,omitempty"`
	MaxOutputLength      *int                                    `json:"maxOutputLength,omitempty"`
	Functions            []ChatCompletionFunctions               `json:"functions,omitempty"`
	FunctionCall         CreateChatCompletionRequestFunctionCall `json:"functionCall,omitempty"`
	Suffix               *string                                 `json:"suffix,omitempty"`
	Timeout              *int                                    `json:"timeout,omitempty"`
	AdditionalParameters map[string]any                          `json:"additionalParameters,omitempty"`
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
