package gotool

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ppipada/flexigpt-app/internal/fileutil"
	"github.com/ppipada/flexigpt-app/internal/tool/spec"
)

const (
	ReadFileFuncID      = "github.com/ppipada/flexigpt-app/internal/builtin/gotool/fs.ReadFile"
	ListDirectoryFuncID = "github.com/ppipada/flexigpt-app/internal/builtin/gotool/fs.ListDirectory"
	SearchFilesFuncID   = "github.com/ppipada/flexigpt-app/internal/builtin/gotool/fs.SearchFiles"
	StatPathFuncID      = "github.com/ppipada/flexigpt-app/internal/builtin/gotool/fs.StatPath"
	InspectImageFuncID  = "github.com/ppipada/flexigpt-app/internal/builtin/gotool/fs.InspectImage"
)

type ReadFileArgs struct {
	Path     string `json:"path"`               // required
	Encoding string `json:"encoding,omitempty"` // "text" (default) | "binary"
}

const maxReadBytes = 16 * 1024 * 1024 // 16MB safety limit

// ReadFile reads a file from disk and returns its contents.
// If Encoding == "binary" the output is base64-encoded.
func ReadFile(_ context.Context, args ReadFileArgs) ([]spec.ToolStoreOutputUnion, error) {
	// Normalize and validate encoding.
	enc := fileutil.ReadEncoding(strings.TrimSpace(args.Encoding))
	if enc == "" {
		enc = fileutil.ReadEncodingText
	}
	if enc != fileutil.ReadEncodingText && enc != fileutil.ReadEncodingBinary {
		return nil, errors.New(`encoding must be "text" or "binary"`)
	}

	path := strings.TrimSpace(args.Path)
	if path == "" {
		return nil, errors.New("path is required")
	}

	// Basic filesystem sanity checks.
	pi, err := fileutil.StatPath(path)
	if err != nil {
		return nil, err
	}
	if !pi.Exists {
		return nil, fmt.Errorf("path does not exist: %s", path)
	}
	if pi.IsDir {
		return nil, fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Detect MIME / extension where possible.
	mimeType, extMode, mimeErr := fileutil.MIMEForLocalFile(path)
	ext := strings.ToLower(filepath.Ext(path))

	if enc == fileutil.ReadEncodingText {
		// If we can't even detect the MIME, be conservative and refuse.
		if mimeErr != nil {
			return nil, fmt.Errorf("cannot read %q as text (MIME detection failed: %w)", path, mimeErr)
		}

		isPDF := ext == string(fileutil.ExtPDF) || mimeType == fileutil.MIMEApplicationPDF

		if isPDF {
			// PDF: use the same extraction logic as attachments.
			//
			// Implement this helper in fileutil, using the same underlying PDF
			// extraction code that getTextBlock/buildPDFTextOrFileBlock use.
			text, err := fileutil.ExtractPDFTextSafe(path, maxReadBytes)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(text) == "" {
				return nil, fmt.Errorf("no extractable text found in PDF %q", path)
			}

			return []spec.ToolStoreOutputUnion{
				{
					Kind: spec.ToolStoreOutputKindText,
					TextItem: &spec.ToolStoreOutputText{
						Text: text,
					},
				},
			}, nil
		}

		// Non‑PDF: only allow clearly text-like files.
		if extMode != fileutil.ExtensionModeText {
			return nil, fmt.Errorf(
				"cannot read non-text file %q as text; use encoding \"binary\" instead",
				path,
			)
		}

		// Normal text file: read and validate UTF‑8.
		data, err := fileutil.ReadFile(path, fileutil.ReadEncodingText)
		if err != nil {
			return nil, err
		}
		if !utf8.ValidString(data) {
			return nil, fmt.Errorf(
				"file %q is not valid UTF-8 text; use encoding \"binary\" instead",
				path,
			)
		}

		return []spec.ToolStoreOutputUnion{
			{
				Kind: spec.ToolStoreOutputKindText,
				TextItem: &spec.ToolStoreOutputText{
					Text: data,
				},
			},
		}, nil
	}

	// Binary mode: base64-encode and return, like before.
	data, err := fileutil.ReadFile(path, fileutil.ReadEncodingBinary)
	if err != nil {
		return nil, err
	}

	baseName := filepath.Base(path)
	if baseName == "" {
		baseName = "file"
	}

	// Prefer MIMEForLocalFile result if available; otherwise fall back to extension mapping.
	var mt string
	if mimeErr == nil && mimeType != "" {
		mt = string(mimeType)
	} else {
		if ext == "" {
			ext = strings.ToLower(filepath.Ext(baseName))
		}
		mt = mime.TypeByExtension(ext)
	}
	if mt == "" {
		mt = "application/octet-stream"
	}

	if strings.HasPrefix(mt, "image/") {
		return []spec.ToolStoreOutputUnion{
			{
				Kind: spec.ToolStoreOutputKindImage,
				ImageItem: &spec.ToolStoreOutputImage{
					Detail:    spec.ImageDetailAuto,
					ImageName: baseName,
					ImageMIME: mt,
					ImageData: data, // base64-encoded
				},
			},
		}, nil
	}

	return []spec.ToolStoreOutputUnion{
		{
			Kind: spec.ToolStoreOutputKindFile,
			FileItem: &spec.ToolStoreOutputFile{
				FileName: baseName,
				FileMIME: mt,
				FileData: data, // base64-encoded
			},
		},
	}, nil
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
		SizeBytes: pathInfo.Size,
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
	info, err := fileutil.ReadImage(args.Path, false)
	if err != nil {
		return nil, err
	}
	return &InspectImageOut{
		Exists:    info.Exists,
		Width:     info.Width,
		Height:    info.Height,
		Format:    info.Format,
		SizeBytes: info.Size,
		ModTime:   info.ModTime,
	}, nil
}
