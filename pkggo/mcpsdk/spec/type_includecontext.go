package spec

var enumValues_IncludeContext = []string{
	"allServers",
	"none",
	"thisServer",
}

// IncludeContext
type IncludeContext struct {
	*StringUnion
}

// NewIncludeContext creates a new IncludeContext with the provided value.
func NewIncludeContext(
	value string,
) *IncludeContext {
	stringUnion := NewStringUnion(enumValues_IncludeContext...)
	_ = stringUnion.SetValue(value)
	return &IncludeContext{StringUnion: stringUnion}
}

// UnmarshalJSON implements json.Unmarshaler for IncludeContext.
func (r *IncludeContext) UnmarshalJSON(b []byte) error {
	if r.StringUnion == nil {
		// Initialize with allowed values if not already initialized
		r.StringUnion = NewStringUnion(
			enumValues_IncludeContext...)
	}
	return r.StringUnion.UnmarshalJSON(b)
}

// MarshalJSON implements json.Marshaler for IncludeContext.
func (r *IncludeContext) MarshalJSON() ([]byte, error) {
	return r.StringUnion.MarshalJSON()
}

var IncludeContextNone *IncludeContext = NewIncludeContext("none")
var IncludeContextAllServers *IncludeContext = NewIncludeContext("allServers")
var IncludeContextThisServer *IncludeContext = NewIncludeContext("thisServer")
