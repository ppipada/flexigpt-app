package store

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/tool/spec"
)

// Re-use the simple “name” pattern from the prompt validator.
var nameRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// validateTool performs structural validation of a Tool object.
func validateTool(t *spec.Tool) error {
	if t == nil {
		return errors.New("tool is nil")
	}
	if strings.TrimSpace(t.DisplayName) == "" {
		return errors.New("displayName is empty")
	}
	if t.Slug == "" {
		return errors.New("slug is empty")
	}
	if t.ArgSchema == nil {
		return errors.New("argSchema is missing")
	}
	if t.OutputSchema == nil {
		return errors.New("outputSchema is missing")
	}

	// Type / implementation sanity.
	switch t.Type {
	case spec.ToolTypeGo:
		if t.GoImpl == nil || strings.TrimSpace(t.GoImpl.Func) == "" {
			return errors.New("goImpl.func is required for type 'go'")
		}
		if t.HTTP != nil {
			return errors.New("httpImpl must be unset for type 'go'")
		}
	case spec.ToolTypeHTTP:
		if t.HTTP == nil {
			return errors.New("httpImpl is required for type 'http'")
		}
		if strings.TrimSpace(t.HTTP.Request.URLTemplate) == "" {
			return errors.New("httpImpl.request.urlTemplate is empty")
		}
		if t.GoImpl != nil {
			return errors.New("goImpl must be unset for type 'http'")
		}
	default:
		return fmt.Errorf("invalid type %q", t.Type)
	}

	// Tags.
	seen := map[string]struct{}{}
	for i, tg := range t.Tags {
		if !nameRE.MatchString(tg) {
			return fmt.Errorf("tags[%d]: invalid tag %q", i, tg)
		}
		if _, dup := seen[tg]; dup {
			return fmt.Errorf("duplicate tag %q", tg)
		}
		seen[tg] = struct{}{}
	}
	return nil
}
