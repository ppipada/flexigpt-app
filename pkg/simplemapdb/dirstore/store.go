package dirstore

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	simplemapdbFileStore "github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
)

// PartitionProvider defines an interface for determining the partition directory for a file.
type PartitionProvider interface {
	GetPartitionDir(filename string) string
	ListPartitions(
		baseDir string,
		sortOrder string,
		pageToken string,
		pageSize int,
	) ([]string, string, error)
}

// MapDirectoryStore manages multiple MapFileStores within a directory.
type MapDirectoryStore struct {
	baseDir           string
	pageSize          int
	PartitionProvider PartitionProvider
	listeners         []simplemapdbFileStore.Listener
}

// Option is a functional option for configuring the MapDirectoryStore.
type Option func(*MapDirectoryStore)

// WithPageSize sets the default page size for pagination.
func WithPageSize(size int) Option {
	return func(mds *MapDirectoryStore) {
		mds.pageSize = size
	}
}

// WithPartitionProvider sets a custom partition provider.
func WithPartitionProvider(provider PartitionProvider) Option {
	return func(mds *MapDirectoryStore) {
		mds.PartitionProvider = provider
	}
}

// WithListeners registers one or more listeners when the directory store is
// created.
func WithListeners(ls ...simplemapdbFileStore.Listener) Option {
	return func(mds *MapDirectoryStore) {
		mds.listeners = append(mds.listeners, ls...)
	}
}

// NewMapDirectoryStore initializes a new MapDirectoryStore with the given base directory and options.
func NewMapDirectoryStore(
	baseDir string,
	createIfNotExists bool,
	opts ...Option,
) (*MapDirectoryStore, error) {
	// Resolve the base directory path.
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base directory path: %w", err)
	}

	// Ensure the base directory exists or create it if allowed.
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		if createIfNotExists {
			if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", baseDir, err)
			}
		} else {
			return nil, fmt.Errorf("directory %s does not exist", baseDir)
		}
	}

	mds := &MapDirectoryStore{
		baseDir: baseDir,
		// Default page size.
		pageSize: 25,
		// Default to no partitioning.
		PartitionProvider: &NoPartitionProvider{},
	}

	for _, opt := range opts {
		opt(mds)
	}

	return mds, nil
}

// SetFileData creates or truncates a file and sets the provided data.
// If the data is nil, it initializes an empty map.
func (mds *MapDirectoryStore) SetFileData(filename string, data map[string]any) error {
	// Check if the filename contains any directory components.
	if strings.Contains(filename, string(os.PathSeparator)) {
		return fmt.Errorf("filename should not contain directory components: %s", filename)
	}
	if data == nil {
		return fmt.Errorf("cannot set nil data to file %s", filename)
	}

	partitionDir := mds.PartitionProvider.GetPartitionDir(filename)
	filePath := filepath.Join(mds.baseDir, partitionDir, filename)

	// Ensure the partition directory exists.
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return fmt.Errorf(
			"failed to create partition directory %s: %w",
			filepath.Dir(filePath),
			err,
		)
	}

	// Create or truncate the file.
	store, err := simplemapdbFileStore.NewMapFileStore(
		filePath,
		data,
		simplemapdbFileStore.WithCreateIfNotExists(true),
		simplemapdbFileStore.WithListeners(mds.listeners...),
	)
	if err != nil {
		return fmt.Errorf("failed to create or truncate file %s: %w", filename, err)
	}

	// Set data.
	if err := store.SetAll(data); err != nil {
		return fmt.Errorf("failed to set data for file %s: %w", filename, err)
	}

	return nil
}

// GetFileData returns the data from the specified file in the store.
func (mds *MapDirectoryStore) GetFileData(
	filename string,
	forceFetch bool,
) (map[string]any, error) {
	partitionDir := mds.PartitionProvider.GetPartitionDir(filename)
	filePath := filepath.Join(mds.baseDir, partitionDir, filename)
	store, err := simplemapdbFileStore.NewMapFileStore(
		filePath,
		map[string]any{"k": "v"},
		simplemapdbFileStore.WithListeners(mds.listeners...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get file %s: %w", filename, err)
	}
	return store.GetAll(forceFetch)
}

// DeleteFile removes the file with the given filename from the base directory.
func (mds *MapDirectoryStore) DeleteFile(filename string) error {
	partitionDir := mds.PartitionProvider.GetPartitionDir(filename)
	filePath := filepath.Join(mds.baseDir, partitionDir, filename)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", filename, err)
	}
	return nil
}

