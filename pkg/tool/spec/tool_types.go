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
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	InvokeSlug  string `json:"invokeSlug,omitempty"`
	IsEnabled   bool   `json:"isEnabled"`

	Parameters []ToolParameter `json:"parameters"`

	Version    string    `json:"version"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

// Collection of tools that often travel together (e.g. "Web + Math").
type ToolBundle struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	InvokeSlug  string `json:"invokeSlug,omitempty"`
	IsEnabled   bool   `json:"isEnabled"`

	Tools []ToolSpec `json:"tools"`

	Version    string    `json:"version"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}
