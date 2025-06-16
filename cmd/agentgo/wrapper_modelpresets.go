package main

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/middleware"
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
	modelStore "github.com/ppipada/flexigpt-app/pkg/model/store"
)

// ModelPresetsStoreWrapper is a wrapper around ModelPresetsStore that provides non-contextual APIs.
type ModelPresetsStoreWrapper struct {
	store *modelStore.ModelPresetsStore
}

func InitModelPresetsStoreWrapper(m *ModelPresetsStoreWrapper, filename string) error {
	if m == nil {
		panic("Initializing model presets store without a object")
	}
	modelPresetsStore := &modelStore.ModelPresetsStore{}
	err := modelStore.InitModelPresetsStore(modelPresetsStore, filename)
	if err != nil {
		return err
	}
	m.store = modelPresetsStore
	return nil
}

func (w *ModelPresetsStoreWrapper) GetAllModelPresets(
	req *spec.GetAllModelPresetsRequest,
) (*spec.GetAllModelPresetsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.GetAllModelPresetsResponse, error) {
		return w.store.GetAllModelPresets(context.Background(), req)
	})
}

func (w *ModelPresetsStoreWrapper) CreateModelPresets(
	req *spec.CreateModelPresetsRequest,
) (*spec.CreateModelPresetsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.CreateModelPresetsResponse, error) {
		return w.store.CreateModelPresets(context.Background(), req)
	})
}

func (w *ModelPresetsStoreWrapper) DeleteModelPresets(
	req *spec.DeleteModelPresetsRequest,
) (*spec.DeleteModelPresetsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.DeleteModelPresetsResponse, error) {
		return w.store.DeleteModelPresets(context.Background(), req)
	})
}

func (w *ModelPresetsStoreWrapper) AddModelPreset(
	req *spec.AddModelPresetRequest,
) (*spec.AddModelPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.AddModelPresetResponse, error) {
		return w.store.AddModelPreset(context.Background(), req)
	})
}

func (w *ModelPresetsStoreWrapper) DeleteModelPreset(
	req *spec.DeleteModelPresetRequest,
) (*spec.DeleteModelPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.DeleteModelPresetResponse, error) {
		return w.store.DeleteModelPreset(context.Background(), req)
	})
}
