package spec

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
)

const (
	ToolBundlesMetaFileName      = "tools.bundles.json"
	ToolDBFileName               = "tools.fts.sqlite"
	ToolBuiltInOverlayDBFileName = "toolsbuiltin.overlay.sqlite"

	// Current on-disk schema version.
	SchemaVersion        = "2025-07-01"
	DefaultHTTPTimeoutMs = 10_000
	DefaultHTTPEncoding  = "json"
	DefaultHTTPErrorMode = "fail"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidDir     = errors.New("invalid directory")
	ErrConflict       = errors.New("resource already exists")

	ErrBuiltInBundleNotFound = errors.New("bundle not found in built-in data")
	ErrBundleNotFound        = errors.New("bundle not found")
	ErrBundleDisabled        = errors.New("bundle is disabled")
	ErrBundleDeleting        = errors.New("bundle is being deleted")
	ErrBundleNotEmpty        = errors.New("bundle still contains tools")

	ErrToolNotFound = errors.New("tool not found")

	ErrBuiltInReadOnly = errors.New("built-in resource is read-only")

	ErrFTSDisabled = errors.New("FTS is disabled")
)

type (
	ToolType   string
	JSONSchema = json.RawMessage
)

const (
	ToolTypeGo   ToolType = "go"
	ToolTypeHTTP ToolType = "http"
)

// Register-by-name pattern for Go tools.
type GoToolImpl struct {
	// Fully-qualified registration key, e.g.
	//   "github.com/acme/flexigpt/tools.Weather"
	Func string `json:"func" validate:"required"`
}

// Simple auth descriptor (can be extended later).
type HTTPAuth struct {
	Type          string `json:"type"`
	In            string `json:"in,omitempty"`   // "header" | "query"  (apiKey only)
	Name          string `json:"name,omitempty"` // header/query key
	ValueTemplate string `json:"valueTemplate"`  // may contain ${SECRET}
}
type HTTPRequest struct {
	Method      string            `json:"method,omitempty"`    // default "GET"
	URLTemplate string            `json:"urlTemplate"`         // http(s)://… may contain ${var}
	Query       map[string]string `json:"query,omitempty"`     // k:${var}
	Headers     map[string]string `json:"headers,omitempty"`   // k:${var}
	Body        string            `json:"body,omitempty"`      // raw or template
	Auth        *HTTPAuth         `json:"auth,omitempty"`      // see below
	TimeoutMs   int               `json:"timeoutMs,omitempty"` // default 10 000
}

type HTTPResponse struct {
	SuccessCodes []int  `json:"successCodes,omitempty"` // default: any 2xx
	Encoding     string `json:"encoding,omitempty"`     // "json"(dflt) | "text"
	Selector     string `json:"selector,omitempty"`     // JSONPath / JMESPath / regexp
	ErrorMode    string `json:"errorMode,omitempty"`    // "fail"(dflt) | "empty"
}

type HTTPToolImpl struct {
	Request  HTTPRequest  `json:"request"`
	Response HTTPResponse `json:"response"`
}

type Tool struct {
	SchemaVersion string                      `json:"schemaVersion"`
	ID            bundleitemutils.ItemID      `json:"id"` // UUID-v7
	Slug          bundleitemutils.ItemSlug    `json:"slug"`
	Version       bundleitemutils.ItemVersion `json:"version"` // opaque

	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	ArgSchema    JSONSchema `json:"argSchema"`    // validated pre-invoke
	OutputSchema JSONSchema `json:"outputSchema"` // validated post-invoke

	Type   ToolType      `json:"type"`
	GoImpl *GoToolImpl   `json:"goImpl,omitempty"`
	HTTP   *HTTPToolImpl `json:"httpImpl,omitempty"`

	IsEnabled  bool      `json:"isEnabled"`
	IsBuiltIn  bool      `json:"isBuiltIn"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type ToolBundle struct {
	SchemaVersion string                     `json:"schemaVersion"`
	ID            bundleitemutils.BundleID   `json:"id"` // UUID-v7
	Slug          bundleitemutils.BundleSlug `json:"slug"`

	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`

	IsEnabled     bool       `json:"isEnabled"`
	IsBuiltIn     bool       `json:"isBuiltIn"`
	CreatedAt     time.Time  `json:"createdAt"`
	ModifiedAt    time.Time  `json:"modifiedAt"`
	SoftDeletedAt *time.Time `json:"softDeletedAt,omitempty"`
}

type AllBundles struct {
	Bundles map[bundleitemutils.BundleID]ToolBundle `json:"bundles"`
}
