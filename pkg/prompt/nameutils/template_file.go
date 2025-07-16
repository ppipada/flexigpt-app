package nameutils

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

// FileInfo holds information about a template file.
type FileInfo struct {
	Slug     spec.TemplateSlug
	Version  spec.TemplateVersion
	FileName string
}

// BuildTemplateFileInfo returns the canonical filename for a (slug, version) pair.
// Both slug and version must be validated before calling.
// The filename is "<url-escaped-slug>_<url-escaped-version>.json".
func BuildTemplateFileInfo(slug spec.TemplateSlug, version spec.TemplateVersion) (FileInfo, error) {
	if err := ValidateTemplateSlug(slug); err != nil {
		return FileInfo{}, err
	}
	if err := ValidateTemplateVersion(version); err != nil {
		return FileInfo{}, err
	}
	fname := fmt.Sprintf("%s_%s.%s",
		url.PathEscape(string(slug)),
		url.PathEscape(string(version)),
		spec.PromptTemplateFileExtension,
	)
	return FileInfo{Slug: slug, Version: version, FileName: fname}, nil
}

// ParseTemplateFileName parses a template filename into slug and version.
// Uses the last underscore as delimiter.
func ParseTemplateFileName(fn string) (FileInfo, error) {
	fn = filepath.Base(fn)
	ext := "." + spec.PromptTemplateFileExtension
	if !strings.HasSuffix(fn, ext) {
		return FileInfo{}, fmt.Errorf("not a %q file: %q", ext, fn)
	}
	trim := strings.TrimSuffix(fn, ext)
	idx := strings.LastIndex(trim, "_")
	if idx < 1 || idx == len(trim)-1 {
		return FileInfo{}, fmt.Errorf("invalid template filename: %q", fn)
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
	if err := ValidateTemplateSlug(spec.TemplateSlug(slug)); err != nil {
		return FileInfo{}, err
	}
	if err := ValidateTemplateVersion(spec.TemplateVersion(ver)); err != nil {
		return FileInfo{}, err
	}
	return FileInfo{
		Slug:     spec.TemplateSlug(slug),
		Version:  spec.TemplateVersion(ver),
		FileName: fn,
	}, nil
}
