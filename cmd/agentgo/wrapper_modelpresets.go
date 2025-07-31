// Package main (or your own wrapper package) provides a small helper layer
// that turns the context-aware APIs exposed by ModelPresetStore into simple
// “context-less” helpers while adding the panic-to-error recovery middleware.
package main

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/middleware"
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	modelpresetStore "github.com/ppipada/flexigpt-app/pkg/modelpreset/store"
)

type ModelPresetStoreWrapper struct {
	store *modelpresetStore.ModelPresetStore
}

// InitModelPresetStoreWrapper initialises the wrapped store in `baseDir`.
func InitModelPresetStoreWrapper(m *ModelPresetStoreWrapper, baseDir string) error {
	if m == nil {
		panic("initialising model-preset store wrapper on a nil receiver")
	}
	s, err := modelpresetStore.NewModelPresetStore(baseDir)
	if err != nil {
		return err
	}
	m.store = s
	return nil
}

func (w *ModelPresetStoreWrapper) PutProviderPreset(
	req *spec.PutProviderPresetRequest,
) (*spec.PutProviderPresetResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.PutProviderPresetResponse, error) {
		return w.store.PutProviderPreset(context.Background(), req)
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
		return w.store.DeleteProviderPreset(context.Background(), req)
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
