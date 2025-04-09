package settingstore_test

import (
	"os"
	"strings"
	"testing"

	"github.com/flexigpt/flexiui/pkg/settingstore"
	"github.com/flexigpt/flexiui/pkg/settingstore/spec"
)

func initTestFile(filePath string) error {
	// Check if the file exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, so remove it
		err := os.Remove(filePath)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return err
	}
	return nil
}

func TestSettingStore_GetAllSettings(t *testing.T) {
	// Create a temporary file for testing
	filename := "test_settings.json"
	err := initTestFile(filename)
	if err != nil {
		t.Fatalf("Failed to init test file: %v", err)
	}
	defer os.Remove(filename)
	store := &settingstore.SettingStore{}
	err = settingstore.InitSettingStore(store, filename)
	if err != nil {
		t.Fatalf("Failed to create settings store: %v", err)
	}
	ctx := t.Context()
	request := &spec.SetSettingRequest{
		Key: "aiSettings.openai.defaultModel",
		Body: &spec.SetSettingRequestBody{
			Value: "gpt-4o",
		},
	}
	// Set additional data
	_, err = store.SetSetting(
		ctx,
		request,
	)
	if err != nil {
		t.Fatalf("Failed to set additional setting: %v", err)
	}
	// Set sensitive data
	_, err = store.SetSetting(ctx, &spec.SetSettingRequest{
		Key: "aiSettings.openai.apiKey",
		Body: &spec.SetSettingRequestBody{
			Value: "sensitiveApiKey",
		},
	})
	if err != nil {
		t.Fatalf("Failed to set sensitive setting: %v", err)
	}

	tests := []struct {
		name          string
		setupFunc     func() error
		expectedError string
	}{
		{
			name: "GetAllSettings_HappyPath",
			setupFunc: func() error {
				return nil
			},
			expectedError: "",
		},
		{
			name: "GetAllSettings_InvalidSchema",
			setupFunc: func() error {
				// Manually create an invalid schema file
				return os.WriteFile(filename, []byte(`{"invalid": "schema"}`), 0o600)
			},
			expectedError: "unknown field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setupFunc(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			got, err := store.GetAllSettings(
				t.Context(),
				&spec.GetAllSettingsRequest{ForceFetch: true},
			)
			if err != nil && tt.expectedError == "" {
				t.Errorf("GetAllSettings() unexpected error = %v", err)
				return
			}
			if err == nil && tt.expectedError != "" {
				t.Errorf("GetAllSettings() expected error = %v, got nil", tt.expectedError)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf(
					"GetAllSettings() error = %v, expected error containing %v",
					err,
					tt.expectedError,
				)
			}

			if tt.expectedError == "" && got == nil {
				t.Errorf("GetAllSettings() got = %v, want non-nil", got)
			}

			if tt.expectedError == "" {
				// Verify additional settings
				defaultModel := got.Body.AISettings["openai"].DefaultModel
				if defaultModel != "gpt-4o" {
					t.Errorf(
						"GetAllSettings() defaultModel mismatch, expected 'gpt-4o' setting to be present",
					)
				}

			}
		})
	}
}

func TestSettingStore_SetSetting(t *testing.T) {
	filename := "test_settings.json"
	err := initTestFile(filename)
	if err != nil {
		t.Fatalf("Failed to init test file: %v", err)
	}
	defer os.Remove(filename)

	store := &settingstore.SettingStore{}
	err = settingstore.InitSettingStore(store, filename)
	if err != nil {
		t.Fatalf("Failed to create settings store: %v", err)
	}

	tests := []struct {
		name          string
		key           string
		value         any
		wantErr       bool
		expectedError string
	}{
		{"SetSetting_ValidKey", "app.defaultProvider", "ProviderNameOpenAI", false, ""},
		{"SetSetting_KeyNotInDefaultData", "aiSettings.openai2.origin", "O2Origin", false, ""},
		{
			"SetSetting_InvalidKey",
			"app.invalidKey",
			"ProviderNameOpenAI",
			true,
			"unknown field \"invalidKey\"",
		},
		{
			"SetSetting_TypeMismatch",
			"app.defaultProvider",
			123,
			true,
			"cannot unmarshal number",
		},
		{"SetSetting_SensitiveKey", "aiSettings.openai.apiKey", "newApiKey", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set sensitive data
			_, err = store.SetSetting(t.Context(), &spec.SetSettingRequest{
				Key: tt.key,
				Body: &spec.SetSettingRequestBody{
					Value: tt.value,
				},
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("SetSetting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("SetSetting() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}
