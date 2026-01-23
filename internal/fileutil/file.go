package fileutil

import (
	"encoding/base64"
	"errors"
	"io"
	"io/fs"
	"os"
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
