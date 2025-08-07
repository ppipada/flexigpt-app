package fts

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

func TestBuildDoc_HappyAndErrors(t *testing.T) {
	bid, bslug, _ := makeBundle(1, true)
	tpl := makeTemplate(0, true)

	docID, vals, ok := buildDoc(bid, bslug, tpl)
	if !ok {
		t.Fatalf("expected ok==true")
	}
	if !strings.HasPrefix(docID, BuiltInDocPrefix) {
		t.Fatalf("docID does not have builtin prefix: %s", docID)
	}
	if vals[compareColumn] == "" {
		t.Fatalf("compareColumn missing")
	}

	// Invalid slug  -> buildDoc must refuse.
	bad := tpl
	bad.Slug = "white space"
	if _, _, ok2 := buildDoc(bid, bslug, bad); ok2 {
		t.Fatalf("expected buildDoc to fail for invalid slug")
	}
}

func TestSyncBuiltInsToFTS_Scenarios(t *testing.T) {
	tmp := mustTempBuiltInDir(t)
	engine := newEngine(t, tmp)
	ctx := t.Context()

	type want struct {
		rows int
	}
	cases := []struct {
		want

		name   string
		before func()                       // pre-populate index if wanted
		lister func() BuiltInLister         // returns the lister used for sync
		after  func(t *testing.T, rows int) // extra custom asserts
	}{
		{
			name:   "initial import inserts everything",
			before: func() {},
			lister: func() BuiltInLister {
				bid, _, bundle := makeBundle(1, true)
				tpl := makeTemplate(1, true)
				return func(ctx context.Context) (map[bundleitemutils.BundleID]spec.PromptBundle,
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate, error,
				) {
					return map[bundleitemutils.BundleID]spec.PromptBundle{bid: bundle},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate{
							bid: {
								tpl.ID: tpl,
							},
						}, nil
				}
			},
			want: want{rows: 1},
		},
		{
			name: "unchanged data does not grow the table",
			before: func() {
				bid, bslug, _ := makeBundle(2, true)
				tpl := makeTemplate(2, true)
				docID, vals, _ := buildDoc(bid, bslug, tpl)
				if err := engine.Upsert(ctx, docID, vals); err != nil {
					t.Fatalf("Upsert pre-seed: %v", err)
				}
			},
			lister: func() BuiltInLister {
				bid, _, bundle := makeBundle(2, true)
				tpl := makeTemplate(2, true)
				return func(ctx context.Context) (map[bundleitemutils.BundleID]spec.PromptBundle,
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate, error,
				) {
					return map[bundleitemutils.BundleID]spec.PromptBundle{bid: bundle},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate{
							bid: {
								tpl.ID: tpl,
							},
						}, nil
				}
			},
			want: want{rows: 1},
		},
		{
			name: "modified mtime triggers update",
			before: func() {
				bid, bslug, _ := makeBundle(3, true)
				tpl := makeTemplate(3, true)

				docID, vals, _ := buildDoc(bid, bslug, tpl)
				vals[compareColumn] = "2000-01-01T00:00:00Z" // force out-dated
				if err := engine.Upsert(ctx, docID, vals); err != nil {
					t.Fatalf("Upsert pre-seed: %v", err)
				}
			},
			lister: func() BuiltInLister {
				bid, _, bundle := makeBundle(3, true)
				tpl := makeTemplate(3, true) // fresh mtime
				return func(ctx context.Context) (map[bundleitemutils.BundleID]spec.PromptBundle,
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate, error,
				) {
					return map[bundleitemutils.BundleID]spec.PromptBundle{bid: bundle},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate{
							bid: {
								tpl.ID: tpl,
							},
						}, nil
				}
			},
			after: func(t *testing.T, _ int) {
				t.Helper()
				// Ensure the mtime is *not* the old sentinel value any more.
				for id, val := range listAllRows(t, engine) {
					if strings.Contains(id, "bundle-3") && val == "2000-01-01T00:00:00Z" {
						t.Fatalf("row %s not updated", id)
					}
				}
			},
			want: want{rows: 1},
		},
		{
			name: "template vanished -> row deleted",
			before: func() {
				bid, bslug, _ := makeBundle(4, true)
				tpl := makeTemplate(4, false)
				docID, vals, _ := buildDoc(bid, bslug, tpl)
				if err := engine.Upsert(ctx, docID, vals); err != nil {
					t.Fatalf("Upsert pre-seed: %v", err)
				}
			},
			lister: func() BuiltInLister {
				// Return nothing.
				return func(ctx context.Context) (map[bundleitemutils.BundleID]spec.PromptBundle,
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate, error,
				) {
					return map[bundleitemutils.BundleID]spec.PromptBundle{}, map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate{}, nil
				}
			},
			want: want{rows: 0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Isolate each sub-case in its own brand-new DB.
			tmpCase := filepath.Join(tmp, strings.ReplaceAll(tc.name, " ", "_"))
			engineCase := newEngine(t, tmpCase)
			engine = engineCase // used by listAllRows / after callback
			ctx := t.Context()

			if tc.before != nil {
				tc.before()
			}

			lister := tc.lister()
			if err := syncBuiltInsToFTS(ctx, lister, engineCase); err != nil {
				t.Fatalf("syncBuiltInsToFTS: %v", err)
			}

			rows := listAllRows(t, engineCase)
			if got := len(rows); got != tc.rows {
				t.Fatalf("want %d rows, got %d", tc.rows, got)
			}
			if tc.after != nil {
				tc.after(t, len(rows))
			}
		})
	}
}

