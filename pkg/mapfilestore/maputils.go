package mapfilestore

import (
	"fmt"
	"strings"
)

// KeyNotFoundError is a custom error type for missing keys.
type KeyNotFoundError struct {
	Key  string
	Path string
}

// Error implements the error interface.
func (e *KeyNotFoundError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("key '%s' not found at path '%s'", e.Key, e.Path)
	}
	return fmt.Sprintf("key '%s' not found", e.Key)
}

// getValueAtPath retrieves the value at the specified path in the data map.
func getValueAtPath(data interface{}, keys []string) (interface{}, error) {
	parentMap, lastKey, err := navigateToParentMap(data, keys, false)
	if err != nil {
		return nil, err
	}
	val, ok := parentMap[lastKey]
	if !ok {
		path := strings.Join(keys[:len(keys)-1], ".")
		return nil, &KeyNotFoundError{Key: lastKey, Path: path}
	}
	return val, nil
}

// setValueAtPath sets the value at the specified path in the data map.
func setValueAtPath(data interface{}, keys []string, value interface{}) error {
	parentMap, lastKey, err := navigateToParentMap(data, keys, true)
	if err != nil {
		return err
	}
	if lastKey == "" {
		return &KeyNotFoundError{Key: lastKey, Path: strings.Join(keys, ".")}
	}
	parentMap[lastKey] = deepCopyValue(value)
	return nil
}

// deleteValueAtPath deletes the value at the specified path in the data map.
// If path is not found, this is a noop
func deleteValueAtPath(data interface{}, keys []string) error {
	parentMap, lastKey, err := navigateToParentMap(data, keys, false)
	if err != nil {
		if _, ok := err.(*KeyNotFoundError); ok {
			// Path not found, Delete is a noop in this case
			return nil
		}
		return err
	}
	_, ok := parentMap[lastKey]
	if ok {
		delete(parentMap, lastKey)
	}
	return nil
}

// deepCopyValue creates a deep copy of an interface{} value.
func deepCopyValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		newV := make(map[string]interface{})
		for k, val := range v {
			newV[k] = deepCopyValue(val)
		}
		return newV
	case []interface{}:
		sliceCopy := make([]interface{}, len(v))
		for i, elem := range v {
			sliceCopy[i] = deepCopyValue(elem)
		}
		return sliceCopy
	default:
		return v // For basic types, it's safe to return as is
	}
}

// navigateToParentMap navigates to the parent map of the last key in the given path.
// It returns the parent map, the last key, and any error encountered.
// If createMissing is true, it creates any missing maps along the path.
func navigateToParentMap(data interface{}, keys []string, createMissing bool) (map[string]interface{}, string, error) {
	if len(keys) == 0 {
		return nil, "", fmt.Errorf("empty path received")
	}
	current := data
	for i := 0; i < len(keys)-1; i++ {
		key := keys[i]
		m, ok := current.(map[string]interface{})
		if !ok {
			path := strings.Join(keys[:i], ".")
			return nil, "", fmt.Errorf("path '%s' is not a map", path)
		}
		next, ok := m[key]
		if !ok {
			if createMissing && key != "" {
				// Create nested map
				newMap := make(map[string]interface{})
				m[key] = newMap
				current = newMap
			} else {
				path := strings.Join(keys[:i], ".")
				return nil, "", &KeyNotFoundError{Key: key, Path: path}
			}
		} else {
			current = next
		}
	}

	m, ok := current.(map[string]interface{})
	if !ok {
		path := strings.Join(keys[:len(keys)-1], ".")
		return nil, "", fmt.Errorf("path '%s' is not a map", path)
	}
	lastKey := keys[len(keys)-1]
	return m, lastKey, nil
}
