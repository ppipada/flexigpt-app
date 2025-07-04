package dirstore

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ListDirs returns a paginated and sorted list of directories in the base directory.
func ListDirs(
	baseDir string,
	sortOrder string,
	pageToken string,
	pageSize int,
) (dirs []string, nextPageToken string, err error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	// Sort partitions.
	switch strings.ToLower(sortOrder) {
	case SortOrderAscending:
		sort.Strings(dirs)
	case SortOrderDescending:
		sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	default:
		return nil, "", fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	// Decode page token.
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

	// Apply pagination.
	end := min(start+pageSize, len(dirs))

	// Generate next page token.
	if end < len(dirs) {
		nextPageTokenData, _ := json.Marshal(end)
		nextPageToken = base64.StdEncoding.EncodeToString(nextPageTokenData)
	}

	return dirs[start:end], nextPageToken, nil
}
