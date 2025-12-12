package spec

import (
	"time"

	"github.com/ppipada/flexigpt-app/pkg/attachment"
	modelpresetSpec "github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	toolSpec "github.com/ppipada/flexigpt-app/pkg/tool/spec"
)

const (
	ConversationFileExtension = "json"
	MaxPageSize               = 256
	DefaultPageSize           = 12
)

// ConversationMessage represents a message in a conversation.
type ConversationMessage struct {
	ID        string                   `json:"id"`
	CreatedAt *time.Time               `json:"createdAt,omitempty"`
	Role      modelpresetSpec.RoleEnum `json:"role"`
	Content   string                   `json:"content"`
	Name      *string                  `json:"name,omitempty"`
	Details   *string                  `json:"details,omitempty"`

	ReasoningContents []modelpresetSpec.ReasoningContent `json:"reasoningContents,omitempty"`
	ToolChoices       []toolSpec.ToolChoice              `json:"toolChoices,omitempty"`
	Attachments       []attachment.Attachment            `json:"attachments,omitempty"`
	ToolCalls         []toolSpec.ToolCall                `json:"toolCalls,omitempty"`
	ToolOutputs       []toolSpec.ToolOutput              `json:"toolOutputs,omitempty"`
	Usage             *modelpresetSpec.Usage             `json:"usage,omitempty"`
}

// Conversation represents a conversation with messages.
type Conversation struct {
	ID         string                `json:"id"`
	Title      string                `json:"title"`
	CreatedAt  time.Time             `json:"createdAt"`
	ModifiedAt time.Time             `json:"modifiedAt"`
	Messages   []ConversationMessage `json:"messages"`
}
