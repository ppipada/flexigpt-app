package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/builtin"
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
)

const corrupted = "corrupted"

func TestNewBuiltInPresets(t *testing.T) {
	tests := []struct {
		name           string
		setupDir       func(*testing.T) string
		snapshotMaxAge time.Duration
		wantErr        bool
		errContains    string
	}{
		{
			name: "happy_path",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
		},
		{
			name: "zero_snapshot_age_defaults",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			snapshotMaxAge: 0,
		},
		{
			name: "negative_snapshot_age_defaults",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			snapshotMaxAge: -1,
		},
		{
			name: "base_dir_is_file",
			setupDir: func(t *testing.T) string {
				t.Helper()
				tmp := t.TempDir()
				f := filepath.Join(tmp, "file")
				_ = os.WriteFile(f, []byte("dummy"), 0o600)
				return f
			},
			wantErr: true,
		},
		{
			name: "readonly_directory",
			setupDir: func(t *testing.T) string {
				t.Helper()
				tmp := t.TempDir()
				ro := filepath.Join(tmp, "ro")
				_ = os.MkdirAll(ro, 0o444)
				return ro
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			dir := tc.setupDir(t)
			bi, err := NewBuiltInPresets(ctx, dir, tc.snapshotMaxAge)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("error %q does not contain %q", err, tc.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			prov, models, _ := bi.ListBuiltInPresets(ctx)
			if len(prov) == 0 || len(models) == 0 {
				t.Fatal("expected non-empty data")
			}
			if _, err := os.Stat(filepath.Join(dir, spec.ModelPresetsBuiltInOverlayDBFileName)); err != nil {
				t.Fatalf("overlay file missing: %v", err)
			}
			for n, p := range prov {
				if !p.IsBuiltIn {
					t.Errorf("provider %s not flagged built-in", n)
				}
			}
			for pn, mm := range models {
				for mid, mp := range mm {
					if !mp.IsBuiltIn {
						t.Errorf("model %s/%s not flagged built-in", pn, mid)
					}
				}
			}
		})
	}
}

func TestSetProviderEnabled(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T, *BuiltInPresets) (spec.ProviderName, bool)
		wantErr bool
	}{
		{
			name: "toggle_existing_provider",
			setup: func(t *testing.T, bi *BuiltInPresets) (spec.ProviderName, bool) {
				t.Helper()
				prov, _, _ := bi.ListBuiltInPresets(t.Context())
				id, p := anyProvider(prov)
				return id, !p.IsEnabled
			},
		},
		{
			name: "nonexistent_provider",
			setup: func(t *testing.T, _ *BuiltInPresets) (spec.ProviderName, bool) {
				t.Helper()
				return spec.ProviderName("ghost"), true
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			dir := t.TempDir()
			bi, _ := NewBuiltInPresets(ctx, dir, 0)
			pname, enabled := tc.setup(t, bi)
			_, err := bi.SetProviderEnabled(ctx, pname, enabled)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected: %v", err)
			}
			prov, _, _ := bi.ListBuiltInPresets(ctx)
			if prov[pname].IsEnabled != enabled {
				t.Errorf("flag mismatch want %v", enabled)
			}
			// Check overlay state via providerOverlayFlags API.
			flag, ok, err := bi.providerOverlayFlags.GetFlag(ctx, builtInProviderKey(pname))
			if err != nil {
				t.Fatalf("providerOverlayFlags.GetFlag: %v", err)
			}
			if !ok || flag.Value != enabled {
				t.Errorf(
					"overlay mismatch: present=%v, value=%v, want present=true, value=%v",
					ok,
					flag.Value,
					enabled,
				)
			}
		})
	}
}

