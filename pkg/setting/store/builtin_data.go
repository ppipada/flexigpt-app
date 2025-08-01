package store

import (
	"slices"

	"github.com/ppipada/flexigpt-app/pkg/builtin"
	"github.com/ppipada/flexigpt-app/pkg/setting/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
)

// BuiltInAuthKeys lists every key that must never be deleted.
var BuiltInAuthKeys = func() map[spec.AuthKeyType][]spec.AuthKeyName {
	m := map[spec.AuthKeyType][]spec.AuthKeyName{
		spec.AuthKeyTypeProvider: {},
	}
	for _, p := range builtin.BuiltInProviderNames {
		m[spec.AuthKeyTypeProvider] = append(
			m[spec.AuthKeyTypeProvider],
			spec.AuthKeyName(p),
		)
	}
	return m
}()

// DefaultSettingsData is written to disk on first start.
var DefaultSettingsData = func() spec.SettingsSchema {
	ak := spec.AuthKeysSchema{
		spec.AuthKeyTypeProvider: map[spec.AuthKeyName]spec.AuthKey{},
	}
	for _, p := range builtin.BuiltInProviderNames {
		ak[spec.AuthKeyTypeProvider][spec.AuthKeyName(p)] = spec.AuthKey{
			Secret:   "",
			SHA256:   encdec.ComputeSHA(""),
			NonEmpty: false,
		}
	}
	return spec.SettingsSchema{
		SchemaVersion: spec.SchemaVersion,
		AppTheme:      spec.AppTheme{Type: spec.ThemeSystem, Name: string(spec.ThemeSystem)},
		AuthKeys:      ak,
	}
}()

// isBuiltInKey reports whether type/keyName is part of BuiltInAuthKeys.
func isBuiltInKey(t spec.AuthKeyType, name spec.AuthKeyName) bool {
	if names, ok := BuiltInAuthKeys[t]; ok {
		return slices.Contains(names, name)
	}
	return false
}
