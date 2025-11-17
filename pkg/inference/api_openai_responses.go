package inference

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/responses"
	"github.com/openai/openai-go/v2/shared"
	openaiSharedConstant "github.com/openai/openai-go/v2/shared/constant"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

// OpenAIResponsesAPI struct that implements the CompletionProvider interface.
type OpenAIResponsesAPI struct {
	ProviderParams *spec.ProviderParams
	Debug          bool
	client         *openai.Client
}

func NewOpenAIResponsesAPI(
	pi spec.ProviderParams,
	debug bool,
) (*OpenAIResponsesAPI, error) {
	if pi.Name == "" || pi.Origin == "" {
		return nil, errors.New("openai responses api LLM: invalid args")
	}
	return &OpenAIResponsesAPI{
		ProviderParams: &pi,
		Debug:          debug,
	}, nil
}

func (api *OpenAIResponsesAPI) InitLLM(ctx context.Context) error {
	if !api.IsConfigured(ctx) {
		slog.Debug(
			string(
				api.ProviderParams.Name,
			) + ": No API key given. Not initializing OpenAIResponsesAPI LLM object",
		)
		return nil
	}

	opts := []option.RequestOption{
		option.WithAPIKey(api.ProviderParams.APIKey),
	}

	providerURL := modelpresetSpec.DefaultOpenAIOrigin
	if api.ProviderParams.Origin != "" {
		baseURL := api.ProviderParams.Origin
		// Remove trailing slash from baseURL if present.
		baseURL = strings.TrimSuffix(baseURL, "/")

		pathPrefix := api.ProviderParams.ChatCompletionPathPrefix
		// Remove 'responses' from pathPrefix if present,
		// This is because openai sdk adds 'chat/completions' internally.
		pathPrefix = strings.TrimSuffix(
			pathPrefix,
			"responses",
		)
		providerURL = baseURL + pathPrefix
		opts = append(opts, option.WithBaseURL(strings.TrimSuffix(providerURL, "/")))
	}

	for k, v := range api.ProviderParams.DefaultHeaders {
		opts = append(opts, option.WithHeader(strings.TrimSpace(k), strings.TrimSpace(v)))
	}

	if api.ProviderParams.APIKeyHeaderKey != "" &&
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

	c := openai.NewClient(opts...)
	api.client = &c
	slog.Info(
		"openai responses api LLM provider initialized",
		"name",
		string(api.ProviderParams.Name),
		"URL",
		providerURL,
	)
	return nil
}

func (api *OpenAIResponsesAPI) DeInitLLM(ctx context.Context) error {
	api.client = nil
	slog.Info(
		"openai responses api LLM: provider de initialized",
		"name",
		string(api.ProviderParams.Name),
	)
	return nil
}

func (api *OpenAIResponsesAPI) GetProviderInfo(ctx context.Context) *spec.ProviderParams {
	return api.ProviderParams
}

func (api *OpenAIResponsesAPI) IsConfigured(ctx context.Context) bool {
	return api.ProviderParams.APIKey != ""
}

// SetProviderAPIKey sets the key for a provider.
func (api *OpenAIResponsesAPI) SetProviderAPIKey(
	ctx context.Context,
	apiKey string,
) error {
	if apiKey == "" {
		return errors.New("openai responses api LLM: invalid apikey provided")
	}
	if api.ProviderParams == nil {
		return errors.New("openai responses api LLM: no ProviderParams found")
	}

	api.ProviderParams.APIKey = apiKey

	return nil
}

func (api *OpenAIResponsesAPI) BuildCompletionData(
	ctx context.Context,
	modelParams spec.ModelParams,
	currentMessage spec.ChatCompletionDataMessage,
	prevMessages []spec.ChatCompletionDataMessage,
) (*spec.CompletionData, error) {
	return getCompletionData(modelParams, currentMessage, prevMessages), nil
}

