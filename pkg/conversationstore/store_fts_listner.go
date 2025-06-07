package conversationstore

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/filestore"
	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/ftsengine"
)

func NewFTSListner(e *ftsengine.Engine) filestore.Listener {
	return func(ev filestore.Event) {
		switch ev.Op {
		case filestore.OpSetFile, filestore.OpResetFile:
			_ = e.Upsert(context.Background(), ev.File, extract(ev.Data))
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
	// Also accept “Title” from StructToMap.
	if v, ok := m[strings.ToUpper(key[:1])+key[1:]]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

func getUpsertDataForFile(
	ctx context.Context,
	baseDir, fileFullPath string,
) (id string, vals map[string]string, skip bool) {
	if !strings.HasSuffix(fileFullPath, ".json") {
		// Skip non-json files.
		return "", nil, true
	}
	raw, err := os.ReadFile(fileFullPath)
	if err != nil {
		// Skip files that can't be read.
		return "", nil, true
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		// Skip files that can't be decoded.
		return "", nil, true
	}

	vals = extract(m)
	id = fileFullPath
	return id, vals, false
}
