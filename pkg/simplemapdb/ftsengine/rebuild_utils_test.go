package ftsengine

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Helper: create a temp dir, cleanup after test.
func withTempDir(t *testing.T, fn func(dir string)) {
	t.Helper()
	dir, err := os.MkdirTemp("", "ftsengine-test-*")
	if err != nil {
		t.Fatal(err)
	}
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
	if err := os.WriteFile(path, b, 0666); err != nil {
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

// Helper: minimal FTS config.
func minimalConfig(baseDir, dbFile string, cols ...Column) Config {
	return Config{
		BaseDir:    baseDir,
		DBFileName: dbFile,
		Table:      "docs",
		Columns:    cols,
	}
}

// Helper: processFile for test (like your consumer).
func testProcessFile(
	ctx context.Context,
	baseDir, fullPath string,
	getPrevCmp GetPrevCmp,
) (SyncDecision, error) {
	slog.Info("Processing", "file", fullPath)
	if !strings.HasSuffix(fullPath, ".json") {
		return SyncDecision{Skip: true}, nil
	}
	st, err := os.Stat(fullPath)
	if err != nil {
		return SyncDecision{Skip: true}, nil
	}
	mtime := st.ModTime().UTC().Format(time.RFC3339Nano)
	prev := getPrevCmp(fullPath)
	if prev == mtime {
		return SyncDecision{ID: fullPath, Unchanged: true}, nil
	}
	raw, err := os.ReadFile(fullPath)
	if err != nil {
		slog.Error("processFile", "error", err)
		return SyncDecision{Skip: true}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return SyncDecision{Skip: true}, nil
	}
	vals := map[string]string{"title": ""}
	if v, ok := m["title"].(string); ok {
		vals["title"] = v
	}
	vals["mtime"] = mtime
	return SyncDecision{
		ID:     fullPath,
		CmpOut: mtime,
		Vals:   vals,
	}, nil
}

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
			Remove       []string // files to remove after first sync
			Modify       []string // files to touch after first sync
			Add          []fileSpec
			ChangeSchema bool
			WantIDs      []string // expected IDs in FTS after sync
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
					filepath.Join(tmpDir, "x/y/z.json"),
					filepath.Join(tmpDir, "x/y2.json"),
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
				// Setup dirs
				for _, d := range tt.Dirs {
					if err := os.MkdirAll(filepath.Join(tmpDir, d), 0777); err != nil {
						t.Fatal(err)
					}
				}
				// Write files
				for _, f := range tt.Files {
					full := filepath.Join(tmpDir, f.RelPath)
					os.MkdirAll(filepath.Dir(full), 0777)
					writeJSONFile(t, full, map[string]any{"title": f.Title})
				}
				// First sync
				err := SyncDirToFTS(
					context.Background(),
					engine,
					tmpDir,
					"mtime",
					2,
					testProcessFile,
				)
				if err != nil {
					t.Fatalf("first sync: %v", err)
				}
				// Remove files if needed
				for _, rel := range tt.Remove {
					full := filepath.Join(tmpDir, rel)
					if err := os.Remove(full); err != nil {
						t.Fatal(err)
					}
				}
				// Modify files if needed
				for _, rel := range tt.Modify {
					full := filepath.Join(tmpDir, rel)
					touchFile(t, full)
				}
				// Add files if needed
				for _, f := range tt.Add {
					full := filepath.Join(tmpDir, f.RelPath)
					os.MkdirAll(filepath.Dir(full), 0777)
					writeJSONFile(t, full, map[string]any{"title": f.Title})
				}
				// Change schema if needed
				if tt.ChangeSchema {
					engine.Close()
					cfg2 := minimalConfig(tmpDir, dbFile,
						Column{Name: "title"},
						Column{Name: "mtime"},
						Column{Name: "extra"}, // new column
					)
					engine2, err := NewEngine(cfg2)
					if err != nil {
						t.Fatalf("schema change: %v", err)
					}
					engine = engine2
				}
				// Second sync
				err = SyncDirToFTS(
					context.Background(),
					engine,
					tmpDir,
					"mtime",
					2,
					testProcessFile,
				)
				if err != nil {
					t.Fatalf("second sync: %v", err)
				}
				// Check FTS contents
				gotIDs := []string{}
				token := ""
				for {
					rows, next, err := engine.BatchList(
						context.Background(),
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
				// Sort for comparison
				want := append([]string{}, tt.WantIDs...)
				got := append([]string{}, gotIDs...)
				// Order doesn't matter
				if !reflect.DeepEqual(stringSet(want), stringSet(got)) {
					t.Errorf("want IDs %v, got %v", want, got)
				}
				// Clean up for next test
				os.RemoveAll(tmpDir)
				os.MkdirAll(tmpDir, 0777)
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

		// Unreadable file
		badFile := filepath.Join(tmpDir, "bad.json")
		writeJSONFile(t, badFile, map[string]any{"title": "bad"})
		os.Chmod(badFile, 0000)
		defer os.Chmod(badFile, 0666)

		// Invalid JSON
		invalidFile := filepath.Join(tmpDir, "invalid.json")
		os.WriteFile(invalidFile, []byte("{not json"), 0666)

		// Non-json file
		txtFile := filepath.Join(tmpDir, "note.txt")
		os.WriteFile(txtFile, []byte("hello"), 0666)

		err = SyncDirToFTS(context.Background(), engine, tmpDir, "mtime", 2, testProcessFile)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Only valid files should be indexed (none in this case)
		token := ""
		var gotIDs []string
		for {
			rows, next, err := engine.BatchList(
				context.Background(),
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
		empty, err := engine.IsEmpty(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if !empty {
			t.Error("expected empty")
		}
		// Add a file
		vals := map[string]string{"title": "foo", "mtime": time.Now().Format(time.RFC3339Nano)}
		err = engine.Upsert(context.Background(), "id1", vals)
		if err != nil {
			t.Fatal(err)
		}
		empty, err = engine.IsEmpty(context.Background())
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
		engine.Upsert(context.Background(), "id1", vals)
		engine.Upsert(context.Background(), "id2", vals)
		engine.Upsert(context.Background(), "id3", vals)
		engine.Delete(context.Background(), "id2")
		engine.BatchDelete(context.Background(), []string{"id1", "id3"})
		empty, err := engine.IsEmpty(context.Background())
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
		engine.Upsert(
			context.Background(),
			"id1",
			map[string]string{"title": "hello world", "mtime": "1"},
		)
		engine.Upsert(
			context.Background(),
			"id2",
			map[string]string{"title": "foo bar", "mtime": "2"},
		)
		hits, next, err := engine.Search(context.Background(), "hello", "", 10)
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
