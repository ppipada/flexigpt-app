package attachment

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
)

var (
	ErrContentBlockSkipped = errors.New("content block skipped")
	ErrAttachmentModified  = errors.New("attachment modified since snapshot")
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

// BuildAttachmentForFile builds an Attachment for a local filesystem path.
// It inspects the MIME type / extension and chooses an appropriate
// AttachmentKind, default Mode, and AvailableModes.
// The returned attachment is fully populated via PopulateRef().
// Note that this builds a fresh attachment, i.e both original ref and current are populated here.
func BuildAttachmentForFile(pathInfo *fileutil.PathInfo) (*Attachment, error) {
	if pathInfo == nil {
		return nil, errors.New("invalid input pathinfo")
	}

	if !pathInfo.Exists {
		return nil, fmt.Errorf("file does not exist: %s", pathInfo.Path)
	}
	if pathInfo.IsDir {
		return nil, fmt.Errorf("path %q is a directory; expected file", pathInfo.Path)
	}

	mimeType, extMode, err := fileutil.MIMEForLocalFile(pathInfo.Path)
	if err != nil {
		return nil, err
	}

	baseName := filepath.Base(pathInfo.Path)

	switch extMode {
	case fileutil.ExtensionModeImage:
		// Treat images as dedicated image attachments.
		att := &Attachment{
			Kind:  AttachmentImage,
			Label: baseName,
			Mode:  AttachmentModeImage,
			AvailableModes: []AttachmentMode{
				AttachmentModeImage,
			},
			ImageRef: &ImageRef{
				ImageInfo: fileutil.ImageInfo{
					PathInfo: *pathInfo,
				},
			},
		}
		if err := att.PopulateRef(); err != nil {
			return nil, err
		}
		return att, nil

	case fileutil.ExtensionModeText:
		// Source code / markdown / text files: send as text by default.
		att := &Attachment{
			Kind:  AttachmentFile,
			Label: baseName,
			Mode:  AttachmentModeText,
			AvailableModes: []AttachmentMode{
				AttachmentModeText,
			},
			FileRef: &FileRef{
				PathInfo: *pathInfo,
			},
		}
		if err := att.PopulateRef(); err != nil {
			return nil, err
		}
		return att, nil

	case fileutil.ExtensionModeDocument:
		// Documents (PDF, Office, etc.).
		// As of now APIs and we internally only support PDF docs.
		// PDFs can be treated as text (with extraction) or as original file.
		if mimeType != fileutil.MIMEApplicationPDF {
			return buildUnreadableFileAttachment(*pathInfo), nil
		}

		att := &Attachment{
			Kind:  AttachmentFile,
			Label: baseName,
			Mode:  AttachmentModeText,
			AvailableModes: []AttachmentMode{
				AttachmentModeText,
				AttachmentModeFile,
			},
			FileRef: &FileRef{
				PathInfo: *pathInfo,
			},
		}
		if err := att.PopulateRef(); err != nil {
			return nil, err
		}
		return att, nil

	case fileutil.ExtensionModeDefault:
		return buildUnreadableFileAttachment(*pathInfo), nil

	default:
		// Unknown / binary. We still keep it as a file attachment but mark it not-readable so BuildContentBlock
		// produces a short placeholder instead of trying to read it.
		return buildUnreadableFileAttachment(*pathInfo), nil

	}
}

// BuildAttachmentForURL builds an Attachment for a remote URL.
// It does:
//   - infer default mode based on URL path extension
//   - set AvailableModes accordingly
//   - populate URLRef (Normalized / OrigNormalized) via PopulateRef.
func BuildAttachmentForURL(rawURL string) (*Attachment, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil, errors.New("empty url")
	}

	label := trimmed

	// Infer extension from URL pathname, like the TS helper does.
	// If parsing fails, ext stays empty and we fall back to "page".
	ext := ""
	if parsed, err := url.Parse(trimmed); err != nil {
		return nil, errors.New("invalid url")
	} else {
		pathname := strings.ToLower(parsed.Path)
		if pathname != "" {
			parts := strings.Split(pathname, ".")
			ext = parts[len(parts)-1]
		}
	}

	// Choose default mode + available modes based on extension.
	mode := AttachmentModePageContent
	available := []AttachmentMode{
		AttachmentModePageContent,
		AttachmentModeLinkOnly,
	}

	mimeType, err := fileutil.MIMEFromExtensionString(ext)
	if err == nil {
		extMode, ok := fileutil.MIMETypeToExtensionMode[mimeType]
		if ok && extMode == fileutil.ExtensionModeImage {
			mode = AttachmentModeImage
			available = []AttachmentMode{
				AttachmentModeImage,
				AttachmentModeLinkOnly, // "link"
			}
		} else if mimeType == fileutil.MIMEApplicationPDF {
			mode = AttachmentModeFile
			available = []AttachmentMode{
				AttachmentModeText,     // allow "text" view of PDF
				AttachmentModeFile,     // original file
				AttachmentModeLinkOnly, // link-only
			}
		}
	}

	att := &Attachment{
		Kind:           AttachmentURL,
		Label:          label,
		Mode:           mode,
		AvailableModes: available,
		URLRef: &URLRef{
			URL: trimmed,
		},
	}

	// Like BuildAttachmentForFile, ensure the ref is fully populated here.
	if err := att.PopulateRef(); err != nil {
		return nil, err
	}

	return att, nil
}

