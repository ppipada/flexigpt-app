package fileutil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractPDFTextSafe extracts text from a local PDF with a byte limit and panic recovery.
func ExtractPDFTextSafe(path string, maxBytes int) (text string, err error) {
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

// ExtractPDFTextFromBytesSafe extracts text from in-memory PDF bytes with a
// byte limit and panic recovery. It mirrors extractPDFTextSafe but is backed
// by an in-memory reader instead of a file on disk.
func ExtractPDFTextFromBytesSafe(data []byte, maxBytes int) (text string, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Warn("panic during PDF text extraction from bytes", "panic", r)
			err = fmt.Errorf("panic during PDF text extraction: %v", r)
		}
	}()

	if len(data) == 0 {
		return "", errors.New("empty PDF data")
	}

	reader := bytes.NewReader(data)
	r, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", err
	}

	plain, err := r.GetPlainText()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	limited := &io.LimitedReader{
		R: plain,
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
