package spec

import "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"

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
	Name                        spec.ModelName        `json:"name"`
	Stream                      bool                  `json:"stream"`
	MaxPromptLength             int                   `json:"maxPromptLength"`
	MaxOutputLength             int                   `json:"maxOutputLength"`
	Temperature                 *float64              `json:"temperature,omitempty"`
	Reasoning                   *spec.ReasoningParams `json:"reasoning,omitempty"`
	SystemPrompt                string                `json:"systemPrompt"`
	Timeout                     int                   `json:"timeout"`
	AdditionalParametersRawJSON *string               `json:"additionalParametersRawJSON"`
}

// ProviderParams represents information about a provider.
type ProviderParams struct {
	Name                     spec.ProviderName    `json:"name"`
	APIType                  spec.ProviderAPIType `json:"apiType"`
	APIKey                   string               `json:"apiKey"`
	Origin                   string               `json:"origin"`
	ChatCompletionPathPrefix string               `json:"chatCompletionPathPrefix"`
	APIKeyHeaderKey          string               `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string    `json:"defaultHeaders"`
}
