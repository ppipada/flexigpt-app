package inference

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicCompatibleAPI implements CompletionProvider for Anthropics' Messages API.
type AnthropicCompatibleAPI struct {
	ProviderParams *spec.ProviderParams
	Debug          bool
	client         *anthropic.Client
}

// NewAnthropicCompatibleAPI creates a new instance of Anthropics provider.
func NewAnthropicCompatibleAPI(
	pi spec.ProviderParams,
	debug bool,
) (*AnthropicCompatibleAPI, error) {
	if pi.Name == "" || pi.Origin == "" {
		return nil, errors.New("anthropic compatible LLM: invalid args")
	}
	return &AnthropicCompatibleAPI{
		ProviderParams: &pi,
		Debug:          debug,
	}, nil
}

func (api *AnthropicCompatibleAPI) InitLLM(ctx context.Context) error {
	if !api.IsConfigured(ctx) {
		slog.Debug(
			string(
				api.ProviderParams.Name,
			) + ": No API key given. Not initializing Anthropics client",
		)
		return nil
	}

	opts := []option.RequestOption{
		// Sets x-api-key.
		option.WithAPIKey(api.ProviderParams.APIKey),
	}

	providerURL := modelpresetSpec.DefaultAnthropicOrigin
	if api.ProviderParams.Origin != "" {
		baseURL := strings.TrimSuffix(api.ProviderParams.Origin, "/")
		// Remove 'v1/messages' from pathPrefix if present,
		// This is because openai sdk adds 'v1/messages' internally.
		pathPrefix := strings.TrimSuffix(
			api.ProviderParams.ChatCompletionPathPrefix,
			"v1/messages",
		)
		providerURL = baseURL + pathPrefix
		opts = append(opts, option.WithBaseURL(strings.TrimSuffix(providerURL, "/")))
	}

	// Add default headers.
	for k, v := range api.ProviderParams.DefaultHeaders {
		opts = append(opts, option.WithHeader(strings.TrimSpace(k), strings.TrimSpace(v)))
	}

	// If the caller provided a non-standard API key header, attach it.
	if api.ProviderParams.APIKeyHeaderKey != "" &&
		!strings.EqualFold(
			api.ProviderParams.APIKeyHeaderKey,
			modelpresetSpec.DefaultAnthropicAuthorizationHeaderKey,
		) &&
		!strings.EqualFold(
			api.ProviderParams.APIKeyHeaderKey,
			modelpresetSpec.DefaultAuthorizationHeaderKey,
		) {
		opts = append(
			opts,
			option.WithHeader(api.ProviderParams.APIKeyHeaderKey, api.ProviderParams.APIKey),
		)
	}

	newClient := NewDebugHTTPClient(api.Debug, false)
	opts = append(opts, option.WithHTTPClient(newClient))

	c := anthropic.NewClient(opts...)
	api.client = &c
	slog.Info(
		"anthropic compatible LLM provider initialized",
		"name", string(api.ProviderParams.Name),
		"URL", providerURL,
	)
	return nil
}

func (api *AnthropicCompatibleAPI) DeInitLLM(ctx context.Context) error {
	api.client = nil
	slog.Info(
		"anthropic compatible LLM: provider de initialized",
		"name",
		string(api.ProviderParams.Name),
	)
	return nil
}

func (api *AnthropicCompatibleAPI) GetProviderInfo(ctx context.Context) *spec.ProviderParams {
	return api.ProviderParams
}

func (api *AnthropicCompatibleAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderParams.APIKey != ""
}

func (api *AnthropicCompatibleAPI) SetProviderAPIKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return errors.New("anthropic compatible LLM: invalid apikey provided")
	}
	if api.ProviderParams == nil {
		return errors.New("anthropic compatible LLM: no ProviderParams found")
	}
	api.ProviderParams.APIKey = apiKey
	return nil
}

