package debugclient

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
)

func DecodeAndPrintJSON(s string) {
	var obj any
	err := json.Unmarshal([]byte(s), &obj)
	if err != nil {
		slog.Info("json unmarshal error", "msg", err.Error())
	} else {
		PrintJSON(obj)
	}
}

// PrintJSON logs a value as JSON at info level (unchanged helper).
func PrintJSON(v any) {
	p, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		slog.Info("json marshal error", "msg", err.Error())
	} else {
		slog.Info("request params", "json", string(p))
	}
}

// looksLikeBase64 heuristically detects large base64 strings or data URLs.
func looksLikeBase64(s string) bool {
	if len(s) < 128 {
		return false
	}
	// Common pattern: data URLs.
	if strings.Contains(s, "base64,") {
		return true
	}

	// Heuristic: characters limited to base64 charset.
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '+', r == '/', r == '=', r == '\n', r == '\r':
		default:
			return false
		}
	}

	// Optionally require length to be multiple of 4.
	if len(s)%4 != 0 {
		return false
	}
	return true
}

// pointerOf returns a stable pointer identity for maps and slices, or 0
// otherwise; used for cycle detection.
func pointerOf(x any) uintptr {
	rv := reflect.ValueOf(x)
	switch rv.Kind() { //nolint:exhaustive // Cannot cover all types.
	case reflect.Map, reflect.Slice:
		if rv.IsNil() {
			return 0
		}
		return rv.Pointer()
	default:
		return 0
	}
}

// shellQuote quotes a string for POSIX shells using single quotes.
func shellQuote(s string) string {
	//  Pattern ' -> '\''.
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// getDetailsStr pretty-prints any value as JSON for logging.
func getDetailsStr(v any) string {
	s, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Could not marshal object to JSON: %+v", v)
	}
	return string(s)
}
