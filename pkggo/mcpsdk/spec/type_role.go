package spec

import "github.com/flexigpt/flexiui/pkggo/mcpsdk/spec/customtype"

var enumValues_Role = []string{"assistant", "user"}

// Role
type Role struct {
	*customtype.StringUnion
}

// NewRole creates a new Role with the provided value.
func NewRole(value string) *Role {
	stringUnion := customtype.NewStringUnion(enumValues_Role...)
	_ = stringUnion.SetValue(value)
	return &Role{StringUnion: stringUnion}
}

// UnmarshalJSON implements json.Unmarshaler for Role.
func (r *Role) UnmarshalJSON(b []byte) error {
	if r.StringUnion == nil {
		// Initialize with allowed values if not already initialized
		r.StringUnion = customtype.NewStringUnion(enumValues_Role...)
	}
	return r.StringUnion.UnmarshalJSON(b)
}

// MarshalJSON implements json.Marshaler for Role.
func (r *Role) MarshalJSON() ([]byte, error) {
	return r.StringUnion.MarshalJSON()
}

var RoleAssistant *Role = NewRole("assistant")
var RoleUser *Role = NewRole("user")
