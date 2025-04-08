package aiprovider

import (
	"context"
	"errors"
	"log/slog"

	"github.com/flexigpt/flexiui/pkg/aiprovider/anthropic"
	"github.com/flexigpt/flexiui/pkg/aiprovider/deepseek"
	"github.com/flexigpt/flexiui/pkg/aiprovider/google"
	"github.com/flexigpt/flexiui/pkg/aiprovider/huggingface"
	"github.com/flexigpt/flexiui/pkg/aiprovider/llamacpp"
	"github.com/flexigpt/flexiui/pkg/aiprovider/openai"
	"github.com/flexigpt/flexiui/pkg/aiprovider/openaicompat"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"
)

func getInbuiltProviderInfo() map[spec.ProviderName]spec.ProviderInfo {
	return map[spec.ProviderName]spec.ProviderInfo{
		anthropic.ProviderNameAnthropic:     anthropic.AnthropicProviderInfo,
		deepseek.ProviderNameDeepseek:       deepseek.DeepseekProviderInfo,
		google.ProviderNameGoogle:           google.GoogleProviderInfo,
		huggingface.ProviderNameHuggingFace: huggingface.HuggingfaceProviderInfo,
		llamacpp.ProviderNameLlamaCPP:       llamacpp.LlamacppProviderInfo,
		openai.ProviderNameOpenAI:           openai.OpenAIProviderInfo,
	}
}

func getInbuiltProviderAPI(debug bool) map[spec.ProviderName]spec.CompletionProvider {
	return map[spec.ProviderName]spec.CompletionProvider{
		anthropic.ProviderNameAnthropic: anthropic.NewAnthropicAPI(debug),
		deepseek.ProviderNameDeepseek: openaicompat.NewOpenAICompatibleProvider(
			deepseek.DeepseekProviderInfo,
			debug,
		),
		google.ProviderNameGoogle: openaicompat.NewOpenAICompatibleProvider(
			google.GoogleProviderInfo,
			debug,
		),
		huggingface.ProviderNameHuggingFace: huggingface.NewHuggingFaceAPI(),
		llamacpp.ProviderNameLlamaCPP: openaicompat.NewOpenAICompatibleProvider(
			llamacpp.LlamacppProviderInfo,
			debug,
		),
		openai.ProviderNameOpenAI: openaicompat.NewOpenAICompatibleProvider(
			openai.OpenAIProviderInfo,
			debug,
		),
	}
}

// Define the ProviderSetAPI struct.
type ProviderSetAPI struct {
	defaultProvider spec.ProviderName
	providers       map[spec.ProviderName]spec.CompletionProvider
	debug           bool
}

// NewProviderSetAPI creates a new ProviderSet with the specified default provider.
func NewProviderSetAPI(
	defaultInbuiltProvider spec.ProviderName,
	debug bool,
) (*ProviderSetAPI, error) {
	inbuiltProviderSet := getInbuiltProviderInfo()
	_, exists := inbuiltProviderSet[defaultInbuiltProvider]
	if !exists {
		return nil, errors.New("invalid inbuilt provider")
	}
	return &ProviderSetAPI{
		defaultProvider: defaultInbuiltProvider,
		providers:       getInbuiltProviderAPI(debug),
		debug:           debug,
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

	for _, providerInfo := range getInbuiltProviderInfo() {
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

// AddProvider adds a custom provider.
// A provider with same name as inbuilt providers are not allowed.
func (ps *ProviderSetAPI) AddProvider(
	ctx context.Context,
	req *spec.AddProviderRequest,
) (*spec.AddProviderResponse, error) {
	if req == nil || req.Provider == "" {
		return nil, errors.New("got empty provider input")
	}
	_, exists := ps.providers[req.Provider]
	if exists {
		return nil, errors.New(
			"invalid provider: cannot add a provider with same name as an existing provider",
		)
	}

	providerInfo := spec.ProviderInfo{
		Name:         req.Provider,
		APIKey:       req.Body.APIKey,
		Origin:       req.Body.Origin,
		Type:         spec.CustomOpenAICompatible,
		DefaultModel: "",

		APIKeyHeaderKey:          openaicompat.APIKeyHeaderKey,
		DefaultHeaders:           openaicompat.DefaultHeaders,
		ChatCompletionPathPrefix: req.Body.ChatCompletionPathPrefix,
		Models:                   map[spec.ModelName]spec.ModelInfo{},
	}

	ps.providers[req.Provider] = openaicompat.NewOpenAICompatibleProvider(
		providerInfo,
		ps.debug,
	)
	slog.Info("Added", "Provider", req.Provider)
	return &spec.AddProviderResponse{}, nil
}

func (ps *ProviderSetAPI) DeleteProvider(
	ctx context.Context,
	req *spec.DeleteProviderRequest,
) (*spec.DeleteProviderResponse, error) {
	if req == nil || req.Provider == "" {
		return nil, errors.New("got empty provider input")
	}
	papi, exists := ps.providers[req.Provider]
	if !exists {
		return nil, errors.New(
			"invalid provider: provider does not exist",
		)
	}

	pinfo := papi.GetProviderInfo(ctx)
	if pinfo.Type == spec.InbuiltOpenAICompatible || pinfo.Type == spec.InbuiltSpecific {
		return nil, errors.New(
			"invalid provider: cannot delete inbuilt provider",
		)
	}
	delete(ps.providers, req.Provider)
	slog.Info("Deleted", "Provider", req.Provider)
	return &spec.DeleteProviderResponse{}, nil
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
		req.Body.Origin,
		req.Body.ChatCompletionPathPrefix,
	)
	return &spec.SetProviderAttributeResponse{}, err
}

// FetchCompletion processes a completion request for a given provider.
func (ps *ProviderSetAPI) FetchCompletion(
	ctx context.Context,
	req *spec.FetchCompletionRequest,
) (*spec.FetchCompletionResponse, error) {
	if req == nil || req.Body == nil || req.Body.Prompt == "" || req.Body.ModelParams.Name == "" {
		return nil, errors.New("got empty provider/prompt/model input")
	}
	provider := req.Body.Provider
	p, exists := ps.providers[provider]
	if !exists {
		return nil, errors.New("invalid provider")
	}
	resp, err := p.FetchCompletion(
		ctx,
		req.Body.Prompt,
		req.Body.ModelParams,
		req.Body.PrevMessages,
		req.Body.OnStreamData,
	)
	if err != nil {
		return nil, errors.Join(err, errors.New("error in fetch completion"))
	}

	return &spec.FetchCompletionResponse{Body: resp}, nil
}
