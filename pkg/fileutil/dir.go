package fileutil

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var errWalkLimitReached = errors.New("directory walk file limit reached")

const MaxTotalWalkFiles = 1024

// DirectoryOverflowInfo represents a directory subtree whose files were not included.
type DirectoryOverflowInfo struct {
	DirPath      string `json:"dirPath"`      // absolute path to that subdir
	RelativePath string `json:"relativePath"` // relative to the chosen root dir
	FileCount    int    `json:"fileCount"`    // number of files in that subtree
	TotalSize    int64  `json:"totalSize"`    // sum of file sizes in that subtree
}

// WalkDirectoryWithFilesResult is returned when user selects a directory.
type WalkDirectoryWithFilesResult struct {
	DirPath      string                  `json:"dirPath"`
	Files        []PathInfo              `json:"files"`        // included files (flattened)
	OverflowDirs []DirectoryOverflowInfo `json:"overflowDirs"` // dropped subtrees
	MaxFiles     int                     `json:"maxFiles"`     // max number of files returned (after clamping)
	TotalSize    int64                   `json:"totalSize"`    // sum of Files[i].Size
	HasMore      bool                    `json:"hasMore"`      // true if not all files included
}

// WalkDirectoryWithFiles implements:
// - full recursive walk over subdirs
// - if total files <= maxFiles: returns all
// - if total files > maxFiles: drops "largest subdir first" until within limit
// - skips dot-directories, dotfiles, and non-regular files (symlinks, devices, etc.)
// - returns flattened included files + summary of dropped subdirs.
func WalkDirectoryWithFiles(ctx context.Context, dirPath string, maxFiles int) (*WalkDirectoryWithFilesResult, error) {
	if maxFiles <= 0 || maxFiles > MaxTotalWalkFiles {
		maxFiles = MaxTotalWalkFiles
	}
	if dirPath == "" {
		// Empty path. Nothing to walk.
		return &WalkDirectoryWithFilesResult{
			DirPath:      "",
			Files:        []PathInfo{},
			OverflowDirs: nil,
			MaxFiles:     maxFiles,
			TotalSize:    0,
			HasMore:      false,
		}, nil
	}

	absRoot, err := filepath.Abs(dirPath)
	if err == nil {
		dirPath = absRoot
	}

	type scannedFile struct {
		PathInfo PathInfo
		RelPath  string
		DirKey   string
	}

	type dirAgg struct {
		FileCount int
		TotalSize int64
	}

	scanLimit := MaxTotalWalkFiles
	files := make([]scannedFile, 0, scanLimit)
	dirAggByKey := make(map[string]*dirAgg)
	totalFilesScanned := 0
	truncatedWalk := false

	walkErr := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Log and continue.
			slog.Debug("error while walking directory", "path", path, "error", walkErr)
			if path == dirPath {
				return walkErr
			}
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err // propagate context cancellation.
		}

		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				// Skip a dot dir and its subtree.
				return fs.SkipDir
			}
			// Don't need FileInfo for dir.
			return nil
		}

		info, err := d.Info()
		if err != nil {
			slog.Debug("stat error while walking directory", "path", path, "error", err)
			return nil
		}
		mode := info.Mode()
		if !mode.IsRegular() {
			// Skip symlinks, sockets, devices, fifos, etc.
			return nil
		}
		pInfo := getPathInfoFromFileInfo(path, info)

		if strings.HasPrefix(pInfo.Name, ".") {
			// Skip dotfiles.
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			relPath = path
		}

		dirKey := filepath.Dir(relPath)
		if dirKey == "." {
			dirKey = ""
		}

		files = append(files, scannedFile{
			PathInfo: pInfo,
			RelPath:  relPath,
			DirKey:   dirKey,
		})

		// Update aggregates for dirKey and all its ancestors up to root ("").
		curKey := dirKey
		for {
			agg := dirAggByKey[curKey]
			if agg == nil {
				agg = &dirAgg{}
				dirAggByKey[curKey] = agg
			}
			agg.FileCount++
			agg.TotalSize += pInfo.Size

			if curKey == "" {
				break
			}
			parent := filepath.Dir(curKey)
			if parent == "." {
				parent = ""
			}
			if parent == curKey {
				break
			}
			curKey = parent
		}
		totalFilesScanned++
		if totalFilesScanned >= scanLimit {
			truncatedWalk = true
			return errWalkLimitReached
		}

		return nil
	})

	if walkErr != nil && !errors.Is(walkErr, errWalkLimitReached) {
		return nil, walkErr
	}

	totalFiles := len(files)
	if totalFiles == 0 {
		return &WalkDirectoryWithFilesResult{
			DirPath:      dirPath,
			Files:        []PathInfo{},
			OverflowDirs: nil,
			MaxFiles:     maxFiles,
			TotalSize:    0,
			HasMore:      false,
		}, nil
	}

	// If fully within limit and not truncated, just return all.
	if totalFiles <= maxFiles && !truncatedWalk {
		resFiles := make([]PathInfo, 0, totalFiles)
		var totalSize int64
		for _, f := range files {
			totalSize += f.PathInfo.Size
			resFiles = append(resFiles, f.PathInfo)
		}
		return &WalkDirectoryWithFilesResult{
			DirPath:      dirPath,
			Files:        resFiles,
			OverflowDirs: nil,
			MaxFiles:     maxFiles,
			TotalSize:    totalSize,
			HasMore:      false,
		}, nil
	}

	// Need to drop subdirs: "largest subdir first" by file count, then total size.
	type dirEntry struct {
		Key       string
		FileCount int
		TotalSize int64
	}

	dirs := make([]dirEntry, 0, len(dirAggByKey))
	for key, agg := range dirAggByKey {
		if key == "" {
			continue // skip root
		}
		dirs = append(dirs, dirEntry{
			Key:       key,
			FileCount: agg.FileCount,
			TotalSize: agg.TotalSize,
		})
	}

	// Sort by descending file count, then descending total size.
	sort.Slice(dirs, func(i, j int) bool {
		if dirs[i].FileCount == dirs[j].FileCount {
			return dirs[i].TotalSize > dirs[j].TotalSize
		}
		return dirs[i].FileCount > dirs[j].FileCount
	})

	excludedKeys := make([]string, 0)
	remaining := totalFiles

	for _, de := range dirs {
		if remaining <= maxFiles {
			break
		}
		if underExcludedDir(de.Key, excludedKeys) {
			continue
		}
		excludedKeys = append(excludedKeys, de.Key)
		remaining -= de.FileCount
	}

	// Included files: all except those in excluded subtrees.
	included := make([]PathInfo, 0, maxFiles)
	var includedTotalSize int64

	for _, f := range files {
		if underExcludedDir(f.DirKey, excludedKeys) {
			continue
		}
		included = append(included, f.PathInfo)
		includedTotalSize += f.PathInfo.Size
	}

	overflowDirs := make([]DirectoryOverflowInfo, 0, len(excludedKeys)+1)

	// Root-level truncation if we *still* exceed maxFiles (lots of root files).
	if len(included) > maxFiles {
		sort.Slice(included, func(i, j int) bool {
			return included[i].Size < included[j].Size
		})
		truncated := included[maxFiles:]
		included = included[:maxFiles]

		var droppedCount int
		var droppedSize int64
		for _, fs := range truncated {
			droppedCount++
			droppedSize += fs.Size
			includedTotalSize -= fs.Size
		}

		if droppedCount > 0 {
			overflowDirs = append(overflowDirs, DirectoryOverflowInfo{
				DirPath:      dirPath,
				RelativePath: ".",
				FileCount:    droppedCount,
				TotalSize:    droppedSize,
			})
		}
	}

	// Overflow entries for excluded subdirs.
	for _, key := range excludedKeys {
		agg := dirAggByKey[key]
		if agg == nil {
			continue
		}
		overflowDirs = append(overflowDirs, DirectoryOverflowInfo{
			DirPath:      filepath.Join(dirPath, key),
			RelativePath: key,
			FileCount:    agg.FileCount,
			TotalSize:    agg.TotalSize,
		})
	}

	hasMore := truncatedWalk || len(overflowDirs) > 0 || len(included) < totalFiles

	return &WalkDirectoryWithFilesResult{
		DirPath:      dirPath,
		Files:        included,
		OverflowDirs: overflowDirs,
		MaxFiles:     maxFiles,
		TotalSize:    includedTotalSize,
		HasMore:      hasMore,
	}, nil
}

// ListDirectory lists files/dirs in path (default "."), pattern is an optional glob filter (filepath.Match).
func ListDirectory(path, pattern string) ([]string, error) {
	dir := path
	if dir == "" {
		dir = "."
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if pattern != "" {
			matched, _ := filepath.Match(pattern, name)
			if !matched {
				continue
			}
		}
		out = append(out, name)
	}
	return out, nil
}

func underExcludedDir(dirKey string, excluded []string) bool {
	for _, ex := range excluded {
		if dirKey == ex || strings.HasPrefix(dirKey, ex+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}
