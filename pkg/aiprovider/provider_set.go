package aiprovider

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ppipada/flexigpt-app/pkg/aiprovider/api"
	"github.com/ppipada/flexigpt-app/pkg/aiprovider/consts"
	"github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"
)

func getInbuiltProviderAPI(debug bool) map[spec.ProviderName]spec.CompletionProvider {
	return map[spec.ProviderName]spec.CompletionProvider{
		consts.ProviderNameAnthropic: api.NewAnthropicCompatibleAPI(
			consts.AnthropicProviderInfo,
			debug,
		),
		consts.ProviderNameDeepseek: api.NewOpenAICompatibleProvider(
			consts.DeepseekProviderInfo,
			debug,
		),
		consts.ProviderNameGoogle: api.NewOpenAICompatibleProvider(
			consts.GoogleProviderInfo,
			debug,
		),
		consts.ProviderNameHuggingFace: api.NewHuggingFaceCompatibleAPI(
			consts.HuggingfaceProviderInfo,
			debug,
		),
		consts.ProviderNameLlamaCPP: api.NewOpenAICompatibleProvider(
			consts.LlamacppProviderInfo,
			debug,
		),
		consts.ProviderNameOpenAI: api.NewOpenAICompatibleProvider(
			consts.OpenAIProviderInfo,
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
	_, exists := consts.InbuiltProviders[defaultInbuiltProvider]
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

	for _, providerAPI := range ps.providers {
		if providerAPI.IsConfigured(ctx) {
			configuredProviders = append(configuredProviders, *providerAPI.GetProviderInfo(ctx))
		}
	}
	return &spec.GetConfigurationInfoResponse{
		Body: &spec.GetConfigurationInfoResponseBody{
			DefaultProvider:              ps.defaultProvider,
			ConfiguredProviders:          configuredProviders,
			InbuiltProviderModels:        consts.InbuiltProviderModels,
			InbuiltProviderModelDefaults: consts.InbuiltProviderModelDefaults,
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
		Name:                     req.Provider,
		Type:                     spec.CustomOpenAICompatible,
		APIKeyHeaderKey:          consts.OpenAICompatibleAPIKeyHeaderKey,
		DefaultHeaders:           consts.OpenAICompatibleDefaultHeaders,
		APIKey:                   "",
		Origin:                   "",
		ChatCompletionPathPrefix: "",
	}

	ps.providers[req.Provider] = api.NewOpenAICompatibleProvider(
		providerInfo,
		ps.debug,
	)
	if req.Body.APIKey != "" {
		err := ps.providers[req.Provider].SetProviderAPIKey(ctx, req.Body.APIKey)
		if err != nil {
			return nil, err
		}
	}

	if req.Body.Origin != "" || req.Body.ChatCompletionPathPrefix != "" {
		err := ps.providers[req.Provider].SetProviderAttribute(
			ctx,
			&req.Body.Origin,
			&req.Body.ChatCompletionPathPrefix,
		)
		if err != nil {
			return nil, err
		}
	}

	err := ps.providers[req.Provider].InitLLM(ctx)
	if err != nil {
		return nil, err
	}
	slog.Info("AddProvider", "Name", req.Provider)
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
	slog.Info("DeleteProvider", "Name", req.Provider)
	return &spec.DeleteProviderResponse{}, nil
}

// SetProviderAPIKey sets the key for a given provider.
func (ps *ProviderSetAPI) SetProviderAPIKey(
	ctx context.Context,
	req *spec.SetProviderAPIKeyRequest,
) (*spec.SetProviderAPIKeyResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("got empty provider input")
	}
	p, exists := ps.providers[req.Provider]
	if !exists {
		return nil, errors.New("invalid provider")
	}

	err := p.SetProviderAPIKey(
		ctx,
		req.Body.APIKey,
	)
	if err != nil {
		return nil, err
	}
	err = p.InitLLM(ctx)
	if err != nil {
		return nil, err
	}
	return &spec.SetProviderAPIKeyResponse{}, nil
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
		req.Body.Origin,
		req.Body.ChatCompletionPathPrefix,
	)
	if err != nil {
		return nil, err
	}
	err = p.InitLLM(ctx)
	if err != nil {
		return nil, err
	}
	return &spec.SetProviderAttributeResponse{}, nil
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

	var inbuiltModelParams *spec.ModelParams = nil
	if pmodels, providerExists := consts.InbuiltProviderModels[provider]; providerExists {
		if params, paramExists := pmodels[req.Body.ModelParams.Name]; paramExists {
			inbuiltModelParams = &params
		}
	}

	resp, err := p.FetchCompletion(
		ctx,
		p.GetLLMsModel(ctx),
		req.Body.Prompt,
		req.Body.ModelParams,
		inbuiltModelParams,
		req.Body.PrevMessages,
		req.Body.OnStreamData,
	)
	if err != nil {
		return nil, errors.Join(err, errors.New("error in fetch completion"))
	}

	return &spec.FetchCompletionResponse{Body: resp}, nil
}
