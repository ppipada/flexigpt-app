package spec

import (
	"time"
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

// ConversationMessage represents a message in a conversation.
type ConversationMessage struct {
	ID        string               `json:"id"`
	CreatedAt *time.Time           `json:"createdAt,omitempty"`
	Role      ConversationRoleEnum `json:"role"`
	Content   string               `json:"content"`
	Name      *string              `json:"name,omitempty"`
	Details   *string              `json:"details,omitempty"`
}

// Conversation represents a conversation with messages.
type Conversation struct {
	ID         string                `json:"id"`
	Title      string                `json:"title"`
	CreatedAt  time.Time             `json:"createdAt"`
	ModifiedAt time.Time             `json:"modifiedAt"`
	Messages   []ConversationMessage `json:"messages"`
}
