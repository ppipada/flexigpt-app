package aiprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/flexigpt/flexiui/pkg/aiprovider/anthropic"
	"github.com/flexigpt/flexiui/pkg/aiprovider/deepseek"
	"github.com/flexigpt/flexiui/pkg/aiprovider/google"
	"github.com/flexigpt/flexiui/pkg/aiprovider/huggingface"
	"github.com/flexigpt/flexiui/pkg/aiprovider/llamacpp"
	"github.com/flexigpt/flexiui/pkg/aiprovider/openai"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

var AllAIProviders = map[spec.ProviderName]spec.ProviderInfo{
	anthropic.ProviderNameAnthropic:     anthropic.AnthropicProviderInfo,
	deepseek.ProviderNameDeepseek:       deepseek.DeepseekProviderInfo,
	google.ProviderNameGoogle:           google.GoogleProviderInfo,
	huggingface.ProviderNameHuggingFace: huggingface.HuggingfaceProviderInfo,
	llamacpp.ProviderNameLlamaCPP:       llamacpp.LlamacppProviderInfo,
	openai.ProviderNameOpenAI:           openai.OpenAIProviderInfo,
}

// Define the ProviderSetAPI struct.
type ProviderSetAPI struct {
	defaultProvider spec.ProviderName
	providers       map[spec.ProviderName]spec.CompletionProvider
}

// NewProviderSetAPI creates a new ProviderSet with the specified default provider.
func NewProviderSetAPI(defaultProvider spec.ProviderName) (*ProviderSetAPI, error) {
	_, exists := AllAIProviders[defaultProvider]
	if !exists {
		return nil, errors.New("invalid provider")
	}
	return &ProviderSetAPI{
		defaultProvider: defaultProvider,
		providers: map[spec.ProviderName]spec.CompletionProvider{
			anthropic.ProviderNameAnthropic: anthropic.NewAnthropicAPI(),
			deepseek.ProviderNameDeepseek: openai.NewOpenAICompatibleProvider(
				deepseek.DeepseekProviderInfo,
				false,
			),
			google.ProviderNameGoogle: openai.NewOpenAICompatibleProvider(
				google.GoogleProviderInfo,
				false,
			),
			huggingface.ProviderNameHuggingFace: huggingface.NewHuggingFaceAPI(),
			llamacpp.ProviderNameLlamaCPP:       llamacpp.NewLlamaCPPAPI(),
			openai.ProviderNameOpenAI: openai.NewOpenAICompatibleProvider(
				openai.OpenAIProviderInfo,
				false,
			),
		},
	}, nil
}

// GetDefaultProvider returns the default provider.
func (ps *ProviderSetAPI) GetDefaultProvider(
	ctx context.Context,
	req *spec.GetDefaultProviderRequest,
) (*spec.GetDefaultProviderResponse, error) {
	return &spec.GetDefaultProviderResponse{
		Body: &spec.GetDefaultProviderResponseBody{DefaultProvider: ps.defaultProvider},
	}, nil
}

// SetDefaultProvider sets the default provider.
func (ps *ProviderSetAPI) SetDefaultProvider(
	ctx context.Context,
	req *spec.SetDefaultProviderRequest,
) (*spec.SetDefaultProviderResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("got empty provider input")
	}
	_, exists := ps.providers[req.Body.Provider]
	if !exists {
		return nil, errors.New("invalid provider")
	}
	ps.defaultProvider = req.Body.Provider
	return &spec.SetDefaultProviderResponse{}, nil
}

// GetConfigurationInfo returns configuration information.
func (ps *ProviderSetAPI) GetConfigurationInfo(
	ctx context.Context,
	req *spec.GetConfigurationInfoRequest,
) (*spec.GetConfigurationInfoResponse, error) {
	configuredProviders := []spec.ProviderInfo{}

	for _, providerInfo := range AllAIProviders {
		if provider, exists := ps.providers[providerInfo.Name]; exists &&
			provider.IsConfigured(ctx) {
			configuredProviders = append(configuredProviders, providerInfo)
		}
	}
	return &spec.GetConfigurationInfoResponse{
		Body: &spec.GetConfigurationInfoResponseBody{
			DefaultProvider:     ps.defaultProvider,
			ConfiguredProviders: configuredProviders,
		},
	}, nil
}

// SetProviderAttribute sets attributes for a given provider.
func (ps *ProviderSetAPI) SetProviderAttribute(
	ctx context.Context,
	req *spec.SetProviderAttributeRequest,
) (*spec.SetProviderAttributeResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("got empty provider input")
	}
	p, exists := ps.providers[req.Provider]
	if !exists {
		return nil, errors.New("invalid provider")
	}

	err := p.SetProviderAttribute(
		ctx,
		req.Body.APIKey,
		req.Body.DefaultModel,
		req.Body.DefaultTemperature,
		req.Body.Origin,
	)
	return &spec.SetProviderAttributeResponse{}, err
}

// MakeCompletion creates a completion request for a given provider.
func (ps *ProviderSetAPI) MakeCompletion(
	ctx context.Context,
	req *spec.MakeCompletionRequest,
) (*spec.MakeCompletionResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("got empty provider input")
	}
	provider := req.Provider
	p, exists := ps.providers[provider]
	if !exists {
		return nil, errors.New("invalid provider")
	}

	// Get the default model for the provider
	aiProviderInfo, ok := AllAIProviders[provider]
	if !ok {
		return nil, fmt.Errorf("AI provider '%s' not found", provider)
	}

	// Retrieve the model info for the default model
	modelInfo, ok := aiProviderInfo.Models[aiProviderInfo.DefaultModel]
	if !ok {
		return nil, fmt.Errorf(
			"model info for default model '%s' not found",
			aiProviderInfo.DefaultModel,
		)
	}

	// Check if 'model' is specified in inputParams and is a valid model name
	if req.Body.InputParams != nil {
		if modelInterface, ok := req.Body.InputParams["model"]; ok {
			// Ensure the model is a string
			modelStr, ok := modelInterface.(string)
			if !ok {
				return nil, errors.New("'model' in inputParams must be a string")
			}

			// Verify that the model exists and belongs to the provider
			if modelEntry, exists := aiProviderInfo.Models[spec.ModelName(modelStr)]; exists {
				modelInfo = modelEntry
			} else {
				return nil, fmt.Errorf("invalid model '%s' specified for provider '%s'", spec.ModelName(modelStr), provider)
			}
		}
	}

	// Create and return the completion request
	cr, err := p.MakeCompletion(
		ctx,
		modelInfo,
		req.Body.Prompt,
		req.Body.PrevMessages,
		req.Body.InputParams,
	)
	if err != nil {
		return nil, err
	}
	return &spec.MakeCompletionResponse{Body: cr}, nil
}

// FetchCompletion processes a completion request for a given provider.
func (ps *ProviderSetAPI) FetchCompletion(
	ctx context.Context,
	req *spec.FetchCompletionRequest,
) (*spec.FetchCompletionResponse, error) {
	if req == nil || req.Body == nil || req.Body.Input == nil {
		return nil, errors.New("got empty provider input")
	}
	provider := req.Body.Provider
	p, exists := ps.providers[provider]
	if !exists {
		return nil, errors.New("invalid provider")
	}
	resp, err := p.FetchCompletion(ctx, *req.Body.Input, req.Body.OnStreamData)
	if err != nil {
		return nil, errors.Join(err, errors.New("error in fetch completion"))
	}

	return &spec.FetchCompletionResponse{Body: resp}, nil
}
