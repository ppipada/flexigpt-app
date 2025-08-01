package store

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"

	"github.com/ppipada/flexigpt-app/pkg/setting/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
)

type SettingStore struct {
	store      *filestore.MapFileStore
	encEncrypt encdec.EncoderDecoder
}

func NewSettingStore(baseDir string) (*SettingStore, error) {
	st := &SettingStore{
		encEncrypt: encdec.EncryptedStringValueEncoderDecoder{},
	}

	defaultMap, err := encdec.StructWithJSONTagsToMap(DefaultSettingsData)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal default settings: %w", err)
	}

	file := filepath.Join(baseDir, spec.SettingsFile)
	fs, err := filestore.NewMapFileStore(
		file,
		defaultMap,
		filestore.WithCreateIfNotExists(true),
		filestore.WithAutoFlush(true),
		filestore.WithValueEncDecGetter(st.valueEncDecGetter),
		filestore.WithEncoderDecoder(encdec.JSONEncoderDecoder{}),
	)
	if err != nil {
		return nil, fmt.Errorf("filestore init failed: %w", err)
	}

	st.store = fs
	slog.Info("settings store ready", "file", file)
	return st, nil
}

// SetAppTheme validates and persists a new theme.
func (s *SettingStore) SetAppTheme(
	_ context.Context,
	req *spec.SetAppThemeRequest,
) (*spec.SetAppThemeResponse, error) {
	if req == nil || req.Body == nil {
		return nil, spec.ErrInvalidArgument
	}

	theme := &spec.AppTheme{Type: req.Body.Type, Name: req.Body.Name}
	if err := validateTheme(theme); err != nil {
		return nil, err
	}

	val, _ := encdec.StructWithJSONTagsToMap(theme)
	if err := s.store.SetKey([]string{"appTheme"}, val); err != nil {
		return nil, err
	}

	slog.Info("appTheme updated", "type", theme.Type, "name", theme.Name)
	return &spec.SetAppThemeResponse{}, nil
}

// SetAuthKey inserts or updates one auth-key.
func (s *SettingStore) SetAuthKey(
	_ context.Context,
	req *spec.SetAuthKeyRequest,
) (*spec.SetAuthKeyResponse, error) {
	if req == nil || req.Body == nil || req.Type == "" || req.KeyName == "" {
		return nil, spec.ErrInvalidArgument
	}

	nonEmptySecret := req.Body.Secret != ""

	// Build AuthKey record.
	newAk := spec.AuthKey{
		Secret:   req.Body.Secret,
		SHA256:   encdec.ComputeSHA(req.Body.Secret),
		NonEmpty: nonEmptySecret,
	}

	// Persist secret (encrypted) then sha (plain).
	secretPath := []string{"authKeys", string(req.Type), string(req.KeyName), "secret"}
	if err := s.store.SetKey(secretPath, newAk.Secret); err != nil {
		return nil, err
	}
	shaPath := []string{"authKeys", string(req.Type), string(req.KeyName), "sha256"}
	if err := s.store.SetKey(shaPath, newAk.SHA256); err != nil {
		return nil, err
	}
	nonEmptyPath := []string{
		"authKeys",
		string(req.Type),
		string(req.KeyName),
		"nonEmpty",
	}
	if err := s.store.SetKey(nonEmptyPath, newAk.NonEmpty); err != nil {
		return nil, err
	}

	slog.Info("authKey set",
		"type", req.Type, "keyName", req.KeyName,
		"builtIn", isBuiltInKey(req.Type, req.KeyName))
	return &spec.SetAuthKeyResponse{}, nil
}

