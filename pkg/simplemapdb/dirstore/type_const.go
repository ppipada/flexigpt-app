package dirstore

import "errors"

const (
	SortOrderAscending  = "asc"
	SortOrderDescending = "desc"
)

var ErrCannotReadPartition = errors.New("failed to read partition directory")

type FileKey struct {
	FileName string
	XAttr    any
}

// PartitionProvider defines an interface for determining the partition directory for a file.
type PartitionProvider interface {
	GetPartitionDir(key FileKey) (string, error)
	ListPartitions(baseDir, sortOrder, pageToken string,
		pageSize int) ([]string, string, error)
}
