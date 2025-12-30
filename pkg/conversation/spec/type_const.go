package spec

import (
	"time"

	inferencegoSpec "github.com/ppipada/inference-go/spec"
)

const (
	ConversationFileExtension = "json"
	MaxPageSize               = 256
	DefaultPageSize           = 12
	ConversationSchemaVersion = "v1.0.0"
)

// ConversationMessage represents a single *turn* in the conversation.
//
// Examples:
//   - User turn: text + attachments + per-turn tool choices.
//   - Assistant turn: one or more messages, tool calls, tool outputs, reasoning, usage.
type ConversationMessage struct {
	ID        string                   `json:"id"`
	CreatedAt time.Time                `json:"createdAt"`
	Role      inferencegoSpec.RoleEnum `json:"role"`
	Status    inferencegoSpec.Status   `json:"status"`

	// Default model configuration for this turn. This can be empty and would mean that model param have been carried
	// over from previous messages.
	ModelParam *inferencegoSpec.ModelParam `json:"modelParam,omitempty"`

	// Primary content for this turn (text, images, files, etc.).
	// Usually 1 element for user or assistant, but we allow multiple
	// to preserve fidelity if the backend ever returns more.
	Messages []*inferencegoSpec.InputOutputContent `json:"messages,omitempty"`

	// Reasoning chunks associated with this turn (usually assistant).
	Reasoning []*inferencegoSpec.ReasoningContent `json:"reasoning,omitempty"`

	// Tool interactions that conceptually belong to this turn.
	ToolCalls   []inferencegoSpec.ToolCall   `json:"toolCalls,omitempty"`
	ToolOutputs []inferencegoSpec.ToolOutput `json:"toolOutputs,omitempty"`

	// Per-turn tool choices. These are tools you want available for the
	// *next* completion run that this turn initiates. You'll typically
	// merge these with Conversation.ToolChoices when preparing a request.
	ToolChoices []inferencegoSpec.ToolChoice `json:"toolChoices,omitempty"`

	// Usage / error info from the model/provider for this turn
	// (usually attached to assistant turns).
	Usage *inferencegoSpec.Usage `json:"usage,omitempty"`
	Error *inferencegoSpec.Error `json:"error,omitempty"`

	// Arbitrary UI/app metadata (tags, pinned, read state, etc.).
	Meta map[string]any `json:"meta,omitempty"`
}

// Conversation is the full chat, stored as a single JSON file.
type Conversation struct {
	SchemaVersion string    `json:"schemaVersion"`
	ID            string    `json:"id"`
	Title         string    `json:"title,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	ModifiedAt    time.Time `json:"modifiedAt"`

	// Ordered list of turns (messages) in the transcript.
	Messages []ConversationMessage `json:"messages"`

	// Extra metadata for your app (folders, tags, project, etc.).
	Meta map[string]any `json:"meta,omitempty"`
}
