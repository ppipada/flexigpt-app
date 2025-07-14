package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

const corrupted = "corrupted"

// Test helpers.
func anyBundle(m map[spec.BundleID]spec.PromptBundle) (spec.BundleID, spec.PromptBundle) {
	for id, b := range m {
		return id, b
	}
	return "", spec.PromptBundle{}
}

func anyTemplate(
	m map[spec.BundleID]map[spec.TemplateID]spec.PromptTemplate,
) (spec.BundleID, spec.TemplateID, spec.PromptTemplate) {
	for bid, tm := range m {
		for tid, tpl := range tm {
			return bid, tid, tpl
		}
	}
	return "", "", spec.PromptTemplate{}
}

func overlayOnDisk(t *testing.T, dir, group, id string) (present, value bool) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, overlayJSON))
	if err != nil {
		t.Fatalf("cannot read overlay file: %v", err)
	}
	var root map[string]map[string]bool
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("invalid overlay JSON: %v", err)
	}
	if sub, ok := root[group]; ok {
		v, ok := sub[id]
		return ok, v
	}
	return false, false
}

func TestNewBuiltInData(t *testing.T) {
	tests := []struct {
		name           string
		setupDir       func(t *testing.T) string
		snapshotMaxAge time.Duration
		wantErr        bool
		errContains    string
	}{
		{
			name: "happy_path",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
			snapshotMaxAge: time.Hour,
			wantErr:        false,
		},
		{
			name: "zero_snapshot_age_defaults_to_hour",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
			snapshotMaxAge: 0,
			wantErr:        false,
		},
		{
			name: "negative_snapshot_age_defaults_to_hour",
			setupDir: func(t *testing.T) string {
				return t.TempDir()
			},
			snapshotMaxAge: -1,
			wantErr:        false,
		},
		{
			name: "base_dir_is_file",
			setupDir: func(t *testing.T) string {
				tmp := t.TempDir()
				file := filepath.Join(tmp, "not-a-dir")
				if err := os.WriteFile(file, []byte("dummy"), 0o600); err != nil {
					t.Fatalf("setup: %v", err)
				}
				return file
			},
			snapshotMaxAge: time.Hour,
			wantErr:        true,
		},
		{
			name: "readonly_directory",
			setupDir: func(t *testing.T) string {
				tmp := t.TempDir()
				dir := filepath.Join(tmp, "readonly")
				if err := os.MkdirAll(dir, 0o444); err != nil {
					t.Fatalf("setup: %v", err)
				}
				return dir
			},
			snapshotMaxAge: time.Hour,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)

			bi, err := NewBuiltInData(dir, tt.snapshotMaxAge)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify basic functionality.
			bundles, templates := bi.List()
			if len(bundles) == 0 {
				t.Error("expected at least one bundle")
			}
			if len(templates) == 0 {
				t.Error("expected at least one template")
			}

			// Verify overlay file creation.
			if _, err := os.Stat(filepath.Join(dir, overlayJSON)); err != nil {
				t.Errorf("overlay.json not created: %v", err)
			}

			// Verify snapshot max age handling.
			expectedAge := tt.snapshotMaxAge
			if expectedAge <= 0 {
				expectedAge = time.Hour
			}
			if bi.snapshotMaxAge != expectedAge {
				t.Errorf("snapshotMaxAge = %v, want %v", bi.snapshotMaxAge, expectedAge)
			}

			// Verify all bundles are marked as built-in.
			for id, bundle := range bundles {
				if !bundle.IsBuiltIn {
					t.Errorf("bundle %s should be marked as built-in", id)
				}
			}

			// Verify all templates are marked as built-in.
			for bid, tmap := range templates {
				for tid, template := range tmap {
					if !template.IsBuiltIn {
						t.Errorf("template %s/%s should be marked as built-in", bid, tid)
					}
				}
			}
		})
	}
}

