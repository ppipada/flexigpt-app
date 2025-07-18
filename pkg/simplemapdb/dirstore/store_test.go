package dirstore

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// CRUD Tests.

func TestMapDirectoryStore_CRUD(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tests := []struct {
		name               string
		partitionProvider  PartitionProvider
		filename           string
		data               map[string]any
		expectedPartition  string
		expectedFileExists bool
		expectError        bool
	}{
		{
			name:               "NoPartitionProvider - Create File",
			partitionProvider:  &NoPartitionProvider{},
			filename:           "testfile.json",
			data:               map[string]any{"key": "value"},
			expectedPartition:  "",
			expectedFileExists: true,
			expectError:        false,
		},
		{
			name: "MonthPartitionProvider - Create File",
			partitionProvider: &MonthPartitionProvider{
				TimeFn: func(fileKey FileKey) (time.Time, error) { return now, nil },
			},
			filename:           "testfile.json",
			data:               map[string]any{"key": "value"},
			expectedPartition:  now.Format("200601"),
			expectedFileExists: true,
			expectError:        false,
		},
		{
			name:               "NoPartitionProvider - Empty Data",
			partitionProvider:  &NoPartitionProvider{},
			filename:           "emptyfile.json",
			data:               map[string]any{},
			expectedPartition:  "",
			expectedFileExists: true,
			expectError:        false,
		},
		{
			name:               "Invalid Directory",
			partitionProvider:  &NoPartitionProvider{},
			filename:           "invalid/testfile.json",
			data:               map[string]any{"key": "value"},
			expectedPartition:  "",
			expectedFileExists: false,
			expectError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			baseDir := t.TempDir()
			mds, err := NewMapDirectoryStore(
				baseDir,
				true,
				WithPartitionProvider(tt.partitionProvider),
			)
			if err != nil {
				t.Fatalf("failed to create MapDirectoryStore: %v", err)
			}

			err = mds.SetFileData(FileKey{FileName: tt.filename}, tt.data)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			partitionDir := filepath.Join(baseDir, tt.expectedPartition)
			filePath := filepath.Join(partitionDir, tt.filename)

			_, err = os.Stat(filePath)
			if tt.expectedFileExists {
				if os.IsNotExist(err) {
					t.Fatalf("expected file to exist but it does not")
				}
			} else {
				if !os.IsNotExist(err) {
					t.Fatalf("expected file not to exist but it does")
				}
			}

			if tt.expectedFileExists {
				data, err := mds.GetFileData(FileKey{FileName: tt.filename}, false)
				if err != nil {
					t.Fatalf("failed to get file data: %v", err)
				}
				if len(data) != len(tt.data) {
					t.Fatalf("expected data length %d, got %d", len(tt.data), len(data))
				}
				for k, v := range tt.data {
					if data[k] != v {
						t.Fatalf("expected data[%s] = %v, got %v", k, v, data[k])
					}
				}
			}
		})
	}
}

func TestMapDirectoryStore_DeleteFile(t *testing.T) {
	t.Parallel()
	baseDir := t.TempDir()
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(&NoPartitionProvider{}),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	filename := "testfile.json"
	err = mds.SetFileData(FileKey{FileName: filename}, map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("failed to set file data: %v", err)
	}

	filePath := filepath.Join(baseDir, filename)
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		t.Fatalf("expected file to exist but it does not")
	}

	err = mds.DeleteFile(FileKey{FileName: filename})
	if err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}

	_, err = os.Stat(filePath)
	if !os.IsNotExist(err) {
		t.Fatalf("expected file not to exist but it does")
	}
}

// Listing Tests: Basic, Pagination, Filtering, Prefix.

