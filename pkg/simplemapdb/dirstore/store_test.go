package dirstore

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMapDirectoryStore(t *testing.T) {
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
			name:               "Default Partition - Create File",
			partitionProvider:  &NoPartitionProvider{},
			filename:           "testfile.json",
			data:               map[string]any{"key": "value"},
			expectedPartition:  "",
			expectedFileExists: true,
			expectError:        false,
		},
		{
			name: "Month Partition - Create File",
			partitionProvider: &MonthPartitionProvider{
				TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
			},
			filename:           "testfile.json",
			data:               map[string]any{"key": "value"},
			expectedPartition:  time.Now().Format("200601"),
			expectedFileExists: true,
			expectError:        false,
		},
		{
			name:               "Default Partition - No Data",
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
			baseDir := t.TempDir()
			mds, err := NewMapDirectoryStore(
				baseDir,
				true,
				WithPartitionProvider(tt.partitionProvider),
			)
			if err != nil {
				t.Fatalf("failed to create MapDirectoryStore: %v", err)
			}

			err = mds.SetFileData(tt.filename, tt.data)
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

			// Check if the file exists.
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

			// Verify data if file exists.
			if tt.expectedFileExists {
				data, err := mds.GetFileData(tt.filename, false)
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

func TestListFiles(t *testing.T) {
	baseDir := t.TempDir()
	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	// Create some files.
	files := []string{"file1.json", "file2.json", "file3.json"}
	for _, file := range files {
		err := mds.SetFileData(file, map[string]any{"key": "value"})
		if err != nil {
			t.Fatalf("failed to set file data: %v", err)
		}
	}

	tests := []struct {
		name          string
		sortOrder     string
		pageToken     string
		expectedFiles []string
		expectError   bool
	}{
		{
			name:          "List Files Ascending",
			sortOrder:     "asc",
			pageToken:     "",
			expectedFiles: files,
			expectError:   false,
		},
		{
			name:          "List Files Descending",
			sortOrder:     "desc",
			pageToken:     "",
			expectedFiles: []string{"file3.json", "file2.json", "file1.json"},
			expectError:   false,
		},
		{
			name:          "Invalid Sort Order",
			sortOrder:     "invalid",
			pageToken:     "",
			expectedFiles: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, _, err := mds.ListFiles(
				ListingConfig{SortOrder: tt.sortOrder},
				tt.pageToken,
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

			// Extract filenames without partition path.
			var filenames []string
			for _, file := range files {
				_, filename := filepath.Split(file)
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

func TestDeleteFile(t *testing.T) {
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
	err = mds.SetFileData(filename, map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("failed to set file data: %v", err)
	}

	// Verify file exists.
	filePath := filepath.Join(baseDir, filename)
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		t.Fatalf("expected file to exist but it does not")
	}

	// Delete the file.
	err = mds.DeleteFile(filename)
	if err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}

	// Verify file does not exist.
	_, err = os.Stat(filePath)
	if !os.IsNotExist(err) {
		t.Fatalf("expected file not to exist but it does")
	}
}

func TestListPartitionsPagination(t *testing.T) {
	baseDir := t.TempDir()
	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	// Create partition directories.
	partitions := []string{"202301", "202302", "202303"}
	for _, partition := range partitions {
		err := os.Mkdir(filepath.Join(baseDir, partition), os.ModePerm)
		if err != nil {
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
			name:          "List Partitions Ascending",
			sortOrder:     "asc",
			pageToken:     "",
			pageSize:      2,
			expectedParts: []string{"202301", "202302"},
			expectError:   false,
		},
		{
			name:          "List Partitions Descending",
			sortOrder:     "desc",
			pageToken:     "",
			pageSize:      2,
			expectedParts: []string{"202303", "202302"},
			expectError:   false,
		},
		{
			name:          "Invalid Sort Order",
			sortOrder:     "invalid",
			pageToken:     "",
			pageSize:      2,
			expectedParts: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partitions, nextPageToken, err := mds.PartitionProvider.ListPartitions(
				baseDir,
				tt.sortOrder,
				tt.pageToken,
				tt.pageSize,
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

			// Test pagination by checking the next page.
			if nextPageToken != "" {
				partitions, _, err = mds.PartitionProvider.ListPartitions(
					baseDir,
					tt.sortOrder,
					nextPageToken,
					tt.pageSize,
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

func TestListFilesPaginationMonthPartition(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "listdir")
	slog.Info("TestListFilesPaginationMonthPartition", "BaseDir", baseDir)

	// Create some files across multiple partitions.
	partitions := []string{"202301", "202302", "202303"}
	files := []string{"file1.json", "file2.json", "file3.json", "file4.json", "file5.json"}
	testData := map[string]any{"key": "value"}

	for _, partition := range partitions {
		partitionDir := filepath.Join(baseDir, partition)
		err := os.MkdirAll(partitionDir, os.ModePerm)
		if err != nil {
			t.Fatalf("failed to create partition directory: %v", err)
		}

		for _, file := range files {
			filePath := filepath.Join(partitionDir, file)
			fileData, err := json.Marshal(testData)
			if err != nil {
				t.Fatalf("failed to marshal test data: %v", err)
			}
			err = os.WriteFile(filePath, fileData, 0o600)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}
		}
	}

	// Additional test cases for edge scenarios.
	tests := []struct {
		name          string
		sortOrder     string
		pageSize      int
		expectedPages [][]string
		expectError   bool
	}{
		{
			name:      "List Files Ascending with Pagination",
			sortOrder: "asc",
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
			expectError: false,
		},
		{
			name:      "List Files Descending with Pagination",
			sortOrder: "desc",
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
			expectError: false,
		},
		{
			name:      "Exact Page Size",
			sortOrder: "asc",
			pageSize:  5,
			expectedPages: [][]string{
				{
					"202301/file1.json",
					"202301/file2.json",
					"202301/file3.json",
					"202301/file4.json",
					"202301/file5.json",
				},
				{
					"202302/file1.json",
					"202302/file2.json",
					"202302/file3.json",
					"202302/file4.json",
					"202302/file5.json",
				},
				{
					"202303/file1.json",
					"202303/file2.json",
					"202303/file3.json",
					"202303/file4.json",
					"202303/file5.json",
				},
			},
			expectError: false,
		},
		{
			name:      "Single File in Partition",
			sortOrder: "asc",
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
			expectError: false,
		},
		{
			name:      "Empty Partition",
			sortOrder: "asc",
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
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageToken := ""
			for pageIndex, expectedFiles := range tt.expectedPages {
				partitionProvider := &MonthPartitionProvider{
					TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
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
				files, nextPageToken, err := mds.ListFiles(ListingConfig{SortOrder: tt.sortOrder},
					pageToken)
				if tt.expectError {
					if err == nil {
						t.Fatalf("expected error but got none")
					}
					return
				}
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
					if files[i] != expectedFile {
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

func TestListFilesNoPartitionProvider(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "nolistdir")
	slog.Info("TestListFilesNoPartitionProvider", "BaseDir", baseDir)

	// Create some files in a single directory (no partitions).
	files := []string{
		"file1.json",
		"file2.json",
		"file3.json",
		"file4.json",
		"file5.json",
		"file6.json",
		"file7.json",
		"file8.json",
		"file9.json",
	}
	testData := map[string]any{"key": "value"}
	err := os.MkdirAll(baseDir, os.ModePerm)
	if err != nil {
		t.Fatalf("failed to create partition directory: %v", err)
	}

	for _, file := range files {
		filePath := filepath.Join(baseDir, file)
		fileData, err := json.Marshal(testData)
		if err != nil {
			t.Fatalf("failed to marshal test data: %v", err)
		}
		err = os.WriteFile(filePath, fileData, 0o600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
	}

	// Test cases for pagination without partition provider.
	tests := []struct {
		name          string
		sortOrder     string
		pageSize      int
		expectedPages [][]string
		expectError   bool
	}{
		{
			name:      "List Files Ascending with Pagination",
			sortOrder: "asc",
			pageSize:  4,
			expectedPages: [][]string{
				{"file1.json", "file2.json", "file3.json", "file4.json"},
				{"file5.json", "file6.json", "file7.json", "file8.json"},
				{"file9.json"},
			},
			expectError: false,
		},
		{
			name:      "List Files Descending with Pagination",
			sortOrder: "desc",
			pageSize:  4,
			expectedPages: [][]string{
				{"file9.json", "file8.json", "file7.json", "file6.json"},
				{"file5.json", "file4.json", "file3.json", "file2.json"},
				{"file1.json"},
			},
			expectError: false,
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
				files, nextPageToken, err := mds.ListFiles(ListingConfig{SortOrder: tt.sortOrder},
					pageToken)
				if tt.expectError {
					if err == nil {
						t.Fatalf("expected error but got none")
					}
					return
				}
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
					if files[i] != expectedFile {
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

// Helper to create files in partitions
func createFiles(t *testing.T, baseDir string, partitions []string, files []string) {
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

func TestListFiles_FilteredVsNonFiltered(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301", "202302", "202303"}
	files := []string{"a.json", "b.json", "c.json"}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
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
			sortOrder:        "asc",
			filterPartitions: nil,
			expectedFiles: []string{
				"202301/a.json", "202301/b.json", "202301/c.json",
				"202302/a.json", "202302/b.json", "202302/c.json",
				"202303/a.json", "202303/b.json", "202303/c.json",
			},
		},
		{
			name:             "Non-filtered, descending",
			sortOrder:        "desc",
			filterPartitions: nil,
			expectedFiles: []string{
				"202303/c.json", "202303/b.json", "202303/a.json",
				"202302/c.json", "202302/b.json", "202302/a.json",
				"202301/c.json", "202301/b.json", "202301/a.json",
			},
		},
		{
			name:             "Filtered, single partition",
			sortOrder:        "asc",
			filterPartitions: []string{"202302"},
			expectedFiles:    []string{"202302/a.json", "202302/b.json", "202302/c.json"},
		},
		{
			name:             "Filtered, multiple partitions, custom order",
			sortOrder:        "asc",
			filterPartitions: []string{"202303", "202301"},
			expectedFiles: []string{
				"202303/a.json", "202303/b.json", "202303/c.json",
				"202301/a.json", "202301/b.json", "202301/c.json",
			},
		},
		{
			name:             "Filtered, multiple partitions, descending",
			sortOrder:        "desc",
			filterPartitions: []string{"202302", "202301"},
			expectedFiles: []string{
				"202302/c.json", "202302/b.json", "202302/a.json",
				"202301/c.json", "202301/b.json", "202301/a.json",
			},
		},
		{
			name:             "Filtered, empty partition list",
			sortOrder:        "asc",
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
				if files[i] != want {
					t.Errorf("at %d: want %q, got %q", i, want, files[i])
				}
			}
		})
	}
}

func TestListFiles_FilteredPagination(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301", "202302"}
	files := []string{"a.json", "b.json", "c.json", "d.json"}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
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
			sortOrder:        "asc",
			filterPartitions: []string{"202301", "202302"},
			expectedPages: [][]string{
				{"202301/a.json", "202301/b.json", "202301/c.json"},
				{"202301/d.json", "202302/a.json", "202302/b.json"},
				{"202302/c.json", "202302/d.json"},
			},
		},
		{
			name:             "Filtered, paginated, desc",
			sortOrder:        "desc",
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
					if files[i] != want {
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

// --- Error and edge case tests ---

func TestListFiles_InvalidSortOrder(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301"}
	files := []string{"a.json"}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
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
}

func TestListFiles_NonExistentPartition(t *testing.T) {
	baseDir := t.TempDir()
	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	// Filtered mode, but partition does not exist
	files, nextPageToken, err := mds.ListFiles(
		ListingConfig{SortOrder: "asc", FilterPartitions: []string{"doesnotexist"}},
		"",
	)
	if err == nil {
		t.Fatalf("expected error for non-existent partition, got nil")
	}
	if len(files) != 0 {
		t.Fatalf("expected no files, got %v", files)
	}
	if nextPageToken != "" {
		t.Fatalf("expected no next page token, got %q", nextPageToken)
	}
}

func TestListFiles_UnreadablePartitionDir(t *testing.T) {
	baseDir := t.TempDir()
	partition := "202301"
	dir := filepath.Join(baseDir, partition)
	if err := os.MkdirAll(dir, 0o000); err != nil {
		t.Fatalf("failed to create unreadable dir: %v", err)
	}
	defer os.Chmod(dir, 0o755) // restore permissions for cleanup

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	_, _, err = mds.ListFiles(
		ListingConfig{SortOrder: "asc", FilterPartitions: []string{partition}},
		"",
	)
	if err == nil {
		t.Fatal("expected error for unreadable partition dir, got nil")
	}
}

func TestListFiles_InvalidPageToken(t *testing.T) {
	baseDir := t.TempDir()
	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	// Not base64
	_, _, err = mds.ListFiles(ListingConfig{SortOrder: "asc"}, "notbase64!")

	if err == nil {
		t.Fatal("expected error for invalid base64 page token, got nil")
	}

	// Base64 but not JSON
	bad := base64.StdEncoding.EncodeToString([]byte("notjson"))
	_, _, err = mds.ListFiles(ListingConfig{SortOrder: "asc"}, bad)
	if err == nil {
		t.Fatal("expected error for invalid JSON page token, got nil")
	}
}

func TestListFiles_EmptyBaseDir(t *testing.T) {
	baseDir := t.TempDir()
	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	files, nextPageToken, err := mds.ListFiles(ListingConfig{SortOrder: "asc"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no files, got %v", files)
	}
	if nextPageToken != "" {
		t.Fatalf("expected no next page token, got %q", nextPageToken)
	}
}

func TestListFiles_FilterWithNonExistentPartition(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301"}
	files := []string{"a.json"}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	// Filtered mode, one partition exists, one does not
	_, _, err = mds.ListFiles(
		ListingConfig{SortOrder: "asc", FilterPartitions: []string{"202301", "doesnotexist"}},
		"",
	)
	if err == nil {
		t.Fatal("expected error for non-existent partition in filter, got nil")
	}
}

func TestListFiles_PageSizeLargerThanFiles(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301"}
	files := []string{"a.json", "b.json"}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
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

	got, nextPageToken, err := mds.ListFiles(ListingConfig{SortOrder: "asc"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 files, got %d", len(got))
	}
	if nextPageToken != "" {
		t.Fatalf("expected no next page token, got %q", nextPageToken)
	}
}

func TestListFiles_EmptyFilterPartitions(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301"}
	files := []string{"a.json"}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	got, nextPageToken, err := mds.ListFiles(
		ListingConfig{SortOrder: "asc", FilterPartitions: []string{}},
		"",
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
}

func TestListFiles_CorruptedPageToken(t *testing.T) {
	baseDir := t.TempDir()
	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	// Corrupted JSON (valid base64)
	bad := base64.StdEncoding.EncodeToString([]byte("{notjson:"))
	_, _, err = mds.ListFiles(ListingConfig{SortOrder: "asc"}, bad)
	if err == nil {
		t.Fatal("expected error for corrupted JSON page token, got nil")
	}
}

func TestListFiles_FilteredPagination_EmptyPartition(t *testing.T) {
	baseDir := t.TempDir()

	files := []string{"a.json"}
	createFiles(t, baseDir, []string{"202301"}, files)
	// 202302 exists but is empty
	if err := os.MkdirAll(filepath.Join(baseDir, "202302"), 0o755); err != nil {
		t.Fatalf("failed to create partition dir: %v", err)
	}

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
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

	// Should list only the file in 202301, then nothing from 202302
	pageToken := ""
	files1, nextPageToken, err := mds.ListFiles(
		ListingConfig{SortOrder: "asc", FilterPartitions: []string{"202301", "202302"}},
		pageToken,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files1) != 1 || files1[0] != "202301/a.json" {
		t.Fatalf("expected [202301/a.json], got %v", files1)
	}
	if nextPageToken != "" {
		t.Fatalf("expected no next page token, got %q", nextPageToken)
	}
}

func TestListFiles_FilenamePrefixFiltering(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301", "202302"}
	files := []string{
		"apple.json", "apricot.json", "banana.json", "berry.json", "cherry.json",
		"apple_pie.json", "banana_bread.json", "berry_tart.json",
		"zebra.json",
	}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
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

	type want struct {
		files []string
	}
	tests := []struct {
		name             string
		sortOrder        string
		filterPartitions []string
		filenamePrefix   string
		want             want
	}{
		{
			name:           "No prefix, ascending",
			sortOrder:      "asc",
			filenamePrefix: "",
			want: want{files: []string{
				"202301/apple.json", "202301/apple_pie.json", "202301/apricot.json", "202301/banana.json", "202301/banana_bread.json", "202301/berry.json", "202301/berry_tart.json", "202301/cherry.json", "202301/zebra.json",
				"202302/apple.json", "202302/apple_pie.json", "202302/apricot.json", "202302/banana.json", "202302/banana_bread.json", "202302/berry.json", "202302/berry_tart.json", "202302/cherry.json", "202302/zebra.json",
			}},
		},
		{
			name:           "Prefix 'apple', ascending",
			sortOrder:      "asc",
			filenamePrefix: "apple",
			want: want{files: []string{
				"202301/apple.json", "202301/apple_pie.json",
				"202302/apple.json", "202302/apple_pie.json",
			}},
		},
		{
			name:           "Prefix 'banana', ascending",
			sortOrder:      "asc",
			filenamePrefix: "banana",
			want: want{files: []string{
				"202301/banana.json", "202301/banana_bread.json",
				"202302/banana.json", "202302/banana_bread.json",
			}},
		},
		{
			name:           "Prefix 'berry', descending",
			sortOrder:      "desc",
			filenamePrefix: "berry",
			want: want{files: []string{
				"202302/berry_tart.json", "202302/berry.json",
				"202301/berry_tart.json", "202301/berry.json",
			}},
		},
		{
			name:           "Prefix 'z', ascending",
			sortOrder:      "asc",
			filenamePrefix: "z",
			want: want{files: []string{
				"202301/zebra.json", "202302/zebra.json",
			}},
		},
		{
			name:           "Prefix 'notfound', ascending",
			sortOrder:      "asc",
			filenamePrefix: "notfound",
			want:           want{files: []string{}},
		},
		{
			name:             "Prefix '', filtered partition",
			sortOrder:        "asc",
			filterPartitions: []string{"202301"},
			filenamePrefix:   "",
			want: want{files: []string{
				"202301/apple.json", "202301/apple_pie.json", "202301/apricot.json", "202301/banana.json", "202301/banana_bread.json", "202301/berry.json", "202301/berry_tart.json", "202301/cherry.json", "202301/zebra.json",
			}},
		},
		{
			name:             "Prefix 'ap', filtered partition",
			sortOrder:        "asc",
			filterPartitions: []string{"202302"},
			filenamePrefix:   "ap",
			want: want{files: []string{
				"202302/apple.json", "202302/apple_pie.json", "202302/apricot.json",
			}},
		},
		{
			name:             "Prefix 'berry', filtered partition, descending",
			sortOrder:        "desc",
			filterPartitions: []string{"202301"},
			filenamePrefix:   "berry",
			want: want{files: []string{
				"202301/berry_tart.json", "202301/berry.json",
			}},
		},
		{
			name:           "Prefix with underscore",
			sortOrder:      "asc",
			filenamePrefix: "banana_",
			want: want{files: []string{
				"202301/banana_bread.json", "202302/banana_bread.json",
			}},
		},
		{
			name:           "Prefix with special char",
			sortOrder:      "asc",
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
				if files[i] != want {
					t.Errorf("at %d: want %q, got %q", i, want, files[i])
				}
			}
		})
	}
}

func TestListFiles_FilenamePrefixFiltering_Pagination(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301"}
	files := []string{
		"apple.json", "apricot.json", "banana.json", "berry.json", "cherry.json",
		"apple_pie.json", "banana_bread.json", "berry_tart.json",
		"zebra.json",
	}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
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

	type pageWant struct {
		files []string
	}
	tests := []struct {
		name           string
		sortOrder      string
		filenamePrefix string
		expectedPages  []pageWant
	}{
		{
			name:           "Prefix 'apple', ascending, paginated",
			sortOrder:      "asc",
			filenamePrefix: "apple",
			expectedPages: []pageWant{
				{files: []string{"202301/apple.json", "202301/apple_pie.json"}},
			},
		},
		{
			name:           "Prefix 'b', ascending, paginated",
			sortOrder:      "asc",
			filenamePrefix: "b",
			expectedPages: []pageWant{
				{files: []string{"202301/banana.json", "202301/banana_bread.json"}},
				{files: []string{"202301/berry.json", "202301/berry_tart.json"}},
			},
		},
		{
			name:           "Prefix 'berry', ascending, paginated",
			sortOrder:      "asc",
			filenamePrefix: "berry",
			expectedPages: []pageWant{
				{files: []string{"202301/berry.json", "202301/berry_tart.json"}},
			},
		},
		{
			name:           "Prefix 'z', ascending, paginated",
			sortOrder:      "asc",
			filenamePrefix: "z",
			expectedPages: []pageWant{
				{files: []string{"202301/zebra.json"}},
			},
		},
		{
			name:           "Prefix '', ascending, paginated",
			sortOrder:      "asc",
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
					if files[i] != wantFile {
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

func TestListFiles_FilenamePrefixFiltering_EmptyPartition(t *testing.T) {
	baseDir := t.TempDir()
	// 202301 has files, 202302 is empty
	files := []string{"apple.json", "banana.json"}
	createFiles(t, baseDir, []string{"202301"}, files)
	if err := os.MkdirAll(filepath.Join(baseDir, "202302"), 0o755); err != nil {
		t.Fatalf("failed to create partition dir: %v", err)
	}

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
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

	// Should list only the file in 202301 with prefix "apple", then nothing from 202302
	pageToken := ""
	files1, nextPageToken, err := mds.ListFiles(
		ListingConfig{
			SortOrder:        "asc",
			FilterPartitions: []string{"202301", "202302"},
			FilenamePrefix:   "apple",
		},
		pageToken,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files1) != 1 || files1[0] != "202301/apple.json" {
		t.Fatalf("expected [202301/apple.json], got %v", files1)
	}
	if nextPageToken != "" {
		t.Fatalf("expected no next page token, got %q", nextPageToken)
	}
}

func TestListFiles_FilenamePrefixFiltering_NoMatch(t *testing.T) {
	baseDir := t.TempDir()
	partitions := []string{"202301"}
	files := []string{"apple.json", "banana.json"}
	createFiles(t, baseDir, partitions, files)

	partitionProvider := &MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := NewMapDirectoryStore(
		baseDir,
		true,
		WithPartitionProvider(partitionProvider),
	)
	if err != nil {
		t.Fatalf("failed to create MapDirectoryStore: %v", err)
	}

	got, nextPageToken, err := mds.ListFiles(
		ListingConfig{SortOrder: "asc", FilenamePrefix: "zzz"},
		"",
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
}