func (api *AnthropicCompatibleAPI) FetchCompletion(
	ctx context.Context,
	prompt string,
	modelParams spec.ModelParams,
	prevMessages []spec.ChatCompletionRequestMessage,
	onStreamTextData, onStreamThinkingData func(string) error,
) (*spec.CompletionResponse, error) {
	if api.client == nil {
		return nil, errors.New("anthropic compatible LLM: client not initialized")
	}

	input := getCompletionRequest(prompt, modelParams, prevMessages)
	if len(input.Messages) == 0 && strings.TrimSpace(input.ModelParams.SystemPrompt) == "" {
		return nil, errors.New("anthropic compatible LLM: empty input messages")
	}

	msgs, sysParams, err := toAnthropicMessages(
		input.ModelParams.SystemPrompt,
		input.Messages,
	)
	if err != nil {
		return nil, err
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(input.ModelParams.Name),
		MaxTokens: int64(input.ModelParams.MaxOutputLength),
		Messages:  msgs,
	}
	if len(sysParams) > 0 {
		params.System = sysParams
	}

	if rp := input.ModelParams.Reasoning; rp != nil &&
		rp.Type == modelpresetSpec.ReasoningTypeHybridWithTokens &&
		rp.Tokens > 0 {
		tokens := max(rp.Tokens, 1024)
		params.Thinking = anthropic.ThinkingConfigParamOfEnabled(int64(tokens))
	} else if t := input.ModelParams.Temperature; t != nil {
		params.Temperature = anthropic.Float(*t)
	}

	timeout := modelpresetSpec.DefaultAPITimeout
	if input.ModelParams.Timeout > 0 {
		timeout = time.Duration(input.ModelParams.Timeout) * time.Second
	}

	if input.ModelParams.Stream && onStreamTextData != nil && onStreamThinkingData != nil {
		return api.doStreaming(ctx, params, onStreamTextData, onStreamThinkingData, timeout)
	}
	return api.doNonStreaming(ctx, params, timeout)
}

func (api *AnthropicCompatibleAPI) doNonStreaming(
	ctx context.Context,
	params anthropic.MessageNewParams,
	timeout time.Duration,
) (*spec.CompletionResponse, error) {
	completionResp := &spec.CompletionResponse{}
	ctx = AddDebugResponseToCtx(ctx)

	resp, err := api.client.Messages.New(ctx, params, option.WithRequestTimeout(timeout))
	isNilResp := resp == nil || len(resp.Content) == 0
	attachDebugResp(ctx, completionResp, err, isNilResp)
	if isNilResp {
		return completionResp, nil
	}

	respContent := getResponseContentFromAnthropicMessage(resp)
	completionResp.ResponseContent = respContent
	return completionResp, nil
}

func (api *AnthropicCompatibleAPI) doStreaming(
	ctx context.Context,
	params anthropic.MessageNewParams,
	onStreamTextData, onStreamThinkingData func(string) error,
	timeout time.Duration,
) (*spec.CompletionResponse, error) {
	writeTextData, flushTextData := NewBufferedStreamer(
		onStreamTextData,
		FlushInterval,
		FlushChunkSize,
	)
	writeThinkingData, flushThinkingData := NewBufferedStreamer(
		onStreamThinkingData,
		FlushInterval,
		FlushChunkSize,
	)

	completionResp := &spec.CompletionResponse{}
	ctx = AddDebugResponseToCtx(ctx)

	stream := api.client.Messages.NewStreaming(
		ctx,
		params,
		option.WithRequestTimeout(timeout),
	)

	respFull := anthropic.Message{}
	var (
		streamWriteErr      error
		streamAccumulateErr error
	)

	for stream.Next() {
		event := stream.Current()
		err := respFull.Accumulate(event)
		if err != nil {
			streamAccumulateErr = err
			break
		}

		switch eventVariant := event.AsAny().(type) {
		case anthropic.MessageStartEvent:
			// Contains a Message object with empty content i.e message metadata.
		case anthropic.MessageDeltaEvent:
			// Indicating top-level changes to the final Message object metadata.
		case anthropic.MessageStopEvent:
			// Done.
			break
		case anthropic.ContentBlockStopEvent:
			// Content block done.
		case anthropic.ContentBlockStartEvent:
			streamWriteErr = handleContentBlockStartEvent(eventVariant, writeTextData, writeThinkingData)
			if streamWriteErr != nil {
				break
			}
		case anthropic.ContentBlockDeltaEvent:
			streamWriteErr = handleContentBlockDeltaEvent(eventVariant, writeTextData, writeThinkingData)
			if streamWriteErr != nil {
				break
			}
		default:
			// No valid variant.
		}
	}

	if flushTextData != nil {
		flushTextData()
	}

	if flushThinkingData != nil {
		flushThinkingData()
	}

	streamErr := errors.Join(stream.Err(), streamAccumulateErr, streamWriteErr)
	isNilResp := len(respFull.Content) == 0
	attachDebugResp(ctx, completionResp, streamErr, isNilResp)
	if isNilResp {
		return completionResp, nil
	}

	respContent := getResponseContentFromAnthropicMessage(&respFull)
	completionResp.ResponseContent = respContent
	return completionResp, streamErr
}

