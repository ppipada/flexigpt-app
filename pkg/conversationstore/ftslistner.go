package conversationstore

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/conversationstore/spec"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

func NewFTSListner(e *ftsengine.Engine) filestore.Listener {
	return func(ev filestore.Event) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("fts listener panic",
					"op", ev.Op, "file", ev.File, "recover", r,
					"stack", string(debug.Stack()))
			}
		}()
		// Reject files that obviously do not belong to us.
		if ev.File == "" || !strings.HasSuffix(ev.File, ".json") {
			return
		}

		ctx := context.Background()
		switch ev.Op {
		case filestore.OpSetFile, filestore.OpResetFile:
			vals := extractFTS(ev.File, ev.Data)
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

// extractFTS converts the in-memory JSON map (produced by MapFileStore)
// into the column → text map expected by ftsengine.
func extractFTS(fullPath string, m map[string]any) map[string]string {
	var (
		title, _ = stringField(m, "title")
		system   bytes.Buffer
		user     bytes.Buffer
		assist   bytes.Buffer
		fn       bytes.Buffer
		feedback bytes.Buffer
	)

	msgs, _ := m["messages"].([]any)
	for _, raw := range msgs {
		msg, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		txt, _ := msg["content"].(string)
		if txt == "" {
			continue
		}
		switch role {
		case "system":
			system.WriteString(txt + "\n")
		case "user":
			user.WriteString(txt + "\n")
		case "assistant":
			assist.WriteString(txt + "\n")
		case "function":
			fn.WriteString(txt + "\n")
		case "feedback":
			feedback.WriteString(txt + "\n")
		}
	}

	return map[string]string{
		"title":     title,
		"system":    system.String(),
		"user":      user.String(),
		"assistant": assist.String(),
		"function":  fn.String(),
		"feedback":  feedback.String(),
		"mtime":     fileMTime(fullPath),
	}
}

func stringField(m map[string]any, key string) (string, bool) {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	// Also accept “Title” from StructToMap.
	if v, ok := m[strings.ToUpper(key[:1])+key[1:]]; ok {
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

func processFTSDataForFile(
	ctx context.Context,
	baseDir, fullPath string,
	getPrevCmpVal ftsengine.GetPrevCmp,
) (
	ftsengine.SyncDecision, error,
) {
	skipSyncDecision := ftsengine.SyncDecision{
		ID:        fullPath,
		CmpOut:    "",
		Vals:      map[string]string{},
		Unchanged: false,
		Skip:      true,
	}
	if !strings.HasSuffix(fullPath, ".json") {
		return skipSyncDecision, nil
	}
	cmp := fileMTime(fullPath)
	prevCmp := getPrevCmpVal(fullPath)
	if cmp == prevCmp {
		return ftsengine.SyncDecision{
			ID:        fullPath,
			CmpOut:    "",
			Vals:      map[string]string{},
			Unchanged: true,
			Skip:      false,
		}, nil
	}

	raw, err := os.ReadFile(fullPath)
	if err != nil {
		slog.Error("conversation sync fts", "file", fullPath, "read error", err)
		return skipSyncDecision, nil
	}

	pt := spec.Conversation{}
	if err := json.Unmarshal(raw, &pt); err != nil {
		slog.Error("conversation sync fts", "file", fullPath, "non conversation file error", err)
		return skipSyncDecision, nil
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		slog.Error("conversation sync fts", "file", fullPath, "json error", err)
		return skipSyncDecision, nil
	}

	vals := extractFTS(fullPath, m)

	return ftsengine.SyncDecision{
		ID:        fullPath,
		CmpOut:    cmp,
		Vals:      vals,
		Unchanged: false,
		Skip:      false,
	}, nil
}

func StartRebuild(ctx context.Context, baseDir string, e *ftsengine.Engine) {
	go func() {
		_ = ftsengine.SyncDirToFTS(
			ctx,
			e,
			baseDir,
			// Compare column (must exist in Config.Columns).
			"mtime",
			1000,
			processFTSDataForFile,
		)
	}()
}
