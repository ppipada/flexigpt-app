package fileutil

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"time"
)

const DefaultImageMIME = "image/png"

// ImageData holds metadata (and optionally content) for an image file.
type ImageData struct {
	Path      string     `json:"path"`
	Exists    bool       `json:"exists"`
	Width     int        `json:"width,omitempty"`
	Height    int        `json:"height,omitempty"`
	Format    string     `json:"format,omitempty"`   // e.g. "jpeg", "png"
	MIMEType  string     `json:"mimeType,omitempty"` // e.g. "image/jpeg"
	SizeBytes int64      `json:"sizeBytes,omitempty"`
	ModTime   *time.Time `json:"modTime,omitempty"` // UTC

	Base64Data string `json:"base64Data,omitempty"` // optional, if requested
}

// ReadImageInfo inspects an image file and returns its intrinsic metadata.
// If includeBase64 is true, Base64Data will contain the base64-encoded file
// contents. If the file does not exist, Exists == false and err == nil.
// Returns an error if the path is empty, a directory, or not a supported image.
func ReadImageInfo(
	path string,
	includeBase64Data bool,
) (*ImageData, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("path is required")
	}

	pathInfo, err := StatPath(path)
	if err != nil {
		return nil, err
	}

	out := &ImageData{
		Path:      path,
		Exists:    pathInfo.Exists,
		SizeBytes: pathInfo.SizeBytes,
		ModTime:   pathInfo.ModTime,
	}

	if !pathInfo.Exists {
		// Not an error: just report non-existence.
		return out, nil
	}
	if pathInfo.IsDir {
		return nil, errors.New("path points to a directory, expected file")
	}

	// We need to decode the image config; if includeBase64 is true, we can
	// read the whole file once and reuse that data for both config and base64.
	if includeBase64Data {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		cfg, fmtName, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		out.Width = cfg.Width
		out.Height = cfg.Height
		out.Format = fmtName
		out.MIMEType = imageFormatToMIME(fmtName)

		out.Base64Data = base64.StdEncoding.EncodeToString(data)

		return out, nil
	}

	// No base64 requested: just open and decode config.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg, fmtName, err := image.DecodeConfig(f)
	if err != nil {
		return nil, err
	}

	out.Width = cfg.Width
	out.Height = cfg.Height
	out.Format = fmtName
	out.MIMEType = imageFormatToMIME(fmtName)
	return out, nil
}

// ImageFormatToMIME converts an image "format" (e.g. "jpeg", "png")
// into a best-effort MIME type.
func imageFormatToMIME(format string) string {
	f := strings.ToLower(strings.TrimSpace(format))
	switch f {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "bmp":
		return "image/bmp"
	case "":
		// Keep your previous behavior: default to PNG if unknown/empty.
		return DefaultImageMIME
	default:
		// Fallback; most renderers can still display many uncommon types.
		return "application/octet-stream"
	}
}
