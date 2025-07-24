package dirstore

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
)

// MapDirectoryStore manages multiple MapFileStores within a directory.
type MapDirectoryStore struct {
	baseDir           string
	pageSize          int
	PartitionProvider PartitionProvider
	listeners         []filestore.Listener

	// OpenStores caches open MapFileStore instances per file path.
	openStores map[string]*filestore.MapFileStore
	openMu     sync.Mutex
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

// WithListeners registers one or more listeners when the directory store is created.
func WithListeners(ls ...filestore.Listener) Option {
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
		baseDir:           baseDir,
		pageSize:          10,
		PartitionProvider: &NoPartitionProvider{},
		openStores:        make(map[string]*filestore.MapFileStore),
	}

	for _, opt := range opts {
		opt(mds)
	}

	return mds, nil
}

// SetFileData sets the provided data for the given file.
// It is a thin wrapper around Open and SetAll.
func (mds *MapDirectoryStore) SetFileData(fileKey FileKey, data map[string]any) error {
	if data == nil {
		return fmt.Errorf("invalid request for file: %s", fileKey.FileName)
	}
	store, err := mds.Open(fileKey, true, data)
	if err != nil {
		return err
	}
	return store.SetAll(data)
}

// GetFileData returns the data from the specified file in the store.
// It is a thin wrapper around Open and GetAll.
func (mds *MapDirectoryStore) GetFileData(
	fileKey FileKey,
	forceFetch bool,
) (map[string]any, error) {
	// Use a dummy defaultData for opening if file exists.
	store, err := mds.Open(fileKey, false, map[string]any{})
	if err != nil {
		return nil, err
	}
	return store.GetAll(forceFetch)
}

// DeleteFile removes the file with the given filename from the base directory.
// It is a thin wrapper around Open and DeleteFile.
func (mds *MapDirectoryStore) DeleteFile(fileKey FileKey) error {
	store, err := mds.Open(fileKey, false, map[string]any{})
	if err != nil {
		return err
	}

	if err := store.DeleteFile(); err != nil {
		return err
	}
	return mds.Close(fileKey)
}

// Open returns a cached or newly created MapFileStore for the given FileKey.
// It is concurrency-safe and ensures only one instance per file path.
func (mds *MapDirectoryStore) Open(
	fileKey FileKey,
	createIfNotExists bool,
	defaultData map[string]any,
) (*filestore.MapFileStore, error) {
	filePath, err := mds.validateAndGetFilePath(fileKey)
	if err != nil {
		return nil, err
	}

	mds.openMu.Lock()
	defer mds.openMu.Unlock()
	store, ok := mds.openStores[filePath]
	if ok {
		return store, nil
	}

	// Ensure the partition directory exists if creating.
	if createIfNotExists {
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return nil, fmt.Errorf(
				"failed to create partition directory %s: %w",
				filepath.Dir(filePath),
				err,
			)
		}
	}

	// Create a new MapFileStore.
	store, err = filestore.NewMapFileStore(
		filePath,
		defaultData,
		filestore.WithCreateIfNotExists(createIfNotExists),
		filestore.WithListeners(mds.listeners...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open file store for %s: %w", fileKey.FileName, err)
	}

	mds.openStores[filePath] = store

	return store, nil
}

// Close closes the MapFileStore for the given FileKey (if it was opened) and removes it from the cache.
func (mds *MapDirectoryStore) Close(fileKey FileKey) error {
	filePath, err := mds.validateAndGetFilePath(fileKey)
	if err != nil {
		return err
	}

	mds.openMu.Lock()
	store, ok := mds.openStores[filePath]
	if ok {
		delete(mds.openStores, filePath)
	}
	mds.openMu.Unlock()

	if ok {
		return store.Close()
	}
	return nil
}

