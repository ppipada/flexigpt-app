package fileutil

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strings"
)

type ImageInfo struct {
	PathInfo

	Width    int      `json:"width,omitempty"`
	Height   int      `json:"height,omitempty"`
	Format   string   `json:"format,omitempty"`   // e.g. "jpeg", "png"
	MIMEType MIMEType `json:"mimeType,omitempty"` // e.g. "image/jpeg"
}

// ImageData holds metadata (and optionally content) for an image file.
type ImageData struct {
	ImageInfo

	Base64Data string `json:"base64Data,omitempty"` // optional, if requested
}

// ReadImage inspects an image file and returns its intrinsic metadata.
// If includeBase64 is true, Base64Data will contain the base64-encoded file
// contents. If the file does not exist, Exists == false and err == nil.
// Returns an error if the path is empty, a directory, or not a supported image.
func ReadImage(
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

	out := &ImageData{}
	out.Path = pathInfo.Path
	out.Name = pathInfo.Name
	out.Exists = pathInfo.Exists
	out.IsDir = pathInfo.IsDir
	out.Size = pathInfo.Size
	out.ModTime = pathInfo.ModTime

	if !out.Exists {
		// Not an error: just report non-existence.
		return out, nil
	}
	if out.IsDir {
		return nil, errors.New("path points to a directory, expected file")
	}

	// We need to decode the image config; if includeBase64 is true, we can
	// read the whole file once and reuse that data for both config and base64.
	if includeBase64Data {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		reader := bytes.NewReader(data)
		err = decodeImageConfig(out, reader)
		if err != nil {
			return nil, err
		}
		out.Base64Data = base64.StdEncoding.EncodeToString(data)
		return out, nil
	}

	// No base64 requested: just open and decode config.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = decodeImageConfig(out, f)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func decodeImageConfig(info *ImageData, reader io.Reader) error {
	cfg, fmtName, err := image.DecodeConfig(reader)
	if err != nil {
		return err
	}

	info.Width = cfg.Width
	info.Height = cfg.Height
	info.Format = fmtName
	m, err := MIMEFromExtensionString(fmtName)
	if err != nil {
		return errors.New("invalid image format")
	}
	if mode, ok := MIMETypeToExtensionMode[m]; !ok || mode != ExtensionModeImage {
		return errors.New("invalid image format")
	}
	info.MIMEType = m
	return nil
}
