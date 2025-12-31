package spec

import (
	conversationSpec "github.com/ppipada/flexigpt-app/pkg/conversation/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"
	inferencegoSpec "github.com/ppipada/inference-go/spec"
)

type AddProviderRequestBody struct {
	SDKType                  inferencegoSpec.ProviderSDKType `json:"sdkType"`
	Origin                   string                          `json:"origin"`
	ChatCompletionPathPrefix string                          `json:"chatCompletionPathPrefix"`
	APIKeyHeaderKey          string                          `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string               `json:"defaultHeaders"`
}

type AddProviderRequest struct {
	Provider inferencegoSpec.ProviderName `path:"provider" required:"true"`
	Body     *AddProviderRequestBody
}

type AddProviderResponse struct{}

type DeleteProviderRequest struct {
	Provider inferencegoSpec.ProviderName `path:"provider" required:"true"`
}

type DeleteProviderResponse struct{}

type SetProviderAPIKeyRequestBody struct {
	APIKey string `json:"apiKey" required:"true"`
}

type SetProviderAPIKeyRequest struct {
	Provider inferencegoSpec.ProviderName `path:"provider" required:"true"`
	Body     *SetProviderAPIKeyRequestBody
}

type SetProviderAPIKeyResponse struct{}

type CompletionRequestBody struct {
	// Model configuration for this *call*. If nil, the aggregator can fall
	// back to the last non-nil ModelParam from History.
	ModelParam *inferencegoSpec.ModelParam `json:"modelParam,omitempty"`

	// Past turns of the conversation, already persisted.
	History []conversationSpec.ConversationMessage `json:"history"`

	// New user turn to complete. Must have Role=user. Typically will have:
	//   - Attachments (ref attachments),
	//   - either:
	//       * pre-built InputUnion(s) in Inputs, or
	//       * just Messages + Attachments and let the aggregator build
	//         the InputUnion for this turn.
	Current conversationSpec.ConversationMessage `json:"current"`

	// Additional tool-store handles the UI wants to enable for this call.
	// The aggregator will hydrate any missing ToolChoice definitions from
	// tool-store, but if you already provide ToolChoices on History or
	// Current, those will be reused without re-hydration.
	ToolStoreChoices []toolSpec.ToolStoreChoice `json:"toolStoreChoices,omitempty"`
}

type CompletionRequest struct {
	Provider inferencegoSpec.ProviderName `path:"provider" required:"true"`
	Body     *CompletionRequestBody

	OnStreamText     func(text string) error     `json:"-"`
	OnStreamThinking func(thinking string) error `json:"-"`
}

type ToolCallBinding struct {
	// ChoiceID is the inference-go ToolChoice.ID that this tool call refers to.
	ChoiceID string `json:"choiceID"`
	// ToolStoreChoice is the original tool-store handle the UI supplied.
	ToolStoreChoice toolSpec.ToolStoreChoice `json:"toolStoreChoice"`
}

type CompletionResponseBody struct {
	InferenceResponse *inferencegoSpec.FetchCompletionResponse `json:"inferenceResponse,omitempty"`
	ToolCallBindings  []ToolCallBinding                        `json:"toolCallBindings,omitempty"`
}

type CompletionResponse struct {
	Body *CompletionResponseBody
}
