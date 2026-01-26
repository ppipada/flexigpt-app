package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatPath(t *testing.T) {
	dir := t.TempDir()

	filePath := filepath.Join(dir, "file.txt")
	fileContent := "some content"
	writeFile(t, filePath, fileContent)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat test file: %v", err)
	}

	subdirPath := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(subdirPath, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	nonExistentPath := filepath.Join(dir, "missing.txt")

	tests := []struct {
		name            string
		path            string
		wantExists      bool
		wantIsDir       bool
		wantName        string
		wantSize        int64 // -1 means "don't check"
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "empty path",
			path:            "",
			wantErr:         true,
			wantErrContains: "path is required",
		},
		{
			name:       "non-existent path",
			path:       nonExistentPath,
			wantExists: false,
			wantIsDir:  false,
			wantName:   "",
			wantSize:   0,
		},
		{
			name:       "existing file",
			path:       filePath,
			wantExists: true,
			wantIsDir:  false,
			wantName:   filepath.Base(filePath),
			wantSize:   fileInfo.Size(),
		},
		{
			name:       "existing directory",
			path:       subdirPath,
			wantExists: true,
			wantIsDir:  true,
			wantName:   filepath.Base(subdirPath),
			wantSize:   -1, // don't assert exact size (OS-dependent)
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := StatPath(tc.path)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.wantErrContains != "" && !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErrContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected non-nil PathInfo")
			}

			if got.Exists != tc.wantExists {
				t.Errorf("Exists = %v, want %v", got.Exists, tc.wantExists)
			}
			if got.IsDir != tc.wantIsDir {
				t.Errorf("IsDir = %v, want %v", got.IsDir, tc.wantIsDir)
			}
			if tc.wantName != "" && got.Name != tc.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tc.wantName)
			}
			if tc.wantSize >= 0 && got.Size != tc.wantSize {
				t.Errorf("Size = %d, want %d", got.Size, tc.wantSize)
			}

			if tc.wantExists {
				if got.ModTime == nil {
					t.Errorf("ModTime is nil, want non-nil for existing path")
				}
			} else {
				if got.ModTime != nil {
					t.Errorf("ModTime is non-nil for non-existent path")
				}
			}
		})
	}
}

func TestGetPathInfoFromFileInfo(t *testing.T) {
	dir := t.TempDir()

	filePath := filepath.Join(dir, "file.txt")
	writeFile(t, filePath, "file content")
	if err := os.MkdirAll(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	dirPath := filepath.Join(dir, "subdir")

	tests := []struct {
		name  string
		path  string
		isDir bool
	}{
		{
			name:  "file info",
			path:  filePath,
			isDir: false,
		},
		{
			name:  "directory info",
			path:  dirPath,
			isDir: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, err := os.Stat(tc.path)
			if err != nil {
				t.Fatalf("stat failed: %v", err)
			}

			got := GetPathInfoFromFileInfo(tc.path, info)

			if got.Path != tc.path {
				t.Errorf("Path = %q, want %q", got.Path, tc.path)
			}
			if got.Name != info.Name() {
				t.Errorf("Name = %q, want %q", got.Name, info.Name())
			}
			if !got.Exists {
				t.Errorf("Exists = false, want true")
			}
			if got.IsDir != tc.isDir {
				t.Errorf("IsDir = %v, want %v", got.IsDir, tc.isDir)
			}
			if got.Size != info.Size() {
				t.Errorf("Size = %d, want %d", got.Size, info.Size())
			}
			if got.ModTime == nil {
				t.Fatalf("ModTime is nil, want non-nil")
			}
			if !got.ModTime.Equal(info.ModTime().UTC()) {
				t.Errorf("ModTime = %v, want %v", got.ModTime, info.ModTime().UTC())
			}
		})
	}
}

func TestSniffFileMIME(t *testing.T) {
	dir := t.TempDir()

	emptyPath := filepath.Join(dir, "empty.txt")
	mustWriteBytes(t, emptyPath, []byte{})

	textPath := filepath.Join(dir, "text.txt")
	writeFile(t, textPath, "Hello, world!\n")

	utf8Path := filepath.Join(dir, "utf8.txt")
	writeFile(t, utf8Path, "Привет, мир!\n") // UTF-8 text

	binaryPath := filepath.Join(dir, "binary.png")
	// Minimal PNG header; DetectContentType should recognize this as image/png.
	pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	mustWriteBytes(t, binaryPath, pngHeader)

	nonExistentPath := filepath.Join(dir, "no_such_file")

	tests := []struct {
		name            string
		path            string
		wantMIME        string
		wantIsText      bool
		wantErr         bool
		wantErrContains string
		wantIsNotExist  bool
	}{
		{
			name:            "empty path",
			path:            "",
			wantErr:         true,
			wantErrContains: "invalid path",
		},
		{
			name:           "non-existent path",
			path:           nonExistentPath,
			wantErr:        true,
			wantIsNotExist: true,
		},
		{
			name:       "empty file treated as text/plain",
			path:       emptyPath,
			wantMIME:   "text/plain; charset=utf-8",
			wantIsText: true,
		},
		{
			name:       "ASCII text file",
			path:       textPath,
			wantMIME:   "text/plain; charset=utf-8",
			wantIsText: true,
		},
		{
			name:       "UTF-8 text file",
			path:       utf8Path,
			wantMIME:   "text/plain; charset=utf-8",
			wantIsText: true,
		},
		{
			name:       "binary PNG file",
			path:       binaryPath,
			wantMIME:   "image/png",
			wantIsText: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mime, mode, err := SniffFileMIME(tc.path)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.wantErrContains != "" && !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErrContains)
				}
				if tc.wantIsNotExist && !os.IsNotExist(err) {
					t.Fatalf("expected a not-exist error, got: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.wantMIME != "" && mime != MIMEType(tc.wantMIME) {
				t.Errorf("MIME = %q, want %q", mime, tc.wantMIME)
			}
			isText := mode == ExtensionModeText
			if isText != tc.wantIsText {
				t.Errorf("isText = %v, want %v", isText, tc.wantIsText)
			}
		})
	}
}

func TestIsProbablyTextSample(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "empty slice is text",
			data: nil,
			want: true,
		},
		{
			name: "simple ASCII text",
			data: []byte("Hello, world!"),
			want: true,
		},
		{
			name: "text with allowed control characters",
			data: []byte("line1\nline2\tend\r"),
			want: true,
		},
		{
			name: "contains NUL byte",
			data: []byte{'a', 0x00, 'b'},
			want: false,
		},
		{
			name: "too many control characters",
			data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, // many control chars
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isProbablyTextSample(tc.data)
			if got != tc.want {
				t.Errorf("isProbablyTextSample(%v) = %v, want %v", tc.data, got, tc.want)
			}
		})
	}
}

// Helper to write text files in tests.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file %q: %v", path, err)
	}
}

// Helper to write binary files in tests.
func mustWriteBytes(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("failed to write test file %q: %v", path, err)
	}
}
