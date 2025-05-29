package spec

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

type ReasoningType string

const (
	ReasoningTypeHybridWithTokens ReasoningType = "hybridWithTokens"
	ReasoningTypeSingleWithLevels ReasoningType = "singleWithLevels"
)

type ReasoningLevel string

const (
	ReasoningLevelLow    ReasoningLevel = "low"
	ReasoningLevelMedium ReasoningLevel = "medium"
	ReasoningLevelHigh   ReasoningLevel = "high"
)

type ReasoningParams struct {
	Type   ReasoningType  `json:"type"`
	Level  ReasoningLevel `json:"level"`
	Tokens int            `json:"tokens"`
}

type ModelDefaults struct {
	DisplayName string `json:"displayName"`
	IsEnabled   bool   `json:"isEnabled"`
}

// ModelParams represents input information about a model to a completion.
type ModelParams struct {
	Name                 ModelName        `json:"name"`
	Stream               bool             `json:"stream"`
	MaxPromptLength      int              `json:"maxPromptLength"`
	MaxOutputLength      int              `json:"maxOutputLength"`
	Temperature          *float64         `json:"temperature,omitempty"`
	Reasoning            *ReasoningParams `json:"reasoning"`
	SystemPrompt         string           `json:"systemPrompt"`
	Timeout              int              `json:"timeout"`
	AdditionalParameters map[string]any   `json:"additionalParameters"`
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
