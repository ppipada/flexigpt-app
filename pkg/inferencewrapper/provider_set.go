package inferencewrapper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ppipada/inference-go"
	inferencegoSpec "github.com/ppipada/inference-go/spec"
	"github.com/ppipada/mapstore-go/uuidv7filename"

	"github.com/ppipada/flexigpt-app/pkg/attachment"
	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	conversationSpec "github.com/ppipada/flexigpt-app/pkg/conversation/spec"
	"github.com/ppipada/flexigpt-app/pkg/inferencewrapper/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"
	toolStore "github.com/ppipada/flexigpt-app/pkg/tool/store"
)

// ProviderSetAPI is a thin aggregator on top of inference-go's ProviderSetAPI.
// It owns:
//   - provider lifecycle (add/delete/set API key),
//   - attachment/tool hydration,
//   - mapping Conversation+CurrentTurn -> inference-go FetchCompletionRequest.
type ProviderSetAPI struct {
	inner     *inference.ProviderSetAPI
	toolStore *toolStore.ToolStore
}

// NewProviderSetAPI creates a new aggregator.
//
// logger: optional slog logger used by inference-go (may be nil).
// ts:     tool store used to hydrate ToolChoices when needed.
// opts:   additional inference-go ProviderSetOptions (e.g. WithDebugClientBuilder).
func NewProviderSetAPI(
	logger *slog.Logger,
	ts *toolStore.ToolStore,
	opts ...inference.ProviderSetOption,
) (*ProviderSetAPI, error) {
	allOpts := make([]inference.ProviderSetOption, 0, len(opts)+1)
	if logger != nil {
		allOpts = append(allOpts, inference.WithLogger(logger))
	}
	allOpts = append(allOpts, opts...)

	inner, err := inference.NewProviderSetAPI(allOpts...)
	if err != nil {
		return nil, err
	}

	return &ProviderSetAPI{
		inner:     inner,
		toolStore: ts,
	}, nil
}

// AddProvider forwards to inference-go ProviderSetAPI.AddProvider.
func (ps *ProviderSetAPI) AddProvider(
	ctx context.Context,
	req *spec.AddProviderRequest,
) (*spec.AddProviderResponse, error) {
	if req == nil || req.Body == nil || req.Provider == "" || strings.TrimSpace(req.Body.Origin) == "" {
		return nil, errors.New("invalid params")
	}

	cfg := &inference.AddProviderConfig{
		SDKType:                  req.Body.SDKType,
		Origin:                   req.Body.Origin,
		ChatCompletionPathPrefix: req.Body.ChatCompletionPathPrefix,
		APIKeyHeaderKey:          req.Body.APIKeyHeaderKey,
		DefaultHeaders:           req.Body.DefaultHeaders,
	}
	if _, err := ps.inner.AddProvider(ctx, req.Provider, cfg); err != nil {
		return nil, err
	}
	slog.Info("add provider", "name", req.Provider)
	return &spec.AddProviderResponse{}, nil
}

// DeleteProvider forwards to inference-go ProviderSetAPI.DeleteProvider.
func (ps *ProviderSetAPI) DeleteProvider(
	ctx context.Context,
	req *spec.DeleteProviderRequest,
) (*spec.DeleteProviderResponse, error) {
	if req == nil || req.Provider == "" {
		return nil, errors.New("got empty provider input")
	}
	if err := ps.inner.DeleteProvider(ctx, req.Provider); err != nil {
		return nil, err
	}
	slog.Info("deleteProvider", "name", req.Provider)
	return &spec.DeleteProviderResponse{}, nil
}

// SetProviderAPIKey forwards to inference-go ProviderSetAPI.SetProviderAPIKey.
func (ps *ProviderSetAPI) SetProviderAPIKey(
	ctx context.Context,
	req *spec.SetProviderAPIKeyRequest,
) (*spec.SetProviderAPIKeyResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("got empty provider input")
	}
	if err := ps.inner.SetProviderAPIKey(ctx, req.Provider, req.Body.APIKey); err != nil {
		return nil, err
	}
	return &spec.SetProviderAPIKeyResponse{}, nil
}

