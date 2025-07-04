package spec

type PutPromptBundleRequest struct {
	BundleID string                      `path:"bundleID" required:"true"`
	Body     *PutPromptBundleRequestBody `                                json:"body"`
}

type PutPromptBundleRequestBody struct {
	Slug        string `json:"slug"                  required:"true"`
	DisplayName string `json:"displayName"           required:"true"`
	IsEnabled   bool   `json:"isEnabled"             required:"true"`
	Description string `json:"description,omitempty"`
}

type PutPromptBundleResponse struct{}

type DeletePromptBundleRequest struct {
	BundleID string `path:"bundleID"`
}
type DeletePromptBundleResponse struct{}

type PatchPromptBundleRequest struct {
	BundleID string                        `path:"bundleID" required:"true"`
	Body     *PatchPromptBundleRequestBody `                                json:"body"`
}

type PatchPromptBundleRequestBody struct {
	IsEnabled bool `json:"isEnabled" required:"true"`
}

type PatchPromptBundleResponse struct{}

type ListPromptBundlesRequest struct {
	BundleIDs       []string `query:"bundleIDs"`
	IncludeDisabled bool     `query:"includeDisabled"`

	PageSize  int    `query:"pageSize"`
	PageToken string `query:"pageToken"`
}

type ListPromptBundlesResponse struct {
	Body *ListPromptBundlesResponseBody
}

type ListPromptBundlesResponseBody struct {
	PromptBundles []PromptBundle `json:"promptBundles"`
	NextPageToken *string        `json:"nextPageToken,omitempty"`
}

type PutPromptTemplateRequest struct {
	BundleID   string                        `path:"bundleID"   required:"true"`
	TemplateID string                        `path:"templateID" required:"true"`
	Body       *PutPromptTemplateRequestBody `                                  json:"body"`
}

type PutPromptTemplateRequestBody struct {
	DisplayName string   `json:"displayName"           required:"true"`
	Slug        string   `json:"slug"                  required:"true"`
	IsEnabled   bool     `json:"isEnabled"             required:"true"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	// Ordered list of blocks that form the final prompt.
	Blocks []MessageBlock `json:"blocks"                  required:"true"`
	// Declared placeholders.
	Variables []PromptVariable `json:"variables,omitempty"`
	// Helper steps executed before the prompt is sent.
	PreProcessors []PreProcessorCall `json:"preProcessors,omitempty"`

	Version string `json:"version"`
}

type PutPromptTemplateResponse struct{}

type DeletePromptTemplateRequest struct {
	BundleID   string `path:"bundleID"   required:"true"`
	TemplateID string `path:"templateID" required:"true"`
	Version    string `                  required:"true" query:"version"`
}
type DeletePromptTemplateResponse struct{}

type PatchPromptTemplateRequest struct {
	BundleID   string                          `path:"bundleID"   required:"true"`
	TemplateID string                          `path:"templateID" required:"true"`
	Body       *PatchPromptTemplateRequestBody `                                  json:"body"`
}

type PatchPromptTemplateRequestBody struct {
	Version   string `json:"version"   required:"true"`
	IsEnabled bool   `json:"isEnabled" required:"true"`
}

type PatchPromptTemplateResponse struct{}

type GetPromptTemplateRequest struct {
	BundleID   string `path:"bundleID"   required:"true"`
	TemplateID string `path:"templateID" required:"true"`
	Version    string `                                  query:"version,omitempty"`
}
type GetPromptTemplateResponse struct{ Body *PromptTemplate }

type ListPromptTemplatesRequest struct {
	BundleIDs       []string `query:"bundleIDs"`
	Tags            []string `query:"tags"`
	IncludeDisabled bool     `query:"includeDisabled"`
	AllVersions     bool     `query:"allVersions"`

	PageSize  int    `query:"pageSize"`
	PageToken string `query:"pageToken"`
}

type ListPromptTemplatesResponse struct {
	Body *ListPromptTemplatesResponseBody
}

type PromptTemplateListItem struct {
	BundleID        string `json:"bundleID"`
	BundleSlug      string `json:"bundleSlug"`
	TemplateID      string `json:"templateID"`
	TemplateSlug    string `json:"templateSlug"`
	TemplateVersion string `json:"templateVersion"`
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
