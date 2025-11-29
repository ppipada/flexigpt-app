package attachment

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/fileutil"
)

// FileRef carries metadata for file attachments.
type FileRef struct {
	fileutil.PathInfo
}

func (ref *FileRef) PopulateRef() error {
	if ref == nil {
		return errors.New("file attachment missing ref")
	}
	path := strings.TrimSpace(ref.Path)
	if path == "" {
		return errors.New("file attachment missing path")
	}
	pathInfo, err := fileutil.StatPath(path)
	if err != nil {
		return err
	}

	ref.Path = pathInfo.Path
	ref.Name = pathInfo.Name
	ref.Exists = pathInfo.Exists
	ref.IsDir = pathInfo.IsDir
	ref.Size = pathInfo.Size
	ref.ModTime = pathInfo.ModTime

	return nil
}

func inferDefaultFileMode(path string) AttachmentMode {
	ext := strings.ToLower(filepath.Ext(path))
	m, ok := fileutil.ExtensionToMode[ext]
	if !ok {
		// Could be binary; leave as "file" by default.
		return AttachmentModeFile
	}

	switch m {
	case fileutil.ExtensionModeText:
		return AttachmentModeText
	case fileutil.ExtensionModeDocument:
		// PDFs get a special default: send as a file attachment.
		if ext == fileutil.ExtensionPDF {
			return AttachmentModeFile
		}
		// Non‑PDF “documents” (e.g. DOCX) – default to text.
		return AttachmentModeText
	case fileutil.ExtensionModeImage, fileutil.ExtensionModeDefault:
		// In File attachments, images default to binary file mode
		// unless the user explicitly selects "image".
		return AttachmentModeFile
	default:
		return AttachmentModeFile
	}
}

func buildBlocksForLocalFile(ctx context.Context, att *Attachment, mode AttachmentMode) (*ContentBlock, error) {
	path := att.FileRef.Path
	ext := strings.ToLower(filepath.Ext(path))

	switch mode {
	case AttachmentModeText:
		// Special handling for PDFs: try text extraction with panic-safe fallback.
		if ext == fileutil.ExtensionPDF {
			return buildPDFTextOrFileBlock(path)
		}
		// Normal text file.
		text, err := fileutil.ReadFile(path, fileutil.ReadEncodingText)
		if err != nil {
			return nil, err
		}
		return &ContentBlock{
			Kind: ContentBlockText,
			Text: text,
		}, nil

	case AttachmentModeNotReadable,
		AttachmentModePageContent,
		AttachmentModeLinkOnly,
		AttachmentModePRDiff,
		AttachmentModePRPage,
		AttachmentModeCommitDiff,
		AttachmentModeCommitPage:
		// Provide a short note so the model knows the file exists but is not included.
		if txt := att.FormatAsDisplayName(); txt != "" {
			return &ContentBlock{
				Kind: ContentBlockText,
				Text: txt + " (binary file; not readable in this chat)",
			}, nil
		}
		return nil, errors.New("invalid attachment mode")

	case AttachmentModeImage:
		// Treat a file in "image" mode as an image attachment.
		return buildImageBlockFromLocal(path)

	case AttachmentModeFile:
		base64Data, err := fileutil.ReadFile(path, fileutil.ReadEncodingBinary)
		if err != nil {
			return nil, err
		}
		mime := mimeForLocalFile(path)
		return &ContentBlock{
			Kind:       ContentBlockFile,
			Base64Data: base64Data,
			MIMEType:   mime,
			FileName:   filepath.Base(path),
		}, nil

	default:
		return nil, errors.New("invalid attachment mode")
	}
}

func mimeForLocalFile(path string) string {
	// Prefer MIME sniffing via fileutil; fall back to extension-based if needed.
	if mime, _, err := fileutil.DetectFileMIME(path); err == nil && mime != "" {
		return mime
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
