package spec

import "github.com/ppipada/flexigpt-app/pkg/bundleitemutils"

type PutPromptBundleRequest struct {
	BundleID bundleitemutils.BundleID    `path:"bundleID" required:"true"`
	Body     *PutPromptBundleRequestBody `                                json:"body"`
}

type PutPromptBundleRequestBody struct {
	Slug        bundleitemutils.BundleSlug `json:"slug"                  required:"true"`
	DisplayName string                     `json:"displayName"           required:"true"`
	IsEnabled   bool                       `json:"isEnabled"             required:"true"`
	Description string                     `json:"description,omitempty"`
}

type PutPromptBundleResponse struct{}

type DeletePromptBundleRequest struct {
	BundleID bundleitemutils.BundleID `path:"bundleID"`
}
type DeletePromptBundleResponse struct{}

type PatchPromptBundleRequest struct {
	BundleID bundleitemutils.BundleID      `path:"bundleID" required:"true"`
	Body     *PatchPromptBundleRequestBody `                                json:"body"`
}

type PatchPromptBundleRequestBody struct {
	IsEnabled bool `json:"isEnabled" required:"true"`
}

type PatchPromptBundleResponse struct{}

type ListPromptBundlesRequest struct {
	BundleIDs       []bundleitemutils.BundleID `query:"bundleIDs"`
	IncludeDisabled bool                       `query:"includeDisabled"`
	PageSize        int                        `query:"pageSize"`
	PageToken       string                     `query:"pageToken"`
}

type ListPromptBundlesResponse struct {
	Body *ListPromptBundlesResponseBody
}

type ListPromptBundlesResponseBody struct {
	PromptBundles []PromptBundle `json:"promptBundles"`
	NextPageToken *string        `json:"nextPageToken,omitempty"`
}

type PutPromptTemplateRequest struct {
	BundleID     bundleitemutils.BundleID      `path:"bundleID"     required:"true"`
	TemplateSlug bundleitemutils.ItemSlug      `path:"templateSlug" required:"true"`
	Version      bundleitemutils.ItemVersion   `path:"version"      required:"true"`
	Body         *PutPromptTemplateRequestBody `                                    json:"body"`
}

type PutPromptTemplateRequestBody struct {
	// Auto populate template id internally.
	DisplayName string `json:"displayName"           required:"true"`
	IsEnabled   bool   `json:"isEnabled"             required:"true"`
	Description string `json:"description,omitempty"`

	Blocks        []MessageBlock     `json:"blocks"                  required:"true"`
	Tags          []string           `json:"tags,omitempty"`
	Variables     []PromptVariable   `json:"variables,omitempty"`
	PreProcessors []PreProcessorCall `json:"preProcessors,omitempty"`
}

type PutPromptTemplateResponse struct{}

type DeletePromptTemplateRequest struct {
	BundleID     bundleitemutils.BundleID    `path:"bundleID"     required:"true"`
	TemplateSlug bundleitemutils.ItemSlug    `path:"templateSlug" required:"true"`
	Version      bundleitemutils.ItemVersion `path:"version"      required:"true"`
}
type DeletePromptTemplateResponse struct{}

type PatchPromptTemplateRequest struct {
	BundleID     bundleitemutils.BundleID        `path:"bundleID"     required:"true"`
	TemplateSlug bundleitemutils.ItemSlug        `path:"templateSlug" required:"true"`
	Version      bundleitemutils.ItemVersion     `path:"version"      required:"true"`
	Body         *PatchPromptTemplateRequestBody `                                    json:"body"`
}

type PatchPromptTemplateRequestBody struct {
	IsEnabled bool `json:"isEnabled" required:"true"`
}

type PatchPromptTemplateResponse struct{}

type GetPromptTemplateRequest struct {
	BundleID     bundleitemutils.BundleID    `path:"bundleID"     required:"true"`
	TemplateSlug bundleitemutils.ItemSlug    `path:"templateSlug" required:"true"`
	Version      bundleitemutils.ItemVersion `path:"version"      required:"true"`
}
type GetPromptTemplateResponse struct{ Body *PromptTemplate }

type ListPromptTemplatesRequest struct {
	BundleIDs           []bundleitemutils.BundleID `query:"bundleIDs"`
	Tags                []string                   `query:"tags"`
	IncludeDisabled     bool                       `query:"includeDisabled"`
	RecommendedPageSize int                        `query:"recommendedPageSize"`
	PageToken           string                     `query:"pageToken"`
}

type ListPromptTemplatesResponse struct {
	Body *ListPromptTemplatesResponseBody
}

type PromptTemplateListItem struct {
	BundleID        bundleitemutils.BundleID    `json:"bundleID"`
	BundleSlug      bundleitemutils.BundleSlug  `json:"bundleSlug"`
	TemplateSlug    bundleitemutils.ItemSlug    `json:"templateSlug"`
	TemplateVersion bundleitemutils.ItemVersion `json:"templateVersion"`
	IsBuiltIn       bool                        `json:"isBuiltIn"`
}

type ListPromptTemplatesResponseBody struct {
	PromptTemplateListItems []PromptTemplateListItem `json:"promptTemplateListItems"`
	NextPageToken           *string                  `json:"nextPageToken,omitempty"`
}

type SearchPromptTemplatesRequest struct {
	Query           string `query:"q"               required:"true"`
	PageToken       string `query:"pageToken"`
	PageSize        int    `query:"pageSize"`
	IncludeDisabled bool   `query:"includeDisabled"`
}

type SearchPromptTemplatesResponseBody struct {
	PromptTemplateListItems []PromptTemplateListItem `json:"promptTemplateListItems"`
	NextPageToken           *string                  `json:"nextPageToken,omitempty"`
}

type SearchPromptTemplatesResponse struct {
	Body *SearchPromptTemplatesResponseBody
}
