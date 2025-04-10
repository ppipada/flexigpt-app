package settingstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	aiproviderSpec "github.com/flexigpt/flexiui/pkg/aiprovider/spec"
	"github.com/flexigpt/flexiui/pkg/settingstore/spec"
	"github.com/flexigpt/flexiui/pkg/simplemapdb/encdec"
	"github.com/flexigpt/flexiui/pkg/simplemapdb/filestore"
)

type SettingStore struct {
	store       *filestore.MapFileStore
	defaultData spec.SettingsSchema
}

func InitSettingStore(settingStore *SettingStore, filename string) error {
	settingsMap, err := encdec.StructWithJSONTagsToMap(spec.DefaultSettingsData)
	if err != nil {
		return errors.New("could not get map of settings data")
	}
	store, err := filestore.NewMapFileStore(
		filename,
		settingsMap,
		filestore.WithCreateIfNotExists(true),
		filestore.WithAutoFlush(true),
		filestore.WithKeyEncDecsGetter(settingStore.KeyEncDecsGetter),
		filestore.WithEncoder(encdec.JSONEncoderDecoder{}))
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	settingStore.store = store
	settingStore.defaultData = spec.DefaultSettingsData
	return nil
}

func (s *SettingStore) KeyEncDecsGetter(data map[string]any) map[string]encdec.EncoderDecoder {
	keyEncDecs := make(map[string]encdec.EncoderDecoder)
	aiSettings, ok := data["aiSettings"].(map[string]any)
	if !ok {
		return nil
	}
	for providerName := range aiSettings {
		k := fmt.Sprintf("aiSettings.%s.apiKey", providerName)
		keyEncDecs[k] = encdec.EncryptedStringValueEncoderDecoder{}
	}

	return keyEncDecs
}

func (s *SettingStore) GetAllSettings(
	ctx context.Context,
	req *spec.GetAllSettingsRequest,
) (*spec.GetAllSettingsResponse, error) {
	forceFetch := false
	if req != nil {
		forceFetch = req.ForceFetch
	}
	data, err := s.store.GetAll(forceFetch)
	if err != nil {
		return nil, err
	}
	var settings spec.SettingsSchema
	if err := encdec.MapToStructWithJSONTags(data, &settings); err != nil {
		return nil, err
	}

	return &spec.GetAllSettingsResponse{Body: &settings}, nil
}

func (s *SettingStore) SetSetting(
	ctx context.Context,
	req *spec.SetSettingRequest,
) (*spec.SetSettingResponse, error) {
	dotSeparatedKey := req.Key
	keys := strings.Split(dotSeparatedKey, ".")

	// 1. Retrieve the current data from the store
	currentData, err := s.store.GetAll(false)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve current data: %w", err)
	}

	// 2. Make a copy of the current data
	mData := filestore.DeepCopyValue(currentData)

	modifiedData, ok := mData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to copy %q data", dotSeparatedKey)
	}

	// 3. Attempt to set the new value in the cloned map
	if err := filestore.SetValueAtPath(modifiedData, keys, req.Body.Value); err != nil {
		return nil, fmt.Errorf("failed to set value at key %q in map: %w", dotSeparatedKey, err)
	}

	// 4. Validate by converting the entire map to the SettingsSchema struct.
	//    This ensures the new data is compatible with our schema.
	var validatedSettings spec.SettingsSchema
	if err := encdec.MapToStructWithJSONTags(modifiedData, &validatedSettings); err != nil {
		return nil, fmt.Errorf("failed schema validation for key %q: %w", dotSeparatedKey, err)
	}

	// 5. If no error, persist the key/value to the underlying store.
	if err := s.store.SetKey(dotSeparatedKey, req.Body.Value); err != nil {
		return nil, fmt.Errorf("failed to persist key %q: %w", dotSeparatedKey, err)
	}

	return &spec.SetSettingResponse{}, nil
}

