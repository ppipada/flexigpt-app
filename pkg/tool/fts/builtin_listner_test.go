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

	// Check new fields.
	if vals["args"] == "" {
		t.Errorf("args field should not be empty")
	}
	if vals["impl"] != string(tool.Type) {
		t.Errorf("impl field mismatch: got %q want %q", vals["impl"], tool.Type)
	}
	if tool.Type == spec.ToolTypeGo && !strings.Contains(vals["implMeta"], "ToolFunc") {
		t.Errorf("implMeta for Go tool should contain func name")
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
		want

		name   string
		before func()
		lister func() ToolBuiltInLister
		after  func(*testing.T, int)
	}{
		{
			name: "initial import",
			lister: func() ToolBuiltInLister {
				bid, _, b := makeBundle(1, true)
				tool := makeTool(1, true)
				return func() (map[bundleitemutils.BundleID]spec.ToolBundle,
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool, error,
				) {
					return map[bundleitemutils.BundleID]spec.ToolBundle{bid: b},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool{bid: {
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
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool, error,
				) {
					return map[bundleitemutils.BundleID]spec.ToolBundle{bid: b},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool{bid: {
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
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool, error,
				) {
					return map[bundleitemutils.BundleID]spec.ToolBundle{bid: b},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool{bid: {
							tool.ID: tool,
						}}, nil
				}
			},
			after: func(t *testing.T, _ int) {
				t.Helper()
				for id, vals := range listAllRows(t, engine) {
					if strings.Contains(id, "bundle-3") &&
						vals[compareColumn] == "2000-01-01T00:00:00Z" {
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
					map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool, error,
				) {
					return map[bundleitemutils.BundleID]spec.ToolBundle{},
						map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool{}, nil
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
	tools := make(map[bundleitemutils.ItemID]spec.Tool)
	for i := range total {
		tl := makeTool(i, true)
		tools[tl.ID] = tl
	}
	lister := func() (map[bundleitemutils.BundleID]spec.ToolBundle,
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool, error,
	) {
		return map[bundleitemutils.BundleID]spec.ToolBundle{bid: bundle},
			map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool{bid: tools}, nil
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
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool, error,
	) {
		return map[bundleitemutils.BundleID]spec.ToolBundle{bid: bundle},
			map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool{bid: {
				tool.ID: tool,
			}}, nil
	}
	if err := syncBuiltInsToFTS(ctx, lister, engine); err != nil {
		t.Fatalf("sync: %v", err)
	}
	for _, vals := range listAllRows(t, engine) {
		if _, err := time.Parse(time.RFC3339Nano, vals[compareColumn]); err != nil {
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

func TestFTSContent_GoAndHTTP(t *testing.T) {
	tmp := mustTempBuiltInDir(t)
	engine := newEngine(t, tmp)
	ctx := t.Context()

	bid, _, bundle := makeBundle(42, true)
	goTool := makeTool(0, true)   // Go tool
	httpTool := makeTool(1, true) // HTTP tool

	lister := func() (map[bundleitemutils.BundleID]spec.ToolBundle,
		map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool, error,
	) {
		return map[bundleitemutils.BundleID]spec.ToolBundle{bid: bundle},
			map[bundleitemutils.BundleID]map[bundleitemutils.ItemID]spec.Tool{bid: {
				goTool.ID:   goTool,
				httpTool.ID: httpTool,
			}}, nil
	}
	if err := syncBuiltInsToFTS(ctx, lister, engine); err != nil {
		t.Fatalf("sync: %v", err)
	}
	rows := listAllRows(t, engine)
	foundGo, foundHTTP := false, false
	for _, vals := range rows {
		switch vals["impl"] {
		case string(spec.ToolTypeGo):
			foundGo = true
			if !strings.Contains(vals["implMeta"], "ToolFunc") {
				t.Errorf("Go tool implMeta should contain func name, got: %q", vals["implMeta"])
			}
			if !strings.Contains(vals["args"], "foo") ||
				!strings.Contains(vals["args"], "Foo Title") ||
				!strings.Contains(vals["args"], "Foo Desc") {
				t.Errorf("Go tool args field missing expected content: %q", vals["args"])
			}
		case string(spec.ToolTypeHTTP):
			foundHTTP = true
			if !strings.Contains(vals["implMeta"], "POST") ||
				!strings.Contains(vals["implMeta"], "api.example.com") {
				t.Errorf(
					"HTTP tool implMeta should contain method and URL, got: %q",
					vals["implMeta"],
				)
			}
			if !strings.Contains(vals["args"], "bar") ||
				!strings.Contains(vals["args"], "Bar Desc") {
				t.Errorf("HTTP tool args field missing expected content: %q", vals["args"])
			}
		}
		if !strings.Contains(vals["tags"], "common") {
			t.Errorf("tags field missing expected tag: %q", vals["tags"])
		}
	}
	if !foundGo || !foundHTTP {
		t.Errorf("Expected both Go and HTTP tools indexed, got: %+v", rows)
	}
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

// makeTool returns a Tool with either Go or HTTP implementation, and a simple argSchema.
func makeTool(idx int, enabled bool) spec.Tool {
	argSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"foo": { "type": "string", "title": "Foo Title", "description": "Foo Desc" },
			"bar": { "type": "integer", "description": "Bar Desc" }
		}
	}`)
	tags := []string{"tag" + strconv.Itoa(idx), "common"}

	// Alternate between Go and HTTP tools for variety.
	if idx%2 == 0 {
		return spec.Tool{
			ID:           bundleitemutils.ItemID("tool-" + strconv.Itoa(idx)),
			DisplayName:  "Go Tool " + strconv.Itoa(idx),
			Slug:         bundleitemutils.ItemSlug("slug-" + strconv.Itoa(idx)),
			Description:  "desc go",
			Version:      bundleitemutils.ItemVersion("v1"),
			IsEnabled:    enabled,
			CreatedAt:    time.Now().UTC(),
			ModifiedAt:   time.Now().UTC(),
			ArgSchema:    argSchema,
			OutputSchema: json.RawMessage(`{"type":"object"}`),
			Type:         spec.ToolTypeGo,
			GoImpl: &spec.GoToolImpl{
				Func: "github.com/acme/flexigpt/tools.ToolFunc" + strconv.Itoa(idx),
			},
			Tags: tags,
		}
	} else {
		return spec.Tool{
			ID:           bundleitemutils.ItemID("tool-" + strconv.Itoa(idx)),
			DisplayName:  "HTTP Tool " + strconv.Itoa(idx),
			Slug:         bundleitemutils.ItemSlug("slug-" + strconv.Itoa(idx)),
			Description:  "desc http",
			Version:      bundleitemutils.ItemVersion("v1"),
			IsEnabled:    enabled,
			CreatedAt:    time.Now().UTC(),
			ModifiedAt:   time.Now().UTC(),
			ArgSchema:    argSchema,
			OutputSchema: json.RawMessage(`{"type":"object"}`),
			Type:         spec.ToolTypeHTTP,
			HTTP: &spec.HTTPToolImpl{
				Request: spec.HTTPRequest{
					Method:      "POST",
					URLTemplate: "https://api.example.com/do",
				},
				Response: spec.HTTPResponse{
					SuccessCodes: []int{200},
					Encoding:     "json",
				},
			},
			Tags: tags,
		}
	}
}

func listAllRows(t *testing.T, e *ftsengine.Engine) map[string]map[string]string {
	t.Helper()
	ctx := t.Context()
	out := map[string]map[string]string{}
	token := ""
	for {
		part, next, err := e.BatchList(ctx, compareColumn, []string{
			"slug", "displayName", "desc", "args", "tags", "impl", "implMeta", "enabled", "bundleID", "mtime",
		}, token, 200)
		if err != nil {
			t.Fatalf("BatchList: %v", err)
		}
		for _, r := range part {
			out[r.ID] = r.Values
		}
		if next == "" {
			break
		}
		token = next
	}
	return out
}

func newEngine(t *testing.T, dir string) *ftsengine.Engine {
	t.Helper()
	e, err := ftsengine.NewEngine(ftsengine.Config{
		BaseDir:    dir,
		DBFileName: "fts.db",
		Table:      sqliteDBTableName,
		Columns: []ftsengine.Column{
			{Name: "slug", Weight: 1},
			{Name: "displayName", Weight: 2},
			{Name: "desc", Weight: 3},
			{Name: "args", Weight: 4},
			{Name: "tags", Weight: 5},
			{Name: "impl", Weight: 6},
			{Name: "implMeta", Weight: 7},
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

func mustTempBuiltInDir(t *testing.T) string { t.Helper(); return t.TempDir() }
