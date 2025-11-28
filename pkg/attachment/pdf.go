package attachment

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
)

// buildPDFTextOrFileBlock tries to extract PDF text; on failure/panic, falls back to base64 file.
func buildPDFTextOrFileBlock(path string) (*ContentBlock, error) {
	text, err := extractPDFTextSafe(path, maxAttachmentFetchBytes)
	if err == nil && text != "" {
		return &ContentBlock{
			Kind: ContentBlockText,
			Text: text,
		}, nil
	}

	// Fallback: send as file attachment.
	base64Data, err2 := fileutil.ReadFile(path, fileutil.ReadEncodingBinary)
	if err2 != nil {
		if err != nil {
			return nil, fmt.Errorf(
				"pdf text extraction failed (%w); fallback to base64 also failed: %w",
				err, err2,
			)
		}
		return nil, err2
	}

	if err != nil {
		slog.Warn("pdf text extraction failed; falling back to base64 attachment",
			"path", path, "err", err)
	}
	return &ContentBlock{
		Kind:       ContentBlockFile,
		Base64Data: base64Data,
		MIMEType:   "application/pdf",
		FileName:   filepath.Base(path),
	}, nil
}

// extractPDFTextSafe extracts text from a local PDF with a byte limit and panic recovery.
func extractPDFTextSafe(path string, maxBytes int) (text string, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Warn("panic during PDF text extraction", "path", path, "panic", r)
			err = fmt.Errorf("panic during PDF text extraction: %v", r)
		}
	}()

	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	limited := &io.LimitedReader{
		R: reader,
		N: int64(maxBytes),
	}
	if _, err := io.Copy(&buf, limited); err != nil {
		return "", err
	}
	text = strings.TrimSpace(buf.String())
	if text == "" {
		return "", errors.New("empty PDF text after extraction")
	}
	return text, nil
}
