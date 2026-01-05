package fileutil

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidPath      = errors.New("invalid path")
	ErrUnknownExtension = errors.New("unknown extension")
)

type ExtensionMode string

const (
	ExtensionModeText     ExtensionMode = "text"
	ExtensionModeImage    ExtensionMode = "image"
	ExtensionModeDocument ExtensionMode = "document"
	ExtensionModeDefault  ExtensionMode = "default"
)

var AllExtensionModes = []ExtensionMode{
	ExtensionModeText,
	ExtensionModeImage,
	ExtensionModeDocument,
	ExtensionModeDefault,
}

type FileExt string

const (
	ExtTxt      FileExt = ".txt"
	ExtMd       FileExt = ".md"
	ExtMarkdown FileExt = ".markdown"
	ExtLog      FileExt = ".log"
	ExtJSON     FileExt = ".json"
	ExtYAML     FileExt = ".yaml"
	ExtYML      FileExt = ".yml"
	ExtTOML     FileExt = ".toml"
	ExtJS       FileExt = ".js"
	ExtTS       FileExt = ".ts"
	ExtTSX      FileExt = ".tsx"
	ExtJSX      FileExt = ".jsx"
	ExtPY       FileExt = ".py"
	ExtGO       FileExt = ".go"
	ExtRS       FileExt = ".rs"
	ExtJAVA     FileExt = ".java"
	ExtC        FileExt = ".c"
	ExtCPP      FileExt = ".cpp"
	ExtH        FileExt = ".h"
	ExtHPP      FileExt = ".hpp"
	ExtCS       FileExt = ".cs"
	ExtRB       FileExt = ".rb"
	ExtPHP      FileExt = ".php"
	ExtHTML     FileExt = ".html"
	ExtHTM      FileExt = ".htm"
	ExtCSS      FileExt = ".css"
	ExtSCSS     FileExt = ".scss"
	ExtLESS     FileExt = ".less"
	ExtSQL      FileExt = ".sql"
	ExtMod      FileExt = ".mod"
	ExtSum      FileExt = ".sum"
	ExtJSONL    FileExt = ".jsonl"
	ExtShell    FileExt = ".sh"
	ExtSWIFT    FileExt = ".swift"
	ExtM        FileExt = ".m"
	ExtKT       FileExt = ".kt"
	ExtPL       FileExt = ".pl"
	ExtSCALA    FileExt = ".scala"
	ExtHS       FileExt = ".hs"
	ExtLUA      FileExt = ".lua"
	ExtDART     FileExt = ".dart"

	ExtJPG  FileExt = ".jpg"
	ExtJPEG FileExt = ".jpeg"
	ExtPNG  FileExt = ".png"
	ExtGIF  FileExt = ".gif"
	ExtWEBP FileExt = ".webp"
	ExtBMP  FileExt = ".bmp"
	ExtSVG  FileExt = ".svg"

	ExtPDF  FileExt = ".pdf"
	ExtDOC  FileExt = ".doc"
	ExtDOCX FileExt = ".docx"
	ExtPPT  FileExt = ".ppt"
	ExtPPTX FileExt = ".pptx"
	ExtXLS  FileExt = ".xls"
	ExtXLSX FileExt = ".xlsx"
	ExtODT  FileExt = ".odt"
	ExtODS  FileExt = ".ods"
)

type MIMEType string

