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

	html2md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/go-shiori/dom"
	"github.com/markusmobius/go-trafilatura"
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
	converter := html2md.NewConverter("", true, nil)

	markdown, err := converter.ConvertString(cleanedHTML)
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

	cleaned := strings.TrimSpace(dom.OuterHTML(doc))
	if cleaned == "" {
		return "", ErrNoContentExtracted
	}

	return cleaned, nil
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

// peekURLContentType sends a HEAD request and returns the Content-Type header.
// Returns an error on network / HTTP failure and an empty string if the header
// is missing. This is used to quickly determine how to treat a URL without
// downloading the entire body.
func peekURLContentType(ctx context.Context, rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("empty url")
	}
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return "", fmt.Errorf("invalid url %q: %w", rawURL, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("http %d when fetching headers for %s", resp.StatusCode, rawURL)
	}

	return resp.Header.Get("Content-Type"), nil
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
