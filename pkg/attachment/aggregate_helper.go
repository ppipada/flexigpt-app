package attachment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
)

// BuildAttachmentForFile builds an Attachment for a local filesystem path.
// It inspects the MIME type / extension and chooses an appropriate
// AttachmentKind, default Mode, and AvailableContentBlockModes.
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
			Mode:  AttachmentContentBlockModeImage,
			AvailableContentBlockModes: []AttachmentContentBlockMode{
				AttachmentContentBlockModeImage,
			},
			ImageRef: &ImageRef{
				ImageInfo: fileutil.ImageInfo{
					PathInfo: *pathInfo,
				},
			},
		}
		if err := att.PopulateRef(false); err != nil {
			return nil, err
		}
		return att, nil

	case fileutil.ExtensionModeText:
		// Source code / markdown / text files: send as text by default.
		att := &Attachment{
			Kind:  AttachmentFile,
			Label: baseName,
			Mode:  AttachmentContentBlockModeText,
			AvailableContentBlockModes: []AttachmentContentBlockMode{
				AttachmentContentBlockModeText,
			},
			FileRef: &FileRef{
				PathInfo: *pathInfo,
			},
		}
		if err := att.PopulateRef(false); err != nil {
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
			Mode:  AttachmentContentBlockModeText,
			AvailableContentBlockModes: []AttachmentContentBlockMode{
				AttachmentContentBlockModeText,
				AttachmentContentBlockModeFile,
			},
			FileRef: &FileRef{
				PathInfo: *pathInfo,
			},
		}
		if err := att.PopulateRef(false); err != nil {
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

func buildUnreadableFileAttachment(pathInfo fileutil.PathInfo) *Attachment {
	return &Attachment{
		Kind:  AttachmentFile,
		Label: filepath.Base(pathInfo.Path),
		Mode:  AttachmentContentBlockModeNotReadable,
		AvailableContentBlockModes: []AttachmentContentBlockMode{
			AttachmentContentBlockModeNotReadable,
		},
		FileRef: &FileRef{
			PathInfo: pathInfo,
		},
	}
}

// BuildAttachmentForURL builds an Attachment for a remote URL.
// It does:
//   - infer default mode based on URL path extension
//   - set AvailableContentBlockModes accordingly
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
	mode := AttachmentContentBlockModePageContent
	available := []AttachmentContentBlockMode{
		AttachmentContentBlockModePageContent,
		AttachmentContentBlockModeLinkOnly,
	}

	mimeType, err := fileutil.MIMEFromExtensionString(ext)
	if err == nil {
		extMode, ok := fileutil.MIMETypeToExtensionMode[mimeType]
		if ok && extMode == fileutil.ExtensionModeImage {
			mode = AttachmentContentBlockModeImage
			available = []AttachmentContentBlockMode{
				AttachmentContentBlockModeImage,
				AttachmentContentBlockModeLinkOnly, // "link"
			}
		} else if mimeType == fileutil.MIMEApplicationPDF {
			mode = AttachmentContentBlockModeFile
			available = []AttachmentContentBlockMode{
				AttachmentContentBlockModeText,     // allow "text" view of PDF
				AttachmentContentBlockModeFile,     // original file
				AttachmentContentBlockModeLinkOnly, // link-only
			}
		}
	}

	att := &Attachment{
		Kind:                       AttachmentURL,
		Label:                      label,
		Mode:                       mode,
		AvailableContentBlockModes: available,
		URLRef: &URLRef{
			URL: trimmed,
		},
	}

	// Like BuildAttachmentForFile, ensure the ref is fully populated here.
	if err := att.PopulateRef(false); err != nil {
		return nil, err
	}

	return att, nil
}

// BuildContentBlocks converts high-level attachments (file paths, URLs, etc.)
// into provider-agnostic content blocks that can then be adapted for each LLM.
func BuildContentBlocks(ctx context.Context, atts []Attachment, opts ...ContentBlockOption) ([]ContentBlock, error) {
	if len(atts) == 0 {
		return nil, nil
	}
	blocks := make([]ContentBlock, 0, len(atts))
	buildContentOptions := getBuildContentBlockOptions(opts...)
	for _, att := range atts {
		b, err := (&att).BuildContentBlock(ctx, opts...)
		if err != nil {
			switch {
			case errors.Is(err, ErrExistingContentBlock):
				// If content block already existed we should just reattach it.
				b = att.ContentBlock
				slog.Warn("got existing att", "a", att)

			case errors.Is(err, ErrAttachmentModifiedSinceSnapshot) && !buildContentOptions.OverrideOriginal:
				displayBlock, err := att.GetTextBlockWithDisplayNameOnly(
					"attachment modified since this message was sent",
				)
				if err != nil {
					continue
				}
				b = displayBlock
			default:
				slog.Warn("failed to build content block for attachment", "err", err, "attachment", att)
				// Skip this content block. It is ok if the build block skipped this because OnlyIfTextKind was set or
				// any other error.
				continue
			}
		}

		blocks = append(blocks, *b)
	}

	return blocks, nil
}
