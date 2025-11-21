package spec

import (
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"
)

type ChatCompletionRoleEnum string

const (
	System    ChatCompletionRoleEnum = "system"
	Developer ChatCompletionRoleEnum = "developer"
	User      ChatCompletionRoleEnum = "user"
	Assistant ChatCompletionRoleEnum = "assistant"
	Function  ChatCompletionRoleEnum = "function"
	Tool      ChatCompletionRoleEnum = "tool"
)

// ChatCompletionAttachmentKind enumerates contextual attachment categories that can be
// associated with messages sent to the inference layer.
type ChatCompletionAttachmentKind string

const (
	AttachmentFile     ChatCompletionAttachmentKind = "file"
	AttachmentDocIndex ChatCompletionAttachmentKind = "docIndex"
	AttachmentPR       ChatCompletionAttachmentKind = "pr"
	AttachmentCommit   ChatCompletionAttachmentKind = "commit"
	AttachmentSnapshot ChatCompletionAttachmentKind = "snapshot"
)

// ChatCompletionAttachment is a lightweight reference to external context (files, PRs, snapshots, etc.).
type ChatCompletionAttachment struct {
	Kind  ChatCompletionAttachmentKind `json:"kind"`
	Ref   string                       `json:"ref"`
	Label string                       `json:"label"`
}

type ChatCompletionToolChoice struct {
	BundleID    string `json:"bundleID,omitempty"`
	ToolSlug    string `json:"toolSlug"`
	ToolVersion string `json:"toolVersion"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
}

type ChatCompletionDataMessage struct {
	Role    ChatCompletionRoleEnum `json:"role"`
	Content *string                `json:"content,omitempty"`
	Name    *string                `json:"name,omitempty"`
}

type FetchCompletionToolChoice struct {
	BundleID    string         `json:"bundleID,omitempty"`
	ToolSlug    string         `json:"toolSlug"`
	ToolVersion string         `json:"toolVersion"`
	ID          string         `json:"id,omitempty"`
	Tool        *toolSpec.Tool `json:"tool"`
	Description string         `json:"description"`
}

// ModelParams represents input information about a model to a completion.
type ModelParams struct {
	Name                        modelpresetSpec.ModelName        `json:"name"`
	Stream                      bool                             `json:"stream"`
	MaxPromptLength             int                              `json:"maxPromptLength"`
	MaxOutputLength             int                              `json:"maxOutputLength"`
	Temperature                 *float64                         `json:"temperature,omitempty"`
	Reasoning                   *modelpresetSpec.ReasoningParams `json:"reasoning,omitempty"`
	SystemPrompt                string                           `json:"systemPrompt"`
	Timeout                     int                              `json:"timeout"`
	AdditionalParametersRawJSON *string                          `json:"additionalParametersRawJSON"`
}

// ProviderParams represents information about a provider.
type ProviderParams struct {
	Name                     modelpresetSpec.ProviderName    `json:"name"`
	SDKType                  modelpresetSpec.ProviderSDKType `json:"sdkType"`
	APIKey                   string                          `json:"apiKey"`
	Origin                   string                          `json:"origin"`
	ChatCompletionPathPrefix string                          `json:"chatCompletionPathPrefix"`
	APIKeyHeaderKey          string                          `json:"apiKeyHeaderKey"`
	DefaultHeaders           map[string]string               `json:"defaultHeaders"`
}