func TestSetModelPresetEnabled(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T, *BuiltInPresets) (spec.ProviderName, spec.ModelPreset, bool)
		wantErr bool
	}{
		{
			name: "toggle_existing_model",
			setup: func(t *testing.T, bi *BuiltInPresets) (spec.ProviderName, spec.ModelPreset, bool) {
				t.Helper()
				_, models, _ := bi.ListBuiltInPresets(t.Context())
				pn, _, mp := anyModel(models)
				return pn, mp, !mp.IsEnabled
			},
		},
		{
			name: "nonexistent_provider",
			setup: func(*testing.T, *BuiltInPresets) (spec.ProviderName, spec.ModelPreset, bool) {
				return spec.ProviderName("ghost"), spec.ModelPreset{ID: "m"}, true
			},
			wantErr: true,
		},
		{
			name: "nonexistent_model",
			setup: func(t *testing.T, bi *BuiltInPresets) (spec.ProviderName, spec.ModelPreset, bool) {
				t.Helper()
				prov, _, _ := bi.ListBuiltInPresets(t.Context())
				pn, _ := anyProvider(prov)
				return pn, spec.ModelPreset{ID: "ghost"}, true
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			dir := t.TempDir()
			bi, _ := NewBuiltInPresets(ctx, dir, 0)
			pn, mp, enabled := tc.setup(t, bi)
			_, err := bi.SetModelPresetEnabled(ctx, pn, mp.ID, enabled)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected: %v", err)
			}
			_, models, _ := bi.ListBuiltInPresets(ctx)
			if models[pn][mp.ID].IsEnabled != enabled {
				t.Errorf("flag mismatch want %v", enabled)
			}
			// Check overlay state via modelOverlayFlags API.
			flag, ok, err := bi.modelOverlayFlags.GetFlag(ctx, builtInModelKey(mp.ID))
			if err != nil {
				t.Fatalf("modelOverlayFlags.GetFlag: %v", err)
			}
			if !ok || flag.Value != enabled {
				t.Errorf(
					"overlay mismatch: present=%v, value=%v, want present=true, value=%v",
					ok,
					flag.Value,
					enabled,
				)
			}
		})
	}
}

func TestListBuiltInPresets(t *testing.T) {
	ctx := t.Context()
	dir := t.TempDir()
	bi, _ := NewBuiltInPresets(ctx, dir, 0)

	t.Run("independent_copies", func(t *testing.T) {
		p1, m1, _ := bi.ListBuiltInPresets(ctx)
		p2, m2, _ := bi.ListBuiltInPresets(ctx)

		for n := range p1 {
			p1[n] = spec.ProviderPreset{DisplayName: corrupted}
			break
		}
		for pn := range m1 {
			for mid := range m1[pn] {
				mp := m1[pn][mid]
				mp.DisplayName = corrupted
				m1[pn][mid] = mp
				break
			}
			break
		}
		for n, p := range p2 {
			if p.DisplayName == corrupted {
				t.Errorf("provider %s shares memory", n)
			}
		}
		for pn, mm := range m2 {
			for mid, mp := range mm {
				if mp.DisplayName == corrupted {
					t.Errorf("model %s/%s shares memory", pn, mid)
				}
			}
		}
	})

	t.Run("consistent_maps", func(t *testing.T) {
		prov, models, _ := bi.ListBuiltInPresets(ctx)
		for pn := range prov {
			if _, ok := models[pn]; !ok {
				t.Errorf("provider %s has no model map", pn)
			}
		}
		for pn := range models {
			if _, ok := prov[pn]; !ok {
				t.Errorf("model map %s has no provider", pn)
			}
		}
	})
}

func TestConcurrencyPresets(t *testing.T) {
	ctx := t.Context()
	dir := t.TempDir()
	bi, _ := NewBuiltInPresets(ctx, dir, 10*time.Millisecond)

	// Fetch sample keys.
	prov, models, _ := bi.ListBuiltInPresets(ctx)
	pname, _ := anyProvider(prov)
	pn, mid, _ := anyModel(models)

	// Concurrent reads.
	t.Run("reads", func(t *testing.T) {
		var wg sync.WaitGroup
		const loops = 10
		for range loops {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 100 {
					_, _, _ = bi.ListBuiltInPresets(ctx)
				}
			}()
		}
		wg.Wait()
	})

	// Concurrent writes.
	t.Run("writes", func(t *testing.T) {
		var wg sync.WaitGroup
		const writers = 5
		for i := range writers {
			wg.Add(2)
			go func(i int) {
				defer wg.Done()
				_, _ = bi.SetProviderEnabled(ctx, pname, i%2 == 0)
			}(i)
			go func(i int) {
				defer wg.Done()
				_, _ = bi.SetModelPresetEnabled(ctx, pn, mid, i%2 == 0)
			}(i)
		}
		wg.Wait()
	})
}

func TestRebuildSnapshotPresets(t *testing.T) {
	ctx := t.Context()
	dir := t.TempDir()
	bi, _ := NewBuiltInPresets(ctx, dir, time.Hour)

	prov, _, _ := bi.ListBuiltInPresets(ctx)
	pn, p := anyProvider(prov)
	_, err := bi.providerOverlayFlags.SetFlag(ctx, builtInProviderKey(pn), !p.IsEnabled)
	if err != nil {
		t.Fatalf("providerOverlayFlags.SetFlag: %v", err)
	}

	bi.mu.Lock()
	_ = bi.rebuildSnapshot(ctx)
	bi.mu.Unlock()

	prov2, _, _ := bi.ListBuiltInPresets(ctx)
	if prov2[pn].IsEnabled == p.IsEnabled {
		t.Error("rebuild did not apply overlay")
	}
}

