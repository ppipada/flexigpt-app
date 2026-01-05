package attachment

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/internal/fileutil"
)

// FileRef carries metadata for file attachments.
type FileRef struct {
	fileutil.PathInfo

	// Snapshot of the original file state when it was first attached.
	// This lets us detect if the underlying file changed between turns.
	OrigPath    string    `json:"origPath"`
	OrigSize    int64     `json:"origSize"`
	OrigModTime time.Time `json:"origModTime"`
}

func (ref *FileRef) PopulateRef(replaceOrig bool) error {
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

	// Capture original snapshot once.
	if strings.TrimSpace(ref.OrigPath) == "" || replaceOrig {
		ref.OrigPath = pathInfo.Path
		ref.OrigSize = pathInfo.Size
		ref.OrigModTime = *pathInfo.ModTime
	}

	ref.Path = pathInfo.Path
	ref.Name = pathInfo.Name
	ref.Exists = pathInfo.Exists
	ref.IsDir = pathInfo.IsDir
	ref.Size = pathInfo.Size
	ref.ModTime = pathInfo.ModTime

	return nil
}

// IsModified reports whether the current file state differs from the original
// snapshot captured when the attachment was first used.
func (ref *FileRef) IsModified() bool {
	if ref == nil {
		return false
	}
	if strings.TrimSpace(ref.OrigPath) == "" {
		// No baseline; treat as not modified.
		return false
	}
	// If the file no longer exists or any key metadata changed, mark as modified.
	if !ref.Exists {
		return true
	}
	if ref.Path != ref.OrigPath {
		return true
	}
	if ref.Size != ref.OrigSize {
		return true
	}
	if !ref.ModTime.Equal(ref.OrigModTime) {
		return true
	}
	return false
}

func (ref *FileRef) BuildContentBlock(
	attachmentContentBlockMode AttachmentContentBlockMode,
	onlyIfTextKind bool,
) (*ContentBlock, error) {
	path := strings.TrimSpace(ref.Path)
	if path == "" {
		return nil, errors.New("got invalid path")
	}
	switch attachmentContentBlockMode {
	case AttachmentContentBlockModeImage:
		if onlyIfTextKind {
			return nil, ErrNonTextContentBlock
		}
		// Treat a file in "image" mode as an image attachment.
		return buildImageBlockFromLocal(path)

	case AttachmentContentBlockModeFile:
		// File mode needs to handle all kinds of supported files appropriately.
		mimeType, extensionMode, err := fileutil.MIMEForLocalFile(path)
		if err != nil {
			// Could not detect mime, render as unreadable file.
			return nil, errors.Join(ErrUnreadableFile, err)
		}
		switch extensionMode {
		case fileutil.ExtensionModeText:
			// If this is a text type file, we cannot attach it as b64 encoded currently.
			// Right now we are making a safe fallback to send it as text block.
			// Ideally we should not reach here if UI takes care of AttachmentKind and AttachmentContentBlockMode
			// properly.
			return ref.getTextBlock()
		case fileutil.ExtensionModeImage:
			if onlyIfTextKind {
				return nil, ErrNonTextContentBlock
			}
			// Images have to be attached specially.
			// Ideally we should not reach here if UI takes care of AttachmentKind and AttachmentContentBlockMode
			// properly.
			return buildImageBlockFromLocal(path)
		case fileutil.ExtensionModeDocument:
			if onlyIfTextKind {
				return nil, ErrNonTextContentBlock
			}
			// As of now non pdf files are not supported to be attached as b64 files in any APIs.
			// But, as the call is reached here, we are still sending the content back as b64 content.
			// As APIs start supporting other file types, and UI and fetch api handles things this will be seameless.
			base64Data, err := fileutil.ReadFile(path, fileutil.ReadEncodingBinary)
			if err != nil {
				return nil, err
			}

			mStr := string(mimeType)
			fname := filepath.Base(path)
			return &ContentBlock{
				Kind:       ContentBlockFile,
				Base64Data: &base64Data,
				MIMEType:   &mStr,
				FileName:   &fname,
			}, nil

		case fileutil.ExtensionModeDefault:
			return nil, ErrUnreadableFile
		default:
			return nil, ErrUnreadableFile
		}

	case AttachmentContentBlockModeText:
		mimeType, extensionMode, err := fileutil.MIMEForLocalFile(path)
		if err == nil && (extensionMode == fileutil.ExtensionModeText || mimeType == fileutil.MIMEApplicationPDF) {
			// Text mode mimes and pdf with text extraction is supported.
			return ref.getTextBlock()
		}
		// Could not detect mime or non pdf text attachment sent, render as unreadable file.
		return nil, ErrUnreadableFile

	case AttachmentContentBlockModeNotReadable,
		AttachmentContentBlockModePageContent,
		AttachmentContentBlockModeLinkOnly,
		AttachmentContentBlockModePRDiff,
		AttachmentContentBlockModePRPage,
		AttachmentContentBlockModeCommitDiff,
		AttachmentContentBlockModeCommitPage:
		// Provide a short note so the model knows the file exists but is not included.
		return nil, ErrUnreadableFile
	default:
		return nil, errors.New("invalid attachment mode")
	}
}

func (ref *FileRef) getTextBlock() (*ContentBlock, error) {
	path := strings.TrimSpace(ref.Path)
	if path == "" {
		return nil, errors.New("got invalid path")
	}
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
		Text: &text,
	}, nil
}
