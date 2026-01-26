package attachment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
			mime, mode, err := sniffFileMIME(tc.path)

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
