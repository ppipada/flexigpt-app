package attachment

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"
)

// URLRef carries metadata for URL-based attachments.
//
// URL:          the raw user-provided URL (after trimming whitespace).
// Normalized:   a canonicalized string representation used internally.
// OrigNormalized: snapshot of the original Normalized value so we can
//
//	detect whether a URL has been modified in-place.
type URLRef struct {
	URL            string `json:"url"`
	Normalized     string `json:"normalized,omitempty"`
	OrigNormalized string `json:"origNormalized"`
}

// PopulateRef validates and normalizes the URL stored in the URLRef.
// It must be called before the URLRef is used.
//
// It ensures:
//   - The URL is non-empty.
//   - The URL parses successfully.
//   - The URL is absolute (has a scheme/host).
//   - Normalized and OrigNormalized are populated.
func (ref *URLRef) PopulateRef() error {
	if ref == nil {
		return errors.New("url attachment missing ref")
	}
	raw := strings.TrimSpace(ref.URL)
	if raw == "" {
		return errors.New("url attachment missing url")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url %q: %w", raw, err)
	}
	if !parsed.IsAbs() {
		return fmt.Errorf("url %q must be absolute", raw)
	}

	nURL := parsed.String()
	ref.URL = raw
	ref.Normalized = nURL
	if strings.TrimSpace(ref.OrigNormalized) == "" {
		ref.OrigNormalized = nURL
	}
	return nil
}

// IsModified reports whether the URL has been modified from its original
// normalized form. This is useful for detecting in-place edits.
func (ref *URLRef) IsModified() bool {
	if ref == nil {
		return false
	}
	if strings.TrimSpace(ref.OrigNormalized) == "" {
		return false
	}
	return ref.Normalized != ref.OrigNormalized
}

// buildBlocksForURL builds a ContentBlock representation for a URL-based
// attachment, depending on the desired AttachmentMode.
//
// The behaviour is:
//   - AttachmentModeLinkOnly:      always returns a simple text link.
//   - AttachmentModeImage:         tries to fetch the URL as an image and
//     return a ContentBlockImage, else falls
//     back to a simple link.
//   - AttachmentModePageContent / AttachmentModeText:
//     runs the URL through the "page pipeline"
//     (HTML/text/PDF/image handling) and returns
//     a text or file/image block as appropriate.
//   - AttachmentModeFile:          fetches the raw bytes and returns a
//     ContentBlockFile, else falls back to a link.
//   - Any other modes (PR diff/page, commit diff/page, not readable, etc.):
//     safest fallback is a link-only block.
//
// This function assumes that Attachment, AttachmentMode, ContentBlock and
// ContentBlock* constants are defined elsewhere in this package.
func buildBlocksForURL(ctx context.Context, att *Attachment, mode AttachmentMode) (*ContentBlock, error) {
	if att == nil || att.URLRef == nil {
		return nil, errors.New("nil attachment or nil URLRef")
	}

	rawURL := strings.TrimSpace(att.URLRef.URL)
	if rawURL == "" {
		return nil, errors.New("got invalid url")
	}

	switch mode {
	case AttachmentModeLinkOnly:
		// Minimal representation: just the URL (optionally with a label).
		return buildLinkOnlyContentBlock(att), nil

	case AttachmentModeImage:
		// Try to download the URL as an image.
		data, contentType, err := fetchURLBytes(ctx, rawURL, maxAttachmentFetchBytes)
		if err != nil {
			slog.Warn("failed to fetch image url attachment, falling back to link",
				"url", rawURL, "err", err)
			return buildLinkOnlyContentBlock(att), nil
		}
		if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
			// Not actually an image, fall back to a link.
			return buildLinkOnlyContentBlock(att), nil
		}

		base64Data := base64.StdEncoding.EncodeToString(data)
		fname := filenameFromURL(rawURL, contentType)
		return &ContentBlock{
			Kind:       ContentBlockImage,
			Base64Data: &base64Data,
			MIMEType:   &contentType,
			FileName:   &fname,
		}, nil

	case AttachmentModePageContent, AttachmentModeText:
		// New pipeline: image/pdf → html/text → link.
		return buildBlocksForURLPage(ctx, att)

	case AttachmentModeFile:
		// Fetch raw bytes and expose them as a downloadable file.
		data, contentType, err := fetchURLBytes(ctx, rawURL, maxAttachmentFetchBytes)
		if err != nil {
			slog.Warn("failed to fetch url file, falling back to link",
				"url", rawURL, "err", err)
			return buildLinkOnlyContentBlock(att), nil
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		base64Data := base64.StdEncoding.EncodeToString(data)
		fname := filenameFromURL(rawURL, contentType)
		return &ContentBlock{
			Kind:       ContentBlockFile,
			Base64Data: &base64Data,
			MIMEType:   &contentType,
			FileName:   &fname,
		}, nil

	case AttachmentModeNotReadable,
		AttachmentModePRDiff,
		AttachmentModePRPage,
		AttachmentModeCommitDiff,
		AttachmentModeCommitPage:
		// Unknown or special mode: safest fallback is link-only.
		return buildLinkOnlyContentBlock(att), nil

	default:
		return nil, errors.New("unknown attachment mode")
	}
}

