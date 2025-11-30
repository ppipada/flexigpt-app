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

func buildBlocksForLocalFile(ctx context.Context, att *Attachment, mode AttachmentMode) (*ContentBlock, error) {
	path := att.FileRef.Path

	switch mode {
	case AttachmentModeText:
		mimeType, extensionMode, err := fileutil.MIMEForLocalFile(path)
		if err == nil && (extensionMode == fileutil.ExtensionModeText || mimeType == fileutil.MIMEApplicationPDF) {
			// Text mode mimes and pdf with text extraction is supported.
			return getTextBlock(att)
		}
		// Could not detect mime or non pdf text attachment sent, render as unreadable file.
		return getUnreadableBlock(att)

	case AttachmentModeImage:
		// Treat a file in "image" mode as an image attachment.
		return buildImageBlockFromLocal(path)

	case AttachmentModeFile:
		// File mode needs to handle all kinds of supported files appropriately.
		mimeType, extensionMode, err := fileutil.MIMEForLocalFile(path)
		if err != nil {
			// Could not detect mime, render as unreadable file.
			return getUnreadableBlock(att)
		}
		switch extensionMode {
		case fileutil.ExtensionModeText:
			// If this is a text type file, we cannot attach it as b64 encoded currently.
			// Right now we are making a safe fallback to send it as text block.
			// Ideally we should not reach here if UI takes care of AttachmentKind and AttachmentMode properly.
			return getTextBlock(att)
		case fileutil.ExtensionModeImage:
			// Images have to be attached specially.
			// Ideally we should not reach here if UI takes care of AttachmentKind and AttachmentMode properly.
			return buildImageBlockFromLocal(path)
		case fileutil.ExtensionModeDocument:
			// As of now non pdf files are not supported to be attached as b64 files in any APIs.
			// But, as the call is reached here, we are still sending the content back as b64 content.
			// As APIs start supporting other file types, and UI and fetch api handles things this will be seameless.
			base64Data, err := fileutil.ReadFile(path, fileutil.ReadEncodingBinary)
			if err != nil {
				return nil, err
			}

			return &ContentBlock{
				Kind:       ContentBlockFile,
				Base64Data: base64Data,
				MIMEType:   string(mimeType),
				FileName:   filepath.Base(path),
			}, nil

		case fileutil.ExtensionModeDefault:
			return getUnreadableBlock(att)
		default:
			return getUnreadableBlock(att)
		}

	case AttachmentModeNotReadable,
		AttachmentModePageContent,
		AttachmentModeLinkOnly,
		AttachmentModePRDiff,
		AttachmentModePRPage,
		AttachmentModeCommitDiff,
		AttachmentModeCommitPage:
		// Provide a short note so the model knows the file exists but is not included.
		return getUnreadableBlock(att)
	default:
		return nil, errors.New("invalid attachment mode")
	}
}

func getTextBlock(att *Attachment) (*ContentBlock, error) {
	path := att.FileRef.Path
	ext := strings.ToLower(filepath.Ext(path))
	// Special handling for PDFs: try text extraction with panic-safe fallback.
	if ext == string(fileutil.ExtPDF) {
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
}

func getUnreadableBlock(att *Attachment) (*ContentBlock, error) {
	if txt := att.FormatAsDisplayName(); txt != "" {
		return &ContentBlock{
			Kind: ContentBlockText,
			Text: txt + " (binary file; not readable in this chat)",
		}, nil
	}
	return nil, errors.New("invalid attachment mode")
}
