package spec

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

type (
	ModelName    string
	ProviderName string
	ProviderType string
)

const (
	InbuiltSpecific         ProviderType = "inbuiltSpecific"
	InbuiltOpenAICompatible ProviderType = "inbuiltOpenAICompatible"
	CustomOpenAICompatible  ProviderType = "customOpenAICompatible"
)

// ProviderInfo represents information about a provider.
type ProviderInfo struct {
	Name                     ProviderName      `json:"name"`
	APIKey                   string            `json:"apiKey"`
	Origin                   string            `json:"origin"`
	ChatCompletionPathPrefix string            `json:"chatCompletionPathPrefix"`
	APIKeyHeaderKey          string            `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string `json:"defaultHeaders"`
	Type                     ProviderType      `json:"type"`
}

func (p *ProviderInfo) IsConfigured() bool {
	return p.APIKey != ""
}

func Float64Ptr(f float64) *float64 {
	return &f
}

type CompletionProvider interface {
	GetProviderInfo(
		ctx context.Context,
	) *ProviderInfo
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
		modelParams ModelParams,
		inbuiltModelParams *ModelParams,
		prevMessages []ChatCompletionRequestMessage,
		onStreamData func(data string) error,
	) (*CompletionResponse, error)
}
