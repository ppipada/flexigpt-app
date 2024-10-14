package openai

import (
	"context"
	"fmt"

	"github.com/flexigpt/flexiui/pkg/aiprovider/baseutils"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"

	"github.com/tmc/langchaingo/httputil"
	langchainOpenAI "github.com/tmc/langchaingo/llms/openai"
)

// OpenAIAPI struct that implements the CompletionProvider interface
type OpenAIAPI struct {
	*baseutils.BaseAIAPI
	llm *langchainOpenAI.LLM
}

// NewOpenAIAPI creates a new instance of OpenAIAPI with default ProviderInfo
func NewOpenAIAPI() *OpenAIAPI {
	pi := OpenAIProviderInfo
	debug := false
	return &OpenAIAPI{
		BaseAIAPI: baseutils.NewBaseAIAPI(&pi, debug),
	}
}

// SetProviderAttribute sets the attributes for the OpenAIAPI
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
	if api.ProviderInfo.ApiKey != "" {
		options = append(options, langchainOpenAI.WithToken(api.ProviderInfo.ApiKey))
	}
	if api.ProviderInfo.DefaultOrigin != "" {
		options = append(options, langchainOpenAI.WithBaseURL(api.ProviderInfo.DefaultOrigin))
	}
	if api.ProviderInfo.DefaultModel != "" {
		options = append(options, langchainOpenAI.WithModel(string(api.ProviderInfo.DefaultModel)))
	}
	if api.BaseAIAPI.Debug {
		options = append(options, langchainOpenAI.WithHTTPClient(httputil.DebugHTTPClient))
	}

	llm, err := langchainOpenAI.New(options...)
	if err != nil {
		return err
	}
	api.llm = llm
	fmt.Println("Set OpenAI LLM provider")
	return nil
}

// FetchCompletion processes the completion request
func (api *OpenAIAPI) FetchCompletion(
	ctx context.Context,
	input spec.CompletionRequest,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	return api.BaseAIAPI.FetchCompletion(ctx, api.llm, input, onStreamData)
}
