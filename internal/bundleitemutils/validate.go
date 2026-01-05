package bundleitemutils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// maxTokenLength is the maximum allowed length for slugs and versions.
const maxTokenLength = 64

var tagNameRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// ValidateTags checks a slice of tags for format, length, and duplicates.
func ValidateTags(tags []string) error {
	seen := map[string]struct{}{}
	for i, tag := range tags {
		if err := ValidateTag(tag); err != nil {
			return fmt.Errorf("tags[%d]: %w", i, err)
		}
		if _, dup := seen[tag]; dup {
			return fmt.Errorf("duplicate tag %q", tag)
		}
		seen[tag] = struct{}{}
	}
	return nil
}

// ValidateTag checks a single tag for format and length.
func ValidateTag(tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return errors.New("tag is empty")
	}
	if len(tag) > maxTokenLength {
		return fmt.Errorf("tag %q is too long (max %d)", tag, maxTokenLength)
	}
	if !tagNameRE.MatchString(tag) {
		return fmt.Errorf("invalid tag %q", tag)
	}
	return nil
}

func ValidateBundleSlug(s BundleSlug) error   { return validateSlug(string(s)) }
func ValidateItemSlug(s ItemSlug) error       { return validateSlug(string(s)) }
func ValidateItemVersion(v ItemVersion) error { return validateVersion(string(v)) }

// validateSlug validates a slug.
func validateSlug(slug string) error {
	return validateToken(slug, false, ErrInvalidSlug)
}

// validateVersion validates a version string.
func validateVersion(v string) error {
	return validateToken(v, true, ErrInvalidVersion)
}

// validateToken checks that a string contains only allowed runes and is not too long.
// Allowed: Unicode Letter, Unicode Digit, ASCII dash '-'. No dot, underscore, space, slash, etc.
// Returns the provided error if invalid.
func validateToken(tok string, allowDot bool, errToReturn error) error {
	if tok == "" {
		return errToReturn
	}
	runeCount := 0
	for _, r := range tok {
		runeCount++
		if r == '-' {
			continue
		}
		if r == '.' && allowDot {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return fmt.Errorf("input: %s, err: %w", tok, errToReturn)
	}
	if runeCount > maxTokenLength {
		return errToReturn
	}
	return nil
}
