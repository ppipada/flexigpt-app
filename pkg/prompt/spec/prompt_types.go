package spec

import (
	"time"
)

type PromptRoleEnum string

const (
	System    PromptRoleEnum = "system"
	Developer PromptRoleEnum = "developer"
	User      PromptRoleEnum = "user"
	Assistant PromptRoleEnum = "assistant"
	Function  PromptRoleEnum = "function"
)

// Variable types you care about.
type VarType string

const (
	VarString  VarType = "string"
	VarNumber  VarType = "number"
	VarBoolean VarType = "boolean"
	VarEnum    VarType = "enum"
	VarDate    VarType = "date"
)

// How a variable is populated.
type VarSource string

const (
	// Ask the user or parse CLI flags.
	SourceUser VarSource = "user"
	// Fixed value in template.
	SourceStatic VarSource = "static"
	// Run helper tool.
	SourceTool VarSource = "tool"
)

// A single role-tagged chunk of text.
type MessageBlock struct {
	ID      string         `json:"id"`
	Role    PromptRoleEnum `json:"role"`
	Content string         `json:"content"`
	Enabled bool           `json:"enabled,omitempty"`
}

type PromptVariable struct {
	Name        string    `json:"name"`
	Type        VarType   `json:"type"`
	Source      VarSource `json:"source"`
	Description string    `json:"description,omitempty"`

	// For SourceStatic.
	StaticVal string `json:"staticVal,omitempty"`
	// For SourceTool.
	ToolID string `json:"toolID,omitempty"`
	// For VarEnum.
	EnumValues []string `json:"enumValues,omitempty"`

	Required bool `json:"required"`
}

type PromptTemplate struct {
	ID          string   `json:"id"`
	Version     string   `json:"version"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`

	// Ordered list of blocks.
	Blocks []MessageBlock `json:"blocks"`

	// Meta about the placeholders used in Blocks.
	Variables []PromptVariable `json:"variables,omitempty"`

	// Zero-or-more bundles that will be exposed to the models at runtime.
	ToolBundleIDs []string `json:"toolBundleIds,omitempty"`
}
