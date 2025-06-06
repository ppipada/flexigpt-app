package conversationstore

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

func NewFTSListner(e *ftsengine.Engine) filestore.Listener {
	return func(ev filestore.Event) {
		switch ev.Op {
		case filestore.OpSetFile, filestore.OpResetFile:
			_ = e.Upsert(ev.File, extract(ev.Data))
		}
	}
}

// extract converts the in-memory JSON map (produced by MapFileStore)
// into the column → text map expected by ftsengine.
func extract(m map[string]any) map[string]string {
	var (
		title, _ = stringField(m, "title")
		system   bytes.Buffer
		user     bytes.Buffer
		assist   bytes.Buffer
		fn       bytes.Buffer
		feedback bytes.Buffer
	)

	// iterate messages --------------------------------------------------
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
	}
}

func stringField(m map[string]any, key string) (string, bool) {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	// also accept “Title” from StructToMap
	if v, ok := m[strings.ToUpper(key[:1])+key[1:]]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

func rebuild(baseDir string, e *ftsengine.Engine) error {
	// walk the directory only once, if table was empty
	return filepath.WalkDir(baseDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return err
		}
		raw, err := os.ReadFile(p)
		var m map[string]any
		if err == nil {
			if err := json.Unmarshal(raw, &m); err == nil {
				return e.Upsert(p, extract(m))
			}
		}
		return nil
	})
}

func rebuildIfEmpty(baseDir string, e *ftsengine.Engine) {
	if e.IsEmpty() {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("fts rebuild panic", "error", r)
				}
			}()
			if err := rebuild(baseDir, e); err != nil {
				slog.Error("fts rebuild", "error", err)
			}
		}()
	}
}
