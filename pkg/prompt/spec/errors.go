package spec

import "errors"

var (
	ErrInvalidRequest  = errors.New("invalid request")
	ErrInvalidSlug     = errors.New("invalid slug")
	ErrInvalidVersion  = errors.New("invalid version")
	ErrInvalidFilename = errors.New("invalid filename")
	ErrInvalidDir      = errors.New("invalid directory")
	ErrConflict        = errors.New("resource already exists")

	ErrBuiltInBundleNotFound = errors.New("bundle not found in built-in data")
	ErrBundleNotFound        = errors.New("bundle not found")
	ErrBundleDisabled        = errors.New("bundle is disabled")
	ErrBundleDeleting        = errors.New("bundle is being deleted")
	ErrBundleNotEmpty        = errors.New("bundle still contains templates")

	ErrTemplateNotFound        = errors.New("template not found")
	ErrBuiltInTemplateNotFound = errors.New("template not found in built-in data")

	ErrBuiltInReadOnly = errors.New("built-in resource is read-only")

	ErrFTSDisabled            = errors.New("FTS is disabled")
	ErrBundleAttributeMissing = errors.New("missing bundle partition attribute")
)
