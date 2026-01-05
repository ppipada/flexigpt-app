// Package main (or your own wrapper package) provides a small helper layer
// that turns the context-aware APIs exposed by ModelPresetStore into simple
// “context-less” helpers while adding the panic-to-error recovery middleware.
package main

import (
	"context"

	"github.com/ppipada/flexigpt-app/internal/inferencewrapper"
	inferencewrapperSpec "github.com/ppipada/flexigpt-app/internal/inferencewrapper/spec"
	"github.com/ppipada/flexigpt-app/internal/middleware"
	"github.com/ppipada/flexigpt-app/internal/modelpreset/spec"
	modelpresetStore "github.com/ppipada/flexigpt-app/internal/modelpreset/store"
	settingSpec "github.com/ppipada/flexigpt-app/internal/setting/spec"

	inferencegoSpec "github.com/ppipada/inference-go/spec"
)

type ModelPresetStoreWrapper struct {
	store               *modelpresetStore.ModelPresetStore
	settingStoreWrapper *SettingStoreWrapper
	providerSetWrapper  *ProviderSetWrapper
}

// InitModelPresetStoreWrapper initialises the wrapped store in `baseDir`.
func InitModelPresetStoreWrapper(
	m *ModelPresetStoreWrapper,
	settingStoreWrapper *SettingStoreWrapper,
	providerSetWrapper *ProviderSetWrapper,
	baseDir string,
) error {
	if m == nil || settingStoreWrapper == nil || providerSetWrapper == nil {
		panic("initialising model-preset store wrapper on nil receivers")
	}
	s, err := modelpresetStore.NewModelPresetStore(baseDir)
	if err != nil {
		return err
	}
	m.store = s
	m.settingStoreWrapper = settingStoreWrapper
	m.providerSetWrapper = providerSetWrapper
	err = InitProviderSetUsingSettingsAndPresets(
		m,
		m.settingStoreWrapper,
		m.providerSetWrapper,
	)
	if err != nil {
		return err
	}
	return nil
}

func (w *ModelPresetStoreWrapper) PatchDefaultProvider(
	req *spec.PatchDefaultProviderRequest,
) (*spec.PatchDefaultProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.PatchDefaultProviderResponse, error) {
		return w.store.PatchDefaultProvider(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) GetDefaultProvider(
	req *spec.GetDefaultProviderRequest,
) (*spec.GetDefaultProviderResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.GetDefaultProviderResponse, error) {
		return w.store.GetDefaultProvider(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) PutProviderPreset(
	req *spec.PutProviderPresetRequest,
) (*spec.PutProviderPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.PutProviderPresetResponse, error) {
		// First try to delete from provider apis, it is ok if it is not present.
		_, _ = w.providerSetWrapper.DeleteProvider(
			&inferencewrapperSpec.DeleteProviderRequest{
				Provider: inferencegoSpec.ProviderName(string(req.ProviderName)),
			},
		)
		// Then try to add in provider apis, need to skip adding to store if it cannot be added.
		if _, err := w.providerSetWrapper.AddProvider(
			&inferencewrapperSpec.AddProviderRequest{
				Provider: inferencegoSpec.ProviderName(string(req.ProviderName)),
				Body: &inferencewrapperSpec.AddProviderRequestBody{
					SDKType:                  inferencewrapper.ConvertModelPresetToInferencegoSDKType(req.Body.SDKType),
					Origin:                   req.Body.Origin,
					ChatCompletionPathPrefix: req.Body.ChatCompletionPathPrefix,
					APIKeyHeaderKey:          req.Body.APIKeyHeaderKey,
					DefaultHeaders:           req.Body.DefaultHeaders,
				},
			}); err != nil {
			return nil, err
		}
		resp, err := w.store.PutProviderPreset(context.Background(), req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	})
}

func (w *ModelPresetStoreWrapper) PatchProviderPreset(
	req *spec.PatchProviderPresetRequest,
) (*spec.PatchProviderPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.PatchProviderPresetResponse, error) {
		return w.store.PatchProviderPreset(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) DeleteProviderPreset(
	req *spec.DeleteProviderPresetRequest,
) (*spec.DeleteProviderPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.DeleteProviderPresetResponse, error) {
		_, err := w.settingStoreWrapper.DeleteAuthKey(
			&settingSpec.DeleteAuthKeyRequest{
				Type:    settingSpec.AuthKeyTypeProvider,
				KeyName: settingSpec.AuthKeyName(req.ProviderName),
			},
		)
		if err != nil {
			return nil, err
		}
		_, _ = w.providerSetWrapper.DeleteProvider(
			&inferencewrapperSpec.DeleteProviderRequest{
				Provider: inferencegoSpec.ProviderName(string(req.ProviderName)),
			},
		)
		resp, err := w.store.DeleteProviderPreset(context.Background(), req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	})
}

func (w *ModelPresetStoreWrapper) ListProviderPresets(
	req *spec.ListProviderPresetsRequest,
) (*spec.ListProviderPresetsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.ListProviderPresetsResponse, error) {
		return w.store.ListProviderPresets(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) PutModelPreset(
	req *spec.PutModelPresetRequest,
) (*spec.PutModelPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.PutModelPresetResponse, error) {
		return w.store.PutModelPreset(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) PatchModelPreset(
	req *spec.PatchModelPresetRequest,
) (*spec.PatchModelPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.PatchModelPresetResponse, error) {
		return w.store.PatchModelPreset(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) DeleteModelPreset(
	req *spec.DeleteModelPresetRequest,
) (*spec.DeleteModelPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.DeleteModelPresetResponse, error) {
		return w.store.DeleteModelPreset(context.Background(), req)
	})
}