func TestMapDirectoryStore_ListFiles_BasicAndSort(t *testing.T) {
	baseDir := t.TempDir()
	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filekey FileKey) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	files := []string{"file1.json", "file2.json", "file3.json"}
	for _, filename := range files {
		if err := mds.SetFileData(FileKey{FileName: filename}, map[string]any{"key": "value"}); err != nil {
			t.Fatalf("failed to set file data: %v", err)
		}
	}

	tests := []struct {
		name          string
		sortOrder     string
		expectedFiles []string
		expectError   bool
	}{
		{
			name:          "Ascending",
			sortOrder:     SortOrderAscending,
			expectedFiles: files,
		},
		{
			name:          "Descending",
			sortOrder:     SortOrderDescending,
			expectedFiles: []string{"file3.json", "file2.json", "file1.json"},
		},
		{
			name:        "InvalidSortOrder",
			sortOrder:   "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, _, err := mds.ListFiles(ListingConfig{SortOrder: tt.sortOrder}, "")
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var filenames []string
			for _, file := range files {
				_, filename := filepath.Split(file.BaseRelativePath)
				filenames = append(filenames, filename)
			}
			if len(filenames) != len(tt.expectedFiles) {
				t.Fatalf("expected %d files, got %d", len(tt.expectedFiles), len(filenames))
			}
			for i, expectedFile := range tt.expectedFiles {
				if filenames[i] != expectedFile {
					t.Fatalf("expected file %s, got %s", expectedFile, filenames[i])
				}
			}
		})
	}
}

