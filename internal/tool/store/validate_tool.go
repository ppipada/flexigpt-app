package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ppipada/flexigpt-app/internal/bundleitemutils"
	"github.com/ppipada/flexigpt-app/internal/tool/spec"
)

// validateTool performs structural validation of a Tool object.
func validateTool(t *spec.Tool) error {
	if t == nil {
		return errors.New("tool is nil")
	}
	if t.SchemaVersion != spec.SchemaVersion {
		return fmt.Errorf(
			"schemaVersion %q does not match expected %q",
			t.SchemaVersion,
			spec.SchemaVersion,
		)
	}
	if err := bundleitemutils.ValidateItemSlug(t.Slug); err != nil {
		return fmt.Errorf("invalid slug %s: %w", t.Slug, err)
	}
	if err := bundleitemutils.ValidateItemVersion(t.Version); err != nil {
		return fmt.Errorf("invalid version %s: %w", t.Version, err)
	}
	if strings.TrimSpace(t.DisplayName) == "" {
		return errors.New("displayName is empty")
	}
	if strings.TrimSpace(string(t.ID)) == "" {
		return errors.New("id is empty")
	}
	if t.ArgSchema == nil || !json.Valid(t.ArgSchema) {
		return errors.New("argSchema is missing or invalid")
	}

	if t.CreatedAt.IsZero() {
		return errors.New("createdAt is zero")
	}
	if t.ModifiedAt.IsZero() {
		return errors.New("modifiedAt is zero")
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

	if err := bundleitemutils.ValidateTags(t.Tags); err != nil {
		return err
	}
	return nil
}
