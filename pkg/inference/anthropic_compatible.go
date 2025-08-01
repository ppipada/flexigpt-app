package inference

import (
	"context"
	"log/slog"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

// AnthropicCompatibleAPI struct that implements the CompletionProvider interface.
type AnthropicCompatibleAPI struct {
	*BaseAIAPI

	llm *anthropic.LLM
}

// NewAnthropicCompatibleAPI creates a new instance of AnthropicCompatibleAPI with default ProviderParams.
func NewAnthropicCompatibleAPI(
	pi spec.ProviderParams,
	debug bool,
) (*AnthropicCompatibleAPI, error) {
	a, err := NewBaseAIAPI(&pi, debug)
	if err != nil {
		return nil, err
	}
	return &AnthropicCompatibleAPI{
		BaseAIAPI: a,
	}, nil
}

func (api *AnthropicCompatibleAPI) GetLLMsModel(ctx context.Context) llms.Model {
	return api.llm
}

func (api *AnthropicCompatibleAPI) DeInitLLM(ctx context.Context) error {
	api.llm = nil
	slog.Info(
		"anthropic compatible LLM provider de initialized",
		"name",
		string(api.ProviderParams.Name),
	)
	return nil
}

func (api *AnthropicCompatibleAPI) InitLLM(ctx context.Context) error {
	options := []anthropic.Option{}
	providerURL := "https://api.anthropic.com/v1"
	if api.ProviderParams.Origin != "" {
		baseURL := api.ProviderParams.Origin
		// Remove trailing slash from baseURL if present.
		baseURL = strings.TrimSuffix(baseURL, "/")

		pathPrefix := api.ProviderParams.ChatCompletionPathPrefix
		// Remove '/messages' from pathPrefix if present
		// This is because langchaingo adds '/messages' internally.
		pathPrefix = strings.TrimSuffix(pathPrefix, "/messages")
		providerURL = baseURL + pathPrefix
		options = append(options, anthropic.WithBaseURL(providerURL))
	}
	if api.ProviderParams.APIKey == "" {
		slog.Debug("no API key given, not initializing Anthropic LLM object")
		return nil
	}
	options = append(options, anthropic.WithToken(api.ProviderParams.APIKey))
	newClient := NewDebugHTTPClient(api.Debug, false)
	options = append(options, anthropic.WithHTTPClient(newClient))

	llm, err := anthropic.New(options...)
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
