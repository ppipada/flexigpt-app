package aiprovider

import (
	"context"
	"fmt"

	"github.com/flexigpt/flexiui/pkggo/aiprovider/anthropic"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/google"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/huggingface"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/llamacpp"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/openai"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/spec"
)

var AllAIProviders = map[spec.ProviderName]spec.ProviderInfo{
	anthropic.ProviderNameAnthropic:     anthropic.AnthropicProviderInfo,
	google.ProviderNameGoogle:           google.GoogleProviderInfo,
	huggingface.ProviderNameHuggingFace: huggingface.HuggingfaceProviderInfo,
	llamacpp.ProviderNameLlamaCPP:       llamacpp.LlamacppProviderInfo,
	openai.ProviderNameOpenAI:           openai.OpenAIProviderInfo,
}

var AllModelInfo = map[spec.ModelName]spec.ModelInfo{
	anthropic.CLAUDE_3_5_SONNET:              anthropic.AnthropicModels[anthropic.CLAUDE_3_5_SONNET],
	anthropic.CLAUDE_3_OPUS:                  anthropic.AnthropicModels[anthropic.CLAUDE_3_OPUS],
	anthropic.CLAUDE_3_SONNET:                anthropic.AnthropicModels[anthropic.CLAUDE_3_SONNET],
	anthropic.CLAUDE_3_HAIKU:                 anthropic.AnthropicModels[anthropic.CLAUDE_3_HAIKU],
	google.GEMINI_1_5_FLASH:                  google.GoogleModels[google.GEMINI_1_5_FLASH],
	google.GEMINI_1_5_PRO:                    google.GoogleModels[google.GEMINI_1_5_PRO],
	huggingface.DEEPSEEK_CODER_1_3B_INSTRUCT: huggingface.HuggingfaceModels[huggingface.DEEPSEEK_CODER_1_3B_INSTRUCT],
	llamacpp.LLAMA_3:                         llamacpp.LlamacppModels[llamacpp.LLAMA_3],
	llamacpp.LLAMA_3_1:                       llamacpp.LlamacppModels[llamacpp.LLAMA_3_1],
	openai.GPT_O1:                            openai.OpenAIModels[openai.GPT_O1],
	openai.GPT_O1_PREVIEW:                    openai.OpenAIModels[openai.GPT_O1_PREVIEW],
	openai.GPT_O1_MINI:                       openai.OpenAIModels[openai.GPT_O1_MINI],
	openai.GPT_4O:                            openai.OpenAIModels[openai.GPT_4O],
	openai.GPT_4:                             openai.OpenAIModels[openai.GPT_4],
	openai.GPT_3_5_TURBO:                     openai.OpenAIModels[openai.GPT_3_5_TURBO],
	openai.GPT_4O_MINI:                       openai.OpenAIModels[openai.GPT_4O_MINI],
}

// Define the ProviderSetAPI struct
type ProviderSetAPI struct {
	defaultProvider spec.ProviderName
	providers       map[spec.ProviderName]spec.CompletionProvider
}

// NewProviderSetAPI creates a new ProviderSet with the specified default provider
func NewProviderSetAPI(defaultProvider spec.ProviderName) (*ProviderSetAPI, error) {
	_, exists := AllAIProviders[defaultProvider]
	if !exists {
		return nil, fmt.Errorf("Invalid provider")
	}
	return &ProviderSetAPI{
		defaultProvider: defaultProvider,
		providers: map[spec.ProviderName]spec.CompletionProvider{
			anthropic.ProviderNameAnthropic:     anthropic.NewAnthropicAPI(),
			google.ProviderNameGoogle:           google.NewGoogleAPI(),
			huggingface.ProviderNameHuggingFace: huggingface.NewHuggingFaceAPI(),
			llamacpp.ProviderNameLlamaCPP:       llamacpp.NewLlamaCPPAPI(),
			openai.ProviderNameOpenAI:           openai.NewOpenAIAPI(),
		},
	}, nil
}

// GetDefaultProvider returns the default provider
func (ps *ProviderSetAPI) GetDefaultProvider(
	ctx context.Context,
	req *spec.GetDefaultProviderRequest,
) (*spec.GetDefaultProviderResponse, error) {
	return &spec.GetDefaultProviderResponse{
		Body: &spec.GetDefaultProviderResponseBody{DefaultProvider: ps.defaultProvider},
	}, nil
}

// SetDefaultProvider sets the default provider
func (ps *ProviderSetAPI) SetDefaultProvider(
	ctx context.Context,
	req *spec.SetDefaultProviderRequest,
) (*spec.SetDefaultProviderResponse, error) {
	if req == nil || req.Body == nil {
		return nil, fmt.Errorf("Got empty provider input")
	}
	_, exists := ps.providers[req.Body.Provider]
	if !exists {
		return nil, fmt.Errorf("Invalid provider")
	}
	ps.defaultProvider = req.Body.Provider
	return &spec.SetDefaultProviderResponse{}, nil
}

// GetConfigurationInfo returns configuration information
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
		if provider, exists := ps.providers[spec.ProviderName(providerInfo.Name)]; exists &&
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

// SetProviderAttribute sets attributes for a given provider
func (ps *ProviderSetAPI) SetProviderAttribute(
	ctx context.Context,
	req *spec.SetProviderAttributeRequest,
) (*spec.SetProviderAttributeResponse, error) {
	if req == nil || req.Body == nil {
		return nil, fmt.Errorf("Got empty provider input")
	}
	p, exists := ps.providers[req.Provider]
	if !exists {
		return nil, fmt.Errorf("Invalid provider")
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

// MakeCompletion creates a completion request for a given provider
func (ps *ProviderSetAPI) MakeCompletion(
	ctx context.Context,
	req *spec.MakeCompletionRequest,
) (*spec.MakeCompletionResponse, error) {
	if req == nil || req.Body == nil {
		return nil, fmt.Errorf("Got empty provider input")
	}
	provider := req.Provider
	p, exists := ps.providers[provider]
	if !exists {
		return nil, fmt.Errorf("Invalid provider")
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
				return nil, fmt.Errorf("'model' in inputParams must be a string")
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
	// slog.Debug(
	// 	"GetCompletionAPI",
	// 	"Input Parameters",
	// 	fmt.Sprintf("%+v", inputParams),
	// 	"Model Info",
	// 	fmt.Sprintf("%+v", modelInfo),
	// )

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

// FetchCompletion processes a completion request for a given provider
func (ps *ProviderSetAPI) FetchCompletion(
	ctx context.Context,
	req *spec.FetchCompletionRequest,
) (*spec.FetchCompletionResponse, error) {
	if req == nil || req.Body == nil || req.Body.Input == nil {
		return nil, fmt.Errorf("Got empty provider input")
	}
	provider := req.Body.Provider
	p, exists := ps.providers[provider]
	if !exists {
		return nil, fmt.Errorf("Invalid provider")
	}

	resp, err := p.FetchCompletion(ctx, *req.Body.Input, req.Body.OnStreamData)
	if err != nil {
		return nil, fmt.Errorf("Error in fetch completion")
	}

	return &spec.FetchCompletionResponse{Body: resp}, nil
}
