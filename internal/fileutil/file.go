package fileutil

import (
	"encoding/base64"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type ReadEncoding string

const (
	ReadEncodingText   ReadEncoding = "text"
	ReadEncodingBinary ReadEncoding = "binary"
)

type PathInfo struct {
	Path    string     `json:"path"`
	Name    string     `json:"name"`
	Exists  bool       `json:"exists"`
	IsDir   bool       `json:"isDir"`
	Size    int64      `json:"size,omitempty"`
	ModTime *time.Time `json:"modTime,omitempty"`
}

// ReadFile reads a file and returns its contents.
// encoding: "text" (default) or "binary" (base64-encoded output).
func ReadFile(path string, encoding ReadEncoding) (string, error) {
	if path == "" {
		return "", errors.New("path is required")
	}

	// Open the file (no write permissions!)
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	if encoding == ReadEncodingText {
		return string(data), nil
	}
	if encoding == ReadEncodingBinary {
		return base64.StdEncoding.EncodeToString(data), nil
	}
	return "", errors.New(`encoding must be "text" or "binary"`)
}

// StatPath returns basic metadata for the supplied path without mutating the filesystem.
// If the path does not exist, exists == false and err == nil.
func StatPath(path string) (pathInfo *PathInfo, err error) {
	if path == "" {
		return nil, errors.New("path is required")
	}

	pathInfo = &PathInfo{
		Path:   path,
		Exists: false,
	}

	info, e := os.Stat(path)
	if e != nil {
		if errors.Is(e, os.ErrNotExist) {
			return pathInfo, nil
		}
		return nil, e
	}

	p := getPathInfoFromFileInfo(path, info)
	return &p, nil
}

func getPathInfoFromFileInfo(path string, info fs.FileInfo) PathInfo {
	m := info.ModTime().UTC()
	return PathInfo{
		Path:    path,
		Name:    info.Name(),
		Exists:  true,
		IsDir:   info.IsDir(),
		Size:    info.Size(),
		ModTime: &m,
	}
}

// SearchFiles walks root (default ".") recursively and returns up to maxResults
// files whose *path* or UTF-8 text content matches the regexp pattern.
// If maxResults <= 0, it is treated as "no limit".
func SearchFiles(root, pattern string, maxResults int) ([]string, error) {
	if pattern == "" {
		return nil, errors.New("pattern is required")
	}
	if root == "" {
		root = "."
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	limit := maxResults
	if limit <= 0 {
		limit = int(^uint(0) >> 1) // effectively “infinite”
	}

	var matches []string
	walkFn := func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		if re.MatchString(path) {
			matches = append(matches, path)
		} else {
			// Check file content only for (reasonably small) text files.
			if info, _ := d.Info(); info != nil && info.Size() < 10*1024*1024 { // 10 MB guard
				if data, rerr := os.ReadFile(path); rerr == nil && re.Match(data) {
					matches = append(matches, path)
				}
			}
		}
		if len(matches) >= limit {
			return fs.SkipDir // stop walking
		}
		return nil
	}

	if err := filepath.WalkDir(root, walkFn); err != nil {
		return nil, err
	}
	return matches, nil
}
