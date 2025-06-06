package ftsengine

import (
	"path/filepath"
	"strconv"
	"testing"
)

// helper --------------------------------------------------------------------

func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	tmp := t.TempDir()
	e, err := New(Config{
		DBPath: filepath.Join(tmp, "fts.sqlite"),
		Table:  "docs",
		Columns: []Column{
			{Name: "title", Weight: 1},
			{Name: "body", Weight: 5},
		},
	})
	if err != nil {
		t.Fatalf("engine init: %v", err)
	}
	return e
}

/* ------------------------------------------------------------------ */

func TestIsEmptyAndCRUD(t *testing.T) {
	e := newTestEngine(t)
	isEmp, _ := e.IsEmpty(t.Context())
	if !isEmp {
		t.Fatal("new engine should be empty")
	}

	// insert two docs ---------------------------------------------------
	if err := e.Upsert(t.Context(), "doc/alpha", map[string]string{
		"title": "hello world",
		"body":  "ignored",
	}); err != nil {
		t.Fatalf("upsert alpha: %v", err)
	}
	if err := e.Upsert(t.Context(), "doc/bravo", map[string]string{
		"title": "second",
		"body":  "hello world again",
	}); err != nil {
		t.Fatalf("upsert bravo: %v", err)
	}
	isEmp, _ = e.IsEmpty(t.Context())
	if isEmp {
		t.Fatal("index should not be empty after inserts")
	}

	// search – should hit two ------------------------------------------
	hits, next, err := e.Search(t.Context(), "hello", "", 10)
	if err != nil || len(hits) != 2 || next != "" {
		t.Fatalf("search expected 2 hits, got %d (next=%q, err=%v)",
			len(hits), next, err)
	}

	// update one, delete other -----------------------------------------
	if err := e.Upsert(t.Context(), "doc/alpha", map[string]string{
		"title": "updated",
		"body":  "",
	}); err != nil {
		t.Fatalf("update alpha: %v", err)
	}
	_ = e.Delete(t.Context(), "doc/bravo")

	hits, _, _ = e.Search(t.Context(), "hello", "", 10)
	if len(hits) != 0 {
		t.Fatalf("expected 0 hits, got %d", len(hits))
	}
}

/* ------------------------------------------------------------------ */

func TestWeightRanking(t *testing.T) {
	e := newTestEngine(t)

	_ = e.Upsert(t.Context(), "1", map[string]string{
		"title": "alpha winner",
		"body":  "",
	})
	_ = e.Upsert(t.Context(), "2", map[string]string{
		"title": "",
		"body":  "alpha only in body",
	})

	hits, _, err := e.Search(t.Context(), "alpha", "", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 2 {
		t.Fatalf("want 2 hits, got %d", len(hits))
	}
	if hits[0].ID != "1" {
		t.Fatalf("title-match should rank first, got %q", hits[0].ID)
	}
	if hits[0].Score >= hits[1].Score {
		t.Fatalf("bm25 score ordering unexpected: %.3f >= %.3f",
			hits[0].Score, hits[1].Score)
	}
}

/* ------------------------------------------------------------------ */

func TestPaginationToken(t *testing.T) {
	e := newTestEngine(t)

	// 15 docs containing "foo"
	for i := range 15 {
		_ = e.Upsert(t.Context(), "id"+strconv.Itoa(i), map[string]string{
			"title": "",
			"body":  "foo bar",
		})
	}

	token := ""
	seen := map[string]bool{}
	total := 0

	for page := 0; ; page++ {
		hits, next, err := e.Search(t.Context(), "foo", token, 6)
		if err != nil {
			t.Fatalf("page %d: %v", page, err)
		}
		for _, h := range hits {
			if seen[h.ID] {
				t.Fatalf("duplicate id %s across pages", h.ID)
			}
			seen[h.ID] = true
		}
		total += len(hits)
		if next == "" {
			if len(hits) != 3 { // last page
				t.Fatalf("last page size, want 3, got %d", len(hits))
			}
			break
		}
		if len(hits) != 6 {
			t.Fatalf("full pages must have 6 items, got %d", len(hits))
		}
		token = next
	}
	if total != 15 {
		t.Fatalf("expected 15 hits total, got %d", total)
	}
}

/* ------------------------------------------------------------------ */
/*  new edge-case & error-path tests                                  */
/* ------------------------------------------------------------------ */

func TestEdgeCases(t *testing.T) {
	t.Run("constructor failures", func(t *testing.T) {
		if _, err := New(Config{DBPath: ":memory:", Table: "t"}); err == nil {
			t.Error("want error for no columns")
		}
		if _, err := New(Config{
			Table:   "t",
			Columns: []Column{{Name: "c"}},
		}); err == nil {
			t.Error("want error for missing DBPath")
		}
	})

	e := newTestEngine(t)

	t.Run("empty docID rejected", func(t *testing.T) {
		if err := e.Upsert(t.Context(), "", map[string]string{"title": "x"}); err == nil {
			t.Error("expected validation error for empty id")
		}
	})

	t.Run("delete unknown id returns nil error", func(t *testing.T) {
		if err := e.Delete(t.Context(), "does/not/exist"); err != nil {
			t.Errorf("delete unknown: %v", err)
		}
	})

	t.Run("row replacement keeps 1 copy", func(t *testing.T) {
		if err := e.Upsert(t.Context(), "dup", map[string]string{"body": "first"}); err != nil {
			t.Fatal(err)
		}
		if err := e.Upsert(t.Context(), "dup", map[string]string{"body": "second"}); err != nil {
			t.Fatal(err)
		}
		h, _, _ := e.Search(t.Context(), "second", "", 10)
		if len(h) != 1 || h[0].ID != "dup" {
			t.Fatalf("replace failed, hits=%v", h)
		}
	})

	t.Run("IsEmpty resets after all deletes", func(t *testing.T) {
		_ = e.Delete(t.Context(), "dup")
		isEmp, _ := e.IsEmpty(t.Context())
		if !isEmp {
			t.Error("IsEmpty should be true after deleting last row")
		}
	})

	t.Run("token ignored on different query", func(t *testing.T) {
		_ = e.Upsert(t.Context(), "a1", map[string]string{"title": "apple"})
		_ = e.Upsert(t.Context(), "a2", map[string]string{"title": "apple"})

		h1, tok, _ := e.Search(t.Context(), "apple", "", 1)
		if len(h1) != 1 || tok == "" {
			t.Fatalf("setup failed, hits=%d token=%q", len(h1), tok)
		}

		// use token with a DIFFERENT query → offset must reset ⇒ 0 hits
		h2, _, _ := e.Search(t.Context(), "banana", tok, 1)
		if len(h2) != 0 {
			t.Fatalf("token should reset for new query, got %d hits", len(h2))
		}
	})
}