func (s *SettingStore) SetAppSettings(
	ctx context.Context,
	req *spec.SetAppSettingsRequest,
) (*spec.SetAppSettingsResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}
	_, _, _, err := s.getProviderData(req.Body.DefaultProvider, false)
	if err != nil {
		return nil, err
	}

	if err := s.store.SetKey("app.defaultProvider", string(req.Body.DefaultProvider)); err != nil {
		return nil, fmt.Errorf("failed to set app settings: %w", err)
	}
	return &spec.SetAppSettingsResponse{}, nil
}

func (s *SettingStore) AddAISetting(
	ctx context.Context,
	req *spec.AddAISettingRequest,
) (*spec.AddAISettingResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}

	if strings.Contains(string(req.ProviderName), ".") ||
		strings.Contains(string(req.ProviderName), " ") {
		return nil, errors.New("providername cannot have dot or space")
	}
	// Pull current data
	currentData, err := s.store.GetAll(false)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving current data: %w", err)
	}

	aiSettings, ok := currentData["aiSettings"].(map[string]any)
	if !ok {
		return nil, errors.New("aiSettings is missing or not a map")
	}

	// If it already exists, return error
	if _, exists := aiSettings[string(req.ProviderName)]; exists {
		return nil, fmt.Errorf("provider %q already exists", req.ProviderName)
	}

	// Just set
	dotKey := fmt.Sprintf("aiSettings.%s", req.ProviderName)
	val, err := encdec.StructWithJSONTagsToMap(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to add provider %q: %w", req.ProviderName, err)
	}
	if err := s.store.SetKey(dotKey, val); err != nil {
		return nil, fmt.Errorf("failed to add provider %q: %w", req.ProviderName, err)
	}

	return &spec.AddAISettingResponse{}, nil
}

func (s *SettingStore) DeleteAISetting(
	ctx context.Context,
	req *spec.DeleteAISettingRequest,
) (*spec.DeleteAISettingResponse, error) {
	providerName := req.ProviderName

	// Pull current data
	currentData, err := s.store.GetAll(false)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving current data: %w", err)
	}

	aiSettings, ok := currentData["aiSettings"].(map[string]any)
	if !ok {
		return nil, errors.New("aiSettings missing or not a map")
	}

	if _, exists := aiSettings[string(providerName)]; !exists {
		return nil, fmt.Errorf("provider %q does not exist", providerName)
	}

	if err := s.store.DeleteKey(fmt.Sprintf("aiSettings.%s", providerName)); err != nil {
		return nil, fmt.Errorf("failed to delete provider %q: %w", providerName, err)
	}

	return &spec.DeleteAISettingResponse{}, nil
}

func (s *SettingStore) getProviderData(
	providerName aiproviderSpec.ProviderName,
	forceFetch bool,
) (currentData, aiSettings, providerData map[string]any, err error) {
	// 1) GetAll with no forced fetch (or use the boolean if you sometimes need forced)
	currentData, err = s.store.GetAll(forceFetch)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to retrieve current data: %w", err)
	}

	// 2) aiSettings must be a map
	aiSettingsRaw, ok := currentData["aiSettings"]
	if !ok {
		return nil, nil, nil, errors.New("aiSettings missing from store")
	}
	aiSettings, ok = aiSettingsRaw.(map[string]any)
	if !ok {
		return nil, nil, nil, errors.New("aiSettings is not a map[string]any")
	}

	// 3) Check if provider exists
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
	req *spec.SetAISettingAPIKeyRequest,
) (*spec.SetAISettingAPIKeyResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}

	// If APIKey is empty, do nothing
	if req.Body.APIKey == "" {
		return &spec.SetAISettingAPIKeyResponse{}, nil
	}

	// Check provider exists
	_, _, _, err := s.getProviderData(req.ProviderName, false)
	if err != nil {
		return nil, err
	}

	// Construct dot-key and set
	dotKey := fmt.Sprintf("aiSettings.%s.apiKey", req.ProviderName)
	if err := s.store.SetKey(dotKey, req.Body.APIKey); err != nil {
		return nil, fmt.Errorf("failed to set AI API key: %w", err)
	}
	return &spec.SetAISettingAPIKeyResponse{}, nil
}