// BuildContentBlocks converts high-level attachments (file paths, URLs, etc.)
// into provider-agnostic content blocks that can then be adapted for each LLM.
func BuildContentBlocks(ctx context.Context, atts []Attachment, overrideOriginal bool) ([]ContentBlock, error) {
	if len(atts) == 0 {
		return nil, nil
	}
	blocks := make([]ContentBlock, 0, len(atts))

	for _, att := range atts {
		b, err := (&att).BuildContentBlock(ctx, overrideOriginal)
		if err != nil {
			if errors.Is(err, ErrAttachmentModified) && overrideOriginal {
				txt := att.FormatAsDisplayName()
				if txt == "" {
					txt = "[Attachment]"
				}
				txt += " (attachment modified since this message was sent)"
				b = &ContentBlock{
					Kind: ContentBlockText,
					Text: &txt,
				}
			} else {
				continue
			}
		}
		if err != nil {
		}
		blocks = append(blocks, *b)
	}

	return blocks, nil
}

func (att *Attachment) BuildContentBlock(ctx context.Context, overrideOriginal bool) (*ContentBlock, error) {
	// Ensure refs are populated; caller may have done this earlier,
	// but calling again on a populated ref is cheap to do in actual read data path.
	if err := (att).PopulateRef(); err != nil {
		return nil, err
	}

	// If this is an attachment from a previous turn whose underlying ref (file path, size, mod time, etc.) has changed,
	// do not re-send the possibly mismatched content.
	if !overrideOriginal && att.isModifiedSinceSnapshot() {
		return nil, ErrAttachmentModified
	}

	mode := att.Mode
	switch mode {
	case AttachmentModeText, AttachmentModeFile,
		AttachmentModeImage,
		AttachmentModePageContent,
		AttachmentModeLinkOnly:
	// Ok.
	case AttachmentModePRDiff,
		AttachmentModePRPage,
		AttachmentModeCommitDiff,
		AttachmentModeCommitPage, AttachmentModeNotReadable, "":
		return getUnreadableBlock(att)

	default:
		return nil, errors.New("invalid processing mode for attachment")
	}

	switch att.Kind {
	case AttachmentFile:
		if att.FileRef == nil || !att.FileRef.Exists {
			return nil, errors.New("invalid file ref for attachment")
		}
		return buildBlocksForLocalFile(ctx, att, mode)

	case AttachmentImage:
		if att.ImageRef == nil || !att.ImageRef.Exists {
			return nil, errors.New("invalid image ref for attachment")
		}
		return buildImageBlockFromLocal(att.ImageRef.Path)

	case AttachmentURL:
		if att.URLRef == nil || att.URLRef.URL == "" {
			return nil, errors.New("invalid url ref for attachment")
		}
		return buildBlocksForURL(ctx, att, mode)

	case AttachmentDocIndex, AttachmentPR, AttachmentCommit:
		return getUnreadableBlock(att)

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
