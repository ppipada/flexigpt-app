package store

import (
	"errors"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
)

type bundlePartitionAttr string

type BundlePartitionProvider struct{}

func (b *BundlePartitionProvider) GetPartitionDir(key dirstore.FileKey) (string, error) {
	if key.FileName == "" {
		return "", errors.New("empty filename")
	}

	bundleDir, ok := key.XAttr.(bundlePartitionAttr)
	if !ok || bundleDir == "" {
		return "", errors.New("missing bundle partition attrs")
	}
	return string(bundleDir), nil
}

// ListPartitions returns a paginated and sorted list of partition directories in the base directory.
func (b *BundlePartitionProvider) ListPartitions(
	baseDir string,
	sortOrder string,
	pageToken string,
	pageSize int,
) (partitions []string, nextPageToken string, err error) {
	return dirstore.ListDirs(baseDir, sortOrder, pageToken, pageSize)
}
