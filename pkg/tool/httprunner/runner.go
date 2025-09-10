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

	"github.com/jmespath-community/go-jmespath"
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
	args map[string]any,
) (output any, metaData map[string]any, err error) {
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

	// Body.
	var body io.Reader
	if strings.TrimSpace(req.Body) != "" && methodAllowsBody(method) {
		expanded, err := expandTemplate(req.Body, args, secret)
		if err != nil {
			return nil, nil, fmt.Errorf("body expansion failed: %w", err)
		}
		body = bytes.NewBufferString(expanded)
		// Default content-type if not set and body present.
		if headers.Get("Content-Type") == "" {
			headers.Set("Content-Type", "application/json; charset=utf-8")
		}
	}

	// Extra headers (override any existing).
	for k, v := range r.extraHeaders {
		headers.Set(k, v)
	}

	// Build request.
	httpReq, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, nil, fmt.Errorf("http.NewRequest: %w", err)
	}
	httpReq.Header = headers
	// Default Accept for JSON encoding.
	encoding := resp.Encoding
	if encoding == "" {
		encoding = spec.DefaultHTTPEncoding
	}
	if encoding == spec.JSONEncoding && httpReq.Header.Get("Accept") == "" {
		httpReq.Header.Set("Accept", "application/json")
	}

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
		if errorMode == "empty" {
			return nil, metaData, nil
		}
		return nil, metaData, fmt.Errorf("http status %d not in success set", httpResp.StatusCode)
	}

	// Decode per encoding.
	switch strings.ToLower(encoding) {
	case spec.JSONEncoding:
		var jd any
		if len(data) > 0 {
			if err := json.Unmarshal(data, &jd); err != nil {
				return nil, metaData, fmt.Errorf("json decode failed: %w", err)
			}
		}
		if strings.TrimSpace(resp.Selector) == "" {
			output = jd
			return output, metaData, nil
		}
		// If prefixed "re:", apply regexp to compact JSON string; else treat selector as JMESPath.
		sel := strings.TrimSpace(resp.Selector)
		if strings.HasPrefix(sel, "re:") {
			compact := compactJSON(data)
			output, err = applyRegexpSelector(strings.TrimPrefix(sel, "re:"), compact)
			if err != nil {
				return nil, metaData, err
			}
			return output, metaData, nil
		}
		output, err = jmespath.Search(sel, jd)
		if err != nil {
			return nil, metaData, fmt.Errorf("selector (JMESPath) failed: %w", err)
		}
		return output, metaData, nil

	case "text":
		txt := string(data)
		if strings.TrimSpace(resp.Selector) == "" {
			return txt, metaData, nil
		}
		output, err = applyRegexpSelector(resp.Selector, txt)
		if err != nil {
			return nil, metaData, err
		}
		return output, metaData, nil

	default:
		return nil, metaData, fmt.Errorf("unsupported encoding: %s", encoding)
	}
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
			cur, ok = m[seg]
			if !ok {
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

func applyRegexpSelector(pattern, input string) (any, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return input, nil
	}
	// Accept raw pattern or /pattern/flags (flags ignored).
	if strings.HasPrefix(pattern, "/") && strings.Count(pattern, "/") >= 2 {
		last := strings.LastIndex(pattern, "/")
		pattern = pattern[1:last]
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regexp: %w", err)
	}
	m := re.FindStringSubmatch(input)
	if m == nil {
		// No match -> empty result.
		return nil, errors.New("selector did not match output")
	}
	if len(m) > 1 {
		return m[1], nil
	}
	return m[0], nil
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
	if impl.Response.Encoding != "" {
		enc := strings.ToLower(impl.Response.Encoding)
		if enc != spec.JSONEncoding && enc != spec.TextEncoding {
			return fmt.Errorf("unsupported response encoding: %s", impl.Response.Encoding)
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

func compactJSON(b []byte) string {
	var buf bytes.Buffer
	_ = json.Compact(&buf, b)
	return buf.String()
}
