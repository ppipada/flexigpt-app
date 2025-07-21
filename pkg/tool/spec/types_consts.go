package spec

import (
	"errors"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
)

const (
	ToolBundlesMetaFileName    = "tools.bundles.json"
	ToolDBFileName             = "tools.fts.sqlite"
	ToolBuiltInOverlayFileName = "toolsbuiltin.overlay.json"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidDir     = errors.New("invalid directory")
	ErrConflict       = errors.New("resource already exists")

	ErrBuiltInToolBundleNotFound = errors.New("tool bundle not found in built-in data")
	ErrToolBundleNotFound        = errors.New("tool bundle not found")
	ErrToolBundleDisabled        = errors.New("tool bundle is disabled")
	ErrToolBundleDeleting        = errors.New("tool bundle is being deleted")
	ErrToolBundleNotEmpty        = errors.New("tool bundle still contains tools")

	ErrToolNotFound        = errors.New("tool not found")
	ErrBuiltInToolNotFound = errors.New("tool not found in built-in data")

	ErrBuiltInReadOnly = errors.New("built-in resource is read-only")

	ErrFTSDisabled = errors.New("FTS is disabled")
)

// Primitive parameter kinds.
type ParamType string

const (
	ParamString  ParamType = "string"
	ParamNumber  ParamType = "number"
	ParamBoolean ParamType = "boolean"
	ParamEnum    ParamType = "enum"
)

// One parameter accepted by a tool.
type ToolParameter struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Description string    `json:"description,omitempty"`
	Required    bool      `json:"required"`
	EnumValues  []string  `json:"enumValues,omitempty"`
}

// One callable function.
type ToolSpec struct {
	ID          bundleitemutils.ItemID   `json:"id"`
	DisplayName string                   `json:"displayName"`
	Slug        bundleitemutils.ItemSlug `json:"slug"`
	IsEnabled   bool                     `json:"isEnabled"`
	Description string                   `json:"description,omitempty"`
	Tags        []string                 `json:"tags,omitempty"`
	Parameters  []ToolParameter          `json:"parameters,omitempty"`

	Version    bundleitemutils.ItemVersion `json:"version"`
	CreatedAt  time.Time                   `json:"createdAt"`
	ModifiedAt time.Time                   `json:"modifiedAt"`
	IsBuiltIn  bool                        `json:"isBuiltIn"`
}

// Hard grouping and distribution unit.
type ToolBundle struct {
	ID            bundleitemutils.BundleID   `json:"id"`
	Slug          bundleitemutils.BundleSlug `json:"slug"`
	DisplayName   string                     `json:"displayName,omitempty"`
	Description   string                     `json:"description,omitempty"`
	IsEnabled     bool                       `json:"isEnabled"`
	CreatedAt     time.Time                  `json:"createdAt"`
	ModifiedAt    time.Time                  `json:"modifiedAt"`
	IsBuiltIn     bool                       `json:"isBuiltIn"`
	SoftDeletedAt *time.Time                 `json:"softDeletedAt,omitempty"`
}

type AllToolBundles struct {
	ToolBundles map[bundleitemutils.BundleID]ToolBundle `json:"toolBundles"`
}
