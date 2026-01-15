package attachment

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/flexigpt/flexigpt-app/internal/fileutil"
)

// ImageRef carries metadata for image attachments.
type ImageRef struct {
	fileutil.ImageInfo

	// Original snapshot (for change detection across turns).
	OrigPath    string    `json:"origPath"`
	OrigSize    int64     `json:"origSize"`
	OrigModTime time.Time `json:"origModTime"`
}

func (ref *ImageRef) PopulateRef(replaceOrig bool) error {
	if ref == nil {
		return errors.New("image attachment missing ref")
	}
	path := strings.TrimSpace(ref.Path)
	if path == "" {
		return errors.New("image attachment missing path")
	}
	info, err := fileutil.ReadImage(path, false)
	if err != nil {
		return err
	}

	if strings.TrimSpace(ref.OrigPath) == "" || replaceOrig {
		ref.OrigPath = info.Path
		ref.OrigSize = info.Size
		ref.OrigModTime = *info.ModTime
	}

	ref.Path = info.Path
	ref.Name = info.Name
	ref.Exists = info.Exists
	ref.IsDir = info.IsDir
	ref.Size = info.Size
	ref.ModTime = info.ModTime
	ref.Width = info.Width
	ref.Height = info.Height
	ref.Format = info.Format
	ref.MIMEType = info.MIMEType
	return nil
}

func (ref *ImageRef) IsModified() bool {
	if ref == nil {
		return false
	}
	if strings.TrimSpace(ref.OrigPath) == "" {
		return false
	}
	if !ref.Exists {
		return true
	}
	if ref.Path != ref.OrigPath {
		return true
	}
	if ref.Size != ref.OrigSize {
		return true
	}
	if !ref.ModTime.Equal(ref.OrigModTime) {
		return true
	}
	return false
}

func (ref *ImageRef) BuildContentBlock() (*ContentBlock, error) {
	return buildImageBlockFromLocal(ref.Path)
}

func buildImageBlockFromLocal(path string) (*ContentBlock, error) {
	info, err := fileutil.ReadImage(path, true)
	if err != nil {
		return nil, err
	}
	mStr := string(info.MIMEType)
	fname := filepath.Base(path)
	return &ContentBlock{
		Kind:       ContentBlockImage,
		Base64Data: &info.Base64Data,
		MIMEType:   &mStr,
		FileName:   &fname,
	}, nil
}
