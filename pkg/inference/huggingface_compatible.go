package inference

import (
	"context"
	"log/slog"

	"github.com/ppipada/flexigpt-app/pkg/model/spec"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/huggingface"
)

// HuggingFaceCompatibleAPI struct that implements the CompletionProvider interface.
type HuggingFaceCompatibleAPI struct {
	*BaseAIAPI

	llm *huggingface.LLM
}

// NewHuggingFaceCompatibleAPI creates a new instance of HuggingFaceCompatibleAPI with default ProviderInfo.
func NewHuggingFaceCompatibleAPI(pi spec.ProviderInfo, debug bool) *HuggingFaceCompatibleAPI {
	return &HuggingFaceCompatibleAPI{
		BaseAIAPI: NewBaseAIAPI(&pi, debug),
	}
}

func (api *HuggingFaceCompatibleAPI) GetLLMsModel(ctx context.Context) llms.Model {
	return api.llm
}

func (api *HuggingFaceCompatibleAPI) InitLLM(ctx context.Context) error {
	options := []huggingface.Option{}

	providerURL := "https://api-inference.huggingface.co"
	if api.ProviderInfo.Origin != "" {
		providerURL = api.ProviderInfo.Origin
		options = append(options, huggingface.WithURL(providerURL))
	}
	// Setting a debug client is not supproted on HF by langchaingo
	// if api.BaseAIAPI.Debug {
	// 	options = append(options, huggingface.WithHTTPClient(httputil.DebugHTTPClient))
	// }.
	if api.ProviderInfo.APIKey == "" {
		slog.Debug("no API key given, not initializing Huggingface LLM object")
		return nil
	}
	options = append(options, huggingface.WithToken(api.ProviderInfo.APIKey))
	llm, err := huggingface.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info("LLM provider initialize", "name", string(api.ProviderInfo.Name), "url", providerURL)
	return nil
}
