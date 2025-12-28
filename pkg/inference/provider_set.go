package inference

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"

	"github.com/ppipada/flexigpt-app/pkg/attachment"
	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"
	toolStore "github.com/ppipada/flexigpt-app/pkg/tool/store"
	inferencegoSpec "github.com/ppipada/inference-go/spec"
)

type ProviderSetAPI struct {
	mu sync.RWMutex

	providers map[inferencegoSpec.ProviderName]CompletionProvider
	debug     bool
	toolStore *toolStore.ToolStore
}

// NewProviderSetAPI creates a new ProviderSet with the specified default provider.
func NewProviderSetAPI(
	debug bool,
	ts *toolStore.ToolStore,
) (*ProviderSetAPI, error) {
	return &ProviderSetAPI{
		providers: map[inferencegoSpec.ProviderName]CompletionProvider{},
		debug:     debug,
		toolStore: ts,
	}, nil
}

func (ps *ProviderSetAPI) AddProvider(
	ctx context.Context,
	req *spec.AddProviderRequest,
) (*spec.AddProviderResponse, error) {
	if req == nil || req.Provider == "" || req.Body == nil || req.Body.Origin == "" {
		return nil, errors.New("invalid params")
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	_, exists := ps.providers[req.Provider]
	if exists {
		return nil, errors.New(
			"invalid provider: cannot add a provider with same name as an existing provider, delete first",
		)
	}
	if ok := isProviderSDKTypeSupported(req.Body.SDKType); !ok {
		return nil, errors.New("unsupported provider api type")
	}

	providerInfo := inferencegoSpec.ProviderParam{
		Name:                     req.Provider,
		SDKType:                  req.Body.SDKType,
		APIKey:                   "",
		Origin:                   req.Body.Origin,
		ChatCompletionPathPrefix: req.Body.ChatCompletionPathPrefix,
		APIKeyHeaderKey:          req.Body.APIKeyHeaderKey,
		DefaultHeaders:           req.Body.DefaultHeaders,
	}

	cp, err := getProviderAPI(providerInfo, ps.debug)
	if err != nil {
		return nil, err
	}
	ps.providers[req.Provider] = cp

	slog.Info("add provider", "name", req.Provider)
	return &spec.AddProviderResponse{}, nil
}

func (ps *ProviderSetAPI) DeleteProvider(
	ctx context.Context,
	req *spec.DeleteProviderRequest,
) (*spec.DeleteProviderResponse, error) {
	if req == nil || req.Provider == "" {
		return nil, errors.New("got empty provider input")
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	_, exists := ps.providers[req.Provider]

	if !exists {
		return nil, errors.New(
			"invalid provider: provider does not exist",
		)
	}
	delete(ps.providers, req.Provider)
	slog.Info("deleteProvider", "name", req.Provider)
	return &spec.DeleteProviderResponse{}, nil
}

// SetProviderAPIKey sets the key for a given provider.
func (ps *ProviderSetAPI) SetProviderAPIKey(
	ctx context.Context,
	req *spec.SetProviderAPIKeyRequest,
) (*spec.SetProviderAPIKeyResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("got empty provider input")
	}
	p, exists := ps.providers[req.Provider]
	if !exists {
		return nil, errors.New("invalid provider")
	}
	if req.Body.APIKey == "" {
		err := p.DeInitLLM(ctx)
		return &spec.SetProviderAPIKeyResponse{}, err
	}
	err := p.SetProviderAPIKey(
		ctx,
		req.Body.APIKey,
	)
	if err != nil {
		return nil, err
	}
	err = p.InitLLM(ctx)
	if err != nil {
		return nil, err
	}
	return &spec.SetProviderAPIKeyResponse{}, nil
}

// FetchCompletion processes a completion request for a given provider.
//
// It does prep work first:
//   - message/model validation
//   - construction of FetchCompletionData
//   - tool-store hydration
//   - token-based message filtering
//   - full attachment ContentBlock hydration for all messages
//
// and then invokes the provider-specific LLM API with the prepared FetchCompletionData.
func (ps *ProviderSetAPI) FetchCompletion(
	ctx context.Context,
	req *spec.FetchCompletionRequest,
) (*spec.FetchCompletionResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("got empty fetch completion input")
	}
	provider := req.Provider
	ps.mu.RLock()
	p, exists := ps.providers[provider]
	ps.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid provider")
	}

	// Consolidated prep step: build and hydrate FetchCompletionData from the raw request body.
	completionData, err := ps.prepareFetchCompletionData(ctx, req.Body)
	if err != nil {
		return nil, err
	}

	resp, err := p.FetchCompletion(
		ctx,
		completionData,
		req.OnStreamTextData,
		req.OnStreamThinkingData,
	)
	if err != nil {
		return nil, errors.Join(err, errors.New("error in fetch completion"))
	}

	return resp, nil
}