func TestSetBundleEnabled(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, bi *BuiltInData) (spec.BundleID, bool)
		wantErr bool
	}{
		{
			name: "enable_existing_bundle",
			setup: func(t *testing.T, bi *BuiltInData) (spec.BundleID, bool) {
				bundles, _ := bi.List()
				id, bundle := anyBundle(bundles)
				return id, !bundle.IsEnabled
			},
			wantErr: false,
		},
		{
			name: "disable_existing_bundle",
			setup: func(t *testing.T, bi *BuiltInData) (spec.BundleID, bool) {
				bundles, _ := bi.List()
				id, bundle := anyBundle(bundles)
				return id, !bundle.IsEnabled
			},
			wantErr: false,
		},
		{
			name: "nonexistent_bundle",
			setup: func(t *testing.T, bi *BuiltInData) (spec.BundleID, bool) {
				return spec.BundleID("does-not-exist"), true
			},
			wantErr: true,
		},
		{
			name: "empty_bundle_id",
			setup: func(t *testing.T, bi *BuiltInData) (spec.BundleID, bool) {
				return spec.BundleID(""), true
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			bi, err := NewBuiltInData(dir, 0)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			bundleID, enabled := tt.setup(t, bi)

			err = bi.SetBundleEnabled(bundleID, enabled)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the change is reflected in the snapshot
			// Note: Due to async rebuild, we might need to wait or force rebuild.
			bundles, _ := bi.List()
			if got := bundles[bundleID].IsEnabled; got != enabled {
				t.Errorf("bundle enabled state = %v, want %v", got, enabled)
			}

			// Verify persistence to disk.
			if present, val := overlayOnDisk(t, dir, "bundles", string(bundleID)); !present ||
				val != enabled {
				t.Errorf(
					"overlay: present=%v val=%v, want present=true val=%v",
					present,
					val,
					enabled,
				)
			}
		})
	}
}

func TestSetTemplateEnabled(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, bi *BuiltInData) (spec.BundleID, spec.TemplateID, bool)
		wantErr bool
	}{
		{
			name: "enable_existing_template",
			setup: func(t *testing.T, bi *BuiltInData) (spec.BundleID, spec.TemplateID, bool) {
				_, templates := bi.List()
				bid, tid, tpl := anyTemplate(templates)
				return bid, tid, !tpl.IsEnabled
			},
			wantErr: false,
		},
		{
			name: "disable_existing_template",
			setup: func(t *testing.T, bi *BuiltInData) (spec.BundleID, spec.TemplateID, bool) {
				_, templates := bi.List()
				bid, tid, tpl := anyTemplate(templates)
				return bid, tid, !tpl.IsEnabled
			},
			wantErr: false,
		},
		{
			name: "nonexistent_bundle",
			setup: func(t *testing.T, bi *BuiltInData) (spec.BundleID, spec.TemplateID, bool) {
				return spec.BundleID("does-not-exist"), spec.TemplateID("template"), true
			},
			wantErr: true,
		},
		{
			name: "nonexistent_template",
			setup: func(t *testing.T, bi *BuiltInData) (spec.BundleID, spec.TemplateID, bool) {
				bundles, _ := bi.List()
				bid, _ := anyBundle(bundles)
				return bid, spec.TemplateID("does-not-exist"), true
			},
			wantErr: true,
		},
		{
			name: "empty_template_id",
			setup: func(t *testing.T, bi *BuiltInData) (spec.BundleID, spec.TemplateID, bool) {
				bundles, _ := bi.List()
				bid, _ := anyBundle(bundles)
				return bid, spec.TemplateID(""), true
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			bi, err := NewBuiltInData(dir, 0)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			bundleID, templateID, enabled := tt.setup(t, bi)

			err = bi.SetTemplateEnabled(bundleID, templateID, enabled)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the change is reflected in the snapshot.
			_, templates := bi.List()
			if got := templates[bundleID][templateID].IsEnabled; got != enabled {
				t.Errorf("template enabled state = %v, want %v", got, enabled)
			}

			// Verify persistence to disk.
			if present, val := overlayOnDisk(t, dir, "templates", string(templateID)); !present ||
				val != enabled {
				t.Errorf(
					"overlay: present=%v val=%v, want present=true val=%v",
					present,
					val,
					enabled,
				)
			}
		})
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	bi, err := NewBuiltInData(dir, 0)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("returns_independent_copies", func(t *testing.T) {
		b1, t1 := bi.List()
		b2, t2 := bi.List()

		// Mutate first copy.
		for id := range b1 {
			b := b1[id]
			b.Slug = corrupted
			b1[id] = b
			break
		}
		for bid := range t1 {
			for tid := range t1[bid] {
				tpl := t1[bid][tid]
				tpl.Slug = corrupted
				t1[bid][tid] = tpl
				break
			}
			break
		}

		// Verify second copy is unaffected.
		for id, b := range b2 {
			if b.Slug == corrupted {
				t.Errorf("bundle %s shares memory with first copy", id)
			}
		}
		for bid, tm := range t2 {
			for tid, tpl := range tm {
				if tpl.Slug == corrupted {
					t.Errorf("template %s/%s shares memory with first copy", bid, tid)
				}
			}
		}
	})

	t.Run("returns_consistent_data", func(t *testing.T) {
		bundles, templates := bi.List()

		// Verify all bundles have corresponding template maps.
		for bundleID := range bundles {
			if _, exists := templates[bundleID]; !exists {
				t.Errorf("bundle %s has no corresponding template map", bundleID)
			}
		}

		// Verify all template maps have corresponding bundles.
		for bundleID := range templates {
			if _, exists := bundles[bundleID]; !exists {
				t.Errorf("template map for bundle %s has no corresponding bundle", bundleID)
			}
		}
	})

	t.Run("all_items_marked_builtin", func(t *testing.T) {
		bundles, templates := bi.List()

		for id, bundle := range bundles {
			if !bundle.IsBuiltIn {
				t.Errorf("bundle %s not marked as built-in", id)
			}
		}

		for bid, tmap := range templates {
			for tid, template := range tmap {
				if !template.IsBuiltIn {
					t.Errorf("template %s/%s not marked as built-in", bid, tid)
				}
			}
		}
	})
}