// Helper to read and sort files in a partition directory.
func (mds *MapDirectoryStore) readPartitionFiles(
	partitionPath, sortOrder string,
) ([]string, error) {
	files, err := os.ReadDir(partitionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read partition directory %s: %w", partitionPath, err)
	}

	var partitionFileNames []string
	for _, file := range files {
		if !file.IsDir() {
			partitionFileNames = append(partitionFileNames, file.Name())
		}
	}

	switch strings.ToLower(sortOrder) {
	case "asc":
		sort.Strings(partitionFileNames)
	case "desc":
		sort.Sort(sort.Reverse(sort.StringSlice(partitionFileNames)))
	default:
		return nil, fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	return partitionFileNames, nil
}

func (mds *MapDirectoryStore) ListFiles(
	initialSortOrder, pageToken string,
) (filenames []string, nextPageToken string, err error) {
	// Decode page token.
	var tokenData struct {
		PartitionPageToken string `json:"PartitionPageToken"`
		FileIndex          int    `json:"FileIndex"`
		SortOrder          string `json:"SortOrder"`
	}
	if pageToken != "" {
		tokenBytes, err := base64.StdEncoding.DecodeString(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		if err := json.Unmarshal(tokenBytes, &tokenData); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	} else {
		tokenData.SortOrder = initialSortOrder
	}

	for {
		// Get a paginated list of partition directories.
		partitions, nextPartitionPageToken, err := mds.PartitionProvider.ListPartitions(
			mds.baseDir,
			tokenData.SortOrder,
			tokenData.PartitionPageToken,
			1,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to list partitions: %w", err)
		}

		if len(partitions) == 0 {
			break
		}

		partitionPath := filepath.Join(mds.baseDir, partitions[0])
		partitionFileNames, err := mds.readPartitionFiles(partitionPath, tokenData.SortOrder)
		if err != nil {
			return nil, "", err
		}

		for j := tokenData.FileIndex; j < len(partitionFileNames); j++ {
			filenames = append(filenames, filepath.Join(partitions[0], partitionFileNames[j]))
			if len(filenames) >= mds.pageSize {
				// Check if we are at the end of the current partition.
				if j+1 < len(partitionFileNames) {
					// More files in the current partition.
					nextPageTokenData, _ := json.Marshal(struct {
						PartitionPageToken string `json:"PartitionPageToken"`
						FileIndex          int    `json:"FileIndex"`
						SortOrder          string `json:"SortOrder"`
					}{PartitionPageToken: tokenData.PartitionPageToken, FileIndex: j + 1, SortOrder: tokenData.SortOrder})
					nextPageToken = base64.StdEncoding.EncodeToString(nextPageTokenData)
					return filenames, nextPageToken, nil
				} else {
					// Move to the next partition.
					nextPageTokenData, _ := json.Marshal(struct {
						PartitionPageToken string `json:"PartitionPageToken"`
						FileIndex          int    `json:"FileIndex"`
						SortOrder          string `json:"SortOrder"`
					}{PartitionPageToken: nextPartitionPageToken, FileIndex: 0, SortOrder: tokenData.SortOrder})
					nextPageToken = base64.StdEncoding.EncodeToString(nextPageTokenData)
					return filenames, nextPageToken, nil
				}
			}
		}
		// Reset file index for the next partition.
		tokenData.FileIndex = 0

		// If there are no more partitions to process, break the loop.
		if nextPartitionPageToken == "" {
			break
		}

		// Update the partition page token for the next iteration.
		tokenData.PartitionPageToken = nextPartitionPageToken
	}

	return filenames, "", nil
}
