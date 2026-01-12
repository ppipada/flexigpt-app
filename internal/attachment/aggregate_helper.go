package attachment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/internal/fileutil"
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
// It uses a timeout context to peek and infer type and then give proper options.
func BuildAttachmentForURL(rawURL string) (*Attachment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return BuildAttachmentForURLWithContext(ctx, rawURL)
}

// BuildAttachmentForURLWithContext builds an Attachment for a remote URL.
//
// Mode detection strategy:
//  1. Best effort Content-Type detection via HEAD -> Range GET sniff.
//  2. Fallback to extension-based detection.
//  3. Ultimate fallback: LinkOnly (do not error out due to detection failures).
//
// Note: It can still return an error for invalid/empty/non-absolute URLs because
// PopulateRef enforces validity.
func BuildAttachmentForURLWithContext(ctx context.Context, rawURL string) (*Attachment, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil, errors.New("empty url")
	}

	label := trimmed

	// Parse early mainly to infer extension; PopulateRef will validate absolute URL later.
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, errors.New("invalid url")
	}

	// Extension fallback inference.
	ext := ""
	pathname := strings.ToLower(parsed.Path)
	if pathname != "" {
		parts := strings.Split(pathname, ".")
		ext = parts[len(parts)-1]
	}

	// Start from ultimate fallback: link-only.
	mode := AttachmentContentBlockModeTextLink
	available := []AttachmentContentBlockMode{
		AttachmentContentBlockModeTextLink,
	}

	// 1) Best-effort probe Content-Type (no full body download).
	// If this fails, we do NOT error out: we just log and move to extension fallback.
	if ct, err := peekURLContentType(ctx, trimmed); err == nil {
		ct = normalizeContentType(ct)

		switch {
		case strings.HasPrefix(ct, "image/"):
			mode = AttachmentContentBlockModeImage
			available = []AttachmentContentBlockMode{
				AttachmentContentBlockModeImage,
				AttachmentContentBlockModeTextLink,
			}

		case ct == string(fileutil.MIMEApplicationPDF):
			mode = AttachmentContentBlockModeFile
			available = []AttachmentContentBlockMode{
				AttachmentContentBlockModeText, // allow PDF -> extracted text view
				AttachmentContentBlockModeFile, // raw file
				AttachmentContentBlockModeTextLink,
			}

		case isPlainTextContentType(ct):
			mode = AttachmentContentBlockModeText
			available = []AttachmentContentBlockMode{
				AttachmentContentBlockModeText,
				AttachmentContentBlockModeTextLink,
			}

		case strings.HasPrefix(ct, "text/html"):
			mode = AttachmentContentBlockModePageContent
			available = []AttachmentContentBlockMode{
				AttachmentContentBlockModePageContent,
				AttachmentContentBlockModeTextLink,
			}

		case ct != "":
			// Some other known binary type.
			mode = AttachmentContentBlockModeFile
			available = []AttachmentContentBlockMode{
				AttachmentContentBlockModeFile,
				AttachmentContentBlockModeTextLink,
			}
		}
	} else {
		slog.Debug("content-type probe failed; falling back to extension/link-only",
			"url", trimmed, "err", err)
	}

	// 2) Extension-based fallback ONLY if we still are in LinkOnly mode.
	// (i.e., content-type probe didn't identify anything).
	if mode == AttachmentContentBlockModeTextLink && ext != "" {
		if mimeType, err := fileutil.MIMEFromExtensionString(ext); err == nil {
			switch {
			case func() bool {
				extMode, ok := fileutil.MIMETypeToExtensionMode[mimeType]
				return ok && extMode == fileutil.ExtensionModeImage
			}():
				mode = AttachmentContentBlockModeImage
				available = []AttachmentContentBlockMode{
					AttachmentContentBlockModeImage,
					AttachmentContentBlockModeTextLink,
				}

			case mimeType == fileutil.MIMEApplicationPDF:
				mode = AttachmentContentBlockModeFile
				available = []AttachmentContentBlockMode{
					AttachmentContentBlockModeText,
					AttachmentContentBlockModeFile,
					AttachmentContentBlockModeTextLink,
				}

			case strings.HasPrefix(string(mimeType), "text/"):
				mode = AttachmentContentBlockModeText
				available = []AttachmentContentBlockMode{
					AttachmentContentBlockModeText,
					AttachmentContentBlockModeTextLink,
				}

			default:
				// Unknown-but-present extension -> safest treat as file.
				mode = AttachmentContentBlockModeFile
				available = []AttachmentContentBlockMode{
					AttachmentContentBlockModeFile,
					AttachmentContentBlockModeTextLink,
				}
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

	// Ensure ref is populated/validated (absolute URL requirement, normalized fields, etc.)
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
