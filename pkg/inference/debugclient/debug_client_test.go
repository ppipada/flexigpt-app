package debugclient

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ppipada/flexigpt-app/pkg/inference/spec"
)

// TestContainsSensitiveKey verifies detection of sensitive keys.
func TestContainsSensitiveKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{
			name: "EmptyKeyIsNotSensitive.",
			key:  "",
			want: false,
		},
		{
			name: "AuthorizationIsSensitive.",
			key:  "Authorization",
			want: true,
		},
		{
			name: "ProxyAuthorizationIsSensitive.",
			key:  "Proxy-Authorization",
			want: true,
		},
		{
			name: "ApiKeyIsSensitive.",
			key:  "apiKey",
			want: true,
		},
		{
			name: "XApiKeyIsSensitive.",
			key:  "X-API-KEY",
			want: true,
		},
		{
			name: "SubstringKeyIsSensitive_Monkey.",
			key:  "monkey",
			want: true,
		},
		{
			name: "SubstringKeyIsSensitive_TurKey.",
			key:  "turKey",
			want: true,
		},
		{
			name: "UnrelatedKeyIsNotSensitive.",
			key:  "NotSensitive",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := containsSensitiveKey(tc.key)
			if got != tc.want {
				t.Fatalf("containsSensitiveKey(%q) = %v, want = %v.", tc.key, got, tc.want)
			}
		})
	}
}

// TestRedactHeaders_NilAndEmptyInput verifies behavior for nil and empty inputs.
func TestRedactHeaders_NilAndEmptyInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]any
		want map[string]any
	}{
		{
			name: "NilInputReturnsNil.",
			in:   nil,
			want: nil,
		},
		{
			name: "EmptyMapReturnsEmptyMap.",
			in:   map[string]any{},
			want: map[string]any{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := redactHeaders(tc.in)

			if tc.in == nil {
				if got != nil {
					t.Fatalf("got = %#v, want = nil.", got)
				}
				return
			}

			if got == nil {
				t.Fatalf("got = nil, want = %#v.", tc.want)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got = %#v, want = %#v.", got, tc.want)
			}
		})
	}
}

// TestRedactHeaders_BasicRedaction verifies redaction of sensitive keys in headers.
func TestRedactHeaders_BasicRedaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]any
		want map[string]any
	}{
		{
			name: "NoSensitiveKeysRemainUnchanged.",
			in: map[string]any{
				"Make":  "Honda",
				"Model": "Civic",
				"Year":  2020,
				"New":   true,
			},
			want: map[string]any{
				"Make":  "Honda",
				"Model": "Civic",
				"Year":  2020,
				"New":   true,
			},
		},
		{
			name: "SensitiveKeysAreMaskedCaseInsensitiveAndSubstring.",
			in: map[string]any{
				"Authorization": "Bearer secret",
				"apiKey":        "abc123",
				"monkey":        "banana",
				"turKey":        "sandwich",
				"Key":           "value",
				"VIN":           "JH4DA9350LS000001",
				"Spare":         nil,
			},
			want: map[string]any{
				"Authorization": maskToken,
				"apiKey":        maskToken,
				"monkey":        maskToken,
				"turKey":        maskToken,
				"Key":           maskToken,
				"VIN":           "JH4DA9350LS000001",
				"Spare":         nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := redactHeaders(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got = %#v, want = %#v.", got, tc.want)
			}
		})
	}
}

