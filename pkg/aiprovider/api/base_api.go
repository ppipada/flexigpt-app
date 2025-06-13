package api

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/aiprovider/consts"
	"github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"
	"github.com/tmc/langchaingo/llms"
)

type BaseAIAPI struct {
	ProviderInfo *spec.ProviderInfo
	Debug        bool
}

// NewOpenAIAPI creates a new instance of BaseAIAPI with input ProviderInfo.
func NewBaseAIAPI(p *spec.ProviderInfo, debug bool) *BaseAIAPI {
	return &BaseAIAPI{
		ProviderInfo: p,
		Debug:        debug,
	}
}

// IsConfigured checks if the API is configured.
func (api *BaseAIAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderInfo.APIKey != ""
}

func (api *BaseAIAPI) GetProviderInfo(ctx context.Context) *spec.ProviderInfo {
	return api.ProviderInfo
}

// SetProviderAPIKey sets the key for a provider.
func (api *BaseAIAPI) SetProviderAPIKey(
	ctx context.Context,
	apiKey string,
) error {
	if apiKey == "" {
		return errors.New("invalid apikey provided")
	}
	if api.ProviderInfo == nil {
		return errors.New("no ProviderInfo found")
	}

	api.ProviderInfo.APIKey = apiKey

	return nil
}

// SetProviderAttribute sets the attributes of a provider.
func (api *BaseAIAPI) SetProviderAttribute(
	ctx context.Context,
	origin *string,
	chatCompletionPathPrefix *string,
) error {
	if origin == nil && chatCompletionPathPrefix == nil {
		return errors.New("no attribute provided for set")
	}
	if api.ProviderInfo == nil {
		return errors.New("no ProviderInfo found")
	}
	if origin != nil && *origin != "" {
		api.ProviderInfo.Origin = *origin
	}
	if chatCompletionPathPrefix != nil {
		api.ProviderInfo.ChatCompletionPathPrefix = *chatCompletionPathPrefix
	}

	return nil
}

func trimInbuiltPrompts(systemPrompt, inbuiltPrompt string) string {
	// Split both prompts into lines.
	inbuiltLines := strings.Split(inbuiltPrompt, "\n")
	promptLines := strings.Split(systemPrompt, "\n")

	// Remove matching lines from the start.
	for len(inbuiltLines) > 0 && len(promptLines) > 0 && promptLines[0] == inbuiltLines[0] {
		promptLines = promptLines[1:]
		inbuiltLines = inbuiltLines[1:]
	}
	// Re-join the remaining lines.
	return inbuiltPrompt + "\n" + strings.TrimLeft(strings.Join(promptLines, "\n"), "\n")
}

func (api *BaseAIAPI) getCompletionRequest(
	prompt string,
	modelParams spec.ModelParams,
	inbuiltModelParams *spec.ModelParams,
	prevMessages []spec.ChatCompletionRequestMessage,
) *CompletionRequest {
	completionRequest := CompletionRequest{
		ModelParams: spec.ModelParams{
			Name:                 modelParams.Name,
			AdditionalParameters: modelParams.AdditionalParameters,
			Temperature:          modelParams.Temperature,
			Reasoning:            modelParams.Reasoning,
		},
	}

	// Cannot turn on streaming of it is set false in model.
	if inbuiltModelParams != nil && modelParams.Stream {
		completionRequest.ModelParams.Stream = inbuiltModelParams.Stream
	} else {
		completionRequest.ModelParams.Stream = modelParams.Stream
	}

	if inbuiltModelParams != nil &&
		modelParams.MaxOutputLength > inbuiltModelParams.MaxOutputLength {
		completionRequest.ModelParams.MaxOutputLength = inbuiltModelParams.MaxOutputLength
	} else {
		completionRequest.ModelParams.MaxOutputLength = modelParams.MaxOutputLength
	}

	if inbuiltModelParams != nil &&
		modelParams.MaxPromptLength > inbuiltModelParams.MaxPromptLength {
		completionRequest.ModelParams.MaxPromptLength = inbuiltModelParams.MaxPromptLength
	} else {
		completionRequest.ModelParams.MaxPromptLength = modelParams.MaxPromptLength
	}

	if inbuiltModelParams != nil && inbuiltModelParams.Reasoning != nil &&
		modelParams.Reasoning != nil &&
		modelParams.Reasoning.Type != inbuiltModelParams.Reasoning.Type {
		// Cannot have input reasoning different than models defined reasoning type.
		completionRequest.ModelParams.Reasoning = nil
	}

	reqSystemPrompt := modelParams.SystemPrompt
	if inbuiltModelParams != nil && inbuiltModelParams.SystemPrompt != "" {
		reqSystemPrompt = trimInbuiltPrompts(reqSystemPrompt, inbuiltModelParams.SystemPrompt)
	}
	completionRequest.ModelParams.SystemPrompt = reqSystemPrompt

	// Handle messages.
	messages := slices.Clone(prevMessages)
	if prompt != "" {
		message := spec.ChatCompletionRequestMessage{
			Role:    "user",
			Content: &prompt,
		}
		messages = append(messages, message)
	}
	completionRequest.Messages = messages

	// Assuming filterMessagesByTokenCount is implemented elsewhere.
	completionRequest.Messages = FilterMessagesByTokenCount(
		completionRequest.Messages,
		completionRequest.ModelParams.MaxPromptLength,
	)

	return &completionRequest
}

