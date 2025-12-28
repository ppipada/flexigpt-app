package inference

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
	openaiSharedConstant "github.com/openai/openai-go/v3/shared/constant"

	"github.com/ppipada/flexigpt-app/pkg/attachment"
	"github.com/ppipada/flexigpt-app/pkg/fileutil"
	"github.com/ppipada/flexigpt-app/pkg/inference/debugclient"
	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"

	inferencegoSpec "github.com/ppipada/inference-go/spec"
)

// OpenAIResponsesAPI struct that implements the CompletionProvider interface.
type OpenAIResponsesAPI struct {
	ProviderParams *inferencegoSpec.ProviderParam
	Debug          bool
	client         *openai.Client
}

func NewOpenAIResponsesAPI(
	pi inferencegoSpec.ProviderParam,
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
		// This is because openai sdk adds 'responses' internally.
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

	dbgCfg := debugclient.DefaultDebugConfig
	dbgCfg.LogToSlog = api.Debug

	newClient := debugclient.NewDebugHTTPClient(dbgCfg)

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

func (api *OpenAIResponsesAPI) GetProviderInfo(ctx context.Context) *inferencegoSpec.ProviderParam {
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

func (api *OpenAIResponsesAPI) FetchCompletion(
	ctx context.Context,
	completionData *spec.FetchCompletionData,
	onStreamTextData, onStreamThinkingData func(string) error,
) (*spec.FetchCompletionResponse, error) {
	if api.client == nil {
		return nil, errors.New("openai responses api LLM: client not initialized")
	}
	if completionData == nil || len(completionData.Messages) == 0 {
		return nil, errors.New("openai responses api LLM: empty completion data")
	}

	// Build OpenAI Responses input messages.
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
		Include:         []responses.ResponseIncludable{"reasoning.encrypted_content"},
	}

	sysPrompt := strings.TrimSpace(completionData.ModelParams.SystemPrompt)
	if sysPrompt != "" {
		params.Instructions = openai.String(sysPrompt)
	}
	if completionData.ModelParams.Temperature != nil {
		params.Temperature = openai.Float(*completionData.ModelParams.Temperature)
	}

	if rp := completionData.ModelParams.Reasoning; rp != nil &&
		rp.Type == inferencegoSpec.ReasoningTypeSingleWithLevels {
		switch rp.Level {
		case
			inferencegoSpec.ReasoningLevelNone,
			inferencegoSpec.ReasoningLevelMinimal,
			inferencegoSpec.ReasoningLevelLow,
			inferencegoSpec.ReasoningLevelMedium,
			inferencegoSpec.ReasoningLevelHigh,
			inferencegoSpec.ReasoningLevelXHigh:
			params.Reasoning = shared.ReasoningParam{
				Effort:  shared.ReasoningEffort(string(rp.Level)),
				Summary: shared.ReasoningSummaryAuto,
			}
		default:
			return nil, fmt.Errorf("invalid level %q for singleWithLevels", rp.Level)
		}
	}

	// Build tools + tool name map.
	var toolChoiceNameMap map[string]spec.FetchCompletionToolChoice
	if len(completionData.ToolChoices) > 0 {
		toolDefs, m, err := toOpenAIResponseTools(completionData.ToolChoices)
		if err != nil {
			return nil, err
		}
		if len(toolDefs) > 0 {
			params.Tools = toolDefs
			toolChoiceNameMap = m
		}
	}

	timeout := modelpresetSpec.DefaultAPITimeout
	if completionData.ModelParams.Timeout > 0 {
		timeout = time.Duration(completionData.ModelParams.Timeout) * time.Second
	}

	if completionData.ModelParams.Stream && onStreamTextData != nil && onStreamThinkingData != nil {
		return api.doStreaming(ctx, params, onStreamTextData, onStreamThinkingData, timeout, toolChoiceNameMap)
	}
	return api.doNonStreaming(ctx, params, timeout, toolChoiceNameMap)
}

func (api *OpenAIResponsesAPI) doNonStreaming(
	ctx context.Context,
	params responses.ResponseNewParams,
	timeout time.Duration,
	toolChoiceNameMap map[string]spec.FetchCompletionToolChoice,
) (*spec.FetchCompletionResponse, error) {
	completionResp := &spec.FetchCompletionResponse{Body: &spec.FetchCompletionResponseBody{}}
	ctx = debugclient.AddDebugResponseToCtx(ctx)

	resp, err := api.client.Responses.New(ctx, params, option.WithRequestTimeout(timeout))
	isNilResp := resp == nil || len(resp.Output) == 0

	attachDebugResp(ctx, completionResp.Body, err, isNilResp, "", resp)
	completionResp.Body.Usage = usageFromOpenAIResponse(resp)

	if isNilResp {
		return completionResp, nil
	}

	attachContentFromResponses(resp, toolChoiceNameMap, completionResp.Body)

	return completionResp, nil
}

func (api *OpenAIResponsesAPI) doStreaming(
	ctx context.Context,
	params responses.ResponseNewParams,
	onStreamTextData, onStreamThinkingData func(string) error,
	timeout time.Duration,
	toolChoiceNameMap map[string]spec.FetchCompletionToolChoice,
) (*spec.FetchCompletionResponse, error) {
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

	completionResp := &spec.FetchCompletionResponse{Body: &spec.FetchCompletionResponseBody{}}
	ctx = debugclient.AddDebugResponseToCtx(ctx)
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

	attachDebugResp(ctx, completionResp.Body, streamErr, isNilResp, "", respFull)
	completionResp.Body.Usage = usageFromOpenAIResponse(&respFull)

	if isNilResp {
		return completionResp, nil
	}

	attachContentFromResponses(&respFull, toolChoiceNameMap, completionResp.Body)

	return completionResp, streamErr
}

func toOpenAIResponsesMessages(
	messages []spec.ChatCompletionDataMessage,
	modelName inferencegoSpec.ModelName,
	providerName inferencegoSpec.ProviderName,
) (responses.ResponseInputParam, error) {
	var out responses.ResponseInputParam

	for _, m := range messages {
		// Skip only if there is no text, no attachments, no tool outputs, and no tool calls.
		if m.Content == nil &&
			len(m.Attachments) == 0 &&
			len(m.ToolOutputs) == 0 &&
			len(m.ToolCalls) == 0 {
			continue
		}

		// Normalized trimmed content string (may be empty).
		content := ""
		if m.Content != nil {
			content = strings.TrimSpace(*m.Content)
		}

		switch m.Role {
		case inferencegoSpec.RoleSystem:
			if content == "" {
				continue
			}
			out = append(out, responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(content),
					},
					Role: responses.EasyInputMessageRoleSystem,
					Type: responses.EasyInputMessageTypeMessage,
				},
			})

		case inferencegoSpec.RoleDeveloper:
			if content == "" {
				continue
			}
			out = append(out, responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(content),
					},
					Role: responses.EasyInputMessageRoleDeveloper,
					Type: responses.EasyInputMessageTypeMessage,
				},
			})

		case inferencegoSpec.RoleUser:
			if m.Content == nil && len(m.Attachments) == 0 && len(m.ToolOutputs) == 0 {
				continue
			}
			// Build attachments for this specific user turn (if any).
			var attachmentContent []responses.ResponseInputContentUnionParam
			if len(m.Attachments) > 0 {
				var err error
				attachmentContent, err = attachmentsToOpenAI(m.Attachments)
				if err != nil {
					return nil, err
				}
			}

			// First send tool outputs.
			if len(m.ToolOutputs) > 0 {
				o := toolOutputsToOpenAIResponses(m.ToolOutputs)
				out = append(out, o...)
			}

			var parts responses.ResponseInputMessageContentListParam
			if content != "" {
				parts = append(
					parts,
					responses.ResponseInputContentParamOfInputText(content),
				)
			}

			// No attachments: plain text user message.
			if len(attachmentContent) != 0 {
				parts = append(parts, attachmentContent...)
			}

			if len(parts) > 0 {
				out = append(out, responses.ResponseInputItemParamOfMessage(
					parts,
					responses.EasyInputMessageRoleUser,
				))
			}

		case inferencegoSpec.RoleAssistant:
			// Give thinking blocks first, then content then tools.

			if len(m.ReasoningContents) > 0 {
				parts := make([]responses.ResponseInputItemUnionParam, 0)
				for _, reasoningContent := range m.ReasoningContents {
					if reasoningContent.Type != modelpresetSpec.ReasoningContentTypeOpenAIResponses ||
						reasoningContent.ContentOpenAIResponses == nil {
						// We cannot attach non responses reasoning to reasoning call.
						// This can happen in cross model threads.
						continue
					}

					r := &responses.ResponseReasoningItemParam{
						ID:               reasoningContent.ContentOpenAIResponses.ID,
						EncryptedContent: param.NewOpt(reasoningContent.ContentOpenAIResponses.EncryptedContent),
					}
					r.Summary = make(
						[]responses.ResponseReasoningItemSummaryParam,
						0,
						max(len(reasoningContent.ContentOpenAIResponses.Summary)),
					)

					if len(reasoningContent.ContentOpenAIResponses.Summary) > 0 {
						for _, rs := range reasoningContent.ContentOpenAIResponses.Summary {
							r.Summary = append(r.Summary, responses.ResponseReasoningItemSummaryParam{
								Text: rs,
							})
						}
					}

					if len(reasoningContent.ContentOpenAIResponses.Content) > 0 {

						r.Content = make(
							[]responses.ResponseReasoningItemContentParam,
							0,
							len(reasoningContent.ContentOpenAIResponses.Content),
						)
						for _, rc := range reasoningContent.ContentOpenAIResponses.Content {
							r.Content = append(r.Content, responses.ResponseReasoningItemContentParam{
								Text: rc,
							})
						}
					}

					parts = append(parts, responses.ResponseInputItemUnionParam{
						OfReasoning: r,
					})
				}
				if len(parts) > 0 {
					out = append(out, parts...)
				}
			}

			if content != "" {
				// Here annotation handling is a bit incomplete as of now.
				// Need to collect outputs with IDs and then pass back as proper output blocks when completing
				// citations.
				r := &responses.ResponseInputItemUnionParam{}
				annotations := make([]responses.ResponseOutputTextAnnotationUnionParam, 0)
				if len(m.Citations) > 0 {
					for _, c := range m.Citations {
						if c.Kind == modelpresetSpec.CitationKindURLOpenAIResponses &&
							c.URLCitationOpenAIResponses != nil {
							ra := responses.ResponseOutputTextAnnotationUnionParam{
								OfURLCitation: &responses.ResponseOutputTextAnnotationURLCitationParam{
									EndIndex:   c.URLCitationOpenAIResponses.EndIndex,
									StartIndex: c.URLCitationOpenAIResponses.StartIndex,
									Title:      c.URLCitationOpenAIResponses.Title,
									URL:        c.URLCitationOpenAIResponses.URL,
								},
							}
							annotations = append(annotations, ra)
						}
					}
				}
				if len(annotations) > 0 {
					r.OfOutputMessage = &responses.ResponseOutputMessageParam{
						ID: "tmp", // Need to plug this.
						Content: []responses.ResponseOutputMessageContentUnionParam{
							{
								OfOutputText: &responses.ResponseOutputTextParam{
									Text:        content,
									Annotations: annotations,
								},
							},
						},
						Status: "completed", // Need to plug this.
					}
					r.OfMessage.Content = responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(content),
					}
				} else {
					r.OfMessage = &responses.EasyInputMessageParam{
						Role: responses.EasyInputMessageRoleAssistant,
						Type: responses.EasyInputMessageTypeMessage,
						Content: responses.EasyInputMessageContentUnionParam{
							OfString: openai.String(content),
						},
					}
				}
				out = append(out, *r)
			}

			if len(m.ToolCalls) != 0 {
				parts := make([]responses.ResponseInputItemUnionParam, 0)

				for _, toolCall := range m.ToolCalls {
					switch toolCall.Type {
					case string(openaiSharedConstant.FunctionCall("").Default()):
						fc := responses.ResponseFunctionToolCallParam{
							ID:        param.NewOpt(toolCall.ID),
							CallID:    toolCall.CallID,
							Name:      toolCall.Name,
							Arguments: toolCall.Arguments,
							Type:      openaiSharedConstant.FunctionCall("").Default(),
						}

						parts = append(parts,
							responses.ResponseInputItemUnionParam{
								OfFunctionCall: &fc,
							},
						)

					case string(openaiSharedConstant.CustomToolCall("").Default()):
						cc := responses.ResponseCustomToolCallParam{
							ID:     param.NewOpt(toolCall.ID),
							CallID: toolCall.CallID,
							Name:   toolCall.Name,
							Input:  toolCall.Arguments,
							Type:   openaiSharedConstant.CustomToolCall("").Default(),
						}

						parts = append(parts,
							responses.ResponseInputItemUnionParam{
								OfCustomToolCall: &cc,
							},
						)
					}
				}

				if len(parts) != 0 {
					out = append(out, parts...)
				}
			}

		case inferencegoSpec.RoleFunction, inferencegoSpec.RoleTool:
			if len(m.ToolOutputs) > 0 {
				o := toolOutputsToOpenAIResponses(m.ToolOutputs)
				out = append(out, o...)
			}
			if strings.TrimSpace(content) != "" {
				out = append(out, responses.ResponseInputItemUnionParam{
					OfMessage: &responses.EasyInputMessageParam{
						Content: responses.EasyInputMessageContentUnionParam{
							OfString: openai.String(content),
						},
						Role: responses.EasyInputMessageRoleUser,
						Type: responses.EasyInputMessageTypeMessage,
					},
				})
			}
		}
	}
	return out, nil
}

