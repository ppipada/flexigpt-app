package attachment

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
)

// ImageRef carries metadata for image attachments.
type ImageRef struct {
	Path      string     `json:"path"`
	Exists    bool       `json:"exists,omitempty"`
	Width     int        `json:"width,omitempty"`
	Height    int        `json:"height,omitempty"`
	Format    string     `json:"format,omitempty"`
	SizeBytes int64      `json:"sizeBytes,omitempty"`
	ModTime   *time.Time `json:"modTime,omitempty"`
}

func (ref *ImageRef) PopulateRef() error {
	if ref == nil {
		return errors.New("image attachment missing ref")
	}
	path := strings.TrimSpace(ref.Path)
	if path == "" {
		return errors.New("image attachment missing path")
	}
	info, err := fileutil.ReadImageInfo(path, false)
	if err != nil {
		return err
	}

	ref.Path = path
	ref.Exists = info.Exists
	if info.Exists {
		ref.Width = info.Width
		ref.Height = info.Height
		ref.Format = info.Format
		ref.SizeBytes = info.SizeBytes
		ref.ModTime = info.ModTime
	} else {
		ref.Width = 0
		ref.Height = 0
		ref.Format = ""
		ref.SizeBytes = 0
		ref.ModTime = nil
	}
	return nil
}

func buildImageBlockFromLocal(path string) (*ContentBlock, error) {
	info, err := fileutil.ReadImageInfo(path, true)
	if err != nil {
		return &ContentBlock{}, err
	}
	return &ContentBlock{
		Kind:       ContentBlockImage,
		Base64Data: info.Base64Data,
		MIMEType:   info.MIMEType,
		FileName:   filepath.Base(path),
	}, nil
}