// FetchCompletion processes the completion request.
func (api *BaseAIAPI) FetchCompletion(
	ctx context.Context,
	llm llms.Model,
	prompt string,
	modelParams spec.ModelParams,
	inbuiltModelParams *spec.ModelParams,
	prevMessages []spec.ChatCompletionRequestMessage,
	onStreamData func(data string) error,
) (*CompletionResponse, error) {
	input := api.getCompletionRequest(prompt, modelParams, inbuiltModelParams, prevMessages)
	if len(input.Messages) == 0 {
		return nil, errors.New("empty input messages")
	}
	if llm == nil {
		return nil, errors.New("llm not initialized")
	}

	options := []llms.CallOption{
		llms.WithModel(string(input.ModelParams.Name)),
	}
	// PrintJSON(input.ModelParams).

	if input.ModelParams.Temperature != nil {
		options = append(options, llms.WithTemperature(*input.ModelParams.Temperature))
	}
	if rp := input.ModelParams.Reasoning; rp != nil &&
		rp.Type == spec.ReasoningTypeHybridWithTokens {
		options = append(options, llms.WithReasoning(llms.Reasoning{
			IsEnabled: true,
			Mode:      llms.ReasoningModeTokens,
			Tokens:    rp.Tokens,
		}))
	}
	if rp := input.ModelParams.Reasoning; rp != nil &&
		rp.Type == spec.ReasoningTypeSingleWithLevels {
		options = append(options, llms.WithReasoning(llms.Reasoning{
			IsEnabled: true,
			Mode:      llms.ReasoningModeLevel,
			Level:     llms.ReasoningLevel(rp.Level),
		}))
	}
	options = append(options, llms.WithMaxTokens(input.ModelParams.MaxOutputLength))

	// Wrap onStreamData.
	var write func(string) error
	var flush func()
	if input.ModelParams.Stream && onStreamData != nil {
		write, flush = NewBufferedStreamer(onStreamData, FlushInterval, FlushChunkSize)
		if input.ModelParams.Reasoning != nil {
			streamingReasoningFunc := func(ctx context.Context, reasoningChunk []byte, chunk []byte) error {
				rc := string(reasoningChunk)
				if rc != "" {
					rc = getBlockQuotedReasoning(rc)
				}
				return write(rc + string(chunk))
			}
			options = append(options, llms.WithStreamingReasoningFunc(streamingReasoningFunc))
		} else {
			streamingFunc := func(ctx context.Context, chunk []byte) error {
				return write(string(chunk))
			}
			options = append(options, llms.WithStreamingFunc(streamingFunc))
		}
	}

	content := []llms.MessageContent{}
	if sp := input.ModelParams.SystemPrompt; sp != "" {
		sysmsg := llms.TextParts(llms.ChatMessageTypeSystem, sp)
		if api.ProviderInfo.Name == consts.ProviderNameOpenAI &&
			strings.HasPrefix(string(input.ModelParams.Name), "o") {
			sysmsg = llms.TextParts(llms.ChatMessageTypeDeveloper, sp)
		}
		content = append(content, sysmsg)
	}
	for _, msg := range input.Messages {
		if msg.Content != nil {
			content = append(content, llms.TextParts(LangchainRoleMap[msg.Role], *msg.Content))
		}
	}
	if len(content) == 0 {
		return nil, errors.New("empty input content messages")
	}

	completionResp := &CompletionResponse{}

	ctx = AddDebugResponseToCtx(ctx)
	resp, err := llm.GenerateContent(ctx, content, options...)

	// Make sure buffered data reaches the client.
	if flush != nil {
		flush()
	}

	debugResp, ok := GetDebugHTTPResponse(ctx)
	if err != nil {
		if ok && debugResp != nil && debugResp.ErrorDetails != nil {
			completionResp.ErrorDetails = debugResp.ErrorDetails
			return completionResp, nil
		}
		return nil, err
	}
	if ok && debugResp != nil {
		completionResp.RequestDetails = debugResp.RequestDetails
		completionResp.ErrorDetails = debugResp.ErrorDetails
		completionResp.ResponseDetails = debugResp.ResponseDetails
	}

	// PrintJSON(resp).

	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0] == nil {
		if ok && debugResp != nil {
			if completionResp.ErrorDetails == nil {
				completionResp.ErrorDetails = &APIErrorDetails{
					Message: "got nil response from LLM api",
				}
			} else {
				completionResp.ErrorDetails.Message += " got nil response from LLM api"
			}
			return completionResp, nil
		}
		return nil, errors.New("got nil response from LLM api")
	}

	reasoningContent := ""
	if rc := resp.Choices[0].ReasoningContent; rc != "" {
		reasoningContent = "> Thought process:\n\n" + getBlockQuotedReasoning(rc) + "\n\n"
	}
	full := reasoningContent + resp.Choices[0].Content
	completionResp.RespContent = &full

	return completionResp, nil
}

func getBlockQuotedReasoning(content string) string {
	// Split the content into lines.
	lines := strings.Split(content, "\n")
	// Prepend each line with "> ".
	for i, line := range lines {
		lines[i] = "> " + line
	}
	// Join the lines back together as blockquote.
	return strings.Join(lines, "\n")
}