// FetchCompletion builds a normalized inference-go FetchCompletionRequest from
// app-level conversation types and calls inference-go's FetchCompletion.
func (ps *ProviderSetAPI) FetchCompletion(
	ctx context.Context,
	req *spec.CompletionRequest,
) (*spec.CompletionResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("got empty completion input")
	}
	if req.Provider == "" {
		return nil, errors.New("missing provider")
	}

	body := req.Body

	// Resolve model param for this call (prefer explicit body.ModelParam,
	// otherwise last non-nil ModelParam from history).
	modelParam, err := ps.resolveModelParam(body)
	if err != nil {
		return nil, err
	}
	if modelParam.Name == "" {
		return nil, errors.New("model name is required")
	}

	if len(body.Current.ToolChoices) > 0 {
		return nil, errors.New("prepopulated tool choices are not allowed in fetch completion, need tool store choices")
	}

	// Flatten full conversation (history + current) into InputUnion list.
	inputs, currentInputs, err := ps.buildInputs(ctx, body)
	if err != nil {
		return nil, err
	}
	if len(inputs) == 0 {
		return nil, errors.New("no usable inputs to send to inference-go")
	}

	// Build tool choices for this call.
	toolChoices, choiceBindings, err := ps.buildToolChoices(ctx, body.ToolStoreChoices)
	if err != nil {
		return nil, err
	}

	infReq := &inferencegoSpec.FetchCompletionRequest{
		ModelParam:  *modelParam,
		Inputs:      inputs,
		ToolChoices: toolChoices,
	}

	var opts *inferencegoSpec.FetchCompletionOptions
	if req.OnStreamText != nil || req.OnStreamThinking != nil {
		opts = &inferencegoSpec.FetchCompletionOptions{
			StreamHandler: makeStreamHandler(req.OnStreamText, req.OnStreamThinking),
		}
	}

	b, err := ps.inner.FetchCompletion(ctx, req.Provider, infReq, opts)

	resp := &spec.CompletionResponse{Body: &spec.CompletionResponseBody{
		InferenceResponse:     b,
		HydratedCurrentInputs: currentInputs,
	}}

	if len(choiceBindings) > 0 {
		bindings := make([]spec.ToolCallBinding, 0, len(choiceBindings))
		for choiceID, sc := range choiceBindings {
			bindings = append(bindings, spec.ToolCallBinding{
				ChoiceID:        choiceID,
				ToolStoreChoice: sc,
			})
		}
		resp.Body.ToolCallBindings = bindings
	}

	return resp, err
}

// resolveModelParam chooses the effective ModelParam for this call.
//
// Priority:
//  1. body.ModelParam if non-nil.
//  2. Last non-nil History[i].ModelParam.
//
// If still empty, returns an error.
func (ps *ProviderSetAPI) resolveModelParam(
	body *spec.CompletionRequestBody,
) (*inferencegoSpec.ModelParam, error) {
	if body.ModelParam != nil {
		cp := *body.ModelParam
		return &cp, nil
	}
	// Look backwards through history for the last non-nil ModelParam.
	for i := len(body.History) - 1; i >= 0; i-- {
		if body.History[i].ModelParam != nil {
			cp := *body.History[i].ModelParam
			return &cp, nil
		}
	}
	return &inferencegoSpec.ModelParam{}, nil
}

