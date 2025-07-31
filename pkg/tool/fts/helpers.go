package fts

import (
	"encoding/json"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/tool/spec"
)

// digs into a JSON-Schema object, collecting property names / titles / descriptions.
func extractArgsFromSchema(anySchema any) string {
	sch, ok := anySchema.(map[string]any)
	if !ok {
		return ""
	}

	props, ok := sch["properties"].(map[string]any)
	if !ok {
		return ""
	}

	var sb strings.Builder
	for name, raw := range props {
		sb.WriteString(name)
		sb.WriteString(newline)

		if m, ok := raw.(map[string]any); ok {
			if v, ok := m["title"].(string); ok && v != "" && v != name {
				sb.WriteString(v)
				sb.WriteString(newline)
			}
			if v, ok := m["description"].(string); ok {
				sb.WriteString(v)
				sb.WriteString(newline)
			}
		}
	}
	return sb.String()
}

// Convenience when we already have spec.JSONSchema.
func extractArgsFromRaw(raw spec.JSONSchema) string {
	if len(raw) == 0 {
		return ""
	}
	var m map[string]any
	_ = json.Unmarshal(raw, &m) // ignore error - we just fall back to “”
	return extractArgsFromSchema(m)
}
