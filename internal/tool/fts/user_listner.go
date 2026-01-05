package fts

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/ppipada/mapstore-go"
	"github.com/ppipada/mapstore-go/ftsengine"

	"github.com/ppipada/flexigpt-app/internal/bundleitemutils"

	"github.com/ppipada/flexigpt-app/internal/tool/spec"
)

// StartUserToolsFTSRebuild launches a goroutine to rebuild the FTS index
// for all tool specs under baseDir.
func StartUserToolsFTSRebuild(ctx context.Context, baseDir string, e *ftsengine.Engine) {
	if baseDir == "" || e == nil {
		return
	}

	var once sync.Once
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic in tool fts rebuild",
					"err", rec,
					"stack", debug.Stack())
			}
		}()
		once.Do(func() {
			stat, _ := ftsengine.SyncDirToFTS(
				ctx,
				e,
				baseDir,
				compareColumn,
				ftsSyncBatchSize,
				processFTSSync,
			)
			if stat != nil {
				slog.Info("tool user fts sync", "stat", fmt.Sprintf("%v", stat))
			}
		})
	}()
}

// NewUserToolsFTSListener returns a mapstore.FileListener that updates
// the FTS engine on file changes.
func NewUserToolsFTSListener(e *ftsengine.Engine) mapstore.FileListener {
	return func(ev mapstore.FileEvent) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("tools fts listener panic",
					"op", ev.Op, "file", ev.File, "recover", r,
					"stack", string(debug.Stack()))
			}
		}()
		// Reject files that obviously do not belong to us.
		if ev.File == "" ||
			!strings.HasSuffix(ev.File, "."+bundleitemutils.ItemFileExtension) ||
			strings.HasSuffix(ev.File, spec.ToolDBFileName) ||
			strings.HasSuffix(ev.File, spec.ToolBundlesMetaFileName) ||
			strings.HasSuffix(ev.File, spec.ToolBuiltInOverlayDBFileName) {
			return
		}

		ctx := context.Background()
		switch ev.Op {
		case mapstore.OpSetFile, mapstore.OpResetFile:
			vals := extractFTS(ev.File, ev.Data).ToMap()
			if len(vals) == 0 {
				slog.Warn("tools fts listener: nothing to index", "file", ev.File)
				return
			}
			if err := e.Upsert(ctx, ev.File, vals); err != nil {
				slog.Error("tools fts upsert failed", "file", ev.File, "err", err)
			}
		case mapstore.OpDeleteFile:
			if err := e.Delete(ctx, ev.File); err != nil {
				slog.Error("tools fts delete failed", "file", ev.File, "err", err)
			}
		case mapstore.OpSetKey, mapstore.OpDeleteKey:
			// Do nothing as we dont do key operations.
		}
	}
}

// processFTSSync determines if a file should be (re-)indexed in FTS, and extracts its FTS values.
func processFTSSync(
	ctx context.Context,
	baseDir, fullPath string,
	prevCmp ftsengine.GetPrevCmp,
) (ftsengine.SyncDecision, error) {
	skip := ftsengine.SyncDecision{
		ID:   fullPath,
		Skip: true,
	}

	// Only process files with the correct extension.
	if !strings.HasSuffix(fullPath, "."+bundleitemutils.ItemFileExtension) ||
		strings.HasSuffix(fullPath, spec.ToolDBFileName) ||
		strings.HasSuffix(fullPath, spec.ToolBundlesMetaFileName) ||
		strings.HasSuffix(fullPath, spec.ToolBuiltInOverlayDBFileName) {
		return skip, nil
	}

	cmp := fileMTime(fullPath)
	if cmp == prevCmp(fullPath) {
		return ftsengine.SyncDecision{ID: fullPath, Unchanged: true}, nil
	}

	raw, err := os.ReadFile(fullPath)
	if err != nil {
		slog.Error("tool sync fts", "file", fullPath, "read error", err)
		return skip, nil
	}

	tl := spec.Tool{}
	if err := json.Unmarshal(raw, &tl); err != nil {
		slog.Error("tool sync fts", "file", fullPath, "non tool file error", err)
		return skip, nil
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		slog.Error("tool sync fts got json error", "file", fullPath, "error", err)
		return skip, nil
	}

	vals := extractFTS(fullPath, m).ToMap()
	return ftsengine.SyncDecision{
		ID:     fullPath,
		CmpOut: cmp,
		Vals:   vals,
	}, nil
}

// extractFTS converts an in-memory JSON map to a column-to-text map for FTS indexing.
func extractFTS(fullPath string, m map[string]any) ftsDoc {
	var doc ftsDoc
	s, _ := stringField(m, "slug")
	doc.Slug = bundleitemutils.ItemSlug(s)
	doc.DisplayName, _ = stringField(m, "displayName")
	doc.Desc, _ = stringField(m, "description")
	doc.Enabled = enabledFalse
	doc.MTime = fileMTime(fullPath)
	if v, ok := m["isEnabled"].(bool); ok && v {
		doc.Enabled = enabledTrue
	}
	if arr, ok := m["tags"].([]any); ok {
		for _, t := range arr {
			if s, ok := t.(string); ok {
				doc.Tags += s + newline
			}
		}
	}

	doc.Args = extractArgsFromSchema(m["argSchema"])

	if v, ok := m["type"].(string); ok {
		doc.Impl = v
		switch v {
		case string(spec.ToolTypeGo):
			if gi, ok := m["goImpl"].(map[string]any); ok {
				if fn, ok := gi["func"].(string); ok {
					doc.ImplMeta = fn
				}
			}
		case string(spec.ToolTypeHTTP):
			if hi, ok := m["httpImpl"].(map[string]any); ok {
				if req, ok := hi["request"].(map[string]any); ok {
					if meth, ok := req["method"].(string); ok {
						doc.ImplMeta += meth + " "
					}
					if urlT, ok := req["urlTemplate"].(string); ok {
						doc.ImplMeta += urlT
					}
				}
			}
		}
	}

	if dir := filepath.Base(filepath.Dir(fullPath)); dir != "" {
		if bd, err := bundleitemutils.ParseBundleDir(dir); err == nil {
			doc.BundleID = bd.ID
		}
	}

	if ts, ok := stringField(m, "modifiedAt"); ok && ts != "" {
		doc.MTime = ts
	}
	return doc
}

// stringField extracts a string value for a given key from a map, if present.
func stringField(m map[string]any, key string) (string, bool) {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

// fileMTime returns the file's modification time in RFC3339Nano format, or empty string on error.
func fileMTime(path string) string {
	st, err := os.Stat(path)
	if err != nil {
		return ""
	}
	return st.ModTime().UTC().Format(time.RFC3339Nano)
}