// buildInputs flattens History + Current into a single InputUnion slice.
//
// For History:
//   - If msg.Inputs/Outputs are non-empty, we just replay them in order.
//
// For Current:
//   - If Current.Inputs is non-empty, we replay them as-is.
//   - Otherwise, we build one InputKindInputMessage from:
//   - Current.Messages (user text),
//   - Current.Attachments (hydrated to text/image/file blocks).
func (ps *ProviderSetAPI) buildInputs(
	ctx context.Context,
	body *spec.CompletionRequestBody,
) (all, current []inferencegoSpec.InputUnion, err error) {
	out := make([]inferencegoSpec.InputUnion, 0)

	// 1) History: replay stored unions exactly as they were.
	for _, turn := range body.History {
		// Inputs first, then Outputs, preserving stored order.

		out = append(out, turn.Inputs...)
		for _, outEv := range turn.Outputs {
			// Outputs are not directly part of InputUnion; but for replay
			// we want them to be visible as prior context. We embed them
			// as InputUnion using the matching InputKind* variants.
			out = append(out, outputToInput(outEv))
		}
	}

	cur := body.Current
	if cur.Role != inferencegoSpec.RoleUser {
		return nil, nil, errors.New("current turn must have role=user")
	}

	// If the caller already provided normalized InputUnions, just reuse them.
	currentOut := make([]inferencegoSpec.InputUnion, 0)
	if len(cur.Inputs) > 0 {
		currentOut = append(currentOut, cur.Inputs...)
	}

	// Always process attachments into content items.
	msgContentItems, err := buildContentItemsFromAttachments(ctx, &cur)
	if err != nil {
		return nil, nil, err
	}

	if len(msgContentItems) > 0 {
		// Try to merge into the last user InputMessage if present.
		merged := false
		for i := len(currentOut) - 1; i >= 0; i-- {
			iu := &currentOut[i]
			if iu.Kind == inferencegoSpec.InputKindInputMessage &&
				iu.InputMessage != nil &&
				iu.InputMessage.Role == inferencegoSpec.RoleUser {
				iu.InputMessage.Contents = append(iu.InputMessage.Contents, msgContentItems...)
				merged = true
				break
			}
		}

		if !merged {
			inputMsg := inferencegoSpec.InputOutputContent{
				ID:       "",
				Role:     inferencegoSpec.RoleUser,
				Status:   inferencegoSpec.StatusCompleted,
				Contents: msgContentItems,
			}
			currentOut = append(currentOut, inferencegoSpec.InputUnion{
				Kind:         inferencegoSpec.InputKindInputMessage,
				InputMessage: &inputMsg,
			})
		}
	}
	if len(currentOut) > 0 {
		out = append(out, currentOut...)
	}

	if len(out) == 0 {
		return nil, nil, errors.New("no usable inputs to send to inference-go")
	}

	return out, currentOut, nil
}

// outputToInput converts an OutputUnion from a previous completion into an
// InputUnion so it can be replayed as prior context in the next call.
func outputToInput(o inferencegoSpec.OutputUnion) inferencegoSpec.InputUnion {
	switch o.Kind {
	case inferencegoSpec.OutputKindOutputMessage:
		return inferencegoSpec.InputUnion{
			Kind:          inferencegoSpec.InputKindOutputMessage,
			OutputMessage: o.OutputMessage,
		}
	case inferencegoSpec.OutputKindReasoningMessage:
		return inferencegoSpec.InputUnion{
			Kind:             inferencegoSpec.InputKindReasoningMessage,
			ReasoningMessage: o.ReasoningMessage,
		}
	case inferencegoSpec.OutputKindFunctionToolCall:
		return inferencegoSpec.InputUnion{
			Kind:             inferencegoSpec.InputKindFunctionToolCall,
			FunctionToolCall: o.FunctionToolCall,
		}
	case inferencegoSpec.OutputKindCustomToolCall:
		return inferencegoSpec.InputUnion{
			Kind:           inferencegoSpec.InputKindCustomToolCall,
			CustomToolCall: o.CustomToolCall,
		}
	case inferencegoSpec.OutputKindWebSearchToolCall:
		return inferencegoSpec.InputUnion{
			Kind:              inferencegoSpec.InputKindWebSearchToolCall,
			WebSearchToolCall: o.WebSearchToolCall,
		}
	case inferencegoSpec.OutputKindWebSearchToolOutput:
		return inferencegoSpec.InputUnion{
			Kind:                inferencegoSpec.InputKindWebSearchToolOutput,
			WebSearchToolOutput: o.WebSearchToolOutput,
		}
	default:
		// Unknown kinds are dropped.
		return inferencegoSpec.InputUnion{}
	}
}

