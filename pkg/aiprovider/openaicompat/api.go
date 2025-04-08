package openaicompat

import (
	"context"
	"log/slog"
	"strings"

	"github.com/flexigpt/flexiui/pkg/aiprovider/baseutils"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"

	langchainOpenAI "github.com/tmc/langchaingo/llms/openai"
)

// OpenAICompatibleAPI struct that implements the CompletionProvider interface.
type OpenAICompatibleAPI struct {
	*baseutils.BaseAIAPI
	llm *langchainOpenAI.LLM
}

// NewOpenAICompatibleProvider creates a new instance of OpenAICompatibleProvider with the provided ProviderInfo.
func NewOpenAICompatibleProvider(pi spec.ProviderInfo, debug bool) *OpenAICompatibleAPI {
	return &OpenAICompatibleAPI{
		BaseAIAPI: baseutils.NewBaseAIAPI(&pi, debug),
	}
}

// SetProviderAttribute sets the attributes for the OpenAICompatibleAPI.
func (api *OpenAICompatibleAPI) SetProviderAttribute(
	ctx context.Context,
	apiKey *string,
	origin *string,
	chatCompletionPathPrefix *string,
) error {
	err := api.BaseAIAPI.SetProviderAttribute(
		ctx,
		apiKey,
		origin,
		chatCompletionPathPrefix,
	)
	if err != nil {
		return err
	}
	options := []langchainOpenAI.Option{}

	if api.ProviderInfo.DefaultModel != "" {
		options = append(options, langchainOpenAI.WithModel(string(api.ProviderInfo.DefaultModel)))
	}

	if api.ProviderInfo.Origin != "" {
		baseURL := api.ProviderInfo.Origin
		// Remove trailing slash from baseURL if present
		baseURL = strings.TrimSuffix(baseURL, "/")

		pathPrefix := api.ProviderInfo.ChatCompletionPathPrefix
		// Remove '/chat/completions' from pathPrefix if present
		// This is because langchaingo adds '/chat/completions' internally
		pathPrefix = strings.TrimSuffix(pathPrefix, "/chat/completions")
		slog.Info("URL", "url", baseURL+pathPrefix)
		options = append(options, langchainOpenAI.WithBaseURL(baseURL+pathPrefix))
	}

	if api.ProviderInfo.APIKey == "" {
		slog.Debug(
			string(
				api.ProviderInfo.Name,
			) + ": No API key given. Not initializing OpenAICompatibleAPI LLM object",
		)
		return nil
	}
	options = append(options, langchainOpenAI.WithToken(api.ProviderInfo.APIKey))
	newClient := baseutils.NewDebugHTTPClient(api.BaseAIAPI.Debug)
	options = append(options, langchainOpenAI.WithHTTPClient(newClient))

	llm, err := langchainOpenAI.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info(string(api.ProviderInfo.Name) + ": OpenAICompatibleAPI LLM provider initialized")
	return nil
}

// FetchCompletion processes the completion request.
func (api *OpenAICompatibleAPI) FetchCompletion(
	ctx context.Context,
	prompt string,
	modelParams spec.ModelParams,
	prevMessages []spec.ChatCompletionRequestMessage,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	return api.BaseAIAPI.FetchCompletion(
		ctx,
		api.llm,
		prompt,
		modelParams,
		prevMessages,
		onStreamData,
	)
}
