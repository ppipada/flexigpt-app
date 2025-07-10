// Package store provides FTS (Full Text Search) integration for prompt template storage.
// It handles FTS indexing, extraction, and sync logic for prompt template files.
package store

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

const (
	ftsSortField     = "mtime" // ftsSortField is the field used for FTS sort order.
	ftsSyncBatchSize = 1000    // ftsSyncBatchSize is the batch size for FTS sync.
	enabledTrue      = "true"  // enabledTrue is the string value for enabled templates.
	enabledFalse     = "false" // enabledFalse is the string value for disabled templates.
	newline          = "\n"    // newline is the newline character.
)

type ftsDoc struct {
	Slug        string `fts:"slug"`
	DisplayName string `fts:"displayName"`
	Desc        string `fts:"desc"`
	Messages    string `fts:"messages"`
	Tags        string `fts:"tags"`
	Enabled     string `fts:"enabled"`
	BundleID    string `fts:"bundleId"`
	MTime       string `fts:"mtime"`
}

func (d ftsDoc) ToMap() map[string]string {
	return map[string]string{
		"slug":        d.Slug,
		"displayName": d.DisplayName,
		"desc":        d.Desc,
		"messages":    d.Messages,
		"tags":        d.Tags,
		"enabled":     d.Enabled,
		"bundleId":    d.BundleID,
		"mtime":       d.MTime,
	}
}

// NewFTSListener returns a filestore.Listener that updates the FTS engine on file changes.
func NewFTSListener(e *ftsengine.Engine) filestore.Listener {
	return func(ev filestore.Event) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("fts listener panic",
					"op", ev.Op, "file", ev.File, "recover", r,
					"stack", string(debug.Stack()))
			}
		}()
		// Reject files that obviously do not belong to us.
		if ev.File == "" || !strings.HasSuffix(ev.File, "."+promptTemplateFileExtension) ||
			strings.HasSuffix(ev.File, sqliteDBFileName) ||
			strings.HasSuffix(ev.File, bundlesMetaFileName) {
			return
		}
		ctx := context.Background()
		switch ev.Op {
		case filestore.OpSetFile, filestore.OpResetFile:
			vals := extractFTS(ev.File, ev.Data).ToMap()
			if len(vals) == 0 {
				slog.Warn("fts listener: nothing to index", "file", ev.File)
				return
			}
			if err := e.Upsert(ctx, ev.File, vals); err != nil {
				slog.Error("fts upsert failed", "file", ev.File, "err", err)
			}
		case filestore.OpDeleteFile:
			if err := e.Delete(ctx, ev.File); err != nil {
				slog.Error("fts delete failed", "file", ev.File, "err", err)
			}
		}
	}
}

// extractFTS converts an in-memory JSON map to a column-to-text map for FTS indexing.
// It extracts slug, displayName, description, tags, message blocks, enabled status, bundleId, and mtime.
func extractFTS(fullPath string, m map[string]any) ftsDoc {
	var doc ftsDoc
	doc.Slug, _ = stringField(m, "slug")
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

	if arr, ok := m["blocks"].([]any); ok {
		for _, raw := range arr {
			if blk, ok := raw.(map[string]any); ok {
				if txt, ok := blk["content"].(string); ok {
					doc.Messages += txt + newline
				}
			}
		}
	}

	if dir := filepath.Base(filepath.Dir(fullPath)); dir != "" {
		if bd, err := parseBundleDir(dir); err == nil {
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

// processFTSSync determines if a file should be (re-)indexed in FTS, and extracts its FTS values.
// It skips files that do not have the prompt template extension, or are unchanged.
func processFTSSync(
	ctx context.Context,
	baseDir, fullPath string,
	prevCmp ftsengine.GetPrevCmp,
) (ftsengine.SyncDecision, error) {
	skipSyncDecision := ftsengine.SyncDecision{
		ID:        fullPath,
		CmpOut:    "",
		Vals:      map[string]string{},
		Unchanged: false,
		Skip:      true,
	}

	// Only process files with the correct extension.
	if !strings.HasSuffix(fullPath, "."+promptTemplateFileExtension) ||
		strings.HasSuffix(fullPath, sqliteDBFileName) ||
		strings.HasSuffix(fullPath, bundlesMetaFileName) {
		return skipSyncDecision, nil
	}

	// If Stat failed during getting mtime, below read file will fail and we will skip.
	cmp := fileMTime(fullPath)
	if cmp == prevCmp(fullPath) {
		return ftsengine.SyncDecision{ID: fullPath, Unchanged: true, Skip: false}, nil
	}

	raw, err := os.ReadFile(fullPath)
	if err != nil {
		slog.Error("prompt sync fts", "file", fullPath, "read error", err)
		return skipSyncDecision, nil
	}

	pt := spec.PromptTemplate{}
	if err := json.Unmarshal(raw, &pt); err != nil {
		slog.Error("prompt sync fts", "file", fullPath, "non prompt template file error", err)
		return skipSyncDecision, nil
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		slog.Error("sync fts got json error", "file", fullPath, "error", err)
		return skipSyncDecision, nil
	}

	vals := extractFTS(fullPath, m).ToMap()
	return ftsengine.SyncDecision{
		ID:        fullPath,
		CmpOut:    cmp,
		Vals:      vals,
		Unchanged: false,
		Skip:      false,
	}, nil
}

// StartRebuild launches a goroutine to rebuild the FTS index for all prompt templates under baseDir.
func StartRebuild(ctx context.Context, baseDir string, e *ftsengine.Engine) {
	go func() {
		_ = ftsengine.SyncDirToFTS(
			ctx,
			e,
			baseDir,
			ftsSortField,
			ftsSyncBatchSize,
			processFTSSync,
		)
	}()
}
