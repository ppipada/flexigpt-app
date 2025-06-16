package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ppipada/flexigpt-app/pkg/model/consts"
	"github.com/ppipada/flexigpt-app/pkg/model/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
)

type ModelPresetsStore struct {
	store         *filestore.MapFileStore
	defaultData   spec.ModelPresetsSchema
	encryptEncDec encdec.EncoderDecoder
	keyEncDec     encdec.StringEncoderDecoder
}

func InitModelPresetsStore(mpStore *ModelPresetsStore, filename string) error {
	modelPresetsMap, err := encdec.StructWithJSONTagsToMap(consts.DefaultModelPresetsSchema)
	if err != nil {
		return errors.New("could not get map of model presets data")
	}
	mpStore.defaultData = consts.DefaultModelPresetsSchema
	mpStore.encryptEncDec = encdec.EncryptedStringValueEncoderDecoder{}
	mpStore.keyEncDec = encdec.Base64StringEncoderDecoder{}
	store, err := filestore.NewMapFileStore(
		filename,
		modelPresetsMap,
		filestore.WithCreateIfNotExists(true),
		filestore.WithAutoFlush(true),
		filestore.WithEncoderDecoder(encdec.JSONEncoderDecoder{}))
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	mpStore.store = store
	slog.Info("Model presets store initialization done.")
	return nil
}

func (s *ModelPresetsStore) GetAllModelPresets(
	ctx context.Context,
	req *spec.GetAllModelPresetsRequest,
) (*spec.GetAllModelPresetsResponse, error) {
	forceFetch := false
	if req != nil {
		forceFetch = req.ForceFetch
	}
	data, err := s.store.GetAll(forceFetch)
	if err != nil {
		return nil, err
	}
	var presets spec.ModelPresetsSchema
	if err := encdec.MapToStructWithJSONTags(data, &presets); err != nil {
		return nil, err
	}

	return &spec.GetAllModelPresetsResponse{Body: &presets}, nil
}

func (s *ModelPresetsStore) CreateModelPresets(
	ctx context.Context,
	req *spec.CreateModelPresetsRequest,
) (*spec.CreateModelPresetsResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}

	// Pull current data.
	currentData, err := s.store.GetAll(false)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving current data: %w", err)
	}

	modelPresets, ok := currentData["modelPresets"].(map[string]any)
	if !ok {
		return nil, errors.New("modelPresets is missing or not a map")
	}

	// If it already exists, return error.
	if _, exists := modelPresets[string(req.ProviderName)]; exists {
		return nil, fmt.Errorf("provider %q already exists", req.ProviderName)
	}

	keys := []string{"modelPresets", string(req.ProviderName)}
	val, err := encdec.StructWithJSONTagsToMap(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to add provider %q: %w", req.ProviderName, err)
	}
	if err := s.store.SetKey(keys, val); err != nil {
		return nil, fmt.Errorf("failed to add provider %q: %w", req.ProviderName, err)
	}

	return &spec.CreateModelPresetsResponse{}, nil
}

func (s *ModelPresetsStore) DeleteModelPresets(
	ctx context.Context,
	req *spec.DeleteModelPresetsRequest,
) (*spec.DeleteModelPresetsResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	// Pull current data.
	currentData, err := s.store.GetAll(false)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving current data: %w", err)
	}

	modelPresets, ok := currentData["modelPresets"].(map[string]any)
	if !ok {
		return nil, errors.New("modelPresets missing or not a map")
	}

	if _, exists := modelPresets[string(req.ProviderName)]; !exists {
		return nil, fmt.Errorf("provider %q does not exist", req.ProviderName)
	}

	keys := []string{"modelPresets", string(req.ProviderName)}
	if err := s.store.DeleteKey(keys); err != nil {
		return nil, fmt.Errorf("failed to delete provider %q: %w", req.ProviderName, err)
	}

	return &spec.DeleteModelPresetsResponse{}, nil
}

func (s *ModelPresetsStore) getProviderData(
	providerName spec.ProviderName,
	forceFetch bool,
) (currentData, modelPresets, providerData map[string]any, err error) {
	// 1) GetAll with no forced fetch (or use the boolean if you sometimes need forced).
	currentData, err = s.store.GetAll(forceFetch)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to retrieve current data: %w", err)
	}

	// 2) modelPresets must be a map.
	modelPresetsRaw, ok := currentData["modelPresets"]
	if !ok {
		return nil, nil, nil, errors.New("modelPresets missing from store")
	}
	modelPresets, ok = modelPresetsRaw.(map[string]any)
	if !ok {
		return nil, nil, nil, errors.New("modelPresets is not a map[string]any")
	}

	// 3) Check if provider exists.
	providerRaw, ok := modelPresets[string(providerName)]
	if !ok {
		return nil, nil, nil, fmt.Errorf("provider %q does not exist in modelPresets", providerName)
	}
	providerData, ok = providerRaw.(map[string]any)
	if !ok {
		return nil, nil, nil, fmt.Errorf("provider %q data is not a map[string]any", providerName)
	}

	return currentData, modelPresets, providerData, nil
}

func (s *ModelPresetsStore) AddModelPreset(
	ctx context.Context,
	req *spec.AddModelPresetRequest,
) (*spec.AddModelPresetResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}

	// Confirm provider existence.
	_, _, _, err := s.getProviderData(req.ProviderName, false)
	if err != nil {
		return nil, err
	}

	// Overwrite or create the model.
	keys := []string{
		"modelPresets",
		string(req.ProviderName),
		string(req.ModelName),
	}
	val, err := encdec.StructWithJSONTagsToMap(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed putting preset model %q for provider %q: %w",
			req.ModelName, req.ProviderName, err)
	}
	if err := s.store.SetKey(keys, val); err != nil {
		return nil, fmt.Errorf("failed putting preset model %q for provider %q: %w",
			req.ModelName, req.ProviderName, err)
	}
	return &spec.AddModelPresetResponse{}, nil
}

func (s *ModelPresetsStore) DeleteModelPreset(
	ctx context.Context,
	req *spec.DeleteModelPresetRequest,
) (*spec.DeleteModelPresetResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Confirm provider existence.
	_, _, _, err := s.getProviderData(req.ProviderName, false)
	if err != nil {
		return nil, err
	}

	// Delete the model.
	keys := []string{"modelPresets", string(req.ProviderName), string(req.ModelName)}
	if err := s.store.DeleteKey(keys); err != nil {
		return nil, fmt.Errorf("failed deleting preset model %q for provider %q: %w",
			req.ModelName, req.ProviderName, err)
	}
	return &spec.DeleteModelPresetResponse{}, nil
}
