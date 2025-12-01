package debugclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
)

// DebugConfig controls how HTTP debug information is captured and redacted.
type DebugConfig struct {
	// Enabled - If false, the transport short-circuits to the base RoundTripper and
	// captures nothing.
	Enabled bool

	// CaptureRequestBody - Whether to capture the request body into APIRequestDetails.Data.
	CaptureRequestBody bool

	// CaptureResponseBody - Whether to capture the response body into APIResponseDetails.Data.
	CaptureResponseBody bool

	// StripContent - When true, attempts to remove conversation text (user/assistant messages)
	// and large/base64 payloads, while preserving other metadata (model name,
	// tools, usage, etc.).
	StripContent bool

	// LogToSlog - If true, also log request/response details (and raw response body) to
	// slog at debug level.
	LogToSlog bool
}

var DefaultDebugConfig = DebugConfig{
	Enabled:             true,
	LogToSlog:           false,
	CaptureRequestBody:  true,
	CaptureResponseBody: true,
	StripContent:        true, // Only content typed by user/assistant excluded.
}

type contextKey string

const debugHTTPResponseKey = contextKey("DebugHTTPResponse")

// DebugHTTPResponse wraps http debug info stored on the context.
type DebugHTTPResponse struct {
	RequestDetails  *spec.APIRequestDetails
	ResponseDetails *spec.APIResponseDetails
	ErrorDetails    *spec.APIErrorDetails
}

// LogTransport is a custom http.RoundTripper that captures requests and
// responses (including bodies, depending on DebugConfig).
type LogTransport struct {
	Base http.RoundTripper
	Cfg  DebugConfig
}

// NewDebugHTTPClient creates a new http.Client that uses LogTransport.
func NewDebugHTTPClient(cfg DebugConfig) *http.Client {
	return &http.Client{
		Transport: &LogTransport{
			Base: http.DefaultTransport,
			Cfg:  cfg,
		},
	}
}

// RoundTrip executes a single HTTP transaction and captures debug info
// according to the DebugConfig.
func (t *LogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	if !t.Cfg.Enabled {
		// Debugging disabled; just pass through.
		return base.RoundTrip(req)
	}

	reqCtx := req.Context()
	debugResp, _ := GetDebugHTTPResponse(reqCtx)
	if debugResp == nil {
		// Best-effort container (won't be visible to callers unless they also
		// call AddDebugResponseToCtx, but still useful for slog logs).
		debugResp = &DebugHTTPResponse{}
	}

	// Capture request details (including optional body).
	reqDetails := captureRequestDetails(req, t.Cfg)
	debugResp.RequestDetails = reqDetails

	if t.Cfg.LogToSlog {
		slog.Debug("http_debug: request", "details", getDetailsStr(reqDetails))
	}

	// Perform the request.
	resp, err := base.RoundTrip(req)

	// Capture response details (headers, status, and possibly body).
	var respDetails *spec.APIResponseDetails
	if resp != nil {
		respDetails = captureResponseDetails(resp, t.Cfg, debugResp)
		debugResp.ResponseDetails = respDetails
	}

	// Capture error details if an error occurred.
	if err != nil {
		debugResp.ErrorDetails = &spec.APIErrorDetails{
			Message:         err.Error(),
			RequestDetails:  reqDetails,
			ResponseDetails: respDetails,
		}
	}

	if t.Cfg.LogToSlog {
		if respDetails != nil {
			slog.Debug("http_debug: response", "details", getDetailsStr(respDetails))
		}
		if debugResp.ErrorDetails != nil {
			slog.Debug("http_debug: error", "details", getDetailsStr(debugResp.ErrorDetails))
		}
	}

	return resp, err
}

