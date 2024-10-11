package aiprovider

import (
	"context"
	"errors"

	"github.com/flexigpt/flexiui/pkg/aiprovider/openai"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

// Define the ProviderSetAPI struct
type ProviderSetAPI struct {
	defaultProvider spec.ProviderName
	providers       map[spec.ProviderName]spec.CompletionProvider
}

// NewProviderSetAPI creates a new ProviderSet with the specified default provider
func NewProviderSetAPI(defaultProvider spec.ProviderName) *ProviderSetAPI {
	return &ProviderSetAPI{
		defaultProvider: defaultProvider,
		providers: map[spec.ProviderName]spec.CompletionProvider{
			// spec.ProviderNameAnthropic:   &AnthropicAPI{},
			// spec.ProviderNameGoogle:      &GoogleAPI{},
			// spec.ProviderNameHuggingFace: &HuggingFaceAPI{},
			// spec.ProviderNameLlamaCPP:    &LlamaCPPAPI{},
			spec.ProviderNameOpenAI: openai.NewOpenAIAPI(),
		},
	}
}

// GetDefaultProvider returns the default provider
func (ps *ProviderSetAPI) GetDefaultProvider() spec.ProviderName {
	return ps.defaultProvider
}

// SetDefaultProvider sets the default provider
func (ps *ProviderSetAPI) SetDefaultProvider(provider spec.ProviderName) {
	ps.defaultProvider = provider
}

// GetConfigurationInfo returns configuration information
func (ps *ProviderSetAPI) GetConfigurationInfo() (map[string]interface{}, error) {
	ctx := context.Background()
	configurationInfo := map[string]interface{}{
		"defaultProvider": ps.defaultProvider,
	}
	configuredProviders := []spec.ProviderInfo{}
	configuredModels := []spec.ModelInfo{}

	for _, providerInfo := range spec.AllAIProviders {
		if provider, exists := ps.providers[spec.ProviderName(providerInfo.Name)]; exists &&
			provider.IsConfigured(ctx) {
			configuredProviders = append(configuredProviders, providerInfo)
			for _, modelInfo := range spec.AllModelInfo {
				if modelInfo.Provider == providerInfo.Name {
					configuredModels = append(configuredModels, modelInfo)
				}
			}
		}
	}

	configurationInfo["configuredProviders"] = configuredProviders
	configurationInfo["configuredModels"] = configuredModels
	return configurationInfo, nil
}

// GetProviderInfo returns the provider information for a given provider
func (ps *ProviderSetAPI) GetProviderInfo(
	provider spec.ProviderName,
) (*spec.ProviderInfo, error) {
	ctx := context.Background()
	if p, exists := ps.providers[provider]; exists {
		return p.GetProviderInfo(ctx)
	}
	return nil, errors.New("provider not found")
}

// SetProviderAttribute sets attributes for a given provider
func (ps *ProviderSetAPI) SetProviderAttribute(
	provider spec.ProviderName,
	apiKey *string,
	defaultModel *string,
	defaultTemperature *float64,
	defaultOrigin *string,
) error {
	ctx := context.Background()
	if p, exists := ps.providers[provider]; exists {
		return p.SetProviderAttribute(ctx, apiKey, defaultModel, defaultTemperature, defaultOrigin)
	}
	return errors.New("provider not found")
}

// GetCompletionRequest creates a completion request for a given provider
func (ps *ProviderSetAPI) GetCompletionRequest(
	provider spec.ProviderName,
	prompt string,
	prevMessages []spec.ChatCompletionRequestMessage,
	inputParams map[string]interface{},
	stream bool,
) (*spec.CompletionRequest, error) {
	ctx := context.Background()
	if p, exists := ps.providers[provider]; exists {
		return p.GetCompletionRequest(ctx, prompt, prevMessages, inputParams, stream)
	}
	return nil, errors.New("provider not found")
}

// FetchCompletion processes a completion request for a given provider
func (ps *ProviderSetAPI) FetchCompletion(
	provider spec.ProviderName,
	input spec.CompletionRequest,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	ctx := context.Background()
	if p, exists := ps.providers[provider]; exists {
		return p.FetchCompletion(ctx, input, onStreamData)
	}
	return nil, errors.New("provider not found")
}
