package huggingface

import (
	"context"
	"log/slog"

	"github.com/flexigpt/flexiui/pkggo/aiprovider/baseutils"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/spec"

	langchainHuggingFace "github.com/tmc/langchaingo/llms/huggingface"
)

// HuggingFaceAPI struct that implements the CompletionProvider interface.
type HuggingFaceAPI struct {
	*baseutils.BaseAIAPI
	llm *langchainHuggingFace.LLM
}

// NewHuggingFaceAPI creates a new instance of HuggingFaceAPI with default ProviderInfo.
func NewHuggingFaceAPI() *HuggingFaceAPI {
	pi := HuggingfaceProviderInfo
	debug := false
	return &HuggingFaceAPI{
		BaseAIAPI: baseutils.NewBaseAIAPI(&pi, debug),
	}
}

// SetProviderAttribute sets the attributes for the HuggingFaceAPI.
func (api *HuggingFaceAPI) SetProviderAttribute(
	ctx context.Context,
	apiKey *string,
	defaultModel *string,
	defaultTemperature *float64,
	defaultOrigin *string,
) error {
	err := api.BaseAIAPI.SetProviderAttribute(
		ctx,
		apiKey,
		defaultModel,
		defaultTemperature,
		defaultOrigin,
	)
	if err != nil {
		return err
	}
	options := []langchainHuggingFace.Option{}
	if api.ProviderInfo.ApiKey != "" {
		options = append(options, langchainHuggingFace.WithToken(api.ProviderInfo.ApiKey))
	}
	if api.ProviderInfo.DefaultOrigin != "" {
		options = append(options, langchainHuggingFace.WithURL(api.ProviderInfo.DefaultOrigin))
	}
	if api.ProviderInfo.DefaultModel != "" {
		options = append(
			options,
			langchainHuggingFace.WithModel(string(api.ProviderInfo.DefaultModel)),
		)
	}
	// Setting a debug client is not supproted on HF by langchaingo
	// if api.BaseAIAPI.Debug {
	// 	options = append(options, langchainHuggingFace.WithHTTPClient(httputil.DebugHTTPClient))
	// }

	llm, err := langchainHuggingFace.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info("HuggingFace LLM provider initialized")
	return nil
}

// FetchCompletion processes the completion request.
func (api *HuggingFaceAPI) FetchCompletion(
	ctx context.Context,
	input spec.CompletionRequest,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	return api.BaseAIAPI.FetchCompletion(ctx, api.llm, input, onStreamData)
}