// DeleteAuthKey removes a key unless it is marked built-in.
func (s *SettingStore) DeleteAuthKey(
	_ context.Context,
	req *spec.DeleteAuthKeyRequest,
) (*spec.DeleteAuthKeyResponse, error) {
	if req == nil || req.Type == "" || req.KeyName == "" {
		return nil, spec.ErrInvalidArgument
	}
	if isBuiltInKey(req.Type, req.KeyName) {
		return nil, spec.ErrBuiltInAuthKeyReadOnly
	}

	// Delete the key map entirely (secret + sha).
	keyPath := []string{"authKeys", string(req.Type), string(req.KeyName)}
	if err := s.store.DeleteKey(keyPath); err != nil {
		return nil, err
	}

	// If the type map is now empty, delete it too.
	raw, err := s.store.GetAll(false)
	if err != nil {
		return nil, err
	}

	if akRaw, ok := raw["authKeys"].(map[string]any); ok {
		if typRaw, ok := akRaw[string(req.Type)].(map[string]any); ok && len(typRaw) == 0 {
			_ = s.store.DeleteKey([]string{"authKeys", string(req.Type)})
		}
	}

	slog.Info("authKey deleted", "type", req.Type, "keyName", req.KeyName)
	return &spec.DeleteAuthKeyResponse{}, nil
}

// GetAuthKey returns the decrypted secret for one key.
func (s *SettingStore) GetAuthKey(
	_ context.Context,
	req *spec.GetAuthKeyRequest,
) (*spec.GetAuthKeyResponse, error) {
	if req == nil || req.Type == "" || req.KeyName == "" {
		return nil, spec.ErrInvalidArgument
	}

	raw, err := s.store.GetAll(false)
	if err != nil {
		return nil, err
	}

	var schema spec.SettingsSchema
	if err := encdec.MapToStructWithJSONTags(raw, &schema); err != nil {
		return nil, err
	}

	typData, ok := schema.AuthKeys[req.Type]
	if !ok {
		return nil, spec.ErrAuthKeyNotFound
	}
	ak, ok := typData[req.KeyName]
	if !ok {
		return nil, spec.ErrAuthKeyNotFound
	}

	return &spec.GetAuthKeyResponse{
		Body: &spec.GetAuthKeyResponseBody{
			Secret:   ak.Secret,
			SHA256:   ak.SHA256,
			NonEmpty: ak.NonEmpty,
		},
	}, nil
}

// GetSettings returns the current settings without secrets.
func (s *SettingStore) GetSettings(
	_ context.Context,
	req *spec.GetSettingsRequest,
) (*spec.GetSettingsResponse, error) {
	force := false
	if req != nil {
		force = req.ForceFetch
	}

	raw, err := s.store.GetAll(force)
	if err != nil {
		return nil, err
	}

	var schema spec.SettingsSchema
	if err := encdec.MapToStructWithJSONTags(raw, &schema); err != nil {
		return nil, err
	}

	// Convert to DTO (secrets stripped).
	out := spec.GetSettingsResponse{
		Body: &spec.GetSettingsResponseBody{
			AppTheme: schema.AppTheme,
			AuthKeys: []spec.AuthKeyMeta{},
		},
	}
	for t, m := range schema.AuthKeys {
		for n, ak := range m {
			out.Body.AuthKeys = append(out.Body.AuthKeys, spec.AuthKeyMeta{
				Type:     t,
				KeyName:  n,
				SHA256:   ak.SHA256,
				NonEmpty: ak.NonEmpty,
			})
		}
	}

	// Sort by Type, then by KeyName.
	sort.Slice(out.Body.AuthKeys, func(i, j int) bool {
		if out.Body.AuthKeys[i].Type == out.Body.AuthKeys[j].Type {
			return out.Body.AuthKeys[i].KeyName < out.Body.AuthKeys[j].KeyName
		}
		return out.Body.AuthKeys[i].Type < out.Body.AuthKeys[j].Type
	})

	return &out, nil
}

// valueEncDecGetter returns the encoder/decoder to encrypt secrets.
func (s *SettingStore) valueEncDecGetter(path []string) encdec.EncoderDecoder {
	// AuthKeys / <type> / <keyName> / secret.
	if len(path) == 4 && path[0] == "authKeys" && path[3] == "secret" {
		return s.encEncrypt
	}
	return nil
}
