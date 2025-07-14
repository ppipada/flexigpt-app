package spec

import (
	"time"
)

type (
	BundleID           string
	BundleSlug         string
	TemplateID         string
	TemplateSlug       string
	TemplateVersion    string
	MessageBlockID     string
	PreProcessorCallID string
)

type PromptRoleEnum string

const (
	System    PromptRoleEnum = "system"
	Developer PromptRoleEnum = "developer"
	User      PromptRoleEnum = "user"
	Assistant PromptRoleEnum = "assistant"
)

// Primitive variable kinds.
type VarType string

const (
	VarString  VarType = "string"
	VarNumber  VarType = "number"
	VarBoolean VarType = "boolean"
	VarEnum    VarType = "enum"
	VarDate    VarType = "date"
)

// Where a variable’s value comes from.
type VarSource string

const (
	// Ask UI / CLI.
	SourceUser VarSource = "user"
	// Fixed literal.
	SourceStatic VarSource = "static"
	// Filled by a helper tool.
	SourceTool VarSource = "tool"
)

// One role-tagged chunk of text.
type MessageBlock struct {
	ID      MessageBlockID `json:"id"`
	Role    PromptRoleEnum `json:"role"`
	Content string         `json:"content"`
}

type PromptVariable struct {
	Name        string    `json:"name"`
	Type        VarType   `json:"type"`
	Required    bool      `json:"required"`
	Source      VarSource `json:"source"`
	Description string    `json:"description,omitempty"`

	// SourceStatic.
	StaticVal string `json:"staticVal,omitempty"`
	// SourceTool.
	ToolID string `json:"toolId,omitempty"`
	// VarEnum.
	EnumValues []string `json:"enumValues,omitempty"`

	// Optional default for the var.
	Default string `json:"default,omitempty"`
}

type PreProcessorOnError string

const (
	OnErrorEmpty PreProcessorOnError = "empty"
	OnErrorFail  PreProcessorOnError = "fail"
)

// Runs a helper tool, optionally extracts a JSON sub-path and stores the value into a variable.
type PreProcessorCall struct {
	ID        PreProcessorCallID `json:"id"`
	ToolID    string             `json:"toolId"`
	Arguments map[string]any     `json:"args,omitempty"`

	// Variable name.
	SaveAs string `json:"saveAs"`
	// E.g. "$.weather.tempC".
	PathExpr string `json:"pathExpr,omitempty"`
	// Default empty.
	OnError PreProcessorOnError `json:"onError,omitempty"`
}

type PromptTemplate struct {
	ID          TemplateID `json:"id"`
	DisplayName string     `json:"displayName"`
	// Unique within a bundle.
	Slug        TemplateSlug `json:"slug"`
	IsEnabled   bool         `json:"isEnabled"`
	Description string       `json:"description,omitempty"`
	Tags        []string     `json:"tags,omitempty"`

	// Ordered list of blocks that form the final prompt.
	Blocks []MessageBlock `json:"blocks"`
	// Declared placeholders.
	Variables []PromptVariable `json:"variables,omitempty"`
	// Helper steps executed before the prompt is sent.
	PreProcessors []PreProcessorCall `json:"preProcessors,omitempty"`

	Version    TemplateVersion `json:"version"`
	CreatedAt  time.Time       `json:"createdAt"`
	ModifiedAt time.Time       `json:"modifiedAt"`
	IsBuiltIn  bool            `json:"isBuiltIn"`
}

// Hard grouping & distribution unit.
type PromptBundle struct {
	ID            BundleID   `json:"id"`
	Slug          BundleSlug `json:"slug"`
	DisplayName   string     `json:"displayName,omitempty"`
	Description   string     `json:"description,omitempty"`
	IsEnabled     bool       `json:"isEnabled"`
	CreatedAt     time.Time  `json:"createdAt"`
	ModifiedAt    time.Time  `json:"modifiedAt"`
	IsBuiltIn     bool       `json:"isBuiltIn"`
	SoftDeletedAt *time.Time `json:"softDeletedAt,omitempty"`
}

type AllBundles struct {
	Bundles map[BundleID]PromptBundle `json:"bundles"`
}
