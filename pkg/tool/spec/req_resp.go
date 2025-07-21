package spec

import "github.com/ppipada/flexigpt-app/pkg/bundleitemutils"

type PutToolBundleRequest struct {
	BundleID bundleitemutils.BundleID  `path:"bundleID" required:"true"`
	Body     *PutToolBundleRequestBody `                                json:"body"`
}

type PutToolBundleRequestBody struct {
	Slug        bundleitemutils.BundleSlug `json:"slug"                  required:"true"`
	DisplayName string                     `json:"displayName"           required:"true"`
	IsEnabled   bool                       `json:"isEnabled"             required:"true"`
	Description string                     `json:"description,omitempty"`
}

type PutToolBundleResponse struct{}

type DeleteToolBundleRequest struct {
	BundleID bundleitemutils.BundleID `path:"bundleID"`
}
type DeleteToolBundleResponse struct{}

type PatchToolBundleRequest struct {
	BundleID bundleitemutils.BundleID    `path:"bundleID" required:"true"`
	Body     *PatchToolBundleRequestBody `                                json:"body"`
}

type PatchToolBundleRequestBody struct {
	IsEnabled bool `json:"isEnabled" required:"true"`
}

type PatchToolBundleResponse struct{}

type ListToolBundlesRequest struct {
	BundleIDs       []bundleitemutils.BundleID `query:"bundleIDs"`
	IncludeDisabled bool                       `query:"includeDisabled"`
	PageSize        int                        `query:"pageSize"`
	PageToken       string                     `query:"pageToken"`
}

type ListToolBundlesResponse struct {
	Body *ListToolBundlesResponseBody
}

type ListToolBundlesResponseBody struct {
	ToolBundles   []ToolBundle `json:"toolBundles"`
	NextPageToken *string      `json:"nextPageToken,omitempty"`
}

type PutToolRequest struct {
	BundleID bundleitemutils.BundleID `path:"bundleID" required:"true"`
	ToolSlug bundleitemutils.ItemSlug `path:"toolSlug" required:"true"`
	Body     *PutToolRequestBody      `                                json:"body"`
}

type PutToolRequestBody struct {
	DisplayName string                      `json:"displayName"           required:"true"`
	IsEnabled   bool                        `json:"isEnabled"             required:"true"`
	Version     bundleitemutils.ItemVersion `json:"version"               required:"true"`
	Description string                      `json:"description,omitempty"`
	Tags        []string                    `json:"tags,omitempty"`
	Parameters  []ToolParameter             `json:"parameters,omitempty"`
}

type PutToolResponse struct{}

type DeleteToolRequest struct {
	BundleID bundleitemutils.BundleID    `path:"bundleID" required:"true"`
	ToolSlug bundleitemutils.ItemSlug    `path:"toolSlug" required:"true"`
	Version  bundleitemutils.ItemVersion `                required:"true" query:"version"`
}
type DeleteToolResponse struct{}

type PatchToolRequest struct {
	BundleID bundleitemutils.BundleID `path:"bundleID" required:"true"`
	ToolSlug bundleitemutils.ItemSlug `path:"toolSlug" required:"true"`
	Body     *PatchToolRequestBody    `                                json:"body"`
}

type PatchToolRequestBody struct {
	Version   bundleitemutils.ItemVersion `json:"version"   required:"true"`
	IsEnabled bool                        `json:"isEnabled" required:"true"`
}

type PatchToolResponse struct{}

type GetToolRequest struct {
	BundleID bundleitemutils.BundleID    `path:"bundleID" required:"true"`
	ToolSlug bundleitemutils.ItemSlug    `path:"toolSlug" required:"true"`
	Version  bundleitemutils.ItemVersion `                required:"true" query:"version"`
}
type GetToolResponse struct{ Body *ToolSpec }

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

type ListToolsResponse struct {
	Body *ListToolsResponseBody
}

type ListToolsResponseBody struct {
	ToolListItems []ToolListItem `json:"toolListItems"`
	NextPageToken *string        `json:"nextPageToken,omitempty"`
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