const (
	MIMEEmpty                  MIMEType = ""
	MIMEApplicationOctetStream MIMEType = "application/octet-stream"

	MIMETextPlain       MIMEType = "text/plain; charset=utf-8"
	MIMETextMarkdown    MIMEType = "text/markdown; charset=utf-8"
	MIMETextHTML        MIMEType = "text/html; charset=utf-8"
	MIMETextCSS         MIMEType = "text/css; charset=utf-8"
	MIMEApplicationJSON MIMEType = "application/json"
	MIMEApplicationYAML MIMEType = "application/x-yaml"
	MIMEApplicationTOML MIMEType = "application/toml"
	MIMEApplicationSQL  MIMEType = "application/sql"
	MIMEApplicationJS   MIMEType = "application/javascript"

	MIMEImageJPEG MIMEType = "image/jpeg"
	MIMEImagePNG  MIMEType = "image/png"
	MIMEImageGIF  MIMEType = "image/gif"
	MIMEImageWEBP MIMEType = "image/webp"
	MIMEImageBMP  MIMEType = "image/bmp"
	MIMEImageSVG  MIMEType = "image/svg+xml"

	MIMEApplicationPDF        MIMEType = "application/pdf"
	MIMEApplicationMSWord     MIMEType = "application/msword"
	MIMEApplicationMSPowerPt  MIMEType = "application/vnd.ms-powerpoint"
	MIMEApplicationMSExcel    MIMEType = "application/vnd.ms-excel"
	MIMEApplicationOpenXMLDoc MIMEType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	MIMEApplicationOpenXMLPPT MIMEType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	MIMEApplicationOpenXMLXLS MIMEType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	MIMEApplicationODT        MIMEType = "application/vnd.oasis.opendocument.text"
	MIMEApplicationODS        MIMEType = "application/vnd.oasis.opendocument.spreadsheet"
)

const DefaultImageMIME = MIMEImagePNG

var MIMETypeToExtensionMode = map[MIMEType]ExtensionMode{
	MIMEEmpty:                  ExtensionModeDefault,
	MIMEApplicationOctetStream: ExtensionModeDefault,

	MIMETextPlain:       ExtensionModeText,
	MIMETextMarkdown:    ExtensionModeText,
	MIMETextHTML:        ExtensionModeText,
	MIMETextCSS:         ExtensionModeText,
	MIMEApplicationJSON: ExtensionModeText,
	MIMEApplicationYAML: ExtensionModeText,
	MIMEApplicationTOML: ExtensionModeText,
	MIMEApplicationSQL:  ExtensionModeText,
	MIMEApplicationJS:   ExtensionModeText,

	MIMEImageJPEG: ExtensionModeImage,
	MIMEImagePNG:  ExtensionModeImage,
	MIMEImageGIF:  ExtensionModeImage,
	MIMEImageWEBP: ExtensionModeImage,
	MIMEImageBMP:  ExtensionModeImage,
	MIMEImageSVG:  ExtensionModeImage,

	MIMEApplicationPDF:        ExtensionModeDocument,
	MIMEApplicationMSWord:     ExtensionModeDocument,
	MIMEApplicationMSPowerPt:  ExtensionModeDocument,
	MIMEApplicationMSExcel:    ExtensionModeDocument,
	MIMEApplicationOpenXMLDoc: ExtensionModeDocument,
	MIMEApplicationOpenXMLPPT: ExtensionModeDocument,
	MIMEApplicationOpenXMLXLS: ExtensionModeDocument,
	MIMEApplicationODT:        ExtensionModeDocument,
	MIMEApplicationODS:        ExtensionModeDocument,
}

