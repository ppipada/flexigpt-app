package store_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ppipada/flexigpt-app/pkg/model/spec"
	"github.com/ppipada/flexigpt-app/pkg/model/store"
)

func initTestFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		return os.Remove(filePath)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func newTestStore(t *testing.T, filename string) *store.ModelPresetStore {
	t.Helper()
	if err := initTestFile(filename); err != nil {
		t.Fatalf("Failed to init test file: %v", err)
	}
	s := &store.ModelPresetStore{}
	if err := store.InitModelPresetStore(s, filename); err != nil {
		t.Fatalf("InitModelPresetStore failed: %v", err)
	}
	return s
}

func TestModelPresetStore_GetAllModelPresets(t *testing.T) {
	filename := "test_modelpresets_getall.json"
	defer os.Remove(filename)
	s := newTestStore(t, filename)
	ctx := t.Context()

	// Add a provider for happy path.
	provider := spec.ProviderName("openai2")
	modelName := spec.ModelName("gpt-4")
	preset := spec.ModelPreset{
		Name:         modelName,
		DisplayName:  "GPT-4",
		ShortCommand: "g4",
		IsEnabled:    true,
	}
	body := spec.ProviderModelPresets{
		modelName: preset,
	}
	_, err := s.CreateModelPresets(ctx, &spec.CreateModelPresetsRequest{
		ProviderName: provider,
		Body:         &body,
	})
	if err != nil {
		t.Fatalf("Failed to add provider: %v", err)
	}

	testCases := []struct {
		name          string
		setupFunc     func() error
		forceFetch    bool
		expectedError string
	}{
		{
			name: "HappyPath",
			setupFunc: func() error {
				return nil
			},
			forceFetch:    true,
			expectedError: "",
		},
		{
			name: "InvalidSchema",
			setupFunc: func() error {
				return os.WriteFile(filename, []byte(`{"invalid":"schema"}`), 0o600)
			},
			forceFetch:    true,
			expectedError: "unknown field",
		},
		{
			name: "EmptyFile",
			setupFunc: func() error {
				return os.WriteFile(filename, []byte(``), 0o600)
			},
			forceFetch:    true,
			expectedError: "EOF",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.setupFunc(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			got, err := s.GetAllModelPresets(
				ctx,
				&spec.GetAllModelPresetsRequest{ForceFetch: tc.forceFetch},
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
				if _, ok := got.Body.ModelPresets[provider][modelName]; !ok {
					t.Errorf("expected provider/model to be present")
				}
			}
		})
	}
}

func TestModelPresetStore_CreateModelPresets(t *testing.T) {
	filename := "test_modelpresets_create.json"
	defer os.Remove(filename)
	s := newTestStore(t, filename)
	ctx := t.Context()

	provider := spec.ProviderName("openai2")
	modelName := spec.ModelName("gpt-4")
	preset := spec.ModelPreset{
		Name:         modelName,
		DisplayName:  "GPT-4",
		ShortCommand: "g4",
		IsEnabled:    true,
	}
	body := spec.ProviderModelPresets{
		modelName: preset,
	}

	testCases := []struct {
		name          string
		req           *spec.CreateModelPresetsRequest
		setupFunc     func() error
		wantErr       bool
		expectedError string
	}{
		{
			name:    "NilRequest",
			req:     nil,
			wantErr: true, expectedError: "request or request body cannot be nil",
		},
		{
			name: "NilBody",
			req: &spec.CreateModelPresetsRequest{
				ProviderName: provider,
				Body:         nil,
			},
			wantErr: true, expectedError: "request or request body cannot be nil",
		},
		{
			name: "HappyPath",
			req: &spec.CreateModelPresetsRequest{
				ProviderName: provider,
				Body:         &body,
			},
			wantErr: false,
		},
		{
			name: "DuplicateProvider",
			req: &spec.CreateModelPresetsRequest{
				ProviderName: provider,
				Body:         &body,
			},
			setupFunc: func() error {
				_, err := s.CreateModelPresets(ctx, &spec.CreateModelPresetsRequest{
					ProviderName: provider,
					Body:         &body,
				})
				return err
			},
			wantErr:       true,
			expectedError: "already exists",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupFunc != nil {
				_ = tc.setupFunc()
			}
			_, err := s.CreateModelPresets(ctx, tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("CreateModelPresets() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && tc.expectedError != "" &&
				!strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf(
					"CreateModelPresets() error = %v, want substring %q",
					err,
					tc.expectedError,
				)
			}
		})
	}
}

