package nameutils

import (
	"fmt"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

// BundleDirInfo holds information about a bundle directory.
type BundleDirInfo struct {
	ID      spec.BundleID
	Slug    spec.BundleSlug
	DirName string
}

// SanitizeBundleID removes all characters from id except [a-zA-Z0-9-_].
// If the result is empty, returns "x".
func SanitizeBundleID(id spec.BundleID) string {
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

// BuildBundleDir returns a BundleDirInfo with a directory name safe for all filesystems.
// The directory name is "<sanitizedID>_<slug>".
func BuildBundleDir(id spec.BundleID, slug spec.BundleSlug) (BundleDirInfo, error) {
	if err := ValidateBundleSlug(slug); err != nil {
		return BundleDirInfo{}, fmt.Errorf("buildBundleDir: %w", err)
	}
	dir := fmt.Sprintf("%s_%s", SanitizeBundleID(id), slug)
	return BundleDirInfo{ID: id, Slug: slug, DirName: dir}, nil
}

// ParseBundleDir parses a bundle directory name into its ID and slug parts.
// Splits on the first underscore.
func ParseBundleDir(dir string) (BundleDirInfo, error) {
	parts := strings.SplitN(dir, "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return BundleDirInfo{}, fmt.Errorf("invalid bundle dir: %q", dir)
	}
	return BundleDirInfo{
		ID:      spec.BundleID(parts[0]),
		Slug:    spec.BundleSlug(parts[1]),
		DirName: dir,
	}, nil
}
