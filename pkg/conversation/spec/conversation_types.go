package spec

import (
	"time"
)

// ConversationRoleEnum represents the role of a participant in a conversation.
type ConversationRoleEnum string

const (
	System    ConversationRoleEnum = "system"
	User      ConversationRoleEnum = "user"
	Assistant ConversationRoleEnum = "assistant"
	Function  ConversationRoleEnum = "function"
	Feedback  ConversationRoleEnum = "feedback"
)

// ConversationMessage represents a message in a conversation.
type ConversationMessage struct {
	ID        string               `json:"id"`
	CreatedAt *time.Time           `json:"createdAt,omitempty"`
	Role      ConversationRoleEnum `json:"role"`
	Content   string               `json:"content"`
	Timestamp *string              `json:"timestamp,omitempty"`
	Name      *string              `json:"name,omitempty"`
	Details   *string              `json:"details,omitempty"`
}

// ConversationItem represents a conversation with basic details.
type ConversationItem struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

// Conversation represents a conversation with messages.
type Conversation struct {
	ConversationItem
	ModifiedAt time.Time             `json:"modifiedAt"`
	Messages   []ConversationMessage `json:"messages"`
}

// IConversationAPI defines the interface for conversation-related operations.
type IConversationAPI interface {
	SaveConversation(conversation Conversation) error
	DeleteConversation(id string, title string) error
	GetConversation(id string, title string) (*Conversation, error)
	ListConversations(token *string) ([]ConversationItem, *string, error)
	AddMessageToConversation(id string, title string, newMessage ConversationMessage) error
}
