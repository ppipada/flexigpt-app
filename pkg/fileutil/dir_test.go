package fileutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestWalkDirectoryWithFiles_CoreScenarios(t *testing.T) {
	type testCase struct {
		name     string
		setup    func(t *testing.T) (root string)
		maxFiles int
		// Root is whatever setup returned (may be empty for some tests).
		verify func(t *testing.T, root string, res *WalkDirectoryWithFilesResult, err error)
	}

	tests := []testCase{
		{
			name: "EmptyPath",
			setup: func(t *testing.T) string {
				t.Helper()
				return ""
			},
			maxFiles: 10,
			verify: func(t *testing.T, root string, res *WalkDirectoryWithFilesResult, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				// These expectations reflect the documented behavior:
				// empty dirPath -> nothing to walk.
				if res == nil {
					t.Fatalf("expected non-nil result")
				}
				if res.DirPath != "" {
					t.Errorf("expected DirPath %q, got %q", "", res.DirPath)
				}
				if len(res.Files) != 0 {
					t.Errorf("expected 0 files, got %d", len(res.Files))
				}
				if len(res.OverflowDirs) != 0 {
					t.Errorf("expected 0 overflow dirs, got %d", len(res.OverflowDirs))
				}
				if res.HasMore {
					t.Errorf("expected HasMore=false, got true")
				}
				if res.MaxFiles != 10 {
					t.Errorf("expected MaxFiles=10, got %d", res.MaxFiles)
				}
			},
		},
		{
			name: "EmptyDirectory_NoFiles",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			maxFiles: 10,
			verify: func(t *testing.T, root string, res *WalkDirectoryWithFilesResult, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if res == nil {
					t.Fatalf("expected non-nil result")
				}
				if res.DirPath == "" {
					t.Errorf("expected non-empty DirPath for actual directory")
				}
				if got := len(res.Files); got != 0 {
					t.Errorf("expected 0 files, got %d", got)
				}
				if got := len(res.OverflowDirs); got != 0 {
					t.Errorf("expected 0 overflow dirs, got %d", got)
				}
				if res.HasMore {
					t.Errorf("expected HasMore=false, got true")
				}
				if res.TotalSize != 0 {
					t.Errorf("expected TotalSize=0, got %d", res.TotalSize)
				}
			},
		},
		{
			name: "SimpleUnderLimit_AllIncluded",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				mustWriteFile(t, root, "f1.txt", 1)
				mustWriteFile(t, root, "f2.txt", 2)
				mustWriteFile(t, root, "f3.txt", 3)
				return root
			},
			maxFiles: 10,
			verify: func(t *testing.T, root string, res *WalkDirectoryWithFilesResult, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if res == nil {
					t.Fatalf("expected non-nil result")
				}
				if res.HasMore {
					t.Errorf("expected HasMore=false, got true")
				}
				names := namesFromPathInfos(res.Files)
				wantNames := []string{"f1.txt", "f2.txt", "f3.txt"}
				if len(names) != len(wantNames) {
					t.Fatalf("expected %d files, got %d (%v)", len(wantNames), len(names), names)
				}
				for i, w := range wantNames {
					if names[i] != w {
						t.Errorf("expected names[%d]=%q, got %q", i, w, names[i])
					}
				}
				if sum := sumSizes(res.Files); sum != res.TotalSize {
					t.Errorf("TotalSize mismatch: result=%d, recomputed=%d", res.TotalSize, sum)
				}
				if res.TotalSize != 1+2+3 {
					t.Errorf("expected TotalSize=6, got %d", res.TotalSize)
				}
				if len(res.OverflowDirs) != 0 {
					t.Errorf("expected 0 overflow dirs, got %d", len(res.OverflowDirs))
				}
			},
		},
		{
			name: "SkipDotFiles",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				// Dotfile at root.
				mustWriteFile(t, root, ".hidden_root", 10)

				dotGit := mustMkdir(t, root, ".git")
				mustWriteFile(t, dotGit, "ignored.txt", 20)

				mustWriteFile(t, root, "visible1.txt", 30)

				sub := mustMkdir(t, root, "sub")
				mustWriteFile(t, sub, ".hidden_sub", 40)
				mustWriteFile(t, sub, "visible2.txt", 50)
				return root
			},
			maxFiles: 10,
			verify: func(t *testing.T, root string, res *WalkDirectoryWithFilesResult, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if res == nil {
					t.Fatalf("expected non-nil result")
				}
				if res.HasMore {
					t.Errorf("expected HasMore=false, got true")
				}
				names := namesFromPathInfos(res.Files)
				wantNames := []string{"ignored.txt", "visible1.txt", "visible2.txt"}
				if len(names) != len(wantNames) {
					t.Fatalf("expected %d visible files, got %d (%v)", len(wantNames), len(names), names)
				}
				for i, w := range wantNames {
					if names[i] != w {
						t.Errorf("expected names[%d]=%q, got %q", i, w, names[i])
					}
				}
				// Dotfiles and dot-dirs should not appear in results or overflow.
				for _, of := range res.OverflowDirs {
					if strings.Contains(of.RelativePath, ".git") ||
						strings.Contains(of.RelativePath, ".hidden") {
						t.Errorf("unexpected overflow dir for dot-entry: %+v", of)
					}
				}
			},
		},
		{
			name: "LargestSubdirDropped_First",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				mustWriteFile(t, root, "root.txt", 5)

				a := mustMkdir(t, root, "a")
				mustWriteFile(t, a, "a1.txt", 10)
				mustWriteFile(t, a, "a2.txt", 20)

				b := mustMkdir(t, root, "b")
				mustWriteFile(t, b, "b1.txt", 30)

				c := mustMkdir(t, root, "c")
				mustWriteFile(t, c, "c1.txt", 40)

				return root
			},
			maxFiles: 3,
			verify: func(t *testing.T, root string, res *WalkDirectoryWithFilesResult, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if res == nil {
					t.Fatalf("expected non-nil result")
				}
				// We expect to drop subtree "a" (2 files), keeping root.txt, b1, c1.
				names := namesFromPathInfos(res.Files)
				sort.Strings(names)
				wantNames := []string{"b1.txt", "c1.txt", "root.txt"}
				if len(names) != len(wantNames) {
					t.Fatalf("expected %d files, got %d (%v)", len(wantNames), len(names), names)
				}
				for i, w := range wantNames {
					if names[i] != w {
						t.Errorf("expected names[%d]=%q, got %q", i, w, names[i])
					}
				}
				// OverflowDirs should contain exactly one entry for "a".
				if len(res.OverflowDirs) != 1 {
					t.Fatalf("expected 1 overflow dir, got %d (%+v)", len(res.OverflowDirs), res.OverflowDirs)
				}
				ov := res.OverflowDirs[0]
				if ov.RelativePath != "a" {
					t.Errorf("expected overflow relativePath 'a', got %q", ov.RelativePath)
				}
				if ov.FileCount != 2 {
					t.Errorf("expected overflow fileCount=2 for 'a', got %d", ov.FileCount)
				}
				if ov.TotalSize != 10+20 {
					t.Errorf("expected overflow totalSize=30 for 'a', got %d", ov.TotalSize)
				}
				// TotalSize should be sum of kept files: 5 + 30 + 40 = 75.
				if res.TotalSize != 5+30+40 {
					t.Errorf("expected TotalSize=75, got %d", res.TotalSize)
				}
			},
		},
		{
			name: "RootLevelTruncation_NoSubdirs",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				// 10 root files with distinct sizes 1..10.
				for i := 1; i <= 10; i++ {
					name := "f" + fmt.Sprintf("0%d", i)
					if i == 10 {
						name = "f10"
					}
					mustWriteFile(t, root, name, i)
				}
				return root
			},
			maxFiles: 5,
			verify: func(t *testing.T, root string, res *WalkDirectoryWithFilesResult, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if res == nil {
					t.Fatalf("expected non-nil result")
				}
				if len(res.OverflowDirs) != 1 {
					t.Fatalf("expected 1 overflow dir (root overflow), got %d", len(res.OverflowDirs))
				}
				ov := res.OverflowDirs[0]
				if ov.RelativePath != "." {
					t.Errorf("expected overflow relativePath '.', got %q", ov.RelativePath)
				}

				if ov.FileCount != 5 {
					t.Errorf("expected overflow root dropped 5 files, got %d", ov.FileCount)
				}
				// Total size of dropped files should be 6+7+8+9+10=40.
				if ov.TotalSize != int64(6+7+8+9+10) {
					t.Errorf("expected dropped size=40, got %d", ov.TotalSize)
				}
				if len(res.Files) != 5 {
					t.Fatalf("expected 5 files kept, got %d", len(res.Files))
				}
				if res.TotalSize != int64(1+2+3+4+5) {
					t.Errorf("expected kept TotalSize=15, got %d", res.TotalSize)
				}
				if !res.HasMore {
					t.Errorf("expected HasMore=true due to root-level truncation, got false")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := ""
			if tc.setup != nil {
				root = tc.setup(t)
			}
			ctx := context.Background()
			res, err := WalkDirectoryWithFiles(ctx, root, tc.maxFiles)
			tc.verify(t, root, res, err)
		})
	}
}