// CloseAll closes every cached MapFileStore in this directory instance and clears the cache.
func (mds *MapDirectoryStore) CloseAll() error {
	mds.openMu.Lock()
	stores := make([]*filestore.MapFileStore, 0, len(mds.openStores))
	for _, st := range mds.openStores {
		stores = append(stores, st)
	}
	mds.openStores = make(map[string]*filestore.MapFileStore)
	mds.openMu.Unlock()

	var firstErr error
	for _, st := range stores {
		if err := st.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// ListingConfig holds all options for listing files.
type ListingConfig struct {
	SortOrder        string
	PageSize         int
	FilterPartitions []string // If empty, list all partitions.
	FilenamePrefix   string   // If non-empty, only return files with this prefix.
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

type FileEntry struct {
	BaseRelativePath string
	PartitionName    string
	FileInfo         os.FileInfo
}

// ListFiles lists files according to the config and page token.
func (mds *MapDirectoryStore) ListFiles(
	config ListingConfig,
	pageToken string,
) (fileEntries []FileEntry, nextPageToken string, err error) {
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
		partitionFileInfos, err := mds.readPartitionFiles(
			partitionPath,
			token.SortOrder,
			token.FilenamePrefix,
		)
		if err != nil && errors.Is(err, ErrCannotReadPartition) {
			slog.Debug("skipping listing partition", "error", err)
			token.PartitionFilterPageToken.PartitionIndex++
		} else if err != nil {
			return nil, "", err
		}

		for j := token.FileIndex; j < len(partitionFileInfos); j++ {
			fileEntries = append(
				fileEntries,
				FileEntry{
					BaseRelativePath: filepath.Join(partitionName, partitionFileInfos[j].Name()),
					PartitionName:    partitionName,
					FileInfo:         partitionFileInfos[j],
				},
			)
			if len(fileEntries) > token.PageSize {
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
				return fileEntries[:token.PageSize], nextPageToken, nil
			}
		}
		token.FileIndex = 0

		if isFiltered {
			token.PartitionFilterPageToken.PartitionIndex++
		} else {
			if nextPartitionListingPageToken == "" {
				break
			}
			token.PartitionListingPageToken = nextPartitionListingPageToken
		}
	}

	return fileEntries, "", nil
}

// readPartitionFiles lists files in a partition, sorted and filtered by prefix.
func (mds *MapDirectoryStore) readPartitionFiles(
	partitionPath, sortOrder, filenamePrefix string,
) ([]os.FileInfo, error) {
	files, err := os.ReadDir(partitionPath)
	if err != nil {
		return nil, fmt.Errorf("partition %s: %w", partitionPath, ErrCannotReadPartition)
	}

	var fileInfos []os.FileInfo
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			if filenamePrefix == "" || strings.HasPrefix(name, filenamePrefix) {
				info, err := file.Info()
				if err != nil {
					return nil, fmt.Errorf("cannot stat file %s: %w", name, err)
				}
				fileInfos = append(fileInfos, info)
			}
		}
	}

	// Sort by name.
	sort.Slice(fileInfos, func(i, j int) bool {
		if strings.EqualFold(sortOrder, SortOrderDescending) {
			return fileInfos[i].Name() > fileInfos[j].Name()
		}
		return fileInfos[i].Name() < fileInfos[j].Name()
	})

	return fileInfos, nil
}

// validateAndGetFilePath validates the FileKey and returns the absolute file path.
func (mds *MapDirectoryStore) validateAndGetFilePath(fileKey FileKey) (string, error) {
	if fileKey.FileName == "" {
		return "", fmt.Errorf("invalid request for file: %s", fileKey.FileName)
	}
	// Check if the filename contains any directory components.
	if strings.Contains(fileKey.FileName, string(os.PathSeparator)) {
		return "", fmt.Errorf(
			"filename should not contain directory components: %s",
			fileKey.FileName,
		)
	}
	partitionDir, err := mds.PartitionProvider.GetPartitionDir(fileKey)
	if err != nil {
		return "", fmt.Errorf(
			"could not get partition dir for file: %s, err: %w",
			fileKey.FileName,
			err,
		)
	}
	filePath := filepath.Join(mds.baseDir, partitionDir, fileKey.FileName)
	return filePath, nil
}
