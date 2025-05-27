package api

import (
	"context"
	"log/slog"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"

	"github.com/tmc/langchaingo/llms"
	langchainOpenAI "github.com/tmc/langchaingo/llms/openai"
)

// OpenAICompatibleAPI struct that implements the CompletionProvider interface.
type OpenAICompatibleAPI struct {
	*BaseAIAPI
	llm *langchainOpenAI.LLM
}

// NewOpenAICompatibleProvider creates a new instance of OpenAICompatibleProvider with the provided ProviderInfo.
func NewOpenAICompatibleProvider(pi spec.ProviderInfo, debug bool) *OpenAICompatibleAPI {
	return &OpenAICompatibleAPI{
		BaseAIAPI: NewBaseAIAPI(&pi, debug),
	}
}

func (api *OpenAICompatibleAPI) GetLLMsModel(ctx context.Context) llms.Model {
	return api.llm
}

func (api *OpenAICompatibleAPI) InitLLM(ctx context.Context) error {
	options := []langchainOpenAI.Option{}

	providerURL := "https://api.openai.com/v1"
	if api.ProviderInfo.Origin != "" {
		baseURL := api.ProviderInfo.Origin
		// Remove trailing slash from baseURL if present
		baseURL = strings.TrimSuffix(baseURL, "/")

		pathPrefix := api.ProviderInfo.ChatCompletionPathPrefix
		// Remove '/chat/completions' from pathPrefix if present
		// This is because langchaingo adds '/chat/completions' internally
		pathPrefix = strings.TrimSuffix(pathPrefix, "/chat/completions")
		providerURL = baseURL + pathPrefix
		options = append(options, langchainOpenAI.WithBaseURL(providerURL))
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
	newClient := NewDebugHTTPClient(api.Debug)
	options = append(options, langchainOpenAI.WithHTTPClient(newClient))

	llm, err := langchainOpenAI.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info(
		"OpenAICompatibleAPI LLM provider initialize",
		"Name",
		string(api.ProviderInfo.Name),
		"URL",
		providerURL,
	)
	return nil
}
