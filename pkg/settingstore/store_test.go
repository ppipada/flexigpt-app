package settingstore

import (
	"os"
	"strings"
	"testing"

	"github.com/ppipada/flexigpt-app/pkg/model/spec"
)

// initTestFile removes any existing file with the given path,
// guaranteeing a clean file for each test run.
func initTestFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		return os.Remove(filePath)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// TestSettingStore_GetAllSettings exercises the GetAllSettings method.
func TestSettingStore_GetAllSettings(t *testing.T) {
	filename := "test_settings.json"
	if err := initTestFile(filename); err != nil {
		t.Fatalf("Failed to init test file: %v", err)
	}
	defer os.Remove(filename)

	store := &SettingStore{}
	err := InitSettingStore(store, filename)
	if err != nil {
		t.Fatalf("Failed to create settings store: %v", err)
	}

	// Pre-populate some settings by adding a provider, setting an API key, etc.
	setupCtx := t.Context()

	// Add a new provider "openai2".
	_, err = store.AddAISetting(
		setupCtx,
		&AddAISettingRequest{
			ProviderName: spec.ProviderName("openai2"),
			Body: &AISetting{
				IsEnabled: true,
				Origin:    "https://test-origin",
			},
		},
	)
	if err != nil {
		t.Fatalf("Failed to add AI Setting: %v", err)
	}

	// Set an API key.
	_, err = store.SetAISettingAPIKey(
		setupCtx,
		&SetAISettingAPIKeyRequest{
			ProviderName: spec.ProviderName("openai2"),
			Body: &SetAISettingAPIKeyRequestBody{
				APIKey: "sensitiveApiKey",
			},
		},
	)
	if err != nil {
		t.Fatalf("Failed to set API key: %v", err)
	}

	testCases := []struct {
		name          string
		setupFunc     func() error
		forceFetch    bool
		expectedError string
	}{
		{
			name: "GetAllSettings_HappyPath",
			setupFunc: func() error {
				// Do nothing, file already has valid data.
				return nil
			},
			// Force re-fetch from file.
			forceFetch:    true,
			expectedError: "",
		},
		{
			name: "GetAllSettings_InvalidSchema",
			setupFunc: func() error {
				// Overwrite file with invalid schema.
				return os.WriteFile(filename, []byte(`{"invalid":"schema"}`), 0o600)
			},
			forceFetch:    true,
			expectedError: "unknown field",
		},
		{
			name: "GetAllSettings_EmptyFile",
			setupFunc: func() error {
				// Write an empty file.
				return os.WriteFile(filename, []byte(``), 0o600)
			},
			forceFetch:    true,
			expectedError: "EOF",
		},
	}

	for _, tc := range testCases {
		// Capture range variable.
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.setupFunc(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			got, err := store.GetAllSettings(
				t.Context(),
				&GetAllSettingsRequest{ForceFetch: tc.forceFetch},
			)
			if err != nil && tc.expectedError == "" {
				t.Errorf("unexpected error = %v", err)
				return
			}
			if err == nil && tc.expectedError != "" {
				t.Errorf("expected error containing %q, got nil", tc.expectedError)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("error = %v, expected substring %q", err, tc.expectedError)
			}

			if tc.expectedError == "" && got != nil {
				if _, ok := got.Body.AISettings["openai2"]; !ok {
					t.Errorf("openai2 provider should be present in AISettings")
				}
			}
		})
	}
}

