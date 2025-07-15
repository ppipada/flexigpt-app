package store

import "errors"

var (
	ErrBuiltInBundleNotFound = errors.New("bundle not found in built-in data")
	ErrBundleNotFound        = errors.New("bundle not found")
	ErrBundleDisabled        = errors.New("bundle is disabled")
	ErrBundleDeleting        = errors.New("bundle is being deleted")
	ErrBundleNotEmpty        = errors.New("bundle still contains templates")

	ErrTemplateNotFound        = errors.New("template not found")
	ErrBuiltInTemplateNotFound = errors.New("template not found in built-in data")

	ErrConflict       = errors.New("resource already exists")
	ErrInvalidRequest = errors.New("invalid request")
	ErrFTSDisabled    = errors.New("FTS is disabled")

	ErrBuiltInReadOnly = errors.New("built-in resource is read-only")
)
