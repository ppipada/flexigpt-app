package bundleitemutils

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// FileInfo holds information about a item file.
type FileInfo struct {
	Slug     ItemSlug
	Version  ItemVersion
	FileName string
}

// BuildItemFileInfo returns the canonical filename for a (slug, version) pair.
// Both slug and version must be validated before calling.
// The filename is "<url-escaped-slug>_<url-escaped-version>.json".
func BuildItemFileInfo(slug ItemSlug, version ItemVersion) (FileInfo, error) {
	if err := ValidateItemSlug(slug); err != nil {
		return FileInfo{}, err
	}
	if err := ValidateItemVersion(version); err != nil {
		return FileInfo{}, err
	}
	fname := fmt.Sprintf("%s_%s.%s",
		url.PathEscape(string(slug)),
		url.PathEscape(string(version)),
		ItemFileExtension,
	)
	return FileInfo{Slug: slug, Version: version, FileName: fname}, nil
}

// ParseItemFileName parses a item filename into slug and version.
// Uses the last underscore as delimiter.
func ParseItemFileName(fn string) (FileInfo, error) {
	fn = filepath.Base(fn)
	ext := "." + ItemFileExtension
	if !strings.HasSuffix(fn, ext) {
		return FileInfo{}, fmt.Errorf("not a %q file: %q", ext, fn)
	}
	trim := strings.TrimSuffix(fn, ext)
	idx := strings.LastIndex(trim, "_")
	if idx < 1 || idx == len(trim)-1 {
		return FileInfo{}, fmt.Errorf("invalid item filename: %q", fn)
	}
	slugEsc, verEsc := trim[:idx], trim[idx+1:]
	slug, err := url.PathUnescape(slugEsc)
	if err != nil || slug == "" {
		return FileInfo{}, fmt.Errorf("invalid slug escape in %q: %w", fn, err)
	}
	ver, err := url.PathUnescape(verEsc)
	if err != nil || ver == "" {
		return FileInfo{}, fmt.Errorf("invalid version escape in %q: %w", fn, err)
	}
	if err := ValidateItemSlug(ItemSlug(slug)); err != nil {
		return FileInfo{}, err
	}
	if err := ValidateItemVersion(ItemVersion(ver)); err != nil {
		return FileInfo{}, err
	}
	return FileInfo{
		Slug:     ItemSlug(slug),
		Version:  ItemVersion(ver),
		FileName: fn,
	}, nil
}
