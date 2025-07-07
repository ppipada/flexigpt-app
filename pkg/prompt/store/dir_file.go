// Package store provides helpers for directory and filename layout for prompt templates.
// It ensures all disk layout and naming conventions are enforced in one place.
package store

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
)

// ErrInvalidSlug is returned when a slug is invalid.
var ErrInvalidSlug = errors.New("invalid slug")

// ErrInvalidVersion is returned when a version string is invalid.
var ErrInvalidVersion = errors.New("invalid version")

const (
	promptTemplateFileExtension = "json"
	sqliteDBFileName            = "prompttemplates.fts.sqlite"
	bundlesMetaFileName         = "prompttemplates.bundles.json"
	maxTokenLength              = 64 // maxTokenLength is the maximum allowed length for slugs and versions.
)

// validateToken checks that a string contains only allowed runes and is not too long.
// Allowed: Unicode Letter, Unicode Digit, ASCII dash '-'. No dot, underscore, space, slash, etc.
// Returns the provided error if invalid.
func validateToken(tok string, errToReturn error) error {
	if tok == "" {
		return errToReturn
	}
	runeCount := 0
	for _, r := range tok {
		runeCount++
		if r == '-' {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return errToReturn
	}
	if runeCount > maxTokenLength {
		return errToReturn
	}
	return nil
}

// ValidateSlug validates a bundle or template slug.
func ValidateSlug(slug string) error {
	return validateToken(slug, ErrInvalidSlug)
}

// ValidateVersion validates a template version string.
func ValidateVersion(v string) error {
	return validateToken(v, ErrInvalidVersion)
}

// bundleDirInfo holds information about a bundle directory.
type bundleDirInfo struct {
	ID, Slug, DirName string
}

// sanitizeID removes all characters from id except [a-zA-Z0-9-_].
// If the result is empty, returns "x".
func sanitizeID(id string) string {
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= '0' && r <= '9',
			r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r == '-', r == '_':
			_, _ = b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "x"
	}
	return b.String()
}

// buildBundleDir returns a bundleDirInfo with a directory name safe for all filesystems.
// The directory name is "<sanitizedID>_<slug>".
func buildBundleDir(id, slug string) (bundleDirInfo, error) {
	if err := ValidateSlug(slug); err != nil {
		return bundleDirInfo{}, fmt.Errorf("buildBundleDir: %w", err)
	}
	dir := fmt.Sprintf("%s_%s", sanitizeID(id), slug)
	return bundleDirInfo{ID: id, Slug: slug, DirName: dir}, nil
}

// parseBundleDir parses a bundle directory name into its ID and slug parts.
// Splits on the first underscore.
func parseBundleDir(dir string) (bundleDirInfo, error) {
	parts := strings.SplitN(dir, "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return bundleDirInfo{}, fmt.Errorf("invalid bundle dir: %q", dir)
	}
	return bundleDirInfo{ID: parts[0], Slug: parts[1], DirName: dir}, nil
}

// fileInfo holds information about a template file.
type fileInfo struct {
	Slug, Version, FileName string
}

// templateFileName returns the canonical filename for a (slug, version) pair.
// Both slug and version must be validated before calling.
// The filename is "<url-escaped-slug>_<url-escaped-version>.json".
func templateFileName(slug, version string) (string, error) {
	if err := ValidateSlug(slug); err != nil {
		return "", err
	}
	if err := ValidateVersion(version); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s.%s",
		url.PathEscape(slug),
		url.PathEscape(version),
		promptTemplateFileExtension,
	), nil
}

// parseTemplateFileName parses a template filename into slug and version.
// Uses the last underscore as delimiter.
func parseTemplateFileName(fn string) (fileInfo, error) {
	fn = filepath.Base(fn)
	ext := "." + promptTemplateFileExtension
	if !strings.HasSuffix(fn, ext) {
		return fileInfo{}, fmt.Errorf("not a %q file: %q", ext, fn)
	}
	trim := strings.TrimSuffix(fn, ext)
	idx := strings.LastIndex(trim, "_")
	if idx < 1 || idx == len(trim)-1 {
		return fileInfo{}, fmt.Errorf("invalid template filename: %q", fn)
	}
	slugEsc, verEsc := trim[:idx], trim[idx+1:]
	slug, err := url.PathUnescape(slugEsc)
	if err != nil {
		return fileInfo{}, fmt.Errorf("invalid slug escape in %q: %w", fn, err)
	}
	ver, err := url.PathUnescape(verEsc)
	if err != nil {
		return fileInfo{}, fmt.Errorf("invalid version escape in %q: %w", fn, err)
	}
	if err := ValidateSlug(slug); err != nil {
		return fileInfo{}, err
	}
	if err := ValidateVersion(ver); err != nil {
		return fileInfo{}, err
	}
	return fileInfo{Slug: slug, Version: ver, FileName: fn}, nil
}

// bundlePartitionAttr is a type for bundle partition attributes.
type bundlePartitionAttr string

// BundlePartitionProvider implements dirstore.PartitionProvider for bundle directories.
type BundlePartitionProvider struct{}

// GetPartitionDir returns the partition directory for a given file key.
// The directory is provided via the XAttr field.
func (b *BundlePartitionProvider) GetPartitionDir(key dirstore.FileKey) (string, error) {
	if key.FileName == "" {
		return "", errors.New("empty filename")
	}
	dir, ok := key.XAttr.(bundlePartitionAttr)
	if !ok || dir == "" {
		return "", errors.New("missing bundle partition attribute")
	}
	return string(dir), nil
}

// ListPartitions lists all bundle directories under baseDir.
func (b *BundlePartitionProvider) ListPartitions(
	baseDir, sortOrder, pageToken string,
	pageSize int,
) (dirs []string, nextPageToken string, err error) {
	return dirstore.ListDirs(baseDir, sortOrder, pageToken, pageSize)
}
