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

	"github.com/ppipada/flexigpt-app/pkg/booloverlay"
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
			dir := tc.setupDir(t)
			bi, err := NewBuiltInPresets(dir, tc.snapshotMaxAge)

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

			prov, models, _ := bi.ListBuiltInPresets()
			if len(prov) == 0 || len(models) == 0 {
				t.Fatal("expected non-empty data")
			}
			if _, err := os.Stat(filepath.Join(dir, spec.ModelPresetsBuiltInOverlayFileName)); err != nil {
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
		setup   func(*BuiltInPresets) (spec.ProviderName, bool)
		wantErr bool
	}{
		{
			name: "toggle_existing_provider",
			setup: func(bi *BuiltInPresets) (spec.ProviderName, bool) {
				prov, _, _ := bi.ListBuiltInPresets()
				id, p := anyProvider(prov)
				return id, !p.IsEnabled
			},
		},
		{
			name: "nonexistent_provider",
			setup: func(*BuiltInPresets) (spec.ProviderName, bool) {
				return spec.ProviderName("ghost"), true
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			bi, _ := NewBuiltInPresets(dir, 0)
			pname, enabled := tc.setup(bi)
			_, err := bi.SetProviderEnabled(pname, enabled)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected: %v", err)
			}
			prov, _, _ := bi.ListBuiltInPresets()
			// PrintJSON(prov).
			// PrintJSON(models).
			if prov[pname].IsEnabled != enabled {
				t.Errorf("flag mismatch want %v", enabled)
			}
			if present, val := overlayFlagOnDisk(t, dir, "providers", string(pname)); !present ||
				val != enabled {
				t.Errorf("overlay mismatch")
			}
		})
	}
}

func TestSetModelPresetEnabled(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*BuiltInPresets) (spec.ProviderName, spec.ModelPreset, bool)
		wantErr bool
	}{
		{
			name: "toggle_existing_model",
			setup: func(bi *BuiltInPresets) (spec.ProviderName, spec.ModelPreset, bool) {
				_, models, _ := bi.ListBuiltInPresets()
				pn, _, mp := anyModel(models)
				return pn, mp, !mp.IsEnabled
			},
		},
		{
			name: "nonexistent_provider",
			setup: func(*BuiltInPresets) (spec.ProviderName, spec.ModelPreset, bool) {
				return spec.ProviderName("ghost"), spec.ModelPreset{ID: "m"}, true
			},
			wantErr: true,
		},
		{
			name: "nonexistent_model",
			setup: func(bi *BuiltInPresets) (spec.ProviderName, spec.ModelPreset, bool) {
				prov, _, _ := bi.ListBuiltInPresets()
				pn, _ := anyProvider(prov)
				return pn, spec.ModelPreset{ID: "ghost"}, true
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			bi, _ := NewBuiltInPresets(dir, 0)
			pn, mp, enabled := tc.setup(bi)
			_, err := bi.SetModelPresetEnabled(pn, mp.ID, enabled)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected: %v", err)
			}
			_, models, _ := bi.ListBuiltInPresets()
			if models[pn][mp.ID].IsEnabled != enabled {
				t.Errorf("flag mismatch want %v", enabled)
			}
			if present, val := overlayFlagOnDisk(t, dir, "models", string(mp.ID)); !present ||
				val != enabled {
				t.Errorf("overlay mismatch")
			}
		})
	}
}

func TestListBuiltInPresets(t *testing.T) {
	dir := t.TempDir()
	bi, _ := NewBuiltInPresets(dir, 0)

	t.Run("independent_copies", func(t *testing.T) {
		p1, m1, _ := bi.ListBuiltInPresets()
		p2, m2, _ := bi.ListBuiltInPresets()

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
		prov, models, _ := bi.ListBuiltInPresets()
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
	dir := t.TempDir()
	bi, _ := NewBuiltInPresets(dir, 10*time.Millisecond)

	// Fetch sample keys.
	prov, models, _ := bi.ListBuiltInPresets()
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
					_, _, _ = bi.ListBuiltInPresets()
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
				_, _ = bi.SetProviderEnabled(pname, i%2 == 0)
			}(i)
			go func(i int) {
				defer wg.Done()
				_, _ = bi.SetModelPresetEnabled(pn, mid, i%2 == 0)
			}(i)
		}
		wg.Wait()
	})
}

func TestRebuildSnapshotPresets(t *testing.T) {
	dir := t.TempDir()
	bi, _ := NewBuiltInPresets(dir, time.Hour)

	prov, _, _ := bi.ListBuiltInPresets()
	pn, p := anyProvider(prov)
	_, _ = bi.store.SetFlag(builtInProviderKey(pn), !p.IsEnabled)

	bi.mu.Lock()
	_ = bi.rebuildSnapshot()
	bi.mu.Unlock()

	prov2, _, _ := bi.ListBuiltInPresets()
	if prov2[pn].IsEnabled == p.IsEnabled {
		t.Error("rebuild did not apply overlay")
	}
}

func TestAsyncRebuildPresets(t *testing.T) {
	dir := t.TempDir()
	bi, _ := NewBuiltInPresets(dir, time.Millisecond)

	prov, _, _ := bi.ListBuiltInPresets()
	pn, p := anyProvider(prov)
	time.Sleep(2 * time.Millisecond)
	_, _ = bi.SetProviderEnabled(pn, !p.IsEnabled)
	time.Sleep(10 * time.Millisecond)

	prov2, _, _ := bi.ListBuiltInPresets()
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
	pn := spec.ProviderName("demo")
	mpid := spec.ModelPresetID("m1")

	schema := buildHappySchema(pn, mpid)
	fsys := fstest.MapFS{builtin.BuiltInModelPresetsJSON: {Data: schema}}

	dir := t.TempDir()
	bi, err := NewBuiltInPresets(dir, 0, WithModelPresetsFS(fsys, "."))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	prov, models, _ := bi.ListBuiltInPresets()
	if len(prov) != 1 || len(models) != 1 {
		t.Fatalf("want 1/1 objects, got %d/%d", len(prov), len(models))
	}

	// Toggle provider.
	p := prov[pn]
	_, _ = bi.SetProviderEnabled(pn, !p.IsEnabled)
	if present, val := overlayFlagOnDisk(t, dir, "providers", string(pn)); !present ||
		val == p.IsEnabled {
		t.Fatal("provider overlay not updated")
	}

	// Toggle model preset.
	mp := models[pn][mpid]
	_, _ = bi.SetModelPresetEnabled(pn, mpid, !mp.IsEnabled)
	if present, val := overlayFlagOnDisk(t, dir, "models", string(mpid)); !present ||
		val == mp.IsEnabled {
		t.Fatal("model overlay not updated")
	}
}

func overlayFlagOnDisk(t *testing.T, dir, group, id string) (present, val bool) {
	t.Helper()
	present = false
	val = false
	t.Helper()
	bytes, err := os.ReadFile(filepath.Join(dir, spec.ModelPresetsBuiltInOverlayFileName))
	if err != nil {
		t.Fatalf("read overlay: %v", err)
	}
	var root map[string]map[string]booloverlay.Flag
	if err := json.Unmarshal(bytes, &root); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sub, ok := root[group]; ok {
		if f, ok := sub[id]; ok {
			present, val = true, f.Enabled
			return present, val
		}
	}
	return present, val
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
	return NewBuiltInPresets(t.TempDir(), time.Hour, WithModelPresetsFS(mem, "."))
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
