package main

import (
	"context"

	aiproviderConsts "github.com/flexigpt/flexiui/pkg/aiprovider/consts"
	aiproviderSpec "github.com/flexigpt/flexiui/pkg/aiprovider/spec"

	"github.com/flexigpt/flexiui/pkg/settingstore"

	"github.com/flexigpt/flexiui/pkg/settingstore/spec"
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

func InitProviderSetUsingSettings(s *SettingStoreWrapper, p *ProviderSetWrapper) error {
	allSettingsResponse, err := s.GetAllSettings(&spec.GetAllSettingsRequest{})
	if err != nil || allSettingsResponse.Body == nil {
		return err
	}

	for providerName, aiSetting := range allSettingsResponse.Body.AISettings {
		if _, exists := aiproviderConsts.InbuiltProviderModels[providerName]; exists {
			// Update inbuilt providers
			if aiSetting.APIKey != "" {
				_, err = p.SetProviderAPIKey(
					&aiproviderSpec.SetProviderAPIKeyRequest{
						Provider: providerName,
						Body: &aiproviderSpec.SetProviderAPIKeyRequestBody{
							APIKey: aiSetting.APIKey,
						},
					},
				)
				if err != nil {
					return err
				}
			}

			_, err = p.SetProviderAttribute(
				&aiproviderSpec.SetProviderAttributeRequest{
					Provider: providerName,
					Body: &aiproviderSpec.SetProviderAttributeRequestBody{
						Origin:                   &aiSetting.Origin,
						ChatCompletionPathPrefix: &aiSetting.ChatCompletionPathPrefix,
					},
				},
			)
			if err != nil {
				return err
			}
		} else {
			// Add custom providers
			_, err := p.AddProvider(&aiproviderSpec.AddProviderRequest{
				Provider: providerName,
				Body: &aiproviderSpec.AddProviderRequestBody{
					APIKey:                   aiSetting.APIKey,
					Origin:                   aiSetting.Origin,
					ChatCompletionPathPrefix: aiSetting.ChatCompletionPathPrefix,
				},
			})
			if err != nil {
				return err
			}
		}
	}
	_, err = p.SetDefaultProvider(&aiproviderSpec.SetDefaultProviderRequest{
		Body: &aiproviderSpec.SetDefaultProviderRequestBody{
			Provider: allSettingsResponse.Body.App.DefaultProvider,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// GetAllSettings retrieves all settings without requiring a context.
func (w *SettingStoreWrapper) GetAllSettings(
	req *spec.GetAllSettingsRequest,
) (*spec.GetAllSettingsResponse, error) {
	return w.store.GetAllSettings(context.Background(), req)
}

// SetSetting sets a setting without requiring a context.
// Disable updating via dot key
// func (w *SettingStoreWrapper) SetSetting(
// 	req *spec.SetSettingRequest,
// ) (*spec.SetSettingResponse, error) {
// 	return w.store.SetSetting(context.Background(), req)
// }

// SetAppSettings updates the "app" portion of settings without requiring a context.
func (w *SettingStoreWrapper) SetAppSettings(
	req *spec.SetAppSettingsRequest,
) (*spec.SetAppSettingsResponse, error) {
	return w.store.SetAppSettings(context.Background(), req)
}

// AddAISetting creates a new AI provider without requiring a context.
func (w *SettingStoreWrapper) AddAISetting(
	req *spec.AddAISettingRequest,
) (*spec.AddAISettingResponse, error) {
	return w.store.AddAISetting(context.Background(), req)
}

// DeleteAISetting removes an existing AI provider without requiring a context.
func (w *SettingStoreWrapper) DeleteAISetting(
	req *spec.DeleteAISettingRequest,
) (*spec.DeleteAISettingResponse, error) {
	return w.store.DeleteAISetting(context.Background(), req)
}

// SetAISettingAPIKey updates the API key of a provider without requiring a context.
func (w *SettingStoreWrapper) SetAISettingAPIKey(
	req *spec.SetAISettingAPIKeyRequest,
) (*spec.SetAISettingAPIKeyResponse, error) {
	return w.store.SetAISettingAPIKey(context.Background(), req)
}

// SetAISettingAttrs partially updates AI provider attributes without requiring a context.
func (w *SettingStoreWrapper) SetAISettingAttrs(
	req *spec.SetAISettingAttrsRequest,
) (*spec.SetAISettingAttrsResponse, error) {
	return w.store.SetAISettingAttrs(context.Background(), req)
}

// AddModelSetting creates or replaces a single model setting for a provider without requiring a context.
func (w *SettingStoreWrapper) AddModelSetting(
	req *spec.AddModelSettingRequest,
) (*spec.AddModelSettingResponse, error) {
	return w.store.AddModelSetting(context.Background(), req)
}

// DeleteModelSetting creates or replaces a single model setting for a provider without requiring a context.
func (w *SettingStoreWrapper) DeleteModelSetting(
	req *spec.DeleteModelSettingRequest,
) (*spec.DeleteModelSettingResponse, error) {
	return w.store.DeleteModelSetting(context.Background(), req)
}