func toolOutputsToOpenAIResponses(
	toolOutputs []toolSpec.ToolOutput,
) []responses.ResponseInputItemUnionParam {
	out := make([]responses.ResponseInputItemUnionParam, 0, len(toolOutputs))

	// Tool outputs with a callID can be represented as proper tool messages.
	for _, o := range toolOutputs {
		callID := strings.TrimSpace(o.CallID)
		if callID == "" {
			continue
		}
		toolContent := strings.TrimSpace(o.RawOutput)
		if toolContent == "" {
			toolContent = strings.TrimSpace(o.Summary)
		}
		if toolContent == "" {
			continue
		}
		out = append(out, responses.ResponseInputItemParamOfFunctionCallOutput(callID, toolContent))
	}

	// Any tool outputs without a callID are rendered as plain user text
	// so the model still sees the information.
	var orphanOutputs []toolSpec.ToolOutput
	for _, o := range toolOutputs {
		if strings.TrimSpace(o.CallID) == "" {
			orphanOutputs = append(orphanOutputs, o)
		}
	}
	if len(orphanOutputs) > 0 {
		if text := renderToolOutputsAsText(orphanOutputs); text != "" {
			o := responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Content: responses.EasyInputMessageContentUnionParam{
						OfString: openai.String(text),
					},
					Role: responses.EasyInputMessageRoleUser,
					Type: responses.EasyInputMessageTypeMessage,
				},
			}

			out = append(out, o)
		}
	}
	return out
}

