package spec

var enumValuesIncludeContext = []string{
	"allServers",
	"none",
	"thisServer",
}

// IncludeContext options from all or this server.
type IncludeContext struct {
	*StringUnion
}

// NewIncludeContext creates a new IncludeContext with the provided value.
func NewIncludeContext(
	value string,
) *IncludeContext {
	stringUnion := NewStringUnion(enumValuesIncludeContext...)
	_ = stringUnion.SetValue(value)
	return &IncludeContext{StringUnion: stringUnion}
}

// UnmarshalJSON implements json.Unmarshaler for IncludeContext.
func (r *IncludeContext) UnmarshalJSON(b []byte) error {
	if r.StringUnion == nil {
		// Initialize with allowed values if not already initialized.
		r.StringUnion = NewStringUnion(
			enumValuesIncludeContext...)
	}
	return r.StringUnion.UnmarshalJSON(b)
}

// MarshalJSON implements json.Marshaler for IncludeContext.
func (r *IncludeContext) MarshalJSON() ([]byte, error) {
	return r.StringUnion.MarshalJSON()
}

var (
	IncludeContextNone       *IncludeContext = NewIncludeContext("none")
	IncludeContextAllServers *IncludeContext = NewIncludeContext("allServers")
	IncludeContextThisServer *IncludeContext = NewIncludeContext("thisServer")
)
