package main

import (
	"context"

	"github.com/flexigpt/flexiui/pkggo/settingstore"

	"github.com/flexigpt/flexiui/pkggo/settingstore/spec"
)

// SettingStoreWrapper is a wrapper around SettingStore that provides non-contextual APIs.
type SettingStoreWrapper struct {
	store *settingstore.SettingStore
}

func InitSettingStoreWrapper(s *SettingStoreWrapper, filename string) error {
	if s == nil {
		panic("Initializing settings store without a object")
	}
	settingStore := &settingstore.SettingStore{}
	err := settingstore.InitSettingStore(settingStore, filename)
	if err != nil {
		return err
	}
	s.store = settingStore
	return nil
}

// GetAllSettings retrieves all settings without requiring a context.
func (w *SettingStoreWrapper) GetAllSettings(
	req *spec.GetAllSettingsRequest,
) (*spec.GetAllSettingsResponse, error) {
	return w.store.GetAllSettings(context.Background(), req)
}

// SetSetting sets a setting without requiring a context.
func (w *SettingStoreWrapper) SetSetting(
	req *spec.SetSettingRequest,
) (*spec.SetSettingResponse, error) {
	return w.store.SetSetting(context.Background(), req)
}
