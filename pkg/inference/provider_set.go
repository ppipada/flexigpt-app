package inference

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"
	toolStore "github.com/ppipada/flexigpt-app/pkg/tool/store"
)

type ProviderSetAPI struct {
	providers map[modelpresetSpec.ProviderName]CompletionProvider
	debug     bool
	toolStore *toolStore.ToolStore
}

// NewProviderSetAPI creates a new ProviderSet with the specified default provider.
func NewProviderSetAPI(
	debug bool,
	ts *toolStore.ToolStore,
) (*ProviderSetAPI, error) {
	return &ProviderSetAPI{
		providers: map[modelpresetSpec.ProviderName]CompletionProvider{},
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
	_, exists := ps.providers[req.Provider]
	if exists {
		return nil, errors.New(
			"invalid provider: cannot add a provider with same name as an existing provider, delete first",
		)
	}
	if ok := isProviderSDKTypeSupported(req.Body.SDKType); !ok {
		return nil, errors.New("unsupported provider api type")
	}

	providerInfo := spec.ProviderParams{
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
	if req == nil || req.Body == nil || req.Body.ModelParams.Name == "" ||
		req.Body.CurrentMessage.Content == nil ||
		req.Body.CurrentMessage.Role != spec.User {
		return nil, errors.New("got empty provider/model input")
	}
	if req.Body.CurrentMessage.Content == nil || req.Body.CurrentMessage.Role != spec.User {
		return nil, errors.New("got invalid current message input")
	}
	if strings.TrimSpace(*req.Body.CurrentMessage.Content) == "" {
		return nil, errors.New("got empty current message input")
	}
	provider := req.Provider
	p, exists := ps.providers[provider]
	if !exists {
		return nil, errors.New("invalid provider")
	}

	prevMessages := make([]spec.ChatCompletionDataMessage, 0, len(req.Body.PrevMessages))
	for _, m := range req.Body.PrevMessages {
		prevMessages = append(prevMessages, convertBuildMessageToChatMessage(m))
	}
	currentMessage := convertBuildMessageToChatMessage(req.Body.CurrentMessage)

	resp, err := p.BuildCompletionData(
		ctx,
		req.Body.ModelParams,
		currentMessage,
		prevMessages,
	)
	if err != nil {
		return nil, errors.Join(err, errors.New("error in building completion data"))
	}
	if err := ps.attachToolsToCompletionData(ctx, resp); err != nil {
		return nil, err
	}

	return &spec.BuildCompletionDataResponse{Body: resp}, nil
}

// FetchCompletion processes a completion request for a given provider.
func (ps *ProviderSetAPI) FetchCompletion(
	ctx context.Context,
	req *spec.FetchCompletionRequest,
) (*spec.FetchCompletionResponse, error) {
	if req == nil || req.Body == nil || req.Body.CompletionData == nil {
		return nil, errors.New("got empty fetch completion input")
	}
	provider := req.Provider
	p, exists := ps.providers[provider]
	if !exists {
		return nil, errors.New("invalid provider")
	}

	resp, err := p.FetchCompletion(
		ctx,
		req.Body.CompletionData,
		req.Body.OnStreamTextData,
		req.Body.OnStreamThinkingData,
	)
	if err != nil {
		return nil, errors.Join(err, errors.New("error in fetch completion"))
	}

	return &spec.FetchCompletionResponse{Body: resp}, nil
}

func (ps *ProviderSetAPI) attachToolsToCompletionData(
	ctx context.Context,
	data *spec.CompletionData,
) error {
	if data == nil {
		return errors.New("invalid completion data: nil")
	}

	type toolKey struct {
		bundleID string
		toolSlug string
		version  string
	}

	attachmentIndex := map[toolKey]map[string]struct{}{}

	for _, msg := range data.Messages {
		for _, att := range msg.ToolAttachments {
			bundleID := strings.TrimSpace(att.BundleID)
			toolSlug := strings.TrimSpace(att.ToolSlug)
			version := strings.TrimSpace(att.ToolVersion)

			if bundleID == "" || toolSlug == "" || version == "" {
				return errors.New(
					"invalid tool attachment: bundleID, toolSlug and toolVersion are required",
				)
			}

			k := toolKey{
				bundleID: bundleID,
				toolSlug: toolSlug,
				version:  version,
			}
			if _, ok := attachmentIndex[k]; !ok {
				attachmentIndex[k] = make(map[string]struct{})
			}
			if id := strings.TrimSpace(att.ID); id != "" {
				attachmentIndex[k][id] = struct{}{}
			}
		}
	}

	if len(attachmentIndex) == 0 {
		data.Tools = nil
		return nil
	}

	if ps.toolStore == nil {
		return errors.New("tool store not configured for provider set")
	}

	keys := make([]toolKey, 0, len(attachmentIndex))
	for k := range attachmentIndex {
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

	tools := make([]spec.CompletionTool, 0, len(keys))
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

		completionTool := spec.CompletionTool{
			BundleID: k.bundleID,
			Tool:     tool,
		}
		if ids := attachmentIndex[k]; len(ids) > 0 {
			completionTool.AttachmentIDs = make([]string, 0, len(ids))
			for id := range ids {
				completionTool.AttachmentIDs = append(completionTool.AttachmentIDs, id)
			}
			sort.Strings(completionTool.AttachmentIDs)
		}
		tools = append(tools, completionTool)
	}

	data.Tools = tools
	return nil
}

func convertBuildMessageToChatMessage(
	msg spec.BuildCompletionDataMessage,
) spec.ChatCompletionDataMessage {
	out := spec.ChatCompletionDataMessage{
		Role: msg.Role,
	}
	if msg.Content != nil {
		c := *msg.Content
		out.Content = &c
	}
	if msg.Name != nil {
		n := *msg.Name
		out.Name = &n
	}
	if len(msg.ToolAttachments) > 0 {
		attachments := make([]spec.ChatCompletionToolAttachment, 0, len(msg.ToolAttachments))
		for _, att := range msg.ToolAttachments {
			attachment := spec.ChatCompletionToolAttachment{
				BundleID:    strings.TrimSpace(att.BundleID),
				ToolSlug:    strings.TrimSpace(att.ToolSlug),
				ToolVersion: strings.TrimSpace(att.ToolVersion),
			}
			if id := strings.TrimSpace(att.ID); id != "" {
				attachment.ID = id
			}
			attachments = append(attachments, attachment)
		}
		out.ToolAttachments = attachments
	}
	return out
}

func isProviderSDKTypeSupported(t modelpresetSpec.ProviderSDKType) bool {
	if t == modelpresetSpec.ProviderSDKTypeAnthropic ||
		t == modelpresetSpec.ProviderSDKTypeOpenAIChatCompletions ||
		t == modelpresetSpec.ProviderSDKTypeOpenAIResponses {
		return true
	}
	return false
}

func getProviderAPI(p spec.ProviderParams, debug bool) (CompletionProvider, error) {
	switch p.SDKType {
	case modelpresetSpec.ProviderSDKTypeAnthropic:
		return NewAnthropicMessagesAPI(p, debug)

	case modelpresetSpec.ProviderSDKTypeOpenAIChatCompletions:
		return NewOpenAIChatCompletionsAPI(p, debug)

	case modelpresetSpec.ProviderSDKTypeOpenAIResponses:
		return NewOpenAIResponsesAPI(p, debug)
	}

	return nil, errors.New("invalid provider api type")
}