func attachmentsToOpenAI(
	atts []attachment.Attachment,
) ([]responses.ResponseInputContentUnionParam, error) {
	if len(atts) == 0 {
		return nil, nil
	}
	out := make([]responses.ResponseInputContentUnionParam, 0, len(atts))

	for _, att := range atts {
		if att.ContentBlock == nil {
			continue
		}
		b := att.ContentBlock

		txt := ""
		if b.Text != nil {
			txt = strings.TrimSpace(*b.Text)
		}
		bData := ""
		if b.Base64Data != nil {
			bData = strings.TrimSpace(*b.Base64Data)
		}
		mime := ""
		if b.MIMEType != nil {
			mime = strings.TrimSpace(*b.MIMEType)
		}
		fname := ""
		if b.FileName != nil {
			fname = strings.TrimSpace(*b.FileName)
		}
		switch b.Kind {
		case attachment.ContentBlockText:
			if txt == "" {
				continue
			}
			out = append(
				out,
				responses.ResponseInputContentParamOfInputText(txt),
			)

		case attachment.ContentBlockImage:
			if bData == "" {
				continue
			}
			if mime == "" {
				mime = string(fileutil.DefaultImageMIME)
			}
			dataURL := fmt.Sprintf("data:%s;base64,%s", mime, bData)
			img := responses.ResponseInputImageParam{
				Detail:   responses.ResponseInputImageDetailAuto,
				ImageURL: param.NewOpt(dataURL),
			}
			out = append(
				out,
				responses.ResponseInputContentUnionParam{OfInputImage: &img},
			)

		case attachment.ContentBlockFile:
			if bData == "" {
				continue
			}

			if mime == "" {
				mime = string(fileutil.MIMEApplicationOctetStream)
			}
			dataURL := fmt.Sprintf("data:%s;base64,%s", mime, bData)
			fileParam := responses.ResponseInputFileParam{
				FileData: param.NewOpt(dataURL),
				Filename: param.NewOpt(fname),
				Type:     openaiSharedConstant.InputFile("").Default(),
			}
			out = append(
				out,
				responses.ResponseInputContentUnionParam{OfInputFile: &fileParam},
			)

		default:
			continue
		}
	}
	return out, nil
}

