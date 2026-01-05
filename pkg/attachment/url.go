package attachment

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
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
func (ref *URLRef) PopulateRef(replaceOrig bool) error {
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
	if strings.TrimSpace(ref.OrigNormalized) == "" || replaceOrig {
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

// BuildContentBlock builds a ContentBlock representation for a URL-based
// attachment, depending on the desired AttachmentContentBlockMode.
//
// The behaviour is:
//   - AttachmentContentBlockModeLinkOnly:      always returns a simple text link.
//   - AttachmentContentBlockModeImage:         tries to fetch the URL as an image and
//     return a ContentBlockImage, else falls
//     back to a simple link.
//   - AttachmentContentBlockModePageContent / AttachmentContentBlockModeText:
//     runs the URL through the "page pipeline"
//     (HTML/text/PDF/image handling) and returns
//     a text or file/image block as appropriate.
//   - AttachmentContentBlockModeFile:          fetches the raw bytes and returns a
//     ContentBlockFile, else falls back to a link.
//   - Any other modes (PR diff/page, commit diff/page, not readable, etc.):
//     safest fallback is a link-only block.
//
// This function assumes that Attachment, AttachmentContentBlockMode, ContentBlock and
// ContentBlock* constants are defined elsewhere in this package.
func (ref *URLRef) BuildContentBlock(
	ctx context.Context,
	attachmentContentBlockMode AttachmentContentBlockMode,
	onlyIfTextKind bool,
) (*ContentBlock, error) {
	rawURL := strings.TrimSpace(ref.URL)
	if rawURL == "" {
		return nil, errors.New("got invalid url")
	}

	switch attachmentContentBlockMode {

	case AttachmentContentBlockModeImage:
		if onlyIfTextKind {
			return nil, ErrNonTextContentBlock
		}
		return ref.buildImageBlockFromURL(ctx)

	case AttachmentContentBlockModeFile:
		if onlyIfTextKind {
			return nil, ErrNonTextContentBlock
		}
		return ref.buildFileBlockFromURL(ctx)

	case AttachmentContentBlockModeLinkOnly:
		// Minimal representation: just the URL (optionally with a label).
		return ref.buildLinkOnlyContentBlock(), nil

	case AttachmentContentBlockModePageContent, AttachmentContentBlockModeText:
		// New pipeline: image/pdf → html/text → link.
		return ref.buildBlocksForURLPage(ctx, onlyIfTextKind)

	case AttachmentContentBlockModeNotReadable,
		AttachmentContentBlockModePRDiff,
		AttachmentContentBlockModePRPage,
		AttachmentContentBlockModeCommitDiff,
		AttachmentContentBlockModeCommitPage:
		// Unknown or special mode: safest fallback is link-only.
		return ref.buildLinkOnlyContentBlock(), nil

	default:
		return nil, errors.New("unknown attachment mode")
	}
}

// buildBlocksForURLPage handles web pages or unknown content-types using a
// multi-step pipeline:
//
//  1. If the URL is an image (by Content-Type), it is fetched as an image
//     block (same behaviour as AttachmentContentBlockModeImage).
//  2. If the URL is a PDF, we try to extract text into a text block; on
//     failure we fall back to a file block (or link-only when text-only
//     was explicitly requested).
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
func (ref *URLRef) buildBlocksForURLPage(ctx context.Context, onlyIfTextKind bool) (*ContentBlock, error) {
	rawURL := strings.TrimSpace(ref.URL)
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

	// 1) Images: encode as Base64 image block (same behavior as in image mode).
	if strings.HasPrefix(lowerCT, "image/") {
		if onlyIfTextKind {
			return nil, ErrNonTextContentBlock
		}
		return ref.buildImageBlockFromURL(ctx)
	}

	// 2) PDF: prefer text extraction; fall back to file/link.
	if strings.HasPrefix(lowerCT, "application/pdf") || strings.HasSuffix(lowerURL, ".pdf") {
		cb, err := ref.buildPDFTextOrFileBlockFromURL(ctx)
		if err != nil {
			slog.Warn("failed to process pdf url, falling back to link-only",
				"url", rawURL, "err", err)

			return ref.buildLinkOnlyContentBlock(), nil
		}

		if onlyIfTextKind && cb.Kind != ContentBlockText {
			// We managed to download the PDF but could not extract text
			// (so the helper fell back to a file block). Don't return
			// non-text content when text-only was explicitly requested.
			return nil, ErrNonTextContentBlock
		}

		return cb, nil
	}

	// 3) Plain text (non-HTML): fetch as text and send directly.
	//    This preserves the previous behaviour for text/plain, etc.
	if isPlainTextContentType(lowerCT) {
		data, _, err := fetchURLBytes(ctx, rawURL, maxPageContentBytes)
		if err != nil {
			slog.Warn("failed to fetch text url attachment, falling back to link",
				"url", rawURL, "err", err)
			return ref.buildLinkOnlyContentBlock(), nil
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
			return ref.buildLinkOnlyContentBlock(), nil
		}

		slog.Warn("failed to extract readable content from url, falling back to raw fetch",
			"url", rawURL, "err", err)

		// As a last resort, fetch up to maxPageContentBytes and include raw text.
		data, _, err2 := fetchURLBytes(ctx, rawURL, maxPageContentBytes)
		if err2 != nil {
			slog.Warn("failed to fetch url content for fallback, using link-only",
				"url", rawURL, "err", err2)
			return ref.buildLinkOnlyContentBlock(), nil
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

// buildPDFTextOrFileBlockFromURL fetches a PDF from a URL and tries to extract
// its text content. On success, it returns a text ContentBlock; on failure or
// panic it falls back to a base64-encoded file ContentBlock.
//
// This mirrors buildPDFTextOrFileBlock (local file) but works with URL bytes.
func (ref *URLRef) buildPDFTextOrFileBlockFromURL(ctx context.Context) (*ContentBlock, error) {
	rawURL := strings.TrimSpace(ref.URL)
	if rawURL == "" {
		return nil, errors.New("got invalid url")
	}
	// Reuse the same safety limit as other attachment fetches.
	data, contentType, err := fetchURLBytes(ctx, rawURL, maxAttachmentFetchBytes)
	if err != nil {
		return nil, err
	}

	text, err := extractPDFTextFromBytesSafe(data, maxAttachmentFetchBytes)
	if err == nil && text != "" {
		return &ContentBlock{
			Kind: ContentBlockText,
			Text: &text,
		}, nil
	}

	// Fallback: send as file attachment.
	if err != nil {
		slog.Warn("pdf text extraction from url failed; falling back to base64 attachment",
			"url", rawURL, "err", err)
	}

	if contentType == "" {
		contentType = string(fileutil.MIMEApplicationPDF)
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	fname := filenameFromURL(rawURL, contentType)
	return &ContentBlock{
		Kind:       ContentBlockFile,
		Base64Data: &base64Data,
		MIMEType:   &contentType,
		FileName:   &fname,
	}, nil
}

func (ref *URLRef) buildImageBlockFromURL(ctx context.Context) (*ContentBlock, error) {
	rawURL := strings.TrimSpace(ref.URL)
	if rawURL == "" {
		return nil, errors.New("got invalid url")
	}
	// Try to download the URL as an image.
	data, contentType, err := fetchURLBytes(ctx, rawURL, maxAttachmentFetchBytes)
	if err != nil {
		slog.Warn("failed to fetch image url attachment, falling back to link",
			"url", rawURL, "err", err)
		return ref.buildLinkOnlyContentBlock(), nil
	}
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		// Not actually an image, fall back to a link.
		return ref.buildLinkOnlyContentBlock(), nil
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	fname := filenameFromURL(rawURL, contentType)
	return &ContentBlock{
		Kind:       ContentBlockImage,
		Base64Data: &base64Data,
		MIMEType:   &contentType,
		FileName:   &fname,
	}, nil
}

// Fetch raw bytes and expose them as a downloadable file.
func (ref *URLRef) buildFileBlockFromURL(ctx context.Context) (*ContentBlock, error) {
	rawURL := strings.TrimSpace(ref.URL)
	if rawURL == "" {
		return nil, errors.New("got invalid url")
	}
	data, contentType, err := fetchURLBytes(ctx, rawURL, maxAttachmentFetchBytes)
	if err != nil {
		slog.Warn("failed to fetch url file, falling back to link",
			"url", rawURL, "err", err)
		return ref.buildLinkOnlyContentBlock(), nil
	}
	if contentType == "" {
		if strings.HasSuffix(strings.ToLower(rawURL), ".pdf") {
			contentType = string(fileutil.MIMEApplicationPDF)
		} else {
			contentType = string(fileutil.MIMEApplicationOctetStream)
		}
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	fname := filenameFromURL(rawURL, contentType)
	return &ContentBlock{
		Kind:       ContentBlockFile,
		Base64Data: &base64Data,
		MIMEType:   &contentType,
		FileName:   &fname,
	}, nil
}

// buildLinkOnlyContentBlock returns a simple text block that represents the
// URL as a human-readable link. If the attachment has a label distinct from
// the URL, it is included as "label (url)".
func (ref *URLRef) buildLinkOnlyContentBlock() *ContentBlock {
	return &ContentBlock{
		Kind: ContentBlockText,
		Text: &ref.URL,
	}
}
