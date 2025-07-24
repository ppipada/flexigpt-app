package ftsengine

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestSyncDirToFTS_TableDriven(t *testing.T) {
	withTempDir(t, func(tmpDir string) {
		dbFile := "fts.db"
		// Use only "title" and "mtime" columns for simplicity.
		cfg := minimalConfig(tmpDir, dbFile,
			Column{Name: "title"},
			Column{Name: "mtime"},
		)
		engine, err := NewEngine(cfg)
		if err != nil {
			t.Fatalf("engine init: %v", err)
		}
		defer engine.Close()

		type fileSpec struct {
			RelPath string
			Title   string
		}

		tests := []struct {
			Name         string
			Files        []fileSpec
			Dirs         []string
			Remove       []string
			Modify       []string
			Add          []fileSpec
			ChangeSchema bool
			WantIDs      []string
		}{
			{
				Name: "flat files",
				Files: []fileSpec{
					{"a.json", "A"},
					{"b.json", "B"},
				},
				WantIDs: []string{
					filepath.Join(tmpDir, "a.json"),
					filepath.Join(tmpDir, "b.json"),
				},
			},
			{
				Name: "hierarchical tree",
				Files: []fileSpec{
					{"x/y/z.json", "Z"},
					{"x/y2.json", "Y2"},
				},
				Dirs: []string{"x"},
				WantIDs: []string{
					filepath.Join(tmpDir, "x", "y", "z.json"),
					filepath.Join(tmpDir, "x", "y2.json"),
				},
			},
			{
				Name: "delete file after sync",
				Files: []fileSpec{
					{"a.json", "A"},
					{"b.json", "B"},
				},
				Remove: []string{"a.json"},
				WantIDs: []string{
					filepath.Join(tmpDir, "b.json"),
				},
			},
			{
				Name: "add file after sync",
				Files: []fileSpec{
					{"a.json", "A"},
				},
				Add: []fileSpec{
					{"b.json", "B"},
				},
				WantIDs: []string{
					filepath.Join(tmpDir, "a.json"),
					filepath.Join(tmpDir, "b.json"),
				},
			},
			{
				Name:    "empty tree",
				Files:   nil,
				WantIDs: nil,
			},
			{
				Name: "modify file after sync",
				Files: []fileSpec{
					{"a.json", "A"},
				},
				Modify: []string{"a.json"},
				WantIDs: []string{
					filepath.Join(tmpDir, "a.json"),
				},
			},
			{
				Name: "change schema",
				Files: []fileSpec{
					{"a.json", "A"},
				},
				ChangeSchema: true,
				WantIDs: []string{
					filepath.Join(tmpDir, "a.json"),
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.Name, func(t *testing.T) {
				// Setup dirs.
				for _, d := range tt.Dirs {
					if err := os.MkdirAll(filepath.Join(tmpDir, d), 0o777); err != nil {
						t.Fatal(err)
					}
				}
				// Write files.
				for _, f := range tt.Files {
					full := filepath.Join(tmpDir, f.RelPath)
					_ = os.MkdirAll(filepath.Dir(full), 0o777)
					writeJSONFile(t, full, map[string]any{"title": f.Title})
				}
				// First sync.
				err := SyncDirToFTS(
					t.Context(),
					engine,
					tmpDir,
					"mtime",
					2,
					testProcessFile,
				)
				if err != nil {
					t.Fatalf("first sync: %v", err)
				}
				// Remove files if needed.
				for _, rel := range tt.Remove {
					full := filepath.Join(tmpDir, rel)
					if err := os.Remove(full); err != nil {
						t.Fatal(err)
					}
				}
				// Modify files if needed.
				for _, rel := range tt.Modify {
					full := filepath.Join(tmpDir, rel)
					touchFile(t, full)
				}
				// Add files if needed.
				for _, f := range tt.Add {
					full := filepath.Join(tmpDir, f.RelPath)
					_ = os.MkdirAll(filepath.Dir(full), 0o777)
					writeJSONFile(t, full, map[string]any{"title": f.Title})
				}
				// Change schema if needed.
				if tt.ChangeSchema {
					engine.Close()
					cfg2 := minimalConfig(tmpDir, dbFile,
						Column{Name: "title"},
						Column{Name: "mtime"},
						Column{Name: "extra"},
					)
					engine2, err := NewEngine(cfg2)
					if err != nil {
						t.Fatalf("schema change: %v", err)
					}
					engine = engine2
				}
				// Second sync.
				err = SyncDirToFTS(
					t.Context(),
					engine,
					tmpDir,
					"mtime",
					2,
					testProcessFile,
				)
				if err != nil {
					t.Fatalf("second sync: %v", err)
				}
				// Check FTS contents.
				gotIDs := []string{}
				token := ""
				for {
					rows, next, err := engine.BatchList(
						t.Context(),
						"mtime",
						[]string{"mtime"},
						token,
						100,
					)
					if err != nil {
						t.Fatalf("batchlist: %v", err)
					}
					for _, r := range rows {
						gotIDs = append(gotIDs, r.ID)
					}
					if next == "" {
						break
					}
					token = next
				}
				// Sort for comparison.
				want := slices.Clone(tt.WantIDs)
				got := slices.Clone(gotIDs)
				// Order doesn't matter.
				if !reflect.DeepEqual(stringSet(want), stringSet(got)) {
					t.Errorf("want IDs %v, got %v", want, got)
				}
				// Clean up for next test.
				os.RemoveAll(tmpDir)
				_ = os.MkdirAll(tmpDir, 0o777)
			})
		}
	})
}

// Helper: set of strings for order-insensitive comparison.
func stringSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func TestSyncDirToFTS_ErrorCases(t *testing.T) {
	withTempDir(t, func(tmpDir string) {
		dbFile := "fts.db"
		cfg := minimalConfig(tmpDir, dbFile,
			Column{Name: "title"},
			Column{Name: "mtime"},
		)
		engine, err := NewEngine(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()

		// Unreadable file.
		badFile := filepath.Join(tmpDir, "bad.json")
		writeJSONFile(t, badFile, map[string]any{"title": "bad"})
		_ = os.Chmod(badFile, 0o000)
		defer func() { _ = os.Chmod(badFile, 0o666) }()

		// Invalid JSON.
		invalidFile := filepath.Join(tmpDir, "invalid.json")
		_ = os.WriteFile(invalidFile, []byte("{not json"), 0o600)

		// Non-json file.
		txtFile := filepath.Join(tmpDir, "note.txt")
		_ = os.WriteFile(txtFile, []byte("hello"), 0o600)

		err = SyncDirToFTS(t.Context(), engine, tmpDir, "mtime", 2, testProcessFile)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Only valid files should be indexed (none in this case).
		token := ""
		var gotIDs []string
		for {
			rows, next, err := engine.BatchList(
				t.Context(),
				"mtime",
				[]string{"mtime"},
				token,
				100,
			)
			if err != nil {
				t.Fatalf("batchlist: %v", err)
			}
			for _, r := range rows {
				gotIDs = append(gotIDs, r.ID)
			}
			if next == "" {
				break
			}
			token = next
		}
		if len(gotIDs) != 0 {
			t.Errorf("expected no indexed files, got %v", gotIDs)
		}
	})
}

func TestFTSEngine_IsEmpty(t *testing.T) {
	withTempDir(t, func(tmpDir string) {
		cfg := minimalConfig(tmpDir, "fts.db",
			Column{Name: "title"},
			Column{Name: "mtime"},
		)
		engine, err := NewEngine(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()
		empty, err := engine.IsEmpty(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if !empty {
			t.Error("expected empty")
		}
		// Add a file.
		vals := map[string]string{"title": "foo", "mtime": time.Now().Format(time.RFC3339Nano)}
		err = engine.Upsert(t.Context(), "id1", vals)
		if err != nil {
			t.Fatal(err)
		}
		empty, err = engine.IsEmpty(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if empty {
			t.Error("expected not empty")
		}
	})
}

func TestFTSEngine_DeleteAndBatchDelete(t *testing.T) {
	withTempDir(t, func(tmpDir string) {
		cfg := minimalConfig(tmpDir, "fts.db",
			Column{Name: "title"},
			Column{Name: "mtime"},
		)
		engine, err := NewEngine(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()
		vals := map[string]string{"title": "foo", "mtime": time.Now().Format(time.RFC3339Nano)}
		_ = engine.Upsert(t.Context(), "id1", vals)
		_ = engine.Upsert(t.Context(), "id2", vals)
		_ = engine.Upsert(t.Context(), "id3", vals)
		_ = engine.Delete(t.Context(), "id2")
		_ = engine.BatchDelete(t.Context(), []string{"id1", "id3"})
		empty, err := engine.IsEmpty(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if !empty {
			t.Error("expected empty after deletes")
		}
	})
}

func TestFTSEngine_Search(t *testing.T) {
	withTempDir(t, func(tmpDir string) {
		cfg := minimalConfig(tmpDir, "fts.db",
			Column{Name: "title"},
			Column{Name: "mtime"},
		)
		engine, err := NewEngine(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()
		_ = engine.Upsert(
			t.Context(),
			"id1",
			map[string]string{"title": "hello world", "mtime": "1"},
		)
		_ = engine.Upsert(
			t.Context(),
			"id2",
			map[string]string{"title": "foo bar", "mtime": "2"},
		)
		hits, next, err := engine.Search(t.Context(), "hello", "", 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(hits) != 1 || hits[0].ID != "id1" {
			t.Errorf("unexpected search hits: %+v", hits)
		}
		if next != "" {
			t.Errorf("unexpected next token: %q", next)
		}
	})
}

// Helper: processFile for test (like your consumer).
func testProcessFile(
	ctx context.Context,
	baseDir, fullPath string,
	getPrevCmp GetPrevCmp,
) (SyncDecision, error) {
	slog.Info("processing", "file", fullPath)
	if !strings.HasSuffix(fullPath, ".json") {
		return SyncDecision{Skip: true}, nil
	}
	st, err := os.Stat(fullPath)
	if err != nil {
		return SyncDecision{Skip: true}, err
	}
	mtime := st.ModTime().UTC().Format(time.RFC3339Nano)
	prev := getPrevCmp(fullPath)
	if prev == mtime {
		return SyncDecision{ID: fullPath, Unchanged: true}, nil
	}
	syncDecision := SyncDecision{Skip: true}
	raw, err := os.ReadFile(fullPath)
	if err == nil {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err == nil {
			vals := map[string]string{"title": ""}
			if v, ok := m["title"].(string); ok {
				vals["title"] = v
			}
			vals["mtime"] = mtime
			syncDecision = SyncDecision{
				ID:     fullPath,
				CmpOut: mtime,
				Vals:   vals,
			}
		}
	}
	return syncDecision, nil
}

// Helper: minimal FTS config.
func minimalConfig(baseDir, dbFile string, cols ...Column) Config {
	return Config{
		BaseDir:    baseDir,
		DBFileName: dbFile,
		Table:      "docs",
		Columns:    cols,
	}
}

// Helper: create a temp dir, cleanup after test.
func withTempDir(t *testing.T, fn func(dir string)) {
	t.Helper()
	dir := t.TempDir()
	defer os.RemoveAll(dir)
	fn(dir)
}

// Helper: write a JSON file.
func writeJSONFile(t *testing.T, path string, m map[string]any) {
	t.Helper()
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

// Helper: touch file to update mtime.
func touchFile(t *testing.T, path string) {
	t.Helper()
	now := time.Now().Add(time.Duration(1) * time.Second)
	if err := os.Chtimes(path, now, now); err != nil {
		t.Fatal(err)
	}
}
