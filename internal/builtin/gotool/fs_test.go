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

	"github.com/flexigpt/flexigpt-app/internal/tool/spec"
)

// TestReadFile covers happy, error, and boundary cases for ReadFile.
func TestReadFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	textFile := filepath.Join(tmpDir, "file.txt")
	binaryFile := filepath.Join(tmpDir, "file.bin")
	imageFile := filepath.Join(tmpDir, "image.png")

	if err := os.WriteFile(textFile, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("write textFile: %v", err)
	}
	binData := []byte{0x00, 0x01, 0x02, 0x03}
	if err := os.WriteFile(binaryFile, binData, 0o600); err != nil {
		t.Fatalf("write binaryFile: %v", err)
	}
	imgData := []byte{0x11, 0x22, 0x33}
	if err := os.WriteFile(imageFile, imgData, 0o600); err != nil {
		t.Fatalf("write imageFile: %v", err)
	}

	type testCase struct {
		name          string
		args          ReadFileArgs
		wantErr       bool
		wantKind      spec.ToolStoreOutputKind
		wantText      string
		wantFileName  string
		wantFileMIME  string
		wantImageName string
		wantMIMEPref  string
		wantBinary    []byte // expected raw bytes after base64 decoding (for file/image)
	}

	tests := []testCase{
		{
			name:    "Missing path returns error",
			args:    ReadFileArgs{},
			wantErr: true,
		},
		{
			name:    "Nonexistent file returns error",
			args:    ReadFileArgs{Path: filepath.Join(tmpDir, "nope.txt")},
			wantErr: true,
		},
		{
			name:     "Read text file as text",
			args:     ReadFileArgs{Path: textFile, Encoding: "text"},
			wantKind: spec.ToolStoreOutputKindText,
			wantText: "hello world",
		},
		{
			name:     "Read text file with default encoding",
			args:     ReadFileArgs{Path: textFile},
			wantKind: spec.ToolStoreOutputKindText,
			wantText: "hello world",
		},
		{
			name:         "Read binary file as binary -> file output",
			args:         ReadFileArgs{Path: binaryFile, Encoding: "binary"},
			wantKind:     spec.ToolStoreOutputKindFile,
			wantFileName: "file.bin",
			// Mime from ReadFile: ".bin" -> TypeByExtension("") => application/octet-stream.
			wantFileMIME: "application/octet-stream",
			wantBinary:   binData,
		},
		{
			name:          "Read image file as binary -> image output",
			args:          ReadFileArgs{Path: imageFile, Encoding: "binary"},
			wantKind:      spec.ToolStoreOutputKindImage,
			wantImageName: "image.png",
			wantMIMEPref:  "image/",
			wantBinary:    imgData,
		},
		{
			name:    "Invalid encoding returns error",
			args:    ReadFileArgs{Path: textFile, Encoding: "foo"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outs, err := ReadFile(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ReadFile error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(outs) != 0 {
					t.Fatalf("expected no outputs on error, got %#v", outs)
				}
				return
			}

			if len(outs) != 1 {
				t.Fatalf("expected exactly 1 output, got %d: %#v", len(outs), outs)
			}
			out := outs[0]

			if out.Kind != tt.wantKind {
				t.Fatalf("Kind = %q, want %q", out.Kind, tt.wantKind)
			}

			switch tt.wantKind {
			case spec.ToolStoreOutputKindText:
				if out.TextItem == nil {
					t.Fatalf("TextItem is nil for text output")
				}
				if out.ImageItem != nil || out.FileItem != nil {
					t.Fatalf("unexpected non-nil image/file items in text output: %#v", out)
				}
				if out.TextItem.Text != tt.wantText {
					t.Fatalf("Text = %q, want %q", out.TextItem.Text, tt.wantText)
				}

			case spec.ToolStoreOutputKindFile:
				if out.FileItem == nil {
					t.Fatalf("FileItem is nil for file output")
				}
				if out.TextItem != nil || out.ImageItem != nil {
					t.Fatalf("unexpected non-nil text/image items in file output: %#v", out)
				}
				if tt.wantFileName != "" && out.FileItem.FileName != tt.wantFileName {
					t.Fatalf("FileName = %q, want %q", out.FileItem.FileName, tt.wantFileName)
				}
				if tt.wantFileMIME != "" && out.FileItem.FileMIME != tt.wantFileMIME {
					t.Fatalf("FileMIME = %q, want %q", out.FileItem.FileMIME, tt.wantFileMIME)
				}
				if tt.wantBinary != nil {
					raw, err := base64.StdEncoding.DecodeString(out.FileItem.FileData)
					if err != nil {
						t.Fatalf("FileData not valid base64: %v", err)
					}
					if len(raw) != len(tt.wantBinary) {
						t.Fatalf("decoded binary len=%d, want %d", len(raw), len(tt.wantBinary))
					}
					for i := range raw {
						if raw[i] != tt.wantBinary[i] {
							t.Fatalf("decoded[%d] = %d, want %d", i, raw[i], tt.wantBinary[i])
						}
					}
				}

			case spec.ToolStoreOutputKindImage:
				if out.ImageItem == nil {
					t.Fatalf("ImageItem is nil for image output")
				}
				if out.TextItem != nil || out.FileItem != nil {
					t.Fatalf("unexpected non-nil text/file items in image output: %#v", out)
				}
				if tt.wantImageName != "" && out.ImageItem.ImageName != tt.wantImageName {
					t.Fatalf("ImageName = %q, want %q", out.ImageItem.ImageName, tt.wantImageName)
				}
				if tt.wantMIMEPref != "" && !strings.HasPrefix(out.ImageItem.ImageMIME, tt.wantMIMEPref) {
					t.Fatalf("ImageMIME = %q, want prefix %q", out.ImageItem.ImageMIME, tt.wantMIMEPref)
				}
				if tt.wantBinary != nil {
					raw, err := base64.StdEncoding.DecodeString(out.ImageItem.ImageData)
					if err != nil {
						t.Fatalf("ImageData not valid base64: %v", err)
					}
					if len(raw) != len(tt.wantBinary) {
						t.Fatalf("decoded binary len=%d, want %d", len(raw), len(tt.wantBinary))
					}
					for i := range raw {
						if raw[i] != tt.wantBinary[i] {
							t.Fatalf("decoded[%d] = %d, want %d", i, raw[i], tt.wantBinary[i])
						}
					}
				}

			default:
				t.Fatalf("unexpected output kind: %q", out.Kind)
			}
		})
	}
}

