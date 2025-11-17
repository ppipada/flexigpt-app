package fts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
	"github.com/ppipada/mapstore-go"
	"github.com/ppipada/mapstore-go/ftsengine"
)

func TestFTSListener_Integration(t *testing.T) {
	tmp := mustTempDir(t)
	defer mustRemoveAll(t, tmp)

	// Setup FTS engine.
	engine, err := ftsengine.NewEngine(ftsengine.Config{
		BaseDir:    tmp,
		DBFileName: "fts.db",
		Table:      sqliteDBTableName,
		Columns: []ftsengine.Column{
			{Name: "slug", Weight: 1},
			{Name: "displayName", Weight: 2},
			{Name: "desc", Weight: 3},
			{Name: "messages", Weight: 4},
			{Name: "tags", Weight: 5},
			{Name: "enabled", Unindexed: true},
			{Name: "bundleID", Unindexed: true},
			{Name: "mtime", Unindexed: true},
		},
	})
	if err != nil {
		t.Fatalf("ftsengine.NewEngine: %v", err)
	}

	listener := NewUserPromptsFTSListener(engine)

	// Prepare a bundle dir and template file.
	bundleID := "b1"
	bundleSlug := "bundle-slug"
	bundleDir := filepath.Join(tmp, bundleDirName(bundleID, bundleSlug))
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	tplSlug := "myslug"
	tplVersion := "v1"
	tplFile := filepath.Join(
		bundleDir,
		tplSlug+"."+tplVersion+"."+bundleitemutils.ItemFileExtension,
	)

	// Compose a prompt template.
	tpl := spec.PromptTemplate{
		ID:          "id1",
		DisplayName: "Test Template",
		Slug:        bundleitemutils.ItemSlug(tplSlug),
		IsEnabled:   true,
		Description: "desc here",
		Tags:        []string{"foo", "bar"},
		Blocks: []spec.MessageBlock{
			{ID: "1", Role: spec.User, Content: "Hello"},
			{ID: "2", Role: spec.Assistant, Content: "World"},
		},
		Version:    bundleitemutils.ItemVersion(tplVersion),
		CreatedAt:  time.Now().UTC(),
		ModifiedAt: time.Now().UTC(),
	}

	// Write the template file.
	mustWriteJSON(t, tplFile, tpl)

	// Simulate a SetFile event.
	ev := mapstore.FileEvent{
		Op:   mapstore.OpSetFile,
		File: tplFile,
		Data: func() map[string]any {
			var m map[string]any
			_ = json.Unmarshal(mustReadFile(t, tplFile), &m)
			return m
		}(),
	}
	listener(ev)

	// Wait for FTS to index (sqlite is synchronous, but just in case).
	time.Sleep(100 * time.Millisecond)

	// Search for "Hello" (should match).
	ctx := t.Context()
	hits, next, err := engine.Search(ctx, "Hello", "", 10)
	if err != nil {
		t.Fatalf("fts search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if !strings.Contains(hits[0].ID, tplFile) {
		t.Errorf("hit ID mismatch: %v", hits[0].ID)
	}
	if next != "" {
		t.Errorf("unexpected next token: %v", next)
	}

	// Simulate a DeleteFile event.
	ev = mapstore.FileEvent{
		Op:   mapstore.OpDeleteFile,
		File: tplFile,
	}
	listener(ev)

	// Wait for FTS to update.
	time.Sleep(100 * time.Millisecond)

	// Search again, should be gone.
	hits, _, err = engine.Search(ctx, "Hello", "", 10)
	if err != nil {
		t.Fatalf("fts search: %v", err)
	}
	if len(hits) != 0 {
		t.Errorf("expected 0 hits after delete, got %d", len(hits))
	}
}

func TestExtractFTS_AllFields(t *testing.T) {
	now := time.Now().UTC()
	m := map[string]any{
		"slug":        "sluggy",
		"displayName": "disp",
		"description": "desc",
		"tags":        []any{"a", "b"},
		"blocks": []any{
			map[string]any{"content": "msg1"},
			map[string]any{"content": "msg2"},
		},
		"isEnabled":  true,
		"modifiedAt": now.Format(time.RFC3339Nano),
	}
	fakePath := "/tmp/bundleb1_sluggy/sluggy.v1.prompt.json"
	vals := extractFTS(fakePath, m).ToMap()
	if vals["slug"] != "sluggy" {
		t.Errorf("slug: got %q", vals["slug"])
	}
	if vals["displayName"] != "disp" {
		t.Errorf("displayName: got %q", vals["displayName"])
	}
	if vals["desc"] != "desc" {
		t.Errorf("desc: got %q", vals["desc"])
	}
	if !strings.Contains(vals["tags"], "a") || !strings.Contains(vals["tags"], "b") {
		t.Errorf("tags: got %q", vals["tags"])
	}
	if !strings.Contains(vals["messages"], "msg1") || !strings.Contains(vals["messages"], "msg2") {
		t.Errorf("messages: got %q", vals["messages"])
	}
	if vals["enabled"] != "true" {
		t.Errorf("enabled: got %q", vals["enabled"])
	}
	if vals["bundleID"] != "bundleb1" {
		t.Errorf("bundleID: got %q", vals["bundleID"])
	}
	if vals["mtime"] != now.Format(time.RFC3339Nano) {
		t.Errorf("mtime: got %q", vals["mtime"])
	}
}

func TestExtractFTS_EmptyAndEdgeCases(t *testing.T) {
	// No tags, no blocks, disabled, missing fields.
	m := map[string]any{
		"slug":      "",
		"isEnabled": false,
	}
	vals := extractFTS("/tmp/bundles__bid__slug/slug.v1.prompt.json", m).ToMap()
	if vals["enabled"] != "false" {
		t.Errorf("enabled: got %q", vals["enabled"])
	}
	if vals["tags"] != "" {
		t.Errorf("tags: got %q", vals["tags"])
	}
	if vals["messages"] != "" {
		t.Errorf("messages: got %q", vals["messages"])
	}
}

func TestProcessFTSSync_SkipNonPromptFiles(t *testing.T) {
	ctx := t.Context()
	tmp := mustTempDir(t)
	defer mustRemoveAll(t, tmp)

	// Non-prompt file.
	otherFile := filepath.Join(tmp, "foo.txt")
	_ = os.WriteFile(otherFile, []byte("hi"), 0o600)

	dec, err := processFTSSync(ctx, tmp, otherFile, func(string) string { return "" })
	if err != nil {
		t.Fatalf("processFTSSync: %v", err)
	}
	if !dec.Skip {
		t.Errorf("expected skip for non-prompt file")
	}
}

func TestProcessFTSSync_Unchanged(t *testing.T) {
	ctx := t.Context()
	tmp := mustTempDir(t)
	defer mustRemoveAll(t, tmp)

	// Write a prompt file.
	bundleID := "bid"
	bundleSlug := "slug"
	bundleDir := filepath.Join(tmp, bundleDirName(bundleID, bundleSlug))
	_ = os.MkdirAll(bundleDir, 0o755)
	tplFile := filepath.Join(bundleDir, "myslug.v1.prompt.json")
	tpl := spec.PromptTemplate{
		ID:          "id1",
		DisplayName: "Test",
		Slug:        "myslug",
		IsEnabled:   true,
		Version:     "v1",
		CreatedAt:   time.Now().UTC(),
		ModifiedAt:  time.Now().UTC(),
	}
	mustWriteJSON(t, tplFile, tpl)

	cmp := fileMTime(tplFile)
	dec, err := processFTSSync(ctx, tmp, tplFile, func(string) string { return cmp })
	if err != nil {
		t.Fatalf("processFTSSync: %v", err)
	}
	if !dec.Unchanged {
		t.Errorf("expected unchanged")
	}
}

func TestProcessFTSSync_BadJSON(t *testing.T) {
	ctx := t.Context()
	tmp := mustTempDir(t)
	defer mustRemoveAll(t, tmp)

	bundleID := "bid"
	bundleSlug := "slug"
	bundleDir := filepath.Join(tmp, bundleDirName(bundleID, bundleSlug))
	_ = os.MkdirAll(bundleDir, 0o755)
	tplFile := filepath.Join(bundleDir, "myslug.v1.prompt.json")
	_ = os.WriteFile(tplFile, []byte("{notjson"), 0o600)

	dec, err := processFTSSync(ctx, tmp, tplFile, func(string) string { return "" })
	if err != nil {
		t.Fatalf("processFTSSync: %v", err)
	}
	if dec.Skip != true {
		t.Errorf("expected skip for bad json")
	}
}

func mustTempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
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

func bundleDirName(id, slug string) string {
	return "bundles__" + id + "__" + slug
}
