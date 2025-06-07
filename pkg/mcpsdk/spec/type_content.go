package spec

var enumValuesContentType = []string{"text", "image", "resource"}

// ContentType is a union of all contents supported as resources.
type ContentType struct {
	*StringUnion
}

// NewContentType creates a new ContentType with the provided value.
func NewContentType(value string) *ContentType {
	stringUnion := NewStringUnion(enumValuesContentType...)
	_ = stringUnion.SetValue(value)
	return &ContentType{StringUnion: stringUnion}
}

// UnmarshalJSON implements json.Unmarshaler for ContentType.
func (c *ContentType) UnmarshalJSON(b []byte) error {
	if c.StringUnion == nil {
		// Initialize with allowed values if not already initialized.
		c.StringUnion = NewStringUnion(enumValuesContentType...)
	}
	return c.StringUnion.UnmarshalJSON(b)
}

// MarshalJSON implements json.Marshaler for ContentType.
func (c *ContentType) MarshalJSON() ([]byte, error) {
	return c.StringUnion.MarshalJSON()
}

// Content.
var (
	ContentTypeText             = NewContentType("text")
	ContentTypeImage            = NewContentType("image")
	ContentTypeEmbeddedResource = NewContentType("resource")
)

type Content struct {
	// Type corresponds to the JSON schema field "type".
	Type ContentType `json:"type"                  yaml:"type"                  mapstructure:"type"`
	// Annotations corresponds to the JSON schema field "annotations".
	Annotations *Annotations `json:"annotations,omitempty" yaml:"annotations,omitempty" mapstructure:"annotations,omitempty"`
	// TextContent only: The text content of the message.
	Text *string `json:"text"                  yaml:"text"                  mapstructure:"text"`

	// ImageContent only: The base64-encoded image data.
	Data *string `json:"data" yaml:"data" mapstructure:"data"`

	// ImageContent only: The MIME type of the image. Different providers may support different image types.
	MimeType *string `json:"mimeType" yaml:"mimeType" mapstructure:"mimeType"`

	// EmbeddedResourceContent only: Resource corresponds to the JSON schema field "resource".
	Resource *ResourceContent `json:"resource" yaml:"resource" mapstructure:"resource"`
}

type ResourceContent struct {
	// The MIME type of this resource, if known.
	MimeType *string `json:"mimeType,omitempty" yaml:"mimeType,omitempty" mapstructure:"mimeType,omitempty"`

	// The text of the item. This must only be set if the item can actually be
	// represented as text (not binary data).
	Text *string `json:"text,omitempty" yaml:"text" mapstructure:"text"`

	// A base64-encoded string representing the binary data of the item.
	Blob *string `json:"blob,omitempty" yaml:"blob" mapstructure:"blob"`

	// The URI of this resource.
	URI string `json:"uri" yaml:"uri" mapstructure:"uri"`
}

// A template description for resources available on the server.
type ResourceTemplate struct {
	// Annotations corresponds to the JSON schema field "annotations".
	Annotations *Annotations `json:"annotations,omitempty"`

	// A URI template (according to RFC 6570) that can be used to construct resource
	// URIs.
	URITemplate string `json:"uriTemplate"`

	// A human-readable name for the type of resource this template refers to.
	//
	// This can be used by clients to populate UI elements.
	Name string `json:"name"`

	// A description of what this template is for.
	//
	// This can be used by clients to improve the LLM's understanding of available
	// resources. It can be thought of like a "hint" to the model.
	Description *string `json:"description,omitempty"`

	// The MIME type for all resources that match this template. This should only be
	// included if all resources matching this template have the same type.
	MimeType *string `json:"mimeType,omitempty"`
}
