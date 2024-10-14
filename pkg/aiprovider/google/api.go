package google

import (
	"context"
	"fmt"

	"github.com/flexigpt/flexiui/pkg/aiprovider/baseutils"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"

	"github.com/tmc/langchaingo/httputil"
	langchainGoogle "github.com/tmc/langchaingo/llms/googleai"
)

// GoogleAPI struct that implements the CompletionProvider interface
type GoogleAPI struct {
	*baseutils.BaseAIAPI
	llm *langchainGoogle.GoogleAI
}

// NewGoogleAPI creates a new instance of GoogleAPI with default ProviderInfo
func NewGoogleAPI() *GoogleAPI {
	pi := GoogleProviderInfo
	debug := false
	return &GoogleAPI{
		BaseAIAPI: baseutils.NewBaseAIAPI(&pi, debug),
	}
}

// SetProviderAttribute sets the attributes for the GoogleAPI
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
	options := []langchainGoogle.Option{}
	if api.ProviderInfo.ApiKey != "" {
		options = append(options, langchainGoogle.WithAPIKey(api.ProviderInfo.ApiKey))
	}
	// googles official sdk provides option to override the endpoitn but langchain doesnt expose it.
	// if api.ProviderInfo.DefaultOrigin != "" {
	// 	options = append(options, langchainGoogle.WithBaseURL(api.ProviderInfo.DefaultOrigin))
	// }
	if api.ProviderInfo.DefaultModel != "" {
		options = append(
			options,
			langchainGoogle.WithDefaultModel(string(api.ProviderInfo.DefaultModel)),
		)
	}

	if api.BaseAIAPI.Debug {
		options = append(options, langchainGoogle.WithHTTPClient(httputil.DebugHTTPClient))
	}

	options = append(
		options,
		langchainGoogle.WithDefaultTemperature(api.ProviderInfo.DefaultTemperature),
	)
	llm, err := langchainGoogle.New(ctx, options...)
	if err != nil {
		return err
	}
	api.llm = llm
	fmt.Println("Set Google LLM provider")
	return nil
}

// FetchCompletion processes the completion request
func (api *GoogleAPI) FetchCompletion(
	ctx context.Context,
	input spec.CompletionRequest,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	return api.BaseAIAPI.FetchCompletion(ctx, api.llm, input, onStreamData)
}
