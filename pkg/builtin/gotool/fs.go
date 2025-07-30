package gotool

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ReadFileArgs struct {
	Path     string `json:"path"`               // required
	Encoding string `json:"encoding,omitempty"` // "text" (default) | "binary"
}

type ReadFileOut struct {
	Content string `json:"content"`
}

// ReadFile reads a file from disk and returns its contents.
// If Encoding == "binary" the output is base64-encoded.
func ReadFile(_ context.Context, args ReadFileArgs) (*ReadFileOut, error) {
	if args.Path == "" {
		return nil, errors.New("path is required")
	}
	// Open the file (no write permissions!)
	f, err := os.Open(args.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	enc := strings.ToLower(args.Encoding)
	if enc == "" || enc == "text" {
		return &ReadFileOut{Content: string(data)}, nil
	}
	if enc == "binary" {
		return &ReadFileOut{Content: base64.StdEncoding.EncodeToString(data)}, nil
	}
	return nil, errors.New(`encoding must be "text" or "binary"`)
}

type ListDirectoryArgs struct {
	Path    string `json:"path,omitempty"`    // default "."
	Pattern string `json:"pattern,omitempty"` // Optional glob
}
type ListDirectoryOut struct {
	Entries []string `json:"entries"`
}

// ListDirectory lists files / dirs in Path. If Pattern is supplied, the
// results are filtered via filepath.Match.
func ListDirectory(_ context.Context, args ListDirectoryArgs) (*ListDirectoryOut, error) {
	dir := args.Path
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
		if args.Pattern != "" {
			matched, _ := filepath.Match(args.Pattern, name)
			if !matched {
				continue
			}
		}
		out = append(out, name)
	}
	return &ListDirectoryOut{Entries: out}, nil
}

type SearchFilesArgs struct {
	Root       string `json:"root,omitempty"` // default "."
	Pattern    string `json:"pattern"`        // required (RE2)
	MaxResults int    `json:"maxResults,omitempty"`
}
type SearchFilesOut struct {
	Matches []string `json:"matches"`
}

// SearchFiles walks Root (recursively) and returns up to MaxResults files
// whose *path* or *UTF-8 text content* match the supplied regexp.
func SearchFiles(_ context.Context, args SearchFilesArgs) (*SearchFilesOut, error) {
	if args.Pattern == "" {
		return nil, errors.New("pattern is required")
	}
	root := args.Root
	if root == "" {
		root = "."
	}
	re, err := regexp.Compile(args.Pattern)
	if err != nil {
		return nil, err
	}
	limit := args.MaxResults
	if limit <= 0 {
		limit = int(^uint(0) >> 1) // “infinite”
	}

	var matches []string
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
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
	return &SearchFilesOut{Matches: matches}, nil
}