func TestAsyncRebuildPresets(t *testing.T) {
	ctx := t.Context()
	dir := t.TempDir()
	bi, _ := NewBuiltInPresets(ctx, dir, time.Millisecond)

	prov, _, _ := bi.ListBuiltInPresets(ctx)
	pn, p := anyProvider(prov)
	time.Sleep(2 * time.Millisecond)
	_, _ = bi.SetProviderEnabled(ctx, pn, !p.IsEnabled)
	time.Sleep(10 * time.Millisecond)

	prov2, _, _ := bi.ListBuiltInPresets(ctx)
	if prov2[pn].IsEnabled == p.IsEnabled {
		t.Error("async rebuild missed change")
	}
}

func Test_NewBuiltInPresets_SyntheticFS_Errors(t *testing.T) {
	provName := spec.ProviderName("demoProv")
	modelID := spec.ModelPresetID("model1")

	t.Run("missing_json", func(t *testing.T) {
		_, err := newPresetsFromFS(t, fstest.MapFS{})
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("want fs.ErrNotExist, got %v", err)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		fsys := fstest.MapFS{builtin.BuiltInModelPresetsJSON: {Data: []byte("{ nope ]")}}
		_, err := newPresetsFromFS(t, fsys)
		if err == nil {
			t.Fatal("want error")
		}
	})

	t.Run("no_providers", func(t *testing.T) {
		empty, _ := json.Marshal(spec.PresetsSchema{
			SchemaVersion:   spec.SchemaVersion,
			ProviderPresets: map[spec.ProviderName]spec.ProviderPreset{},
		})
		fsys := fstest.MapFS{builtin.BuiltInModelPresetsJSON: {Data: empty}}
		_, err := newPresetsFromFS(t, fsys)
		if err == nil || (!strings.Contains(err.Error(), "no providers") &&
			!strings.Contains(err.Error(), "no default provider in builtin")) {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("default_model_missing", func(t *testing.T) {
		s := buildSchemaDefaultMissing(provName, modelID)
		fsys := fstest.MapFS{builtin.BuiltInModelPresetsJSON: {Data: s}}
		_, err := newPresetsFromFS(t, fsys)
		if err == nil || !strings.Contains(err.Error(), "defaultModelPresetID") {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("duplicate_model_across_providers", func(t *testing.T) {
		s := buildSchemaDuplicateModel(modelID)
		fsys := fstest.MapFS{builtin.BuiltInModelPresetsJSON: {Data: s}}
		_, err := newPresetsFromFS(t, fsys)
		if err == nil || !strings.Contains(err.Error(), "duplicate modelPresetID") {
			t.Fatalf("unexpected: %v", err)
		}
	})
}

func Test_NewBuiltInPresets_SyntheticFS_HappyAndCRUD(t *testing.T) {
	ctx := t.Context()
	pn := spec.ProviderName("demo")
	mpid := spec.ModelPresetID("m1")

	schema := buildHappySchema(pn, mpid)
	fsys := fstest.MapFS{builtin.BuiltInModelPresetsJSON: {Data: schema}}

	dir := t.TempDir()
	bi, err := NewBuiltInPresets(ctx, dir, 0, WithModelPresetsFS(fsys, "."))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	prov, models, _ := bi.ListBuiltInPresets(ctx)
	if len(prov) != 1 || len(models) != 1 {
		t.Fatalf("want 1/1 objects, got %d/%d", len(prov), len(models))
	}

	// Toggle provider.
	p := prov[pn]
	_, _ = bi.SetProviderEnabled(ctx, pn, !p.IsEnabled)
	flag, ok, err := bi.providerOverlayFlags.GetFlag(ctx, builtInProviderKey(pn))
	if err != nil {
		t.Fatalf("providerOverlayFlags.GetFlag: %v", err)
	}
	if !ok || flag.Value == p.IsEnabled {
		t.Fatal("provider overlay not updated")
	}

	// Toggle model preset.
	mp := models[pn][mpid]
	_, _ = bi.SetModelPresetEnabled(ctx, pn, mpid, !mp.IsEnabled)
	mflag, ok, err := bi.modelOverlayFlags.GetFlag(ctx, builtInModelKey(mpid))
	if err != nil {
		t.Fatalf("modelOverlayFlags.GetFlag: %v", err)
	}
	if !ok || mflag.Value == mp.IsEnabled {
		t.Fatal("model overlay not updated")
	}
}

func anyProvider(
	m map[spec.ProviderName]spec.ProviderPreset,
) (spec.ProviderName, spec.ProviderPreset) {
	for n, p := range m {
		return n, p
	}
	return "", spec.ProviderPreset{}
}

func anyModel(m map[spec.ProviderName]map[spec.ModelPresetID]spec.ModelPreset,
) (spec.ProviderName, spec.ModelPresetID, spec.ModelPreset) {
	for pn, mm := range m {
		for mid, mp := range mm {
			return pn, mid, mp
		}
	}
	return "", "", spec.ModelPreset{}
}

func newPresetsFromFS(t *testing.T, mem fs.FS) (*BuiltInPresets, error) {
	t.Helper()
	ctx := t.Context()
	return NewBuiltInPresets(ctx, t.TempDir(), time.Hour, WithModelPresetsFS(mem, "."))
}

func buildSchemaDefaultMissing(pn spec.ProviderName, mpid spec.ModelPresetID) []byte {
	model := makeModelPreset(mpid)
	pp := spec.ProviderPreset{
		SchemaVersion:            spec.SchemaVersion,
		Name:                     pn,
		DisplayName:              "Demo",
		APIType:                  spec.ProviderAPITypeOpenAICompatible,
		IsEnabled:                true,
		CreatedAt:                time.Now(),
		ModifiedAt:               time.Now(),
		Origin:                   "https://x",
		ChatCompletionPathPrefix: spec.OpenAICompatibleChatCompletionPathPrefix,
		DefaultModelPresetID:     "ghost",
		ModelPresets:             map[spec.ModelPresetID]spec.ModelPreset{mpid: model},
	}
	s := spec.PresetsSchema{
		SchemaVersion:   spec.SchemaVersion,
		DefaultProvider: pn,
		ProviderPresets: map[spec.ProviderName]spec.ProviderPreset{pn: pp},
	}
	b, _ := json.Marshal(s)
	return b
}

func buildSchemaDuplicateModel(mpid spec.ModelPresetID) []byte {
	p1 := spec.ProviderName("p1")
	p2 := spec.ProviderName("p2")

	model := makeModelPreset(mpid)
	makeProv := func(name spec.ProviderName) spec.ProviderPreset {
		return spec.ProviderPreset{
			SchemaVersion:            spec.SchemaVersion,
			Name:                     name,
			DisplayName:              "X",
			APIType:                  spec.ProviderAPITypeOpenAICompatible,
			IsEnabled:                true,
			CreatedAt:                time.Now(),
			ModifiedAt:               time.Now(),
			Origin:                   "o",
			ChatCompletionPathPrefix: spec.OpenAICompatibleChatCompletionPathPrefix,
			ModelPresets:             map[spec.ModelPresetID]spec.ModelPreset{mpid: model},
		}
	}
	s := spec.PresetsSchema{
		SchemaVersion:   spec.SchemaVersion,
		DefaultProvider: p1,
		ProviderPresets: map[spec.ProviderName]spec.ProviderPreset{
			p1: makeProv(p1),
			p2: makeProv(p2),
		},
	}
	b, _ := json.Marshal(s)
	return b
}

func buildHappySchema(pn spec.ProviderName, mpid spec.ModelPresetID) []byte {
	model := makeModelPreset(mpid)
	pp := spec.ProviderPreset{
		SchemaVersion:            spec.SchemaVersion,
		Name:                     pn,
		DisplayName:              "Demo",
		APIType:                  spec.ProviderAPITypeOpenAICompatible,
		IsEnabled:                true,
		CreatedAt:                time.Now(),
		ModifiedAt:               time.Now(),
		Origin:                   "https://example.com",
		ChatCompletionPathPrefix: spec.OpenAICompatibleChatCompletionPathPrefix,
		DefaultModelPresetID:     mpid,
		ModelPresets:             map[spec.ModelPresetID]spec.ModelPreset{mpid: model},
	}
	s := spec.PresetsSchema{
		SchemaVersion:   spec.SchemaVersion,
		DefaultProvider: pp.Name,
		ProviderPresets: map[spec.ProviderName]spec.ProviderPreset{pn: pp},
	}
	b, _ := json.Marshal(s)
	return b
}

func makeModelPreset(mpid spec.ModelPresetID) spec.ModelPreset {
	return spec.ModelPreset{
		SchemaVersion: spec.SchemaVersion,
		ID:            mpid,
		Name:          spec.ModelName(mpid),
		DisplayName:   spec.ModelDisplayName("Demo Model"),
		Slug:          "demo",
		IsEnabled:     true,
		CreatedAt:     time.Now(),
		ModifiedAt:    time.Now(),
	}
}

func PrintJSON(v any) {
	p, err := json.MarshalIndent(v, "", "")
	if err == nil {
		fmt.Print("request params", "json", string(p))
	}
}
