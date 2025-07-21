package bundleitemutils

import (
	"fmt"
	"unicode"
)

// maxTokenLength is the maximum allowed length for slugs and versions.
const maxTokenLength = 64

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

// validateSlug validates a slug.
func validateSlug(slug string) error {
	return validateToken(slug, false, ErrInvalidSlug)
}

// validateVersion validates a version string.
func validateVersion(v string) error {
	return validateToken(v, true, ErrInvalidVersion)
}

func ValidateBundleSlug(s BundleSlug) error   { return validateSlug(string(s)) }
func ValidateItemSlug(s ItemSlug) error       { return validateSlug(string(s)) }
func ValidateItemVersion(v ItemVersion) error { return validateVersion(string(v)) }