func TestMapDirectoryStore_ListFiles_NoPartitionProvider_Pagination(t *testing.T) {
	baseDir := t.TempDir()
	files := []string{
		"file1.json", "file2.json", "file3.json", "file4.json", "file5.json",
		"file6.json", "file7.json", "file8.json", "file9.json",
	}
	testData := map[string]any{"key": "value"}
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	for _, file := range files {
		filePath := filepath.Join(baseDir, file)
		fileData, err := json.Marshal(testData)
		if err != nil {
			t.Fatalf("failed to marshal test data: %v", err)
		}
		if err := os.WriteFile(filePath, fileData, 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
	}

	tests := []struct {
		name          string
		sortOrder     string
		pageSize      int
		expectedPages [][]string
	}{
		{
			name:      "Ascending",
			sortOrder: SortOrderAscending,
			pageSize:  4,
			expectedPages: [][]string{
				{"file1.json", "file2.json", "file3.json", "file4.json"},
				{"file5.json", "file6.json", "file7.json", "file8.json"},
				{"file9.json"},
			},
		},
		{
			name:      "Descending",
			sortOrder: SortOrderDescending,
			pageSize:  4,
			expectedPages: [][]string{
				{"file9.json", "file8.json", "file7.json", "file6.json"},
				{"file5.json", "file4.json", "file3.json", "file2.json"},
				{"file1.json"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageToken := ""
			mds, err := NewMapDirectoryStore(
				baseDir,
				true,
				WithPageSize(tt.pageSize),
			)
			if err != nil {
				t.Fatalf("failed to create MapDirectoryStore: %v", err)
			}
			for pageIndex, expectedFiles := range tt.expectedPages {
				files, nextPageToken, err := mds.ListFiles(
					ListingConfig{SortOrder: tt.sortOrder},
					pageToken,
				)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(files) != len(expectedFiles) {
					t.Fatalf(
						"expected %d files on page %d, got %d",
						len(expectedFiles),
						pageIndex+1,
						len(files),
					)
				}
				for i, expectedFile := range expectedFiles {
					if files[i].BaseRelativePath != expectedFile {
						t.Fatalf(
							"expected file %s on page %d, got %s",
							expectedFile,
							pageIndex+1,
							files[i],
						)
					}
				}
				pageToken = nextPageToken
			}
		})
	}
}

func TestMapDirectoryStore_ListFiles_MultiPartition_Pagination(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "listdir")
	partitions := []string{"202301", "202302", "202303"}
	files := []string{"file1.json", "file2.json", "file3.json", "file4.json", "file5.json"}
	testData := map[string]any{"key": "value"}

	for _, partition := range partitions {
		partitionDir := filepath.Join(baseDir, partition)
		if err := os.MkdirAll(partitionDir, os.ModePerm); err != nil {
			t.Fatalf("failed to create partition directory: %v", err)
		}
		for _, file := range files {
			filePath := filepath.Join(partitionDir, file)
			fileData, err := json.Marshal(testData)
			if err != nil {
				t.Fatalf("failed to marshal test data: %v", err)
			}
			if err := os.WriteFile(filePath, fileData, 0o600); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}
		}
	}

	tests := []struct {
		name          string
		sortOrder     string
		pageSize      int
		expectedPages [][]string
	}{
		{
			name:      "Ascending",
			sortOrder: SortOrderAscending,
			pageSize:  4,
			expectedPages: [][]string{
				{
					"202301/file1.json",
					"202301/file2.json",
					"202301/file3.json",
					"202301/file4.json",
				},
				{
					"202301/file5.json",
					"202302/file1.json",
					"202302/file2.json",
					"202302/file3.json",
				},
				{
					"202302/file4.json",
					"202302/file5.json",
					"202303/file1.json",
					"202303/file2.json",
				},
				{"202303/file3.json", "202303/file4.json", "202303/file5.json"},
			},
		},
		{
			name:      "Descending",
			sortOrder: SortOrderDescending,
			pageSize:  4,
			expectedPages: [][]string{
				{
					"202303/file5.json",
					"202303/file4.json",
					"202303/file3.json",
					"202303/file2.json",
				},
				{
					"202303/file1.json",
					"202302/file5.json",
					"202302/file4.json",
					"202302/file3.json",
				},
				{
					"202302/file2.json",
					"202302/file1.json",
					"202301/file5.json",
					"202301/file4.json",
				},
				{"202301/file3.json", "202301/file2.json", "202301/file1.json"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageToken := ""
			for pageIndex, expectedFiles := range tt.expectedPages {
				partitionProvider := &MonthPartitionProvider{
					TimeFn: func(filekey FileKey) (time.Time, error) { return time.Now(), nil },
				}
				mds, err := NewMapDirectoryStore(
					baseDir,
					true,
					WithPartitionProvider(partitionProvider),
					WithPageSize(tt.pageSize),
				)
				if err != nil {
					t.Fatalf("failed to create MapDirectoryStore: %v", err)
				}
				files, nextPageToken, err := mds.ListFiles(
					ListingConfig{SortOrder: tt.sortOrder},
					pageToken,
				)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(files) != len(expectedFiles) {
					t.Fatalf(
						"expected %d files on page %d, got %d",
						len(expectedFiles),
						pageIndex+1,
						len(files),
					)
				}
				for i, expectedFile := range expectedFiles {
					if files[i].BaseRelativePath != expectedFile {
						t.Fatalf(
							"expected file %s on page %d, got %s",
							expectedFile,
							pageIndex+1,
							files[i],
						)
					}
				}
				pageToken = nextPageToken
			}
		})
	}
}

// Listing Tests: Filtering by Partition and FileName Prefix.

func TestMapDirectoryStore_ListFiles_FilteredPartitions(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301", "202302", "202303"}
	files := []string{"a.json", "b.json", "c.json"}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filekey FileKey) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
		WithPageSize(10),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	tests := []struct {
		name             string
		sortOrder        string
		filterPartitions []string
		expectedFiles    []string
	}{
		{
			name:             "Non-filtered, ascending",
			sortOrder:        SortOrderAscending,
			filterPartitions: nil,
			expectedFiles: []string{
				"202301/a.json", "202301/b.json", "202301/c.json",
				"202302/a.json", "202302/b.json", "202302/c.json",
				"202303/a.json", "202303/b.json", "202303/c.json",
			},
		},
		{
			name:             "Non-filtered, descending",
			sortOrder:        SortOrderDescending,
			filterPartitions: nil,
			expectedFiles: []string{
				"202303/c.json", "202303/b.json", "202303/a.json",
				"202302/c.json", "202302/b.json", "202302/a.json",
				"202301/c.json", "202301/b.json", "202301/a.json",
			},
		},
		{
			name:             "Filtered, single partition",
			sortOrder:        SortOrderAscending,
			filterPartitions: []string{"202302"},
			expectedFiles:    []string{"202302/a.json", "202302/b.json", "202302/c.json"},
		},
		{
			name:             "Filtered, multiple partitions, custom order",
			sortOrder:        SortOrderAscending,
			filterPartitions: []string{"202303", "202301"},
			expectedFiles: []string{
				"202303/a.json", "202303/b.json", "202303/c.json",
				"202301/a.json", "202301/b.json", "202301/c.json",
			},
		},
		{
			name:             "Filtered, multiple partitions, descending",
			sortOrder:        SortOrderDescending,
			filterPartitions: []string{"202302", "202301"},
			expectedFiles: []string{
				"202302/c.json", "202302/b.json", "202302/a.json",
				"202301/c.json", "202301/b.json", "202301/a.json",
			},
		},
		{
			name:             "Filtered, empty partition list",
			sortOrder:        SortOrderAscending,
			filterPartitions: []string{},
			expectedFiles: []string{
				"202301/a.json", "202301/b.json", "202301/c.json",
				"202302/a.json", "202302/b.json", "202302/c.json",
				"202303/a.json", "202303/b.json", "202303/c.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, nextPageToken, err := mds.ListFiles(
				ListingConfig{SortOrder: tt.sortOrder, FilterPartitions: tt.filterPartitions},
				"",
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if nextPageToken != "" {
				t.Fatalf("expected no next page token, got %q", nextPageToken)
			}
			if len(files) != len(tt.expectedFiles) {
				t.Fatalf("expected %d files, got %d", len(tt.expectedFiles), len(files))
			}
			for i, want := range tt.expectedFiles {
				if files[i].BaseRelativePath != want {
					t.Errorf("at %d: want %q, got %q", i, want, files[i])
				}
			}
		})
	}
}

