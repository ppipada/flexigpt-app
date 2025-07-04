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

// extractFTS converts in-memory JSON map to column → text map for FTS.
func extractFTS(fullPath string, m map[string]any) map[string]string {
	var (
		slug, _    = stringField(m, "slug")
		display, _ = stringField(m, "displayName")
		desc, _    = stringField(m, "description")
		tagsBuf    bytes.Buffer
		blockBuf   bytes.Buffer
		enabled    = "false"
	)
	if v, ok := m["isEnabled"].(bool); ok && v {
		enabled = "true"
	}
	if arr, ok := m["tags"].([]any); ok {
		for _, t := range arr {
			if s, ok := t.(string); ok {
				tagsBuf.WriteString(s + "\n")
			}
		}
	}
	if arr, ok := m["blocks"].([]any); ok {
		for _, raw := range arr {
			if blk, ok := raw.(map[string]any); ok {
				if txt, ok := blk["content"].(string); ok {
					blockBuf.WriteString(txt + "\n")
				}
			}
		}
	}

	bundleId := ""
	if dir := filepath.Base(filepath.Dir(fullPath)); dir != "" {
		if bd, err := parseBundleDir(dir); err == nil {
			bundleId = bd.ID
		}
	}

	return map[string]string{
		"slug":        slug,
		"displayName": display,
		"desc":        desc,
		"messages":    blockBuf.String(),
		"tags":        tagsBuf.String(),
		"enabled":     enabled,
		"bundleId":    bundleId,
		"mtime":       fileMTime(fullPath),
	}
}

func stringField(m map[string]any, key string) (string, bool) {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

func fileMTime(path string) string {
	st, err := os.Stat(path)
	if err != nil {
		return ""
	}
	return st.ModTime().UTC().Format(time.RFC3339Nano)
}

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

	if !strings.HasSuffix(fullPath, "."+promptTemplateFileExtension) {
		return skipSyncDecision, nil
	}
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

func StartRebuild(ctx context.Context, baseDir string, e *ftsengine.Engine) {
	go func() {
		_ = ftsengine.SyncDirToFTS(
			ctx,
			e,
			baseDir,
			"mtime",
			1000,
			processFTSSync,
		)
	}()
}
