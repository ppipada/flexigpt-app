package settingstore

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	modelSpec "github.com/ppipada/flexigpt-app/pkg/model/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
)

type SettingStore struct {
	store         *filestore.MapFileStore
	defaultData   SettingsSchema
	encryptEncDec encdec.EncoderDecoder
	keyEncDec     encdec.StringEncoderDecoder
}

func InitSettingStore(settingStore *SettingStore, filename string) error {
	settingsMap, err := encdec.StructWithJSONTagsToMap(DefaultSettingsData)
	if err != nil {
		return errors.New("could not get map of settings data")
	}
	settingStore.defaultData = DefaultSettingsData
	settingStore.encryptEncDec = encdec.EncryptedStringValueEncoderDecoder{}
	settingStore.keyEncDec = encdec.Base64StringEncoderDecoder{}
	store, err := filestore.NewMapFileStore(
		filename,
		settingsMap,
		filestore.WithCreateIfNotExists(true),
		filestore.WithAutoFlush(true),
		filestore.WithValueEncDecGetter(settingStore.ValueEncDecGetter),
		// Lets not encode decode keys for now
		// filestore.WithKeyEncDecGetter(settingStore.KeyEncDecGetter).
		filestore.WithEncoderDecoder(encdec.JSONEncoderDecoder{}))
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	settingStore.store = store
	slog.Info("Store initialization done.")
	return nil
}

func (s *SettingStore) ValueEncDecGetter(pathSoFar []string) encdec.EncoderDecoder {
	if len(pathSoFar) == 3 {
		if pathSoFar[2] == "apiKey" {
			return s.encryptEncDec
		}
	}
	return nil
}

// Func (s *SettingStore) KeyEncDecGetter(pathSoFar []string) encdec.StringEncoderDecoder {
// 	// 1) If pathSoFar == ["aiSettings", <providerName>], encode providerName
// 	if len(pathSoFar) == 2 && pathSoFar[1] == "aiSettings" {
// 		return s.keyEncDec
// 	}
// 	// 2) If pathSoFar == ["aiSettings", <providerName>, modelPresets, <modelName>], encode modelName
// 	if len(pathSoFar) == 4 &&
// 		pathSoFar[0] == "aiSettings" &&
// 		pathSoFar[2] == "modelPresets" {
// 		return s.keyEncDec
// 	}
// 	// Otherwise, no key-encoding
// 	return nil
// }.

func (s *SettingStore) GetAllSettings(
	ctx context.Context,
	req *GetAllSettingsRequest,
) (*GetAllSettingsResponse, error) {
	forceFetch := false
	if req != nil {
		forceFetch = req.ForceFetch
	}
	data, err := s.store.GetAll(forceFetch)
	if err != nil {
		return nil, err
	}
	var settings SettingsSchema
	if err := encdec.MapToStructWithJSONTags(data, &settings); err != nil {
		return nil, err
	}

	return &GetAllSettingsResponse{Body: &settings}, nil
}

func (s *SettingStore) SetAppSettings(
	ctx context.Context,
	req *SetAppSettingsRequest,
) (*SetAppSettingsResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}
	_, _, _, err := s.getProviderData(req.Body.DefaultProvider, false)
	if err != nil {
		return nil, err
	}

	if err := s.store.SetKey([]string{"app", "defaultProvider"}, string(req.Body.DefaultProvider)); err != nil {
		return nil, fmt.Errorf("failed to set app settings: %w", err)
	}
	return &SetAppSettingsResponse{}, nil
}

func (s *SettingStore) AddAISetting(
	ctx context.Context,
	req *AddAISettingRequest,
) (*AddAISettingResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}

	// Pull current data.
	currentData, err := s.store.GetAll(false)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving current data: %w", err)
	}

	aiSettings, ok := currentData["aiSettings"].(map[string]any)
	if !ok {
		return nil, errors.New("aiSettings is missing or not a map")
	}

	// If it already exists, return error.
	if _, exists := aiSettings[string(req.ProviderName)]; exists {
		return nil, fmt.Errorf("provider %q already exists", req.ProviderName)
	}

	keys := []string{"aiSettings", string(req.ProviderName)}
	val, err := encdec.StructWithJSONTagsToMap(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to add provider %q: %w", req.ProviderName, err)
	}
	if err := s.store.SetKey(keys, val); err != nil {
		return nil, fmt.Errorf("failed to add provider %q: %w", req.ProviderName, err)
	}

	return &AddAISettingResponse{}, nil
}