// buildContentItemsFromTurn builds InputOutputContentItemUnion list for the
// current user turn from its Messages (text only) and Attachments.
//
// Attachment rules:
//   - If Attachment.ContentBlock is already present, we reuse it.
//   - Otherwise we call BuildContentBlocks WITHOUT ForceFetch, so any
//     existing ContentBlock is kept and only missing ones are hydrated.
func buildContentItemsFromAttachments(
	ctx context.Context,
	turn *conversationSpec.ConversationMessage,
) ([]inferencegoSpec.InputOutputContentItemUnion, error) {
	items := make([]inferencegoSpec.InputOutputContentItemUnion, 0)
	if len(turn.Attachments) == 0 {
		return items, nil
	}

	blocks, err := attachment.BuildContentBlocks(
		ctx,
		turn.Attachments,
		attachment.WithOverrideOriginalContentBlock(true),
		// IMPORTANT: do NOT ForceFetch; this ensures that if a ContentBlock
		// is already present (e.g., from a stored conversation), we reuse it.
		attachment.WithOnlyTextKindContentBlock(false),
	)
	if err != nil {
		return nil, err
	}

	for _, b := range blocks {
		switch b.Kind {
		case attachment.ContentBlockText:
			if b.Text == nil {
				continue
			}
			txt := strings.TrimSpace(*b.Text)
			if txt == "" {
				continue
			}
			items = append(items, inferencegoSpec.InputOutputContentItemUnion{
				Kind: inferencegoSpec.ContentItemKindText,
				TextItem: &inferencegoSpec.ContentItemText{
					Text: txt,
				},
			})

		case attachment.ContentBlockImage:
			if b.Base64Data == nil {
				continue
			}
			data := strings.TrimSpace(*b.Base64Data)
			if data == "" {
				continue
			}
			mime := inferencegoSpec.DefaultImageDataMIME
			if b.MIMEType != nil && strings.TrimSpace(*b.MIMEType) != "" {
				mime = strings.TrimSpace(*b.MIMEType)
			}
			name := ""
			if b.FileName != nil {
				name = strings.TrimSpace(*b.FileName)
			}
			items = append(items, inferencegoSpec.InputOutputContentItemUnion{
				Kind: inferencegoSpec.ContentItemKindImage,
				ImageItem: &inferencegoSpec.ContentItemImage{
					ImageName: name,
					ImageMIME: mime,
					ImageData: data,
					Detail:    inferencegoSpec.ImageDetailAuto,
				},
			})

		case attachment.ContentBlockFile:
			if b.Base64Data == nil {
				continue
			}
			data := strings.TrimSpace(*b.Base64Data)
			if data == "" {
				continue
			}
			mime := inferencegoSpec.DefaultFileDataMIME
			if b.MIMEType != nil && strings.TrimSpace(*b.MIMEType) != "" {
				mime = strings.TrimSpace(*b.MIMEType)
			}
			name := ""
			if b.FileName != nil {
				name = strings.TrimSpace(*b.FileName)
			}
			items = append(items, inferencegoSpec.InputOutputContentItemUnion{
				Kind: inferencegoSpec.ContentItemKindFile,
				FileItem: &inferencegoSpec.ContentItemFile{
					FileName: name,
					FileMIME: mime,
					FileData: data,
				},
			})
		}
	}

	return items, nil
}

func (ps *ProviderSetAPI) buildToolChoices(
	ctx context.Context,
	toolStoreChoices []toolSpec.ToolStoreChoice,
) ([]inferencegoSpec.ToolChoice, map[string]toolSpec.ToolStoreChoice, error) {
	out := make([]inferencegoSpec.ToolChoice, 0)
	choiceBindings := make(map[string]toolSpec.ToolStoreChoice)
	if len(toolStoreChoices) == 0 {
		return nil, nil, nil
	}

	if ps.toolStore == nil {
		return nil, nil, errors.New("tool store not configured for provider set")
	}

	for _, sc := range toolStoreChoices {
		if sc.BundleID == "" || sc.ToolSlug == "" || strings.TrimSpace(sc.ToolVersion) == "" {
			return nil, nil, fmt.Errorf(
				"invalid tool store choice: bundleID/toolSlug/toolVersion required: %+v",
				sc,
			)
		}

		tc, choiceID, err := ps.hydrateToolChoice(ctx, sc)
		if err != nil {
			return nil, nil, err
		}
		if choiceID == "" {
			// Somehow we received a empty choice id. Skip this.
			continue
		}
		out = append(out, *tc)
		choiceBindings[choiceID] = sc
	}

	if len(out) == 0 {
		return nil, nil, nil
	}
	return out, choiceBindings, nil
}