func (api *OpenAIResponsesAPI) FetchCompletion(
	ctx context.Context,
	completionData *spec.CompletionData,
	onStreamTextData, onStreamThinkingData func(string) error,
) (*spec.CompletionResponse, error) {
	if api.client == nil {
		return nil, errors.New("openai responses api LLM: client not initialized")
	}
	if completionData == nil || len(completionData.Messages) == 0 {
		return nil, errors.New("openai responses api LLM: empty completion data")
	}

	// Build OpenAI chat messages.
	msgs, err := toOpenAIResponsesMessages(
		completionData.Messages,
		completionData.ModelParams.Name,
		api.ProviderParams.Name,
	)
	if err != nil {
		return nil, err
	}

	params := responses.ResponseNewParams{
		Model:           shared.ChatModel(completionData.ModelParams.Name),
		MaxOutputTokens: openai.Int(int64(completionData.ModelParams.MaxOutputLength)),
		Input:           responses.ResponseNewParamsInputUnion{OfInputItemList: msgs},
		Store:           openai.Bool(false),
	}
	sysPrompt := strings.TrimSpace(completionData.ModelParams.SystemPrompt)
	if sysPrompt != "" {
		params.Instructions = openai.String(sysPrompt)
	}
	if completionData.ModelParams.Temperature != nil {
		params.Temperature = openai.Float(*completionData.ModelParams.Temperature)
	}

	if rp := completionData.ModelParams.Reasoning; rp != nil &&
		rp.Type == modelpresetSpec.ReasoningTypeSingleWithLevels {
		switch rp.Level {
		case
			modelpresetSpec.ReasoningLevelNone,
			modelpresetSpec.ReasoningLevelMinimal,
			modelpresetSpec.ReasoningLevelLow,
			modelpresetSpec.ReasoningLevelMedium,
			modelpresetSpec.ReasoningLevelHigh:
			params.Reasoning = shared.ReasoningParam{
				Effort: shared.ReasoningEffort(string(rp.Level)),
				// No option to tweak reasoning summary for now.
				Summary: shared.ReasoningSummaryAuto,
			}
		default:
			return nil, fmt.Errorf("invalid level %q for singleWithLevels", rp.Level)

		}
	}
	if len(completionData.Tools) > 0 {
		toolDefs, err := toOpenAIResponseTools(completionData.Tools)
		if err != nil {
			return nil, err
		}
		if len(toolDefs) > 0 {
			params.Tools = toolDefs
		}
	}
	timeout := modelpresetSpec.DefaultAPITimeout
	if completionData.ModelParams.Timeout > 0 {
		timeout = time.Duration(completionData.ModelParams.Timeout) * time.Second
	}

	if completionData.ModelParams.Stream && onStreamTextData != nil && onStreamThinkingData != nil {
		return api.doStreaming(ctx, params, onStreamTextData, onStreamThinkingData, timeout)
	}
	return api.doNonStreaming(ctx, params, timeout)
}

func (api *OpenAIResponsesAPI) doNonStreaming(
	ctx context.Context,
	params responses.ResponseNewParams,
	timeout time.Duration,
) (*spec.CompletionResponse, error) {
	completionResp := &spec.CompletionResponse{}
	ctx = AddDebugResponseToCtx(ctx)
	resp, err := api.client.Responses.New(ctx, params, option.WithRequestTimeout(timeout))
	isNilResp := resp == nil || len(resp.Output) == 0
	attachDebugResp(ctx, completionResp, err, isNilResp)
	if isNilResp {
		return completionResp, nil
	}

	completionResp.ResponseContent = getResponseContentFromOpenAIOutput(resp)
	return completionResp, nil
}

func (api *OpenAIResponsesAPI) doStreaming(
	ctx context.Context,
	params responses.ResponseNewParams,
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

	var respFull responses.Response

	completionResp := &spec.CompletionResponse{}
	ctx = AddDebugResponseToCtx(ctx)
	stream := api.client.Responses.NewStreaming(
		ctx,
		params,
		option.WithRequestTimeout(timeout),
	)

	var streamWriteErr error
	for stream.Next() {
		chunk := stream.Current()
		if chunk.Type == "response.output_text.delta" {
			streamWriteErr = writeTextData(chunk.Delta)
			if streamWriteErr != nil {
				break
			}
		}
		if chunk.Type == "response.reasoning_summary_text.delta" {
			streamWriteErr = writeThinkingData(chunk.Delta)
			if streamWriteErr != nil {
				break
			}
		}
		if chunk.Type == "response.reasoning_text.delta" {
			streamWriteErr = writeThinkingData(chunk.Delta)
			if streamWriteErr != nil {
				break
			}
		}

		if chunk.Type == "response.completed" {
			respFull = chunk.Response
			break
		}
		if chunk.Type == "response.failed" {
			respFull = chunk.Response
			streamWriteErr = fmt.Errorf("API failed, %s", respFull.Error.RawJSON())
			break
		}
		if chunk.Type == "response.incomplete" {
			respFull = chunk.Response
			streamWriteErr = fmt.Errorf(
				"API finished as incomplete, %s",
				respFull.IncompleteDetails.Reason,
			)
			break
		}
	}
	if flushTextData != nil {
		flushTextData()
	}

	if flushThinkingData != nil {
		flushThinkingData()
	}

	streamErr := errors.Join(stream.Err(), streamWriteErr)
	isNilResp := len(respFull.Output) == 0
	attachDebugResp(ctx, completionResp, streamErr, isNilResp)
	if isNilResp {
		return completionResp, nil
	}

	slog.Info("resp json", "full", respFull.RawJSON())

	respContent := getResponseContentFromOpenAIOutput(&respFull)
	completionResp.ResponseContent = respContent
	return completionResp, streamErr
}

