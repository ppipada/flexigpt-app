package main

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/middleware"
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	modelpresetStore "github.com/ppipada/flexigpt-app/pkg/modelpreset/store"
)

// ModelPresetStoreWrapper is a wrapper around ModelPresetStore that provides non-contextual APIs.
type ModelPresetStoreWrapper struct {
	store *modelpresetStore.ModelPresetStore
}

func InitModelPresetStoreWrapper(m *ModelPresetStoreWrapper, filename string) error {
	if m == nil {
		panic("Initializing model presets store without a object")
	}
	modelPresetStore := &modelpresetStore.ModelPresetStore{}
	err := modelpresetStore.InitModelPresetStore(modelPresetStore, filename)
	if err != nil {
		return err
	}
	m.store = modelPresetStore
	return nil
}

func (w *ModelPresetStoreWrapper) GetAllModelPresets(
	req *spec.GetAllModelPresetsRequest,
) (*spec.GetAllModelPresetsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.GetAllModelPresetsResponse, error) {
		return w.store.GetAllModelPresets(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) CreateProviderPreset(
	req *spec.CreateProviderPresetRequest,
) (*spec.CreateProviderPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.CreateProviderPresetResponse, error) {
		return w.store.CreateProviderPreset(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) DeleteProviderPreset(
	req *spec.DeleteProviderPresetRequest,
) (*spec.DeleteProviderPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.DeleteProviderPresetResponse, error) {
		return w.store.DeleteProviderPreset(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) AddModelPreset(
	req *spec.AddModelPresetRequest,
) (*spec.AddModelPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.AddModelPresetResponse, error) {
		return w.store.AddModelPreset(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) DeleteModelPreset(
	req *spec.DeleteModelPresetRequest,
) (*spec.DeleteModelPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.DeleteModelPresetResponse, error) {
		return w.store.DeleteModelPreset(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) SetDefaultModelPreset(
	req *spec.SetDefaultModelPresetRequest,
) (*spec.SetDefaultModelPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.SetDefaultModelPresetResponse, error) {
		return w.store.SetDefaultModelPreset(context.Background(), req)
	})
}
