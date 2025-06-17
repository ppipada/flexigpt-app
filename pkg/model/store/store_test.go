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
	modelPresetID := spec.ModelPresetID("gpt4")
	preset := spec.ModelPreset{
		ID:           modelPresetID,
		Name:         modelName,
		DisplayName:  "GPT-4",
		ShortCommand: "g4",
		IsEnabled:    true,
	}
	body := spec.ProviderPreset{
		DefaultModelPresetID: modelPresetID,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			modelPresetID: preset,
		},
	}
	_, err := s.CreateProviderPreset(ctx, &spec.CreateProviderPresetRequest{
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
				if _, ok := got.Body.ProviderPresets[provider].ModelPresets[modelPresetID]; !ok {
					t.Errorf("expected provider/model to be present")
				}
			}
		})
	}
}

func TestModelPresetStore_CreateProviderPreset(t *testing.T) {
	filename := "test_modelpresets_create.json"
	defer os.Remove(filename)
	s := newTestStore(t, filename)
	ctx := t.Context()

	provider := spec.ProviderName("openai2")
	modelPresetID := spec.ModelPresetID("gpt4")
	modelName := spec.ModelName("gpt-4")
	preset := spec.ModelPreset{
		ID:           modelPresetID,
		Name:         modelName,
		DisplayName:  "GPT-4",
		ShortCommand: "g4",
		IsEnabled:    true,
	}
	body := spec.ProviderPreset{
		DefaultModelPresetID: modelPresetID,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			modelPresetID: preset,
		},
	}

	testCases := []struct {
		name          string
		req           *spec.CreateProviderPresetRequest
		setupFunc     func() error
		wantErr       bool
		expectedError string
	}{
		{
			name:    "NilRequest",
			req:     nil,
			wantErr: true, expectedError: "invalid request",
		},
		{
			name: "NilBody",
			req: &spec.CreateProviderPresetRequest{
				ProviderName: provider,
				Body:         nil,
			},
			wantErr: true, expectedError: "invalid request",
		},
		{
			name: "MissingDefaultModelPresetID",
			req: &spec.CreateProviderPresetRequest{
				ProviderName: provider,
				Body: &spec.ProviderPreset{
					DefaultModelPresetID: "",
					ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
						modelPresetID: preset,
					},
				},
			},
			wantErr: true, expectedError: "invalid request",
		},
		{
			name: "MissingModelPresets",
			req: &spec.CreateProviderPresetRequest{
				ProviderName: provider,
				Body: &spec.ProviderPreset{
					DefaultModelPresetID: modelPresetID,
					ModelPresets:         nil,
				},
			},
			wantErr: true, expectedError: "invalid request",
		},
		{
			name: "EmptyModelPresets",
			req: &spec.CreateProviderPresetRequest{
				ProviderName: provider,
				Body: &spec.ProviderPreset{
					DefaultModelPresetID: modelPresetID,
					ModelPresets:         map[spec.ModelPresetID]spec.ModelPreset{},
				},
			},
			wantErr: true, expectedError: "invalid request",
		},
		{
			name: "HappyPath",
			req: &spec.CreateProviderPresetRequest{
				ProviderName: provider,
				Body:         &body,
			},
			wantErr: false,
		},
		{
			name: "DuplicateProvider",
			req: &spec.CreateProviderPresetRequest{
				ProviderName: provider,
				Body:         &body,
			},
			setupFunc: func() error {
				_, err := s.CreateProviderPreset(ctx, &spec.CreateProviderPresetRequest{
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
			_, err := s.CreateProviderPreset(ctx, tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("CreateProviderPreset() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && tc.expectedError != "" &&
				!strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf(
					"CreateProviderPreset() error = %v, want substring %q",
					err,
					tc.expectedError,
				)
			}
		})
	}
}

func TestModelPresetStore_DeleteProviderPreset(t *testing.T) {
	filename := "test_modelpresets_delete.json"
	defer os.Remove(filename)
	s := newTestStore(t, filename)
	ctx := t.Context()

	provider := spec.ProviderName("openai2")
	modelPresetID := spec.ModelPresetID("gpt4")
	modelName := spec.ModelName("gpt-4")
	body := spec.ProviderPreset{
		DefaultModelPresetID: modelPresetID,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			modelPresetID: {
				ID:           modelPresetID,
				Name:         modelName,
				DisplayName:  "GPT-4",
				ShortCommand: "g4",
				IsEnabled:    true,
			},
		},
	}
	_, err := s.CreateProviderPreset(ctx, &spec.CreateProviderPresetRequest{
		ProviderName: provider,
		Body:         &body,
	})
	if err != nil {
		t.Fatalf("Failed to add provider: %v", err)
	}

	testCases := []struct {
		name          string
		req           *spec.DeleteProviderPresetRequest
		wantErr       bool
		setupFunc     func() error
		expectedError string
	}{
		{
			name:          "NilRequest",
			req:           nil,
			wantErr:       true,
			expectedError: "invalid request",
		},
		{
			name: "DeleteNonExistingProvider",
			req: &spec.DeleteProviderPresetRequest{
				ProviderName: "cohere",
			},
			wantErr:       true,
			expectedError: "does not exist",
		},
		{
			name: "DeleteExistingProvider",
			req: &spec.DeleteProviderPresetRequest{
				ProviderName: provider,
			},
			wantErr: false,
		},
		{
			name: "DeleteAgain",
			req: &spec.DeleteProviderPresetRequest{
				ProviderName: provider,
			},
			setupFunc: func() error {
				_, _ = s.DeleteProviderPreset(
					ctx,
					&spec.DeleteProviderPresetRequest{ProviderName: provider},
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
			_, err := s.DeleteProviderPreset(ctx, tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("DeleteProviderPreset() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && tc.expectedError != "" &&
				!strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf(
					"DeleteProviderPreset() error = %v, want substring %q",
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
	modelPresetID := spec.ModelPresetID("gpt4")
	modelName := spec.ModelName("gpt-4")
	// Add provider first.
	providerBody := spec.ProviderPreset{
		DefaultModelPresetID: modelPresetID,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			spec.ModelPresetID("g1"): {
				ID:           "g1",
				Name:         "g-1",
				DisplayName:  "G-1",
				ShortCommand: "g1",
				IsEnabled:    true,
			},
		},
	}
	_, err := s.CreateProviderPreset(ctx, &spec.CreateProviderPresetRequest{
		ProviderName: provider,
		Body:         &providerBody,
	})
	if err != nil {
		t.Fatalf("Failed to add provider: %v", err)
	}

	preset := spec.ModelPreset{
		ID:           modelPresetID,
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
			expectedError: "invalid request",
		},
		{
			name: "NilBody",
			req: &spec.AddModelPresetRequest{
				ProviderName:  provider,
				ModelPresetID: modelPresetID,
				Body:          nil,
			},
			wantErr:       true,
			expectedError: "invalid request",
		},
		{
			name: "EmptyProviderName",
			req: &spec.AddModelPresetRequest{
				ProviderName:  "",
				ModelPresetID: modelPresetID,
				Body:          &preset,
			},
			wantErr:       true,
			expectedError: "invalid request",
		},
		{
			name: "EmptyModelPresetID",
			req: &spec.AddModelPresetRequest{
				ProviderName:  provider,
				ModelPresetID: "",
				Body:          &preset,
			},
			wantErr:       true,
			expectedError: "invalid request",
		},
		{
			name: "EmptyBodyID",
			req: &spec.AddModelPresetRequest{
				ProviderName:  provider,
				ModelPresetID: modelPresetID,
				Body: &spec.ModelPreset{
					ID:           "",
					Name:         modelName,
					DisplayName:  "GPT-4",
					ShortCommand: "g4",
					IsEnabled:    true,
				},
			},
			wantErr:       true,
			expectedError: "invalid request",
		},
		{
			name: "UnknownProvider",
			req: &spec.AddModelPresetRequest{
				ProviderName:  "cohere",
				ModelPresetID: "foo",
				Body: &spec.ModelPreset{
					ID:           "foo",
					Name:         "foo",
					DisplayName:  "Foo",
					ShortCommand: "f",
					IsEnabled:    true,
				},
			},
			wantErr:       true,
			expectedError: "does not exist",
		},
		{
			name: "ValidRequest",
			req: &spec.AddModelPresetRequest{
				ProviderName:  provider,
				ModelPresetID: modelPresetID,
				Body:          &preset,
			},
			wantErr: false,
		},
		{
			name: "OverwriteExistingModel",
			req: &spec.AddModelPresetRequest{
				ProviderName:  provider,
				ModelPresetID: modelPresetID,
				Body: &spec.ModelPreset{
					ID:           modelPresetID,
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
	modelPresetID := spec.ModelPresetID("gpt4")
	modelName := spec.ModelName("gpt-4")
	providerBody := spec.ProviderPreset{
		DefaultModelPresetID: modelPresetID,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			spec.ModelPresetID("g1"): {
				ID:           "g1",
				Name:         "g-1",
				DisplayName:  "G-1",
				ShortCommand: "g1",
				IsEnabled:    true,
			},
		},
	}
	_, err := s.CreateProviderPreset(ctx, &spec.CreateProviderPresetRequest{
		ProviderName: provider,
		Body:         &providerBody,
	})
	if err != nil {
		t.Fatalf("Failed to add provider: %v", err)
	}
	_, err = s.AddModelPreset(ctx, &spec.AddModelPresetRequest{
		ProviderName:  provider,
		ModelPresetID: modelPresetID,
		Body: &spec.ModelPreset{
			ID:           modelPresetID,
			Name:         modelName,
			DisplayName:  "GPT-4",
			ShortCommand: "g4",
			IsEnabled:    true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to add model preset: %v", err)
	}

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
				ProviderName:  "cohere",
				ModelPresetID: "foo",
			},
			wantErr:       true,
			expectedError: "does not exist",
		},
		{
			name: "DeleteExistingModel",
			req: &spec.DeleteModelPresetRequest{
				ProviderName:  provider,
				ModelPresetID: modelPresetID,
			},
			wantErr: false,
		},
		{
			name: "DeleteAgain",
			req: &spec.DeleteModelPresetRequest{
				ProviderName:  provider,
				ModelPresetID: modelPresetID,
			},
			setupFunc: func() error {
				_, _ = s.DeleteModelPreset(ctx, &spec.DeleteModelPresetRequest{
					ProviderName:  provider,
					ModelPresetID: modelPresetID,
				})
				return nil
			},
			wantErr: false,
		},
		{
			name: "EmptyModelPresetID",
			req: &spec.DeleteModelPresetRequest{
				ProviderName:  provider,
				ModelPresetID: "",
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
	modelPresetID := spec.ModelPresetID("persistmodelid")
	modelName := spec.ModelName("persistmodel")
	body := spec.ProviderPreset{
		DefaultModelPresetID: modelPresetID,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			modelPresetID: {
				ID:           modelPresetID,
				Name:         modelName,
				DisplayName:  "Persist Model",
				ShortCommand: "pm",
				IsEnabled:    true,
			},
		},
	}
	_, err := s.CreateProviderPreset(ctx, &spec.CreateProviderPresetRequest{
		ProviderName: provider,
		Body:         &body,
	})
	if err != nil {
		t.Fatalf("CreateProviderPreset failed: %v", err)
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
	for pname, pp := range resp.Body.ProviderPresets {
		if pname == provider {
			if _, ok := pp.ModelPresets[modelPresetID]; ok {
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
	modelPresetID1 := spec.ModelPresetID("minmodelid")
	modelName1 := spec.ModelName("minmodel")
	modelPresetID2 := spec.ModelPresetID("fullmodelid")
	modelName2 := spec.ModelName("fullmodel")

	// Create provider with empty ModelPresets, but valid DefaultModelPresetID (must be present and in map).
	emptyPreset := spec.ModelPreset{
		ID:           modelPresetID1,
		Name:         modelName1,
		DisplayName:  "",
		ShortCommand: "",
		IsEnabled:    false,
	}
	providerBody := spec.ProviderPreset{
		DefaultModelPresetID: modelPresetID1,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			modelPresetID1: emptyPreset,
		},
	}
	_, err := s.CreateProviderPreset(ctx, &spec.CreateProviderPresetRequest{
		ProviderName: provider,
		Body:         &providerBody,
	})
	if err != nil {
		t.Fatalf("CreateProviderPreset failed: %v", err)
	}

	// Add model with all optional fields set.
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
	fullPreset := spec.ModelPreset{
		ID:                   modelPresetID2,
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
		ProviderName:  provider,
		ModelPresetID: modelPresetID2,
		Body:          &fullPreset,
	})
	if err != nil {
		t.Fatalf("AddModelPreset (full) failed: %v", err)
	}
	// Check round-trip.
	resp, err := s.GetAllModelPresets(ctx, &spec.GetAllModelPresetsRequest{})
	if err != nil {
		t.Fatalf("GetAllModelPresets failed: %v", err)
	}
	got := resp.Body.ProviderPresets[provider].ModelPresets[modelPresetID2]
	if !reflect.DeepEqual(got, fullPreset) {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, fullPreset)
	}
}

func TestModelPresetStore_SetDefaultModelPreset(t *testing.T) {
	filename := "test_modelpresets_setdefault.json"
	defer os.Remove(filename)
	s := newTestStore(t, filename)
	ctx := t.Context()

	provider := spec.ProviderName("openai2")
	modelPresetID1 := spec.ModelPresetID("gpt4")
	modelPresetID2 := spec.ModelPresetID("gpt3")
	modelName1 := spec.ModelName("gpt-4")
	modelName2 := spec.ModelName("gpt-3")

	// Add provider with two model presets.
	body := spec.ProviderPreset{
		DefaultModelPresetID: modelPresetID1,
		ModelPresets: map[spec.ModelPresetID]spec.ModelPreset{
			modelPresetID1: {
				ID:           modelPresetID1,
				Name:         modelName1,
				DisplayName:  "GPT-4",
				ShortCommand: "g4",
				IsEnabled:    true,
			},
			modelPresetID2: {
				ID:           modelPresetID2,
				Name:         modelName2,
				DisplayName:  "GPT-3",
				ShortCommand: "g3",
				IsEnabled:    true,
			},
		},
	}
	_, err := s.CreateProviderPreset(ctx, &spec.CreateProviderPresetRequest{
		ProviderName: provider,
		Body:         &body,
	})
	if err != nil {
		t.Fatalf("Failed to add provider: %v", err)
	}

	testCases := []struct {
		name          string
		req           *spec.SetDefaultModelPresetRequest
		setupFunc     func() error
		wantErr       bool
		expectedError string
		verifyFunc    func(t *testing.T)
	}{
		{
			name:          "NilRequest",
			req:           nil,
			wantErr:       true,
			expectedError: "invalid request",
		},
		{
			name: "NilBody",
			req: &spec.SetDefaultModelPresetRequest{
				ProviderName: provider,
				Body:         nil,
			},
			wantErr:       true,
			expectedError: "invalid request",
		},
		{
			name: "EmptyModelPresetID",
			req: &spec.SetDefaultModelPresetRequest{
				ProviderName: provider,
				Body:         &spec.SetDefaultModelPresetRequestBody{ModelPresetID: ""},
			},
			wantErr:       true,
			expectedError: "invalid request",
		},
		{
			name: "UnknownProvider",
			req: &spec.SetDefaultModelPresetRequest{
				ProviderName: "unknownprov",
				Body:         &spec.SetDefaultModelPresetRequestBody{ModelPresetID: modelPresetID1},
			},
			wantErr:       true,
			expectedError: "does not exist",
		},
		{
			name: "UnknownModelPresetID",
			req: &spec.SetDefaultModelPresetRequest{
				ProviderName: provider,
				Body:         &spec.SetDefaultModelPresetRequestBody{ModelPresetID: "notfound"},
			},
			wantErr:       true,
			expectedError: "model preset not found",
		},
		{
			name: "HappyPath",
			req: &spec.SetDefaultModelPresetRequest{
				ProviderName: provider,
				Body:         &spec.SetDefaultModelPresetRequestBody{ModelPresetID: modelPresetID2},
			},
			wantErr: false,
			verifyFunc: func(t *testing.T) {
				resp, err := s.GetAllModelPresets(ctx, &spec.GetAllModelPresetsRequest{})
				if err != nil {
					t.Fatalf("GetAllModelPresets failed: %v", err)
				}
				got := resp.Body.ProviderPresets[provider].DefaultModelPresetID
				if got != modelPresetID2 {
					t.Errorf("DefaultModelPresetID = %v, want %v", got, modelPresetID2)
				}
			},
		},
		{
			name: "ChangeDefaultAgain",
			req: &spec.SetDefaultModelPresetRequest{
				ProviderName: provider,
				Body:         &spec.SetDefaultModelPresetRequestBody{ModelPresetID: modelPresetID1},
			},
			wantErr: false,
			verifyFunc: func(t *testing.T) {
				resp, err := s.GetAllModelPresets(ctx, &spec.GetAllModelPresetsRequest{})
				if err != nil {
					t.Fatalf("GetAllModelPresets failed: %v", err)
				}
				got := resp.Body.ProviderPresets[provider].DefaultModelPresetID
				if got != modelPresetID1 {
					t.Errorf("DefaultModelPresetID = %v, want %v", got, modelPresetID1)
				}
			},
		},
		{
			name: "ModelPresetIDPresentButNotInMap",
			req: &spec.SetDefaultModelPresetRequest{
				ProviderName: provider,
				Body:         &spec.SetDefaultModelPresetRequestBody{ModelPresetID: "ghost"},
			},
			wantErr:       true,
			expectedError: "model preset not found",
		},
		{
			name: "ModelPresetIDIsEmptyString",
			req: &spec.SetDefaultModelPresetRequest{
				ProviderName: provider,
				Body:         &spec.SetDefaultModelPresetRequestBody{ModelPresetID: ""},
			},
			wantErr:       true,
			expectedError: "invalid request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupFunc != nil {
				_ = tc.setupFunc()
			}
			_, err := s.SetDefaultModelPreset(ctx, tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("SetDefaultModelPreset() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && tc.expectedError != "" &&
				!strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf(
					"SetDefaultModelPreset() error = %v, want substring %q",
					err,
					tc.expectedError,
				)
			}
			if err == nil && tc.verifyFunc != nil {
				tc.verifyFunc(t)
			}
		})
	}
}

func boolPtr(b bool) *bool { return &b }
