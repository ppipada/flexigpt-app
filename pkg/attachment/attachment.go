package attachment

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
)

var (
	ErrNonTextContentBlock             = errors.New("content block is not of kind - text")
	ErrExistingContentBlock            = errors.New("content block already exists")
	ErrAttachmentModifiedSinceSnapshot = errors.New("attachment modified since snapshot")
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
	Text *string `json:"text,omitempty"`

	// For Kind == image or file: Base64Data + MIMEType are populated.
	Base64Data *string `json:"base64Data,omitempty"`
	MIMEType   *string `json:"mimeType,omitempty"`
	FileName   *string `json:"fileName,omitempty"`
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

	ContentBlock *ContentBlock `json:"contentBlock,omitempty"`
}

type DirectoryAttachmentsResult struct {
	DirPath      string                           `json:"dirPath"`
	Attachments  []Attachment                     `json:"attachments"`  // included attachments (flattened)
	OverflowDirs []fileutil.DirectoryOverflowInfo `json:"overflowDirs"` // directories not fully included
	MaxFiles     int                              `json:"maxFiles"`     // max number of files returned (after clamping)
	TotalSize    int64                            `json:"totalSize"`    // sum of Files[i].Size
	HasMore      bool                             `json:"hasMore"`      // true if not all content included
}

type buildContentBlockOptions struct {
	OverrideOriginal bool
	OnlyIfTextKind   bool
	ForceFetch       bool
}

type ContentBlockOption func(*buildContentBlockOptions)

// WithOverrideOriginalContentBlock. Default false.
// False = If this is an attachment from a previous turn whose underlying ref (file path, size, mod time, etc.) has
// changed, do not re-send the possibly mismatched content.
func WithOverrideOriginalContentBlock(override bool) ContentBlockOption {
	return func(o *buildContentBlockOptions) {
		o.OverrideOriginal = override
	}
}

// WithOnlyTextKindContentBlock. Default False.
// False = Build content block of any kind.
// True = Return content block only if it is of kind text, else error.
func WithOnlyTextKindContentBlock(textOnly bool) ContentBlockOption {
	return func(o *buildContentBlockOptions) {
		o.OnlyIfTextKind = textOnly
	}
}

// WithForceFetchContentBlock. Default False.
// False = Dont fetch the content block if already present in attachment.
// True = Override content block even if present.
func WithForceFetchContentBlock(forceFetch bool) ContentBlockOption {
	return func(o *buildContentBlockOptions) {
		o.ForceFetch = forceFetch
	}
}

// BuildContentBlock function builds and returns a content block for an attachment.
// It does NOT attach the content block to the attachment.
func (att *Attachment) BuildContentBlock(ctx context.Context, opts ...ContentBlockOption,
) (*ContentBlock, error) {
	buildContentOptions := getBuildContentBlockOptions(opts...)

	// Ensure refs are populated; caller may have done this earlier,
	// but calling again on a populated ref is cheap to do in actual read data path.
	if err := (att).PopulateRef(buildContentOptions.OverrideOriginal); err != nil {
		return nil, err
	}

	if !buildContentOptions.ForceFetch && att.ContentBlock != nil {
		return nil, ErrExistingContentBlock
	}

	if !buildContentOptions.OverrideOriginal && att.isModifiedSinceSnapshot() {
		return nil, ErrAttachmentModifiedSinceSnapshot
	}

	switch att.Kind {
	case AttachmentDocIndex, AttachmentPR, AttachmentCommit:
		return getUnreadableBlock(att)

	case AttachmentImage:
		if att.ImageRef == nil || !att.ImageRef.Exists {
			return nil, errors.New("invalid image ref for attachment")
		}
		if buildContentOptions.OnlyIfTextKind {
			return nil, ErrNonTextContentBlock
		}
		return buildImageBlockFromLocal(att.ImageRef.Path)

	case AttachmentFile:
		if att.FileRef == nil || !att.FileRef.Exists {
			return nil, errors.New("invalid file ref for attachment")
		}
		return buildBlocksForLocalFile(ctx, att, buildContentOptions.OnlyIfTextKind)

	case AttachmentURL:
		if att.URLRef == nil || att.URLRef.URL == "" {
			return nil, errors.New("invalid url ref for attachment")
		}
		return buildBlocksForURL(ctx, att, buildContentOptions.OnlyIfTextKind)

	default:
		return nil, errors.New("unknown attachment kind")
	}
}

func (att *Attachment) PopulateRef(replaceOrig bool) error {
	switch att.Kind {
	case AttachmentFile:
		if att.FileRef == nil {
			return errors.New("no file ref for file attachment")
		}
		if err := att.FileRef.PopulateRef(replaceOrig); err != nil {
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
		if err := att.ImageRef.PopulateRef(replaceOrig); err != nil {
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
		if err := att.URLRef.PopulateRef(replaceOrig); err != nil {
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
		if err := att.GenericRef.PopulateRef(replaceOrig); err != nil {
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

func (att *Attachment) isModifiedSinceSnapshot() bool {
	switch att.Kind {
	case AttachmentFile:
		if att.FileRef != nil {
			return att.FileRef.IsModified()
		}
	case AttachmentImage:
		if att.ImageRef != nil {
			return att.ImageRef.IsModified()
		}
	case AttachmentURL:
		if att.URLRef != nil {
			return att.URLRef.IsModified()
		}
	case AttachmentDocIndex, AttachmentPR, AttachmentCommit:
		if att.GenericRef != nil {
			return att.GenericRef.IsModified()
		}
	default:
		return false
	}
	return false
}

func buildUnreadableFileAttachment(pathInfo fileutil.PathInfo) *Attachment {
	return &Attachment{
		Kind:  AttachmentFile,
		Label: filepath.Base(pathInfo.Path),
		Mode:  AttachmentModeNotReadable,
		AvailableModes: []AttachmentMode{
			AttachmentModeNotReadable,
		},
		FileRef: &FileRef{
			PathInfo: pathInfo,
		},
	}
}

func getBuildContentBlockOptions(opts ...ContentBlockOption) *buildContentBlockOptions {
	options := &buildContentBlockOptions{
		OverrideOriginal: false,
		OnlyIfTextKind:   false,
		ForceFetch:       false,
	}

	// Apply user-specified options.
	for _, opt := range opts {
		opt(options)
	}
	return options
}
