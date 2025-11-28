package gotool

import (
	"context"
	"errors"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
)

const (
	ReadFileFuncID      = "github.com/ppipada/flexigpt-app/pkg/builtin/gotool/fs.ReadFile"
	ListDirectoryFuncID = "github.com/ppipada/flexigpt-app/pkg/builtin/gotool/fs.ListDirectory"
	SearchFilesFuncID   = "github.com/ppipada/flexigpt-app/pkg/builtin/gotool/fs.SearchFiles"
	StatPathFuncID      = "github.com/ppipada/flexigpt-app/pkg/builtin/gotool/fs.StatPath"
	InspectImageFuncID  = "github.com/ppipada/flexigpt-app/pkg/builtin/gotool/fs.InspectImage"
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
	enc := fileutil.ReadEncoding(args.Encoding)
	if enc == "" {
		enc = fileutil.ReadEncodingText
	}
	if enc != fileutil.ReadEncodingText && enc != fileutil.ReadEncodingBinary {
		return nil, errors.New(`encoding must be "text" or "binary"`)
	}
	data, err := fileutil.ReadFile(args.Path, enc)
	if err != nil {
		return nil, err
	}
	return &ReadFileOut{Content: data}, nil
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
	entries, err := fileutil.ListDirectory(args.Path, args.Pattern)
	if err != nil {
		return nil, err
	}
	return &ListDirectoryOut{Entries: entries}, nil
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
	matches, err := fileutil.SearchFiles(args.Root, args.Pattern, args.MaxResults)
	if err != nil {
		return nil, err
	}
	return &SearchFilesOut{Matches: matches}, nil
}

type StatPathArgs struct {
	Path string `json:"path"`
}

type StatPathOut struct {
	Exists    bool       `json:"exists"`
	IsDir     bool       `json:"isDir"`
	SizeBytes int64      `json:"sizeBytes,omitempty"`
	ModTime   *time.Time `json:"modTime,omitempty"`
}

// StatPath returns basic metadata for the supplied path without mutating the file system.
func StatPath(_ context.Context, args StatPathArgs) (*StatPathOut, error) {
	pathInfo, err := fileutil.StatPath(args.Path)
	if err != nil {
		return nil, err
	}
	return &StatPathOut{
		Exists:    pathInfo.Exists,
		IsDir:     pathInfo.IsDir,
		SizeBytes: pathInfo.SizeBytes,
		ModTime:   pathInfo.ModTime,
	}, nil
}

type InspectImageArgs struct {
	Path string `json:"path"`
}

type InspectImageOut struct {
	Exists    bool       `json:"exists"`
	Width     int        `json:"width,omitempty"`
	Height    int        `json:"height,omitempty"`
	Format    string     `json:"format,omitempty"`
	SizeBytes int64      `json:"sizeBytes,omitempty"`
	ModTime   *time.Time `json:"modTime,omitempty"`
}

// InspectImage inspects an image file and returns its intrinsic metadata.
func InspectImage(ctx context.Context, args InspectImageArgs) (*InspectImageOut, error) {
	info, err := fileutil.ReadImageInfo(args.Path, false)
	if err != nil {
		return nil, err
	}
	return &InspectImageOut{
		Exists:    info.Exists,
		Width:     info.Width,
		Height:    info.Height,
		Format:    info.Format,
		SizeBytes: info.SizeBytes,
		ModTime:   info.ModTime,
	}, nil
}
