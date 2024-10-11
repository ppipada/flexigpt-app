package spec

import "context"

type ProviderSetAPI interface {
	GetDefaultProvider(ctx context.Context) (ProviderName, error)
	SetDefaultProvider(ctx context.Context, provider ProviderName) error
	GetConfigurationInfo(ctx context.Context) (map[string]interface{}, error)
	GetProviderInfo(ctx context.Context, provider ProviderName) (ProviderInfo, error)
	SetAttribute(ctx context.Context, provider ProviderName, apiKey *string, defaultModel *ModelName, defaultTemperature *float64, defaultOrigin *string) error

	GetCompletionRequest(ctx context.Context, provider ProviderName, prompt string, prevMessages []ChatCompletionRequestMessage, inputParams map[string]interface{}, stream bool) (CompletionRequest, error)
	Completion(ctx context.Context, provider ProviderName, input CompletionRequest, onStreamData func(data string) error) (*CompletionResponse, error)
}