func TestConcurrency(t *testing.T) {
	dir := t.TempDir()
	bi, err := NewBuiltInData(dir, time.Millisecond*10) // Short rebuild interval
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	bundles, templates := bi.List()
	if len(bundles) == 0 || len(templates) == 0 {
		t.Skip("no test data available")
	}

	bundleID, _ := anyBundle(bundles)
	templateBundleID, templateID, _ := anyTemplate(templates)

	t.Run("concurrent_reads", func(t *testing.T) {
		var wg sync.WaitGroup
		const numReaders = 10

		for range numReaders {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 100 {
					bundles, templates := bi.List()
					if len(bundles) == 0 || len(templates) == 0 {
						t.Errorf("got empty results during concurrent read")
						return
					}
				}
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent_writes", func(t *testing.T) {
		var wg sync.WaitGroup
		const numWriters = 5

		for i := range numWriters {
			wg.Add(2)
			go func(i int) {
				defer wg.Done()
				enabled := i%2 == 0
				if err := bi.SetBundleEnabled(bundleID, enabled); err != nil {
					t.Errorf("SetBundleEnabled failed: %v", err)
				}
			}(i)

			go func(i int) {
				defer wg.Done()
				enabled := i%2 == 0
				if err := bi.SetTemplateEnabled(templateBundleID, templateID, enabled); err != nil {
					t.Errorf("SetTemplateEnabled failed: %v", err)
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("concurrent_read_write", func(t *testing.T) {
		var wg sync.WaitGroup
		done := make(chan struct{})

		// Start readers.
		for range 5 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					default:
						bi.List()
					}
				}
			}()
		}

		// Start writers.
		for i := range 3 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for j := range 10 {
					enabled := (i+j)%2 == 0
					if e := bi.SetBundleEnabled(bundleID, enabled); e != nil {
						t.Errorf("SetBundleEnabled failed: %v", e)
					}
					if e := bi.SetTemplateEnabled(templateBundleID, templateID, enabled); e != nil {
						t.Errorf("SetTemplateEnabled failed: %v", e)
					}
				}
			}(i)
		}

		// Let it run for a bit.
		time.Sleep(100 * time.Millisecond)
		close(done)
		wg.Wait()
	})
}

func TestRebuildSnapshot(t *testing.T) {
	dir := t.TempDir()
	bi, err := NewBuiltInData(dir, time.Hour) // Long interval to control rebuilds
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	bundles, _ := bi.List()
	bundleID, bundle := anyBundle(bundles)

	// Change the overlay directly.
	if err := bi.store.SetEnabled(BuiltInBundleID(bundleID), !bundle.IsEnabled); err != nil {
		t.Fatalf("failed to set overlay: %v", err)
	}

	// Force rebuild.
	if err := bi.rebuildSnapshot(); err != nil {
		t.Fatalf("rebuildSnapshot failed: %v", err)
	}

	// Verify the change is reflected.
	newBundles, _ := bi.List()
	if newBundles[bundleID].IsEnabled == bundle.IsEnabled {
		t.Error("rebuild did not apply overlay changes")
	}
}

func TestAsyncRebuild(t *testing.T) {
	dir := t.TempDir()
	bi, err := NewBuiltInData(dir, time.Millisecond) // Very short interval
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	bundles, _ := bi.List()
	bundleID, bundle := anyBundle(bundles)

	// Trigger async rebuild by making the snapshot stale.
	time.Sleep(2 * time.Millisecond)

	// Make a change that should trigger rebuild.
	if err := bi.SetBundleEnabled(bundleID, !bundle.IsEnabled); err != nil {
		t.Fatalf("SetBundleEnabled failed: %v", err)
	}

	// Give async rebuild time to complete.
	time.Sleep(10 * time.Millisecond)

	// Verify the change is eventually reflected.
	newBundles, _ := bi.List()
	if newBundles[bundleID].IsEnabled == bundle.IsEnabled {
		t.Error("async rebuild did not apply changes")
	}
}
