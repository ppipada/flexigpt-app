package spec

import (
	"time"
)

const (
	ConversationFileExtension = "json"
	MaxPageSize               = 256
	DefaultPageSize           = 12
)

// ConversationRoleEnum represents the role of a participant in a conversation.
type ConversationRoleEnum string

const (
	ConversationRoleSystem    ConversationRoleEnum = "system"
	ConversationRoleUser      ConversationRoleEnum = "user"
	ConversationRoleAssistant ConversationRoleEnum = "assistant"
	ConversationRoleFunction  ConversationRoleEnum = "function"
	ConversationRoleFeedback  ConversationRoleEnum = "feedback"
)

// ConversationAttachmentKind enumerates high-level categories of contextual attachments
// that a message can reference.
type ConversationAttachmentKind string

const (
	AttachmentFile     ConversationAttachmentKind = "file"
	AttachmentImage    ConversationAttachmentKind = "image"
	AttachmentDocIndex ConversationAttachmentKind = "docIndex"
	AttachmentPR       ConversationAttachmentKind = "pr"
	AttachmentCommit   ConversationAttachmentKind = "commit"
	AttachmentSnapshot ConversationAttachmentKind = "snapshot"
)

type ConversationFileRef struct {
	Path string `json:"path"`
}

type ConversationImageRef struct {
	Path string `json:"path"`
}

type ConversationHandleRef struct {
	Handle string `json:"handle"`
}

// ConversationAttachment stores a lightweight reference to contextual data so that
// the composer can rehydrate attachments when re-editing a prior turn.
type ConversationAttachment struct {
	Kind  ConversationAttachmentKind `json:"kind"`
	Label string                     `json:"label"`

	FileRef    *ConversationFileRef   `json:"fileRef,omitempty"`
	ImageRef   *ConversationImageRef  `json:"imageRef,omitempty"`
	GenericRef *ConversationHandleRef `json:"genericRef,omitempty"`
}

type ConversationToolChoice struct {
	BundleID    string `json:"bundleID"`
	ToolSlug    string `json:"toolSlug"`
	ToolVersion string `json:"toolVersion"`
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
	ID          string `json:"id,omitempty"`
}

// ConversationMessage represents a message in a conversation.
type ConversationMessage struct {
	ID          string                   `json:"id"`
	CreatedAt   *time.Time               `json:"createdAt,omitempty"`
	Role        ConversationRoleEnum     `json:"role"`
	Content     string                   `json:"content"`
	Name        *string                  `json:"name,omitempty"`
	Details     *string                  `json:"details,omitempty"`
	ToolChoices []ConversationToolChoice `json:"toolChoices,omitempty"`
	Attachments []ConversationAttachment `json:"attachments,omitempty"`
}

// Conversation represents a conversation with messages.
type Conversation struct {
	ID         string                `json:"id"`
	Title      string                `json:"title"`
	CreatedAt  time.Time             `json:"createdAt"`
	ModifiedAt time.Time             `json:"modifiedAt"`
	Messages   []ConversationMessage `json:"messages"`
}
