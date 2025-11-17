package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/builtin/gotool"
	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/tool/localregistry"
	"github.com/ppipada/flexigpt-app/pkg/tool/spec"
	"github.com/ppipada/mapstore-go/jsonencdec"
)

func TestInvokeTool(t *testing.T) {
	t.Parallel()

	type (
		handlerFn func(w http.ResponseWriter, r *http.Request)
		mkToolFn  func(baseURL string) spec.HTTPToolImpl
		verifyFn  func(t *testing.T, resp *spec.InvokeToolResponse, err error)
	)

	// Common minimal tool builder: GET with query/header templating.
	defaultTool := func(baseURL, pathSuffix string) spec.HTTPToolImpl {
		return spec.HTTPToolImpl{
			Request: spec.HTTPRequest{
				Method:      "GET",
				URLTemplate: baseURL + pathSuffix,
				Query: map[string]string{
					"q":   "${msg}",
					"sec": "${SECRET}",
				},
				Headers: map[string]string{
					"X-Auth": "token ${SECRET}",
				},
			},
			Response: spec.HTTPResponse{},
		}
	}

	tests := []struct {
		name            string
		handler         handlerFn
		mkTool          mkToolFn
		args            string
		httpOptions     *spec.InvokeHTTPOptions
		disableBundle   bool
		disableTool     bool
		wantErrIs       error
		wantErrContains string
		verify          verifyFn
	}{
		{
			name: "http_get_success_with_templating_and_meta",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Echo query/header into JSON.
				q := r.URL.Query().Get("q")
				sec := r.URL.Query().Get("sec")
				auth := r.Header.Get("X-Auth")
				extra := r.Header.Get("X-Extra")
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"q":     q,
					"sec":   sec,
					"auth":  auth,
					"extra": extra,
				})
			},
			mkTool: func(baseURL string) spec.HTTPToolImpl {
				impl := defaultTool(baseURL, "/echo")
				return impl
			},
			args: `{"msg":"hello"}`,
			httpOptions: &spec.InvokeHTTPOptions{
				Secrets:      map[string]string{"SECRET": "s3cr3t"},
				ExtraHeaders: map[string]string{"X-Extra": "1"},
			},
			verify: func(t *testing.T, resp *spec.InvokeToolResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("InvokeTool error: %v", err)
				}
				if resp == nil || resp.Body == nil {
					t.Fatalf("nil response/body")
				}
				if resp.Body.IsBuiltIn {
					t.Fatalf("expected IsBuiltIn=false for user tool")
				}
				if resp.Body.Meta == nil {
					t.Fatalf("expected meta, got nil")
				}
				if typ, _ := resp.Body.Meta["type"].(string); typ != string(spec.ToolTypeHTTP) {
					t.Fatalf("meta.type = %v, want http", typ)
				}
				if _, ok := resp.Body.Meta["status"]; !ok {
					t.Fatalf("meta.status missing")
				}

				var got map[string]any
				if err := json.Unmarshal([]byte(resp.Body.Output), &got); err != nil {
					t.Fatalf("unmarshal output: %v; raw: %q", err, resp.Body.Output)
				}
				if got["q"] != "hello" {
					t.Fatalf("q = %v, want hello", got["q"])
				}
				if got["sec"] != "s3cr3t" {
					t.Fatalf("sec = %v, want s3cr3t", got["sec"])
				}
				if got["auth"] != "token s3cr3t" {
					t.Fatalf("auth = %v, want 'token s3cr3t'", got["auth"])
				}
				if got["extra"] != "1" {
					t.Fatalf("extra = %v, want 1", got["extra"])
				}
			},
		},
		{
			name: "bundle_disabled",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			},
			mkTool:        func(baseURL string) spec.HTTPToolImpl { return defaultTool(baseURL, "/ok") },
			args:          `{"msg":"x"}`,
			disableBundle: true,
			wantErrIs:     spec.ErrBundleDisabled,
		},
		{
			name: "tool_disabled",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			},
			mkTool:      func(baseURL string) spec.HTTPToolImpl { return defaultTool(baseURL, "/ok") },
			args:        `{"msg":"x"}`,
			disableTool: true,
			wantErrIs:   spec.ErrToolDisabled,
		},
		{
			name: "non_json_response_errors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				_, _ = w.Write([]byte("ok"))
			},
			mkTool:          func(baseURL string) spec.HTTPToolImpl { return defaultTool(baseURL, "/plain") },
			args:            `{}`,
			wantErrContains: "non-JSON Content-Type",
		},
		{
			name: "no_content_204_returns_null",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			mkTool: func(baseURL string) spec.HTTPToolImpl { return defaultTool(baseURL, "/nocontent") },
			args:   `{}`,
			verify: func(t *testing.T, resp *spec.InvokeToolResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("InvokeTool error: %v", err)
				}
				if resp == nil || resp.Body == nil {
					t.Fatalf("nil response/body")
				}
				if resp.Body.Output != "null" {
					t.Fatalf("Output = %q, want %q", resp.Body.Output, "null")
				}
			},
		},
		{
			name: "custom_success_codes_mismatch_errors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Return 200 while the tool requires 201.
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			},
			mkTool: func(baseURL string) spec.HTTPToolImpl {
				impl := defaultTool(baseURL, "/status")
				impl.Response.SuccessCodes = []int{http.StatusCreated}
				return impl
			},
			args:            `{}`,
			wantErrContains: "not in success set",
		},
		{
			name: "error_mode_empty_suppresses_error_and_returns_empty_output",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Return 404; errorMode: empty should suppress errors and yield empty output.
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"err": "nope"})
			},
			mkTool: func(baseURL string) spec.HTTPToolImpl {
				impl := defaultTool(baseURL, "/404")
				impl.Response.ErrorMode = "empty"
				return impl
			},
			args: `{}`,
			verify: func(t *testing.T, resp *spec.InvokeToolResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("InvokeTool error: %v", err)
				}
				if resp == nil || resp.Body == nil {
					t.Fatalf("nil response/body")
				}
				if resp.Body.Output != "" {
					t.Fatalf("Output = %q, want empty", resp.Body.Output)
				}
				if resp.Body.Meta == nil {
					t.Fatalf("meta should be present")
				}
				if st, ok := resp.Body.Meta["status"].(int); !ok || st != http.StatusNotFound {
					t.Fatalf(
						"meta.status = %v, want %d",
						resp.Body.Meta["status"],
						http.StatusNotFound,
					)
				}
			},
		},
		{
			name: "timeout_override_errors",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Sleep long enough to exceed override timeout.
				time.Sleep(150 * time.Millisecond)
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			},
			mkTool: func(baseURL string) spec.HTTPToolImpl {
				return defaultTool(baseURL, "/slow")
			},
			args: `{}`,
			httpOptions: &spec.InvokeHTTPOptions{
				TimeoutMs: 20,
			},
			wantErrContains: "http.Do", // generic marker; actual error may vary
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Test HTTP server for this case.
			srv := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer srv.Close()

			baseDir := t.TempDir()
			ts, err := NewToolStore(baseDir, WithFTS(false))
			if err != nil {
				t.Fatalf("NewToolStore: %v", err)
			}
			defer ts.Close()

			// Create bundle and tool.
			const (
				bundleID   = bundleitemutils.BundleID("bundle-1")
				bundleSlug = bundleitemutils.BundleSlug("bundle1")
				toolSlug   = bundleitemutils.ItemSlug("tool1")
				version    = bundleitemutils.ItemVersion("v1")
			)

			if _, err := ts.PutToolBundle(context.Background(), &spec.PutToolBundleRequest{
				BundleID: bundleID,
				Body: &spec.PutToolBundleRequestBody{
					Slug:        bundleSlug,
					DisplayName: "Bundle 1",
					IsEnabled:   true,
					Description: "test bundle",
				},
			}); err != nil {
				t.Fatalf("PutToolBundle: %v", err)
			}

			// Tool (HTTP-only custom tools).
			impl := tc.mkTool(srv.URL)
			if _, err := ts.PutTool(context.Background(), &spec.PutToolRequest{
				BundleID: bundleID,
				ToolSlug: toolSlug,
				Version:  version,
				Body: &spec.PutToolRequestBody{
					DisplayName:  "Tool 1",
					Description:  "test tool",
					Tags:         []string{"t"},
					IsEnabled:    true,
					ArgSchema:    "{}",
					OutputSchema: "{}",
					Type:         spec.ToolTypeHTTP,
					HTTP:         &impl,
				},
			}); err != nil {
				t.Fatalf("PutTool: %v", err)
			}

			// Optional: disable bundle or tool prior to invocation.
			if tc.disableBundle {
				if _, err := ts.PatchToolBundle(context.Background(), &spec.PatchToolBundleRequest{
					BundleID: bundleID,
					Body:     &spec.PatchToolBundleRequestBody{IsEnabled: false},
				}); err != nil {
					t.Fatalf("PatchToolBundle: %v", err)
				}
			}
			if tc.disableTool {
				if _, err := ts.PatchTool(context.Background(), &spec.PatchToolRequest{
					BundleID: bundleID,
					ToolSlug: toolSlug,
					Version:  version,
					Body:     &spec.PatchToolRequestBody{IsEnabled: false},
				}); err != nil {
					t.Fatalf("PatchTool: %v", err)
				}
			}

			resp, err := ts.InvokeTool(context.Background(), &spec.InvokeToolRequest{
				BundleID: bundleID,
				ToolSlug: toolSlug,
				Version:  version,
				Body: &spec.InvokeToolRequestBody{
					Args:        tc.args,
					HTTPOptions: tc.httpOptions,
				},
			})

			if tc.wantErrIs != nil {
				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("InvokeTool error = %v, want errors.Is(..., %v)", err, tc.wantErrIs)
				}
				return
			}
			if tc.wantErrContains != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Fatalf("InvokeTool err = %v, want contains %q", err, tc.wantErrContains)
				}
				return
			}

			if tc.verify != nil {
				tc.verify(t, resp, err)
				return
			}

			if err != nil {
				t.Fatalf("InvokeTool error: %v", err)
			}
			if resp == nil || resp.Body == nil {
				t.Fatalf("nil response/body")
			}
			if resp.Body.Meta == nil {
				t.Fatalf("meta is nil")
			}
			if typ, _ := resp.Body.Meta["type"].(string); typ != string(spec.ToolTypeHTTP) {
				t.Fatalf("meta.type = %v, want http", typ)
			}
		})
	}
}

