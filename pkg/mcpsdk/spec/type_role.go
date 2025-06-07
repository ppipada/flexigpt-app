package spec

var enumValuesRole = []string{"assistant", "user"}

// Role of the actor.
type Role struct {
	*StringUnion
}

// NewRole creates a new Role with the provided value.
func NewRole(value string) *Role {
	stringUnion := NewStringUnion(enumValuesRole...)
	_ = stringUnion.SetValue(value)
	return &Role{StringUnion: stringUnion}
}

// UnmarshalJSON implements json.Unmarshaler for Role.
func (r *Role) UnmarshalJSON(b []byte) error {
	if r.StringUnion == nil {
		// Initialize with allowed values if not already initialized.
		r.StringUnion = NewStringUnion(enumValuesRole...)
	}
	return r.StringUnion.UnmarshalJSON(b)
}

// MarshalJSON implements json.Marshaler for Role.
func (r *Role) MarshalJSON() ([]byte, error) {
	return r.StringUnion.MarshalJSON()
}

var (
	RoleAssistant *Role = NewRole("assistant")
	RoleUser      *Role = NewRole("user")
)
