package spec

import "time"

// Simplified JSON-schema node for a single function parameter.
type ToolParameter struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required"`
	EnumValues  []string `json:"enumValues,omitempty"`
}

// One callable function exposed to the LLM.
type ToolSpec struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	InvokeSlugs []string        `json:"invokeSlugs,omitempty"`
	Description string          `json:"description"`
	Parameters  []ToolParameter `json:"parameters"`
	// Future-proofing hooks
	// Run via sandbox?
	SafeMode bool `json:"safeMode,omitempty"`
	// Free-form tag bag.
	Meta       map[string]string `json:"meta,omitempty"`
	Version    int               `json:"version"`
	CreatedAt  time.Time         `json:"createdAt"`
	ModifiedAt time.Time         `json:"modifiedAt"`
}

// Collection of tools that often travel together (e.g. “Web + Math”).
type ToolBundle struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Tools       []ToolSpec `json:"tools"`

	Version    int       `json:"version"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}
