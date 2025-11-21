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
	AttachmentDocIndex ConversationAttachmentKind = "docIndex"
	AttachmentPR       ConversationAttachmentKind = "pr"
	AttachmentCommit   ConversationAttachmentKind = "commit"
	AttachmentSnapshot ConversationAttachmentKind = "snapshot"
)

// ConversationAttachment is a lightweight handle to external context (files, PRs, snapshots, etc.) that the model can
// read.
type ConversationAttachment struct {
	Kind  ConversationAttachmentKind `json:"kind"`
	Ref   string                     `json:"ref"`   // ID, slug, etc.
	Label string                     `json:"label"` // human-friendly label for UI
}

type ConversationToolChoice struct {
	BundleID    string `json:"bundleID,omitempty"`
	ToolSlug    string `json:"toolSlug"`
	ToolVersion string `json:"toolVersion"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
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