var ExtensionToMIMEType = map[FileExt]MIMEType{
	ExtTxt:      MIMETextPlain,
	ExtMd:       MIMETextMarkdown,
	ExtMarkdown: MIMETextMarkdown,
	ExtLog:      MIMETextPlain,
	ExtJSON:     MIMEApplicationJSON,
	ExtYAML:     MIMEApplicationYAML,
	ExtYML:      MIMEApplicationYAML,
	ExtTOML:     MIMEApplicationTOML,
	ExtJS:       MIMEApplicationJS,
	ExtTS:       MIMETextPlain,
	ExtTSX:      MIMETextPlain,
	ExtJSX:      MIMETextPlain,
	ExtPY:       MIMETextPlain,
	ExtGO:       MIMETextPlain,
	ExtRS:       MIMETextPlain,
	ExtJAVA:     MIMETextPlain,
	ExtC:        MIMETextPlain,
	ExtCPP:      MIMETextPlain,
	ExtH:        MIMETextPlain,
	ExtHPP:      MIMETextPlain,
	ExtCS:       MIMETextPlain,
	ExtRB:       MIMETextPlain,
	ExtPHP:      MIMETextPlain,
	ExtHTML:     MIMETextHTML,
	ExtHTM:      MIMETextHTML,
	ExtCSS:      MIMETextCSS,
	ExtSCSS:     MIMETextPlain,
	ExtLESS:     MIMETextPlain,
	ExtSQL:      MIMEApplicationSQL,
	ExtMod:      MIMETextPlain,
	ExtSum:      MIMETextPlain,
	ExtJSONL:    MIMETextPlain,
	ExtShell:    MIMETextPlain,
	ExtSWIFT:    MIMETextPlain,
	ExtM:        MIMETextPlain,
	ExtKT:       MIMETextPlain,
	ExtPL:       MIMETextPlain,
	ExtSCALA:    MIMETextPlain,
	ExtHS:       MIMETextPlain,
	ExtLUA:      MIMETextPlain,
	ExtDART:     MIMETextPlain,

	ExtJPG:  MIMEImageJPEG,
	ExtJPEG: MIMEImageJPEG,
	ExtPNG:  MIMEImagePNG,
	ExtGIF:  MIMEImageGIF,
	ExtWEBP: MIMEImageWEBP,
	ExtBMP:  MIMEImageBMP,
	ExtSVG:  MIMEImageSVG,

	ExtPDF:  MIMEApplicationPDF,
	ExtDOC:  MIMEApplicationMSWord,
	ExtDOCX: MIMEApplicationOpenXMLDoc,
	ExtPPT:  MIMEApplicationMSPowerPt,
	ExtPPTX: MIMEApplicationOpenXMLPPT,
	ExtXLS:  MIMEApplicationMSExcel,
	ExtXLSX: MIMEApplicationOpenXMLXLS,
	ExtODT:  MIMEApplicationODT,
	ExtODS:  MIMEApplicationODS,
}

var ModeToExtensions = func() map[ExtensionMode][]FileExt {
	m := make(map[ExtensionMode][]FileExt, len(AllExtensionModes))

	for ext, mimeType := range ExtensionToMIMEType {
		if mode, ok := MIMETypeToExtensionMode[mimeType]; ok {
			m[mode] = append(m[mode], ext)
		} else {
			m[ExtensionModeDefault] = append(m[ExtensionModeDefault], ext)
		}
	}
	return m
}()

// FileFilter is used to build file dialog filters (Patterns like "*.png;*.jpg").
type FileFilter struct {
	DisplayName string
	Extensions  []FileExt // normalized, e.g. ".png".
}

// Pattern converts a FileFilter into a semicolon-separated pattern string.
// W.g. "*.png;*.jpg".
func (f FileFilter) Pattern() string {
	if len(f.Extensions) == 0 {
		return "*"
	}
	out := make([]string, 0, len(f.Extensions))
	for _, ext := range f.Extensions {
		e := strings.TrimSpace(string(ext))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		out = append(out, "*"+e)
	}
	if len(out) == 0 {
		return "*"
	}
	return strings.Join(out, ";")
}

var (
	FileFilterAllFiles = FileFilter{
		DisplayName: "All Files",
		Extensions:  nil, // Pattern() => "*".
	}

	FileFilterTextMarkdown = FileFilter{
		DisplayName: "Text",
		Extensions:  ModeToExtensions[ExtensionModeText],
	}

	FileFilterDocuments = FileFilter{
		DisplayName: "Documents",
		Extensions:  ModeToExtensions[ExtensionModeDocument],
	}

	FileFilterImages = FileFilter{
		DisplayName: "Images",
		Extensions:  ModeToExtensions[ExtensionModeImage],
	}

	DefaultFileFilters = []FileFilter{
		FileFilterAllFiles,
		FileFilterTextMarkdown,
		FileFilterDocuments,
		FileFilterImages,
	}
)

// MIMEForLocalFile returns a best-effort MIME type, it tries to see if extension based mime can be detected.
// If not, it tries to sniff mime using magic bits from file.
func MIMEForLocalFile(path string) (mimeType MIMEType, mode ExtensionMode, err error) {
	if path == "" {
		return MIMEEmpty, ExtensionModeDefault, ErrInvalidPath
	}
	ext := filepath.Ext(path)
	if ext == "" {
		// No extension: fall back to sniffing.
		return SniffFileMIME(path)
	}

	// First try extension-based detection.
	mimeType, err = MIMEFromExtensionString(ext)
	if err == nil && mimeType != MIMEApplicationOctetStream && mimeType != MIMEEmpty {
		if m, ok := MIMETypeToExtensionMode[mimeType]; ok {
			return mimeType, m, nil
		}
		// Heuristic: treat any text/* media type as text mode even if we don't
		// have an explicit entry in MIMETypeToExtensionMode (e.g. text/csv).
		if strings.HasPrefix(string(mimeType), "text/") {
			return mimeType, ExtensionModeText, nil
		}
		// Known MIME type but no explicit mode mapping and not obviously text.
		return mimeType, ExtensionModeDefault, nil
	}

	// Unknown / generic : fall back to sniffing.
	return SniffFileMIME(path)
}

