package dirstore

// NoPartitionProvider is a default implementation that treats the base directory as a single partition.
type NoPartitionProvider struct{}

// GetPartitionDir returns an empty string, indicating no partitioning.
func (p *NoPartitionProvider) GetPartitionDir(filename string) string {
	return ""
}

// ListPartitions returns a single partition representing the base directory.
func (p *NoPartitionProvider) ListPartitions(
	baseDir string,
	sortOrder string,
	pageToken string,
	pageSize int,
) ([]string, string, error) {
	return []string{""}, "", nil
}