// TestSanitizeBodyForDebug_PlainText verifies behaviour for non-JSON bodies.
func TestSanitizeBodyForDebug_PlainText(t *testing.T) {
	t.Parallel()

	longBase64 := strings.Repeat("A", 128) // Valid-looking base64, length >= 128 and %4 == 0.

	tests := []struct {
		name        string
		raw         string
		cfg         DebugConfig
		want        any
		description string
	}{
		{
			name: "PlainTextWithoutStripReturnsAsIs.",
			raw:  "hello world",
			cfg: DebugConfig{
				StripContent: false,
			},
			want:        "hello world",
			description: "No JSON, no strip => return raw string.",
		},
		{
			name: "PlainTextWithStripReturnsAsIsIfNotBase64.",
			raw:  "not base64",
			cfg: DebugConfig{
				StripContent: true,
			},
			want:        "not base64",
			description: "StripContent=true but not base64-like => keep text.",
		},
		{
			name: "LongBase64WithStripIsOmitted.",
			raw:  longBase64,
			cfg: DebugConfig{
				StripContent: true,
			},
			want: fmt.Sprintf(
				"[omitted: %d bytes base64 data]",
				len(longBase64),
			),
			description: "Base64-like string is replaced by placeholder.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := sanitizeBodyForDebug([]byte(tc.raw), true, tc.cfg)
			gotStr, ok := got.(string)
			if !ok {
				t.Fatalf("expected string, got %T: %#v.", got, got)
			}
			if gotStr != tc.want {
				t.Fatalf("got = %q, want = %q. %s", gotStr, tc.want, tc.description)
			}
		})
	}
}

// TestSanitizeBodyForDebug_JSON_SensitiveKeys verifies that sensitive keys are
// masked in JSON bodies, regardless of StripContent.
func TestSanitizeBodyForDebug_JSON_SensitiveKeys(t *testing.T) {
	t.Parallel()

	const body = `{
  "apiKey": "secret",
  "nested": {
    "authorization": "Bearer token",
    "value": 42
  }
}`

	tests := []struct {
		name string
		cfg  DebugConfig
	}{
		{
			name: "StripContentFalseStillRedactsSensitiveKeys.",
			cfg: DebugConfig{
				StripContent: false,
			},
		},
		{
			name: "StripContentTrueAlsoRedactsSensitiveKeys.",
			cfg: DebugConfig{
				StripContent: true,
			},
		},
	}

	want := map[string]any{
		"apiKey": maskToken,
		"nested": map[string]any{
			"authorization": maskToken,
			// JSON numbers unmarshal as float64.
			"value": float64(42),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := sanitizeBodyForDebug([]byte(body), true, tc.cfg)
			gotMap, ok := got.(map[string]any)
			if !ok {
				t.Fatalf("expected map[string]any, got %T: %#v.", got, got)
			}
			if !reflect.DeepEqual(gotMap, want) {
				t.Fatalf("got = %#v, want = %#v.", gotMap, want)
			}
		})
	}
}

// TestSanitizeBodyForDebug_JSON_MessageContent verifies that user/assistant
// message content is scrubbed only when StripContent is true.
func TestSanitizeBodyForDebug_JSON_MessageContent(t *testing.T) {
	t.Parallel()

	const body = `{
  "role": "user",
  "content": "Hello world",
  "extra": "ok"
}`

	tests := []struct {
		name        string
		cfg         DebugConfig
		wantContent string
		wantExtra   string
		description string
	}{
		{
			name: "StripContentFalseKeepsMessageText.",
			cfg: DebugConfig{
				StripContent: false,
			},
			wantContent: "Hello world",
			wantExtra:   "ok",
			description: "Message content is not scrubbed when StripContent=false.",
		},
		{
			name: "StripContentTrueScrubsMessageText.",
			cfg: DebugConfig{
				StripContent: true,
			},
			wantContent: ommitedTextContentStr,
			wantExtra:   "ok",
			description: "Message content is scrubbed when StripContent=true.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := sanitizeBodyForDebug([]byte(body), true, tc.cfg)
			gotMap, ok := got.(map[string]any)
			if !ok {
				t.Fatalf("expected map[string]any, got %T: %#v.", got, got)
			}

			content, _ := gotMap[contentStr].(string)
			if content != tc.wantContent {
				t.Fatalf(
					"content got = %q, want = %q. %s",
					content,
					tc.wantContent,
					tc.description,
				)
			}

			extra, _ := gotMap["extra"].(string)
			if extra != tc.wantExtra {
				t.Fatalf(
					"extra got = %q, want = %q. %s",
					extra,
					tc.wantExtra,
					tc.description,
				)
			}
		})
	}
}

