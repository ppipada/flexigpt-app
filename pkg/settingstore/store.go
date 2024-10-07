package settingstore

import (
	"fmt"
	"strings"

	"github.com/flexigpt/flexiui/pkg/settingstore/spec"
	"github.com/flexigpt/flexiui/pkg/simplemapdb/encdec"
	"github.com/flexigpt/flexiui/pkg/simplemapdb/filestore"
)

type SettingsStore struct {
	store       *filestore.MapFileStore
	defaultData spec.SettingsSchema
}

func NewSettingStore(filename string) (*SettingsStore, error) {
	keyEncDecs := make(map[string]encdec.EncoderDecoder)
	for _, key := range spec.SensitiveKeys {
		keyEncDecs[key] = encdec.EncryptedStringValueEncoderDecoder{}
	}
	settingsMap, err := encdec.StructWithJSONTagsToMap(spec.DefaultSettingsData)
	if err != nil {
		return nil, fmt.Errorf("Could not get map of settings data")
	}
	store, err := filestore.NewMapFileStore(
		filename,
		settingsMap,
		filestore.WithCreateIfNotExists(true),
		filestore.WithAutoFlush(true),
		filestore.WithKeyEncoders(keyEncDecs),
		filestore.WithEncoder(encdec.JSONEncoderDecoder{}))
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %v", err)
	}

	return &SettingsStore{store: store, defaultData: spec.DefaultSettingsData}, nil
}

func (s *SettingsStore) GetAllSettings(forceFetch bool) (*spec.SettingsSchema, error) {
	data, err := s.store.GetAll(forceFetch)
	if err != nil {
		return nil, err
	}
	var settings spec.SettingsSchema
	// Assuming a function to map data to settings
	if err := encdec.MapToStructWithJSONTags(data, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *SettingsStore) SetSetting(dotSeparatedKey string, value interface{}) error {
	keys := strings.Split(dotSeparatedKey, ".")
	defaultDataMap, err := encdec.StructWithJSONTagsToMap(s.defaultData)
	if err != nil {
		return err
	}

	var currentSchema interface{}
	currentSchema = defaultDataMap
	for _, key := range keys {
		switch currentTypedSchema := currentSchema.(type) {
		case map[string]interface{}:
			if _, ok := currentTypedSchema[key]; !ok {
				return fmt.Errorf("invalid key: %s", dotSeparatedKey)
			}
			currentSchema = currentTypedSchema[key]
		default:
			return fmt.Errorf("invalid key: %s", dotSeparatedKey)
		}
	}

	expectedType := fmt.Sprintf("%T", currentSchema)
	valueType := fmt.Sprintf("%T", value)

	if expectedType != valueType {
		return fmt.Errorf("type mismatch for key \"%s\": expected %s, got %s", dotSeparatedKey, expectedType, valueType)
	}

	return s.store.SetKey(dotSeparatedKey, value)
}
