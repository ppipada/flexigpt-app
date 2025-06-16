package main

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/middleware"
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
	modelStore "github.com/ppipada/flexigpt-app/pkg/model/store"
)

// ModelPresetStoreWrapper is a wrapper around ModelPresetStore that provides non-contextual APIs.
type ModelPresetStoreWrapper struct {
	store *modelStore.ModelPresetStore
}

func InitModelPresetStoreWrapper(m *ModelPresetStoreWrapper, filename string) error {
	if m == nil {
		panic("Initializing model presets store without a object")
	}
	modelPresetStore := &modelStore.ModelPresetStore{}
	err := modelStore.InitModelPresetStore(modelPresetStore, filename)
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

func (w *ModelPresetStoreWrapper) CreateModelPresets(
	req *spec.CreateModelPresetsRequest,
) (*spec.CreateModelPresetsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.CreateModelPresetsResponse, error) {
		return w.store.CreateModelPresets(context.Background(), req)
	})
}

func (w *ModelPresetStoreWrapper) DeleteModelPresets(
	req *spec.DeleteModelPresetsRequest,
) (*spec.DeleteModelPresetsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.DeleteModelPresetsResponse, error) {
		return w.store.DeleteModelPresets(context.Background(), req)
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
