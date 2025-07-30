package spec

import (
	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
)

type PutToolBundleRequestBody struct {
	Slug        bundleitemutils.BundleSlug `json:"slug"                  required:"true"`
	DisplayName string                     `json:"displayName"           required:"true"`
	IsEnabled   bool                       `json:"isEnabled"             required:"true"`
	Description string                     `json:"description,omitempty"`
}

type PutToolBundleRequest struct {
	BundleID bundleitemutils.BundleID  `path:"bundleID" required:"true"`
	Body     *PutToolBundleRequestBody `                                json:"body"`
}

type PutToolBundleResponse struct{}

type DeleteToolBundleRequest struct {
	BundleID bundleitemutils.BundleID `path:"bundleID" required:"true"`
}
type DeleteToolBundleResponse struct{}

type PatchToolBundleRequestBody struct {
	IsEnabled bool `json:"isEnabled" required:"true"`
}

type PatchToolBundleRequest struct {
	BundleID bundleitemutils.BundleID    `path:"bundleID" required:"true"`
	Body     *PatchToolBundleRequestBody `                                json:"body"`
}

type PatchToolBundleResponse struct{}

// BundlePageToken is the opaque cursor used when paging over tool-bundles.
type BundlePageToken struct {
	BundleIDs       []bundleitemutils.BundleID `json:"ids,omitempty"` // Optional bundle-ID filter.
	IncludeDisabled bool                       `json:"d,omitempty"`   // Include disabled bundles?
	PageSize        int                        `json:"s"`             // Requested page-size.
	CursorMod       string                     `json:"t,omitempty"`   // RFC-3339-nano modification timestamp.
	CursorID        bundleitemutils.BundleID   `json:"id,omitempty"`  // Tie-breaker for equal timestamps.
}

type ListToolBundlesRequest struct {
	BundleIDs       []bundleitemutils.BundleID `query:"bundleIDs"`
	IncludeDisabled bool                       `query:"includeDisabled"`
	PageSize        int                        `query:"pageSize"`
	PageToken       string                     `query:"pageToken"`
}

type ListToolBundlesResponseBody struct {
	ToolBundles   []ToolBundle `json:"toolBundles"`
	NextPageToken *string      `json:"nextPageToken,omitempty"`
}

type ListToolBundlesResponse struct {
	Body *ListToolBundlesResponseBody
}

type PutToolRequestBody struct {
	DisplayName string   `json:"displayName"           required:"true"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsEnabled   bool     `json:"isEnabled"             required:"true"`

	ArgSchema    JSONSchema `json:"argSchema"    required:"true"`
	OutputSchema JSONSchema `json:"outputSchema" required:"true"`

	Type   ToolType      `json:"type"               required:"true"`
	GoImpl *GoToolImpl   `json:"goImpl,omitempty"`
	HTTP   *HTTPToolImpl `json:"httpImpl,omitempty"`
}

type PutToolRequest struct {
	BundleID bundleitemutils.BundleID    `path:"bundleID" required:"true"`
	ToolSlug bundleitemutils.ItemSlug    `path:"toolSlug" required:"true"`
	Version  bundleitemutils.ItemVersion `path:"version"  required:"true"`
	Body     *PutToolRequestBody         `                                json:"body"`
}
type PutToolResponse struct{}

type DeleteToolRequest struct {
	BundleID bundleitemutils.BundleID    `path:"bundleID" required:"true"`
	ToolSlug bundleitemutils.ItemSlug    `path:"toolSlug" required:"true"`
	Version  bundleitemutils.ItemVersion `path:"version"  required:"true"`
}
type DeleteToolResponse struct{}

type PatchToolRequestBody struct {
	IsEnabled bool `json:"isEnabled" required:"true"`
}

type PatchToolRequest struct {
	BundleID bundleitemutils.BundleID    `path:"bundleID" required:"true"`
	ToolSlug bundleitemutils.ItemSlug    `path:"toolSlug" required:"true"`
	Version  bundleitemutils.ItemVersion `path:"version"  required:"true"`

	Body *PatchToolRequestBody `json:"body"`
}

type PatchToolResponse struct{}

type GetToolRequest struct {
	BundleID bundleitemutils.BundleID    `path:"bundleID" required:"true"`
	ToolSlug bundleitemutils.ItemSlug    `path:"toolSlug" required:"true"`
	Version  bundleitemutils.ItemVersion `path:"version"  required:"true"`
}
type GetToolResponse struct{ Body *Tool }

// ToolPageToken is the opaque cursor used when paging over individual tool
// versions (ListTools API).
type ToolPageToken struct {
	RecommendedPageSize int                        `json:"ps,omitempty"`
	IncludeDisabled     bool                       `json:"d,omitempty"`
	BundleIDs           []bundleitemutils.BundleID `json:"ids,omitempty"`
	Tags                []string                   `json:"tags,omitempty"`
	BuiltInDone         bool                       `json:"bd,omitempty"` // Built-ins already emitted?
	DirTok              string                     `json:"dt,omitempty"` // Directory-store cursor.
}

type ListToolsRequest struct {
	BundleIDs           []bundleitemutils.BundleID `query:"bundleIDs"`
	Tags                []string                   `query:"tags"`
	IncludeDisabled     bool                       `query:"includeDisabled"`
	RecommendedPageSize int                        `query:"recommendedPageSize"`
	PageToken           string                     `query:"pageToken"`
}

type ToolListItem struct {
	BundleID    bundleitemutils.BundleID    `json:"bundleID"`
	BundleSlug  bundleitemutils.BundleSlug  `json:"bundleSlug"`
	ToolSlug    bundleitemutils.ItemSlug    `json:"toolSlug"`
	ToolVersion bundleitemutils.ItemVersion `json:"toolVersion"`
	IsBuiltIn   bool                        `json:"isBuiltIn"`
}

type ListToolsResponseBody struct {
	ToolListItems []ToolListItem `json:"toolListItems"`
	NextPageToken *string        `json:"nextPageToken,omitempty"`
}
type ListToolsResponse struct {
	Body *ListToolsResponseBody
}

type SearchToolsRequest struct {
	Query           string `query:"q"               required:"true"`
	PageToken       string `query:"pageToken"`
	PageSize        int    `query:"pageSize"`
	IncludeDisabled bool   `query:"includeDisabled"`
}

type SearchToolsResponseBody struct {
	ToolListItems []ToolListItem `json:"toolListItems"`
	NextPageToken *string        `json:"nextPageToken,omitempty"`
}
type SearchToolsResponse struct {
	Body *SearchToolsResponseBody
}
