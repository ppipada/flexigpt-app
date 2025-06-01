package dirstore

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// TimeExtractor is a function that returns the creation time of a file.
type TimeExtractor func(filename string) (time.Time, error)

// MonthPartitionProvider decides directories yyyyMM from TimeExtractor.
type MonthPartitionProvider struct {
	TimeFn TimeExtractor
}

// GetPartitionDir implements the PartitionProvider interface.
func (p *MonthPartitionProvider) GetPartitionDir(filename string) string {
	if t, err := p.TimeFn(filename); err == nil {
		return t.Format("200601")
	}
	// Fallback – never explode; put file into current month.
	return time.Now().Format("200601")
}

// ListPartitions returns a paginated and sorted list of partition directories in the base directory.
func (p *MonthPartitionProvider) ListPartitions(
	baseDir string,
	sortOrder string,
	pageToken string,
	pageSize int,
) (partitions []string, nextPageToken string, err error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			partitions = append(partitions, entry.Name())
		}
	}

	// Sort partitions
	switch strings.ToLower(sortOrder) {
	case "asc":
		sort.Strings(partitions)
	case "desc":
		sort.Sort(sort.Reverse(sort.StringSlice(partitions)))
	default:
		return nil, "", fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	// Decode page token
	start := 0
	if pageToken != "" {
		tokenData, err := base64.StdEncoding.DecodeString(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		if err := json.Unmarshal(tokenData, &start); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}

	// Apply pagination
	end := start + pageSize
	if end > len(partitions) {
		end = len(partitions)
	}

	// Generate next page token
	if end < len(partitions) {
		nextPageTokenData, _ := json.Marshal(end)
		nextPageToken = base64.StdEncoding.EncodeToString(nextPageTokenData)
	}

	return partitions[start:end], nextPageToken, nil
}