func TestMapDirectoryStore_ListFiles_FilteredPartitions_Pagination(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301", "202302"}
	files := []string{"a.json", "b.json", "c.json", "d.json"}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filekey FileKey) (time.Time, error) { return time.Now(), nil },
	}
	pageSize := 3
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
		WithPageSize(pageSize),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	tests := []struct {
		name             string
		sortOrder        string
		filterPartitions []string
		expectedPages    [][]string
	}{
		{
			name:             "Filtered, paginated, asc",
			sortOrder:        SortOrderAscending,
			filterPartitions: []string{"202301", "202302"},
			expectedPages: [][]string{
				{"202301/a.json", "202301/b.json", "202301/c.json"},
				{"202301/d.json", "202302/a.json", "202302/b.json"},
				{"202302/c.json", "202302/d.json"},
			},
		},
		{
			name:             "Filtered, paginated, desc",
			sortOrder:        SortOrderDescending,
			filterPartitions: []string{"202302", "202301"},
			expectedPages: [][]string{
				{"202302/d.json", "202302/c.json", "202302/b.json"},
				{"202302/a.json", "202301/d.json", "202301/c.json"},
				{"202301/b.json", "202301/a.json"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageToken := ""
			for pageIdx, wantFiles := range tt.expectedPages {
				files, nextPageToken, err := mds.ListFiles(
					ListingConfig{SortOrder: tt.sortOrder, FilterPartitions: tt.filterPartitions},
					pageToken,
				)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(files) != len(wantFiles) {
					t.Fatalf(
						"page %d: expected %d files, got %d",
						pageIdx+1,
						len(wantFiles),
						len(files),
					)
				}
				for i, want := range wantFiles {
					if files[i].BaseRelativePath != want {
						t.Errorf("page %d, file %d: want %q, got %q", pageIdx+1, i, want, files[i])
					}
				}
				pageToken = nextPageToken
				if pageIdx < len(tt.expectedPages)-1 && pageToken == "" {
					t.Fatalf("expected next page token for page %d, got empty", pageIdx+1)
				}
				if pageIdx == len(tt.expectedPages)-1 && pageToken != "" {
					t.Fatalf("expected no next page token for last page, got %q", pageToken)
				}
			}
		})
	}
}

