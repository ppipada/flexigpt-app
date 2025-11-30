package fileutil

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

// Structure describing the tree used in SearchFiles tests.
type searchTestTree struct {
	root        string
	helloPath   string
	matchPath   string
	contentPath string
	nestedPath  string
	bigPath     string
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()

	textPath := filepath.Join(dir, "sample.txt")
	textContent := "Hello, world!\n"
	writeFile(t, textPath, textContent)

	nonExistentPath := filepath.Join(dir, "does_not_exist.txt")
	binaryContent := base64.StdEncoding.EncodeToString([]byte(textContent))

	tests := []struct {
		name            string
		path            string
		encoding        ReadEncoding
		want            string
		wantErr         bool
		wantErrContains string
		wantIsNotExist  bool
	}{
		{
			name:     "text encoding returns raw content",
			path:     textPath,
			encoding: ReadEncodingText,
			want:     textContent,
		},
		{
			name:     "binary encoding returns base64",
			path:     textPath,
			encoding: ReadEncodingBinary,
			want:     binaryContent,
		},
		{
			name:            "invalid encoding",
			path:            textPath,
			encoding:        ReadEncoding("invalid"),
			wantErr:         true,
			wantErrContains: `encoding must be "text" or "binary"`,
		},
		{
			name:            "empty encoding (zero value) is invalid",
			path:            textPath,
			encoding:        ReadEncoding(""),
			wantErr:         true,
			wantErrContains: `encoding must be "text" or "binary"`,
		},
		{
			name:            "empty path",
			path:            "",
			encoding:        ReadEncodingText,
			wantErr:         true,
			wantErrContains: "path is required",
		},
		{
			name:           "non-existent path",
			path:           nonExistentPath,
			encoding:       ReadEncodingText,
			wantErr:        true,
			wantIsNotExist: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ReadFile(tc.path, tc.encoding)

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
			if got != tc.want {
				t.Errorf("ReadFile(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

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

			got := getPathInfoFromFileInfo(tc.path, info)

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

func TestSearchFilesBasic(t *testing.T) {
	tree := createSearchTestTree(t)

	tests := []struct {
		name            string
		root            string
		pattern         string
		maxResults      int
		wantExactPaths  []string // compare as set if non-nil
		allowedPaths    []string // each result must be one of these (for limit tests)
		wantLen         int      // -1 means "don't check"; overrides len(wantExactPaths) when set
		wantErr         bool
		wantErrContains string
	}{
		{
			name:           "match by filename only",
			root:           tree.root,
			pattern:        "matchname",
			maxResults:     0,
			wantLen:        -1,
			wantExactPaths: []string{tree.matchPath},
		},
		{
			name:           "match by content only",
			root:           tree.root,
			pattern:        "CONTENTPATTERN",
			maxResults:     0,
			wantLen:        -1,
			wantExactPaths: []string{tree.contentPath, tree.nestedPath},
		},
		{
			name:       "maxResults limits number of matches",
			root:       tree.root,
			pattern:    "CONTENTPATTERN",
			maxResults: 1,
			allowedPaths: []string{
				tree.contentPath,
				tree.nestedPath,
			},
			wantLen: 1,
		},
		{
			name:            "pattern is required",
			root:            tree.root,
			pattern:         "",
			maxResults:      0,
			wantErr:         true,
			wantErrContains: "pattern is required",
		},
		{
			name:       "invalid regexp pattern",
			root:       tree.root,
			pattern:    "(",
			maxResults: 0,
			wantErr:    true,
		},
		{
			name:       "non-existent root returns error",
			root:       filepath.Join(tree.root, "no_such_dir"),
			pattern:    "anything",
			maxResults: 0,
			wantErr:    true,
		},
		{
			name:           "maxResults zero treated as unlimited",
			root:           tree.root,
			pattern:        "txt",
			maxResults:     0,
			wantExactPaths: []string{tree.helloPath, tree.matchPath, tree.contentPath, tree.nestedPath},
			wantLen:        -1,
		},
		{
			name:           "negative maxResults treated as unlimited",
			root:           tree.root,
			pattern:        "txt",
			maxResults:     -10,
			wantExactPaths: []string{tree.helloPath, tree.matchPath, tree.contentPath, tree.nestedPath},
			wantLen:        -1,
		},
		{
			name:           "large files are not searched by content",
			root:           tree.root,
			pattern:        "BIGPATTERN",
			maxResults:     0,
			wantExactPaths: []string{},
			wantLen:        -1,
		},
		{
			name:           "root can be a single file",
			root:           tree.helloPath,
			pattern:        "hello",
			maxResults:     0,
			wantExactPaths: []string{tree.helloPath},
			wantLen:        -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SearchFiles(tc.root, tc.pattern, tc.maxResults)

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

			if tc.wantLen >= 0 && len(got) != tc.wantLen {
				t.Fatalf("len(results) = %d, want %d", len(got), tc.wantLen)
			}

			if tc.wantExactPaths != nil {
				if !equalStringSets(got, tc.wantExactPaths) {
					t.Fatalf("results = %#v, want %#v (order-independent)", got, tc.wantExactPaths)
				}
			}

			if tc.allowedPaths != nil {
				for _, p := range got {
					if !containsString(tc.allowedPaths, p) {
						t.Fatalf("result %q not in allowed set %#v", p, tc.allowedPaths)
					}
				}
			}
		})
	}
}

func TestSearchFilesRootDefaultUsesCWD(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "cwdfile.txt")
	writeFile(t, filePath, "some pattern content")

	t.Chdir(root)

	tests := []struct {
		name           string
		pattern        string
		maxResults     int
		wantExactPaths []string
	}{
		{
			name:           "empty root searches current directory by path",
			pattern:        "cwdfile",
			maxResults:     0,
			wantExactPaths: []string{"cwdfile.txt"},
		},
		{
			name:           "empty root searches current directory by content",
			pattern:        "pattern content",
			maxResults:     0,
			wantExactPaths: []string{"cwdfile.txt"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SearchFiles("", tc.pattern, tc.maxResults)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !equalStringSets(got, tc.wantExactPaths) {
				t.Fatalf("results = %#v, want %#v (order-independent)", got, tc.wantExactPaths)
			}
		})
	}
}

func TestSearchFilesConcurrency(t *testing.T) {
	tree := createSearchTestTree(t)

	tests := []struct {
		name        string
		goroutines  int
		iterations  int
		searchRoot  string
		searchPat   string
		expectedSet []string
	}{
		{
			name:        "concurrent searches on same tree",
			goroutines:  8,
			iterations:  5,
			searchRoot:  tree.root,
			searchPat:   "CONTENTPATTERN",
			expectedSet: []string{tree.contentPath, tree.nestedPath},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup
			errCh := make(chan error, tc.goroutines)

			for i := 0; i < tc.goroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < tc.iterations; j++ {
						got, err := SearchFiles(tc.searchRoot, tc.searchPat, 0)
						if err != nil {
							errCh <- fmt.Errorf("goroutine %d: unexpected error: %w", id, err)
							return
						}
						if !equalStringSets(got, tc.expectedSet) {
							errCh <- fmt.Errorf("goroutine %d: unexpected results: %#v, want %#v",
								id, got, tc.expectedSet)
							return
						}
					}
				}(i)
			}

			wg.Wait()
			close(errCh)

			for err := range errCh {
				if err != nil {
					t.Fatalf("concurrent SearchFiles error: %v", err)
				}
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

// Build a deterministic directory tree for SearchFiles tests.
func createSearchTestTree(t *testing.T) searchTestTree {
	t.Helper()

	root := t.TempDir()

	helloPath := filepath.Join(root, "hello.txt")
	writeFile(t, helloPath, "hello world")

	matchPath := filepath.Join(root, "matchname.txt")
	writeFile(t, matchPath, "this file path will match; its content will not")

	contentPath := filepath.Join(root, "content.txt")
	writeFile(t, contentPath, "some CONTENTPATTERN in this file.")

	subdir := filepath.Join(root, "subdir")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("failed to create subdir %q: %v", subdir, err)
	}

	nestedPath := filepath.Join(subdir, "nested.txt")
	writeFile(t, nestedPath, "nested CONTENTPATTERN plus more text.")

	// Big file (>10MB) whose content contains BIGPATTERN but should not be read
	// by SearchFiles because of the size guard.
	bigPath := filepath.Join(subdir, "bigfile.bin")
	f, err := os.Create(bigPath)
	if err != nil {
		t.Fatalf("failed to create big file %q: %v", bigPath, err)
	}
	if _, err := f.WriteString("BIGPATTERN at the beginning of a big file."); err != nil {
		t.Fatalf("failed to write to big file %q: %v", bigPath, err)
	}
	if err := f.Truncate(10*1024*1024 + 1); err != nil {
		t.Fatalf("failed to truncate big file %q: %v", bigPath, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close big file %q: %v", bigPath, err)
	}

	return searchTestTree{
		root:        root,
		helloPath:   helloPath,
		matchPath:   matchPath,
		contentPath: contentPath,
		nestedPath:  nestedPath,
		bigPath:     bigPath,
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

// Helper to compare string slices as sets (order-independent).
func equalStringSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for _, s := range a {
		m[s]++
	}
	for _, s := range b {
		m[s]--
		if m[s] < 0 {
			return false
		}
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

// Helper to check if a slice contains a string.
func containsString(slice []string, target string) bool {
	return slices.Contains(slice, target)
}
