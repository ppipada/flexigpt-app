package store

import (
	"fmt"

	"github.com/ppipada/flexigpt-app/internal/setting/spec"
)

// validateTheme checks a theme for correctness.
func validateTheme(th *spec.AppTheme) error {
	if th == nil {
		return spec.ErrInvalidTheme
	}
	switch th.Type {
	case spec.ThemeSystem:
		if th.Name != spec.ThemeNameSystem {
			return fmt.Errorf(
				"%w: type and name required. input - type %s, name %s",
				spec.ErrInvalidTheme,
				th.Type,
				th.Name,
			)
		}
		return nil
	case spec.ThemeLight:
		if th.Name != spec.ThemeNameLight {
			return fmt.Errorf(
				"%w: type and name required. input - type %s, name %s",
				spec.ErrInvalidTheme,
				th.Type,
				th.Name,
			)
		}
		return nil
	case spec.ThemeDark:
		if th.Name != spec.ThemeNameDark {
			return fmt.Errorf(
				"%w: type and name required. input - type %s, name %s",
				spec.ErrInvalidTheme,
				th.Type,
				th.Name,
			)
		}
		return nil
	case spec.ThemeOther:
		if th.Name == "" {
			return fmt.Errorf("%w: name required", spec.ErrInvalidTheme)
		}
		return nil
	default:
		return spec.ErrInvalidTheme
	}
}
