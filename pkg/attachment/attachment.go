package attachment

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Kind enumerates contextual attachment categories that can be
// associated with messages sent to the inference layer.
type Kind string

const (
	AttachmentFile     Kind = "file"
	AttachmentImage    Kind = "image"
	AttachmentDocIndex Kind = "docIndex"
	AttachmentPR       Kind = "pr"
	AttachmentCommit   Kind = "commit"
	AttachmentSnapshot Kind = "snapshot"
)

// Attachment is a lightweight reference to external context (files, docs, images, etc.).
type Attachment struct {
	Kind  Kind   `json:"kind"`
	Label string `json:"label"`

	// Exactly one field below must be non-nil.
	FileRef    *FileRef   `json:"fileRef,omitempty"`
	ImageRef   *ImageRef  `json:"imageRef,omitempty"`
	GenericRef *HandleRef `json:"genericRef,omitempty"`
}

func (att *Attachment) PopulateRef() error {
	switch att.Kind {
	case AttachmentFile:
		if att.FileRef == nil {
			return errors.New("no file ref for file attachment")
		}
		err := att.FileRef.PopulateRef()
		if err != nil {
			return err
		}

		if att.Label == "" {
			att.Label = filepath.Base(att.FileRef.Path)
		}
		return nil
	case AttachmentImage:
		if att.ImageRef == nil {
			return errors.New("no image ref for image attachment")
		}
		err := att.ImageRef.PopulateRef()
		if err != nil {
			return err
		}

		if att.Label == "" {
			att.Label = filepath.Base(att.ImageRef.Path)
		}
		return nil
	case AttachmentCommit,
		AttachmentDocIndex,
		AttachmentPR,
		AttachmentSnapshot:
		if att.GenericRef == nil {
			return errors.New("no generic ref for attachment")
		}

		err := att.GenericRef.PopulateRef()
		if err != nil {
			return err
		}

		if att.Label == "" {
			att.Label = att.GenericRef.Handle
		}
		return nil

	default:
		return errors.New("unknown attachment kind")
	}
}

// FormatAsDisplayName normalizes an attachment into a short, human-readable
// string that can be injected into text prompts when a richer modality is not
// available (for example, for generic refs or providers without native file support).
func (att *Attachment) FormatAsDisplayName() string {
	label := strings.TrimSpace(att.Label)
	var detail string

	switch att.Kind {
	case AttachmentFile:
		if att.FileRef != nil {
			detail = strings.TrimSpace(att.FileRef.Path)
		}
	case AttachmentImage:
		if att.ImageRef != nil {
			detail = strings.TrimSpace(att.ImageRef.Path)
		}
	case AttachmentDocIndex,
		AttachmentPR,
		AttachmentCommit,
		AttachmentSnapshot:
		if att.GenericRef != nil {
			detail = strings.TrimSpace(att.GenericRef.Handle)
		}
	default:
		detail = ""
	}

	if label == "" {
		label = detail
	}

	if label == "" && detail == "" {
		return ""
	}
	if detail == "" || label == detail {
		return fmt.Sprintf("[Attachment: %s]", label)
	}
	return fmt.Sprintf("[Attachment: %s — %s]", label, detail)
}
