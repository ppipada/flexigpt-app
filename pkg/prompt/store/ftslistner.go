// Package store provides FTS (Full Text Search) integration for prompt template storage.
// It handles FTS indexing, extraction, and sync logic for prompt template files.
package store

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// NewFTSListener returns a filestore.Listener that updates the FTS engine on file changes.
func NewFTSListener(e *ftsengine.Engine) filestore.Listener {
	return func(ev filestore.Event) {
		switch ev.Op {
		case filestore.OpSetFile, filestore.OpResetFile:
			_ = e.Upsert(context.Background(), ev.File, extractFTS(ev.File, ev.Data))
		case filestore.OpDeleteFile:
			_ = e.Delete(context.Background(), ev.File)
		}
	}
}

// extractFTS converts an in-memory JSON map to a column-to-text map for FTS indexing.
// It extracts slug, displayName, description, tags, message blocks, enabled status, bundleId, and mtime.
func extractFTS(fullPath string, m map[string]any) map[string]string {
	var (
		slug, _    = stringField(m, "slug")
		display, _ = stringField(m, "displayName")
		desc, _    = stringField(m, "description")
		tagsBuf    bytes.Buffer
		blockBuf   bytes.Buffer
		enabled    = enabledFalse
		mtime      = fileMTime(fullPath)
		bundleId   = ""
	)

	// Extract enabled status.
	if v, ok := m["isEnabled"].(bool); ok && v {
		enabled = enabledTrue
	}

	// Extract tags as newline-separated string.
	if arr, ok := m["tags"].([]any); ok {
		for _, t := range arr {
			if s, ok := t.(string); ok {
				tagsBuf.WriteString(s)
				tagsBuf.WriteString(newline)
			}
		}
	}

	// Extract message blocks as newline-separated string.
	if arr, ok := m["blocks"].([]any); ok {
		for _, raw := range arr {
			if blk, ok := raw.(map[string]any); ok {
				if txt, ok := blk["content"].(string); ok {
					blockBuf.WriteString(txt)
					blockBuf.WriteString(newline)
				}
			}
		}
	}

	// Extract bundleId from parent directory.
	if dir := filepath.Base(filepath.Dir(fullPath)); dir != "" {
		if bd, err := parseBundleDir(dir); err == nil {
			bundleId = bd.ID
		}
	}

	// Extract modified time.
	if ts, ok := stringField(m, "modifiedAt"); ok && ts != "" {
		mtime = ts
	}

	return map[string]string{
		"slug":        slug,
		"displayName": display,
		"desc":        desc,
		"messages":    blockBuf.String(),
		"tags":        tagsBuf.String(),
		"enabled":     enabled,
		"bundleId":    bundleId,
		"mtime":       mtime,
	}
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
	if !strings.HasSuffix(fullPath, "."+promptTemplateFileExtension) {
		return skipSyncDecision, nil
	}

	// If Stat failed during getting mtime, below read file will fail and we will skip.
	cmp := fileMTime(fullPath)
	if cmp == prevCmp(fullPath) {
		return ftsengine.SyncDecision{ID: fullPath, Unchanged: true, Skip: false}, nil
	}

	syncDecision := skipSyncDecision
	var m map[string]any
	raw, err := os.ReadFile(fullPath)
	if err == nil {
		if err := json.Unmarshal(raw, &m); err == nil {
			syncDecision = ftsengine.SyncDecision{
				ID:        fullPath,
				CmpOut:    cmp,
				Vals:      extractFTS(fullPath, m),
				Unchanged: false,
				Skip:      false,
			}
		}
	}

	return syncDecision, nil
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