func TestReindexOneBuiltIn(t *testing.T) {
	tmp := mustTempBuiltInDir(t)
	engine := newEngine(t, tmp)
	ctx := t.Context()

	bid, bslug, _ := makeBundle(99, true)
	baseTpl := makeTemplate(100, true)

	// Error for nil engine.
	if err := ReindexOneBuiltIn(ctx, bid, bslug, baseTpl, nil); err == nil {
		t.Fatalf("expected error for nil engine")
	}

	// Successful single insert.
	if err := ReindexOneBuiltIn(ctx, bid, bslug, baseTpl, engine); err != nil {
		t.Fatalf("ReindexOneBuiltIn: %v", err)
	}
	if got := len(listAllRows(t, engine)); got != 1 {
		t.Fatalf("expected 1 row, got %d", got)
	}

	// Concurrent upserts for many versions.
	const workers = 20
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tpl := baseTpl
			tpl.Version = bundleitemutils.ItemVersion("v" + strconv.Itoa(i))
			tpl.Slug = bundleitemutils.ItemSlug("slug-conc-" + strconv.Itoa(i))
			_ = ReindexOneBuiltIn(ctx, bid, bslug, tpl, engine)
		}(i)
	}
	wg.Wait()

	if got := len(listAllRows(t, engine)); got != workers+1 /* initial */ {
		t.Fatalf("expected %d rows after concurrent inserts, got %d", workers+1, got)
	}
}

func TestSyncBuiltInsToFTS_BiggerThanBatchSize(t *testing.T) {
	tmp := mustTempBuiltInDir(t)
	engine := newEngine(t, tmp)
	ctx := t.Context()

	// Create more docs than upsertBatchSize (constant in production code).
	const total = upsertBatchSize + 25

	bid, _, bundle := makeBundle(1234, true)
	templates := make(map[bundleitemutils.ItemID]spec.PromptTemplate)
	for i := range total {
		t := makeTemplate(i, true)
		templates[t.ID] = t
	}
	lister := func(ctx context.Context) (map[bundleitemutils.BundleID]spec.PromptBundle,
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate, error,
	) {
		return map[bundleitemutils.BundleID]spec.PromptBundle{bid: bundle},
			map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate{
				bid: templates,
			}, nil
	}

	if err := syncBuiltInsToFTS(ctx, lister, engine); err != nil {
		t.Fatalf("syncBuiltInsToFTS: %v", err)
	}

	if got := len(listAllRows(t, engine)); got != total {
		t.Fatalf("expected %d rows, got %d", total, got)
	}
}

