package attachment

type AttachmentContentBlockKind string

const (
	ContentBlockText  AttachmentContentBlockKind = "text"
	ContentBlockImage AttachmentContentBlockKind = "image"
	ContentBlockFile  AttachmentContentBlockKind = "file"
)

// ContentBlock represents a provider-agnostic chunk of content derived
// from an Attachment. Providers (OpenAI, Anthropic, etc.) adapt this
// into their own message/part formats.
type ContentBlock struct {
	Kind AttachmentContentBlockKind `json:"kind"`

	// For Kind == text: Text is populated.
	Text *string `json:"text,omitempty"`

	// For Kind == image or file: Base64Data + MIMEType are populated.
	Base64Data *string `json:"base64Data,omitempty"`
	MIMEType   *string `json:"mimeType,omitempty"`
	FileName   *string `json:"fileName,omitempty"`
}
