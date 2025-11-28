package attachment

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
)

// URLRef carries metadata for URL-based attachments.
type URLRef struct {
	URL        string `json:"url"`
	Normalized string `json:"normalized,omitempty"`
}

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

	ref.URL = raw
	ref.Normalized = parsed.String()
	return nil
}

func inferDefaultURLMode(rawURL string) AttachmentMode {
	u := strings.ToLower(rawURL)
	for ext, mode := range fileutil.ExtensionToMode {
		if strings.HasSuffix(u, ext) {
			if mode == fileutil.ExtensionModeImage {
				return AttachmentModeImage
			}

			return AttachmentModePageContent
		}
	}

	return AttachmentModePageContent
}

func buildBlocksForURL(ctx context.Context, att *Attachment, mode AttachmentMode) (*ContentBlock, error) {
	rawURL := att.URLRef.URL
	if rawURL == "" {
		return nil, errors.New("got invalid url")
	}

	switch mode {
	case AttachmentModeLinkOnly:
		return buildLinkOnlyContentBlock(att), nil

	case AttachmentModeImage:
		data, contentType, err := fetchURLBytes(ctx, rawURL, maxAttachmentFetchBytes)
		if err != nil {
			slog.Warn("failed to fetch image url attachment, falling back to link",
				"url", rawURL, "err", err)
			return buildLinkOnlyContentBlock(att), nil
		}
		if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
			return buildLinkOnlyContentBlock(att), nil
		}
		base64Data := base64.StdEncoding.EncodeToString(data)
		return &ContentBlock{
			Kind:       ContentBlockImage,
			Base64Data: base64Data,
			MIMEType:   contentType,
			FileName:   filenameFromURL(rawURL, contentType),
		}, nil

	case AttachmentModePageContent, AttachmentModeText:
		// New pipeline: image/pdf → html/text → link.
		return buildBlocksForURLPage(ctx, att)

	case AttachmentModeFile:
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
		return &ContentBlock{
			Kind:       ContentBlockFile,
			Base64Data: base64Data,
			MIMEType:   contentType,
			FileName:   filenameFromURL(rawURL, contentType),
		}, nil

	case AttachmentModeNotReadable,
		AttachmentModePRDiff,
		AttachmentModePRPage,
		AttachmentModeCommitDiff,
		AttachmentModeCommitPage:
		// Unknown mode: safest fallback is link-only.
		return buildLinkOnlyContentBlock(att), nil

	default:
		return nil, errors.New("unknown attachment mode")
	}
}

func buildBlocksForURLPage(ctx context.Context, att *Attachment) (*ContentBlock, error) {
	rawURL := att.URLRef.URL
	if rawURL == "" {
		return nil, errors.New("got invalid url")
	}

	data, contentType, err := fetchURLBytes(ctx, rawURL, maxAttachmentFetchBytes)
	if err != nil {
		slog.Warn("failed to fetch url content, falling back to link",
			"url", rawURL, "err", err)
		return buildLinkOnlyContentBlock(att), nil
	}

	lowerCT := strings.ToLower(contentType)
	lowerURL := strings.ToLower(rawURL)

	// 1) If the mode was chosen as page, then for images we need to encode the image as b64 and send it as "page".
	if strings.HasPrefix(lowerCT, "image/") {
		base64Data := base64.StdEncoding.EncodeToString(data)
		return &ContentBlock{
			Kind:       ContentBlockImage,
			Base64Data: base64Data,
			MIMEType:   contentType,
			FileName:   filenameFromURL(rawURL, contentType),
		}, nil
	}

	// 2) If the mode was chosen as page, then for pdf we need to encode the pdf as b64 and send it as "page".
	if strings.HasPrefix(lowerCT, "application/pdf") || strings.HasSuffix(lowerURL, ".pdf") {
		base64Data := base64.StdEncoding.EncodeToString(data)
		mime := contentType
		if mime == "" {
			mime = "application/pdf"
		}
		return &ContentBlock{
			Kind:       ContentBlockFile,
			Base64Data: base64Data,
			MIMEType:   mime,
			FileName:   filenameFromURL(rawURL, mime),
		}, nil
	}

	// 3) HTML / text?.
	if strings.HasPrefix(lowerCT, "text/") || lowerCT == "" {
		text := string(data)
		prefix := fmt.Sprintf("[Content from %s]\n\n", rawURL)
		return &ContentBlock{
			Kind: ContentBlockText,
			Text: prefix + text,
		}, nil
	}

	// 4) Fallback: link only.
	return buildLinkOnlyContentBlock(att), nil
}

func buildLinkOnlyContentBlock(att *Attachment) *ContentBlock {
	rawURL := att.URLRef.URL
	txt := rawURL
	if label := strings.TrimSpace(att.Label); label != "" && label != rawURL {
		txt = fmt.Sprintf("%s (%s)", label, rawURL)
	}
	return &ContentBlock{
		Kind: ContentBlockText,
		Text: txt,
	}
}

func fetchURLBytes(ctx context.Context, rawURL string, maxBytes int) (data []byte, contentType string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("http %d when fetching %s", resp.StatusCode, rawURL)
	}
	var reader io.Reader = resp.Body
	if maxBytes > 0 {
		reader = io.LimitReader(resp.Body, int64(maxBytes))
	}
	data, err = io.ReadAll(reader)
	if err != nil {
		return nil, "", err
	}
	contentType = resp.Header.Get("Content-Type")
	return data, contentType, nil
}

func filenameFromURL(rawURL, contentType string) string {
	u, err := url.Parse(rawURL)
	if err == nil {
		name := filepath.Base(u.Path)
		if name != "" && name != "/" && name != "." {
			return name
		}
	}
	switch {
	case strings.HasPrefix(strings.ToLower(contentType), "application/pdf"):
		return "document.pdf"
	case strings.HasPrefix(strings.ToLower(contentType), "image/"):
		return "image"
	default:
		return "download"
	}
}
