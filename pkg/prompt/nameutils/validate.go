package nameutils

import (
	"errors"
	"fmt"
	"unicode"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

// ErrInvalidSlug is returned when a slug is invalid.
var ErrInvalidSlug = errors.New("invalid slug")

// ErrInvalidVersion is returned when a version string is invalid.
var ErrInvalidVersion = errors.New("invalid version")

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

// validateSlug validates a bundle or template slug.
func validateSlug(slug string) error {
	return validateToken(slug, false, ErrInvalidSlug)
}

// validateVersion validates a template version string.
func validateVersion(v string) error {
	return validateToken(v, true, ErrInvalidVersion)
}

func ValidateBundleSlug(s spec.BundleSlug) error           { return validateSlug(string(s)) }
func ValidateTemplateSlug(s spec.TemplateSlug) error       { return validateSlug(string(s)) }
func ValidateTemplateVersion(v spec.TemplateVersion) error { return validateVersion(string(v)) }
