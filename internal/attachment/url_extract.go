package attachment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	html2md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/markusmobius/go-trafilatura"
	"github.com/ppipada/flexigpt-app/internal/fileutil"
	"golang.org/x/net/html"
)

// NOTE: These limits are safety rails to avoid pulling huge resources
// into memory or into the LLM context. Tune them to your environment.
const (
	// Max number of bytes we will download and process as "page content"
	// for LLM consumption. If a page is larger than this, we will return
	// ErrPageTooLarge and fall back to link-only at a higher layer.
	maxPageContentBytes = 16 << 20 // 16 MiB
	defaultHTTPTimeout  = 30 * time.Second
)

var (

	// ErrResponseTooLarge is used internally by fetchURLBytes to signal
	// that the remote resource exceeded the configured maxBytes limit.
	ErrResponseTooLarge = errors.New("remote resource larger than the configured limit")

	// ErrPageTooLarge is returned by ExtractReadableMarkdownFromURL when
	// the HTML page exceeds maxPageContentBytes.
	ErrPageTooLarge = errors.New("page content too large")

	// ErrNoContentExtracted is returned when the extractor produces no
	// readable content at all (e.g., empty body, or extraction failure).
	ErrNoContentExtracted = errors.New("no readable content extracted")
)

var sharedHTTPClient = &http.Client{
	Timeout: defaultHTTPTimeout,
	Transport: func() *http.Transport {
		// DefaultTransport is a *http.Transport; clone it so we can tweak it if needed.
		if t, ok := http.DefaultTransport.(*http.Transport); ok {
			return t.Clone()
		}
		// Fallback: new one with reasonable defaults.
		return &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
		}
	}(),
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

// ExtractReadableMarkdownFromURL downloads a web page, runs it through
// go-trafilatura to obtain the main readable content, converts that to
// Markdown, and returns the Markdown text.
//
// It enforces a hard size limit (maxPageContentBytes). If the remote server
// is larger than that, ErrPageTooLarge is returned. Only the HTML of the
// page itself is downloaded; sub-resources (images, scripts, etc.) are
// not fetched.
//
// This function is conservative: it only attempts extraction on content
// types that look like HTML or generic text.
func ExtractReadableMarkdownFromURL(ctx context.Context, rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("empty url")
	}
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return "", fmt.Errorf("invalid url %q: %w", rawURL, err)
	}

	// Fetch the page HTML with a strict byte limit.
	data, contentType, err := fetchURLBytes(ctx, rawURL, maxPageContentBytes)
	if err != nil {
		if errors.Is(err, ErrResponseTooLarge) {
			// Promote to a page-specific sentinel that callers can treat specially.
			return "", ErrPageTooLarge
		}
		return "", fmt.Errorf("fetching url: %w", err)
	}

	if !isProbablyHTMLOrText(contentType, rawURL) {
		return "", fmt.Errorf("unsupported content type %q for page extraction", contentType)
	}

	if len(data) == 0 {
		return "", ErrNoContentExtracted
	}

	// Use go-trafilatura to extract the main content as cleaned HTML.
	cleanedHTML, err := extractMainContentHTMLWithTrafilatura(ctx, rawURL, data)
	if err != nil {
		return "", fmt.Errorf("trafilatura extraction failed: %w", err)
	}
	cleanedHTML = strings.TrimSpace(cleanedHTML)
	if cleanedHTML == "" {
		return "", ErrNoContentExtracted
	}

	// Convert cleaned HTML to Markdown for LLM-friendly structured text.
	markdown, err := html2md.ConvertString(cleanedHTML)
	if err != nil {
		return "", fmt.Errorf("html-to-markdown conversion failed: %w", err)
	}

	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return "", ErrNoContentExtracted
	}

	return markdown, nil
}

