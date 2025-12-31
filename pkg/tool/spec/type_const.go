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

	DefaultHTTPTimeoutMs = 10_000
	JSONEncoding         = "json"
	TextEncoding         = "text"
	DefaultHTTPEncoding  = JSONEncoding
	DefaultHTTPErrorMode = "fail"

	// SchemaVersion  - Current on-disk schema version.
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
	ErrBundleNotEmpty        = errors.New("bundle still contains tools")

	ErrToolNotFound = errors.New("tool not found")

	ErrBuiltInReadOnly = errors.New("built-in resource is read-only")

	ErrFTSDisabled  = errors.New("FTS is disabled")
	ErrToolDisabled = errors.New("tool is disabled")
)

type (
	ToolImplType  string
	JSONRawString = string
	JSONSchema    = json.RawMessage
)

const (
	ToolTypeGo   ToolImplType = "go"
	ToolTypeHTTP ToolImplType = "http"
)

// ToolImplOutputKind describes how a tool's result should be treated at the UX boundary.
type ToolImplOutputKind string

const (
	ToolOutputText ToolImplOutputKind = "text" // result is text or text-like JSON
	ToolOutputBlob ToolImplOutputKind = "blob" // bytes / file-like
	ToolOutputNone ToolImplOutputKind = "none" // no visible payload, side-effect only
)

// GoToolImpl - Register-by-name pattern for Go tools.
type GoToolImpl struct {
	// Fully-qualified registration key, e.g.
	//   "github.com/acme/flexigpt/tools.Weather"
	Func string `json:"func" validate:"required"`
}

// HTTPAuth - Simple auth descriptor (can be extended later).
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
	ErrorMode    string `json:"errorMode,omitempty"`    // "fail"(dflt) | "empty"
}

type HTTPToolImpl struct {
	Request  HTTPRequest  `json:"request"`
	Response HTTPResponse `json:"response"`
}

type ToolStoreChoiceType string

const (
	ToolTypeFunction  ToolStoreChoiceType = "function"
	ToolTypeCustom    ToolStoreChoiceType = "custom"
	ToolTypeWebSearch ToolStoreChoiceType = "webSearch"
)

type ToolStoreChoice struct {
	// BundleID, BundleSlug, ItemID, ItemSlug are string aliases.
	BundleID   bundleitemutils.BundleID   `json:"bundleID"`
	BundleSlug bundleitemutils.BundleSlug `json:"bundleSlug,omitempty"`

	ToolID      bundleitemutils.ItemID   `json:"toolID,omitempty"`
	ToolSlug    bundleitemutils.ItemSlug `json:"toolSlug"`
	ToolVersion string                   `json:"toolVersion"`

	ToolType    ToolStoreChoiceType `json:"toolType"`
	Description string              `json:"description,omitempty"`
	DisplayName string              `json:"displayName,omitempty"`
}

type Tool struct {
	SchemaVersion string                      `json:"schemaVersion"`
	ID            bundleitemutils.ItemID      `json:"id"` // UUID-v7
	Slug          bundleitemutils.ItemSlug    `json:"slug"`
	Version       bundleitemutils.ItemVersion `json:"version"` // opaque

	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	// UserCallable indicates whether the tool can be invoked directly by the user
	// (e.g. from the composer UI before sending a message).
	UserCallable bool `json:"userCallable"`
	// LLMCallable indicates whether the model may call this tool as a function.
	LLMCallable bool `json:"llmCallable"`
	// OutputKind describes how the tool output should be surfaced in the UX.
	OutputKind ToolImplOutputKind `json:"outputKind"`

	ArgSchema    JSONSchema `json:"argSchema"`    // validated pre-invoke
	OutputSchema JSONSchema `json:"outputSchema"` // validated post-invoke

	Type   ToolImplType  `json:"type"`
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
