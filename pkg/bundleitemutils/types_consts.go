package bundleitemutils

import "errors"

const ItemFileExtension = "json"

type (
	BundleID    string
	BundleSlug  string
	ItemID      string
	ItemSlug    string
	ItemVersion string
)

var (
	ErrInvalidSlug     = errors.New("invalid slug")
	ErrInvalidVersion  = errors.New("invalid version")
	ErrInvalidFilename = errors.New("invalid filename")

	ErrBundleAttributeMissing = errors.New("missing bundle partition attribute")
)
