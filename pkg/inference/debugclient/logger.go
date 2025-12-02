package debugclient

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
)

type loggingReadCloser struct {
	io.ReadCloser

	buf       *bytes.Buffer
	debugResp *DebugHTTPResponse
	cfg       DebugConfig

	finalized bool // finalized ensures we only compute & attach Data once, even if both, Read hits EOF and Close is called.
}

func (lc *loggingReadCloser) Read(p []byte) (int, error) {
	n, err := lc.ReadCloser.Read(p)
	if n > 0 {
		lc.buf.Write(p[:n])
	}
	// Many SDKs read until EOF but never call Close() on resp.Body.
	// In that case, we still want to attach the body to ResponseDetails.
	if err == io.EOF {
		lc.finalize()
	}
	return n, err
}

func (lc *loggingReadCloser) Close() error {
	// Always call finalize(), even if Close fails, so we still capture
	// whatever we managed to read.
	err := lc.ReadCloser.Close()
	lc.finalize()
	return err
}

// finalize attaches the buffered response body to debugResp.ResponseDetails.Data
// exactly once, and applies sanitization/redaction.
func (lc *loggingReadCloser) finalize() {
	if lc.finalized {
		return
	}
	lc.finalized = true

	if lc.debugResp == nil || lc.debugResp.ResponseDetails == nil {
		return
	}
	if !lc.cfg.CaptureResponseBody {
		// Should not normally be wrapped in this mode, but guard anyway.
		return
	}

	dataBytes := lc.buf.Bytes()
	if len(dataBytes) == 0 {
		return
	}

	// Process and redact body.
	lc.debugResp.ResponseDetails.Data = sanitizeBodyForDebug(dataBytes, false, lc.cfg)

	if lc.cfg.LogToSlog {
		slog.Debug("http_debug: response body raw", "body", string(dataBytes))
	}
}

// SanitizeJSONForDebug is a helper for other packages (e.g. streaming code)
// to sanitize a JSON body using the same redaction logic as the HTTP debug
// client, without going through RoundTrip.
//
// Only StripContent is honored; other DebugConfig fields are irrelevant here.
func SanitizeJSONForDebug(raw []byte, stripContent bool) any {
	cfg := DebugConfig{StripContent: stripContent}
	return sanitizeBodyForDebug(raw, false, cfg)
}

// sanitizeBodyForDebug parses and redacts a JSON or text body according to
// DebugConfig. It returns the sanitized representation as 'any' suitable for
// APIRequestDetails.Data or APIResponseDetails.Data.
func sanitizeBodyForDebug(raw []byte, isRequest bool, cfg DebugConfig) any {
	if len(raw) == 0 {
		return nil
	}

	// Try to parse as JSON (objects or arrays).
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		// Not JSON; treat as plain text.
		s := string(raw)
		if cfg.StripContent {
			return scrubPlainText(s)
		}
		return s
	}

	s := newScrubber(cfg, isRequest)
	return s.scrub(decoded, 0, scrubContext{})
}

func ScrubAnyForDebug(v any, stripContent bool) any {
	cfg := DebugConfig{StripContent: stripContent}
	s := newScrubber(cfg, false)
	return s.scrub(v, 0, scrubContext{})
}
