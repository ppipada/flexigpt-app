package anthropic

import (
	"context"
	"log/slog"

	"github.com/flexigpt/flexiui/pkg/aiprovider/baseutils"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"

	langchainAnthropic "github.com/tmc/langchaingo/llms/anthropic"
)

// AnthropicAPI struct that implements the CompletionProvider interface.
type AnthropicAPI struct {
	*baseutils.BaseAIAPI
	llm *langchainAnthropic.LLM
}

// NewAnthropicAPI creates a new instance of AnthropicAPI with default ProviderInfo.
func NewAnthropicAPI(debug bool) *AnthropicAPI {
	pi := AnthropicProviderInfo
	return &AnthropicAPI{
		BaseAIAPI: baseutils.NewBaseAIAPI(&pi, debug),
	}
}

// SetProviderAttribute sets the attributes for the AnthropicAPI.
func (api *AnthropicAPI) SetProviderAttribute(
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
	options := []langchainAnthropic.Option{}

	if api.ProviderInfo.Origin != "" {
		options = append(options, langchainAnthropic.WithBaseURL(api.ProviderInfo.Origin))
	}
	if api.ProviderInfo.DefaultModel != "" {
		options = append(
			options,
			langchainAnthropic.WithModel(string(api.ProviderInfo.DefaultModel)),
		)
	}
	if api.ProviderInfo.APIKey == "" {
		slog.Debug("No API key given. Not initializing Anthropic LLM object")
		return nil
	}
	options = append(options, langchainAnthropic.WithToken(api.ProviderInfo.APIKey))
	newClient := baseutils.NewDebugHTTPClient(api.Debug)
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