func TestInvokeTool_InvalidRequest(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	ts, err := NewToolStore(baseDir, WithFTS(false))
	if err != nil {
		t.Fatalf("NewToolStore: %v", err)
	}
	defer ts.Close()

	// Missing required fields.
	_, err = ts.InvokeTool(context.Background(), &spec.InvokeToolRequest{
		BundleID: "",
		ToolSlug: "",
		Version:  "",
		Body:     &spec.InvokeToolRequestBody{Args: "{}"},
	})
	if !errors.Is(err, spec.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

func TestInvokeTool_RequestBodyTemplating_PathQueryHeaderAuth(t *testing.T) {
	t.Parallel()

	// Handler asserts templated values for path, query, header, and Authorization.
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Path element is last segment.
		gotID := path.Base(r.URL.Path)
		if gotID == "" {
			t.Fatalf("empty path base")
		}
		qv := r.URL.Query().Get("id")
		if qv == "" {
			t.Fatalf("missing query 'id'")
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Fatalf("Authorization header unexpected: %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"pathID":  gotID,
			"queryID": qv,
			"auth":    auth,
		})
	}

	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()

	baseDir := t.TempDir()
	ts, err := NewToolStore(baseDir, WithFTS(false))
	if err != nil {
		t.Fatalf("NewToolStore: %v", err)
	}
	defer ts.Close()

	const (
		bundleID   = bundleitemutils.BundleID("bundle-2")
		bundleSlug = bundleitemutils.BundleSlug("bundle2")
		toolSlug   = bundleitemutils.ItemSlug("tool2")
		version    = bundleitemutils.ItemVersion("v1")
	)

	_, err = ts.PutToolBundle(context.Background(), &spec.PutToolBundleRequest{
		BundleID: bundleID,
		Body: &spec.PutToolBundleRequestBody{
			Slug:        bundleSlug,
			DisplayName: "Bundle 2",
			IsEnabled:   true,
		},
	})
	if err != nil {
		t.Fatalf("PutToolBundle: %v", err)
	}

	impl := spec.HTTPToolImpl{
		Request: spec.HTTPRequest{
			Method:      "GET",
			URLTemplate: srv.URL + "/items/${id}",
			Query:       map[string]string{"id": "${id}"},
			Headers:     map[string]string{"X-Trace": "${id}"},
			Auth: &spec.HTTPAuth{
				Type:          "bearer",
				ValueTemplate: "${SECRET}",
			},
		},
		Response: spec.HTTPResponse{},
	}

	_, err = ts.PutTool(context.Background(), &spec.PutToolRequest{
		BundleID: bundleID,
		ToolSlug: toolSlug,
		Version:  version,
		Body: &spec.PutToolRequestBody{
			DisplayName:  "Tool 2",
			IsEnabled:    true,
			ArgSchema:    "{}",
			OutputSchema: "{}",
			Type:         spec.ToolTypeHTTP,
			HTTP:         &impl,
		},
	})
	if err != nil {
		t.Fatalf("PutTool: %v", err)
	}

	resp, err := ts.InvokeTool(context.Background(), &spec.InvokeToolRequest{
		BundleID: bundleID,
		ToolSlug: toolSlug,
		Version:  version,
		Body: &spec.InvokeToolRequestBody{
			Args:        `{"id":"42"}`,
			HTTPOptions: &spec.InvokeHTTPOptions{Secrets: map[string]string{"SECRET": "abc123"}},
		},
	})
	if err != nil {
		t.Fatalf("InvokeTool: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(resp.Body.Output), &got); err != nil {
		t.Fatalf("unmarshal output: %v; raw: %q", err, resp.Body.Output)
	}
	if got["pathID"] != "42" || got["queryID"] != "42" {
		t.Fatalf("templating mismatch: got=%v", got)
	}
	if a, _ := got["auth"].(string); a != "Bearer abc123" {
		t.Fatalf("auth = %v, want 'Bearer abc123'", got["auth"])
	}
}

// TestInvokeTool_Go_CustomRegistered covers invoking user-created Go tools
// by directly inserting Tool records (type=go) into the directory-store.
// We bypass PutTool because it only accepts custom HTTP tools.
func TestInvokeGoCustomRegistered(t *testing.T) {
	t.Parallel()

	type verifyFn func(t *testing.T, resp *spec.InvokeToolResponse, err error)

	tests := []struct {
		name            string
		register        func(t *testing.T) string // returns funcName registered in the default registry
		funcName        string                    // if register is nil, allows unknown func-name case
		args            string
		disableBundle   bool
		disableTool     bool
		wantErrIs       error
		wantErrContains string
		verify          verifyFn
	}{
		{
			name: "echo_success",
			register: func(t *testing.T) string {
				t.Helper()
				funcName := "github.com/ppipada/flexigpt-app/tests/" + sanitizeID(t.Name())
				type Args struct {
					Name  string `json:"name"`
					Times int    `json:"times"`
				}
				type Out struct {
					Msg string `json:"msg"`
				}
				fn := func(ctx context.Context, a Args) (Out, error) {
					if a.Times <= 0 {
						a.Times = 1
					}
					var b strings.Builder
					for i := 0; i < a.Times; i++ {
						b.WriteString(a.Name)
					}
					return Out{Msg: b.String()}, nil
				}
				if err := localregistry.RegisterTyped(localregistry.DefaultGoRegistry, funcName, fn); err != nil {
					t.Fatalf("RegisterTyped: %v", err)
				}
				return funcName
			},
			args: `{"name":"ab","times":3}`,
			verify: func(t *testing.T, resp *spec.InvokeToolResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("InvokeTool error: %v", err)
				}
				if resp == nil || resp.Body == nil {
					t.Fatalf("nil response/body")
				}
				if resp.Body.Meta == nil || resp.Body.Meta["type"] != "go" {
					t.Fatalf("expected go meta")
				}
				type Out struct {
					Msg string `json:"msg"`
				}
				var out Out
				if err := json.Unmarshal([]byte(resp.Body.Output), &out); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if out.Msg != "ababab" {
					t.Fatalf("msg = %q, want %q", out.Msg, "ababab")
				}
			},
		},
		{
			name: "invalid_input_type_rejected",
			register: func(t *testing.T) string {
				t.Helper()
				funcName := "github.com/ppipada/flexigpt-app/tests/" + sanitizeID(t.Name())
				type Args struct {
					Times int `json:"times"`
				}
				type Out struct {
					Ok bool `json:"ok"`
				}
				fn := func(ctx context.Context, a Args) (Out, error) {
					return Out{Ok: a.Times > 0}, nil
				}
				if err := localregistry.RegisterTyped(localregistry.DefaultGoRegistry, funcName, fn); err != nil {
					t.Fatalf("RegisterTyped: %v", err)
				}
				return funcName
			},
			args:            `{"times":"oops"}`, // wrong type for int -> strict decode should fail
			wantErrContains: "invalid input",
		},
		{
			name: "function_returns_error",
			register: func(t *testing.T) string {
				t.Helper()
				funcName := "github.com/ppipada/flexigpt-app/tests/" + sanitizeID(t.Name())
				type Args struct {
					Fail bool `json:"fail"`
				}
				type Out struct{}
				fn := func(ctx context.Context, a Args) (Out, error) {
					if a.Fail {
						return Out{}, errors.New("boom")
					}
					return Out{}, nil
				}
				if err := localregistry.RegisterTyped(localregistry.DefaultGoRegistry, funcName, fn); err != nil {
					t.Fatalf("RegisterTyped: %v", err)
				}
				return funcName
			},
			args:            `{"fail":true}`,
			wantErrContains: "boom",
		},
		{
			name:            "unknown_function",
			register:        nil,
			funcName:        "github.com/ppipada/flexigpt-app/tests/unknownFunc",
			args:            `{}`,
			wantErrContains: "unknown tool",
		},
		{
			name: "respects_context_timeout",
			register: func(t *testing.T) string {
				t.Helper()
				funcName := "github.com/ppipada/flexigpt-app/tests/" + sanitizeID(t.Name())
				type Args struct {
					SleepMs int `json:"sleepMs"`
				}
				type Out struct {
					Ok bool `json:"ok"`
				}
				fn := func(ctx context.Context, a Args) (Out, error) {
					select {
					case <-ctx.Done():
						return Out{}, ctx.Err()
					case <-time.After(time.Duration(a.SleepMs) * time.Millisecond):
						return Out{Ok: true}, nil
					}
				}
				if err := localregistry.RegisterTyped(localregistry.DefaultGoRegistry, funcName, fn); err != nil {
					t.Fatalf("RegisterTyped: %v", err)
				}
				return funcName
			},
			args:            `{"sleepMs":200}`,
			wantErrContains: "context deadline exceeded",
		},
		{
			name: "bundle_disabled",
			register: func(t *testing.T) string {
				t.Helper()
				funcName := "github.com/ppipada/flexigpt-app/tests/" + sanitizeID(t.Name())
				type Args struct{}
				type Out struct{}
				fn := func(ctx context.Context, a Args) (Out, error) { return Out{}, nil }
				if err := localregistry.RegisterTyped(localregistry.DefaultGoRegistry, funcName, fn); err != nil {
					t.Fatalf("RegisterTyped: %v", err)
				}
				return funcName
			},
			args:          `{}`,
			disableBundle: true,
			wantErrIs:     spec.ErrBundleDisabled,
		},
		{
			name: "tool_disabled",
			register: func(t *testing.T) string {
				t.Helper()
				funcName := "github.com/ppipada/flexigpt-app/tests/" + sanitizeID(t.Name())
				type Args struct{}
				type Out struct{}
				fn := func(ctx context.Context, a Args) (Out, error) { return Out{}, nil }
				if err := localregistry.RegisterTyped(localregistry.DefaultGoRegistry, funcName, fn); err != nil {
					t.Fatalf("RegisterTyped: %v", err)
				}
				return funcName
			},
			args:        `{}`,
			disableTool: true,
			wantErrIs:   spec.ErrToolDisabled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			baseDir := t.TempDir()
			ts, err := NewToolStore(baseDir, WithFTS(false))
			if err != nil {
				t.Fatalf("NewToolStore: %v", err)
			}
			defer ts.Close()

			// Create a dedicated bundle per sub-test.
			bundleID := bundleitemutils.BundleID("b-" + sanitizeID(t.Name()))
			bundleSlug := bundleitemutils.BundleSlug("bundle-" + sanitizeID(t.Name()))

			putBundle(t, ts, bundleID, bundleSlug, true)

			// Register or use provided function name.
			funcName := tc.funcName
			if tc.register != nil {
				funcName = tc.register(t)
			}

			// Insert a Go tool (bypass PutTool).
			toolSlug := bundleitemutils.ItemSlug("go-" + sanitizeID(t.Name()))
			version := bundleitemutils.ItemVersion("v1")
			if err := addGoToolFile(t, ts, bundleID, bundleSlug, toolSlug, version, funcName, true); err != nil {
				t.Fatalf("addGoToolFile: %v", err)
			}

			// Optional disable bundle/tool.
			if tc.disableBundle {
				if _, err := ts.PatchToolBundle(context.Background(), &spec.PatchToolBundleRequest{
					BundleID: bundleID,
					Body:     &spec.PatchToolBundleRequestBody{IsEnabled: false},
				}); err != nil {
					t.Fatalf("PatchToolBundle: %v", err)
				}
			}
			if tc.disableTool {
				if _, err := ts.PatchTool(context.Background(), &spec.PatchToolRequest{
					BundleID: bundleID,
					ToolSlug: toolSlug,
					Version:  version,
					Body:     &spec.PatchToolRequestBody{IsEnabled: false},
				}); err != nil {
					t.Fatalf("PatchTool: %v", err)
				}
			}

			// Build invocation context (custom timeout for the timeout scenario).
			ctx := context.Background()
			if strings.Contains(tc.name, "timeout") {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()
			}

			resp, err := ts.InvokeTool(ctx, &spec.InvokeToolRequest{
				BundleID: bundleID,
				ToolSlug: toolSlug,
				Version:  version,
				Body: &spec.InvokeToolRequestBody{
					Args: tc.args,
					// Note: GoOptions currently not used by InvokeTool; context controls timeout.
				},
			})

			if tc.wantErrIs != nil {
				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("err = %v, want errors.Is(..., %v)", err, tc.wantErrIs)
				}
				return
			}
			if tc.wantErrContains != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Fatalf("err = %v, want contains %q", err, tc.wantErrContains)
				}
				return
			}
			if tc.verify != nil {
				tc.verify(t, resp, err)
				return
			}
			if err != nil {
				t.Fatalf("InvokeTool error: %v", err)
			}
			if resp == nil || resp.Body == nil || resp.Body.Meta["type"] != "go" {
				t.Fatalf("unexpected response/meta: %+v", resp)
			}
		})
	}
}