func TestSyncBuiltInsToFTS_CompareColumnStoresMTime(t *testing.T) {
	tmp := mustTempBuiltInDir(t)
	engine := newEngine(t, tmp)
	ctx := t.Context()

	bid, _, bundle := makeBundle(55, true)
	tpl := makeTemplate(1, true)

	lister := func(ctx context.Context) (map[bundleitemutils.BundleID]spec.PromptBundle,
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate, error,
	) {
		return map[bundleitemutils.BundleID]spec.PromptBundle{bid: bundle},
			map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.PromptTemplate{bid: {
				tpl.ID: tpl,
			}}, nil
	}

	if err := syncBuiltInsToFTS(ctx, lister, engine); err != nil {
		t.Fatalf("syncBuiltInsToFTS: %v", err)
	}

	rows := listAllRows(t, engine)
	for _, val := range rows {
		// Should parse as RFC3339Nano.
		if _, err := time.Parse(time.RFC3339Nano, val); err != nil {
			t.Fatalf("compareColumn not RFC3339Nano: %v", err)
		}
	}
}

func TestBuildDoc_JSONRoundTrip(t *testing.T) {
	bid, bslug, _ := makeBundle(777, true)
	tpl := makeTemplate(888, false)

	_, vals, ok := buildDoc(bid, bslug, tpl)
	if !ok {
		t.Fatalf("buildDoc failed unexpectedly")
	}

	// Ensure that we can JSON-encode what buildDoc produced - the real engine will do exactly that when inserting rows.
	if _, err := json.Marshal(vals); err != nil {
		t.Fatalf("values not json-marshable: %v", err)
	}
}

func newEngine(t *testing.T, dir string) *ftsengine.Engine {
	t.Helper()
	e, err := ftsengine.NewEngine(ftsengine.Config{
		BaseDir:    dir,
		DBFileName: "fts.db",
		Table:      sqliteDBTableName,
		Columns: []ftsengine.Column{
			{Name: "slug", Weight: 1},
			{Name: "displayName", Weight: 1},
			{Name: "desc", Weight: 1},
			{Name: "messages", Weight: 1},
			{Name: "tags", Weight: 1},
			{Name: "enabled", Unindexed: true},
			{Name: "bundleID", Unindexed: true},
			{Name: "mtime", Unindexed: true},
		},
	})
	if err != nil {
		t.Fatalf("ftsengine.NewEngine: %v", err)
	}
	return e
}

// listAllRows fetches every row that exists in the index.
// Only the compare column is requested because that is what the production sync cares about.
func listAllRows(t *testing.T, e *ftsengine.Engine) map[string]string {
	t.Helper()
	ctx := t.Context()

	res := map[string]string{}
	token := ""
	for {
		part, next, err := e.BatchList(ctx, compareColumn, []string{compareColumn}, token, 200)
		if err != nil {
			t.Fatalf("BatchList: %v", err)
		}
		for _, r := range part {
			res[r.ID] = r.Values[compareColumn]
		}
		if next == "" {
			break
		}
		token = next
	}
	return res
}

func makeTemplate(idx int, enabled bool) spec.PromptTemplate {
	return spec.PromptTemplate{
		ID:          bundleitemutils.ItemID("tpl-" + strconv.Itoa(idx)),
		DisplayName: "Template " + strconv.Itoa(idx),
		Slug:        bundleitemutils.ItemSlug("slug-" + strconv.Itoa(idx)),
		Description: "desc",
		Version:     bundleitemutils.ItemVersion("v1"),
		IsEnabled:   enabled,
		CreatedAt:   time.Now().UTC(),
		ModifiedAt:  time.Now().UTC(),
		Blocks: []spec.MessageBlock{
			{Role: spec.User, Content: "HelloMsg" + strconv.Itoa(idx)},
		},
		Tags: []string{"tag" + strconv.Itoa(idx)},
	}
}

func makeBundle(
	id int,
	enabled bool,
) (bundleitemutils.BundleID, bundleitemutils.BundleSlug, spec.PromptBundle) {
	bid := bundleitemutils.BundleID("bundle-" + strconv.Itoa(id))
	bslug := bundleitemutils.BundleSlug("bundleslug-" + strconv.Itoa(id))
	return bid, bslug, spec.PromptBundle{
		ID:        bid,
		Slug:      bslug,
		IsEnabled: enabled,
	}
}

func mustTempBuiltInDir(t *testing.T) string { t.Helper(); return t.TempDir() }
