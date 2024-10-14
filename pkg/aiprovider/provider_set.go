package aiprovider

import (
	"context"
	"fmt"

	"github.com/flexigpt/flexiui/pkg/aiprovider/anthropic"
	"github.com/flexigpt/flexiui/pkg/aiprovider/google"
	"github.com/flexigpt/flexiui/pkg/aiprovider/huggingface"
	"github.com/flexigpt/flexiui/pkg/aiprovider/llamacpp"
	"github.com/flexigpt/flexiui/pkg/aiprovider/openai"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
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
func NewProviderSetAPI(defaultProvider spec.ProviderName) *ProviderSetAPI {
	return &ProviderSetAPI{
		defaultProvider: defaultProvider,
		providers: map[spec.ProviderName]spec.CompletionProvider{
			anthropic.ProviderNameAnthropic:     anthropic.NewAnthropicAPI(),
			google.ProviderNameGoogle:           google.NewGoogleAPI(),
			huggingface.ProviderNameHuggingFace: huggingface.NewHuggingFaceAPI(),
			// spec.ProviderNameLlamaCPP:    &LlamaCPPAPI{},
			openai.ProviderNameOpenAI: openai.NewOpenAIAPI(),
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
	return nil, fmt.Errorf("provider not found")
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
	return fmt.Errorf("provider not found")
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
	p, exists := ps.providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider not found")
	}
	return p.GetCompletionRequest(
		ctx,
		AllModelInfo[AllAIProviders[provider].DefaultModel],
		prompt,
		prevMessages,
		inputParams,
		stream,
	)
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
	return nil, fmt.Errorf("provider not found")
}