func handleContentBlockStartEvent(
	eventVariant anthropic.ContentBlockStartEvent,
	writeTextData, writeThinkingData func(string) error,
) error {
	switch cbStartvariant := eventVariant.ContentBlock.AsAny().(type) {
	case anthropic.TextBlock:
		return writeTextData(cbStartvariant.Text)

	case anthropic.ThinkingBlock:
		return writeThinkingData(cbStartvariant.Thinking)

	case anthropic.RedactedThinkingBlock:
	case anthropic.ToolUseBlock:
	case anthropic.ServerToolUseBlock:
	case anthropic.WebSearchToolResultBlock:
	default:
		// No valid variant.
	}
	return nil
}

func handleContentBlockDeltaEvent(
	eventVariant anthropic.ContentBlockDeltaEvent,
	writeTextData, writeThinkingData func(string) error,
) error {
	switch deltaVariant := eventVariant.Delta.AsAny().(type) {
	case anthropic.TextDelta:
		return writeTextData(deltaVariant.Text)

	case anthropic.ThinkingDelta:
		return writeThinkingData(deltaVariant.Thinking)

	case anthropic.InputJSONDelta:
	case anthropic.CitationsDelta:
	case anthropic.SignatureDelta:
	default:
		// No valid variant.
	}
	return nil
}

func toAnthropicMessages(
	systemPrompt string,
	messages []spec.ChatCompletionRequestMessage,
) (msgs []anthropic.MessageParam, sysPrompts []anthropic.TextBlockParam, err error) {
	var out []anthropic.MessageParam
	var sysParts []string

	if s := strings.TrimSpace(systemPrompt); s != "" {
		sysParts = append(sysParts, s)
	}

	for _, m := range messages {
		if m.Content == nil {
			continue
		}
		switch m.Role {
		case spec.System, spec.Developer:
			sysParts = append(sysParts, strings.TrimSpace(*m.Content))
		case spec.User:
			out = append(out, anthropic.NewUserMessage(anthropic.NewTextBlock(*m.Content)))
		case spec.Assistant:
			out = append(out, anthropic.NewAssistantMessage(anthropic.NewTextBlock(*m.Content)))
		case spec.Function, spec.Tool:
			// Anthropic tools need to be processed independently of messages, skip for now.
		default:
			// Ignore unknown roles.
		}
	}
	sysParams := []anthropic.TextBlockParam{}
	if len(sysParts) > 0 {
		// Anthropic seems to support array of prompts too, but lets keeps sys prompt as a single string for now.
		sysStr := strings.Join(sysParts, "\n\n")
		sysParams = append(sysParams, anthropic.TextBlockParam{Text: sysStr})
	}

	return out, sysParams, nil
}

func getResponseContentFromAnthropicMessage(msg *anthropic.Message) []spec.ResponseContent {
	if msg == nil {
		return []spec.ResponseContent{}
	}
	resp := make([]spec.ResponseContent, 0, len(msg.Content))
	for _, content := range msg.Content {
		switch variant := content.AsAny().(type) {
		case anthropic.TextBlock:
			resp = append(resp, spec.ResponseContent{Type: spec.ResponseContentTypeText, Content: variant.Text})
		case anthropic.ThinkingBlock:
			resp = append(resp, spec.ResponseContent{Type: spec.ResponseContentTypeThinking, Content: variant.Thinking})
		case anthropic.RedactedThinkingBlock:
		case anthropic.ToolUseBlock:
		case anthropic.ServerToolUseBlock:
		case anthropic.WebSearchToolResultBlock:
		default:
			// Invalid variant, dont handle as of now.
		}
	}
	return resp
}
