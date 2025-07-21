package bundleitemutils

import (
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
)

// bundlePartitionAttr is a type for bundle partition attributes.
type bundlePartitionAttr string

// BundlePartitionProvider implements dirstore.PartitionProvider for bundle directories.
type BundlePartitionProvider struct{}

// GetPartitionDir returns the partition directory for a given file key.
// The directory is provided via the XAttr field.
func (b *BundlePartitionProvider) GetPartitionDir(key dirstore.FileKey) (string, error) {
	if key.FileName == "" {
		return "", ErrInvalidFilename
	}
	dir, ok := key.XAttr.(bundlePartitionAttr)
	if !ok || dir == "" {
		return "", ErrBundleAttributeMissing
	}
	return string(dir), nil
}

// ListPartitions lists all bundle directories under baseDir.
func (b *BundlePartitionProvider) ListPartitions(
	baseDir, sortOrder, pageToken string,
	pageSize int,
) (dirs []string, nextPageToken string, err error) {
	return dirstore.ListDirs(baseDir, sortOrder, pageToken, pageSize)
}

func GetBundlePartitionFileKey(filename, dirname string) dirstore.FileKey {
	return dirstore.FileKey{FileName: filename, XAttr: bundlePartitionAttr(dirname)}
}
