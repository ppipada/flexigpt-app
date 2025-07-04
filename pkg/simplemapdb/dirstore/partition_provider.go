package dirstore

import (
	"fmt"
	"time"
)

// NoPartitionProvider is a default implementation that treats the base directory as a single partition.
type NoPartitionProvider struct{}

// GetPartitionDir returns an empty string, indicating no partitioning.
func (p *NoPartitionProvider) GetPartitionDir(_ FileKey) (string, error) {
	return "", nil
}

// ListPartitions returns a single partition representing the base directory.
func (p *NoPartitionProvider) ListPartitions(
	baseDir string,
	sortOrder string,
	pageToken string,
	pageSize int,
) (partitions []string, nextPageToken string, err error) {
	return []string{""}, "", nil
}

// TimeExtractor is a function that returns the creation time of a file.
type TimeExtractor func(key FileKey) (time.Time, error)

// MonthPartitionProvider decides directories yyyyMM from TimeExtractor.
type MonthPartitionProvider struct {
	TimeFn TimeExtractor
}

// GetPartitionDir implements the PartitionProvider interface.
func (p *MonthPartitionProvider) GetPartitionDir(key FileKey) (string, error) {
	t, err := p.TimeFn(key)
	if err != nil {
		return "", fmt.Errorf("could not get time for file: %s err: %w", key.FileName, err)
	}
	return t.Format("200601"), nil
}

// ListPartitions returns a paginated and sorted list of partition directories in the base directory.
func (p *MonthPartitionProvider) ListPartitions(
	baseDir string,
	sortOrder string,
	pageToken string,
	pageSize int,
) (partitions []string, nextPageToken string, err error) {
	return ListDirs(baseDir, sortOrder, pageToken, pageSize)
}