// TestSettingStore_SetAppSettings exercises the SetAppSettings method.
func TestSettingStore_SetAppSettings(t *testing.T) {
	filename := "test_appsettings.json"
	if err := initTestFile(filename); err != nil {
		t.Fatalf("Failed to init test file: %v", err)
	}
	defer os.Remove(filename)

	store := &SettingStore{}
	err := InitSettingStore(store, filename)
	if err != nil {
		t.Fatalf("Failed to create settings store: %v", err)
	}

	// Must add a provider first to set it as default.
	ctx := t.Context()
	_, err = store.AddAISetting(ctx, &AddAISettingRequest{
		ProviderName: spec.ProviderName("openai2"),
		Body: &AISetting{
			IsEnabled: true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to add AI provider: %v", err)
	}

	testCases := []struct {
		name          string
		req           *SetAppSettingsRequest
		wantErr       bool
		expectedError string
	}{
		{
			name: "ValidAppSettings",
			req: &SetAppSettingsRequest{
				Body: &AppSettings{
					DefaultProvider: "openai2",
				},
			},
			wantErr:       false,
			expectedError: "",
		},
		{
			name:          "NilRequest",
			req:           nil,
			wantErr:       true,
			expectedError: "request or request body cannot be nil",
		},
		{
			name:          "NilRequestBody",
			req:           &SetAppSettingsRequest{},
			wantErr:       true,
			expectedError: "request or request body cannot be nil",
		},
		{
			name: "UnknownProvider",
			req: &SetAppSettingsRequest{
				Body: &AppSettings{
					DefaultProvider: "unknown-provider",
				},
			},
			wantErr:       true,
			expectedError: "does not exist in aiSettings",
		},
		{
			name: "EmptyProviderName",
			req: &SetAppSettingsRequest{
				Body: &AppSettings{
					DefaultProvider: "",
				},
			},
			wantErr:       true,
			expectedError: "provider \"\" does not exist in aiSettings",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := store.SetAppSettings(t.Context(), tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("SetAppSettings() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("SetAppSettings() error = %v, want substring %q", err, tc.expectedError)
			}
		})
	}
}

// TestSettingStore_AddAISetting exercises the AddAISetting method.
func TestSettingStore_AddAISetting(t *testing.T) {
	filename := "test_add_ai_settings.json"
	if err := initTestFile(filename); err != nil {
		t.Fatalf("Failed to init test file: %v", err)
	}
	defer os.Remove(filename)

	store := &SettingStore{}
	err := InitSettingStore(store, filename)
	if err != nil {
		t.Fatalf("Failed to create settings store: %v", err)
	}
	ctx := t.Context()

	// First, add a provider "openai2".
	_, err = store.AddAISetting(ctx, &AddAISettingRequest{
		ProviderName: spec.ProviderName("openai2"),
		Body: &AISetting{
			IsEnabled: true,
		},
	})
	if err != nil {
		t.Fatalf("Initial AddAISetting() for openai2 failed: %v", err)
	}

	testCases := []struct {
		name          string
		req           *AddAISettingRequest
		wantErr       bool
		expectedError string
	}{
		{
			name:    "NilRequest",
			req:     nil,
			wantErr: true, expectedError: "request or request body cannot be nil",
		},
		{
			name: "NilRequestBody",
			req: &AddAISettingRequest{
				ProviderName: spec.ProviderName("cohere"),
				Body:         nil,
			},
			wantErr:       true,
			expectedError: "request or request body cannot be nil",
		},
		{
			name: "DuplicateProvider",
			req: &AddAISettingRequest{
				ProviderName: spec.ProviderName("openai2"),
				Body:         &AISetting{},
			},
			wantErr:       true,
			expectedError: "already exists",
		},
		{
			name: "ValidNewProvider",
			req: &AddAISettingRequest{
				ProviderName: spec.ProviderName("azure"),
				Body: &AISetting{
					IsEnabled: false,
				},
			},
			wantErr:       false,
			expectedError: "",
		},
		{
			name: "ProviderNameWithSpecialChars",
			req: &AddAISettingRequest{
				ProviderName: spec.ProviderName("some@crazy#provider!"),
				Body: &AISetting{
					IsEnabled: true,
				},
			},
			wantErr:       false,
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := store.AddAISetting(t.Context(), tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("AddAISetting() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("AddAISetting() error = %v, want substring %q", err, tc.expectedError)
			}
		})
	}
}

// TestSettingStore_DeleteAISetting exercises the DeleteAISetting method.
func TestSettingStore_DeleteAISetting(t *testing.T) {
	filename := "test_delete_ai_settings.json"
	if err := initTestFile(filename); err != nil {
		t.Fatalf("Failed to init test file: %v", err)
	}
	defer os.Remove(filename)

	store := &SettingStore{}
	err := InitSettingStore(store, filename)
	if err != nil {
		t.Fatalf("Failed to create settings store: %v", err)
	}
	ctx := t.Context()

	// Add a provider "openai2".
	_, err = store.AddAISetting(ctx, &AddAISettingRequest{
		ProviderName: spec.ProviderName("openai2"),
		Body: &AISetting{
			IsEnabled: true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to add AI Setting for openai2: %v", err)
	}

	testCases := []struct {
		name          string
		req           *DeleteAISettingRequest
		wantErr       bool
		expectedError string
	}{
		{
			name:          "NilRequest",
			req:           nil,
			wantErr:       true,
			expectedError: "request cannot be nil",
		},
		{
			name: "DeleteNonExistingProvider",
			req: &DeleteAISettingRequest{
				ProviderName: spec.ProviderName("cohere"),
			},
			wantErr:       true,
			expectedError: "does not exist",
		},
		{
			name: "DeleteExistingProvider",
			req: &DeleteAISettingRequest{
				ProviderName: spec.ProviderName("openai2"),
			},
			wantErr:       false,
			expectedError: "",
		},
		{
			name: "DeleteEmptyProviderName",
			req: &DeleteAISettingRequest{
				ProviderName: spec.ProviderName(""),
			},
			wantErr:       true,
			expectedError: "provider \"\" does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := store.DeleteAISetting(t.Context(), tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteAISetting() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && tc.expectedError != "" &&
				!strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("DeleteAISetting() error = %v, want substring %q", err, tc.expectedError)
			}
		})
	}
}

// TestSettingStore_SetAISettingAPIKey exercises the SetAISettingAPIKey method.
func TestSettingStore_SetAISettingAPIKey(t *testing.T) {
	filename := "test_set_api_key.json"
	if err := initTestFile(filename); err != nil {
		t.Fatalf("Failed to init test file: %v", err)
	}
	defer os.Remove(filename)

	store := &SettingStore{}
	err := InitSettingStore(store, filename)
	if err != nil {
		t.Fatalf("Failed to create settings store: %v", err)
	}
	ctx := t.Context()

	// Add a provider "openai2".
	_, err = store.AddAISetting(ctx, &AddAISettingRequest{
		ProviderName: spec.ProviderName("openai2"),
		Body:         &AISetting{},
	})
	if err != nil {
		t.Fatalf("Failed to add AI Setting for openai2: %v", err)
	}

	testCases := []struct {
		name          string
		req           *SetAISettingAPIKeyRequest
		wantErr       bool
		expectedError string
	}{
		{
			name:          "NilRequest",
			req:           nil,
			wantErr:       true,
			expectedError: "request or request body cannot be nil",
		},
		{
			name: "NilBody",
			req: &SetAISettingAPIKeyRequest{
				ProviderName: spec.ProviderName("openai2"),
				Body:         nil,
			},
			wantErr:       true,
			expectedError: "request or request body cannot be nil",
		},
		{
			name: "EmptyAPIKeyNoOp",
			req: &SetAISettingAPIKeyRequest{
				ProviderName: spec.ProviderName("openai2"),
				Body: &SetAISettingAPIKeyRequestBody{
					APIKey: "",
				},
			},
			wantErr:       false,
			expectedError: "",
		},
		{
			name: "UnknownProvider",
			req: &SetAISettingAPIKeyRequest{
				ProviderName: spec.ProviderName("cohere"),
				Body: &SetAISettingAPIKeyRequestBody{
					APIKey: "someKey",
				},
			},
			wantErr:       true,
			expectedError: "does not exist in aiSettings",
		},
		{
			name: "SetValidAPIKey",
			req: &SetAISettingAPIKeyRequest{
				ProviderName: spec.ProviderName("openai2"),
				Body: &SetAISettingAPIKeyRequestBody{
					APIKey: "actualAPIKey",
				},
			},
			wantErr:       false,
			expectedError: "",
		},
		{
			name: "ExceedinglyLongAPIKey",
			req: &SetAISettingAPIKeyRequest{
				ProviderName: spec.ProviderName("openai2"),
				Body: &SetAISettingAPIKeyRequestBody{
					// 1KB of 'A'.
					APIKey: strings.Repeat("A", 1024),
				},
			},
			wantErr:       false,
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := store.SetAISettingAPIKey(t.Context(), tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("SetAISettingAPIKey() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf(
					"SetAISettingAPIKey() error = %v, want substring %q",
					err, tc.expectedError,
				)
			}
		})
	}
}

// TestSettingStore_SetAISettingAttrs exercises the SetAISettingAttrs method.
func TestSettingStore_SetAISettingAttrs(t *testing.T) {
	filename := "test_set_ai_attrs.json"
	if err := initTestFile(filename); err != nil {
		t.Fatalf("Failed to init test file: %v", err)
	}
	defer os.Remove(filename)

	store := &SettingStore{}
	err := InitSettingStore(store, filename)
	if err != nil {
		t.Fatalf("Failed to create settings store: %v", err)
	}
	ctx := t.Context()

	// Add one provider.
	_, err = store.AddAISetting(ctx, &AddAISettingRequest{
		ProviderName: spec.ProviderName("openai2"),
		Body:         &AISetting{},
	})
	if err != nil {
		t.Fatalf("Failed to add AI Setting for openai2: %v", err)
	}

	newBool := func(b bool) *bool { return &b }
	newString := func(s string) *string { return &s }

	testCases := []struct {
		name          string
		req           *SetAISettingAttrsRequest
		wantErr       bool
		expectedError string
	}{
		{
			name:          "NilRequest",
			req:           nil,
			wantErr:       true,
			expectedError: "request or request body cannot be nil",
		},
		{
			name: "NilRequestBody",
			req: &SetAISettingAttrsRequest{
				ProviderName: spec.ProviderName("openai2"),
				Body:         nil,
			},
			wantErr:       true,
			expectedError: "request or request body cannot be nil",
		},
		{
			name: "UnknownProvider",
			req: &SetAISettingAttrsRequest{
				ProviderName: spec.ProviderName("cohere"),
				Body: &SetAISettingAttrsRequestBody{
					IsEnabled: newBool(true),
				},
			},
			wantErr:       true,
			expectedError: "does not exist in aiSettings",
		},
		{
			name: "PartialAttrsUpdate",
			req: &SetAISettingAttrsRequest{
				ProviderName: spec.ProviderName("openai2"),
				Body: &SetAISettingAttrsRequestBody{
					IsEnabled: newBool(true),
					Origin:    newString("some-origin"),
				},
			},
			wantErr:       false,
			expectedError: "",
		},
		{
			name: "AllFieldsSet",
			req: &SetAISettingAttrsRequest{
				ProviderName: spec.ProviderName("openai2"),
				Body: &SetAISettingAttrsRequestBody{
					IsEnabled:                newBool(false),
					Origin:                   newString("test-origin"),
					ChatCompletionPathPrefix: newString("/v1/complete-chat"),
				},
			},
			wantErr:       false,
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := store.SetAISettingAttrs(t.Context(), tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("SetAISettingAttrs() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf(
					"SetAISettingAttrs() error = %v, expect substring %q",
					err,
					tc.expectedError,
				)
			}
		})
	}
}
