package spec

var enumValues_Ref = []string{"ref/resource", "ref/prompt"}

// Ref
type Ref struct {
	*StringUnion
}

// NewRef creates a new Ref with the provided value.
func NewRef(value string) *Ref {
	stringUnion := NewStringUnion(enumValues_Ref...)
	_ = stringUnion.SetValue(value)
	return &Ref{StringUnion: stringUnion}
}

// UnmarshalJSON implements json.Unmarshaler for Ref.
func (r *Ref) UnmarshalJSON(b []byte) error {
	if r.StringUnion == nil {
		// Initialize with allowed values if not already initialized
		r.StringUnion = NewStringUnion(enumValues_Ref...)
	}
	return r.StringUnion.UnmarshalJSON(b)
}

// MarshalJSON implements json.Marshaler for Ref.
func (r *Ref) MarshalJSON() ([]byte, error) {
	return r.StringUnion.MarshalJSON()
}

var (
	RefResource *Ref = NewRef("ref/resource")
	RefPrompt   *Ref = NewRef("ref/prompt")
)
