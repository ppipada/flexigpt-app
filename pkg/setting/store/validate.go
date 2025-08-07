package store

import (
	"fmt"

	"github.com/ppipada/flexigpt-app/pkg/setting/spec"
)

// validateTheme checks a theme for correctness.
func validateTheme(th *spec.AppTheme) error {
	if th == nil {
		return spec.ErrInvalidTheme
	}
	switch th.Type {
	case spec.ThemeSystem, spec.ThemeLight, spec.ThemeDark:
		if th.Name != getThemeString(th.Type) {
			return fmt.Errorf("%w: name required", spec.ErrInvalidTheme)
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

func getThemeString(s spec.ThemeType) string {
	return string(s)
}
