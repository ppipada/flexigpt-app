package spec

import (
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

type ChatCompletionRoleEnum string

const (
	System    ChatCompletionRoleEnum = "system"
	Developer ChatCompletionRoleEnum = "developer"
	User      ChatCompletionRoleEnum = "user"
	Assistant ChatCompletionRoleEnum = "assistant"
	Function  ChatCompletionRoleEnum = "function"
	Tool      ChatCompletionRoleEnum = "tool"
)

// ModelParams represents input information about a model to a completion.
type ModelParams struct {
	Name                        modelpresetSpec.ModelName        `json:"name"`
	Stream                      bool                             `json:"stream"`
	MaxPromptLength             int                              `json:"maxPromptLength"`
	MaxOutputLength             int                              `json:"maxOutputLength"`
	Temperature                 *float64                         `json:"temperature,omitempty"`
	Reasoning                   *modelpresetSpec.ReasoningParams `json:"reasoning,omitempty"`
	SystemPrompt                string                           `json:"systemPrompt"`
	Timeout                     int                              `json:"timeout"`
	AdditionalParametersRawJSON *string                          `json:"additionalParametersRawJSON"`
}

// ProviderParams represents information about a provider.
type ProviderParams struct {
	Name                     modelpresetSpec.ProviderName    `json:"name"`
	SDKType                  modelpresetSpec.ProviderSDKType `json:"sdkType"`
	APIKey                   string                          `json:"apiKey"`
	Origin                   string                          `json:"origin"`
	ChatCompletionPathPrefix string                          `json:"chatCompletionPathPrefix"`
	APIKeyHeaderKey          string                          `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string               `json:"defaultHeaders"`
}
