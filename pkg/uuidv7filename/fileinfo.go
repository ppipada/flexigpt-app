package uuidv7filename

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var nonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9-]`)

// UUIDv7FileInfo provides UUIDv7 based filenames "<uuid>_<sanitised-64-char-suffix>.<ext>".
type UUIDv7FileInfo struct {
	ID     string
	Suffix string
	// Without leading dot.
	Extension string
	// Full filename with extension.
	FileName string
	Time     time.Time
}

// Build constructs a filename of the form "<uuid>_<sanitized-suffix>.<extension>".
// Note: The Suffix is lossy- non-alphanumeric characters are replaced with underscores and the suffix is truncated to 64 characters.
// The original suffix cannot be fully recovered from the filename.
func Build(id, suffix, extension string) (UUIDv7FileInfo, error) {
	if id == "" || suffix == "" || extension == "" {
		return UUIDv7FileInfo{}, fmt.Errorf(
			"invalid request. id: %s, suffix: %s extension:%s",
			id,
			suffix,
			extension,
		)
	}
	extension = cleanExt(extension)
	u, err := ExtractUUIDv7(id)
	if err != nil {
		return UUIDv7FileInfo{}, fmt.Errorf("invalid ID: %s err: %w", id, err)
	}

	t, err := extractTimeFromUUIDv7(u)
	if err != nil {
		return UUIDv7FileInfo{}, fmt.Errorf("invalid ID: %s err: %w", id, err)
	}

	if len(suffix) > 64 {
		suffix = suffix[:64]
	}
	suffix = nonAlphaNum.ReplaceAllString(suffix, "_")
	name := fmt.Sprintf("%s_%s.%s", id, suffix, extension)
	return UUIDv7FileInfo{
		ID:        id,
		Suffix:    suffix,
		Extension: extension,
		FileName:  name,
		Time:      t,
	}, nil
}

// Parse extracts the UUID, suffix, and extension from a filename produced by Build.
// Note: The Suffix is only an approximation of the original input as build is lossy.
// Underscores in the filename are converted to spaces, and any original non-alphanumeric characters or underscores cannot be exactly recovered.
func Parse(filename string) (UUIDv7FileInfo, error) {
	base := filepath.Base(filename)
	extension := filepath.Ext(base)
	base = strings.TrimSuffix(base, extension)
	extension = cleanExt(extension)

	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return UUIDv7FileInfo{}, fmt.Errorf("invalid file name: %s", filename)
	}
	id := parts[0]
	suffix := strings.ReplaceAll(parts[1], "_", " ")
	u, err := ExtractUUIDv7(id)
	if err != nil {
		return UUIDv7FileInfo{}, fmt.Errorf("invalid ID: %s err: %w", id, err)
	}
	t, err := extractTimeFromUUIDv7(u)
	if err != nil {
		return UUIDv7FileInfo{}, err
	}

	return UUIDv7FileInfo{
		ID:        id,
		Suffix:    suffix,
		Extension: extension,
		FileName:  filename,
		Time:      t,
	}, nil
}

// ExtractUUIDv7 parses and validates a UUIDv7 string.
func ExtractUUIDv7(s string) (uuid.UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return u, fmt.Errorf("invalid UUID: %w", err)
	}
	if u.Variant() != uuid.RFC4122 {
		return u, fmt.Errorf("UUID %q is not RFC-4122 variant", s)
	}
	if u.Version() != 7 {
		return u, fmt.Errorf("UUID %q is version %d, want 7", s, u.Version())
	}
	return u, nil
}

func NewUUID() (string, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// extractTimeFromUUIDv7 extracts the time from a UUIDv7 object.
func extractTimeFromUUIDv7(u uuid.UUID) (time.Time, error) {
	sec, nsec := u.Time().UnixTime()
	return time.Unix(sec, nsec).UTC(), nil
}

// cleanExt removes a leading dot from the extension, if present.
func cleanExt(ext string) string {
	if strings.HasPrefix(ext, ".") {
		return ext[1:]
	}
	return ext
}