// MIMEFromExtensionString returns the best-known MIME for a given extension string.
// Lookup order is: internal registry -> mime.TypeByExtension.
// If the extension cannot be resolved, it returns MIMEApplicationOctetStream and ErrUnknownExtension.
func MIMEFromExtensionString(ext string) (mimeType MIMEType, err error) {
	if ext == "" {
		return MIMEEmpty, ErrInvalidPath
	}
	e := normalizeExt(ext)
	if mimeType, ok := ExtensionToMIMEType[e]; ok {
		return mimeType, nil
	}

	if t := mime.TypeByExtension(string(e)); t != "" {
		return MIMEType(t), nil
	}

	// Unknown extension: report a generic binary MIME type and an error.
	return MIMEApplicationOctetStream, ErrUnknownExtension
}

// SniffFileMIME inspects the first bytes of a file and returns a best-effort
// MIME type plus a coarse-grained mode (text, image, document, or default).
//
// It uses net/http.DetectContentType and a simple heuristic to decide whether
// the content is probably text vs binary (NUL bytes, control chars, etc.).
func SniffFileMIME(path string) (mimeType MIMEType, mode ExtensionMode, err error) {
	if path == "" {
		return MIMEEmpty, ExtensionModeDefault, ErrInvalidPath
	}

	f, err := os.Open(path)
	if err != nil {
		return MIMEEmpty, ExtensionModeDefault, err
	}
	defer f.Close()

	buf := make([]byte, 4096)
	n, err := f.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return MIMEEmpty, ExtensionModeDefault, err
	}
	sample := buf[:n]
	if len(sample) == 0 {
		// Empty file: treat as text/plain.
		return MIMETextPlain, ExtensionModeText, nil
	}

	mStr := http.DetectContentType(sample)
	mimeType = MIMEType(mStr)

	// See if we have a known, non-generic mode.
	if m, ok := MIMETypeToExtensionMode[mimeType]; ok &&
		mimeType != MIMEApplicationOctetStream &&
		mimeType != MIMEEmpty {
		return mimeType, m, nil
	}

	// Fallback: try to decide whether this looks like text.
	if isProbablyTextSample(sample) {
		// For text-ish content where DetectContentType gave us a generic type,
		// prefer text/plain to application/octet-stream.
		if mimeType == MIMEApplicationOctetStream || mimeType == MIMEEmpty {
			mimeType = MIMETextPlain
		}
		return mimeType, ExtensionModeText, nil
	}

	if mimeType == MIMEEmpty {
		mimeType = MIMEApplicationOctetStream
	}
	return mimeType, ExtensionModeDefault, nil
}

// isProbablyTextSample returns true if the byte sample looks like text.
//
// Very simple heuristic: disallow embedded NULs and too many
// non-printable control characters.
func isProbablyTextSample(p []byte) bool {
	if len(p) == 0 {
		return true
	}
	nulCount := 0
	controlCount := 0
	for _, b := range p {
		if b == 0 {
			nulCount++
			continue
		}
		// Allow tab, newline, carriage return.
		if b < 32 && b != 9 && b != 10 && b != 13 {
			controlCount++
		}
	}
	if nulCount > 0 {
		return false
	}
	// If more than ~10% of bytes are weird control chars, assume binary.
	if controlCount*10 > len(p) {
		return false
	}
	return true
}

// normalizeExt lowercases and ensures a leading '.' for an extension.
func normalizeExt(ext string) FileExt {
	ext = strings.TrimSpace(ext)
	if ext == "" {
		return FileExt("")
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return FileExt(strings.ToLower(ext))
}
