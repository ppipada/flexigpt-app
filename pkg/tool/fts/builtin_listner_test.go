package fts

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
	"github.com/ppipada/flexigpt-app/pkg/tool/spec"
)

func mustTempBuiltInDir(t *testing.T) string { t.Helper(); return t.TempDir() }

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
			{Name: "parameters", Weight: 1},
			{Name: "tags", Weight: 1},
			{Name: "enabled", Unindexed: true},
			{Name: "bundleID", Unindexed: true},
			{Name: "mtime", Unindexed: true},
		},
	})
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	return e
}

func makeBundle(
	id int,
	enabled bool,
) (bundleitemutils.BundleID, bundleitemutils.BundleSlug, spec.ToolBundle) {
	bid := bundleitemutils.BundleID("bundle-" + strconv.Itoa(id))
	bslug := bundleitemutils.BundleSlug("bundleslug-" + strconv.Itoa(id))
	return bid, bslug, spec.ToolBundle{
		ID:        bid,
		Slug:      bslug,
		IsEnabled: enabled,
	}
}

func makeTool(idx int, enabled bool) spec.ToolSpec {
	return spec.ToolSpec{
		ID:          bundleitemutils.ItemID("tool-" + strconv.Itoa(idx)),
		DisplayName: "Tool " + strconv.Itoa(idx),
		Slug:        bundleitemutils.ItemSlug("slug-" + strconv.Itoa(idx)),
		Description: "desc",
		Version:     bundleitemutils.ItemVersion("v1"),
		IsEnabled:   enabled,
		CreatedAt:   time.Now().UTC(),
		ModifiedAt:  time.Now().UTC(),
		Parameters: []spec.ToolParameter{
			{Name: "p" + strconv.Itoa(idx), Type: spec.ParamString},
		},
		Tags: []string{"tag" + strconv.Itoa(idx)},
	}
}

func listAllRows(t *testing.T, e *ftsengine.Engine) map[string]string {
	t.Helper()
	ctx := t.Context()
	out := map[string]string{}
	token := ""
	for {
		part, next, err := e.BatchList(ctx, compareColumn, []string{compareColumn}, token, 200)
		if err != nil {
			t.Fatalf("BatchList: %v", err)
		}
		for _, r := range part {
			out[r.ID] = r.Values[compareColumn]
		}
		if next == "" {
			break
		}
		token = next
	}
	return out
}

func TestToolsBuildDoc_HappyAndErrors(t *testing.T) {
	bid, bslug, _ := makeBundle(1, true)
	tool := makeTool(0, true)

	docID, vals, ok := buildDoc(bid, bslug, tool)
	if !ok {
		t.Fatalf("expected ok")
	}
	if !strings.HasPrefix(docID, BuiltInDocPrefix) {
		t.Fatalf("no builtin prefix: %s", docID)
	}
	if vals[compareColumn] == "" {
		t.Fatalf("missing compare column")
	}

	// Invalid slug should fail.
	bad := tool
	bad.Slug = "white space"
	if _, _, ok := buildDoc(bid, bslug, bad); ok {
		t.Fatalf("expected failure for bad slug")
	}
}

func TestToolsSyncBuiltInsToFTS_Scenarios(t *testing.T) {
	tmp := mustTempBuiltInDir(t)
	engine := newEngine(t, tmp)
	ctx := t.Context()

	type want struct{ rows int }
	cases := []struct {
		name   string
		before func()
		lister func() ToolBuiltInLister
		after  func(*testing.T, int)
		want
	}{
		{
			name: "initial import",
			lister: func() ToolBuiltInLister {
				bid, _, b := makeBundle(1, true)
				tool := makeTool(1, true)
				return func() (map[bundleitemutils.BundleID]spec.ToolBundle,
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec, error,
				) {
					return map[bundleitemutils.BundleID]spec.ToolBundle{bid: b},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec{bid: {
							tool.ID: tool,
						}}, nil
				}
			},
			want: want{rows: 1},
		},
		{
			name: "unchanged data",
			before: func() {
				bid, bslug, _ := makeBundle(2, true)
				tool := makeTool(2, true)
				docID, vals, _ := buildDoc(bid, bslug, tool)
				_ = engine.Upsert(ctx, docID, vals)
			},
			lister: func() ToolBuiltInLister {
				bid, _, b := makeBundle(2, true)
				tool := makeTool(2, true)
				return func() (map[bundleitemutils.BundleID]spec.ToolBundle,
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec, error,
				) {
					return map[bundleitemutils.BundleID]spec.ToolBundle{bid: b},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec{bid: {
							tool.ID: tool,
						}}, nil
				}
			},
			want: want{rows: 1},
		},
		{
			name: "mtime update",
			before: func() {
				bid, bslug, _ := makeBundle(3, true)
				tool := makeTool(3, true)
				docID, vals, _ := buildDoc(bid, bslug, tool)
				vals[compareColumn] = "2000-01-01T00:00:00Z"
				_ = engine.Upsert(ctx, docID, vals)
			},
			lister: func() ToolBuiltInLister {
				bid, _, b := makeBundle(3, true)
				tool := makeTool(3, true)
				return func() (map[bundleitemutils.BundleID]spec.ToolBundle,
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec, error,
				) {
					return map[bundleitemutils.BundleID]spec.ToolBundle{bid: b},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec{bid: {
							tool.ID: tool,
						}}, nil
				}
			},
			after: func(t *testing.T, _ int) {
				for id, val := range listAllRows(t, engine) {
					if strings.Contains(id, "bundle-3") && val == "2000-01-01T00:00:00Z" {
						t.Fatalf("row not updated")
					}
				}
			},
			want: want{rows: 1},
		},
		{
			name: "row deleted",
			before: func() {
				bid, bslug, _ := makeBundle(4, true)
				tool := makeTool(4, false)
				docID, vals, _ := buildDoc(bid, bslug, tool)
				_ = engine.Upsert(ctx, docID, vals)
			},
			lister: func() ToolBuiltInLister {
				return func() (map[bundleitemutils.BundleID]spec.ToolBundle,
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec, error,
				) {
					return map[bundleitemutils.BundleID]spec.ToolBundle{},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec{}, nil
				}
			},
			want: want{rows: 0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dbDir := filepath.Join(tmp, strings.ReplaceAll(tc.name, " ", "_"))
			engineCase := newEngine(t, dbDir)
			engine = engineCase
			ctx := t.Context()

			if tc.before != nil {
				tc.before()
			}
			lister := tc.lister()
			if err := syncBuiltInsToFTS(ctx, lister, engineCase); err != nil {
				t.Fatalf("sync: %v", err)
			}

			rows := listAllRows(t, engineCase)
			if len(rows) != tc.rows {
				t.Fatalf("want %d rows, got %d", tc.rows, len(rows))
			}
			if tc.after != nil {
				tc.after(t, len(rows))
			}
		})
	}
}