func TestModelPresetStore_DeleteModelPresets(t *testing.T) {
	filename := "test_modelpresets_delete.json"
	defer os.Remove(filename)
	s := newTestStore(t, filename)
	ctx := t.Context()

	provider := spec.ProviderName("openai2")
	modelName := spec.ModelName("gpt-4")
	body := spec.ProviderModelPresets{
		modelName: spec.ModelPreset{
			Name:         modelName,
			DisplayName:  "GPT-4",
			ShortCommand: "g4",
			IsEnabled:    true,
		},
	}
	_, _ = s.CreateModelPresets(ctx, &spec.CreateModelPresetsRequest{
		ProviderName: provider,
		Body:         &body,
	})

	testCases := []struct {
		name          string
		req           *spec.DeleteModelPresetsRequest
		wantErr       bool
		setupFunc     func() error
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
			req: &spec.DeleteModelPresetsRequest{
				ProviderName: "cohere",
			},
			wantErr:       true,
			expectedError: "does not exist",
		},
		{
			name: "DeleteExistingProvider",
			req: &spec.DeleteModelPresetsRequest{
				ProviderName: provider,
			},
			wantErr: false,
		},
		{
			name: "DeleteAgain",
			req: &spec.DeleteModelPresetsRequest{
				ProviderName: provider,
			},
			setupFunc: func() error {
				_, _ = s.DeleteModelPresets(
					ctx,
					&spec.DeleteModelPresetsRequest{ProviderName: provider},
				)
				return nil
			},
			wantErr:       true,
			expectedError: "does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupFunc != nil {
				_ = tc.setupFunc()
			}
			_, err := s.DeleteModelPresets(ctx, tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteModelPresets() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && tc.expectedError != "" &&
				!strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf(
					"DeleteModelPresets() error = %v, want substring %q",
					err,
					tc.expectedError,
				)
			}
		})
	}
}

func TestModelPresetStore_AddModelPreset(t *testing.T) {
	filename := "test_modelpresets_addmodel.json"
	defer os.Remove(filename)
	s := newTestStore(t, filename)
	ctx := t.Context()

	provider := spec.ProviderName("openai2")
	_, _ = s.CreateModelPresets(ctx, &spec.CreateModelPresetsRequest{
		ProviderName: provider,
		Body:         &spec.ProviderModelPresets{},
	})

	modelName := spec.ModelName("gpt-4")
	preset := spec.ModelPreset{
		Name:         modelName,
		DisplayName:  "GPT-4",
		ShortCommand: "g4",
		IsEnabled:    true,
	}

	testCases := []struct {
		name          string
		req           *spec.AddModelPresetRequest
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
			req: &spec.AddModelPresetRequest{
				ProviderName: provider,
				ModelName:    "foo",
				Body:         nil,
			},
			wantErr:       true,
			expectedError: "request or request body cannot be nil",
		},
		{
			name: "UnknownProvider",
			req: &spec.AddModelPresetRequest{
				ProviderName: "cohere",
				ModelName:    "foo",
				Body:         &preset,
			},
			wantErr:       true,
			expectedError: "does not exist",
		},
		{
			name: "ValidRequest",
			req: &spec.AddModelPresetRequest{
				ProviderName: provider,
				ModelName:    modelName,
				Body:         &preset,
			},
			wantErr: false,
		},
		{
			name: "OverwriteExistingModel",
			req: &spec.AddModelPresetRequest{
				ProviderName: provider,
				ModelName:    modelName,
				Body: &spec.ModelPreset{
					Name:         modelName,
					DisplayName:  "Changed",
					ShortCommand: "g4",
					IsEnabled:    false,
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.AddModelPreset(ctx, tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("AddModelPreset() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && tc.expectedError != "" &&
				!strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("AddModelPreset() error = %v, want substring %q", err, tc.expectedError)
			}
		})
	}
}

func TestModelPresetStore_DeleteModelPreset(t *testing.T) {
	filename := "test_modelpresets_delmodel.json"
	defer os.Remove(filename)
	s := newTestStore(t, filename)
	ctx := t.Context()

	provider := spec.ProviderName("openai2")
	modelName := spec.ModelName("gpt-4")
	_, _ = s.CreateModelPresets(ctx, &spec.CreateModelPresetsRequest{
		ProviderName: provider,
		Body:         &spec.ProviderModelPresets{},
	})
	_, _ = s.AddModelPreset(ctx, &spec.AddModelPresetRequest{
		ProviderName: provider,
		ModelName:    modelName,
		Body: &spec.ModelPreset{
			Name:         modelName,
			DisplayName:  "GPT-4",
			ShortCommand: "g4",
			IsEnabled:    true,
		},
	})

	testCases := []struct {
		name          string
		req           *spec.DeleteModelPresetRequest
		setupFunc     func() error
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
			name: "UnknownProvider",
			req: &spec.DeleteModelPresetRequest{
				ProviderName: "cohere",
				ModelName:    "foo",
			},
			wantErr:       true,
			expectedError: "does not exist",
		},
		{
			name: "DeleteExistingModel",
			req: &spec.DeleteModelPresetRequest{
				ProviderName: provider,
				ModelName:    modelName,
			},
			wantErr: false,
		},
		{
			name: "DeleteAgain",
			req: &spec.DeleteModelPresetRequest{
				ProviderName: provider,
				ModelName:    modelName,
			},
			setupFunc: func() error {
				_, _ = s.DeleteModelPreset(ctx, &spec.DeleteModelPresetRequest{
					ProviderName: provider,
					ModelName:    modelName,
				})
				return nil
			},
			wantErr: false,
		},
		{
			name: "EmptyModelName",
			req: &spec.DeleteModelPresetRequest{
				ProviderName: provider,
				ModelName:    "",
			},
			// Should not error, just no-op.
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupFunc != nil {
				_ = tc.setupFunc()
			}
			_, err := s.DeleteModelPreset(ctx, tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteModelPreset() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && tc.expectedError != "" &&
				!strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("DeleteModelPreset() error = %v, want substring %q", err, tc.expectedError)
			}
		})
	}
}

