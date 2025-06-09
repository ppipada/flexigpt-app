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

const (
	SortOrderAscending  = "asc"
	SortOrderDescending = "desc"
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

// ListingConfig holds all options for listing files.
type ListingConfig struct {
	SortOrder string
	PageSize  int
	// If empty, list all partitions.
	FilterPartitions []string
	// If non-empty, only return files with this prefix.
	FilenamePrefix string
}

// PartitionFilterPageToken tracks progress through filtered partitions.
type PartitionFilterPageToken struct {
	PartitionIndex   int      `json:"partitionIndex"`
	FilterPartitions []string `json:"filterPartitions"`
}

// PageTokenData encodes all paging state.
type PageTokenData struct {
	FileIndex                 int                       `json:"fileIndex"`
	SortOrder                 string                    `json:"sortOrder"`
	PageSize                  int                       `json:"pageSize"`
	FilenamePrefix            string                    `json:"filenamePrefix,omitempty"`
	PartitionListingPageToken string                    `json:"partitionListingPageToken,omitempty"`
	PartitionFilterPageToken  *PartitionFilterPageToken `json:"partitionFilterPageToken,omitempty"`
}

// readPartitionFiles lists files in a partition, sorted and filtered by prefix.
func (mds *MapDirectoryStore) readPartitionFiles(
	partitionPath, sortOrder, filenamePrefix string,
) ([]string, error) {
	files, err := os.ReadDir(partitionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read partition directory %s: %w", partitionPath, err)
	}

	var partitionFileNames []string
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			if filenamePrefix == "" || strings.HasPrefix(name, filenamePrefix) {
				partitionFileNames = append(partitionFileNames, name)
			}
		}
	}

	switch strings.ToLower(sortOrder) {
	case SortOrderAscending:
		sort.Strings(partitionFileNames)
	case SortOrderDescending:
		sort.Sort(sort.Reverse(sort.StringSlice(partitionFileNames)))
	default:
		return nil, fmt.Errorf("invalid sort order: %s", sortOrder)
	}

	return partitionFileNames, nil
}

// ListFiles lists files according to the config and page token.
func (mds *MapDirectoryStore) ListFiles(
	config ListingConfig,
	pageToken string,
) (filenames []string, nextPageToken string, err error) {
	var token PageTokenData

	// Decode page token or initialize.
	if pageToken != "" {
		tokenBytes, err := base64.StdEncoding.DecodeString(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		if err := json.Unmarshal(tokenBytes, &token); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	} else {
		token.SortOrder = config.SortOrder
		if token.SortOrder == "" {
			token.SortOrder = SortOrderAscending
		}
		token.FileIndex = 0
		token.PageSize = config.PageSize
		if token.PageSize <= 0 {
			token.PageSize = mds.pageSize
		}
		token.FilenamePrefix = config.FilenamePrefix
		if len(config.FilterPartitions) > 0 {
			token.PartitionFilterPageToken = &PartitionFilterPageToken{
				PartitionIndex:   0,
				FilterPartitions: config.FilterPartitions,
			}
		}
	}

	isFiltered := token.PartitionFilterPageToken != nil

	for {
		var partitionName string
		var nextPartitionListingPageToken string

		if isFiltered {
			pfpt := token.PartitionFilterPageToken
			if pfpt.PartitionIndex >= len(pfpt.FilterPartitions) {
				// No more partitions.
				break
			}
			partitionName = pfpt.FilterPartitions[pfpt.PartitionIndex]
		} else {
			partitions, nextToken, err := mds.PartitionProvider.ListPartitions(
				mds.baseDir,
				token.SortOrder,
				token.PartitionListingPageToken,
				1,
			)
			if err != nil {
				return nil, "", fmt.Errorf("failed to list partitions: %w", err)
			}
			if len(partitions) == 0 {
				break
			}
			partitionName = partitions[0]
			nextPartitionListingPageToken = nextToken
		}

		partitionPath := filepath.Join(mds.baseDir, partitionName)
		partitionFileNames, err := mds.readPartitionFiles(
			partitionPath,
			token.SortOrder,
			token.FilenamePrefix,
		)
		if err != nil {
			return nil, "", err
		}

		for j := token.FileIndex; j < len(partitionFileNames); j++ {
			filenames = append(filenames, filepath.Join(partitionName, partitionFileNames[j]))
			if len(filenames) > token.PageSize {
				// Prepare next page token.
				nextToken := PageTokenData{
					SortOrder:      token.SortOrder,
					FileIndex:      j,
					PageSize:       token.PageSize,
					FilenamePrefix: token.FilenamePrefix,
				}
				if isFiltered {
					pfpt := *token.PartitionFilterPageToken
					nextToken.PartitionFilterPageToken = &PartitionFilterPageToken{
						PartitionIndex:   pfpt.PartitionIndex,
						FilterPartitions: pfpt.FilterPartitions,
					}
				} else {
					nextToken.PartitionListingPageToken = token.PartitionListingPageToken
				}
				nextPageTokenBytes, _ := json.Marshal(nextToken)
				nextPageToken = base64.StdEncoding.EncodeToString(nextPageTokenBytes)
				return filenames[:token.PageSize], nextPageToken, nil
			}
		}
		token.FileIndex = 0

		if isFiltered {
			token.PartitionFilterPageToken.PartitionIndex++
			if token.PartitionFilterPageToken.PartitionIndex >= len(
				token.PartitionFilterPageToken.FilterPartitions,
			) {
				break
			}
		} else {
			if nextPartitionListingPageToken == "" {
				break
			}
			token.PartitionListingPageToken = nextPartitionListingPageToken
		}
	}

	return filenames, "", nil
}