// extractMainContentHTMLWithTrafilatura is a thin adapter around the
// github.com/markusmobius/go-trafilatura library. It takes the raw HTML
// bytes we already downloaded and returns cleaned HTML representing the
// main readable content.
//
// It uses Trafilatura's high-level Extract + CreateReadableDocument API,
// which exposes a single main-document DOM node that we then serialize
// back to HTML.
func extractMainContentHTMLWithTrafilatura(_ context.Context, rawURL string, htmlBytes []byte) (string, error) {
	// Parse the URL if possible, but don't fail extraction if it is invalid.
	var parsedURL *url.URL
	if u, err := url.Parse(rawURL); err == nil {
		parsedURL = u
	}

	opts := trafilatura.Options{
		// We care primarily about text for LLM consumption; images add
		// little value and increase context size.
		IncludeImages: false,

		// Some heuristics (e.g. canonical URL detection) may make use of
		// the original page URL if available.
		OriginalURL: parsedURL,

		// Enable Trafilatura's internal fallbacks (e.g., Readability, DOM
		// distiller) when dealing with tricky pages.
		EnableFallback: true,
	}

	reader := bytes.NewReader(htmlBytes)

	result, err := trafilatura.Extract(reader, opts)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", ErrNoContentExtracted
	}

	// Build a minimal, readable document around the extracted content.
	doc := trafilatura.CreateReadableDocument(result)
	if doc == nil {
		return "", ErrNoContentExtracted
	}

	cleaned := strings.TrimSpace(getHTMLString(doc))
	if cleaned == "" {
		return "", ErrNoContentExtracted
	}

	return cleaned, nil
}

// getHTMLString get an escaped HTML serialization of the element and its descendants.
func getHTMLString(node *html.Node) string {
	if node == nil {
		return ""
	}

	var buffer bytes.Buffer
	err := html.Render(&buffer, node)
	if err != nil {
		return ""
	}

	return buffer.String()
}

// fetchURLBytes downloads up to maxBytes bytes from rawURL.
//   - If maxBytes <= 0, it will read the full body (subject to httpClient).
//   - If the remote server advertises a larger Content-Length than maxBytes,
//     or if more than maxBytes bytes are actually read, ErrResponseTooLarge
//     is returned.
//   - On success, it returns the body bytes (at most maxBytes) and the
//     Content-Type header value (which may be empty).
//
// This helper is used both for page extraction and for binary attachments.
func fetchURLBytes(ctx context.Context, rawURL string, maxBytes int) (data []byte, contentType string, err error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, "", errors.New("empty url")
	}
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return nil, "", fmt.Errorf("invalid url %q: %w", rawURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, "", err
	}

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("http %d when fetching %s", resp.StatusCode, rawURL)
	}

	contentType = resp.Header.Get("Content-Type")

	// If we have a size limit and the server advertises a larger body,
	// abort early to avoid reading it at all.
	if maxBytes > 0 && resp.ContentLength > 0 && resp.ContentLength > int64(maxBytes) {
		return nil, contentType, ErrResponseTooLarge
	}

	var reader io.Reader = resp.Body
	if maxBytes > 0 {
		// Use maxBytes+1 so that we can detect if the server tries to send
		// more than we are willing to accept.
		reader = io.LimitReader(resp.Body, int64(maxBytes)+1)
	}

	data, err = io.ReadAll(reader)
	if err != nil {
		return nil, contentType, err
	}

	if maxBytes > 0 && len(data) > maxBytes {
		return nil, contentType, ErrResponseTooLarge
	}

	return data, contentType, nil
}

// peekURLContentType tries to determine the URL's content type without downloading the full body.
//
// Strategy:
//  1. HEAD: use Content-Type header if present and status < 400
//  2. If HEAD fails, is blocked (405/403), or returns no Content-Type:
//     GET with Range: bytes=0-511
//     - prefer Content-Type header if present
//     - else sniff the first bytes via http.DetectContentType.
//
// It returns a *normalized* mime type (lowercased, params stripped), e.g. "text/html", "image/png", "application/pdf".
func peekURLContentType(ctx context.Context, rawURL string) (string, error) {
	const maxPeekBytes = 512

	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("empty url")
	}

	// Validate URL. ParseRequestURI is OK for basic validation, but for URL attachments
	// you typically want absolute URLs; enforce that here.
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid url %q: %w", rawURL, err)
	}
	if !parsed.IsAbs() {
		return "", fmt.Errorf("url %q must be absolute", rawURL)
	}

	var headErr error

	// 1) HEAD attempt.
	{
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, http.NoBody)
		if err != nil {
			return "", err
		}

		resp, err := sharedHTTPClient.Do(req)
		if err != nil {
			headErr = err
		} else {
			_ = resp.Body.Close()

			// Some servers return useful headers even on errors, but generally treat >= 400 as failure.
			if resp.StatusCode < 400 {
				if ct := normalizeContentType(resp.Header.Get("Content-Type")); ct != "" {
					return ct, nil
				}
			} else {
				headErr = fmt.Errorf("http %d when fetching headers for %s", resp.StatusCode, rawURL)
			}
		}
	}

	// 2) GET Range fallback + sniff (reads at most maxPeekBytes).
	{
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			if headErr != nil {
				return "", fmt.Errorf("head failed: %w; range-get build failed: %w", headErr, err)
			}
			return "", err
		}

		req.Header.Set("Range", fmt.Sprintf("bytes=0-%d", maxPeekBytes-1))
		// Encourage uncompressed for more consistent sniffing with Range.
		req.Header.Set("Accept-Encoding", "identity")

		resp, err := sharedHTTPClient.Do(req)
		if err != nil {
			if headErr != nil {
				return "", fmt.Errorf("head failed: %w; range-get failed: %w", headErr, err)
			}
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			if headErr != nil {
				return "", fmt.Errorf("head failed: %w; range-get status %d for %s", headErr, resp.StatusCode, rawURL)
			}
			return "", fmt.Errorf("http %d when range-fetching %s", resp.StatusCode, rawURL)
		}

		// Prefer server-declared Content-Type if present.
		if ct := normalizeContentType(resp.Header.Get("Content-Type")); ct != "" {
			return ct, nil
		}

		// Otherwise sniff from the first bytes.
		buf, _ := io.ReadAll(io.LimitReader(resp.Body, maxPeekBytes))
		if len(buf) > 0 {
			return normalizeContentType(http.DetectContentType(buf)), nil
		}
	}

	// Could not determine, but caller can still proceed with extension heuristics / link-only.
	if headErr != nil {
		return "", fmt.Errorf("unable to determine content-type for %s (head err: %w)", rawURL, headErr)
	}
	return "", fmt.Errorf("unable to determine content-type for %s", rawURL)
}