func TestMapDirectoryStore_ListFiles_FilenamePrefixFiltering(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301", "202302"}
	files := []string{
		"apple.json", "apricot.json", "banana.json", "berry.json", "cherry.json",
		"apple_pie.json", "banana_bread.json", "berry_tart.json", "zebra.json",
	}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filekey FileKey) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
		WithPageSize(20),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	type want struct{ files []string }
	tests := []struct {
		name             string
		sortOrder        string
		filterPartitions []string
		filenamePrefix   string
		want             want
	}{
		{
			name:           "No prefix, ascending",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "",
			want: want{files: []string{
				"202301/apple.json", "202301/apple_pie.json", "202301/apricot.json", "202301/banana.json", "202301/banana_bread.json", "202301/berry.json", "202301/berry_tart.json", "202301/cherry.json", "202301/zebra.json",
				"202302/apple.json", "202302/apple_pie.json", "202302/apricot.json", "202302/banana.json", "202302/banana_bread.json", "202302/berry.json", "202302/berry_tart.json", "202302/cherry.json", "202302/zebra.json",
			}},
		},
		{
			name:           "Prefix 'apple', ascending",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "apple",
			want: want{files: []string{
				"202301/apple.json", "202301/apple_pie.json",
				"202302/apple.json", "202302/apple_pie.json",
			}},
		},
		{
			name:           "Prefix 'banana', ascending",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "banana",
			want: want{files: []string{
				"202301/banana.json", "202301/banana_bread.json",
				"202302/banana.json", "202302/banana_bread.json",
			}},
		},
		{
			name:           "Prefix 'berry', descending",
			sortOrder:      SortOrderDescending,
			filenamePrefix: "berry",
			want: want{files: []string{
				"202302/berry_tart.json", "202302/berry.json",
				"202301/berry_tart.json", "202301/berry.json",
			}},
		},
		{
			name:           "Prefix 'z', ascending",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "z",
			want: want{files: []string{
				"202301/zebra.json", "202302/zebra.json",
			}},
		},
		{
			name:           "Prefix 'notfound', ascending",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "notfound",
			want:           want{files: []string{}},
		},
		{
			name:             "Prefix '', filtered partition",
			sortOrder:        SortOrderAscending,
			filterPartitions: []string{"202301"},
			filenamePrefix:   "",
			want: want{files: []string{
				"202301/apple.json", "202301/apple_pie.json", "202301/apricot.json", "202301/banana.json", "202301/banana_bread.json", "202301/berry.json", "202301/berry_tart.json", "202301/cherry.json", "202301/zebra.json",
			}},
		},
		{
			name:             "Prefix 'ap', filtered partition",
			sortOrder:        SortOrderAscending,
			filterPartitions: []string{"202302"},
			filenamePrefix:   "ap",
			want: want{files: []string{
				"202302/apple.json", "202302/apple_pie.json", "202302/apricot.json",
			}},
		},
		{
			name:             "Prefix 'berry', filtered partition, descending",
			sortOrder:        SortOrderDescending,
			filterPartitions: []string{"202301"},
			filenamePrefix:   "berry",
			want: want{files: []string{
				"202301/berry_tart.json", "202301/berry.json",
			}},
		},
		{
			name:           "Prefix with underscore",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "banana_",
			want: want{files: []string{
				"202301/banana_bread.json", "202302/banana_bread.json",
			}},
		},
		{
			name:           "Prefix with special char",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "apple_",
			want: want{files: []string{
				"202301/apple_pie.json", "202302/apple_pie.json",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, nextPageToken, err := mds.ListFiles(
				ListingConfig{
					SortOrder:        tt.sortOrder,
					FilterPartitions: tt.filterPartitions,
					FilenamePrefix:   tt.filenamePrefix,
				},
				"",
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if nextPageToken != "" {
				t.Fatalf("expected no next page token, got %q", nextPageToken)
			}
			if len(files) != len(tt.want.files) {
				t.Fatalf("expected %d files, got %d: %v", len(tt.want.files), len(files), files)
			}
			for i, want := range tt.want.files {
				if files[i].BaseRelativePath != want {
					t.Errorf("at %d: want %q, got %q", i, want, files[i])
				}
			}
		})
	}
}

func TestMapDirectoryStore_ListFiles_FilenamePrefixFiltering_Pagination(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301"}
	files := []string{
		"apple.json", "apricot.json", "banana.json", "berry.json", "cherry.json",
		"apple_pie.json", "banana_bread.json", "berry_tart.json", "zebra.json",
	}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filekey FileKey) (time.Time, error) { return time.Now(), nil },
	}
	pageSize := 2
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
		WithPageSize(pageSize),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	type pageWant struct{ files []string }
	tests := []struct {
		name           string
		sortOrder      string
		filenamePrefix string
		expectedPages  []pageWant
	}{
		{
			name:           "Prefix 'apple', ascending, paginated",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "apple",
			expectedPages: []pageWant{
				{files: []string{"202301/apple.json", "202301/apple_pie.json"}},
			},
		},
		{
			name:           "Prefix 'b', ascending, paginated",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "b",
			expectedPages: []pageWant{
				{files: []string{"202301/banana.json", "202301/banana_bread.json"}},
				{files: []string{"202301/berry.json", "202301/berry_tart.json"}},
			},
		},
		{
			name:           "Prefix 'berry', ascending, paginated",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "berry",
			expectedPages: []pageWant{
				{files: []string{"202301/berry.json", "202301/berry_tart.json"}},
			},
		},
		{
			name:           "Prefix 'z', ascending, paginated",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "z",
			expectedPages: []pageWant{
				{files: []string{"202301/zebra.json"}},
			},
		},
		{
			name:           "Prefix '', ascending, paginated",
			sortOrder:      SortOrderAscending,
			filenamePrefix: "",
			expectedPages: []pageWant{
				{files: []string{"202301/apple.json", "202301/apple_pie.json"}},
				{files: []string{"202301/apricot.json", "202301/banana.json"}},
				{files: []string{"202301/banana_bread.json", "202301/berry.json"}},
				{files: []string{"202301/berry_tart.json", "202301/cherry.json"}},
				{files: []string{"202301/zebra.json"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageToken := ""
			for pageIdx, want := range tt.expectedPages {
				files, nextPageToken, err := mds.ListFiles(
					ListingConfig{
						SortOrder:      tt.sortOrder,
						FilenamePrefix: tt.filenamePrefix,
					},
					pageToken,
				)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(files) != len(want.files) {
					t.Fatalf(
						"page %d: expected %d files, got %d: %v",
						pageIdx+1,
						len(want.files),
						len(files),
						files,
					)
				}
				for i, wantFile := range want.files {
					if files[i].BaseRelativePath != wantFile {
						t.Errorf(
							"page %d, file %d: want %q, got %q",
							pageIdx+1,
							i,
							wantFile,
							files[i],
						)
					}
				}
				pageToken = nextPageToken
				if pageIdx < len(tt.expectedPages)-1 && pageToken == "" {
					t.Fatalf("expected next page token for page %d, got empty", pageIdx+1)
				}
				if pageIdx == len(tt.expectedPages)-1 && pageToken != "" {
					t.Fatalf("expected no next page token for last page, got %q", pageToken)
				}
			}
		})
	}
}

// Listing Tests: Partition Listing & Pagination.

func TestMapDirectoryStore_ListPartitions_Pagination(t *testing.T) {
	baseDir := t.TempDir()
	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filekey FileKey) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	partitions := []string{"202301", "202302", "202303"}
	for _, partition := range partitions {
		if err := os.Mkdir(filepath.Join(baseDir, partition), os.ModePerm); err != nil {
			t.Fatalf("failed to create partition directory: %v", err)
		}
	}

	tests := []struct {
		name          string
		sortOrder     string
		pageToken     string
		pageSize      int
		expectedParts []string
		expectError   bool
	}{
		{
			name:          "Ascending",
			sortOrder:     SortOrderAscending,
			pageToken:     "",
			pageSize:      2,
			expectedParts: []string{"202301", "202302"},
		},
		{
			name:          "Descending",
			sortOrder:     SortOrderDescending,
			pageToken:     "",
			pageSize:      2,
			expectedParts: []string{"202303", "202302"},
		},
		{
			name:        "InvalidSortOrder",
			sortOrder:   "invalid",
			pageToken:   "",
			pageSize:    2,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partitions, nextPageToken, err := mds.PartitionProvider.ListPartitions(
				baseDir, tt.sortOrder, tt.pageToken, tt.pageSize,
			)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(partitions) != len(tt.expectedParts) {
				t.Fatalf("expected %d partitions, got %d", len(tt.expectedParts), len(partitions))
			}
			for i, expectedPart := range tt.expectedParts {
				if partitions[i] != expectedPart {
					t.Fatalf("expected partition %s, got %s", expectedPart, partitions[i])
				}
			}
			if nextPageToken != "" {
				partitions, _, err = mds.PartitionProvider.ListPartitions(
					baseDir, tt.sortOrder, nextPageToken, tt.pageSize,
				)
				if err != nil {
					t.Fatalf("unexpected error on next page: %v", err)
				}
				if len(partitions) != 1 {
					t.Fatalf("expected 1 partition on next page, got %d", len(partitions))
				}
			}
		})
	}
}