// TestInvokeTool_Go_BuiltIns exercises the inbuilt Go tools you provided:
// - ReadFile
// - ListDirectory
// - SearchFiles
// We synthesize user Tool records that reference those registered function IDs.
func TestInvokeTool_Go_BuiltIns(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	// Create some files for read/list/search.
	fileA := filepath.Join(tmp, "a.txt")
	fileB := filepath.Join(tmp, "b.md")
	if err := os.WriteFile(fileA, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}
	if err := os.WriteFile(fileB, []byte("lorem ipsum"), 0o600); err != nil {
		t.Fatalf("write b.md: %v", err)
	}

	tests := []struct {
		name     string
		funcName string
		args     string
		verify   func(t *testing.T, out json.RawMessage)
	}{
		{
			name:     "ReadFile_text_success",
			funcName: gotool.ReadFileFuncID,
			args:     fmt.Sprintf(`{"path":%q,"encoding":"text"}`, fileA),
			verify: func(t *testing.T, out json.RawMessage) {
				t.Helper()
				var o gotool.ReadFileOut
				if err := json.Unmarshal(out, &o); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if o.Content != "hello world" {
					t.Fatalf("content=%q, want %q", o.Content, "hello world")
				}
			},
		},
		{
			name:     "ListDirectory_glob",
			funcName: gotool.ListDirectoryFuncID,
			args:     fmt.Sprintf(`{"path":%q,"pattern":"*.txt"}`, tmp),
			verify: func(t *testing.T, out json.RawMessage) {
				t.Helper()
				var o gotool.ListDirectoryOut
				if err := json.Unmarshal(out, &o); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				has := func(name string) bool {
					return slices.Contains(o.Entries, name)
				}
				if !has("a.txt") || has("b.md") {
					t.Fatalf("entries=%v, expected only a.txt with pattern", o.Entries)
				}
			},
		},
		{
			name:     "SearchFiles_regex",
			funcName: gotool.SearchFilesFuncID,
			args:     fmt.Sprintf(`{"root":%q,"pattern":"hello"}`, tmp),
			verify: func(t *testing.T, out json.RawMessage) {
				t.Helper()
				var o gotool.SearchFilesOut
				if err := json.Unmarshal(out, &o); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				found := slices.Contains(o.Matches, fileA)
				if !found {
					t.Fatalf("expected to find %q in matches; got %v", fileA, o.Matches)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			baseDir := t.TempDir()
			ts, err := NewToolStore(baseDir, WithFTS(false))
			if err != nil {
				t.Fatalf("NewToolStore: %v", err)
			}
			defer ts.Close()

			// Bundle per case.
			bundleID := bundleitemutils.BundleID("b-" + sanitizeID(t.Name()))
			bundleSlug := bundleitemutils.BundleSlug("bundle-" + sanitizeID(t.Name()))
			t.Logf("slug: %s", bundleSlug)
			putBundle(t, ts, bundleID, bundleSlug, true)

			// Write the Go tool referencing the built-in function name.
			toolSlug := bundleitemutils.ItemSlug("go-" + sanitizeID(t.Name()))
			version := bundleitemutils.ItemVersion("v1")
			if err := addGoToolFile(t, ts, bundleID, bundleSlug, toolSlug, version, tc.funcName, true); err != nil {
				t.Fatalf("addGoToolFile: %v", err)
			}

			resp, err := ts.InvokeTool(context.Background(), &spec.InvokeToolRequest{
				BundleID: bundleID,
				ToolSlug: toolSlug,
				Version:  version,
				Body: &spec.InvokeToolRequestBody{
					Args: tc.args,
				},
			})
			if err != nil {
				t.Fatalf("InvokeTool: %v", err)
			}
			if resp == nil || resp.Body == nil {
				t.Fatalf("nil response/body")
			}
			tc.verify(t, json.RawMessage(resp.Body.Output))
		})
	}
}

// addGoToolFile inserts a Go-type Tool record directly into the directory store.
// This bypasses PutTool (which only accepts HTTP tools) so we can exercise the Go path.
func addGoToolFile(
	t *testing.T,
	ts *ToolStore,
	bundleID bundleitemutils.BundleID,
	bundleSlug bundleitemutils.BundleSlug,
	toolSlug bundleitemutils.ItemSlug,
	version bundleitemutils.ItemVersion,
	funcName string,
	enabled bool,
) error {
	t.Helper()

	dirInfo, err := bundleitemutils.BuildBundleDir(bundleID, bundleSlug)
	if err != nil {
		return err
	}
	finf, err := bundleitemutils.BuildItemFileInfo(toolSlug, version)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	tool := spec.Tool{
		SchemaVersion: spec.SchemaVersion,
		ID:            bundleitemutils.ItemID("go-" + string(toolSlug) + "-" + string(version)),
		Slug:          toolSlug,
		Version:       version,
		DisplayName:   "Go Tool " + string(toolSlug),
		Description:   "test go tool",
		Tags:          []string{"go"},
		ArgSchema:     json.RawMessage(`{}`),
		OutputSchema:  json.RawMessage(`{}`),
		Type:          spec.ToolTypeGo,
		GoImpl:        &spec.GoToolImpl{Func: funcName},
		HTTP:          nil,
		IsEnabled:     enabled,
		IsBuiltIn:     false,
		CreatedAt:     now,
		ModifiedAt:    now,
	}
	if err := validateTool(&tool); err != nil {
		return fmt.Errorf("validateTool: %w", err)
	}

	mp, _ := jsonencdec.StructWithJSONTagsToMap(tool)
	return ts.toolStore.SetFileData(
		bundleitemutils.GetBundlePartitionFileKey(finf.FileName, dirInfo.DirName),
		mp,
	)
}

func putBundle(
	t *testing.T,
	ts *ToolStore,
	id bundleitemutils.BundleID,
	slug bundleitemutils.BundleSlug,
	enabled bool,
) {
	t.Helper()
	if _, err := ts.PutToolBundle(context.Background(), &spec.PutToolBundleRequest{
		BundleID: id,
		Body: &spec.PutToolBundleRequestBody{
			Slug:        slug,
			DisplayName: "bundle " + string(slug),
			IsEnabled:   enabled,
		},
	}); err != nil {
		t.Fatalf("PutToolBundle: %v", err)
	}
}

// sanitizeID produces a slug-ish identifier from a test name for use in IDs, slugs and func names.
func sanitizeID(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
