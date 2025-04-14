package api

import (
	"context"
	"log/slog"
	"strings"

	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"

	"github.com/tmc/langchaingo/llms"
	langchainAnthropic "github.com/tmc/langchaingo/llms/anthropic"
)

// AnthropicCompatibleAPI struct that implements the CompletionProvider interface.
type AnthropicCompatibleAPI struct {
	*BaseAIAPI
	llm *langchainAnthropic.LLM
}

// NewAnthropicCompatibleAPI creates a new instance of AnthropicCompatibleAPI with default ProviderInfo.
func NewAnthropicCompatibleAPI(pi spec.ProviderInfo, debug bool) *AnthropicCompatibleAPI {
	return &AnthropicCompatibleAPI{
		BaseAIAPI: NewBaseAIAPI(&pi, debug),
	}
}

func (api *AnthropicCompatibleAPI) GetLLMsModel(ctx context.Context) llms.Model {
	return api.llm
}

func (api *AnthropicCompatibleAPI) InitLLM(ctx context.Context) error {
	options := []langchainAnthropic.Option{}
	providerURL := "https://api.anthropic.com/v1"
	if api.ProviderInfo.Origin != "" {
		baseURL := api.ProviderInfo.Origin
		// Remove trailing slash from baseURL if present
		baseURL = strings.TrimSuffix(baseURL, "/")

		pathPrefix := api.ProviderInfo.ChatCompletionPathPrefix
		// Remove '/messages' from pathPrefix if present
		// This is because langchaingo adds '/messages' internally
		pathPrefix = strings.TrimSuffix(pathPrefix, "/messages")
		providerURL = baseURL + pathPrefix
		options = append(options, langchainAnthropic.WithBaseURL(providerURL))
	}
	if api.ProviderInfo.APIKey == "" {
		slog.Debug("No API key given. Not initializing Anthropic LLM object")
		return nil
	}
	options = append(options, langchainAnthropic.WithToken(api.ProviderInfo.APIKey))
	newClient := NewDebugHTTPClient(api.Debug)
	options = append(options, langchainAnthropic.WithHTTPClient(newClient))

	llm, err := langchainAnthropic.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info("LLM provider initialize", "Name", string(api.ProviderInfo.Name), "URL", providerURL)
	return nil
}