func captureRequestDetails(req *http.Request, cfg DebugConfig) *spec.APIRequestDetails {
	if req == nil {
		return nil
	}
	headers := make(map[string]any, len(req.Header))
	for key, values := range req.Header {
		headers[key] = strings.Join(values, ", ")
	}
	headers = redactHeaders(headers)

	params := make(map[string]any)
	if req.URL != nil {
		q := req.URL.Query()
		for key := range q {
			vals := q[key]
			if len(vals) == 1 {
				params[key] = vals[0]
			} else if len(vals) > 1 {
				params[key] = vals
			}
		}
	}

	var data any
	if cfg.CaptureRequestBody && req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil && len(bodyBytes) > 0 {
			data = sanitizeBodyForDebug(bodyBytes, true, cfg)
		}
		// Reset body so it can be read by the underlying transport & SDK.
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else if !cfg.CaptureRequestBody && req.Body != nil {
		// Indicate that body exists but was intentionally not captured.
		data = "[omitted: request body not captured by debug configuration]"
	}

	urlStr := ""
	if req.URL != nil {
		urlStr = req.URL.String()
	}
	method := req.Method

	apireq := &spec.APIRequestDetails{
		URL:     &urlStr,
		Method:  &method,
		Headers: headers,
		Params:  params,
		Data:    data,
	}

	curl := generateCurlCommand(apireq, cfg)
	apireq.CurlCommand = &curl

	return apireq
}

func captureResponseDetails(
	resp *http.Response,
	cfg DebugConfig,
	debugResp *DebugHTTPResponse,
) *spec.APIResponseDetails {
	if resp == nil {
		return nil
	}

	headers := make(map[string]any, len(resp.Header))
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}
	headers = redactHeaders(headers)

	respDetails := &spec.APIResponseDetails{
		Status:  resp.StatusCode,
		Headers: headers,
	}

	// Wrap the body if we want to capture it.
	if cfg.CaptureResponseBody && resp.Body != nil {
		buffer := new(bytes.Buffer)
		resp.Body = &loggingReadCloser{
			ReadCloser: resp.Body,
			buf:        buffer,
			debugResp:  debugResp,
			cfg:        cfg,
		}
	}

	return respDetails
}

// generateCurlCommand builds a (mostly) copy-pasteable curl command from
// APIRequestDetails. It uses the already-redacted Data and Headers.
func generateCurlCommand(config *spec.APIRequestDetails, cfg DebugConfig) string {
	if config == nil || config.URL == nil || config.Method == nil {
		return ""
	}

	var b strings.Builder

	method := strings.ToUpper(*config.Method)
	b.WriteString("curl")
	if method != "" {
		b.WriteString(" -X ")
		b.WriteString(method)
	}

	if config.URL != nil {
		escapedURL := shellQuote(*config.URL)
		b.WriteString(" ")
		b.WriteString(escapedURL)
	}

	// Headers (sorted for stability).
	if len(config.Headers) > 0 {
		keys := make([]string, 0, len(config.Headers))
		for k := range config.Headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := config.Headers[k]
			headerStr := fmt.Sprintf("%s: %v", k, v)
			b.WriteString(" \\\n  -H ")
			b.WriteString(shellQuote(headerStr))
		}
	}

	if config.Data != nil {
		bodyBytes, err := json.MarshalIndent(config.Data, "", "  ")
		if err == nil {
			b.WriteString(" \\\n  --data-raw ")
			b.WriteString(shellQuote(string(bodyBytes)))
		}
	}

	return b.String()
}

// AddDebugResponseToCtx sets up a DebugHTTPResponse container on the context.
// All SDK calls that should capture HTTP debug must use this context.
func AddDebugResponseToCtx(ctx context.Context) context.Context {
	debugResp := &DebugHTTPResponse{}
	return context.WithValue(ctx, debugHTTPResponseKey, debugResp)
}

// GetDebugHTTPResponse retrieves the DebugHTTPResponse from context.
func GetDebugHTTPResponse(ctx context.Context) (*DebugHTTPResponse, bool) {
	debugResp, ok := ctx.Value(debugHTTPResponseKey).(*DebugHTTPResponse)
	return debugResp, ok
}
