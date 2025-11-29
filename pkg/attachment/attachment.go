package attachment

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type ContentBlockKind string

const (
	ContentBlockText  ContentBlockKind = "text"
	ContentBlockImage ContentBlockKind = "image"
	ContentBlockFile  ContentBlockKind = "file"
)
const maxAttachmentFetchBytes = 16 * 1024 * 1024 // 16MB safety limit

// ContentBlock represents a provider-agnostic chunk of content derived
// from an Attachment. Providers (OpenAI, Anthropic, etc.) adapt this
// into their own message/part formats.
type ContentBlock struct {
	Kind ContentBlockKind `json:"kind"`

	// For Kind == text: Text is populated.
	Text string `json:"text,omitempty"`

	// For Kind == image or file: Base64Data + MIMEType are populated.
	Base64Data string `json:"base64Data,omitempty"`
	MIMEType   string `json:"mimeType,omitempty"`
	FileName   string `json:"fileName,omitempty"`
}

// AttachmentKind enumerates contextual attachment categories that can be
// associated with messages sent to the inference layer.
type AttachmentKind string

const (
	AttachmentFile     AttachmentKind = "file"
	AttachmentImage    AttachmentKind = "image"
	AttachmentURL      AttachmentKind = "url"
	AttachmentDocIndex AttachmentKind = "docIndex"
	AttachmentPR       AttachmentKind = "pr"
	AttachmentCommit   AttachmentKind = "commit"
)

// AttachmentMode describes how a given attachment should be used
// for the current turn.
type AttachmentMode string

const (
	AttachmentModeText        AttachmentMode = "text"         // "Text content"
	AttachmentModeFile        AttachmentMode = "file"         // "File (original format)"
	AttachmentModeImage       AttachmentMode = "image"        // Image rendering
	AttachmentModePageContent AttachmentMode = "page"         // "Page content" for HTML/URLs
	AttachmentModeLinkOnly    AttachmentMode = "link"         // "Link only" – no fetch
	AttachmentModeNotReadable AttachmentMode = "not-readable" // Binary/unknown – cannot process

	AttachmentModePRDiff     AttachmentMode = "pr-diff"
	AttachmentModePRPage     AttachmentMode = "pr-page"
	AttachmentModeCommitDiff AttachmentMode = "commit-diff"
	AttachmentModeCommitPage AttachmentMode = "commit-page"
)

// Attachment is a lightweight reference to external context (files, docs, images, etc.).
type Attachment struct {
	Kind  AttachmentKind `json:"kind"`
	Label string         `json:"label"`

	// Mode selected for this attachment for the current turn.
	// Frontend picks smart defaults; backend can override missing values.
	Mode AttachmentMode `json:"mode,omitempty"`
	// Optional: allowed modes for this attachment (primarily for UI).
	AvailableModes []AttachmentMode `json:"availableModes,omitempty"`

	// Exactly one field below should be non-nil.
	FileRef    *FileRef    `json:"fileRef,omitempty"`
	ImageRef   *ImageRef   `json:"imageRef,omitempty"`
	URLRef     *URLRef     `json:"urlRef,omitempty"`
	GenericRef *GenericRef `json:"genericRef,omitempty"`
}

// BuildContentBlocks converts high-level attachments (file paths, URLs, etc.)
// into provider-agnostic content blocks that can then be adapted for each LLM.
func BuildContentBlocks(ctx context.Context, atts []Attachment) ([]ContentBlock, error) {
	if len(atts) == 0 {
		return nil, nil
	}
	blocks := make([]ContentBlock, 0, len(atts))

	for _, att := range atts {
		b, err := (&att).BuildContentBlock(ctx)
		if err != nil {
			continue
		}
		blocks = append(blocks, *b)
	}

	return blocks, nil
}

