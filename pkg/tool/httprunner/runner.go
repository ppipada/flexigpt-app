package httprunner

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
	"github.com/ppipada/flexigpt-app/pkg/tool/spec"
)

// HTTPToolRunner executes an HTTPToolImpl. Safe for concurrent use.
type HTTPToolRunner struct {
	impl spec.HTTPToolImpl

	overrideTimeoutMs int
	extraHeaders      map[string]string
	secrets           map[string]string

	// HTTP Clients keyed by effective timeout (ms). Each client is safe for concurrent use.
	clientsMu sync.RWMutex
	clients   map[int]*http.Client
}

// Functional options for HTTPToolRunner.
type HTTPOption func(*HTTPToolRunner)

// WithHTTPTimeoutMs sets an invocation timeout override (milliseconds).
func WithHTTPTimeoutMs(ms int) HTTPOption {
	return func(r *HTTPToolRunner) {
		if ms > 0 {
			r.overrideTimeoutMs = ms
		}
	}
}

// WithHTTPExtraHeaders injects/overrides request headers.
func WithHTTPExtraHeaders(h map[string]string) HTTPOption {
	return func(r *HTTPToolRunner) {
		if len(h) == 0 {
			return
		}
		if r.extraHeaders == nil {
			r.extraHeaders = make(map[string]string, len(h))
		}
		maps.Copy(r.extraHeaders, h)
	}
}

// WithHTTPSecrets provides secrets used in templating (e.g., ${SECRET}).
func WithHTTPSecrets(s map[string]string) HTTPOption {
	return func(r *HTTPToolRunner) {
		if len(s) == 0 {
			return
		}
		if r.secrets == nil {
			r.secrets = make(map[string]string, len(s))
		}
		maps.Copy(r.secrets, s)
	}
}

func NewHTTPToolRunner(impl spec.HTTPToolImpl, opts ...HTTPOption) (*HTTPToolRunner, error) {
	if err := validateHTTPImpl(&impl); err != nil {
		return nil, err
	}
	r := &HTTPToolRunner{
		impl:    impl,
		clients: make(map[int]*http.Client),
	}
	for _, o := range opts {
		o(r)
	}
	return r, nil
}