func TestModelPresetStore_Persistence(t *testing.T) {
	filename := "test_modelpresets_persist.json"
	defer os.Remove(filename)
	s := newTestStore(t, filename)
	ctx := t.Context()

	provider := spec.ProviderName("persistprov")
	modelName := spec.ModelName("persistmodel")
	body := spec.ProviderModelPresets{
		modelName: spec.ModelPreset{
			Name:         modelName,
			DisplayName:  "Persist Model",
			ShortCommand: "pm",
			IsEnabled:    true,
		},
	}
	_, err := s.CreateModelPresets(ctx, &spec.CreateModelPresetsRequest{
		ProviderName: provider,
		Body:         &body,
	})
	if err != nil {
		t.Fatalf("CreateModelPresets failed: %v", err)
	}

	// Re-open store.
	s2 := &store.ModelPresetStore{}
	if err := store.InitModelPresetStore(s2, filename); err != nil {
		t.Fatalf("InitModelPresetStore failed: %v", err)
	}

	resp, err := s2.GetAllModelPresets(ctx, &spec.GetAllModelPresetsRequest{})
	if err != nil {
		t.Fatalf("GetAllModelPresets failed: %v", err)
	}
	found := false
	for pname, pm := range resp.Body.ModelPresets {
		if pname == provider {
			if _, ok := pm[modelName]; ok {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("persisted provider/model not found after reload")
	}
}

func TestModelPresetStore_BoundaryCases(t *testing.T) {
	filename := "test_modelpresets_boundary.json"
	defer os.Remove(filename)
	s := newTestStore(t, filename)
	ctx := t.Context()
	provider := spec.ProviderName("boundaryprov")
	// Empty provider model preset.
	body := spec.ProviderModelPresets{}
	_, err := s.CreateModelPresets(ctx, &spec.CreateModelPresetsRequest{
		ProviderName: provider,
		Body:         &body,
	})
	if err != nil {
		t.Fatalf("CreateModelPresets failed: %v", err)
	}
	// Add model with minimal fields.
	modelName := spec.ModelName("minmodel")
	preset := spec.ModelPreset{
		Name:         modelName,
		DisplayName:  "",
		ShortCommand: "",
		IsEnabled:    false,
	}
	_, err = s.AddModelPreset(ctx, &spec.AddModelPresetRequest{
		ProviderName: provider,
		ModelName:    modelName,
		Body:         &preset,
	})
	if err != nil {
		t.Fatalf("AddModelPreset failed: %v", err)
	}
	// Add model with all optional fields set.
	modelName2 := spec.ModelName("fullmodel")
	temp := 0.5
	maxPrompt := 1000
	maxOutput := 200
	timeout := 30
	systemPrompt := "sys"
	reasoning := &spec.ReasoningParams{
		Type:   spec.ReasoningTypeHybridWithTokens,
		Level:  spec.ReasoningLevelHigh,
		Tokens: 42,
	}
	preset2 := spec.ModelPreset{
		Name:                 modelName2,
		DisplayName:          "Full",
		ShortCommand:         "f",
		IsEnabled:            true,
		Stream:               boolPtr(true),
		MaxPromptLength:      &maxPrompt,
		MaxOutputLength:      &maxOutput,
		Temperature:          &temp,
		Reasoning:            reasoning,
		SystemPrompt:         &systemPrompt,
		Timeout:              &timeout,
		AdditionalParameters: map[string]any{"foo": "bar"},
	}
	_, err = s.AddModelPreset(ctx, &spec.AddModelPresetRequest{
		ProviderName: provider,
		ModelName:    modelName2,
		Body:         &preset2,
	})
	if err != nil {
		t.Fatalf("AddModelPreset (full) failed: %v", err)
	}
	// Check round-trip.
	resp, err := s.GetAllModelPresets(ctx, &spec.GetAllModelPresetsRequest{})
	if err != nil {
		t.Fatalf("GetAllModelPresets failed: %v", err)
	}
	got := resp.Body.ModelPresets[provider][modelName2]
	if !reflect.DeepEqual(got, preset2) {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, preset2)
	}
}

func boolPtr(b bool) *bool { return &b }
