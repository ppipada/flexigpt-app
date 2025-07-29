package main

import (
	"context"
	"log/slog"

	inferenceSpec "github.com/ppipada/flexigpt-app/pkg/inference/spec"
	"github.com/ppipada/flexigpt-app/pkg/middleware"
	modelConsts "github.com/ppipada/flexigpt-app/pkg/model/consts"
	"github.com/ppipada/flexigpt-app/pkg/settingstore"
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
	allSettingsResponse, err := s.GetAllSettings(&settingstore.GetAllSettingsRequest{})
	if err != nil || allSettingsResponse.Body == nil {
		return err
	}

	for providerName, aiSetting := range allSettingsResponse.Body.AISettings {
		if _, exists := modelConsts.InbuiltProviderModels[providerName]; exists {
			// Update inbuilt providers.
			if aiSetting.APIKey != "" {
				_, err = p.SetProviderAPIKey(
					&inferenceSpec.SetProviderAPIKeyRequest{
						Provider: providerName,
						Body: &inferenceSpec.SetProviderAPIKeyRequestBody{
							APIKey: aiSetting.APIKey,
						},
					},
				)
				if err != nil {
					return err
				}
			}

			_, err = p.SetProviderAttribute(
				&inferenceSpec.SetProviderAttributeRequest{
					Provider: providerName,
					Body: &inferenceSpec.SetProviderAttributeRequestBody{
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
			_, err := p.AddProvider(&inferenceSpec.AddProviderRequest{
				Provider: providerName,
				Body: &inferenceSpec.AddProviderRequestBody{
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
	_, err = p.SetDefaultProvider(&inferenceSpec.SetDefaultProviderRequest{
		Body: &inferenceSpec.SetDefaultProviderRequestBody{
			Provider: allSettingsResponse.Body.App.DefaultProvider,
		},
	})
	if err != nil {
		return err
	}
	slog.Info("initProviderSetUsingSettings Done.")
	return nil
}

func (w *SettingStoreWrapper) GetAllSettings(
	req *settingstore.GetAllSettingsRequest,
) (*settingstore.GetAllSettingsResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingstore.GetAllSettingsResponse, error) {
		return w.store.GetAllSettings(context.Background(), req)
	})
}

func (w *SettingStoreWrapper) SetAppSettings(
	req *settingstore.SetAppSettingsRequest,
) (*settingstore.SetAppSettingsResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingstore.SetAppSettingsResponse, error) {
		return w.store.SetAppSettings(context.Background(), req)
	})
}

func (w *SettingStoreWrapper) AddAISetting(
	req *settingstore.AddAISettingRequest,
) (*settingstore.AddAISettingResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingstore.AddAISettingResponse, error) {
		return w.store.AddAISetting(context.Background(), req)
	})
}

func (w *SettingStoreWrapper) DeleteAISetting(
	req *settingstore.DeleteAISettingRequest,
) (*settingstore.DeleteAISettingResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingstore.DeleteAISettingResponse, error) {
		return w.store.DeleteAISetting(context.Background(), req)
	})
}

func (w *SettingStoreWrapper) SetAISettingAPIKey(
	req *settingstore.SetAISettingAPIKeyRequest,
) (*settingstore.SetAISettingAPIKeyResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingstore.SetAISettingAPIKeyResponse, error) {
		return w.store.SetAISettingAPIKey(context.Background(), req)
	})
}

func (w *SettingStoreWrapper) SetAISettingAttrs(
	req *settingstore.SetAISettingAttrsRequest,
) (*settingstore.SetAISettingAttrsResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingstore.SetAISettingAttrsResponse, error) {
		return w.store.SetAISettingAttrs(context.Background(), req)
	})
}
