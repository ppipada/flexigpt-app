package inference

import (
	"context"
	"log/slog"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// OpenAICompatibleAPI struct that implements the CompletionProvider interface.
type OpenAICompatibleAPI struct {
	*BaseAIAPI

	llm *openai.LLM
}

// NewOpenAICompatibleProvider creates a new instance of OpenAICompatibleProvider with the provided ProviderParams.
func NewOpenAICompatibleProvider(pi spec.ProviderParams, debug bool) (*OpenAICompatibleAPI, error) {
	a, err := NewBaseAIAPI(&pi, debug)
	if err != nil {
		return nil, err
	}
	return &OpenAICompatibleAPI{
		BaseAIAPI: a,
	}, nil
}

func (api *OpenAICompatibleAPI) GetLLMsModel(ctx context.Context) llms.Model {
	return api.llm
}

func (api *OpenAICompatibleAPI) DeInitLLM(ctx context.Context) error {
	api.llm = nil
	slog.Info(
		"openai compatible LLM provider de initialized",
		"name",
		string(api.ProviderParams.Name),
	)
	return nil
}

func (api *OpenAICompatibleAPI) InitLLM(ctx context.Context) error {
	options := []openai.Option{}

	providerURL := "https://api.openai.com/v1"
	if api.ProviderParams.Origin != "" {
		baseURL := api.ProviderParams.Origin
		// Remove trailing slash from baseURL if present.
		baseURL = strings.TrimSuffix(baseURL, "/")

		pathPrefix := api.ProviderParams.ChatCompletionPathPrefix
		// Remove '/chat/completions' from pathPrefix if present,
		// This is because langchaingo adds '/chat/completions' internally.
		pathPrefix = strings.TrimSuffix(pathPrefix, "/chat/completions")
		providerURL = baseURL + pathPrefix
		options = append(options, openai.WithBaseURL(providerURL))
	}

	if api.ProviderParams.APIKey == "" {
		slog.Debug(
			string(
				api.ProviderParams.Name,
			) + ": No API key given. Not initializing OpenAICompatibleAPI LLM object",
		)
		return nil
	}
	options = append(options, openai.WithToken(api.ProviderParams.APIKey))
	newClient := NewDebugHTTPClient(api.Debug, false)
	options = append(options, openai.WithHTTPClient(newClient))

	llm, err := openai.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info(
		"openai compatible LLM provider initialized",
		"name",
		string(api.ProviderParams.Name),
		"URL",
		providerURL,
	)
	return nil
}
