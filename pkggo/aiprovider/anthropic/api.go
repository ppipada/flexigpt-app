package anthropic

import (
	"context"
	"log/slog"

	"github.com/flexigpt/flexiui/pkggo/aiprovider/baseutils"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/spec"

	langchainAnthropic "github.com/tmc/langchaingo/llms/anthropic"
)

// AnthropicAPI struct that implements the CompletionProvider interface.
type AnthropicAPI struct {
	*baseutils.BaseAIAPI
	llm *langchainAnthropic.LLM
}

// NewAnthropicAPI creates a new instance of AnthropicAPI with default ProviderInfo.
func NewAnthropicAPI() *AnthropicAPI {
	pi := AnthropicProviderInfo
	debug := false
	return &AnthropicAPI{
		BaseAIAPI: baseutils.NewBaseAIAPI(&pi, debug),
	}
}

// SetProviderAttribute sets the attributes for the AnthropicAPI.
func (api *AnthropicAPI) SetProviderAttribute(
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
	options := []langchainAnthropic.Option{}
	if api.ProviderInfo.ApiKey != "" {
		options = append(options, langchainAnthropic.WithToken(api.ProviderInfo.ApiKey))
	}
	if api.ProviderInfo.DefaultOrigin != "" {
		options = append(options, langchainAnthropic.WithBaseURL(api.ProviderInfo.DefaultOrigin))
	}
	if api.ProviderInfo.DefaultModel != "" {
		options = append(
			options,
			langchainAnthropic.WithModel(string(api.ProviderInfo.DefaultModel)),
		)
	}
	newClient := baseutils.NewDebugHTTPClient(api.BaseAIAPI.Debug)
	options = append(options, langchainAnthropic.WithHTTPClient(newClient))

	llm, err := langchainAnthropic.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info("Anthropic LLM provider initialized")
	return nil
}

// FetchCompletion processes the completion request.
func (api *AnthropicAPI) FetchCompletion(
	ctx context.Context,
	input spec.CompletionRequest,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	return api.BaseAIAPI.FetchCompletion(ctx, api.llm, input, onStreamData)
}
