package spec

import (
	"errors"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
)

const (
	PromptBundlesMetaFileName      = "prompts.bundles.json"
	PromptDBFileName               = "prompts.fts.sqlite"
	PromptBuiltInOverlayDBFileName = "promptsbuiltin.overlay.sqlite"

	// SchemaVersion is the current on-disk schema version.
	SchemaVersion = "2025-07-01"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidDir     = errors.New("invalid directory")
	ErrConflict       = errors.New("resource already exists")

	ErrBuiltInBundleNotFound = errors.New("bundle not found in built-in data")
	ErrBundleNotFound        = errors.New("bundle not found")
	ErrBundleDisabled        = errors.New("bundle is disabled")
	ErrBundleDeleting        = errors.New("bundle is being deleted")
	ErrBundleNotEmpty        = errors.New("bundle still contains templates")
	ErrTemplateNotFound      = errors.New("template not found")

	ErrBuiltInReadOnly = errors.New("built-in resource is read-only")

	ErrFTSDisabled = errors.New("FTS is disabled")
)

type (
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

type VarType string

const (
	VarString  VarType = "string"
	VarNumber  VarType = "number"
	VarBoolean VarType = "boolean"
	VarEnum    VarType = "enum"
	VarDate    VarType = "date"
)

type VarSource string

const (
	// SourceUser: Ask UI / CLI.
	SourceUser VarSource = "user"
	// SourceStatic: Fixed literal.
	SourceStatic VarSource = "static"
	// SourceTool: Filled by a helper tool.
	SourceTool VarSource = "tool"
)

// MessageBlock - One role-tagged chunk of text.
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
	ToolBundleID bundleitemutils.BundleID    `json:"toolBundleID,omitempty"`
	ToolSlug     bundleitemutils.ItemSlug    `json:"toolSlug,omitempty"`
	ToolVersion  bundleitemutils.ItemVersion `json:"toolVersion,omitempty"`
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

// PreProcessorCall - Runs a helper tool, optionally extracts a JSON sub-path and stores the value into a variable.
type PreProcessorCall struct {
	ID           PreProcessorCallID          `json:"id"`
	ToolBundleID bundleitemutils.BundleID    `json:"toolBundleID"`
	ToolSlug     bundleitemutils.ItemSlug    `json:"toolSlug"`
	ToolVersion  bundleitemutils.ItemVersion `json:"toolVersion"`
	Arguments    map[string]any              `json:"args,omitempty"`

	// Variable name.
	SaveAs string `json:"saveAs"`
	// E.g. "$.weather.tempC".
	PathExpr string `json:"pathExpr,omitempty"`
	// Default empty.
	OnError PreProcessorOnError `json:"onError,omitempty"`
}

type PromptTemplate struct {
	SchemaVersion string                   `json:"schemaVersion"`
	ID            bundleitemutils.ItemID   `json:"id"`
	Slug          bundleitemutils.ItemSlug `json:"slug"`
	IsEnabled     bool                     `json:"isEnabled"`

	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	// Ordered list of blocks that form the final prompt.
	Blocks []MessageBlock `json:"blocks"`
	// Declared placeholders.
	Variables []PromptVariable `json:"variables,omitempty"`
	// Helper steps executed before the prompt is sent.
	PreProcessors []PreProcessorCall `json:"preProcessors,omitempty"`

	Version    bundleitemutils.ItemVersion `json:"version"`
	CreatedAt  time.Time                   `json:"createdAt"`
	ModifiedAt time.Time                   `json:"modifiedAt"`
	IsBuiltIn  bool                        `json:"isBuiltIn"`
}

// PromptBundle is a hard grouping & distribution unit.
type PromptBundle struct {
	SchemaVersion string                     `json:"schemaVersion"`
	ID            bundleitemutils.BundleID   `json:"id"`
	Slug          bundleitemutils.BundleSlug `json:"slug"`

	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`

	IsEnabled     bool       `json:"isEnabled"`
	CreatedAt     time.Time  `json:"createdAt"`
	ModifiedAt    time.Time  `json:"modifiedAt"`
	IsBuiltIn     bool       `json:"isBuiltIn"`
	SoftDeletedAt *time.Time `json:"softDeletedAt,omitempty"`
}

type AllBundles struct {
	Bundles map[bundleitemutils.BundleID]PromptBundle `json:"bundles"`
}
