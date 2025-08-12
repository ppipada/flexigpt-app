package uuidv7filename

import (
	"strings"
	"testing"
	"time"
)

const (
	validUUIDv7   = "018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e"
	fileExtension = "json"
)

func TestBuild(t *testing.T) {
	longSuffix := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec. Ipsto facto."
	sanitizedLong := "Lorem_ipsum_dolor_sit_amet__consectetur_adipiscing_elit__Donec__Ipsto_facto"
	tests := []struct {
		name        string
		id          string
		suffix      string
		extension   string
		wantName    string
		wantSuffix  string
		wantExt     string
		expectError bool
	}{
		{
			name:       "simple",
			id:         validUUIDv7,
			suffix:     "Chat",
			extension:  fileExtension,
			wantName:   validUUIDv7 + "_Chat.json",
			wantSuffix: "Chat",
			wantExt:    fileExtension,
		},
		{
			name:       "special chars",
			id:         validUUIDv7,
			suffix:     "Chat with AI!",
			extension:  fileExtension,
			wantName:   validUUIDv7 + "_Chat_with_AI_.json",
			wantSuffix: "Chat_with_AI_",
			wantExt:    fileExtension,
		},
		{
			name:       "hyphen",
			id:         validUUIDv7,
			suffix:     "Chat with-AI!",
			extension:  fileExtension,
			wantName:   validUUIDv7 + "_Chat_with-AI_.json",
			wantSuffix: "Chat_with-AI_",
			wantExt:    fileExtension,
		},
		{
			name:       "extension with dot",
			id:         validUUIDv7,
			suffix:     "Chat",
			extension:  ".json",
			wantName:   validUUIDv7 + "_Chat.json",
			wantSuffix: "Chat",
			wantExt:    fileExtension,
		},
		{
			name:       "long suffix truncated",
			id:         validUUIDv7,
			suffix:     longSuffix,
			extension:  "txt",
			wantName:   validUUIDv7 + "_" + sanitizedLong[:64] + ".txt",
			wantSuffix: sanitizedLong[:64],
			wantExt:    "txt",
		},
		{
			name:        "missing id",
			id:          "",
			suffix:      "Chat",
			extension:   fileExtension,
			expectError: true,
		},
		{
			name:        "missing suffix",
			id:          validUUIDv7,
			suffix:      "",
			extension:   fileExtension,
			expectError: true,
		},
		{
			name:        "missing extension",
			id:          validUUIDv7,
			suffix:      "Chat",
			extension:   "",
			expectError: true,
		},
		{
			name:        "invalid uuid",
			id:          "not-a-uuid",
			suffix:      "Chat",
			extension:   fileExtension,
			expectError: true,
		},
		{
			name:       "suffix with only non-alphanum",
			id:         validUUIDv7,
			suffix:     "!@#$%^&*()",
			extension:  fileExtension,
			wantName:   validUUIDv7 + "___________.json",
			wantSuffix: "__________",
			wantExt:    fileExtension,
		},
		{
			name:       "suffix with underscores",
			id:         validUUIDv7,
			suffix:     "foo_bar_baz",
			extension:  fileExtension,
			wantName:   validUUIDv7 + "_foo_bar_baz.json",
			wantSuffix: "foo_bar_baz",
			wantExt:    fileExtension,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, err := Build(tc.id, tc.suffix, tc.extension)
			if (err != nil) != tc.expectError {
				t.Fatalf("want error %v, got %v", tc.expectError, err)
			}
			if err != nil {
				return
			}
			if info.FileName != tc.wantName {
				t.Errorf("want FileName %q, got %q", tc.wantName, info.FileName)
			}
			if info.Suffix != tc.wantSuffix {
				t.Errorf("want Suffix %q, got %q", tc.wantSuffix, info.Suffix)
			}
			if info.Extension != tc.wantExt {
				t.Errorf("want Extension %q, got %q", tc.wantExt, info.Extension)
			}
			if info.ID != tc.id {
				t.Errorf("want ID %q, got %q", tc.id, info.ID)
			}
			if info.Time.IsZero() {
				t.Errorf("want non-zero Time, got zero")
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		wantID      string
		wantSuffix  string
		wantExt     string
		expectError bool
	}{
		{
			name:       "valid simple",
			filename:   validUUIDv7 + "_Chat.json",
			wantID:     validUUIDv7,
			wantSuffix: "Chat",
			wantExt:    fileExtension,
		},
		{
			name:       "valid with spaces",
			filename:   validUUIDv7 + "_Chat_with_AI_.json",
			wantID:     validUUIDv7,
			wantSuffix: "Chat with AI ",
			wantExt:    fileExtension,
		},
		{
			name:       "valid with underscores",
			filename:   validUUIDv7 + "_foo_bar_baz.json",
			wantID:     validUUIDv7,
			wantSuffix: "foo bar baz",
			wantExt:    fileExtension,
		},
		{
			name:       "valid with only underscores",
			filename:   validUUIDv7 + "__________.json",
			wantID:     validUUIDv7,
			wantSuffix: "         ",
			wantExt:    fileExtension,
		},
		{
			name:        "invalid filename no underscore",
			filename:    "whatever.json",
			expectError: true,
		},
		{
			name:        "invalid uuid",
			filename:    "not-a-uuid_Chat.json",
			expectError: true,
		},
		{
			name:       "missing extension",
			filename:   validUUIDv7 + "_Chat",
			wantID:     validUUIDv7,
			wantSuffix: "Chat",
			wantExt:    "",
		},
		{
			name:        "empty filename",
			filename:    "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, err := Parse(tc.filename)
			if (err != nil) != tc.expectError {
				t.Fatalf("want error %v, got %v", tc.expectError, err)
			}
			if err != nil {
				return
			}
			if info.ID != tc.wantID {
				t.Errorf("want ID %q, got %q", tc.wantID, info.ID)
			}
			if info.Suffix != tc.wantSuffix {
				t.Errorf("want Suffix %q, got %q", tc.wantSuffix, info.Suffix)
			}
			if info.Extension != tc.wantExt {
				t.Errorf("want Extension %q, got %q", tc.wantExt, info.Extension)
			}
			if info.FileName != tc.filename {
				t.Errorf("want FileName %q, got %q", tc.filename, info.FileName)
			}
			if info.Time.IsZero() {
				t.Errorf("want non-zero Time, got zero")
			}
		})
	}
}

func TestBuildParse_RoundTrip(t *testing.T) {
	id := validUUIDv7
	suffix := "Round Trip Chat!"
	extension := fileExtension
	expectedSuffix := "Round Trip Chat "

	info, err := Build(id, suffix, extension)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	parsed, err := Parse(info.FileName)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if parsed.ID != id {
		t.Errorf("round-trip failed: want ID %q, got %q", id, parsed.ID)
	}
	if parsed.Suffix != expectedSuffix {
		t.Errorf("round-trip failed: want Suffix %q, got %q", expectedSuffix, parsed.Suffix)
	}
	if parsed.Extension != extension {
		t.Errorf("round-trip failed: want Extension %q, got %q", extension, parsed.Extension)
	}
}

func TestBuildParse_ExtensionVariants(t *testing.T) {
	id := validUUIDv7
	suffix := "Test"
	extensions := []string{fileExtension, ".json", "txt", ".txt"}
	for _, ext := range extensions {
		info, err := Build(id, suffix, ext)
		if err != nil {
			t.Fatalf("unexpected build error for ext %q: %v", ext, err)
		}
		parsed, err := Parse(info.FileName)
		if err != nil {
			t.Fatalf("unexpected parse error for ext %q: %v", ext, err)
		}
		if parsed.Extension != strings.TrimPrefix(ext, ".") {
			t.Errorf("extension round-trip failed: input %q, got %q", ext, parsed.Extension)
		}
	}
}

func TestBuild_SuffixSanitization(t *testing.T) {
	id := validUUIDv7
	extension := fileExtension
	suffix := "a!b@c#d$e%f^g&h*i(j)k_l+m=n{o}p|q\\r:s;t'u\"v<w>x,y.z"
	wantSanitized := "a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z"
	info, err := Build(id, suffix, extension)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if info.Suffix != wantSanitized {
		t.Errorf("want sanitized Suffix %q, got %q", wantSanitized, info.Suffix)
	}
}

func TestParse_InvalidCases(t *testing.T) {
	cases := []string{
		"",
		"no_underscore",
		"not-a-uuid_suffix.json",
		"018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e.json",
	}
	for _, filename := range cases {
		_, err := Parse(filename)
		if err == nil {
			t.Errorf("expected error for filename %q, got nil", filename)
		}
	}
}

func TestBuild_InvalidUUIDv7(t *testing.T) {
	// Valid v4 UUID, not v7.
	id := "018f1e3e-7c89-4b4b-8a3b-6f8e8f8e8f8e"
	_, err := Build(id, "Chat", fileExtension)
	if err == nil {
		t.Errorf("expected error for non-v7 UUID, got nil")
	}
}

func TestBuildParse_Suffix64Limit(t *testing.T) {
	id := validUUIDv7
	extension := fileExtension
	suffix := strings.Repeat("a", 100)
	info, err := Build(id, suffix, extension)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if len(info.Suffix) != 64 {
		t.Errorf("want Suffix length 64, got %d", len(info.Suffix))
	}
	parsed, err := Parse(info.FileName)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(parsed.Suffix) != 64 {
		t.Errorf("want parsed Suffix length 64, got %d", len(parsed.Suffix))
	}
}

func TestParse_TimeExtraction(t *testing.T) {
	id := validUUIDv7
	suffix := "Chat"
	extension := fileExtension
	info, err := Build(id, suffix, extension)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	parsed, err := Parse(info.FileName)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if parsed.Time.IsZero() {
		t.Errorf("want non-zero Time, got zero")
	}
	if parsed.Time.After(time.Now().Add(1 * time.Minute)) {
		t.Errorf("parsed time is in the future: %v", parsed.Time)
	}
}
