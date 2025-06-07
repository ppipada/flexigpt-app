package spec

var enumValuesRef = []string{"ref/resource", "ref/prompt"}

// Ref to a resource or prompt.
type Ref struct {
	*StringUnion
}

// NewRef creates a new Ref with the provided value.
func NewRef(value string) *Ref {
	stringUnion := NewStringUnion(enumValuesRef...)
	_ = stringUnion.SetValue(value)
	return &Ref{StringUnion: stringUnion}
}

// UnmarshalJSON implements json.Unmarshaler for Ref.
func (r *Ref) UnmarshalJSON(b []byte) error {
	if r.StringUnion == nil {
		// Initialize with allowed values if not already initialized.
		r.StringUnion = NewStringUnion(enumValuesRef...)
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
