package gotool

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadFile covers happy, error, and boundary cases for ReadFile.
func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	textFile := filepath.Join(tmpDir, "file.txt")
	binaryFile := filepath.Join(tmpDir, "file.bin")
	_ = os.WriteFile(textFile, []byte("hello world"), 0o600)
	_ = os.WriteFile(binaryFile, []byte{0x00, 0x01, 0x02, 0x03}, 0o600)

	tests := []struct {
		name     string
		args     ReadFileArgs
		want     string
		wantErr  bool
		isBinary bool
	}{
		{
			name:    "Missing path returns error.",
			args:    ReadFileArgs{},
			wantErr: true,
		},
		{
			name:    "Nonexistent file returns error.",
			args:    ReadFileArgs{Path: filepath.Join(tmpDir, "nope.txt")},
			wantErr: true,
		},
		{
			name: "Read text file as text.",
			args: ReadFileArgs{Path: textFile, Encoding: "text"},
			want: "hello world",
		},
		{
			name: "Read text file as default encoding.",
			args: ReadFileArgs{Path: textFile},
			want: "hello world",
		},
		{
			name:     "Read binary file as binary.",
			args:     ReadFileArgs{Path: binaryFile, Encoding: "binary"},
			want:     base64.StdEncoding.EncodeToString([]byte{0x00, 0x01, 0x02, 0x03}),
			isBinary: true,
		},
		{
			name:    "Invalid encoding returns error.",
			args:    ReadFileArgs{Path: textFile, Encoding: "foo"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := ReadFile(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error: %v, got: %v", tt.wantErr, err)
			}
			if err == nil && out.Content != tt.want {
				t.Errorf("Expected content %q, got %q", tt.want, out.Content)
			}
			if tt.isBinary && err == nil {
				// Check that output is valid base64.
				if _, err := base64.StdEncoding.DecodeString(out.Content); err != nil {
					t.Errorf("Output is not valid base64: %v", err)
				}
			}
		})
	}
}

// TestListDirectory covers happy, error, and pattern cases for ListDirectory.
func TestListDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0o600)
	_ = os.WriteFile(filepath.Join(tmpDir, "b.md"), []byte("b"), 0o600)
	_ = os.Mkdir(filepath.Join(tmpDir, "subdir"), 0o755)

	tests := []struct {
		name    string
		args    ListDirectoryArgs
		want    []string
		wantErr bool
	}{
		{
			name: "List all entries.",
			args: ListDirectoryArgs{Path: tmpDir},
			want: []string{"a.txt", "b.md", "subdir"},
		},
		{
			name: "List with pattern.",
			args: ListDirectoryArgs{Path: tmpDir, Pattern: "*.md"},
			want: []string{"b.md"},
		},
		{
			name: "List with pattern no match.",
			args: ListDirectoryArgs{Path: tmpDir, Pattern: "*.go"},
			want: []string{},
		},
		{
			name:    "Nonexistent directory returns error.",
			args:    ListDirectoryArgs{Path: filepath.Join(tmpDir, "nope")},
			wantErr: true,
		},
		{
			name: "Default path lists current dir.",
			args: ListDirectoryArgs{},
			// Can't predict entries, just check no error.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := ListDirectory(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error: %v, got: %v", tt.wantErr, err)
			}
			if err == nil && tt.want != nil {
				got := make(map[string]bool)
				for _, e := range out.Entries {
					got[e] = true
				}
				for _, e := range tt.want {
					if !got[e] {
						t.Errorf("Expected entry %q not found in %v", e, out.Entries)
					}
				}
				if len(out.Entries) != len(tt.want) {
					t.Errorf("Expected %d entries, got %d", len(tt.want), len(out.Entries))
				}
			}
		})
	}
}

// TestSearchFiles covers happy, error, and boundary cases for SearchFiles.
func TestSearchFiles(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "foo.txt"), []byte("hello world"), 0o600)
	_ = os.WriteFile(filepath.Join(tmpDir, "bar.md"), []byte("goodbye world"), 0o600)
	_ = os.Mkdir(filepath.Join(tmpDir, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(tmpDir, "sub", "baz.txt"), []byte("baz content"), 0o600)
	largeFile := filepath.Join(tmpDir, "large.txt")
	largeContent := strings.Repeat("x", 11*1024*1024)
	_ = os.WriteFile(largeFile, []byte(largeContent), 0o600)

	tests := []struct {
		name       string
		args       SearchFilesArgs
		want       []string
		wantErr    bool
		shouldFind func([]string) bool
	}{
		{
			name:    "Missing pattern returns error.",
			args:    SearchFilesArgs{Root: tmpDir},
			wantErr: true,
		},
		{
			name:    "Invalid regexp returns error.",
			args:    SearchFilesArgs{Root: tmpDir, Pattern: "["},
			wantErr: true,
		},
		{
			name: "Match file path.",
			args: SearchFilesArgs{Root: tmpDir, Pattern: "foo\\.txt"},
			want: []string{filepath.Join(tmpDir, "foo.txt")},
		},
		{
			name: "Match file content.",
			args: SearchFilesArgs{Root: tmpDir, Pattern: "goodbye"},
			want: []string{filepath.Join(tmpDir, "bar.md")},
		},
		{
			name: "Match in subdirectory.",
			args: SearchFilesArgs{Root: tmpDir, Pattern: "baz"},
			want: []string{filepath.Join(tmpDir, "sub", "baz.txt")},
		},
		{
			name: "MaxResults limits output.",
			args: SearchFilesArgs{Root: tmpDir, Pattern: "txt", MaxResults: 1},
			shouldFind: func(matches []string) bool {
				return len(matches) == 1 && strings.HasSuffix(matches[0], ".txt")
			},
		},
		{
			name: "Large file does not match content.",
			args: SearchFilesArgs{Root: tmpDir, Pattern: "x{100,}"},
			want: []string{}, // Should not match large.txt content due to size guard.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := SearchFiles(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error: %v, got: %v", tt.wantErr, err)
			}
			if err == nil {
				if tt.shouldFind != nil {
					if !tt.shouldFind(out.Matches) {
						t.Errorf("Custom match function failed for matches: %v", out.Matches)
					}
				} else if tt.want != nil {
					wantMap := make(map[string]bool)
					for _, w := range tt.want {
						wantMap[w] = true
					}
					gotMap := make(map[string]bool)
					for _, g := range out.Matches {
						gotMap[g] = true
					}
					for w := range wantMap {
						if !gotMap[w] {
							t.Errorf("Expected match %q not found in %v", w, out.Matches)
						}
					}
					if len(out.Matches) != len(tt.want) {
						t.Errorf("Expected %d matches, got %d", len(tt.want), len(out.Matches))
					}
				}
			}
		})
	}
}
