package attachment

import (
	"errors"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
)

// FileRef carries metadata for file attachments.
type FileRef struct {
	Path      string     `json:"path"`
	Exists    bool       `json:"exists,omitempty"`
	SizeBytes int64      `json:"sizeBytes,omitempty"`
	ModTime   *time.Time `json:"modTime,omitempty"`
}

func (ref *FileRef) PopulateRef() error {
	if ref == nil {
		return errors.New("file attachment missing ref")
	}
	path := strings.TrimSpace(ref.Path)
	if path == "" {
		return errors.New("file attachment missing path")
	}
	pathInfo, err := fileutil.StatPath(path)
	if err != nil {
		return err
	}

	ref.Path = path
	ref.Exists = pathInfo.Exists
	if pathInfo.Exists {
		ref.SizeBytes = pathInfo.SizeBytes
		ref.ModTime = pathInfo.ModTime
	} else {
		ref.SizeBytes = 0
		ref.ModTime = nil
	}
	return nil
}
