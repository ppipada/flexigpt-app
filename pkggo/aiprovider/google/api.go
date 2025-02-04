package google

import (
	"context"
	"log/slog"

	"github.com/flexigpt/flexiui/pkggo/aiprovider/baseutils"
	"github.com/flexigpt/flexiui/pkggo/aiprovider/spec"

	langchainOpenAI "github.com/tmc/langchaingo/llms/openai"
)

// GoogleAPI struct that implements the CompletionProvider interface.
type GoogleAPI struct {
	*baseutils.BaseAIAPI
	llm *langchainOpenAI.LLM
}

// NewGoogleAPI creates a new instance of GoogleAPI with default ProviderInfo.
func NewGoogleAPI() *GoogleAPI {
	pi := GoogleProviderInfo
	debug := false
	return &GoogleAPI{
		BaseAIAPI: baseutils.NewBaseAIAPI(&pi, debug),
	}
}

// SetProviderAttribute sets the attributes for the GoogleAPI.
func (api *GoogleAPI) SetProviderAttribute(
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
	if api.ProviderInfo.DefaultOrigin != "" {
		options = append(
			options,
			langchainOpenAI.WithBaseURL(
				api.ProviderInfo.DefaultOrigin+api.ProviderInfo.ChatCompletionPathPrefix,
			),
		)
	}
	if api.ProviderInfo.DefaultModel != "" {
		options = append(options, langchainOpenAI.WithModel(string(api.ProviderInfo.DefaultModel)))
	}
	newClient := baseutils.NewDebugHTTPClient(api.BaseAIAPI.Debug)
	options = append(options, langchainOpenAI.WithHTTPClient(newClient))

	llm, err := langchainOpenAI.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	slog.Info("Google LLM provider initialized")
	return nil
}

// FetchCompletion processes the completion request.
func (api *GoogleAPI) FetchCompletion(
	ctx context.Context,
	input spec.CompletionRequest,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	return api.BaseAIAPI.FetchCompletion(ctx, api.llm, input, onStreamData)
}
