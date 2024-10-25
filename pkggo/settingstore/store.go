package settingstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/flexigpt/flexiui/pkggo/settingstore/spec"
	"github.com/flexigpt/flexiui/pkggo/simplemapdb/encdec"
	"github.com/flexigpt/flexiui/pkggo/simplemapdb/filestore"
)

type SettingStore struct {
	store       *filestore.MapFileStore
	defaultData spec.SettingsSchema
}

func InitSettingStore(settingStore *SettingStore, filename string) error {
	keyEncDecs := make(map[string]encdec.EncoderDecoder)
	for _, key := range spec.SensitiveKeys {
		keyEncDecs[key] = encdec.EncryptedStringValueEncoderDecoder{}
	}
	settingsMap, err := encdec.StructWithJSONTagsToMap(spec.DefaultSettingsData)
	if err != nil {
		return fmt.Errorf("Could not get map of settings data")
	}
	store, err := filestore.NewMapFileStore(
		filename,
		settingsMap,
		filestore.WithCreateIfNotExists(true),
		filestore.WithAutoFlush(true),
		filestore.WithKeyEncoders(keyEncDecs),
		filestore.WithEncoder(encdec.JSONEncoderDecoder{}))
	if err != nil {
		return fmt.Errorf("failed to create store: %v", err)
	}

	settingStore.store = store
	settingStore.defaultData = spec.DefaultSettingsData
	return nil
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
	// Assuming a function to map data to settings
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
	defaultDataMap, err := encdec.StructWithJSONTagsToMap(s.defaultData)
	if err != nil {
		return nil, err
	}

	var currentSchema interface{}
	currentSchema = defaultDataMap
	for _, key := range keys {
		switch currentTypedSchema := currentSchema.(type) {
		case map[string]interface{}:
			if _, ok := currentTypedSchema[key]; !ok {
				return nil, fmt.Errorf("invalid key: %s", dotSeparatedKey)
			}
			currentSchema = currentTypedSchema[key]
		default:
			return nil, fmt.Errorf("invalid key: %s", dotSeparatedKey)
		}
	}

	expectedType := fmt.Sprintf("%T", currentSchema)
	valueType := fmt.Sprintf("%T", req.Body.Value)

	if expectedType != valueType {
		return nil, fmt.Errorf(
			"type mismatch for key \"%s\": expected %s, got %s",
			dotSeparatedKey,
			expectedType,
			valueType,
		)
	}

	err = s.store.SetKey(dotSeparatedKey, req.Body.Value)

	return &spec.SetSettingResponse{}, err
}
