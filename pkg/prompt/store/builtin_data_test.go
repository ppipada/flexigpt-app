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
	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
	"github.com/ppipada/flexigpt-app/pkg/uuidv7filename"
)

const corrupted = "corrupted"

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
				t.Helper()
				return t.TempDir()
			},
			snapshotMaxAge: time.Hour,
			wantErr:        false,
		},
		{
			name: "zero_snapshot_age_defaults_to_hour",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			snapshotMaxAge: 0,
			wantErr:        false,
		},
		{
			name: "negative_snapshot_age_defaults_to_hour",
			setupDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			snapshotMaxAge: -1,
			wantErr:        false,
		},
		{
			name: "base_dir_is_file",
			setupDir: func(t *testing.T) string {
				t.Helper()
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
				t.Helper()
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
			bundles, templates, _ := bi.ListBuiltInData()
			if len(bundles) == 0 {
				t.Error("expected at least one bundle")
			}
			if len(templates) == 0 {
				t.Error("expected at least one template")
			}

			// Verify overlay file creation.
			if _, err := os.Stat(filepath.Join(dir, spec.PromptBuiltInOverlayFileName)); err != nil {
				t.Errorf("overlay file not created: %v", err)
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
		setup   func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, bool)
		wantErr bool
	}{
		{
			name: "enable_existing_bundle",
			setup: func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, bool) {
				t.Helper()
				bundles, _, _ := bi.ListBuiltInData()
				id, bundle := anyBundle(bundles)
				return id, !bundle.IsEnabled
			},
			wantErr: false,
		},
		{
			name: "disable_existing_bundle",
			setup: func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, bool) {
				t.Helper()
				bundles, _, _ := bi.ListBuiltInData()
				id, bundle := anyBundle(bundles)
				return id, !bundle.IsEnabled
			},
			wantErr: false,
		},
		{
			name: "nonexistent_bundle",
			setup: func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, bool) {
				t.Helper()
				return bundleitemutils.BundleID("does-not-exist"), true
			},
			wantErr: true,
		},
		{
			name: "empty_bundle_id",
			setup: func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, bool) {
				t.Helper()
				return bundleitemutils.BundleID(""), true
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

			_, err = bi.SetBundleEnabled(bundleID, enabled)

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
			bundles, _, _ := bi.ListBuiltInData()
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
		setup   func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, spec.PromptTemplate, bool)
		wantErr bool
	}{
		{
			name: "enable_existing_template",
			setup: func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, spec.PromptTemplate, bool) {
				t.Helper()
				_, templates, _ := bi.ListBuiltInData()
				bid, _, tpl := anyTemplate(templates)
				return bid, tpl, !tpl.IsEnabled
			},
			wantErr: false,
		},
		{
			name: "disable_existing_template",
			setup: func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, spec.PromptTemplate, bool) {
				t.Helper()
				_, templates, _ := bi.ListBuiltInData()
				bid, _, tpl := anyTemplate(templates)
				return bid, tpl, !tpl.IsEnabled
			},
			wantErr: false,
		},
		{
			name: "nonexistent_bundle",
			setup: func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, spec.PromptTemplate, bool) {
				t.Helper()
				return bundleitemutils.BundleID(
						"does-not-exist",
					), spec.PromptTemplate{
						ID:      bundleitemutils.ItemID("template"),
						Slug:    "template",
						Version: "a",
					}, true
			},
			wantErr: true,
		},
		{
			name: "nonexistent_template",
			setup: func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, spec.PromptTemplate, bool) {
				t.Helper()
				bundles, _, _ := bi.ListBuiltInData()
				bid, _ := anyBundle(bundles)
				return bid, spec.PromptTemplate{
						ID:      bundleitemutils.ItemID("does-not-exist"),
						Slug:    "does-not-exist",
						Version: "a",
					},
					true
			},
			wantErr: true,
		},
		{
			name: "empty_template_id",
			setup: func(t *testing.T, bi *BuiltInData) (bundleitemutils.BundleID, spec.PromptTemplate, bool) {
				t.Helper()
				bundles, _, _ := bi.ListBuiltInData()
				bid, _ := anyBundle(bundles)
				return bid, spec.PromptTemplate{
					ID:      "",
					Slug:    "",
					Version: "a",
				}, true
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

			bundleID, template, enabled := tt.setup(t, bi)

			_, err = bi.SetTemplateEnabled(bundleID, template.Slug, template.Version, enabled)

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
			_, templates, _ := bi.ListBuiltInData()
			if got := templates[bundleID][template.ID].IsEnabled; got != enabled {
				t.Errorf("template enabled state = %v, want %v", got, enabled)
			}

			// Verify persistence to disk.
			if present, val := overlayOnDisk(t, dir, "templates", string(template.ID)); !present ||
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
		b1, t1, _ := bi.ListBuiltInData()
		b2, t2, _ := bi.ListBuiltInData()

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
		bundles, templates, _ := bi.ListBuiltInData()

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
		bundles, templates, _ := bi.ListBuiltInData()

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

	bundles, templates, _ := bi.ListBuiltInData()
	if len(bundles) == 0 || len(templates) == 0 {
		t.Skip("no test data available")
	}

	bundleID, _ := anyBundle(bundles)
	templateBundleID, _, template := anyTemplate(templates)

	t.Run("concurrent_reads", func(t *testing.T) {
		var wg sync.WaitGroup
		const numReaders = 10

		for range numReaders {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 100 {
					bundles, templates, _ := bi.ListBuiltInData()
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
				if _, err := bi.SetBundleEnabled(bundleID, enabled); err != nil {
					t.Errorf("SetBundleEnabled failed: %v", err)
				}
			}(i)

			go func(i int) {
				defer wg.Done()
				enabled := i%2 == 0
				if _, err := bi.SetTemplateEnabled(templateBundleID, template.Slug, template.Version, enabled); err != nil {
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
						_, _, _ = bi.ListBuiltInData()
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
					if _, e := bi.SetBundleEnabled(bundleID, enabled); e != nil {
						t.Errorf("SetBundleEnabled failed: %v", e)
					}
					if _, e := bi.SetTemplateEnabled(templateBundleID, template.Slug, template.Version, enabled); e != nil {
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

	bundles, _, _ := bi.ListBuiltInData()
	bundleID, bundle := anyBundle(bundles)

	// Change the overlay directly.
	_, err = bi.store.SetFlag(BuiltInBundleID(bundleID), !bundle.IsEnabled)
	if err != nil {
		t.Fatalf("failed to set overlay: %v", err)
	}

	// Force rebuild.
	if err := bi.rebuildSnapshot(); err != nil {
		t.Fatalf("rebuildSnapshot failed: %v", err)
	}

	// Verify the change is reflected.
	newBundles, _, _ := bi.ListBuiltInData()
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

	bundles, _, _ := bi.ListBuiltInData()
	bundleID, bundle := anyBundle(bundles)

	// Trigger async rebuild by making the snapshot stale.
	time.Sleep(2 * time.Millisecond)

	// Make a change that should trigger rebuild.
	if _, err := bi.SetBundleEnabled(bundleID, !bundle.IsEnabled); err != nil {
		t.Fatalf("SetBundleEnabled failed: %v", err)
	}

	// Give async rebuild time to complete.
	time.Sleep(10 * time.Millisecond)

	// Verify the change is eventually reflected.
	newBundles, _, _ := bi.ListBuiltInData()
	if newBundles[bundleID].IsEnabled == bundle.IsEnabled {
		t.Error("async rebuild did not apply changes")
	}
}

func Test_NewBuiltInData_SyntheticFS_Errors(t *testing.T) {
	bundleID := newUUID(t)
	slug := "demo"

	t.Run("missing_bundles_json", func(t *testing.T) {
		_, err := newFromFS(t, fstest.MapFS{})
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("want fs.ErrNotExist, got %v", err)
		}
	})

	t.Run("invalid_bundles_json", func(t *testing.T) {
		fsys := fstest.MapFS{
			builtin.BuiltInPromptBundlesJSON: {Data: []byte("{ oops ]")},
		}
		_, err := newFromFS(t, fsys)
		if err == nil || !strings.Contains(err.Error(), "invalid") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("no_bundles_in_manifest", func(t *testing.T) {
		empty, _ := json.Marshal(
			spec.AllBundles{Bundles: map[bundleitemutils.BundleID]spec.PromptBundle{}},
		)
		fsys := fstest.MapFS{builtin.BuiltInPromptBundlesJSON: {Data: empty}}
		_, err := newFromFS(t, fsys)
		if err == nil || !strings.Contains(err.Error(), "contains no bundles") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("bundle_has_no_templates", func(t *testing.T) {
		dir := fmt.Sprintf("%s_%s", bundleID, slug)
		fsys := fstest.MapFS{
			builtin.BuiltInPromptBundlesJSON: {Data: buildManifest(bundleID, slug)},
			dir:                              &fstest.MapFile{Mode: fs.ModeDir},
		}
		_, err := newFromFS(t, fsys)
		if err == nil || !strings.Contains(err.Error(), "has no templates") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("bundle_dir_not_in_manifest", func(t *testing.T) {
		ghostID := newUUID(t)
		ghostSlug := "ghost"
		dir := fmt.Sprintf("%s_%s", ghostID, ghostSlug)

		fn, raw, _ := buildTemplate(t, ghostSlug, "v1")
		fsys := fstest.MapFS{
			builtin.BuiltInPromptBundlesJSON: {Data: buildManifest(bundleID, slug)},
			dir:                              &fstest.MapFile{Mode: fs.ModeDir},
			dir + "/" + fn:                   {Data: raw},
		}
		_, err := newFromFS(t, fsys)
		if err == nil ||
			!strings.Contains(
				err.Error(),
				"not in "+builtin.BuiltInPromptBundlesJSON,
			) {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("slug_mismatch_between_dir_and_manifest", func(t *testing.T) {
		// Dir slug = wrong, manifest slug = demo.
		dir := fmt.Sprintf("%s_%s", bundleID, "wrong")

		fn, raw, _ := buildTemplate(t, "wrong", "v1")
		fsys := fstest.MapFS{
			builtin.BuiltInPromptBundlesJSON: {Data: buildManifest(bundleID, slug)},
			dir:                              &fstest.MapFile{Mode: fs.ModeDir},
			dir + "/" + fn:                   {Data: raw},
		}
		_, err := newFromFS(t, fsys)
		if err == nil || !strings.Contains(err.Error(), "dir slug") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("filename_slug_version_mismatch", func(t *testing.T) {
		dir := fmt.Sprintf("%s_%s", bundleID, slug)

		// Template JSON says slug=demo,ver=v1.
		_, raw, _ := buildTemplate(t, slug, "v1")
		// But store it under file name demo_v2.json.
		fsys := fstest.MapFS{
			builtin.BuiltInPromptBundlesJSON: {Data: buildManifest(bundleID, slug)},
			dir:                              &fstest.MapFile{Mode: fs.ModeDir},
			dir + "/demo_v2.json":            {Data: raw},
		}
		_, err := newFromFS(t, fsys)
		if err == nil || !strings.Contains(err.Error(), "filename") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func Test_NewBuiltInData_SyntheticFS_HappyAndCRUD(t *testing.T) {
	bid := newUUID(t)
	slug := "demo"

	// Build valid dir template.
	dir := fmt.Sprintf("%s_%s", bid, slug)
	fn, rawTpl, _ := buildTemplate(t, slug, "v1")

	mem := fstest.MapFS{
		builtin.BuiltInPromptBundlesJSON: {Data: buildManifest(bid, slug)},
		dir:                              &fstest.MapFile{Mode: fs.ModeDir},
		dir + "/" + fn:                   {Data: rawTpl},
	}

	tmp := t.TempDir()
	bi, err := NewBuiltInData(tmp, 0, WithBundlesFS(mem, "."))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bundles, tpls, _ := bi.ListBuiltInData()
	if len(bundles) != 1 || len(tpls) != 1 {
		t.Fatalf("want 1/1 objects, got %d/%d", len(bundles), len(tpls))
	}

	bundleID, bundle := anyBundle(bundles)
	if _, err := bi.SetBundleEnabled(bundleID, !bundle.IsEnabled); err != nil {
		t.Fatalf("SetBundleEnabled: %v", err)
	}
	if present, val := overlayOnDisk(t, tmp, "bundles", string(bundleID)); !present ||
		val == bundle.IsEnabled {
		t.Fatalf("bundle overlay not updated")
	}

	_, tmplID, tpl := anyTemplate(tpls)
	if _, err := bi.SetTemplateEnabled(bundleID, tpl.Slug, tpl.Version, !tpl.IsEnabled); err != nil {
		t.Fatalf("SetTemplateEnabled: %v", err)
	}
	if present, val := overlayOnDisk(t, tmp, "templates", string(tmplID)); !present ||
		val == tpl.IsEnabled {
		t.Fatalf("template overlay not updated")
	}
}

func overlayOnDisk(t *testing.T, dir, group, id string) (present, value bool) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, spec.PromptBuiltInOverlayFileName))
	if err != nil {
		t.Fatalf("cannot read overlay file: %v", err)
	}
	var root map[string]map[string]booloverlay.Flag
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("invalid overlay JSON: %v", err)
	}
	if sub, ok := root[group]; ok {
		v, ok := sub[id]
		return ok, v.Enabled
	}
	return false, false
}

func anyTemplate(
	m map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate,
) (bundleitemutils.BundleID, bundleitemutils.ItemID, spec.PromptTemplate) {
	for bid, tm := range m {
		for tid, tpl := range tm {
			return bid, tid, tpl
		}
	}
	return "", "", spec.PromptTemplate{}
}

func anyBundle(
	m map[bundleitemutils.BundleID]spec.PromptBundle,
) (bundleitemutils.BundleID, spec.PromptBundle) {
	for id, b := range m {
		return id, b
	}
	return "", spec.PromptBundle{}
}

// buildManifest returns the JSON for a manifest that holds exactly one bundle.
func buildManifest(bundleID, slug string) []byte {
	manifest := spec.AllBundles{
		Bundles: map[bundleitemutils.BundleID]spec.PromptBundle{
			bundleitemutils.BundleID(bundleID): {
				ID:        bundleitemutils.BundleID(bundleID),
				Slug:      bundleitemutils.BundleSlug(slug),
				IsEnabled: true,
			},
		},
	}
	b, _ := json.Marshal(manifest)
	return b
}

// buildTemplate returns filename (slug_version.json), raw JSON and the template ID.
func buildTemplate(t *testing.T, slug, ver string) (fileName string, raw []byte, tplID string) {
	t.Helper()
	tplID = newUUID(t)
	tpl := spec.PromptTemplate{
		SchemaVersion: spec.SchemaVersion,
		ID:            bundleitemutils.ItemID(tplID),
		Slug:          bundleitemutils.ItemSlug(slug),
		Version:       bundleitemutils.ItemVersion(ver),
		DisplayName:   slug,
		Blocks: []spec.MessageBlock{
			{
				ID:      spec.MessageBlockID("a"),
				Role:    spec.System,
				Content: "hey",
			},
		},
		IsEnabled:  true,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}
	raw, _ = json.Marshal(tpl) // cannot fail
	fileName = fmt.Sprintf("%s_%s.json", slug, ver)
	return fileName, raw, tplID
}

// Constructor helper that injects the given fs.FS.
func newFromFS(t *testing.T, mem fs.FS) (*BuiltInData, error) {
	t.Helper()
	return NewBuiltInData(t.TempDir(), time.Hour, WithBundlesFS(mem, "."))
}

// newUUID returns a v7-UUID as string or fails the test.
func newUUID(t *testing.T) string {
	t.Helper()
	u, err := uuidv7filename.NewUUID()
	if err != nil {
		t.Fatalf("uuidv7: %v", err)
	}
	return u
}