// isProbablyHTMLOrText checks whether the given Content-Type / URL combination
// is something we can reasonably try to treat as HTML or plain text for
// extraction. It is deliberately conservative.
func isProbablyHTMLOrText(contentType, rawURL string) bool {
	if contentType == "" {
		// Fall back to URL extension heuristic if we have no Content-Type.
		ext := strings.ToLower(filepath.Ext(rawURL))
		switch ext {
		case ".html", ".htm", ".xhtml", "":
			return true
		default:
			return false
		}
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		// If parsing fails, be conservative and bail out.
		return false
	}
	mt := strings.ToLower(mediaType)

	if strings.Contains(mt, "html") {
		return true
	}
	if strings.HasPrefix(mt, "text/") {
		return true
	}
	if mt == "application/xhtml+xml" {
		return true
	}
	return false
}

// isPlainTextContentType reports whether a content type is "plain text-ish"
// but not HTML. This is used to preserve previous behaviour for text/plain
// and similar, bypassing the HTML extractor and feeding the text directly.
func isPlainTextContentType(lowerCT string) bool {
	if lowerCT == "" {
		return false
	}

	if strings.HasPrefix(lowerCT, "text/plain") {
		return true
	}
	if strings.HasPrefix(lowerCT, "text/markdown") {
		return true
	}
	if strings.HasPrefix(lowerCT, "text/x-markdown") {
		return true
	}
	// Add more variants as needed, e.g. text/csv if you want to treat it
	// as raw text rather than run it through Trafilatura.
	return false
}

// inferContentType tries, in order:
//  1. normalized header content-type
//  2. extension-based hints for common types (pdf)
//  3. sniff bytes via http.DetectContentType (first 512 bytes)
//  4. default octet-stream
func inferContentType(headerCT, rawURL string, data []byte) string {
	if mt := normalizeContentType(headerCT); mt != "" {
		return mt
	}
	lowerURL := strings.ToLower(strings.TrimSpace(rawURL))
	if strings.HasSuffix(lowerURL, ".pdf") {
		return string(fileutil.MIMEApplicationPDF)
	}
	if len(data) > 0 {
		return normalizeContentType(http.DetectContentType(data))
	}
	return string(fileutil.MIMEApplicationOctetStream)
}

// normalizeContentType returns a stable, comparable mime type:
// - strips parameters (e.g. "; charset=utf-8")
// - lowercases the media type
// - returns "" if input is empty/whitespace.
func normalizeContentType(ct string) string {
	ct = strings.TrimSpace(ct)
	if ct == "" {
		return ""
	}
	mt, _, err := mime.ParseMediaType(ct)
	if err == nil && strings.TrimSpace(mt) != "" {
		return strings.ToLower(strings.TrimSpace(mt))
	}
	// Fallback: best-effort strip params.
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = ct[:i]
	}
	return strings.ToLower(strings.TrimSpace(ct))
}
