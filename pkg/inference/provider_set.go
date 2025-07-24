package inference

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ppipada/flexigpt-app/pkg/model/consts"
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

// Define the ProviderSetAPI struct.
type ProviderSetAPI struct {
	defaultProvider spec.ProviderName
	providers       map[spec.ProviderName]CompletionProvider
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
	req *SetDefaultProviderRequest,
) (*SetDefaultProviderResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("got empty provider input")
	}
	_, exists := ps.providers[req.Body.Provider]
	if !exists {
		return nil, errors.New("invalid provider")
	}
	ps.defaultProvider = req.Body.Provider
	return &SetDefaultProviderResponse{}, nil
}

// GetConfigurationInfo returns configuration information.
func (ps *ProviderSetAPI) GetConfigurationInfo(
	ctx context.Context,
	req *GetConfigurationInfoRequest,
) (*GetConfigurationInfoResponse, error) {
	configuredProviders := []spec.ProviderInfo{}

	for _, providerAPI := range ps.providers {
		if providerAPI.IsConfigured(ctx) {
			configuredProviders = append(configuredProviders, *providerAPI.GetProviderInfo(ctx))
		}
	}
	return &GetConfigurationInfoResponse{
		Body: &GetConfigurationInfoResponseBody{
			DefaultProvider:       ps.defaultProvider,
			ConfiguredProviders:   configuredProviders,
			InbuiltProviderModels: consts.InbuiltProviderModels,
		},
	}, nil
}

// AddProvider adds a custom provider.
// A provider with same name as inbuilt providers are not allowed.
func (ps *ProviderSetAPI) AddProvider(
	ctx context.Context,
	req *AddProviderRequest,
) (*AddProviderResponse, error) {
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

	ps.providers[req.Provider] = NewOpenAICompatibleProvider(
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
	slog.Info("add provider", "name", req.Provider)
	return &AddProviderResponse{}, nil
}

func (ps *ProviderSetAPI) DeleteProvider(
	ctx context.Context,
	req *DeleteProviderRequest,
) (*DeleteProviderResponse, error) {
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
	slog.Info("deleteProvider", "name", req.Provider)
	return &DeleteProviderResponse{}, nil
}

// SetProviderAPIKey sets the key for a given provider.
func (ps *ProviderSetAPI) SetProviderAPIKey(
	ctx context.Context,
	req *SetProviderAPIKeyRequest,
) (*SetProviderAPIKeyResponse, error) {
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
	return &SetProviderAPIKeyResponse{}, nil
}

// SetProviderAttribute sets attributes for a given provider.
func (ps *ProviderSetAPI) SetProviderAttribute(
	ctx context.Context,
	req *SetProviderAttributeRequest,
) (*SetProviderAttributeResponse, error) {
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
	return &SetProviderAttributeResponse{}, nil
}

// FetchCompletion processes a completion request for a given provider.
func (ps *ProviderSetAPI) FetchCompletion(
	ctx context.Context,
	req *FetchCompletionRequest,
) (*FetchCompletionResponse, error) {
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
		p.GetLLMsModel(ctx),
		req.Body.Prompt,
		req.Body.ModelParams,
		req.Body.PrevMessages,
		req.Body.OnStreamData,
	)
	if err != nil {
		return nil, errors.Join(err, errors.New("error in fetch completion"))
	}

	return &FetchCompletionResponse{Body: resp}, nil
}

func getInbuiltProviderAPI(debug bool) map[spec.ProviderName]CompletionProvider {
	return map[spec.ProviderName]CompletionProvider{
		consts.ProviderNameAnthropic: NewAnthropicCompatibleAPI(
			consts.AnthropicProviderInfo,
			debug,
		),
		consts.ProviderNameDeepseek: NewOpenAICompatibleProvider(
			consts.DeepseekProviderInfo,
			debug,
		),
		consts.ProviderNameGoogle: NewOpenAICompatibleProvider(
			consts.GoogleProviderInfo,
			debug,
		),
		consts.ProviderNameHuggingFace: NewHuggingFaceCompatibleAPI(
			consts.HuggingfaceProviderInfo,
			debug,
		),
		consts.ProviderNameLlamaCPP: NewOpenAICompatibleProvider(
			consts.LlamacppProviderInfo,
			debug,
		),
		consts.ProviderNameOpenAI: NewOpenAICompatibleProvider(
			consts.OpenAIProviderInfo,
			debug,
		),
	}
}