// hydrateToolChoice loads the Tool definition from tool-store and converts it
// into an inference-go ToolChoice. This is only called when we don't already
// have a ToolChoice persisted in the conversation for the same tool.
func (ps *ProviderSetAPI) hydrateToolChoice(
	ctx context.Context,
	sc toolSpec.ToolStoreChoice,
) (toolChoice *inferencegoSpec.ToolChoice, choiceID string, err error) {
	req := &toolSpec.GetToolRequest{
		BundleID: sc.BundleID,
		ToolSlug: sc.ToolSlug,
		Version:  bundleitemutils.ItemVersion(sc.ToolVersion),
	}
	resp, err := ps.toolStore.GetTool(ctx, req)
	if err != nil {
		return nil, "", fmt.Errorf(
			"failed to load tool %s/%s@%s: %w",
			sc.BundleID,
			sc.ToolSlug,
			sc.ToolVersion,
			err,
		)
	}
	if resp == nil || resp.Body == nil {
		return nil, "", fmt.Errorf(
			"tool %s/%s@%s not found",
			sc.BundleID,
			sc.ToolSlug,
			sc.ToolVersion,
		)
	}
	tool := resp.Body
	if !tool.IsEnabled {
		return nil, "", fmt.Errorf(
			"tool %s/%s@%s is disabled",
			sc.BundleID,
			sc.ToolSlug,
			sc.ToolVersion,
		)
	}
	if !tool.LLMCallable {
		return nil, "", fmt.Errorf(
			"tool %s/%s@%s is not LLM-callable",
			sc.BundleID, sc.ToolSlug, sc.ToolVersion,
		)
	}
	name := string(sc.ToolSlug)
	desc := tool.Description
	if desc == "" {
		desc = sc.Description
	}
	choiceID, _ = uuidv7filename.NewUUIDv7String()
	if choiceID == "" {
		choiceID = name
	}

	tc := &inferencegoSpec.ToolChoice{
		Type:        inferencegoSpec.ToolType(sc.ToolType),
		ID:          choiceID,
		Name:        name,
		Description: desc,
	}

	// If this is a web-search tool, populate minimal WebSearchArguments so that
	// toolChoicesToOpenAIResponseTools will actually emit a web_search tool.
	if sc.ToolType == toolSpec.ToolTypeWebSearch {
		tc.WebSearchArguments = &inferencegoSpec.WebSearchToolChoiceItem{
			SearchContextSize: "medium", // "low" | "medium" | "high"
		}
	} else {
		argSchema, err := decodeToolArgSchema(string(tool.ArgSchema))
		if err != nil {
			return nil, "", fmt.Errorf(
				"invalid argSchema for %s/%s@%s: %w",
				sc.BundleID,
				sc.ToolSlug,
				sc.ToolVersion,
				err,
			)
		}
		tc.Arguments = argSchema
	}

	return tc, choiceID, nil
}

func decodeToolArgSchema(raw toolSpec.JSONRawString) (map[string]any, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return map[string]any{"type": "object"}, nil
	}
	var schema map[string]any
	if err := json.Unmarshal([]byte(s), &schema); err != nil {
		return nil, err
	}
	if len(schema) == 0 {
		schema = map[string]any{"type": "object"}
	}
	return schema, nil
}

// makeStreamHandler adapts inference-go streaming into the legacy
// text/thinking callback pair.
func makeStreamHandler(
	onText func(string) error,
	onThinking func(string) error,
) inferencegoSpec.StreamHandler {
	if onText == nil && onThinking == nil {
		return nil
	}
	return func(ev inferencegoSpec.StreamEvent) error {
		switch ev.Kind {
		case inferencegoSpec.StreamContentKindText:
			if onText != nil && ev.Text != nil {
				return onText(ev.Text.Text)
			}
		case inferencegoSpec.StreamContentKindThinking:
			if onThinking != nil && ev.Thinking != nil {
				return onThinking(ev.Thinking.Text)
			}
		}
		return nil
	}
}

// ConvertModelPresetToInferencegoSDKType maps the app's provider SDK enum to
// inference-go's ProviderSDKType.
func ConvertModelPresetToInferencegoSDKType(inSDK modelpresetSpec.ProviderSDKType) inferencegoSpec.ProviderSDKType {
	switch inSDK {
	case modelpresetSpec.ProviderSDKTypeAnthropic:
		return inferencegoSpec.ProviderSDKTypeAnthropic
	case modelpresetSpec.ProviderSDKTypeOpenAIChatCompletions:
		return inferencegoSpec.ProviderSDKTypeOpenAIChatCompletions
	case modelpresetSpec.ProviderSDKTypeOpenAIResponses:
		return inferencegoSpec.ProviderSDKTypeOpenAIResponses
	default:
		return ""
	}
}