func toOpenAIResponsesMessages(
	messages []spec.ChatCompletionDataMessage,
	modelName modelpresetSpec.ModelName,
	providerName modelpresetSpec.ProviderName,
) (responses.ResponseInputParam, error) {
	var out responses.ResponseInputParam

	for _, m := range messages {
		if m.Content == nil {
			continue
		}
		switch m.Role {
		case spec.System:
			// This is in case additional dev or system message is present in the array itself.
			inputParam := responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(*m.Content),
					},
					Role: responses.EasyInputMessageRoleSystem,
					Type: responses.EasyInputMessageTypeMessage,
				},
			}
			out = append(out, inputParam)
		case spec.Developer:
			// This is in case additional dev or system message is present in the array itself.
			inputParam := responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(*m.Content),
					},
					Role: responses.EasyInputMessageRoleDeveloper,
					Type: responses.EasyInputMessageTypeMessage,
				},
			}
			out = append(out, inputParam)

		case spec.User:
			// This is in case additional dev or system message is present in the array itself.
			inputParam := responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(*m.Content),
					},
					Role: responses.EasyInputMessageRoleUser,
					Type: responses.EasyInputMessageTypeMessage,
				},
			}
			out = append(out, inputParam)

		case spec.Assistant:
			// This is in case additional dev or system message is present in the array itself.
			inputParam := responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(*m.Content),
					},
					Role: responses.EasyInputMessageRoleAssistant,
					Type: responses.EasyInputMessageTypeMessage,
				},
			}
			out = append(out, inputParam)

		case spec.Function, spec.Tool:
		}
	}
	return out, nil
}

func getResponseContentFromOpenAIOutput(
	inputResp *responses.Response,
) []spec.ResponseContent {
	if inputResp == nil || len(inputResp.Output) == 0 {
		return []spec.ResponseContent{}
	}
	resp := make([]spec.ResponseContent, 0, 2)
	var outputText strings.Builder
	var thinkingSummaryText strings.Builder
	var thinkingText strings.Builder

	for _, item := range inputResp.Output {
		if item.Type == string(openaiSharedConstant.Message("").Default()) {
			for _, content := range item.Content {
				if content.Type == string(openaiSharedConstant.OutputText("").Default()) {
					outputText.WriteString(content.Text)
				}
			}
		}
		if item.Type == string(openaiSharedConstant.Reasoning("").Default()) {
			ti := item.AsReasoning()
			for _, content := range ti.Content {
				thinkingText.WriteString(content.Text)
			}
			for _, content := range ti.Summary {
				thinkingSummaryText.WriteString(content.Text)
			}
		}
	}

	thinkingSummaryStr := thinkingSummaryText.String()
	if thinkingSummaryStr != "" {
		resp = append(
			resp,
			spec.ResponseContent{
				Type:    spec.ResponseContentTypeThinkingSummary,
				Content: thinkingSummaryStr,
			},
		)
	}

	thinkingStr := thinkingText.String()
	if thinkingStr != "" {
		resp = append(
			resp,
			spec.ResponseContent{Type: spec.ResponseContentTypeThinking, Content: thinkingStr},
		)
	}
	outStr := outputText.String()

	resp = append(
		resp,
		spec.ResponseContent{Type: spec.ResponseContentTypeText, Content: outStr},
	)

	return resp
}

func toOpenAIResponseTools(
	tools []spec.CompletionTool,
) ([]responses.ToolUnionParam, error) {
	if len(tools) == 0 {
		return nil, nil
	}
	out := make([]responses.ToolUnionParam, 0, len(tools))
	for _, ct := range tools {
		schema, err := decodeToolArgSchema(ct.Tool.ArgSchema)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid tool schema for %s/%s@%s: %w",
				ct.BundleID,
				ct.Tool.Slug,
				ct.Tool.Version,
				err,
			)
		}
		tool := responses.FunctionToolParam{
			Name:       toolFunctionName(ct),
			Parameters: schema,
			Type:       openaiSharedConstant.Function("function"),
		}
		if desc := toolDescription(ct); desc != "" {
			tool.Description = openai.String(desc)
		}
		out = append(out, responses.ToolUnionParam{OfFunction: &tool})
	}
	return out, nil
}
