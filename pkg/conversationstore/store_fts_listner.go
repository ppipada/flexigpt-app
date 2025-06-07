package conversationstore

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

func NewFTSListner(e *ftsengine.Engine) filestore.Listener {
	return func(ev filestore.Event) {
		switch ev.Op {
		case filestore.OpSetFile, filestore.OpResetFile:
			_ = e.Upsert(context.Background(), ev.File, extract(ev.File, ev.Data))
		}
	}
}

// extract converts the in-memory JSON map (produced by MapFileStore)
// into the column → text map expected by ftsengine.
func extract(fullPath string, m map[string]any) map[string]string {
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

	// Default skip file.
	syncDecision := skipSyncDecision

	// Heavy part only if time stamp differs.
	var m map[string]any
	raw, err := os.ReadFile(fullPath)
	if err == nil {
		if err := json.Unmarshal(raw, &m); err == nil {
			// Invalid JSON → skip.
			vals := extract(fullPath, m)
			syncDecision = ftsengine.SyncDecision{
				ID:        fullPath,
				CmpOut:    cmp,
				Vals:      vals,
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
			// Compare column (must exist in Config.Columns).
			"mtime",
			1000,
			processFTSDataForFile,
		)
	}()
}
