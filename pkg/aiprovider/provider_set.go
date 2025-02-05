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

var AllModelInfo = map[spec.ModelName]spec.ModelInfo{
	anthropic.Claude35Sonnet:             anthropic.AnthropicModels[anthropic.Claude35Sonnet],
	anthropic.Claude35Haiku:              anthropic.AnthropicModels[anthropic.Claude35Haiku],
	anthropic.Claude3Opus:                anthropic.AnthropicModels[anthropic.Claude3Opus],
	anthropic.Claude3Sonnet:              anthropic.AnthropicModels[anthropic.Claude3Sonnet],
	anthropic.Claude3Haiku:               anthropic.AnthropicModels[anthropic.Claude3Haiku],
	deepseek.DeepseekChat:                deepseek.DeepseekModels[deepseek.DeepseekChat],
	deepseek.DeepseekReasoner:            deepseek.DeepseekModels[deepseek.DeepseekReasoner],
	google.Gemini2FlashExp:               google.GoogleModels[google.Gemini2FlashExp],
	google.Gemini15Flash:                 google.GoogleModels[google.Gemini15Flash],
	google.Gemini15Pro:                   google.GoogleModels[google.Gemini15Pro],
	huggingface.DeepseekCoder13BInstruct: huggingface.HuggingfaceModels[huggingface.DeepseekCoder13BInstruct],
	llamacpp.Llama3:                      llamacpp.LlamacppModels[llamacpp.Llama3],
	llamacpp.Llama31:                     llamacpp.LlamacppModels[llamacpp.Llama31],
	openai.GPTO3Mini:                     openai.OpenAIModels[openai.GPTO3Mini],
	openai.GPTO1:                         openai.OpenAIModels[openai.GPTO1],
	openai.GPTO1Preview:                  openai.OpenAIModels[openai.GPTO1Preview],
	openai.GPTO1Mini:                     openai.OpenAIModels[openai.GPTO1Mini],
	openai.GPT4O:                         openai.OpenAIModels[openai.GPT4O],
	openai.GPT4:                          openai.OpenAIModels[openai.GPT4],
	openai.GPT35Turbo:                    openai.OpenAIModels[openai.GPT35Turbo],
	openai.GPT4OMini:                     openai.OpenAIModels[openai.GPT4OMini],
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
	configurationInfo := map[string]interface{}{
		"defaultProvider": ps.defaultProvider,
	}
	configuredProviders := []spec.ProviderInfo{}
	configuredModels := []spec.ModelInfo{}

	for _, providerInfo := range AllAIProviders {
		if provider, exists := ps.providers[providerInfo.Name]; exists &&
			provider.IsConfigured(ctx) {
			configuredProviders = append(configuredProviders, providerInfo)
			for _, modelInfo := range AllModelInfo {
				if modelInfo.Provider == providerInfo.Name {
					configuredModels = append(configuredModels, modelInfo)
				}
			}
		}
	}

	configurationInfo["configuredProviders"] = configuredProviders
	configurationInfo["configuredModels"] = configuredModels
	return &spec.GetConfigurationInfoResponse{
		Body: &spec.GetConfigurationInfoResponseBody{ConfigurationInfo: configurationInfo},
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
		req.Body.DefaultOrigin,
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

	defaultModelName := aiProviderInfo.DefaultModel
	// Retrieve the model info for the default model
	modelInfo, ok := AllModelInfo[defaultModelName]
	if !ok {
		return nil, fmt.Errorf("model info for default model '%s' not found", defaultModelName)
	}

	// Check if 'model' is specified in inputParams and is a valid model name
	if req.Body.InputParams != nil {
		if modelInterface, ok := req.Body.InputParams["model"]; ok {
			// Ensure the model is a string
			modelStr, ok := modelInterface.(string)
			if !ok {
				return nil, errors.New("'model' in inputParams must be a string")
			}

			modelName := spec.ModelName(modelStr)

			// Verify that the model exists and belongs to the provider
			if modelEntry, exists := AllModelInfo[modelName]; exists &&
				modelEntry.Provider == provider {
				modelInfo = modelEntry
			} else {
				return nil, fmt.Errorf("invalid model '%s' specified for provider '%s'", modelName, provider)
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