// prepareFetchCompletionData consolidates all pre-processing.
//   - Validates model + current message.
//   - Ensures current message is from user and has text / attachments / tool outputs.
//   - Copies model params and messages into a new FetchCompletionData (via getCompletionData),
//     normalizing fields (e.g. zeroing Name, copying attachments).
//   - Hydrates tools from the tool store into FetchCompletionData.ToolChoices.
//   - Applies token-based message filtering with FilterMessagesByTokenCount.
//   - After filtering, fully hydrates attachment.ContentBlock for all messages so that SDK
//     wrappers only need to translate common types to provider-specific ones.
func (ps *ProviderSetAPI) prepareFetchCompletionData(
	ctx context.Context,
	body *spec.FetchCompletionRequestBody,
) (*spec.FetchCompletionData, error) {
	if body == nil || body.ModelParams.Name == "" {
		return nil, errors.New("got empty provider/model input")
	}
	if body.CurrentMessage.Role != inferencegoSpec.RoleUser {
		return nil, errors.New("current message must be from user")
	}

	hasText := body.CurrentMessage.Content != nil &&
		strings.TrimSpace(*body.CurrentMessage.Content) != ""
	hasAttachments := len(body.CurrentMessage.Attachments) > 0
	hasToolOutputs := len(body.CurrentMessage.ToolOutputs) > 0
	if !hasText && !hasAttachments && !hasToolOutputs {
		return nil, errors.New("current message must have text or attachments or tooloutputs")
	}

	// Build the initial FetchCompletionData: copy model params, prev messages, and current message.
	completionData := getCompletionData(
		body.ModelParams,
		body.CurrentMessage,
		body.PrevMessages,
	)

	// Attach tool choices: convert the lightweight ToolChoice handles provided
	// by the caller into FetchCompletionToolChoice entries, then hydrate them
	// with full tool definitions from the tool store.
	if len(body.ToolChoices) > 0 {
		out := make([]spec.FetchCompletionToolChoice, 0, len(body.ToolChoices))
		for _, c := range body.ToolChoices {
			out = append(out, spec.FetchCompletionToolChoice{
				ToolChoice: c,
				Tool:       nil, // hydrated later in attachToolsToCompletionData
			})
		}
		completionData.ToolChoices = out
		if err := ps.attachToolsToCompletionData(ctx, completionData); err != nil {
			return nil, err
		}
	}

	// Token-based message filtering: limit the prompt to MaxPromptLength.
	completionData.Messages = FilterMessagesByTokenCount(
		completionData.Messages,
		completionData.ModelParams.MaxPromptLength,
	)
	// After filtering, fully hydrate attachments for all remaining messages so that
	// provider-specific wrappers only translate already-hydrated ContentBlocks.
	if err := hydrateAttachmentsInMessages(ctx, completionData.Messages); err != nil {
		return nil, err
	}

	return completionData, nil
}

