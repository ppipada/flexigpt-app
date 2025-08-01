package inference

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

// Define the ProviderSetAPI struct.
type ProviderSetAPI struct {
	providers map[modelpresetSpec.ProviderName]CompletionProvider
	debug     bool
}

// NewProviderSetAPI creates a new ProviderSet with the specified default provider.
func NewProviderSetAPI(
	debug bool,
) (*ProviderSetAPI, error) {
	return &ProviderSetAPI{
		providers: map[modelpresetSpec.ProviderName]CompletionProvider{},
		debug:     debug,
	}, nil
}

func (ps *ProviderSetAPI) AddProvider(
	ctx context.Context,
	req *spec.AddProviderRequest,
) (*spec.AddProviderResponse, error) {
	if req == nil || req.Provider == "" || req.Body == nil || req.Body.Origin == "" {
		return nil, errors.New("invalid params")
	}
	_, exists := ps.providers[req.Provider]
	if exists {
		return nil, errors.New(
			"invalid provider: cannot add a provider with same name as an existing provider, delete first",
		)
	}
	if ok := isProviderAPITypeSupported(req.Body.APIType); !ok {
		return nil, errors.New("unsupported provider api type")
	}

	providerInfo := spec.ProviderParams{
		Name:                     req.Provider,
		APIType:                  req.Body.APIType,
		APIKey:                   "",
		Origin:                   req.Body.Origin,
		ChatCompletionPathPrefix: req.Body.ChatCompletionPathPrefix,
		APIKeyHeaderKey:          req.Body.APIKeyHeaderKey,
		DefaultHeaders:           req.Body.DefaultHeaders,
	}

	cp, err := getProviderAPI(providerInfo, ps.debug)
	if err != nil {
		return nil, err
	}
	ps.providers[req.Provider] = cp

	slog.Info("add provider", "name", req.Provider)
	return &spec.AddProviderResponse{}, nil
}

func (ps *ProviderSetAPI) DeleteProvider(
	ctx context.Context,
	req *spec.DeleteProviderRequest,
) (*spec.DeleteProviderResponse, error) {
	if req == nil || req.Provider == "" {
		return nil, errors.New("got empty provider input")
	}
	_, exists := ps.providers[req.Provider]
	if !exists {
		return nil, errors.New(
			"invalid provider: provider does not exist",
		)
	}
	delete(ps.providers, req.Provider)
	slog.Info("deleteProvider", "name", req.Provider)
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
	if req.Body.APIKey == "" {
		err := p.DeInitLLM(ctx)
		return &spec.SetProviderAPIKeyResponse{}, err
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
		p.GetLLMsModel(ctx),
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

func isProviderAPITypeSupported(t modelpresetSpec.ProviderAPIType) bool {
	if t == modelpresetSpec.ProviderAPITypeOpenAICompatible ||
		t == modelpresetSpec.ProviderAPITypeAnthropicCompatible ||
		t == modelpresetSpec.ProviderAPITypeHuggingFaceCompatible {
		return true
	}
	return false
}

func getProviderAPI(p spec.ProviderParams, debug bool) (CompletionProvider, error) {
	switch p.APIType {
	case modelpresetSpec.ProviderAPITypeAnthropicCompatible:
		return NewAnthropicCompatibleAPI(p, debug)

	case modelpresetSpec.ProviderAPITypeHuggingFaceCompatible:
		return NewHuggingFaceCompatibleAPI(p, debug)

	case modelpresetSpec.ProviderAPITypeOpenAICompatible:
		return NewOpenAICompatibleProvider(p, debug)
	}

	return nil, errors.New("invalid provider api type")
}
