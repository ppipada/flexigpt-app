package settingstore_test

import (
	"os"
	"strings"
	"testing"

	"github.com/flexigpt/flexiui/pkg/settingstore"
)

func TestSettingStore_GetAllSettings(t *testing.T) {
	// Create a temporary file for testing
	filename := "test_settings.json"
	defer os.Remove(filename)
	store := &settingstore.SettingStore{}
	err := settingstore.InitSettingStore(store, filename)
	if err != nil {
		t.Fatalf("Failed to create settings store: %v", err)
	}

	// Set additional data
	err = store.SetSetting("aiSettings.openai.additionalSettings", map[string]interface{}{"internal": 1, "new": [][]string{
		{"1", "2"},
		{"1"},
	}})
	if err != nil {
		t.Fatalf("Failed to set additional setting: %v", err)
	}

	// Set sensitive data
	err = store.SetSetting("aiSettings.openai.apiKey", "sensitiveApiKey")
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
				return os.WriteFile(filename, []byte(`{"invalid": "schema"}`), 0644)
			},
			expectedError: "unknown field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setupFunc(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			got, err := store.GetAllSettings(true)
			if err != nil && tt.expectedError == "" {
				t.Errorf("GetAllSettings() unexpected error = %v", err)
				return
			}
			if err == nil && tt.expectedError != "" {
				t.Errorf("GetAllSettings() expected error = %v, got nil", tt.expectedError)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("GetAllSettings() error = %v, expected error containing %v", err, tt.expectedError)
			}

			if tt.expectedError == "" && got == nil {
				t.Errorf("GetAllSettings() got = %v, want non-nil", got)
			}

			if tt.expectedError == "" {
				// Verify additional settings
				additionalSettings := got.AISettings["openai"].AdditionalSettings
				if customVal, ok := additionalSettings["new"]; !ok || customVal == nil {
					t.Errorf("GetAllSettings() additional setting mismatch, expected 'new' setting to be present")
				}

			}
		})
	}
}

func TestSettingStore_SetSetting(t *testing.T) {
	store := &settingstore.SettingStore{}
	err := settingstore.InitSettingStore(store, "test_settings.json")
	if err != nil {
		t.Fatalf("Failed to create settings store: %v", err)
	}

	tests := []struct {
		name          string
		key           string
		value         interface{}
		wantErr       bool
		expectedError string
	}{
		{"SetSetting_ValidKey", "app.defaultProvider", "OPENAI", false, ""},
		{"SetSetting_InvalidKey", "app.invalidKey", "OPENAI", true, "invalid key: app.invalidKey"},
		{"SetSetting_TypeMismatch", "app.defaultProvider", 123, true, "type mismatch for key \"app.defaultProvider\": expected string, got int"},
		{"SetSetting_SensitiveKey", "aiSettings.openai.apiKey", "newApiKey", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.SetSetting(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetSetting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.expectedError {
				t.Errorf("SetSetting() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}
