package dirstore_test

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
)

func TestMapDirectoryStore(t *testing.T) {
	tests := []struct {
		name               string
		partitionProvider  dirstore.PartitionProvider
		filename           string
		data               map[string]any
		expectedPartition  string
		expectedFileExists bool
		expectError        bool
	}{
		{
			name:               "Default Partition - Create File",
			partitionProvider:  &dirstore.NoPartitionProvider{},
			filename:           "testfile.json",
			data:               map[string]any{"key": "value"},
			expectedPartition:  "",
			expectedFileExists: true,
			expectError:        false,
		},
		{
			name: "Month Partition - Create File",
			partitionProvider: &dirstore.MonthPartitionProvider{
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
			partitionProvider:  &dirstore.NoPartitionProvider{},
			filename:           "emptyfile.json",
			data:               map[string]any{},
			expectedPartition:  "",
			expectedFileExists: true,
			expectError:        false,
		},
		{
			name:               "Invalid Directory",
			partitionProvider:  &dirstore.NoPartitionProvider{},
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
			mds, err := dirstore.NewMapDirectoryStore(
				baseDir,
				true,
				dirstore.WithPartitionProvider(tt.partitionProvider),
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
	partitionProvider := &dirstore.MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := dirstore.NewMapDirectoryStore(
		baseDir,
		true,
		dirstore.WithPartitionProvider(partitionProvider),
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
			files, _, err := mds.ListFiles(tt.sortOrder, tt.pageToken)
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
	mds, err := dirstore.NewMapDirectoryStore(
		baseDir,
		true,
		dirstore.WithPartitionProvider(&dirstore.NoPartitionProvider{}),
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
	partitionProvider := &dirstore.MonthPartitionProvider{
		TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
	}
	mds, err := dirstore.NewMapDirectoryStore(
		baseDir,
		true,
		dirstore.WithPartitionProvider(partitionProvider),
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
				partitionProvider := &dirstore.MonthPartitionProvider{
					TimeFn: func(filename string) (time.Time, error) { return time.Now(), nil },
				}
				mds, err := dirstore.NewMapDirectoryStore(
					baseDir,
					true,
					dirstore.WithPartitionProvider(partitionProvider),
					dirstore.WithPageSize(tt.pageSize),
				)
				if err != nil {
					t.Fatalf("failed to create MapDirectoryStore: %v", err)
				}
				files, nextPageToken, err := mds.ListFiles(tt.sortOrder, pageToken)
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
			mds, err := dirstore.NewMapDirectoryStore(
				baseDir,
				true,
				dirstore.WithPageSize(tt.pageSize),
			)
			if err != nil {
				t.Fatalf("failed to create MapDirectoryStore: %v", err)
			}

			for pageIndex, expectedFiles := range tt.expectedPages {
				files, nextPageToken, err := mds.ListFiles(tt.sortOrder, pageToken)
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
