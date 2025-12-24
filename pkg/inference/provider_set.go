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

func (ps *ProviderSetAPI) BuildCompletionData(
	ctx context.Context,
	req *spec.BuildCompletionDataRequest,
) (*spec.BuildCompletionDataResponse, error) {
	if req == nil || req.Body == nil || req.Body.ModelParams.Name == "" || req.Body.CurrentMessage.Content == nil {
		return nil, errors.New("got empty provider/model input")
	}
	if req.Body.CurrentMessage.Role != modelpresetSpec.RoleUser {
		return nil, errors.New("current message must be from user")
	}
	hasText := req.Body.CurrentMessage.Content != nil &&
		strings.TrimSpace(*req.Body.CurrentMessage.Content) != ""
	hasAttachments := len(req.Body.CurrentMessage.Attachments) > 0
	hasToolOutputs := len(req.Body.CurrentMessage.ToolOutputs) > 0
	if !hasText && !hasAttachments && !hasToolOutputs {
		return nil, errors.New("current message must have text or attachments or tooloutputs")
	}

	provider := req.Provider
	ps.mu.RLock()
	p, exists := ps.providers[provider]
	ps.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid provider")
	}

	resp, err := p.BuildCompletionData(
		ctx,
		req.Body.ModelParams,
		req.Body.CurrentMessage,
		req.Body.PrevMessages,
	)
	if err != nil {
		return nil, errors.Join(err, errors.New("error in building completion data"))
	}

	// Attach tool choices: convert the lightweight ToolChoice handles
	// provided by the caller into FetchCompletionToolChoice entries, then hydrate
	// them with full tool definitions from the tool store.
	if len(req.Body.ToolChoices) > 0 {
		out := make([]spec.FetchCompletionToolChoice, 0, len(req.Body.ToolChoices))
		for _, c := range req.Body.ToolChoices {
			out = append(out, spec.FetchCompletionToolChoice{
				ToolChoice: c,
				Tool:       nil, // hydrated later in attachToolsToCompletionData
			})
		}
		resp.ToolChoices = out
		if err := ps.attachToolsToCompletionData(ctx, resp); err != nil {
			return nil, err
		}
	}

	// For the current message, try to attach the text blocks for attachments in a best effort basis.
	if len(resp.Messages) > 0 && len(resp.Messages[len(resp.Messages)-1].Attachments) > 0 {
		currentAtts := resp.Messages[len(resp.Messages)-1].Attachments
		for idx := range currentAtts {
			a, err := currentAtts[idx].BuildContentBlock(
				ctx,
				attachment.WithForceFetchContentBlock(true),
				attachment.WithOnlyTextKindContentBlock(true),
				attachment.WithOverrideOriginalContentBlock(true),
			)
			if err == nil && a != nil {
				// This is the only place where we are actually setting content block to the attachment.
				currentAtts[idx].ContentBlock = a
			}
		}
	}

	// Assuming filterMessagesByTokenCount is implemented elsewhere.
	resp.Messages = FilterMessagesByTokenCount(
		resp.Messages,
		resp.ModelParams.MaxPromptLength,
	)

	return &spec.BuildCompletionDataResponse{Body: resp}, nil
}

// FetchCompletion processes a completion request for a given provider.
func (ps *ProviderSetAPI) FetchCompletion(
	ctx context.Context,
	req *spec.FetchCompletionRequest,
) (*spec.FetchCompletionResponse, error) {
	if req == nil || req.Body == nil || req.Body.FetchCompletionData == nil {
		return nil, errors.New("got empty fetch completion input")
	}
	provider := req.Provider
	ps.mu.RLock()
	p, exists := ps.providers[provider]
	ps.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid provider")
	}

	resp, err := p.FetchCompletion(
		ctx,
		req.Body.FetchCompletionData,
		req.Body.OnStreamTextData,
		req.Body.OnStreamThinkingData,
	)
	if err != nil {
		return nil, errors.Join(err, errors.New("error in fetch completion"))
	}

	return resp, nil
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
