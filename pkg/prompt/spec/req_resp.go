package spec

type PutPromptBundleRequest struct {
	BundleID BundleID                    `path:"bundleID" required:"true"`
	Body     *PutPromptBundleRequestBody `                                json:"body"`
}

type PutPromptBundleRequestBody struct {
	Slug        BundleSlug `json:"slug"                  required:"true"`
	DisplayName string     `json:"displayName"           required:"true"`
	IsEnabled   bool       `json:"isEnabled"             required:"true"`
	Description string     `json:"description,omitempty"`
}

type PutPromptBundleResponse struct{}

type DeletePromptBundleRequest struct {
	BundleID BundleID `path:"bundleID"`
}
type DeletePromptBundleResponse struct{}

type PatchPromptBundleRequest struct {
	BundleID BundleID                      `path:"bundleID" required:"true"`
	Body     *PatchPromptBundleRequestBody `                                json:"body"`
}

type PatchPromptBundleRequestBody struct {
	IsEnabled bool `json:"isEnabled" required:"true"`
}

type PatchPromptBundleResponse struct{}

type ListPromptBundlesRequest struct {
	BundleIDs       []BundleID `query:"bundleIDs"`
	IncludeDisabled bool       `query:"includeDisabled"`
	PageSize        int        `query:"pageSize"`
	PageToken       string     `query:"pageToken"`
}

type ListPromptBundlesResponse struct {
	Body *ListPromptBundlesResponseBody
}

type ListPromptBundlesResponseBody struct {
	PromptBundles []PromptBundle `json:"promptBundles"`
	NextPageToken *string        `json:"nextPageToken,omitempty"`
}

type PutPromptTemplateRequest struct {
	BundleID     BundleID                      `path:"bundleID"     required:"true"`
	TemplateSlug TemplateSlug                  `path:"templateSlug" required:"true"`
	Body         *PutPromptTemplateRequestBody `                                    json:"body"`
}

type PutPromptTemplateRequestBody struct {
	// Auto populate template id internally.
	DisplayName   string             `json:"displayName"             required:"true"`
	IsEnabled     bool               `json:"isEnabled"               required:"true"`
	Description   string             `json:"description,omitempty"`
	Tags          []string           `json:"tags,omitempty"`
	Blocks        []MessageBlock     `json:"blocks"                  required:"true"`
	Variables     []PromptVariable   `json:"variables,omitempty"`
	PreProcessors []PreProcessorCall `json:"preProcessors,omitempty"`
	Version       TemplateVersion    `json:"version"`
}

type PutPromptTemplateResponse struct{}

type DeletePromptTemplateRequest struct {
	BundleID     BundleID        `path:"bundleID"     required:"true"`
	TemplateSlug TemplateSlug    `path:"templateSlug" required:"true"`
	Version      TemplateVersion `                    required:"true" query:"version"`
}
type DeletePromptTemplateResponse struct{}

type PatchPromptTemplateRequest struct {
	BundleID     BundleID                        `path:"bundleID"     required:"true"`
	TemplateSlug TemplateSlug                    `path:"templateSlug" required:"true"`
	Body         *PatchPromptTemplateRequestBody `                                    json:"body"`
}

type PatchPromptTemplateRequestBody struct {
	Version   TemplateVersion `json:"version"   required:"true"`
	IsEnabled bool            `json:"isEnabled" required:"true"`
}

type PatchPromptTemplateResponse struct{}

type GetPromptTemplateRequest struct {
	BundleID     BundleID        `path:"bundleID"     required:"true"`
	TemplateSlug TemplateSlug    `path:"templateSlug" required:"true"`
	Version      TemplateVersion `                    required:"true" query:"version"`
}
type GetPromptTemplateResponse struct{ Body *PromptTemplate }

type ListPromptTemplatesRequest struct {
	BundleIDs           []BundleID `query:"bundleIDs"`
	Tags                []string   `query:"tags"`
	IncludeDisabled     bool       `query:"includeDisabled"`
	RecommendedPageSize int        `query:"recommendedPageSize"`
	PageToken           string     `query:"pageToken"`
}

type ListPromptTemplatesResponse struct {
	Body *ListPromptTemplatesResponseBody
}

type PromptTemplateListItem struct {
	BundleID        BundleID        `json:"bundleID"`
	BundleSlug      BundleSlug      `json:"bundleSlug"`
	TemplateSlug    TemplateSlug    `json:"templateSlug"`
	TemplateVersion TemplateVersion `json:"templateVersion"`
	IsBuiltIn       bool            `json:"isBuiltIn"`
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
