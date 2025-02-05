package openai

import (
	"context"
	"log/slog"
	"strings"

	"github.com/flexigpt/flexiui/pkg/aiprovider/baseutils"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"

	langchainOpenAI "github.com/tmc/langchaingo/llms/openai"
)

// OpenAIAPI struct that implements the CompletionProvider interface.
type OpenAIAPI struct {
	*baseutils.BaseAIAPI
	llm *langchainOpenAI.LLM
}

// NewOpenAICompatibleProvider creates a new instance of OpenAICompatibleProvider with the provided ProviderInfo.
func NewOpenAICompatibleProvider(pi spec.ProviderInfo, debug bool) *OpenAIAPI {
	return &OpenAIAPI{
		BaseAIAPI: baseutils.NewBaseAIAPI(&pi, debug),
	}
}

// SetProviderAttribute sets the attributes for the OpenAIAPI.
func (api *OpenAIAPI) SetProviderAttribute(
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
	options := []langchainOpenAI.Option{}
	if api.ProviderInfo.APIKey != "" {
		options = append(options, langchainOpenAI.WithToken(api.ProviderInfo.APIKey))
	}
	if api.ProviderInfo.DefaultModel != "" {
		options = append(options, langchainOpenAI.WithModel(string(api.ProviderInfo.DefaultModel)))
	}

	if api.ProviderInfo.DefaultOrigin != "" {
		baseURL := api.ProviderInfo.DefaultOrigin
		// Remove trailing slash from baseURL if present
		baseURL = strings.TrimSuffix(baseURL, "/")

		pathPrefix := api.ProviderInfo.ChatCompletionPathPrefix
		// Remove '/chat/completions' from pathPrefix if present
		// This is because langchaingo adds '/chat/completions' internally
		pathPrefix = strings.TrimSuffix(pathPrefix, "/chat/completions")

		options = append(options, langchainOpenAI.WithBaseURL(baseURL+pathPrefix))
	}

	newClient := baseutils.NewDebugHTTPClient(api.BaseAIAPI.Debug)
	options = append(options, langchainOpenAI.WithHTTPClient(newClient))

	llm, err := langchainOpenAI.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info("OpenAI LLM provider initialized")
	return nil
}

// FetchCompletion processes the completion request.
func (api *OpenAIAPI) FetchCompletion(
	ctx context.Context,
	input spec.CompletionRequest,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	return api.BaseAIAPI.FetchCompletion(ctx, api.llm, input, onStreamData)
}
