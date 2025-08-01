package inference

import (
	"context"
	"log/slog"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/huggingface"
)

// HuggingFaceCompatibleAPI struct that implements the CompletionProvider interface.
type HuggingFaceCompatibleAPI struct {
	*BaseAIAPI

	llm *huggingface.LLM
}

// NewHuggingFaceCompatibleAPI creates a new instance of HuggingFaceCompatibleAPI with default ProviderParams.
func NewHuggingFaceCompatibleAPI(
	pi spec.ProviderParams,
	debug bool,
) (*HuggingFaceCompatibleAPI, error) {
	a, err := NewBaseAIAPI(&pi, debug)
	if err != nil {
		return nil, err
	}
	return &HuggingFaceCompatibleAPI{
		BaseAIAPI: a,
	}, nil
}

func (api *HuggingFaceCompatibleAPI) GetLLMsModel(ctx context.Context) llms.Model {
	return api.llm
}

func (api *HuggingFaceCompatibleAPI) DeInitLLM(ctx context.Context) error {
	api.llm = nil
	slog.Info(
		"huggingface compatible LLM provider de initialized",
		"name",
		string(api.ProviderParams.Name),
	)
	return nil
}

func (api *HuggingFaceCompatibleAPI) InitLLM(ctx context.Context) error {
	options := []huggingface.Option{}

	providerURL := "https://api-inference.huggingface.co"
	if api.ProviderParams.Origin != "" {
		providerURL = api.ProviderParams.Origin
		options = append(options, huggingface.WithURL(providerURL))
	}
	// Setting a debug client is not supproted on HF by langchaingo
	// if api.BaseAIAPI.Debug {
	// 	options = append(options, huggingface.WithHTTPClient(httputil.DebugHTTPClient))
	// }.
	if api.ProviderParams.APIKey == "" {
		slog.Debug("no API key given, not initializing Huggingface LLM object")
		return nil
	}
	options = append(options, huggingface.WithToken(api.ProviderParams.APIKey))
	llm, err := huggingface.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info(
		"LLM provider initialize",
		"name",
		string(api.ProviderParams.Name),
		"url",
		providerURL,
	)
	return nil
}
