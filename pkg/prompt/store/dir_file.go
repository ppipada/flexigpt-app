package store

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/dirstore"
)

const (
	promptTemplateFileExtension = "json"
	sqliteDBFileName            = "prompttemplates.fts.sqlite"
	bundlesMetaFileName         = "prompttemplates.bundles.json"
)

// bundleDirInfo holds information about a bundle directory.
type bundleDirInfo struct {
	ID, Slug, DirName string
}

// sanitizeID removes all characters from id except [a-zA-Z0-9-_].
// If the result is empty, returns "x".
func sanitizeID(id spec.BundleID) string {
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
func buildBundleDir(id spec.BundleID, slug spec.BundleSlug) (bundleDirInfo, error) {
	if err := slug.Validate(); err != nil {
		return bundleDirInfo{}, fmt.Errorf("buildBundleDir: %w", err)
	}
	dir := fmt.Sprintf("%s_%s", sanitizeID(id), slug)
	return bundleDirInfo{ID: string(id), Slug: string(slug), DirName: dir}, nil
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
func templateFileName(slug spec.TemplateSlug, version spec.TemplateVersion) (string, error) {
	if err := slug.Validate(); err != nil {
		return "", err
	}
	if err := version.Validate(); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s.%s",
		url.PathEscape(string(slug)),
		url.PathEscape(string(version)),
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
	if err := spec.TemplateSlug(slug).Validate(); err != nil {
		return fileInfo{}, err
	}
	if err := spec.TemplateVersion(ver).Validate(); err != nil {
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