func getCompletionData(
	modelParams inferencegoSpec.ModelParam,
	currentMessage spec.ChatCompletionDataMessage,
	prevMessages []spec.ChatCompletionDataMessage,
) *spec.FetchCompletionData {
	completionData := spec.FetchCompletionData{
		ModelParams: inferencegoSpec.ModelParam{
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

	msgs := make([]spec.ChatCompletionDataMessage, 0, len(prevMessages))
	for _, m := range prevMessages {
		msgs = append(msgs, convertBuildMessageToChatMessage(m))
	}
	msgs = append(msgs, convertBuildMessageToChatMessage(currentMessage))

	completionData.Messages = msgs

	return &completionData
}

func convertBuildMessageToChatMessage(
	msg spec.ChatCompletionDataMessage,
) spec.ChatCompletionDataMessage {
	// Keep the role, erase the name.
	out := spec.ChatCompletionDataMessage{
		Role: msg.Role,
		Name: nil,
	}
	if msg.Content != nil {
		c := *msg.Content
		out.Content = &c
	}

	if len(msg.ReasoningContents) > 0 {
		out.ReasoningContents = msg.ReasoningContents
	}

	if len(msg.Attachments) > 0 {
		out.Attachments = attachment.CopyAttachments(msg.Attachments)
	}
	if len(msg.ToolCalls) > 0 {
		out.ToolCalls = msg.ToolCalls
	}
	if len(msg.ToolOutputs) > 0 {
		out.ToolOutputs = msg.ToolOutputs
	}
	return out
}

func (ps *ProviderSetAPI) attachToolsToCompletionData(
	ctx context.Context,
	data *spec.FetchCompletionData,
) error {
	if data == nil {
		return errors.New("invalid completion data: nil")
	}

	type toolKey struct {
		bundleID string
		toolSlug string
		version  string
	}

	toolChoiceIndex := map[toolKey]map[string]struct{}{}

	for _, att := range data.ToolChoices {
		bundleID := att.BundleID
		toolSlug := att.ToolSlug
		version := strings.TrimSpace(att.ToolVersion)

		if bundleID == "" || toolSlug == "" || version == "" {
			return errors.New(
				"invalid tool attachment: bundleID, toolSlug and toolVersion are required",
			)
		}

		k := toolKey{
			bundleID: string(bundleID),
			toolSlug: string(toolSlug),
			version:  version,
		}
		if _, ok := toolChoiceIndex[k]; !ok {
			toolChoiceIndex[k] = make(map[string]struct{})
		}
		if id := strings.TrimSpace(string(att.ToolID)); id != "" {
			toolChoiceIndex[k][id] = struct{}{}
		}
	}

	if len(toolChoiceIndex) == 0 {
		data.ToolChoices = nil
		return nil
	}

	if ps.toolStore == nil {
		return errors.New("tool store not configured for provider set")
	}

	keys := make([]toolKey, 0, len(toolChoiceIndex))
	for k := range toolChoiceIndex {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].bundleID != keys[j].bundleID {
			return keys[i].bundleID < keys[j].bundleID
		}
		if keys[i].toolSlug != keys[j].toolSlug {
			return keys[i].toolSlug < keys[j].toolSlug
		}
		return keys[i].version < keys[j].version
	})

	tools := make([]spec.FetchCompletionToolChoice, 0, len(keys))
	for _, k := range keys {
		req := &toolSpec.GetToolRequest{
			BundleID: bundleitemutils.BundleID(k.bundleID),
			ToolSlug: bundleitemutils.ItemSlug(k.toolSlug),
			Version:  bundleitemutils.ItemVersion(k.version),
		}
		resp, err := ps.toolStore.GetTool(ctx, req)
		if err != nil {
			return fmt.Errorf(
				"failed to load tool %s/%s@%s: %w",
				k.bundleID,
				k.toolSlug,
				k.version,
				err,
			)
		}
		if resp == nil || resp.Body == nil {
			return fmt.Errorf(
				"tool %s/%s@%s not found",
				k.bundleID,
				k.toolSlug,
				k.version,
			)
		}
		tool := *resp.Body
		if !tool.IsEnabled {
			return fmt.Errorf(
				"tool %s/%s@%s is disabled",
				k.bundleID,
				k.toolSlug,
				k.version,
			)
		}

		completionTool := spec.FetchCompletionToolChoice{
			Tool: &tool,
		}
		completionTool.BundleID = bundleitemutils.BundleID(k.bundleID)
		completionTool.ToolSlug = bundleitemutils.ItemSlug(k.toolSlug)
		completionTool.ToolVersion = k.version
		// CompletionTool.BundleSlug = "asd", bundle slug not available here, any may not be available from input.
		completionTool.ToolID = tool.ID
		completionTool.Description = tool.Description
		completionTool.DisplayName = tool.DisplayName

		tools = append(tools, completionTool)
	}

	data.ToolChoices = tools
	return nil
}

