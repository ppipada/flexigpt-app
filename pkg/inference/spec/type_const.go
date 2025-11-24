package spec

import (
	"time"

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
	AttachmentImage    ChatCompletionAttachmentKind = "image"
	AttachmentDocIndex ChatCompletionAttachmentKind = "docIndex"
	AttachmentPR       ChatCompletionAttachmentKind = "pr"
	AttachmentCommit   ChatCompletionAttachmentKind = "commit"
	AttachmentSnapshot ChatCompletionAttachmentKind = "snapshot"
)

// ChatCompletionFileRef carries metadata for file attachments.
type ChatCompletionFileRef struct {
	Path      string     `json:"path"`
	Exists    bool       `json:"exists"`
	SizeBytes int64      `json:"sizeBytes,omitempty"`
	ModTime   *time.Time `json:"modTime,omitempty"`
}

// ChatCompletionImageRef carries metadata for image attachments.
type ChatCompletionImageRef struct {
	Path      string     `json:"path"`
	Exists    bool       `json:"exists"`
	Width     int        `json:"width,omitempty"`
	Height    int        `json:"height,omitempty"`
	Format    string     `json:"format,omitempty"`
	SizeBytes int64      `json:"sizeBytes,omitempty"`
	ModTime   *time.Time `json:"modTime,omitempty"`
}

// ChatCompletionHandleRef preserves legacy handle-style references (PRs, doc indexes, etc.).
type ChatCompletionHandleRef struct {
	Handle string `json:"handle"`
}

// ChatCompletionAttachment is a lightweight reference to external context (files, docs, images, etc.).
type ChatCompletionAttachment struct {
	Kind  ChatCompletionAttachmentKind `json:"kind"`
	Label string                       `json:"label"`

	// Exactly one field below must be non-nil.
	FileRef    *ChatCompletionFileRef   `json:"fileRef,omitempty"`
	ImageRef   *ChatCompletionImageRef  `json:"imageRef,omitempty"`
	GenericRef *ChatCompletionHandleRef `json:"genericRef,omitempty"`
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
