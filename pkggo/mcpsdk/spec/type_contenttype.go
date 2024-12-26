package spec

import "github.com/flexigpt/flexiui/pkggo/mcpsdk/spec/customtype"

var enumValues_ContentType = []string{"text", "image", "resource"}

// ContentType
type ContentType struct {
	*customtype.StringUnion
}

// NewContentType creates a new ContentType with the provided value.
func NewContentType(value string) *ContentType {
	stringUnion := customtype.NewStringUnion(enumValues_ContentType...)
	_ = stringUnion.SetValue(value)
	return &ContentType{StringUnion: stringUnion}
}

// UnmarshalJSON implements json.Unmarshaler for ContentType.
func (c *ContentType) UnmarshalJSON(b []byte) error {
	if c.StringUnion == nil {
		// Initialize with allowed values if not already initialized
		c.StringUnion = customtype.NewStringUnion(enumValues_ContentType...)
	}
	return c.StringUnion.UnmarshalJSON(b)
}

// MarshalJSON implements json.Marshaler for ContentType.
func (c *ContentType) MarshalJSON() ([]byte, error) {
	return c.StringUnion.MarshalJSON()
}

// Content
var ContentTypeText = NewContentType("text")
var ContentTypeImage = NewContentType("image")
var ContentTypeEmbeddedResource = NewContentType("resource")
