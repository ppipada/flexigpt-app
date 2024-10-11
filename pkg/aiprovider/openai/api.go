package openai

import (
	"context"
	"errors"
	"fmt"

	"github.com/flexigpt/flexiui/pkg/aiprovider/baseutils"
	"github.com/flexigpt/flexiui/pkg/aiprovider/spec"

	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/llms"
	langchainOpenAI "github.com/tmc/langchaingo/llms/openai"
)

// OpenAIAPI struct that implements the CompletionProvider interface
type OpenAIAPI struct {
	ProviderInfo *spec.ProviderInfo
	llm          *langchainOpenAI.LLM
	debug        bool
}

// NewOpenAIAPI creates a new instance of OpenAIAPI with default ProviderInfo
func NewOpenAIAPI() *OpenAIAPI {
	pi := spec.OpenAIProviderInfo

	return &OpenAIAPI{
		ProviderInfo: &pi,
		debug:        false,
	}
}

// IsConfigured checks if the API is configured
func (api *OpenAIAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderInfo.IsConfigured()
}

// GetProviderInfo returns the provider information
func (api *OpenAIAPI) GetProviderInfo(ctx context.Context) (*spec.ProviderInfo, error) {
	if api.ProviderInfo == nil {
		return nil, errors.New("provider info is not set")
	}
	return api.ProviderInfo, nil
}

// GetCompletionRequest creates a new completion request
func (api *OpenAIAPI) GetCompletionRequest(
	ctx context.Context,
	prompt string,
	prevMessages []spec.ChatCompletionRequestMessage,
	inputParams map[string]interface{},
	stream bool,
) (*spec.CompletionRequest, error) {
	resp := baseutils.GetCompletionRequest(
		prompt,
		prevMessages,
		inputParams,
		stream,
		api.ProviderInfo,
	)
	return &resp, nil
}

// SetProviderAttribute sets the attributes for the OpenAIAPI
func (api *OpenAIAPI) SetProviderAttribute(
	ctx context.Context,
	apiKey *string,
	defaultModel *string,
	defaultTemperature *float64,
	defaultOrigin *string,
) error {
	if apiKey == nil && defaultModel == nil && defaultTemperature == nil && defaultOrigin == nil {
		fmt.Println("No attribute provided for set, returning.")
		return nil
	}
	if api.ProviderInfo == nil {
		fmt.Println("No ProviderInfo found, returning.")
		return nil
	}
	baseutils.SetProviderAttribute(
		api.ProviderInfo,
		apiKey,
		defaultModel,
		defaultTemperature,
		defaultOrigin,
	)
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
	if api.debug {
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
	if len(input.Messages) == 0 {
		return nil, fmt.Errorf("Empty input messages")
	}
	options := []llms.CallOption{
		llms.WithModel(input.Model),
		llms.WithTemperature(input.Temperature),
	}
	if input.MaxOutputLength != nil {
		options = append(options, llms.WithMaxTokens(*input.MaxOutputLength))
	}
	if input.Stream && onStreamData != nil {
		streamingFunc := func(ctx context.Context, chunk []byte) error {
			err := onStreamData(string(chunk))
			return err
		}
		options = append(options, llms.WithStreamingFunc(streamingFunc))
	}

	content := []llms.MessageContent{}
	if input.SystemPrompt != nil {
		content = append(content, llms.TextParts(llms.ChatMessageTypeSystem, *input.SystemPrompt))
	}
	for _, msg := range input.Messages {
		role := baseutils.LangchainRoleMap[msg.Role]
		if msg.Content != nil {
			content = append(content, llms.TextParts(role, *msg.Content))
		}
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("Empty input content messages")
	}

	resp, err := api.llm.GenerateContent(ctx, content, options...)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return nil, fmt.Errorf("Got nil response from LLM api")
	}

	return &spec.CompletionResponse{
		RespContent: &resp.Choices[0].Content,
	}, nil
}