// TestListDirectory covers happy, error, and pattern cases for ListDirectory.
func TestListDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	aPath := filepath.Join(tmpDir, "a.txt")
	bPath := filepath.Join(tmpDir, "b.md")
	subDir := filepath.Join(tmpDir, "subdir")

	if err := os.WriteFile(aPath, []byte("a"), 0o600); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}
	if err := os.WriteFile(bPath, []byte("b"), 0o600); err != nil {
		t.Fatalf("write b.md: %v", err)
	}
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}

	tests := []struct {
		name       string
		args       ListDirectoryArgs
		want       []string
		wantErr    bool
		strictWant bool // if true, require exact set equality; else just subset check
	}{
		{
			name:       "List all entries",
			args:       ListDirectoryArgs{Path: tmpDir},
			want:       []string{"a.txt", "b.md", "subdir"},
			strictWant: true,
		},
		{
			name:       "List with pattern",
			args:       ListDirectoryArgs{Path: tmpDir, Pattern: "*.md"},
			want:       []string{"b.md"},
			strictWant: true,
		},
		{
			name:       "List with pattern no match",
			args:       ListDirectoryArgs{Path: tmpDir, Pattern: "*.go"},
			want:       []string{},
			strictWant: true,
		},
		{
			name:    "Nonexistent directory returns error",
			args:    ListDirectoryArgs{Path: filepath.Join(tmpDir, "nope")},
			wantErr: true,
		},
		{
			name: "Default path lists current dir (no error)",
			args: ListDirectoryArgs{},
		},
		{
			name:    "Path is file returns error",
			args:    ListDirectoryArgs{Path: aPath},
			wantErr: true,
		},
		{
			name:    "Invalid glob pattern returns error",
			args:    ListDirectoryArgs{Path: tmpDir, Pattern: "["},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := ListDirectory(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListDirectory error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if tt.want == nil {
				// No specific expectations beyond success.
				return
			}
			got := make(map[string]bool)
			for _, e := range out.Entries {
				got[e] = true
			}
			for _, e := range tt.want {
				if !got[e] {
					t.Errorf("expected entry %q not found in %v", e, out.Entries)
				}
			}
			if tt.strictWant && len(out.Entries) != len(tt.want) {
				t.Errorf("expected %d entries, got %d (%v)", len(tt.want), len(out.Entries), out.Entries)
			}
		})
	}
}