func toOpenAIResponseTools(
	tools []spec.FetchCompletionToolChoice,
) ([]responses.ToolUnionParam, map[string]spec.FetchCompletionToolChoice, error) {
	if len(tools) == 0 {
		return nil, nil, nil
	}

	ordered, nameMap := buildToolChoiceNameMapping(tools)
	out := make([]responses.ToolUnionParam, 0, len(ordered))

	for _, tw := range ordered {
		ct := tw.Choice

		schema, err := decodeToolArgSchema(ct.Tool.ArgSchema)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"invalid tool schema for %s/%s@%s: %w",
				ct.BundleID,
				ct.Tool.Slug,
				ct.Tool.Version,
				err,
			)
		}

		tool := responses.FunctionToolParam{
			// Use the derived name based only on slug + disambiguation.
			Name:       tw.Name,
			Parameters: schema,
			Type:       openaiSharedConstant.Function("function"),
		}
		if desc := toolDescription(ct); desc != "" {
			tool.Description = openai.String(desc)
		}

		out = append(out, responses.ToolUnionParam{OfFunction: &tool})
	}

	return out, nameMap, nil
}

func attachContentFromResponses(
	inputResp *responses.Response,
	toolChoiceNameMap map[string]spec.FetchCompletionToolChoice,
	fetchCompletionResponseBody *spec.FetchCompletionResponseBody,
) {
	if inputResp == nil || len(inputResp.Output) == 0 {
		c := ""
		fetchCompletionResponseBody.Content = &c
		return
	}

	var outputText strings.Builder
	reasoningItems := make([]modelpresetSpec.ReasoningContent, 0)
	toolCalls := make([]toolSpec.ToolCall, 0)
	citations := make([]modelpresetSpec.Citation, 0)
	for _, item := range inputResp.Output {
		switch item.Type {
		case string(openaiSharedConstant.Message("").Default()):
			for _, content := range item.Content {
				if content.Type == string(openaiSharedConstant.OutputText("").Default()) {
					outputText.WriteString(content.Text)
					outputText.WriteString("\n")
					if len(content.Annotations) > 0 {
						for _, ca := range content.Annotations {
							if ca.Type == string(openaiSharedConstant.URLCitation("").Default()) {
								mc := modelpresetSpec.Citation{
									Kind: modelpresetSpec.CitationKindURLOpenAIResponses,
									URLCitationOpenAIResponses: &modelpresetSpec.URLCitationOpenAIResponses{
										URL:        ca.URL,
										Title:      ca.Title,
										StartIndex: ca.StartIndex,
										EndIndex:   ca.EndIndex,
									},
								}
								citations = append(citations, mc)
							}
						}
					}
				}
			}

		case string(openaiSharedConstant.Reasoning("").Default()):
			ti := item.AsReasoning()
			reasoningItem := modelpresetSpec.ReasoningContent{
				Type: modelpresetSpec.ReasoningContentTypeOpenAIResponses,
				ContentOpenAIResponses: &modelpresetSpec.ReasoningContentOpenAIResponses{
					ID:               ti.ID,
					EncryptedContent: ti.EncryptedContent,
				},
			}
			if len(ti.Summary) > 0 {
				reasoningItem.ContentOpenAIResponses.Summary = make([]string, 0, len(ti.Summary))
				for _, content := range ti.Summary {
					reasoningItem.ContentOpenAIResponses.Summary = append(
						reasoningItem.ContentOpenAIResponses.Summary,
						content.Text,
					)
				}
			}

			if len(ti.Content) > 0 {
				reasoningItem.ContentOpenAIResponses.Content = make([]string, 0, len(ti.Content))
				for _, content := range ti.Content {
					reasoningItem.ContentOpenAIResponses.Content = append(
						reasoningItem.ContentOpenAIResponses.Content,
						content.Text,
					)
				}
			}
			reasoningItems = append(reasoningItems, reasoningItem)

		case string(openaiSharedConstant.FunctionCall("").Default()):
			fn := item.AsFunctionCall()
			if fn.CallID == "" || strings.TrimSpace(fn.Name) == "" {
				continue
			}
			name := fn.Name
			var toolChoice *toolSpec.ToolChoice
			if toolChoiceNameMap != nil {
				if ct, ok := toolChoiceNameMap[name]; ok {
					// Add the actual choice to response.
					toolChoice = &ct.ToolChoice
				}
			}

			toolCalls = append(
				toolCalls,
				toolSpec.ToolCall{
					ID:         fn.ID,
					CallID:     fn.CallID,
					Name:       fn.Name,
					Arguments:  fn.Arguments,
					Type:       item.Type,
					Status:     string(fn.Status),
					ToolChoice: toolChoice,
				},
			)

		case string(openaiSharedConstant.CustomToolCall("").Default()):
			fn := item.AsCustomToolCall()
			if fn.CallID == "" || strings.TrimSpace(fn.Name) == "" {
				continue
			}
			name := fn.Name
			var toolChoice *toolSpec.ToolChoice
			if toolChoiceNameMap != nil {
				if ct, ok := toolChoiceNameMap[name]; ok {
					toolChoice = &ct.ToolChoice
				}
			}

			toolCalls = append(
				toolCalls,
				toolSpec.ToolCall{
					ID:         fn.ID,
					CallID:     fn.CallID,
					Name:       fn.Name,
					Arguments:  fn.Input,
					Type:       item.Type,
					ToolChoice: toolChoice,
				},
			)
		}
	}

	outStr := outputText.String()
	fetchCompletionResponseBody.Content = &outStr
	if len(reasoningItems) > 0 {
		fetchCompletionResponseBody.ReasoningContents = reasoningItems
	}
	if len(toolCalls) > 0 {
		fetchCompletionResponseBody.ToolCalls = toolCalls
	}
	if len(citations) > 0 {
		fetchCompletionResponseBody.Citations = citations
	}
}

// usageFromOpenAIResponse normalizes OpenAI Responses API usage into spec.Usage.
func usageFromOpenAIResponse(resp *responses.Response) *inferencegoSpec.Usage {
	uOut := &inferencegoSpec.Usage{}
	if resp == nil {
		return uOut
	}

	u := resp.Usage

	uOut.InputTokensTotal = u.InputTokens
	uOut.InputTokensCached = u.InputTokensDetails.CachedTokens
	uOut.InputTokensUncached = max(u.InputTokens-u.InputTokensDetails.CachedTokens, 0)
	uOut.OutputTokens = u.OutputTokens
	uOut.ReasoningTokens = u.OutputTokensDetails.ReasoningTokens

	return uOut
}
