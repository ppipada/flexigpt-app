package api

import (
	"context"
	"log/slog"

	"github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"

	"github.com/tmc/langchaingo/llms"
	langchainHuggingFace "github.com/tmc/langchaingo/llms/huggingface"
)

// HuggingFaceCompatibleAPI struct that implements the CompletionProvider interface.
type HuggingFaceCompatibleAPI struct {
	*BaseAIAPI
	llm *langchainHuggingFace.LLM
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
	options := []langchainHuggingFace.Option{}

	providerURL := "https://api-inference.huggingface.co"
	if api.ProviderInfo.Origin != "" {
		providerURL = api.ProviderInfo.Origin
		options = append(options, langchainHuggingFace.WithURL(providerURL))
	}
	// Setting a debug client is not supproted on HF by langchaingo
	// if api.BaseAIAPI.Debug {
	// 	options = append(options, langchainHuggingFace.WithHTTPClient(httputil.DebugHTTPClient))
	// }
	if api.ProviderInfo.APIKey == "" {
		slog.Debug("No API key given. Not initializing Huggingface LLM object")
		return nil
	}
	options = append(options, langchainHuggingFace.WithToken(api.ProviderInfo.APIKey))
	llm, err := langchainHuggingFace.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info("LLM provider initialize", "Name", string(api.ProviderInfo.Name), "URL", providerURL)
	return nil
}
