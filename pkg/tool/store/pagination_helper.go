package store

import (
	"encoding/base64"
	"encoding/json"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
)

// bundlePageToken is the opaque cursor used when paging over tool-bundles.
type bundlePageToken struct {
	BundleIDs       []bundleitemutils.BundleID `json:"ids,omitempty"` // Optional bundle-ID filter.
	IncludeDisabled bool                       `json:"d,omitempty"`   // Include disabled bundles?
	PageSize        int                        `json:"s"`             // Requested page-size.
	CursorMod       string                     `json:"t,omitempty"`   // RFC-3339-nano modification timestamp.
	CursorID        bundleitemutils.BundleID   `json:"id,omitempty"`  // Tie-breaker for equal timestamps.
}

// toolPageToken is the opaque cursor used when paging over individual tool
// versions (ListTools API).
type toolPageToken struct {
	RecommendedPageSize int                        `json:"ps,omitempty"`
	IncludeDisabled     bool                       `json:"d,omitempty"`
	BundleIDs           []bundleitemutils.BundleID `json:"ids,omitempty"`
	Tags                []string                   `json:"tags,omitempty"`
	BuiltInDone         bool                       `json:"bd,omitempty"` // Built-ins already emitted?
	DirTok              string                     `json:"dt,omitempty"` // Directory-store cursor.
}

// base64JSONEncode encodes an arbitrary value as URL-safe base64 JSON.
func base64JSONEncode[T any](t T) string {
	raw, _ := json.Marshal(t)
	return base64.RawURLEncoding.EncodeToString(raw)
}

// base64JSONDecode inverse of base64JSONEncode.
func base64JSONDecode[T any](s string) (T, error) {
	var t T
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(raw, &t)
	return t, err
}

// nullableStr returns &s unless s=="" in which case it returns nil.
func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
