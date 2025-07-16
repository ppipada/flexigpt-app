package store

import (
	"encoding/base64"
	"encoding/json"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

type bundlePageToken struct {
	BundleIDs       []spec.BundleID `json:"ids,omitempty"` // optional filter
	IncludeDisabled bool            `json:"d,omitempty"`   // include disabled bundles?
	PageSize        int             `json:"s"`             // page size
	CursorMod       string          `json:"t,omitempty"`   // RFC-3339-nano
	CursorID        spec.BundleID   `json:"id,omitempty"`
}

type templatePageToken struct {
	RecommendedPageSize int             `json:"ps,omitempty"`
	IncludeDisabled     bool            `json:"d,omitempty"`
	BundleIDs           []spec.BundleID `json:"ids,omitempty"`
	Tags                []string        `json:"tags,omitempty"`
	BuiltInDone         bool            `json:"bd,omitempty"`
	DirTok              string          `json:"dt,omitempty"`
}

func base64JSONEncode[T any](t T) string {
	raw, _ := json.Marshal(t)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func base64JSONDecode[T any](s string) (T, error) {
	var t T
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(raw, &t)
	return t, err
}

// nullableStr returns a pointer to the string, or nil if the string is empty.
func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