func (r *HTTPToolRunner) Run(
	ctx context.Context,
	inArgs json.RawMessage,
) (output json.RawMessage, metaData map[string]any, err error) {
	req := r.impl.Request
	resp := r.impl.Response
	timeoutMs := req.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = spec.DefaultHTTPTimeoutMs
	}
	if r.overrideTimeoutMs > 0 {
		timeoutMs = r.overrideTimeoutMs
	}

	client := r.client(timeoutMs)
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}

	secret := ""
	if r.secrets != nil {
		secret = r.secrets["SECRET"]
	}

	// Decode args into map for templating. Non-object args will simply result in no substitutions.
	args, _ := encdec.DecodeJSONRaw[map[string]any](inArgs)

	// Build URL with templating.
	uStr, err := expandTemplate(req.URLTemplate, args, secret)
	if err != nil {
		return nil, nil, fmt.Errorf("urlTemplate expansion failed: %w", err)
	}
	u, err := url.Parse(uStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid URL after expansion: %w", err)
	}

	// Apply query params.
	q := u.Query()
	for k, v := range req.Query {
		expanded, err := expandTemplate(v, args, secret)
		if err != nil {
			return nil, nil, fmt.Errorf("query[%s] expansion failed: %w", k, err)
		}
		if expanded != "" {
			q.Set(k, expanded)
		}
	}
	u.RawQuery = q.Encode()

	// Headers.
	headers := make(http.Header)
	for k, v := range req.Headers {
		expanded, err := expandTemplate(v, args, secret)
		if err != nil {
			return nil, nil, fmt.Errorf("header[%s] expansion failed: %w", k, err)
		}
		if expanded != "" {
			headers.Set(k, expanded)
		}
	}

	// Auth.
	if req.Auth != nil {
		if err := applyAuth(headers, u, req.Auth, args, secret); err != nil {
			return nil, nil, fmt.Errorf("auth apply failed: %w", err)
		}
	}

	// Body (JSON only).
	var body io.Reader
	if strings.TrimSpace(req.Body) != "" && methodAllowsBody(method) {
		expanded, err := expandTemplate(req.Body, args, secret)
		if err != nil {
			return nil, nil, fmt.Errorf("body expansion failed: %w", err)
		}
		expandedBytes := []byte(expanded)
		if !json.Valid(expandedBytes) {
			return nil, nil, errors.New("request body must be valid JSON after templating")
		}
		body = bytes.NewReader(expandedBytes)
		// Force JSON content-type if not set.
		if headers.Get("Content-Type") == "" {
			headers.Set("Content-Type", "application/json; charset=utf-8")
		}
	}

	// Extra headers (override any existing).
	for k, v := range r.extraHeaders {
		headers.Set(k, v)
	}

	// Force JSON Accept header by default.
	if headers.Get("Accept") == "" {
		headers.Set("Accept", "application/json")
	}

	// Build request.
	httpReq, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, nil, fmt.Errorf("http.NewRequest: %w", err)
	}
	httpReq.Header = headers

	start := time.Now()
	httpResp, err := client.Do(httpReq)
	dur := time.Since(start)
	if err != nil {
		return nil, nil, fmt.Errorf("http.Do: %w", err)
	}
	defer httpResp.Body.Close()

	isSuccess := isSuccessStatus(httpResp.StatusCode, resp.SuccessCodes)
	errorMode := resp.ErrorMode
	if errorMode == "" {
		errorMode = spec.DefaultHTTPErrorMode
	}

	data, readErr := io.ReadAll(httpResp.Body)
	if readErr != nil {
		return nil, nil, fmt.Errorf("read response: %w", readErr)
	}

	metaData = map[string]any{
		"type":        "http",
		"url":         u.String(),
		"status":      httpResp.StatusCode,
		"durationMs":  dur.Milliseconds(),
		"contentType": httpResp.Header.Get("Content-Type"),
	}

	if !isSuccess {
		if strings.EqualFold(errorMode, "empty") {
			return nil, metaData, nil
		}
		return nil, metaData, fmt.Errorf("http status %d not in success set", httpResp.StatusCode)
	}

	// JSON-only response handling.
	ct := httpResp.Header.Get("Content-Type")
	if len(data) == 0 {
		// No body -> return JSON null.
		return json.RawMessage("null"), metaData, nil
	}
	if !isJSONContentType(ct) {
		return nil, metaData, fmt.Errorf("non-JSON Content-Type: %q", ct)
	}
	if !json.Valid(data) {
		return nil, metaData, errors.New("response body is not valid JSON")
	}
	return json.RawMessage(data), metaData, nil
}

func (r *HTTPToolRunner) client(timeoutMs int) *http.Client {
	r.clientsMu.RLock()
	if c, ok := r.clients[timeoutMs]; ok {
		r.clientsMu.RUnlock()
		return c
	}
	r.clientsMu.RUnlock()

	r.clientsMu.Lock()
	defer r.clientsMu.Unlock()
	if c, ok := r.clients[timeoutMs]; ok {
		return c
	}
	c := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}
	r.clients[timeoutMs] = c
	return c
}

func methodAllowsBody(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func isSuccessStatus(status int, successCodes []int) bool {
	if len(successCodes) == 0 {
		return status >= 200 && status < 300
	}
	return slices.Contains(successCodes, status)
}

// Only accept explicit JSON types: application/json or any +json subtype.
func isJSONContentType(ct string) bool {
	if ct == "" {
		return false // when body is present and content-type missing, treat as non-JSON
	}
	ct = strings.ToLower(ct)
	// Strip parameters like "; charset=utf-8".
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = ct[:i]
	}
	ct = strings.TrimSpace(ct)
	return ct == "application/json" || strings.HasSuffix(ct, "+json")
}

