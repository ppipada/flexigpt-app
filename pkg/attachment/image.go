package attachment

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
)

// ImageRef carries metadata for image attachments.
type ImageRef struct {
	fileutil.ImageInfo
}

func (ref *ImageRef) PopulateRef() error {
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

func buildImageBlockFromLocal(path string) (*ContentBlock, error) {
	info, err := fileutil.ReadImage(path, true)
	if err != nil {
		return &ContentBlock{}, err
	}
	return &ContentBlock{
		Kind:       ContentBlockImage,
		Base64Data: info.Base64Data,
		MIMEType:   string(info.MIMEType),
		FileName:   filepath.Base(path),
	}, nil
}
