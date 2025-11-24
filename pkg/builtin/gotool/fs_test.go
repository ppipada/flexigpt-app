package gotool

import (
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
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

func TestStatPath(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sample.txt")
	if err := os.WriteFile(filePath, []byte("hi"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	res, err := StatPath(context.Background(), StatPathArgs{Path: filePath})
	if err != nil {
		t.Fatalf("StatPath returned error: %v", err)
	}
	if !res.Exists || res.IsDir {
		t.Fatalf("expected file to exist and not be dir: %+v", res)
	}
	if res.SizeBytes != 2 {
		t.Fatalf("expected size 2, got %d", res.SizeBytes)
	}
	if res.ModTime == nil {
		t.Fatalf("expected mod time to be set")
	}
	dirRes, err := StatPath(context.Background(), StatPathArgs{Path: tmpDir})
	if err != nil {
		t.Fatalf("StatPath dir error: %v", err)
	}
	if !dirRes.Exists || !dirRes.IsDir {
		t.Fatalf("expected dir to exist and be dir: %+v", dirRes)
	}
	nonExistent, err := StatPath(context.Background(), StatPathArgs{Path: filepath.Join(tmpDir, "missing.txt")})
	if err != nil {
		t.Fatalf("StatPath missing error: %v", err)
	}
	if nonExistent.Exists {
		t.Fatalf("expected missing path to report Exists=false")
	}
	if _, err := StatPath(context.Background(), StatPathArgs{}); err == nil {
		t.Fatalf("expected error for empty path")
	}
}

func TestInspectImage(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "img.png")
	img := image.NewRGBA(image.Rect(0, 0, 8, 6))
	for y := range 6 {
		for x := range 8 {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatalf("encode png: %v", err)
	}
	_ = f.Close()
	out, err := InspectImage(context.Background(), InspectImageArgs{Path: imgPath})
	if err != nil {
		t.Fatalf("InspectImage error: %v", err)
	}
	if !out.Exists {
		t.Fatalf("expected image to exist")
	}
	if out.Width != 8 || out.Height != 6 {
		t.Fatalf("unexpected dimensions: %+v", out)
	}
	if out.Format != "png" {
		t.Fatalf("expected png format, got %q", out.Format)
	}
	textPath := filepath.Join(tmpDir, "notimg.txt")
	if err := os.WriteFile(textPath, []byte("plain"), 0o600); err != nil {
		t.Fatalf("write text: %v", err)
	}
	if _, err := InspectImage(context.Background(), InspectImageArgs{Path: textPath}); err == nil {
		t.Fatalf("expected error for non-image file")
	}
	if _, err := InspectImage(context.Background(), InspectImageArgs{Path: tmpDir}); err == nil {
		t.Fatalf("expected error for directory path")
	}
	if _, err := InspectImage(context.Background(), InspectImageArgs{}); err == nil {
		t.Fatalf("expected error for empty path")
	}
}