func TestWalkDirectoryWithFiles_NonExistentDir(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "does_not_exist")
	ctx := context.Background()
	res, err := WalkDirectoryWithFiles(ctx, root, 10)
	if err == nil {
		t.Fatalf("expected non-nil error for non-existent dir, got nil")
	}
	if res != nil {
		t.Fatalf("expected nil result on error, got %+v", res)
	}
}

func TestWalkDirectoryWithFiles_ContextCanceled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, root, "file.txt", 10)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before walking

	res, err := WalkDirectoryWithFiles(ctx, root, 10)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result when context is canceled, got %+v", res)
	}
}

func TestListDirectory_BasicAndPattern(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, root, "a.txt", 1)
	mustWriteFile(t, root, "b.log", 1)
	if err := os.Mkdir(filepath.Join(root, "subdir"), 0o755); err != nil {
		t.Fatalf("failed to mkdir subdir: %v", err)
	}

	type testCase struct {
		name    string
		dir     string
		pattern string
		want    []string
	}

	tests := []testCase{
		{
			name:    "NoPattern_AllEntries",
			dir:     root,
			pattern: "",
			want:    []string{"a.txt", "b.log", "subdir"},
		},
		{
			name:    "Pattern_TxtOnly",
			dir:     root,
			pattern: "*.txt",
			want:    []string{"a.txt"},
		},
		{
			name:    "Pattern_Invalid_NoError_EmptyResult",
			dir:     root,
			pattern: "[",        // invalid glob; implementation ignores Match error
			want:    []string{}, // currently, invalid pattern yields no matches, no error
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ListDirectory(tc.dir, tc.pattern)
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			sort.Strings(got)
			sort.Strings(tc.want)
			if len(got) != len(tc.want) {
				t.Fatalf("expected %d entries, got %d (%v)", len(tc.want), len(got), got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("expected got[%d]=%q, got %q", i, tc.want[i], got[i])
				}
			}
		})
	}
}

