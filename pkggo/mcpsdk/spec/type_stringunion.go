package spec

import (
	"encoding/json"
	"fmt"
)

// StringUnion represents a string value that must be one of a predefined set of allowed values.
type StringUnion struct {
	Value         string
	allowedValues map[string]struct{}
}

// NewStringUnion creates a new StringUnion with the given allowed values.
func NewStringUnion(allowedValues ...string) *StringUnion {
	allowed := make(map[string]struct{}, len(allowedValues))
	for _, v := range allowedValues {
		allowed[v] = struct{}{}
	}
	return &StringUnion{
		allowedValues: allowed,
	}
}

// SetValue sets the value of the StringUnion after validating it.
func (s *StringUnion) SetValue(value string) error {
	if _, ok := s.allowedValues[value]; !ok {
		return fmt.Errorf("invalid value %q, expected one of %v", value, s.AllowedValues())
	}
	s.Value = value
	return nil
}

// AllowedValues returns a slice of allowed values.
func (s *StringUnion) AllowedValues() []string {
	keys := make([]string, 0, len(s.allowedValues))
	for k := range s.allowedValues {
		keys = append(keys, k)
	}
	return keys
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *StringUnion) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	return s.SetValue(v)
}

// MarshalJSON implements json.Marshaler.
func (s *StringUnion) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
}