// buildBlocksForURLPage handles web pages or unknown content-types using a
// multi-step pipeline:
//
//  1. If the URL is an image (by Content-Type), it is fetched as an image
//     block (same behaviour as AttachmentModeImage).
//  2. If the URL is a PDF, it is fetched as a file block.
//  3. If the URL is plain text (non-HTML), it is fetched as text and sent
//     directly with a small prefix.
//  4. Otherwise, we assume HTML/unknown and use the Trafilatura +
//     HTML→Markdown pipeline via ExtractReadableMarkdownFromURL.
//     - If the page is too large (ErrPageTooLarge), we fall back to
//     a link-only representation to avoid blowing up context size.
//     - If extraction fails for any other reason, we fall back to
//     fetching raw text (up to maxPageContentBytes); if that fails
//     also, we fall back to link-only.
//
// The result is always a single ContentBlock.
func buildBlocksForURLPage(ctx context.Context, att *Attachment) (*ContentBlock, error) {
	if att == nil || att.URLRef == nil {
		return nil, errors.New("nil attachment or nil URLRef")
	}

	rawURL := strings.TrimSpace(att.URLRef.URL)
	if rawURL == "" {
		return nil, errors.New("got invalid url")
	}

	// First, peek at the Content-Type so we can distinguish images/PDFs
	// (which should be treated as binary attachments) from HTML/text pages.
	contentType, err := peekURLContentType(ctx, rawURL)
	if err != nil {
		slog.Warn("failed to HEAD url, assuming HTML and trying extractor",
			"url", rawURL, "err", err)
		// We'll treat it as HTML/unknown below and let the extractor deal with it.
	}

	lowerCT := strings.ToLower(contentType)
	lowerURL := strings.ToLower(rawURL)

	// 1) Images: encode as Base64 image block (same behaviour as in image mode).
	if strings.HasPrefix(lowerCT, "image/") {
		data, ct, err := fetchURLBytes(ctx, rawURL, maxAttachmentFetchBytes)
		if err != nil {
			slog.Warn("failed to fetch image url attachment, falling back to link",
				"url", rawURL, "err", err)
			return buildLinkOnlyContentBlock(att), nil
		}
		if ct == "" {
			ct = contentType
		}
		if ct == "" {
			ct = "image/*"
		}

		base64Data := base64.StdEncoding.EncodeToString(data)
		fname := filenameFromURL(rawURL, ct)
		return &ContentBlock{
			Kind:       ContentBlockImage,
			Base64Data: &base64Data,
			MIMEType:   &ct,
			FileName:   &fname,
		}, nil
	}

	// 2) PDF: encode as Base64 file block.
	if strings.HasPrefix(lowerCT, "application/pdf") || strings.HasSuffix(lowerURL, ".pdf") {
		data, ct, err := fetchURLBytes(ctx, rawURL, maxAttachmentFetchBytes)
		if err != nil {
			slog.Warn("failed to fetch pdf url attachment, falling back to link",
				"url", rawURL, "err", err)
			return buildLinkOnlyContentBlock(att), nil
		}
		if ct == "" {
			ct = "application/pdf"
		}

		base64Data := base64.StdEncoding.EncodeToString(data)
		fname := filenameFromURL(rawURL, ct)
		return &ContentBlock{
			Kind:       ContentBlockFile,
			Base64Data: &base64Data,
			MIMEType:   &ct,
			FileName:   &fname,
		}, nil
	}

	// 3) Plain text (non-HTML): fetch as text and send directly.
	//    This preserves the previous behaviour for text/plain, etc.
	if isPlainTextContentType(lowerCT) {
		data, _, err := fetchURLBytes(ctx, rawURL, maxPageContentBytes)
		if err != nil {
			slog.Warn("failed to fetch text url attachment, falling back to link",
				"url", rawURL, "err", err)
			return buildLinkOnlyContentBlock(att), nil
		}
		prefix := fmt.Sprintf("[Content from %s]\n\n", rawURL)
		txt := prefix + string(data)
		return &ContentBlock{
			Kind: ContentBlockText,
			Text: &txt,
		}, nil
	}

	// 4) HTML / unknown: use the Trafilatura + HTML→Markdown pipeline.
	markdown, err := ExtractReadableMarkdownFromURL(ctx, rawURL)
	if err != nil {
		if errors.Is(err, ErrPageTooLarge) {
			// For very large pages, do NOT pull content into the LLM context.
			slog.Warn("url page too large, falling back to link-only mode",
				"url", rawURL, "err", err)
			return buildLinkOnlyContentBlock(att), nil
		}

		slog.Warn("failed to extract readable content from url, falling back to raw fetch",
			"url", rawURL, "err", err)

		// As a last resort, fetch up to maxPageContentBytes and include raw text.
		data, _, err2 := fetchURLBytes(ctx, rawURL, maxPageContentBytes)
		if err2 != nil {
			slog.Warn("failed to fetch url content for fallback, using link-only",
				"url", rawURL, "err", err2)
			return buildLinkOnlyContentBlock(att), nil
		}
		txt := fmt.Sprintf("[Content from %s]\n\n", rawURL) + string(data)
		return &ContentBlock{
			Kind: ContentBlockText,
			Text: &txt,
		}, nil
	}

	txt := fmt.Sprintf("[Content from %s]\n\n", rawURL) + markdown
	return &ContentBlock{
		Kind: ContentBlockText,
		Text: &txt,
	}, nil
}

// buildLinkOnlyContentBlock returns a simple text block that represents the
// URL as a human-readable link. If the attachment has a label distinct from
// the URL, it is included as "label (url)".
func buildLinkOnlyContentBlock(att *Attachment) *ContentBlock {
	rawURL := ""
	if att != nil && att.URLRef != nil {
		rawURL = strings.TrimSpace(att.URLRef.URL)
	}

	txt := rawURL
	if label := strings.TrimSpace(att.Label); label != "" && label != rawURL {
		txt = fmt.Sprintf("%s (%s)", label, rawURL)
	}

	return &ContentBlock{
		Kind: ContentBlockText,
		Text: &txt,
	}
}

// filenameFromURL attempts to derive a reasonable file name from the URL path.
// If it cannot, it falls back to a generic name based on the MIME type.
func filenameFromURL(rawURL, contentType string) string {
	u, err := url.Parse(rawURL)
	if err == nil {
		name := filepath.Base(u.Path)
		if name != "" && name != "/" && name != "." {
			return name
		}
	}

	contentType = strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.HasPrefix(contentType, "application/pdf"):
		return "document.pdf"
	case strings.HasPrefix(contentType, "image/"):
		return "image"
	default:
		return "download"
	}
}
