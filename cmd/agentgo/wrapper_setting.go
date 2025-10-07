package main

import (
	"context"

	inferenceSpec "github.com/ppipada/flexigpt-app/pkg/inference/spec"
	"github.com/ppipada/flexigpt-app/pkg/middleware"
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"

	settingSpec "github.com/ppipada/flexigpt-app/pkg/setting/spec"
	settingStore "github.com/ppipada/flexigpt-app/pkg/setting/store"
)

type SettingStoreWrapper struct {
	store              *settingStore.SettingStore
	providerSetWrapper *ProviderSetWrapper
}

// InitSettingStoreWrapper boots the underlying store and remembers the pointer.
func InitSettingStoreWrapper(
	w *SettingStoreWrapper,
	providerSetWrapper *ProviderSetWrapper,
	baseDir string,
) error {
	if w == nil || providerSetWrapper == nil {
		panic("initialising SettingStoreWrapper with <nil> receivers")
	}
	ss, err := settingStore.NewSettingStore(baseDir)
	if err != nil {
		return err
	}
	w.store = ss
	w.providerSetWrapper = providerSetWrapper
	return nil
}

func (w *SettingStoreWrapper) SetAppTheme(
	req *settingSpec.SetAppThemeRequest,
) (*settingSpec.SetAppThemeResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingSpec.SetAppThemeResponse, error) {
		return w.store.SetAppTheme(context.Background(), req)
	})
}

func (w *SettingStoreWrapper) GetSettings(
	req *settingSpec.GetSettingsRequest,
) (*settingSpec.GetSettingsResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingSpec.GetSettingsResponse, error) {
		return w.store.GetSettings(context.Background(), req)
	})
}

func (w *SettingStoreWrapper) GetAuthKey(
	req *settingSpec.GetAuthKeyRequest,
) (*settingSpec.GetAuthKeyResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingSpec.GetAuthKeyResponse, error) {
		return w.store.GetAuthKey(context.Background(), req)
	})
}

func (w *SettingStoreWrapper) SetAuthKey(
	req *settingSpec.SetAuthKeyRequest,
) (*settingSpec.SetAuthKeyResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingSpec.SetAuthKeyResponse, error) {
		if req.Type == settingSpec.AuthKeyTypeProvider {
			_, err := w.providerSetWrapper.SetProviderAPIKey(
				&inferenceSpec.SetProviderAPIKeyRequest{
					Provider: spec.ProviderName(req.KeyName),
					Body:     &inferenceSpec.SetProviderAPIKeyRequestBody{APIKey: req.Body.Secret},
				},
			)
			if err != nil {
				return nil, err
			}
		}
		resp, err := w.store.SetAuthKey(context.Background(), req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	})
}

func (w *SettingStoreWrapper) DeleteAuthKey(
	req *settingSpec.DeleteAuthKeyRequest,
) (*settingSpec.DeleteAuthKeyResponse, error) {
	return middleware.WithRecoveryResp(func() (*settingSpec.DeleteAuthKeyResponse, error) {
		resp, err := w.store.DeleteAuthKey(context.Background(), req)
		if err != nil {
			return nil, err
		}
		if req.Type == settingSpec.AuthKeyTypeProvider {
			_, _ = w.providerSetWrapper.SetProviderAPIKey(
				&inferenceSpec.SetProviderAPIKeyRequest{
					Provider: spec.ProviderName(req.KeyName),
					Body:     &inferenceSpec.SetProviderAPIKeyRequestBody{APIKey: ""},
				},
			)
		}
		return resp, nil
	})
}
