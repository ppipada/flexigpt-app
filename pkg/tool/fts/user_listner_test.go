package fts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
	"github.com/ppipada/flexigpt-app/pkg/tool/spec"
)

func TestToolsFTSListener_Integration(t *testing.T) {
	tmp := mustTempDir(t)
	defer mustRemoveAll(t, tmp)

	engine, err := ftsengine.NewEngine(ftsengine.Config{
		BaseDir:    tmp,
		DBFileName: "fts.db",
		Table:      sqliteDBTableName,
		Columns: []ftsengine.Column{
			{Name: "slug", Weight: 1},
			{Name: "displayName", Weight: 2},
			{Name: "desc", Weight: 3},
			{Name: "parameters", Weight: 4},
			{Name: "tags", Weight: 5},
			{Name: "enabled", Unindexed: true},
			{Name: "bundleID", Unindexed: true},
			{Name: "mtime", Unindexed: true},
		},
	})
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	listener := NewUserToolsFTSListener(engine)

	// Build bundle dir + tool file.
	bundleID := "b1"
	bundleSlug := "bundle-slug"
	bundleDir := filepath.Join(tmp, bundleDirName(bundleID, bundleSlug))
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	toolSlug := "mytool"
	toolVersion := "v1"
	toolFile := filepath.Join(
		bundleDir,
		toolSlug+"."+toolVersion+"."+bundleitemutils.ItemFileExtension,
	)

	tool := spec.ToolSpec{
		ID:          "id1",
		DisplayName: "Test Tool",
		Slug:        bundleitemutils.ItemSlug(toolSlug),
		IsEnabled:   true,
		Description: "tool desc",
		Tags:        []string{"foo", "bar"},
		Parameters: []spec.ToolParameter{
			{
				Name:        "param1",
				Type:        spec.ParamString,
				Description: "first parameter",
				Required:    true,
			},
		},
		Version:    bundleitemutils.ItemVersion(toolVersion),
		CreatedAt:  time.Now().UTC(),
		ModifiedAt: time.Now().UTC(),
	}

	mustWriteJSON(t, toolFile, tool)

	ev := filestore.Event{
		Op:   filestore.OpSetFile,
		File: toolFile,
		Data: func() map[string]any {
			var m map[string]any
			_ = json.Unmarshal(mustReadFile(t, toolFile), &m)
			return m
		}(),
	}
	listener(ev)

	time.Sleep(100 * time.Millisecond)

	ctx := t.Context()
	hits, next, err := engine.Search(ctx, "param1", "", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if !strings.Contains(hits[0].ID, toolFile) {
		t.Errorf("unexpected hit ID: %s", hits[0].ID)
	}
	if next != "" {
		t.Errorf("unexpected next token: %q", next)
	}

	// Delete event.
	ev = filestore.Event{
		Op:   filestore.OpDeleteFile,
		File: toolFile,
	}
	listener(ev)

	time.Sleep(100 * time.Millisecond)

	hits, _, err = engine.Search(ctx, "param1", "", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 0 {
		t.Errorf("expected 0 hits, got %d", len(hits))
	}
}

func TestToolsExtractFTS_AllFields(t *testing.T) {
	now := time.Now().UTC()
	m := map[string]any{
		"slug":        "sluggy",
		"displayName": "disp",
		"description": "desc",
		"tags":        []any{"a", "b"},
		"parameters": []any{
			map[string]any{"name": "p1", "description": "d1"},
			map[string]any{"name": "p2"},
		},
		"isEnabled":  true,
		"modifiedAt": now.Format(time.RFC3339Nano),
	}
	fakePath := "/tmp/bundles__b1__sluggy/sluggy.v1.tool.json"
	vals := extractFTS(fakePath, m).ToMap()

	if vals["slug"] != "sluggy" {
		t.Errorf("slug: %q", vals["slug"])
	}
	if vals["displayName"] != "disp" {
		t.Errorf("displayName: %q", vals["displayName"])
	}
	if vals["desc"] != "desc" {
		t.Errorf("desc: %q", vals["desc"])
	}
	if !strings.Contains(vals["parameters"], "p1") || !strings.Contains(vals["parameters"], "p2") {
		t.Errorf("parameters: %q", vals["parameters"])
	}
	if !strings.Contains(vals["tags"], "a") || !strings.Contains(vals["tags"], "b") {
		t.Errorf("tags: %q", vals["tags"])
	}
	if vals["enabled"] != "true" {
		t.Errorf("enabled: %q", vals["enabled"])
	}
	if vals["mtime"] != now.Format(time.RFC3339Nano) {
		t.Errorf("mtime: %q", vals["mtime"])
	}
}

func TestToolsExtractFTS_EmptyAndEdgeCases(t *testing.T) {
	m := map[string]any{
		"slug":      "",
		"isEnabled": false,
	}
	vals := extractFTS("/tmp/bundles__bid__slug/slug.v1.tool.json", m).ToMap()
	if vals["enabled"] != "false" {
		t.Errorf("enabled: %q", vals["enabled"])
	}
	if vals["tags"] != "" {
		t.Errorf("tags: %q", vals["tags"])
	}
	if vals["parameters"] != "" {
		t.Errorf("parameters: %q", vals["parameters"])
	}
}

func TestToolsProcessFTSSync_SkipNonToolFiles(t *testing.T) {
	ctx := t.Context()
	tmp := mustTempDir(t)
	defer mustRemoveAll(t, tmp)

	other := filepath.Join(tmp, "some.txt")
	_ = os.WriteFile(other, []byte("hi"), 0o600)

	dec, err := processFTSSync(ctx, tmp, other, func(string) string { return "" })
	if err != nil {
		t.Fatalf("processFTSSync: %v", err)
	}
	if !dec.Skip {
		t.Errorf("expected skip")
	}
}

func TestToolsProcessFTSSync_Unchanged(t *testing.T) {
	ctx := t.Context()
	tmp := mustTempDir(t)
	defer mustRemoveAll(t, tmp)

	bundleID := "bid"
	bundleSlug := "slug"
	bundleDir := filepath.Join(tmp, bundleDirName(bundleID, bundleSlug))
	_ = os.MkdirAll(bundleDir, 0o755)
	toolFile := filepath.Join(bundleDir, "mytool.v1.tool.json")

	tool := spec.ToolSpec{
		ID:          "id1",
		DisplayName: "Tool",
		Slug:        "mytool",
		IsEnabled:   true,
		Version:     "v1",
		CreatedAt:   time.Now().UTC(),
		ModifiedAt:  time.Now().UTC(),
	}
	mustWriteJSON(t, toolFile, tool)

	cmp := fileMTime(toolFile)

	dec, err := processFTSSync(ctx, tmp, toolFile, func(string) string { return cmp })
	if err != nil {
		t.Fatalf("processFTSSync: %v", err)
	}
	if !dec.Unchanged {
		t.Errorf("expected unchanged")
	}
}

func TestToolsProcessFTSSync_BadJSON(t *testing.T) {
	ctx := t.Context()
	tmp := mustTempDir(t)
	defer mustRemoveAll(t, tmp)

	bundleDir := filepath.Join(tmp, bundleDirName("bid", "slug"))
	_ = os.MkdirAll(bundleDir, 0o755)
	toolFile := filepath.Join(bundleDir, "bad.v1.tool.json")
	_ = os.WriteFile(toolFile, []byte("{notjson"), 0o600)

	dec, err := processFTSSync(ctx, tmp, toolFile, func(string) string { return "" })
	if err != nil {
		t.Fatalf("processFTSSync: %v", err)
	}
	if !dec.Skip {
		t.Errorf("expected skip")
	}
}

func bundleDirName(id, slug string) string {
	return "bundles__" + id + "__" + slug
}

func mustRemoveAll(t *testing.T, dir string) {
	t.Helper()
	_ = os.RemoveAll(dir)
}

func mustWriteJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return b
}

func mustTempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}
