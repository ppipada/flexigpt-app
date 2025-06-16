package inference

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
