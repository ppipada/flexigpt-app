package main

import (
	"context"
	"log/slog"

	"github.com/ppipada/flexigpt-app/pkg/inference"
	"github.com/ppipada/flexigpt-app/pkg/middleware"
	modelConsts "github.com/ppipada/flexigpt-app/pkg/model/consts"
	"github.com/ppipada/flexigpt-app/pkg/settingstore"
	"github.com/ppipada/flexigpt-app/pkg/settingstore/spec"
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
		if _, exists := modelConsts.InbuiltProviderModels[providerName]; exists {
			// Update inbuilt providers.
			if aiSetting.APIKey != "" {
				_, err = p.SetProviderAPIKey(
					&inference.SetProviderAPIKeyRequest{
						Provider: providerName,
						Body: &inference.SetProviderAPIKeyRequestBody{
							APIKey: aiSetting.APIKey,
						},
					},
				)
				if err != nil {
					return err
				}
			}

			_, err = p.SetProviderAttribute(
				&inference.SetProviderAttributeRequest{
					Provider: providerName,
					Body: &inference.SetProviderAttributeRequestBody{
						Origin:                   &aiSetting.Origin,
						ChatCompletionPathPrefix: &aiSetting.ChatCompletionPathPrefix,
					},
				},
			)
			if err != nil {
				return err
			}
		} else {
			// Add custom providers.
			_, err := p.AddProvider(&inference.AddProviderRequest{
				Provider: providerName,
				Body: &inference.AddProviderRequestBody{
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
	_, err = p.SetDefaultProvider(&inference.SetDefaultProviderRequest{
		Body: &inference.SetDefaultProviderRequestBody{
			Provider: allSettingsResponse.Body.App.DefaultProvider,
		},
	})
	if err != nil {
		return err
	}
	slog.Info("InitProviderSetUsingSettings Done.")
	return nil
}

// GetAllSettings retrieves all settings without requiring a context.
func (w *SettingStoreWrapper) GetAllSettings(
	req *spec.GetAllSettingsRequest,
) (*spec.GetAllSettingsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.GetAllSettingsResponse, error) {
		return w.store.GetAllSettings(context.Background(), req)
	})
}

// SetAppSettings updates the "app" portion of settings without requiring a context.
func (w *SettingStoreWrapper) SetAppSettings(
	req *spec.SetAppSettingsRequest,
) (*spec.SetAppSettingsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.SetAppSettingsResponse, error) {
		return w.store.SetAppSettings(context.Background(), req)
	})
}

// AddAISetting creates a new AI provider without requiring a context.
func (w *SettingStoreWrapper) AddAISetting(
	req *spec.AddAISettingRequest,
) (*spec.AddAISettingResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.AddAISettingResponse, error) {
		return w.store.AddAISetting(context.Background(), req)
	})
}

// DeleteAISetting removes an existing AI provider without requiring a context.
func (w *SettingStoreWrapper) DeleteAISetting(
	req *spec.DeleteAISettingRequest,
) (*spec.DeleteAISettingResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.DeleteAISettingResponse, error) {
		return w.store.DeleteAISetting(context.Background(), req)
	})
}

// SetAISettingAPIKey updates the API key of a provider without requiring a context.
func (w *SettingStoreWrapper) SetAISettingAPIKey(
	req *spec.SetAISettingAPIKeyRequest,
) (*spec.SetAISettingAPIKeyResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.SetAISettingAPIKeyResponse, error) {
		return w.store.SetAISettingAPIKey(context.Background(), req)
	})
}

// SetAISettingAttrs partially updates AI provider attributes without requiring a context.
func (w *SettingStoreWrapper) SetAISettingAttrs(
	req *spec.SetAISettingAttrsRequest,
) (*spec.SetAISettingAttrsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.SetAISettingAttrsResponse, error) {
		return w.store.SetAISettingAttrs(context.Background(), req)
	})
}

// AddModelPreset creates or replaces a single model setting for a provider without requiring a context.
func (w *SettingStoreWrapper) AddModelPreset(
	req *spec.AddModelPresetRequest,
) (*spec.AddModelPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.AddModelPresetResponse, error) {
		return w.store.AddModelPreset(context.Background(), req)
	})
}

// DeleteModelPreset creates or replaces a single model setting for a provider without requiring a context.
func (w *SettingStoreWrapper) DeleteModelPreset(
	req *spec.DeleteModelPresetRequest,
) (*spec.DeleteModelPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.DeleteModelPresetResponse, error) {
		return w.store.DeleteModelPreset(context.Background(), req)
	})
}