// applyAuth applies HTTP auth directives. For Type "apiKey", In must be "header" or "query".
// For "bearer": sets "Authorization: Bearer <value>".
// For "basic": sets "Authorization: Basic base64(user:pass)" where <value> is "user:pass".
func applyAuth(
	h http.Header,
	u *url.URL,
	a *spec.HTTPAuth,
	args map[string]any,
	secret string,
) error {
	if a == nil {
		return nil
	}
	val, err := expandTemplate(a.ValueTemplate, args, secret)
	if err != nil {
		return fmt.Errorf("auth valueTemplate expansion: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(a.Type)) {
	case "apikey", "api_key", "api-key":
		loc := strings.ToLower(strings.TrimSpace(a.In))
		if loc != "header" && loc != "query" {
			return errors.New(`auth "in" must be "header" or "query" for apiKey`)
		}
		if strings.TrimSpace(a.Name) == "" {
			return errors.New(`auth "name" is required for apiKey`)
		}
		if loc == "header" {
			h.Set(a.Name, val)
		} else {
			q := u.Query()
			q.Set(a.Name, val)
			u.RawQuery = q.Encode()
		}
	case "bearer":
		h.Set("Authorization", "Bearer "+val)
	case "basic":
		enc := base64.StdEncoding.EncodeToString([]byte(val))
		h.Set("Authorization", "Basic "+enc)
	default:
		return fmt.Errorf("unsupported auth type: %s", a.Type)
	}
	return nil
}

// expandTemplate replaces ${path} tokens with values resolved from args (dot-path with [idx]).
// ${SECRET} is replaced with the provided secret if present.
func expandTemplate(s string, args map[string]any, secret string) (string, error) {
	if s == "" {
		return "", nil
	}
	re := regexp.MustCompile(`\$\{([a-zA-Z0-9_.\[\]-]+)\}`)
	out := re.ReplaceAllStringFunc(s, func(m string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(m, "${"), "}")
		if name == "SECRET" {
			return secret
		}
		v, ok := resolvePath(args, name)
		if !ok || v == nil {
			return ""
		}
		return fmt.Sprintf("%v", v)
	})
	return out, nil
}

// resolvePath resolves dot/array-index path into args (e.g. user.name, items[0].id).
func resolvePath(root any, path string) (any, bool) {
	cur := root
	remain := path
	for remain != "" {
		var seg string
		if i := strings.IndexAny(remain, ".["); i >= 0 {
			seg = remain[:i]
			remain = remain[i:]
		} else {
			seg = remain
			remain = ""
		}
		if seg != "" {
			m, ok := cur.(map[string]any)
			if !ok {
				return nil, false
			}
			var ok2 bool
			cur, ok2 = m[seg]
			if !ok2 {
				return nil, false
			}
		}
		for strings.HasPrefix(remain, "[") {
			end := strings.Index(remain, "]")
			if end <= 1 {
				return nil, false
			}
			idxStr := remain[1:end]
			remain = remain[end+1:]
			idx, err := strconv.Atoi(idxStr)
			if err != nil || idx < 0 {
				return nil, false
			}
			arr, ok := toSlice(cur)
			if !ok || idx >= len(arr) {
				return nil, false
			}
			cur = arr[idx]
			remain = strings.TrimPrefix(remain, ".")
		}
		remain = strings.TrimPrefix(remain, ".")
	}
	return cur, true
}

func toSlice(v any) ([]any, bool) {
	switch x := v.(type) {
	case []any:
		return x, true
	case []map[string]any:
		out := make([]any, len(x))
		for i := range x {
			out[i] = x[i]
		}
		return out, true
	default:
		return nil, false
	}
}

func validateHTTPImpl(impl *spec.HTTPToolImpl) error {
	if impl == nil {
		return errors.New("httpImpl is nil")
	}
	if strings.TrimSpace(impl.Request.URLTemplate) == "" {
		return errors.New("httpImpl.request.urlTemplate is empty")
	}
	// Enforce method sanity if set.
	if impl.Request.Method != "" {
		m := strings.ToUpper(strings.TrimSpace(impl.Request.Method))
		switch m {
		case http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions:
			// Ok.
		default:
			return fmt.Errorf("unsupported http method: %s", m)
		}
	}
	// SuccessCodes sanity.
	for _, c := range impl.Response.SuccessCodes {
		if c < 100 || c > 599 {
			return fmt.Errorf("invalid success code: %d", c)
		}
	}
	if impl.Response.ErrorMode != "" {
		em := strings.ToLower(impl.Response.ErrorMode)
		if em != "fail" && em != "empty" {
			return fmt.Errorf("invalid errorMode: %s", impl.Response.ErrorMode)
		}
	}
	return nil
}