// hydrateAttachmentsInMessages ensures that every attachment in the final
// message list has its ContentBlock populated, using the same semantics that
// were previously implemented inside each SDK wrapper.
//
// For the last user message (the current turn), we always rebuild attachment
// content blocks (OverrideOriginal + ForceFetch) so that we see the latest
// file/url state, even if a text-only block was created earlier for token
// counting. For older messages, we respect any existing ContentBlock snapshot
// and only build new ones when missing.
func hydrateAttachmentsInMessages(
	ctx context.Context,
	messages []spec.ChatCompletionDataMessage,
) error {
	if len(messages) == 0 {
		return nil
	}

	// Identify the last user message index; this corresponds to the current turn.
	lastUserIdx := -1
	for i, m := range messages {
		if m.Role == inferencegoSpec.RoleUser {
			lastUserIdx = i
		}
	}

	for idx := range messages {
		overrideOriginal := idx == lastUserIdx
		if err := hydrateAttachmentsForMessage(ctx, &messages[idx], overrideOriginal); err != nil {
			return err
		}
	}
	return nil
}

func hydrateAttachmentsForMessage(
	ctx context.Context,
	m *spec.ChatCompletionDataMessage,
	overrideOriginal bool,
) error {
	if m == nil || len(m.Attachments) == 0 {
		return nil
	}

	for i := range m.Attachments {
		att := &m.Attachments[i]

		// For prior messages, keep any existing ContentBlock snapshot to avoid
		// re-fetching files/URLs unnecessarily.
		if att.ContentBlock != nil && !overrideOriginal {
			continue
		}

		cb, err := att.BuildContentBlock(
			ctx,
			attachment.WithOverrideOriginalContentBlock(overrideOriginal),
			attachment.WithOnlyTextKindContentBlock(false),
			// For the current user turn we always force a fresh read, even if a
			// text-only block was created earlier for token counting.
			attachment.WithForceFetchContentBlock(overrideOriginal),
		)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
				// Respect context cancellation/timeout for the whole request.
				return ctx.Err()
			}
			switch {
			case errors.Is(err, attachment.ErrExistingContentBlock):
				// Attachment already had a content block and we chose not to override it.
				continue

			case errors.Is(err, attachment.ErrAttachmentModifiedSinceSnapshot) && !overrideOriginal:
				// Underlying resource changed since it was originally attached; fall back
				// to a small text placeholder so the model is aware something existed here.
				txt := att.FormatAsDisplayName()
				if txt == "" {
					txt = "[Attachment]"
				}
				txt += " (attachment modified since this message was sent)"
				cb = &attachment.ContentBlock{
					Kind: attachment.ContentBlockText,
					Text: &txt,
				}

			default:
				// Non-fatal: log and skip this attachment.
				slog.Warn("failed to hydrate attachment content block",
					"err", err,
					"attachment", *att,
				)
				continue
			}
		}

		if cb != nil {
			att.ContentBlock = cb
		}
	}

	return nil
}

func isProviderSDKTypeSupported(t inferencegoSpec.ProviderSDKType) bool {
	if t == inferencegoSpec.ProviderSDKTypeAnthropic ||
		t == inferencegoSpec.ProviderSDKTypeOpenAIChatCompletions ||
		t == inferencegoSpec.ProviderSDKTypeOpenAIResponses {
		return true
	}
	return false
}

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

func getProviderAPI(p inferencegoSpec.ProviderParam, debug bool) (CompletionProvider, error) {
	switch p.SDKType {
	case inferencegoSpec.ProviderSDKTypeAnthropic:
		return NewAnthropicMessagesAPI(p, debug)

	case inferencegoSpec.ProviderSDKTypeOpenAIChatCompletions:
		return NewOpenAIChatCompletionsAPI(p, debug)

	case inferencegoSpec.ProviderSDKTypeOpenAIResponses:
		return NewOpenAIResponsesAPI(p, debug)
	}

	return nil, errors.New("invalid provider api type")
}
