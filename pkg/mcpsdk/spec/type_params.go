package spec

import (
	"encoding/json"
	"fmt"
)

// AdditionalParams is a map that may contain a special `_meta` field.
type AdditionalParams map[string]any

// UnmarshalJSON implements the json.Unmarshaler interface for AdditionalParams.
func (a *AdditionalParams) UnmarshalJSON(data []byte) error {
	// Unmarshal data into a temporary map.
	var tempMap map[string]any
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return err
	}

	// Initialize the map if it is nil.
	if *a == nil {
		*a = make(map[string]any)
	}

	// Iterate over the tempMap.
	for key, value := range tempMap {
		if key == "_meta" {
			// Ensure that _meta is a map[string]any.
			if metaMap, ok := value.(map[string]any); ok {
				(*a)[key] = metaMap
			} else {
				return fmt.Errorf("expected '_meta' to be an object, got %T", value)
			}
		} else {
			(*a)[key] = value
		}
	}

	return nil
}

// MarshalJSON implements the json.Marshaler interface for AdditionalParams.
func (a AdditionalParams) MarshalJSON() ([]byte, error) {
	// Simply marshal the map a to JSON.
	return json.Marshal(map[string]any(a))
}
