package inference

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/builtin"
	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	"github.com/tmc/langchaingo/llms"
)

type CompletionProvider interface {
	InitLLM(ctx context.Context) error
	DeInitLLM(ctx context.Context) error
	GetProviderInfo(
		ctx context.Context,
	) *spec.ProviderParams
	IsConfigured(ctx context.Context) bool
	SetProviderAPIKey(
		ctx context.Context,
		apiKey string,
	) error
	GetLLMsModel(ctx context.Context) llms.Model
	FetchCompletion(
		ctx context.Context,
		llm llms.Model,
		prompt string,
		modelParams spec.ModelParams,
		prevMessages []spec.ChatCompletionRequestMessage,
		onStreamData func(data string) error,
	) (*spec.CompletionResponse, error)
}

var LangchainRoleMap = map[spec.ChatCompletionRoleEnum]llms.ChatMessageType{
	// No developer prompt support in langchain as of now.
	spec.Developer: llms.ChatMessageTypeSystem,
	spec.System:    llms.ChatMessageTypeSystem,
	spec.User:      llms.ChatMessageTypeHuman,
	spec.Assistant: llms.ChatMessageTypeAI,
	spec.Function:  llms.ChatMessageTypeTool,
}

type BaseAIAPI struct {
	ProviderParams *spec.ProviderParams
	Debug          bool
}

// NewOpenAIAPI creates a new instance of BaseAIAPI with input ProviderParams.
func NewBaseAIAPI(p *spec.ProviderParams, debug bool) (*BaseAIAPI, error) {
	if p == nil || p.Name == "" || p.Origin == "" {
		return nil, errors.New("invalid args")
	}
	return &BaseAIAPI{
		ProviderParams: p,
		Debug:          debug,
	}, nil
}

// IsConfigured checks if the API is configured.
func (api *BaseAIAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderParams.APIKey != ""
}

func (api *BaseAIAPI) GetProviderInfo(ctx context.Context) *spec.ProviderParams {
	return api.ProviderParams
}

// SetProviderAPIKey sets the key for a provider.
func (api *BaseAIAPI) SetProviderAPIKey(
	ctx context.Context,
	apiKey string,
) error {
	if apiKey == "" {
		return errors.New("invalid apikey provided")
	}
	if api.ProviderParams == nil {
		return errors.New("no ProviderParams found")
	}

	api.ProviderParams.APIKey = apiKey

	return nil
}

// FetchCompletion processes the completion request.
func (api *BaseAIAPI) FetchCompletion(
	ctx context.Context,
	llm llms.Model,
	prompt string,
	modelParams spec.ModelParams,
	prevMessages []spec.ChatCompletionRequestMessage,
	onStreamData func(data string) error,
) (*spec.CompletionResponse, error) {
	input := getCompletionRequest(prompt, modelParams, prevMessages)
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
		rp.Type == modelpresetSpec.ReasoningTypeHybridWithTokens {
		options = append(options, llms.WithReasoning(llms.Reasoning{
			IsEnabled: true,
			Mode:      llms.ReasoningModeTokens,
			Tokens:    rp.Tokens,
		}))
	}
	if rp := input.ModelParams.Reasoning; rp != nil &&
		rp.Type == modelpresetSpec.ReasoningTypeSingleWithLevels {
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
		if api.ProviderParams.Name == builtin.ProviderNameOpenAI &&
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

	completionResp := &spec.CompletionResponse{}

	ctx = AddDebugResponseToCtx(ctx)
	resp, err := llm.GenerateContent(ctx, content, options...)

	// Make sure buffered data reaches the client.
	if flush != nil {
		flush()
	}

	isNilResp := resp == nil || len(resp.Choices) == 0 || resp.Choices[0] == nil
	attachDebugResp(ctx, completionResp, err, isNilResp)
	if isNilResp {
		return completionResp, nil
	}

	reasoningContent := ""
	if rc := resp.Choices[0].ReasoningContent; rc != "" {
		reasoningContent = "> Thought process:\n\n" + getBlockQuotedReasoning(rc) + "\n\n"
	}
	full := reasoningContent + resp.Choices[0].Content
	completionResp.RespContent = &full

	return completionResp, nil
}

// attachDebugResp adds HTTP-debug information and error context—without panics.
//
// – ctx may or may not contain debug information.
// – respErr is the transport/SDK error (may be nil).
// – isNilResp tells whether the model returned an empty/invalid response.
func attachDebugResp(
	ctx context.Context,
	completionResp *spec.CompletionResponse,
	respErr error,
	isNilResp bool,
) {
	if completionResp == nil {
		return
	}

	debugResp, _ := GetDebugHTTPResponse(ctx)

	// Always attach request/response debug info if available.
	if debugResp != nil {
		completionResp.RequestDetails = debugResp.RequestDetails
		completionResp.ResponseDetails = debugResp.ResponseDetails
	}

	// Gather error-message fragments.
	var msgParts []string
	if debugResp != nil && debugResp.ErrorDetails != nil {
		if m := strings.TrimSpace(debugResp.ErrorDetails.Message); m != "" {
			msgParts = append(msgParts, m)
		}
	}
	if respErr != nil {
		msgParts = append(msgParts, respErr.Error())
	}
	if isNilResp {
		msgParts = append(msgParts, "got nil response from LLM api")
	}

	if len(msgParts) == 0 {
		// Nothing to write; leave ErrorDetails as-is (nil or previously set).
		return
	}

	// Prepare ErrorDetails without aliasing the debug struct pointer.
	if debugResp != nil && debugResp.ErrorDetails != nil {
		ed := *debugResp.ErrorDetails
		ed.Message = strings.Join(msgParts, "; ")
		completionResp.ErrorDetails = &ed
	} else {
		completionResp.ErrorDetails = &spec.APIErrorDetails{
			Message: strings.Join(msgParts, "; "),
		}
	}
}

func getCompletionRequest(
	prompt string,
	modelParams spec.ModelParams,
	prevMessages []spec.ChatCompletionRequestMessage,
) *spec.CompletionRequest {
	completionRequest := spec.CompletionRequest{
		ModelParams: spec.ModelParams{
			Name:                        modelParams.Name,
			Stream:                      modelParams.Stream,
			MaxPromptLength:             modelParams.MaxPromptLength,
			MaxOutputLength:             modelParams.MaxOutputLength,
			Temperature:                 modelParams.Temperature,
			Reasoning:                   modelParams.Reasoning,
			SystemPrompt:                modelParams.SystemPrompt,
			Timeout:                     modelParams.Timeout,
			AdditionalParametersRawJSON: modelParams.AdditionalParametersRawJSON,
		},
	}

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

func TrimInbuiltPrompts(systemPrompt, inbuiltPrompt string) string {
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