func (att *Attachment) BuildContentBlock(ctx context.Context) (*ContentBlock, error) {
	// Ensure refs are populated; caller may have done this earlier,
	// but calling again on a populated ref is cheap.
	if err := (att).PopulateRef(); err != nil {
		return nil, err
	}

	mode := att.Mode
	if mode == "" {
		mode = att.inferDefaultMode()
	}

	switch att.Kind {
	case AttachmentImage:
		if att.ImageRef == nil || !att.ImageRef.Exists {
			return nil, errors.New("invalid image ref for attachment")
		}
		return buildImageBlockFromLocal(att.ImageRef.Path)

	case AttachmentFile:
		if att.FileRef == nil || !att.FileRef.Exists {
			return nil, errors.New("invalid file ref for attachment")
		}
		return buildBlocksForLocalFile(ctx, att, mode)

	case AttachmentURL:
		if att.URLRef == nil || att.URLRef.URL == "" {
			return nil, errors.New("invalid url ref for attachment")
		}
		return buildBlocksForURL(ctx, att, mode)

	case AttachmentDocIndex, AttachmentPR, AttachmentCommit:
		// For now, treat as a simple text mention.
		if txt := (att).FormatAsDisplayName(); txt != "" {
			return &ContentBlock{
				Kind: ContentBlockText,
				Text: txt,
			}, nil
		}
		return nil, errors.New("invalid attachment kind")
	default:
		return nil, errors.New("unknown attachment kind")
	}
}

func (att *Attachment) PopulateRef() error {
	switch att.Kind {
	case AttachmentFile:
		if att.FileRef == nil {
			return errors.New("no file ref for file attachment")
		}
		if err := att.FileRef.PopulateRef(); err != nil {
			return err
		}
		if att.Label == "" {
			att.Label = filepath.Base(att.FileRef.Path)
		}

		return nil

	case AttachmentImage:
		if att.ImageRef == nil {
			return errors.New("no image ref for image attachment")
		}
		if err := att.ImageRef.PopulateRef(); err != nil {
			return err
		}
		if att.Label == "" {
			att.Label = filepath.Base(att.ImageRef.Path)
		}
		if att.Mode == "" {
			att.Mode = AttachmentModeImage
			att.AvailableModes = []AttachmentMode{AttachmentModeImage}
		}
		return nil

	case AttachmentURL:
		if att.URLRef == nil {
			return errors.New("no url ref for url attachment")
		}
		if err := att.URLRef.PopulateRef(); err != nil {
			return err
		}
		if att.Label == "" {
			att.Label = att.URLRef.URL
		}
		// Mode/AvailableModes are usually set by frontend; you could
		// add backend inference here (by extension or content-type)
		// if att.Mode == "".
		return nil

	case AttachmentCommit,
		AttachmentDocIndex,
		AttachmentPR:
		if att.GenericRef == nil {
			return errors.New("no generic ref for attachment")
		}
		if err := att.GenericRef.PopulateRef(); err != nil {
			return err
		}
		if att.Label == "" {
			att.Label = att.GenericRef.Handle
		}
		return nil

	default:
		return errors.New("unknown attachment kind")
	}
}

// FormatAsDisplayName normalizes an attachment into a short, human-readable
// string that can be injected into text prompts when a richer modality is not
// available.
func (att *Attachment) FormatAsDisplayName() string {
	label := strings.TrimSpace(att.Label)
	var detail string

	switch att.Kind {
	case AttachmentFile:
		if att.FileRef != nil {
			detail = strings.TrimSpace(att.FileRef.Path)
		}
	case AttachmentImage:
		if att.ImageRef != nil {
			detail = strings.TrimSpace(att.ImageRef.Path)
		}
	case AttachmentURL:
		if att.URLRef != nil {
			detail = strings.TrimSpace(att.URLRef.URL)
		}
	case AttachmentDocIndex,
		AttachmentPR,
		AttachmentCommit:
		if att.GenericRef != nil {
			detail = strings.TrimSpace(att.GenericRef.Handle)
		}
	default:
		detail = ""
	}

	if label == "" {
		label = detail
	}
	if label == "" && detail == "" {
		return ""
	}
	if detail == "" || label == detail {
		return fmt.Sprintf("[Attachment: %s]", label)
	}
	return fmt.Sprintf("[Attachment: %s — %s]", label, detail)
}

// inferDefaultMode picks a reasonable default mode when client left it empty.
// It respects Kind + file extension but does NOT override explicit user choices.
func (att *Attachment) inferDefaultMode() AttachmentMode {
	switch att.Kind {
	case AttachmentImage:
		return AttachmentModeImage

	case AttachmentFile:
		if att.FileRef == nil {
			return AttachmentModeFile
		}
		return inferDefaultFileMode(att.FileRef.Path)

	case AttachmentURL:
		if att.URLRef == nil {
			return AttachmentModeLinkOnly
		}
		return inferDefaultURLMode(att.URLRef.URL)

	case AttachmentDocIndex, AttachmentPR, AttachmentCommit:
		return AttachmentModeText

	default:
		return AttachmentModeText
	}
}
