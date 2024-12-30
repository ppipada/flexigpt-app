package types

import (
	"encoding/json"
	"fmt"
)

type IntString struct {
	Value interface{}
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (is *IntString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal data into an int
	var intValue int
	if err := json.Unmarshal(data, &intValue); err == nil {
		is.Value = intValue
		return nil
	}

	// Try to unmarshal data into a string
	var strValue string
	if err := json.Unmarshal(data, &strValue); err == nil {
		is.Value = strValue
		return nil
	}

	// If neither int nor string, return an error
	return fmt.Errorf("IntString must be a string or an integer")
}

// MarshalJSON implements the json.Marshaler interface.
func (is IntString) MarshalJSON() ([]byte, error) {
	switch v := is.Value.(type) {
	case int:
		return json.Marshal(v)
	case string:
		return json.Marshal(v)
	default:
		return nil, fmt.Errorf("IntString contains unsupported type")
	}
}

// Helper methods
func (is IntString) IsInt() bool {
	_, ok := is.Value.(int)
	return ok
}

func (is IntString) IsString() bool {
	_, ok := is.Value.(string)
	return ok
}

func (is IntString) IntValue() (int, bool) {
	v, ok := is.Value.(int)
	return v, ok
}

func (is IntString) StringValue() (string, bool) {
	v, ok := is.Value.(string)
	return v, ok
}