func (s *SettingStore) DeleteAISetting(
	ctx context.Context,
	req *DeleteAISettingRequest,
) (*DeleteAISettingResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	// Pull current data.
	currentData, err := s.store.GetAll(false)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving current data: %w", err)
	}

	aiSettings, ok := currentData["aiSettings"].(map[string]any)
	if !ok {
		return nil, errors.New("aiSettings missing or not a map")
	}

	if _, exists := aiSettings[string(req.ProviderName)]; !exists {
		return nil, fmt.Errorf("provider %q does not exist", req.ProviderName)
	}

	keys := []string{"aiSettings", string(req.ProviderName)}
	if err := s.store.DeleteKey(keys); err != nil {
		return nil, fmt.Errorf("failed to delete provider %q: %w", req.ProviderName, err)
	}

	return &DeleteAISettingResponse{}, nil
}

func (s *SettingStore) getProviderData(
	providerName modelSpec.ProviderName,
	forceFetch bool,
) (currentData, aiSettings, providerData map[string]any, err error) {
	// 1) GetAll with no forced fetch (or use the boolean if you sometimes need forced).
	currentData, err = s.store.GetAll(forceFetch)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to retrieve current data: %w", err)
	}

	// 2) aiSettings must be a map.
	aiSettingsRaw, ok := currentData["aiSettings"]
	if !ok {
		return nil, nil, nil, errors.New("aiSettings missing from store")
	}
	aiSettings, ok = aiSettingsRaw.(map[string]any)
	if !ok {
		return nil, nil, nil, errors.New("aiSettings is not a map[string]any")
	}

	// 3) Check if provider exists.
	providerRaw, ok := aiSettings[string(providerName)]
	if !ok {
		return nil, nil, nil, fmt.Errorf("provider %q does not exist in aiSettings", providerName)
	}
	providerData, ok = providerRaw.(map[string]any)
	if !ok {
		return nil, nil, nil, fmt.Errorf("provider %q data is not a map[string]any", providerName)
	}

	return currentData, aiSettings, providerData, nil
}

func (s *SettingStore) SetAISettingAPIKey(
	ctx context.Context,
	req *SetAISettingAPIKeyRequest,
) (*SetAISettingAPIKeyResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}

	// If APIKey is empty, do nothing.
	if req.Body.APIKey == "" {
		return &SetAISettingAPIKeyResponse{}, nil
	}

	// Check provider exists.
	_, _, _, err := s.getProviderData(req.ProviderName, false)
	if err != nil {
		return nil, err
	}

	keys := []string{"aiSettings", string(req.ProviderName), "apiKey"}
	if err := s.store.SetKey(keys, req.Body.APIKey); err != nil {
		return nil, fmt.Errorf("failed to set AI API key: %w", err)
	}
	return &SetAISettingAPIKeyResponse{}, nil
}

func (s *SettingStore) SetAISettingAttrs(
	ctx context.Context,
	req *SetAISettingAttrsRequest,
) (*SetAISettingAttrsResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}

	// Make sure provider is in store.
	_, _, _, err := s.getProviderData(req.ProviderName, false)
	if err != nil {
		return nil, err
	}

	// For each non-nil field, set via dot key if it’s not empty (for strings).
	if req.Body.IsEnabled != nil {
		if err := s.store.SetKey([]string{"aiSettings", string(req.ProviderName), "isEnabled"}, *req.Body.IsEnabled); err != nil {
			return nil, fmt.Errorf("failed updating isEnabled: %w", err)
		}
	}
	if req.Body.Origin != nil && *req.Body.Origin != "" {
		if err := s.store.SetKey([]string{"aiSettings", string(req.ProviderName), "origin"}, *req.Body.Origin); err != nil {
			return nil, fmt.Errorf("failed updating origin: %w", err)
		}
	}
	if req.Body.ChatCompletionPathPrefix != nil && *req.Body.ChatCompletionPathPrefix != "" {
		if err := s.store.SetKey([]string{"aiSettings", string(req.ProviderName), "chatCompletionPathPrefix"}, *req.Body.ChatCompletionPathPrefix); err != nil {
			return nil, fmt.Errorf("failed updating chatCompletionPathPrefix: %w", err)
		}
	}

	return &SetAISettingAttrsResponse{}, nil
}