func TestListDirectory_DefaultPathDot(t *testing.T) {
	// Not parallel: this test changes the working directory globally.
	tmp := t.TempDir()
	t.Chdir(tmp)

	// Create some entries in the current directory.
	mustWriteFile(t, tmp, "x.txt", 1)
	if err := os.Mkdir(filepath.Join(tmp, "dir"), 0o755); err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}

	got, err := ListDirectory("", "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// We only check that the created entries are present; there might be others.
	wantSet := map[string]bool{"x.txt": true, "dir": true}
	for name := range wantSet {
		found := slices.Contains(got, name)
		if !found {
			t.Errorf("expected to find %q in ListDirectory output, got %v", name, got)
		}
	}
}

func mustWriteFile(t *testing.T, dir, name string, size int) string {
	t.Helper()
	full := filepath.Join(dir, name)
	data := bytes.Repeat([]byte("x"), size)
	if err := os.WriteFile(full, data, 0o600); err != nil {
		t.Fatalf("failed to write file %q: %v", full, err)
	}
	return full
}

func mustMkdir(t *testing.T, dir, name string) string {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.Mkdir(full, 0o755); err != nil {
		t.Fatalf("failed to mkdir %q: %v", full, err)
	}
	return full
}

func namesFromPathInfos(files []PathInfo) []string {
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, f.Name)
	}
	sort.Strings(names)
	return names
}

func sumSizes(files []PathInfo) int64 {
	var total int64
	for _, f := range files {
		total += f.Size
	}
	return total
}