// TestSanitizeBodyForDebug_JSON_StructuredContent verifies that structured
// content segments preserve metadata but scrub text and base64 data.
func TestSanitizeBodyForDebug_JSON_StructuredContent(t *testing.T) {
	t.Parallel()

	base64Data := strings.Repeat("A", 128)

	// Build JSON from a Go value to avoid float64 surprises.
	input := map[string]any{
		"role": "assistant",
		contentStr: []any{
			map[string]any{
				"type":  textStr,
				textStr: "Segment text",
			},
			map[string]any{
				"type":      "input_image",
				"image_url": "https://example.com/image.png",
				"data":      base64Data,
			},
		},
	}
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal test input: %v.", err)
	}

	cfg := DebugConfig{
		StripContent: true,
	}

	got := sanitizeBodyForDebug(raw, false, cfg)
	gotMap, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T: %#v.", got, got)
	}

	// Check role.
	if role, _ := gotMap["role"].(string); role != "assistant" {
		t.Fatalf("role got = %q, want = %q.", role, "assistant")
	}

	contentAny, ok := gotMap[contentStr]
	if !ok {
		t.Fatalf("content key missing in result: %#v.", gotMap)
	}
	contentSlice, ok := contentAny.([]any)
	if !ok {
		t.Fatalf("content is %T, want []any: %#v.", contentAny, contentAny)
	}
	if len(contentSlice) != 2 {
		t.Fatalf("content length got = %d, want = 2.", len(contentSlice))
	}

	// First segment: type=text, text should be scrubbed.
	seg0, ok := contentSlice[0].(map[string]any)
	if !ok {
		t.Fatalf("content[0] is %T, want map[string]any: %#v.", contentSlice[0], contentSlice[0])
	}

	if segType, _ := seg0["type"].(string); segType != textStr {
		t.Fatalf("segment[0].type got = %q, want = %q.", segType, textStr)
	}

	if text, _ := seg0[textStr].(string); text != ommitedTextContentStr {
		t.Fatalf("segment[0].text got = %q, want = %q.", text, ommitedTextContentStr)
	}

	// Second segment: type=input_image, image_url preserved, base64 data scrubbed.
	seg1, ok := contentSlice[1].(map[string]any)
	if !ok {
		t.Fatalf("content[1] is %T, want map[string]any: %#v.", contentSlice[1], contentSlice[1])
	}
	if segType, _ := seg1["type"].(string); segType != "input_image" {
		t.Fatalf("segment[1].type got = %q, want = %q.", segType, "input_image")
	}
	if url, _ := seg1["image_url"].(string); url != "https://example.com/image.png" {
		t.Fatalf("segment[1].image_url got = %q, want = %q.", url, "https://example.com/image.png")
	}
	data, _ := seg1["data"].(string)
	wantData := fmt.Sprintf("[omitted: %d bytes base64 data]", len(base64Data))
	if data != wantData {
		t.Fatalf("segment[1].data got = %q, want = %q.", data, wantData)
	}
}

// TestScrubber_Cycles verifies that cyclic references are handled safely and
// produce the cycleToken ("<cycle>").
func TestScrubber_Cycles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   func() map[string]any
		want map[string]any
	}{
		{
			name: "MapSelfCycleYieldsCycleToken.",
			in: func() map[string]any {
				vehicle := map[string]any{
					"Make": "Toyota",
					"VIN":  "JT2JA82J1R0000001",
				}
				vehicle["Self"] = vehicle
				return map[string]any{"Vehicle": vehicle}
			},
			want: map[string]any{
				"Vehicle": map[string]any{
					"Make": "Toyota",
					"VIN":  "JT2JA82J1R0000001",
					"Self": cycleToken,
				},
			},
		},
		{
			name: "SliceSelfCycleYieldsCycleToken.",
			in: func() map[string]any {
				garage := []any{"BMW"}
				garage = append(garage, nil)
				garage[1] = garage // Self-reference.
				return map[string]any{"Garage": garage}
			},
			want: map[string]any{
				"Garage": []any{"BMW", cycleToken},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orig := tc.in()
			s := newScrubber(DebugConfig{StripContent: true}, true)
			gotAny := s.scrub(orig, 0, scrubContext{})
			got, ok := gotAny.(map[string]any)
			if !ok {
				t.Fatalf("expected map[string]any, got %T: %#v.", gotAny, gotAny)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got = %#v, want = %#v.", got, tc.want)
			}
		})
	}
}