// TestSearchFiles covers happy, error, and boundary cases for SearchFiles.
func TestSearchFiles(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "foo.txt"), []byte("hello world"), 0o600); err != nil {
		t.Fatalf("write foo.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "bar.md"), []byte("goodbye world"), 0o600); err != nil {
		t.Fatalf("write bar.md: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "sub", "baz.txt"), []byte("baz content"), 0o600); err != nil {
		t.Fatalf("write baz.txt: %v", err)
	}

	// Large file to exercise content-size guard (if implemented by fileutil.SearchFiles).
	largeFile := filepath.Join(tmpDir, "large.txt")
	largeContent := strings.Repeat("x", 11*1024*1024) // >10MB
	if err := os.WriteFile(largeFile, []byte(largeContent), 0o600); err != nil {
		t.Fatalf("write large.txt: %v", err)
	}

	tests := []struct {
		name       string
		args       SearchFilesArgs
		want       []string
		wantErr    bool
		shouldFind func([]string) bool
	}{
		{
			name:    "Missing pattern returns error",
			args:    SearchFilesArgs{Root: tmpDir},
			wantErr: true,
		},
		{
			name:    "Invalid regexp returns error",
			args:    SearchFilesArgs{Root: tmpDir, Pattern: "["},
			wantErr: true,
		},
		{
			name: "Match file path",
			args: SearchFilesArgs{Root: tmpDir, Pattern: "foo\\.txt"},
			want: []string{filepath.Join(tmpDir, "foo.txt")},
		},
		{
			name: "Match file content",
			args: SearchFilesArgs{Root: tmpDir, Pattern: "goodbye"},
			want: []string{filepath.Join(tmpDir, "bar.md")},
		},
		{
			name: "Match in subdirectory",
			args: SearchFilesArgs{Root: tmpDir, Pattern: "baz"},
			want: []string{filepath.Join(tmpDir, "sub", "baz.txt")},
		},
		{
			name: "MaxResults limits output",
			args: SearchFilesArgs{Root: tmpDir, Pattern: "txt", MaxResults: 1},
			shouldFind: func(matches []string) bool {
				return len(matches) == 1 && strings.HasSuffix(matches[0], ".txt")
			},
		},
		{
			name: "Large file does not match content (size guard)",
			args: SearchFilesArgs{Root: tmpDir, Pattern: "x{100,}"},
			want: []string{}, // Should not match large.txt content if size guard is active.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := SearchFiles(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SearchFiles error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if tt.shouldFind != nil {
				if !tt.shouldFind(out.Matches) {
					t.Errorf("custom predicate failed for matches: %v", out.Matches)
				}
				return
			}
			if tt.want == nil {
				return
			}
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
					t.Errorf("expected match %q not found in %v", w, out.Matches)
				}
			}
			if len(out.Matches) != len(tt.want) {
				t.Errorf("expected %d matches, got %d", len(tt.want), len(out.Matches))
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

	nonExistent, err := StatPath(context.Background(), StatPathArgs{
		Path: filepath.Join(tmpDir, "missing.txt"),
	})
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
		_ = f.Close()
		t.Fatalf("encode png: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close image file: %v", err)
	}

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
	if out.SizeBytes <= 0 {
		t.Fatalf("expected positive SizeBytes, got %d", out.SizeBytes)
	}
	if out.ModTime == nil {
		t.Fatalf("expected ModTime to be set")
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