func TestToolsReindexOneBuiltIn(t *testing.T) {
	tmp := mustTempBuiltInDir(t)
	engine := newEngine(t, tmp)
	ctx := t.Context()

	bid, bslug, _ := makeBundle(99, true)
	base := makeTool(100, true)

	if err := ReindexOneBuiltInTool(ctx, bid, bslug, base, nil); err == nil {
		t.Fatalf("want error for nil engine")
	}
	if err := ReindexOneBuiltInTool(ctx, bid, bslug, base, engine); err != nil {
		t.Fatalf("ReindexOneBuiltInTool: %v", err)
	}
	if len(listAllRows(t, engine)) != 1 {
		t.Fatalf("want 1 row")
	}

	const workers = 20
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tool := base
			tool.Version = bundleitemutils.ItemVersion("v" + strconv.Itoa(i))
			tool.Slug = bundleitemutils.ItemSlug("slug-conc-" + strconv.Itoa(i))
			_ = ReindexOneBuiltInTool(ctx, bid, bslug, tool, engine)
		}(i)
	}
	wg.Wait()

	if len(listAllRows(t, engine)) != workers+1 {
		t.Fatalf("want %d rows, got %d", workers+1, len(listAllRows(t, engine)))
	}
}

func TestToolsSyncBuiltInsToFTS_BiggerThanBatchSize(t *testing.T) {
	tmp := mustTempBuiltInDir(t)
	engine := newEngine(t, tmp)
	ctx := t.Context()

	const total = upsertBatchSize + 25
	bid, _, bundle := makeBundle(1234, true)
	tools := make(map[bundleitemutils.ItemID]spec.ToolSpec)
	for i := range total {
		tl := makeTool(i, true)
		tools[tl.ID] = tl
	}
	lister := func() (map[bundleitemutils.BundleID]spec.ToolBundle,
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec, error,
	) {
		return map[bundleitemutils.BundleID]spec.ToolBundle{bid: bundle},
			map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec{bid: tools}, nil
	}
	if err := syncBuiltInsToFTS(ctx, lister, engine); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if got := len(listAllRows(t, engine)); got != total {
		t.Fatalf("want %d rows, got %d", total, got)
	}
}

func TestToolsSyncBuiltInsToFTS_CompareColumnStoresMTime(t *testing.T) {
	tmp := mustTempBuiltInDir(t)
	engine := newEngine(t, tmp)
	ctx := t.Context()

	bid, _, bundle := makeBundle(55, true)
	tool := makeTool(1, true)
	lister := func() (map[bundleitemutils.BundleID]spec.ToolBundle,
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec, error,
	) {
		return map[bundleitemutils.BundleID]spec.ToolBundle{bid: bundle},
			map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.ToolSpec{bid: {
				tool.ID: tool,
			}}, nil
	}
	if err := syncBuiltInsToFTS(ctx, lister, engine); err != nil {
		t.Fatalf("sync: %v", err)
	}
	for _, v := range listAllRows(t, engine) {
		if _, err := time.Parse(time.RFC3339Nano, v); err != nil {
			t.Fatalf("compare column not RFC3339Nano: %v", err)
		}
	}
}

func TestToolsBuildDoc_JSONRoundTrip(t *testing.T) {
	bid, bslug, _ := makeBundle(777, true)
	tool := makeTool(888, false)
	_, vals, ok := buildDoc(bid, bslug, tool)
	if !ok {
		t.Fatalf("buildDoc failed")
	}
	if _, err := json.Marshal(vals); err != nil {
		t.Fatalf("values not json-able: %v", err)
	}
}