// TestScrubber_Immutability verifies that input is not mutated and output is a deep copy.
func TestScrubber_Immutability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   func() map[string]any
	}{
		{
			name: "DeepCopyPreventsMutationLeaks.",
			in: func() map[string]any {
				return map[string]any{
					"Car": map[string]any{
						"Make":  "Honda",
						"Model": "Civic",
						"Key":   "topSecret",
					},
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orig := tc.in()
			s := newScrubber(DebugConfig{StripContent: true}, true)
			cleanAny := s.scrub(orig, 0, scrubContext{})
			clean, ok := cleanAny.(map[string]any)
			if !ok {
				t.Fatalf("expected map[string]any, got %T: %#v.", cleanAny, cleanAny)
			}

			// Mutate original after sanitization.
			origCar, _ := orig["Car"].(map[string]any)
			origCar["Model"] = "Accord"
			origCar["Key"] = "changed"

			// Ensure the sanitized copy did not change.
			cleanCar, _ := clean["Car"].(map[string]any)
			if got, want := cleanCar["Model"], any("Civic"); got != want {
				t.Fatalf("sanitized copy mutated: Model got = %v, want = %v.", got, want)
			}
			if got, want := cleanCar["Key"], any(maskToken); got != want {
				t.Fatalf("sanitized copy mutated: Key got = %v, want = %v.", got, want)
			}

			// Mutate the sanitized copy and ensure the original did not change.
			cleanCar["Model"] = "Integra"
			if got, want := origCar["Model"], any("Accord"); got != want {
				t.Fatalf(
					"original mutated by changing sanitized copy: got = %v, want = %v.",
					got,
					want,
				)
			}
		})
	}
}

// TestGenerateCurlCommand_Basic verifies that generateCurlCommand produces a
// roughly copy-pasteable curl command that includes redacted headers and JSON body.
func TestGenerateCurlCommand_Basic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		detail *spec.APIRequestDetails
	}{
		{
			name: "BasicPostWithJsonBodyAndHeaders.",
			detail: func() *spec.APIRequestDetails {
				urlStr := "https://api.example.com/v1/test"
				method := "POST"
				return &spec.APIRequestDetails{
					URL:    &urlStr,
					Method: &method,
					Headers: map[string]any{
						"Authorization": maskToken, // Already redacted.
						"X-Custom":      "value",
					},
					Data: map[string]any{
						"foo": "bar",
					},
				}
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := DebugConfig{}
			curl := generateCurlCommand(tc.detail, cfg)

			if !strings.HasPrefix(curl, "curl") {
				t.Fatalf("curl command must start with 'curl', got: %q.", curl)
			}

			if !strings.Contains(curl, " -X POST") {
				t.Fatalf("curl command must contain method '-X POST', got: %q.", curl)
			}

			if tc.detail.URL != nil && !strings.Contains(curl, *tc.detail.URL) {
				t.Fatalf("curl command must contain URL %q, got: %q.", *tc.detail.URL, curl)
			}

			// Check that headers appear, with Authorization value already redacted.
			if !strings.Contains(curl, "Authorization: ***") {
				t.Fatalf("curl command must contain redacted Authorization header, got: %q.", curl)
			}
			if !strings.Contains(curl, "X-Custom: value") {
				t.Fatalf("curl command must contain X-Custom header, got: %q.", curl)
			}

			// Check that JSON body is present.
			if !strings.Contains(curl, `"foo": "bar"`) {
				t.Fatalf("curl command must include JSON body, got: %q.", curl)
			}

			// Check that --data-raw is used.
			if !strings.Contains(curl, "--data-raw") {
				t.Fatalf("curl command must use --data-raw for the body, got: %q.", curl)
			}
		})
	}
}