// Listing Tests: Edge Cases & Error Handling.

func TestMapDirectoryStore_ListFiles_ErrorsAndEdgeCases(t *testing.T) {
	t.Parallel()

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filekey FileKey) (time.Time, error) { return time.Now(), nil },
	}

	t.Run("InvalidSortOrder", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		_, _, err = mds.ListFiles(ListingConfig{SortOrder: "notasort"}, "")
		if err == nil {
			t.Fatal("expected error for invalid sort order, got nil")
		}
	})

	t.Run("NonExistentPartition", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		files, nextPageToken, err := mds.ListFiles(
			ListingConfig{
				SortOrder:        SortOrderAscending,
				FilterPartitions: []string{"doesnotexist"},
			},
			"",
		)
		if err != nil {
			t.Fatalf(
				"expected partition skipped for non-existent partition in filter, got err %s",
				err,
			)
		}
		if len(files) != 0 {
			t.Fatalf("expected no files, got %v", files)
		}
		if nextPageToken != "" {
			t.Fatalf("expected no next page token, got %q", nextPageToken)
		}
	})

	t.Run("UnreadablePartitionDir", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		partition := "202301"
		dir := filepath.Join(baseDir, partition)
		if err := os.MkdirAll(dir, 0o000); err != nil {
			t.Fatalf("failed to create unreadable dir: %v", err)
		}
		defer func() { _ = os.Chmod(dir, 0o755) }()
		_, _, err = mds.ListFiles(
			ListingConfig{SortOrder: SortOrderAscending, FilterPartitions: []string{partition}}, "",
		)
		if err != nil {
			t.Fatalf(
				"expected partition skipped for unreadable partition in filter, got err %s",
				err,
			)
		}
	})

	t.Run("InvalidPageToken", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		_, _, err = mds.ListFiles(ListingConfig{SortOrder: SortOrderAscending}, "notbase64!")
		if err == nil {
			t.Fatal("expected error for invalid base64 page token, got nil")
		}
		bad := base64.StdEncoding.EncodeToString([]byte("notjson"))
		_, _, err = mds.ListFiles(ListingConfig{SortOrder: SortOrderAscending}, bad)
		if err == nil {
			t.Fatal("expected error for invalid JSON page token, got nil")
		}
	})

	t.Run("CorruptedPageToken", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		bad := base64.StdEncoding.EncodeToString([]byte("{notjson:"))
		_, _, err = mds.ListFiles(ListingConfig{SortOrder: SortOrderAscending}, bad)
		if err == nil {
			t.Fatal("expected error for corrupted JSON page token, got nil")
		}
	})

	t.Run("EmptyBaseDir", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		files, nextPageToken, err := mds.ListFiles(ListingConfig{SortOrder: SortOrderAscending}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Fatalf("expected no files, got %v", files)
		}
		if nextPageToken != "" {
			t.Fatalf("expected no next page token, got %q", nextPageToken)
		}
	})

	t.Run("FilterWithNonExistentPartition", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		partitions := []string{"202302"}
		files := []string{"a.json"}
		createFiles(t, baseDir, partitions, files)
		_, _, err = mds.ListFiles(
			ListingConfig{
				SortOrder:        SortOrderAscending,
				FilterPartitions: []string{"202301", "doesnotexist"},
			},
			"",
		)
		if err != nil {
			t.Fatalf(
				"expected partition skipped for non-existent partition in filter, got err %s",
				err,
			)
		}
	})

	t.Run("PageSizeLargerThanFiles", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		partitions := []string{"202303"}
		files := []string{"a.json", "b.json"}
		createFiles(t, baseDir, partitions, files)
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
			WithPageSize(10),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		got, nextPageToken, err := mds.ListFiles(ListingConfig{SortOrder: SortOrderAscending}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 files, got %d", len(got))
		}
		if nextPageToken != "" {
			t.Fatalf("expected no next page token, got %q", nextPageToken)
		}
	})

	t.Run("EmptyFilterPartitions", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		partitions := []string{"202305"}
		files := []string{"a.json"}
		createFiles(t, baseDir, partitions, files)
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		got, nextPageToken, err := mds.ListFiles(
			ListingConfig{SortOrder: SortOrderAscending, FilterPartitions: []string{}}, "",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 file, got %d", len(got))
		}
		if nextPageToken != "" {
			t.Fatalf("expected no next page token, got %q", nextPageToken)
		}
	})

	t.Run("FilenamePrefixFiltering_NoMatch", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		partitions := []string{"202306"}
		files := []string{"apple.json", "banana.json"}
		createFiles(t, baseDir, partitions, files)
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		got, nextPageToken, err := mds.ListFiles(
			ListingConfig{SortOrder: SortOrderAscending, FilenamePrefix: "zzz"}, "",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected 0 files, got %v", got)
		}
		if nextPageToken != "" {
			t.Fatalf("expected no next page token, got %q", nextPageToken)
		}
	})

	t.Run("FilteredPagination_EmptyPartition", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		files := []string{"a.json"}
		createFiles(t, baseDir, []string{"202307"}, files)
		if err := os.MkdirAll(filepath.Join(baseDir, "202302"), 0o755); err != nil {
			t.Fatalf("failed to create partition dir: %v", err)
		}
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
			WithPageSize(1),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		files1, nextPageToken, err := mds.ListFiles(
			ListingConfig{
				SortOrder:        SortOrderAscending,
				FilterPartitions: []string{"202307", "202302"},
			},
			"",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files1) != 1 || files1[0].BaseRelativePath != "202307/a.json" {
			t.Fatalf("expected [202307/a.json], got %v", files1)
		}
		if nextPageToken != "" {
			t.Fatalf("expected no next page token, got %q", nextPageToken)
		}
	})

	t.Run("FilenamePrefixFiltering_EmptyPartition", func(t *testing.T) {
		t.Parallel()
		baseDir := t.TempDir()
		files := []string{"apple.json", "banana.json"}
		createFiles(t, baseDir, []string{"202308"}, files)
		if err := os.MkdirAll(filepath.Join(baseDir, "202309"), 0o755); err != nil {
			t.Fatalf("failed to create partition dir: %v", err)
		}
		mds, err := NewMapDirectoryStore(
			baseDir,
			true,
			WithPartitionProvider(partitionProvider),
			WithPageSize(1),
		)
		if err != nil {
			t.Fatalf("failed to create MapDirectoryStore: %v", err)
		}
		files1, nextPageToken, err := mds.ListFiles(
			ListingConfig{
				SortOrder:        SortOrderAscending,
				FilterPartitions: []string{"202308", "202309"},
				FilenamePrefix:   "apple",
			}, "",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files1) != 1 || files1[0].BaseRelativePath != "202308/apple.json" {
			t.Fatalf("expected [202308/apple.json], got %v", files1)
		}
		if nextPageToken != "" {
			t.Fatalf("expected no next page token, got %q", nextPageToken)
		}
	})
}

// Helpers.

func createFiles(t *testing.T, baseDir string, partitions, files []string) {
	t.Helper()
	for _, partition := range partitions {
		dir := filepath.Join(baseDir, partition)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create partition dir: %v", err)
		}
		for _, file := range files {
			path := filepath.Join(dir, file)
			if err := os.WriteFile(path, []byte(`{"k":"v"}`), 0o600); err != nil {
				t.Fatalf("failed to write file: %v", err)
			}
		}
	}
}
