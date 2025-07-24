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

type ModelPresetStore struct {
	store         *filestore.MapFileStore
	defaultData   spec.PresetsSchema
	encryptEncDec encdec.EncoderDecoder
	keyEncDec     encdec.StringEncoderDecoder
}

func InitModelPresetStore(mpStore *ModelPresetStore, filename string) error {
	modelPresetsMap, err := encdec.StructWithJSONTagsToMap(consts.DefaultPresetsSchema)
	if err != nil {
		return errors.New("could not get map of model presets data")
	}
	mpStore.defaultData = consts.DefaultPresetsSchema
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

func (s *ModelPresetStore) GetAllModelPresets(
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
	var presets spec.PresetsSchema
	if err := encdec.MapToStructWithJSONTags(data, &presets); err != nil {
		return nil, err
	}

	return &spec.GetAllModelPresetsResponse{Body: &presets}, nil
}

func (s *ModelPresetStore) CreateProviderPreset(
	ctx context.Context,
	req *spec.CreateProviderPresetRequest,
) (*spec.CreateProviderPresetResponse, error) {
	if req == nil || req.Body == nil || req.Body.DefaultModelPresetID == "" ||
		req.Body.ModelPresets == nil ||
		len(req.Body.ModelPresets) == 0 {
		return nil, errors.New("invalid request")
	}

	// Pull current data.
	currentData, err := s.store.GetAll(false)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving current data: %w", err)
	}

	providerPresets, ok := currentData["providerPresets"].(map[string]any)
	if !ok {
		return nil, errors.New("providerPresets is missing or not a map")
	}

	// If it already exists, return error.
	if _, exists := providerPresets[string(req.ProviderName)]; exists {
		return nil, fmt.Errorf("provider %q already exists", req.ProviderName)
	}

	keys := []string{"providerPresets", string(req.ProviderName)}
	val, err := encdec.StructWithJSONTagsToMap(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to add provider %q: %w", req.ProviderName, err)
	}
	if err := s.store.SetKey(keys, val); err != nil {
		return nil, fmt.Errorf("failed to add provider %q: %w", req.ProviderName, err)
	}

	return &spec.CreateProviderPresetResponse{}, nil
}

func (s *ModelPresetStore) DeleteProviderPreset(
	ctx context.Context,
	req *spec.DeleteProviderPresetRequest,
) (*spec.DeleteProviderPresetResponse, error) {
	if req == nil {
		return nil, errors.New("invalid request")
	}
	// Pull current data.
	currentData, err := s.store.GetAll(false)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving current data: %w", err)
	}

	providerPresets, ok := currentData["providerPresets"].(map[string]any)
	if !ok {
		return nil, errors.New("providerPresets missing or not a map")
	}

	if _, exists := providerPresets[string(req.ProviderName)]; !exists {
		return nil, fmt.Errorf("provider %q does not exist", req.ProviderName)
	}

	keys := []string{"providerPresets", string(req.ProviderName)}
	if err := s.store.DeleteKey(keys); err != nil {
		return nil, fmt.Errorf("failed to delete provider %q: %w", req.ProviderName, err)
	}

	return &spec.DeleteProviderPresetResponse{}, nil
}

func (s *ModelPresetStore) AddModelPreset(
	ctx context.Context,
	req *spec.AddModelPresetRequest,
) (*spec.AddModelPresetResponse, error) {
	if req == nil || req.Body == nil || req.ProviderName == "" || req.ModelPresetID == "" ||
		req.Body.ID == "" || req.Body.Name == "" || req.Body.DisplayName == "" || req.Body.ShortCommand == "" {
		return nil, errors.New("invalid request")
	}

	// Confirm provider existence.
	_, _, _, err := s.getProviderData(req.ProviderName, false)
	if err != nil {
		return nil, err
	}

	// Overwrite or create the model.
	keys := []string{
		"providerPresets",
		string(req.ProviderName),
		"modelPresets",
		string(req.ModelPresetID),
	}
	val, err := encdec.StructWithJSONTagsToMap(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed putting preset model %q for provider %q: %w",
			req.ModelPresetID, req.ProviderName, err)
	}
	if err := s.store.SetKey(keys, val); err != nil {
		return nil, fmt.Errorf("failed putting preset model %q for provider %q: %w",
			req.ModelPresetID, req.ProviderName, err)
	}
	return &spec.AddModelPresetResponse{}, nil
}

func (s *ModelPresetStore) DeleteModelPreset(
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
	keys := []string{
		"providerPresets",
		string(req.ProviderName),
		"modelPresets",
		string(req.ModelPresetID),
	}
	if err := s.store.DeleteKey(keys); err != nil {
		return nil, fmt.Errorf("failed deleting preset model %q for provider %q: %w",
			req.ModelPresetID, req.ProviderName, err)
	}
	return &spec.DeleteModelPresetResponse{}, nil
}

func (s *ModelPresetStore) SetDefaultModelPreset(
	ctx context.Context,
	req *spec.SetDefaultModelPresetRequest,
) (*spec.SetDefaultModelPresetResponse, error) {
	if req == nil || req.Body == nil || req.Body.ModelPresetID == "" {
		return nil, errors.New("invalid request")
	}

	// Make sure provider is in store.
	_, _, providerPreset, err := s.getProviderData(req.ProviderName, false)
	if err != nil {
		return nil, err
	}

	modelPresetsRaw, ok := providerPreset["modelPresets"]
	if !ok {
		return nil, fmt.Errorf(
			"provider %q: modelpresets doesnt exist",
			req.ProviderName,
		)
	}

	modelPresets, ok := modelPresetsRaw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"provider %q modelpresets data is not a map[string]any",
			req.ProviderName,
		)
	}

	if _, ok := modelPresets[string(req.Body.ModelPresetID)]; !ok {
		return nil, errors.New("model preset not found. id: " + string(req.Body.ModelPresetID))
	}
	keys := []string{
		"providerPresets",
		string(req.ProviderName),
		"defaultModelPresetID",
	}
	if err := s.store.SetKey(keys, req.Body.ModelPresetID); err != nil {
		return nil, fmt.Errorf("failed updating defaultModelPresetID: %w", err)
	}
	return &spec.SetDefaultModelPresetResponse{}, nil
}

func (s *ModelPresetStore) getProviderData(
	providerName spec.ProviderName,
	forceFetch bool,
) (allData, allProviderPresets, providerPreset map[string]any, err error) {
	// 1) GetAll with no forced fetch (or use the boolean if you sometimes need forced).
	allData, err = s.store.GetAll(forceFetch)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to retrieve current data: %w", err)
	}

	// 2) providerPresets must be a map.
	providerPresetsRaw, ok := allData["providerPresets"]
	if !ok {
		return nil, nil, nil, errors.New("providerPresets missing from store")
	}
	allProviderPresets, ok = providerPresetsRaw.(map[string]any)
	if !ok {
		return nil, nil, nil, errors.New("providerPresets is not a map[string]any")
	}

	// 3) Check if provider exists.
	providerRaw, ok := allProviderPresets[string(providerName)]
	if !ok {
		return nil, nil, nil, fmt.Errorf(
			"provider %q does not exist in providerPresets",
			providerName,
		)
	}
	providerPreset, ok = providerRaw.(map[string]any)
	if !ok {
		return nil, nil, nil, fmt.Errorf("provider %q data is not a map[string]any", providerName)
	}

	return allData, allProviderPresets, providerPreset, nil
}