func (s *SettingStore) SetAISettingAttrs(
	ctx context.Context,
	req *spec.SetAISettingAttrsRequest,
) (*spec.SetAISettingAttrsResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}

	// Make sure provider is in store
	_, _, _, err := s.getProviderData(req.ProviderName, false)
	if err != nil {
		return nil, err
	}

	// For each non-nil field, set via dot key if it’s not empty (for strings)
	if req.Body.IsEnabled != nil {
		dotKey := fmt.Sprintf("aiSettings.%s.isEnabled", req.ProviderName)
		if err := s.store.SetKey(dotKey, *req.Body.IsEnabled); err != nil {
			return nil, fmt.Errorf("failed updating isEnabled: %w", err)
		}
	}
	if req.Body.Origin != nil && *req.Body.Origin != "" {
		dotKey := fmt.Sprintf("aiSettings.%s.origin", req.ProviderName)
		if err := s.store.SetKey(dotKey, *req.Body.Origin); err != nil {
			return nil, fmt.Errorf("failed updating origin: %w", err)
		}
	}
	if req.Body.ChatCompletionPathPrefix != nil && *req.Body.ChatCompletionPathPrefix != "" {
		dotKey := fmt.Sprintf("aiSettings.%s.chatCompletionPathPrefix", req.ProviderName)
		if err := s.store.SetKey(dotKey, *req.Body.ChatCompletionPathPrefix); err != nil {
			return nil, fmt.Errorf("failed updating chatCompletionPathPrefix: %w", err)
		}
	}
	if req.Body.DefaultModel != nil && *req.Body.DefaultModel != "" {
		dotKey := fmt.Sprintf("aiSettings.%s.defaultModel", req.ProviderName)
		if err := s.store.SetKey(dotKey, string(*req.Body.DefaultModel)); err != nil {
			return nil, fmt.Errorf("failed updating defaultModel: %w", err)
		}
	}

	return &spec.SetAISettingAttrsResponse{}, nil
}

func (s *SettingStore) AddModelSetting(
	ctx context.Context,
	req *spec.AddModelSettingRequest,
) (*spec.AddModelSettingResponse, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or request body cannot be nil")
	}

	// Confirm provider existence
	_, _, _, err := s.getProviderData(req.ProviderName, false)
	if err != nil {
		return nil, err
	}
	if strings.Contains(string(req.ModelName), ".") ||
		strings.Contains(string(req.ModelName), " ") {
		return nil, errors.New("modelname cannot have dot or space")
	}

	// Overwrite or create the model
	dotKey := fmt.Sprintf("aiSettings.%s.modelSettings.%s", req.ProviderName, req.ModelName)
	val, err := encdec.StructWithJSONTagsToMap(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed setting model %q for provider %q: %w",
			req.ModelName, req.ProviderName, err)
	}
	if err := s.store.SetKey(dotKey, val); err != nil {
		return nil, fmt.Errorf("failed setting model %q for provider %q: %w",
			req.ModelName, req.ProviderName, err)
	}
	return &spec.AddModelSettingResponse{}, nil
}

func (s *SettingStore) DeleteModelSetting(
	ctx context.Context,
	req *spec.DeleteModelSettingRequest,
) (*spec.DeleteModelSettingResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Confirm provider existence
	_, _, _, err := s.getProviderData(req.ProviderName, false)
	if err != nil {
		return nil, err
	}

	// Delete the model
	dotKey := fmt.Sprintf("aiSettings.%s.modelSettings.%s", req.ProviderName, req.ModelName)
	if err := s.store.DeleteKey(dotKey); err != nil {
		return nil, fmt.Errorf("failed deleting model %q for provider %q: %w",
			req.ModelName, req.ProviderName, err)
	}
	return &spec.DeleteModelSettingResponse{}, nil
}
