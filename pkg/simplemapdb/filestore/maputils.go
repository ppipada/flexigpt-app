package filestore

import (
	"errors"
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

// GetValueAtPath retrieves the value at the specified path in the data map.
func GetValueAtPath(data any, keys []string) (any, error) {
	parentMap, lastKey, err := NavigateToParentMap(data, keys, false)
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

// SetValueAtPath sets the value at the specified path in the data map.
func SetValueAtPath(data any, keys []string, value any) error {
	parentMap, lastKey, err := NavigateToParentMap(data, keys, true)
	if err != nil {
		return err
	}
	if lastKey == "" {
		return &KeyNotFoundError{Key: lastKey, Path: strings.Join(keys, ".")}
	}
	parentMap[lastKey] = DeepCopyValue(value)
	return nil
}

// DeleteValueAtPath deletes the value at the specified path in the data map.
// If path is not found, this is a noop.
func DeleteValueAtPath(data any, keys []string) error {
	parentMap, lastKey, err := NavigateToParentMap(data, keys, false)
	if err != nil {
		var kne *KeyNotFoundError
		if errors.As(err, &kne) {
			// Path not found, Delete is a noop in this case.
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

// DeepCopyValue creates a deep copy of an any value.
func DeepCopyValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		newV := make(map[string]any)
		for k, val := range v {
			newV[k] = DeepCopyValue(val)
		}
		return newV
	case []any:
		sliceCopy := make([]any, len(v))
		for i, elem := range v {
			sliceCopy[i] = DeepCopyValue(elem)
		}
		return sliceCopy
	default:
		// For basic types, it's safe to return as is.
		return v
	}
}

// NavigateToParentMap navigates to the parent map of the last key in the given path.
// It returns the parent map, the last key, and any error encountered.
// If createMissing is true, it creates any missing maps along the path.
func NavigateToParentMap(
	data any,
	keys []string,
	createMissing bool,
) (parentMap map[string]any, lastKey string, err error) {
	if len(keys) == 0 {
		return nil, "", errors.New("empty path received")
	}
	current := data
	for i := range len(keys) - 1 {
		key := keys[i]
		m, ok := current.(map[string]any)
		if !ok {
			path := strings.Join(keys[:i], ".")
			return nil, "", fmt.Errorf("path '%s' is not a map", path)
		}
		next, ok := m[key]
		if !ok {
			if createMissing && key != "" {
				// Create nested map.
				newMap := make(map[string]any)
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

	parentMap, ok := current.(map[string]any)
	if !ok {
		path := strings.Join(keys[:len(keys)-1], ".")
		return nil, "", fmt.Errorf("path '%s' is not a map", path)
	}
	lastKey = keys[len(keys)-1]
	return parentMap, lastKey, nil
}
