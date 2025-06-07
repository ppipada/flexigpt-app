package filestore

import (
	"reflect"
	"testing"
)

// TestGetValueAtPath tests the GetValueAtPath function.
func TestGetValueAtPath(t *testing.T) {
	tests := []struct {
		name      string
		data      any
		keys      []string
		wantValue any
		wantErr   bool
	}{
		{
			name: "Happy path - valid nested keys",
			data: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "value1",
					},
				},
			},
			keys:      []string{"a", "b", "c"},
			wantValue: "value1",
			wantErr:   false,
		},
		{
			name: "Error path - key not found",
			data: map[string]any{
				"a": map[string]any{
					"b": "value2",
				},
			},
			keys:    []string{"a", "b", "c"},
			wantErr: true,
		},
		{
			name: "Error path - path not a map",
			data: map[string]any{
				"a": "not a map",
			},
			keys:    []string{"a", "b"},
			wantErr: true,
		},
		{
			name: "Boundary case - empty keys",
			data: map[string]any{
				"a": "value3",
			},
			keys:    []string{},
			wantErr: true,
		},
		{
			name:    "Boundary case - nil data",
			data:    nil,
			keys:    []string{"a"},
			wantErr: true,
		},
		{
			name: "Happy path - root level key",
			data: map[string]any{
				"root": "rootValue",
			},
			keys:      []string{"root"},
			wantValue: "rootValue",
			wantErr:   false,
		},
		{
			name: "Happy path - value is a map",
			data: map[string]any{
				"mapKey": map[string]any{
					"innerKey": "innerValue",
				},
			},
			keys: []string{"mapKey"},
			wantValue: map[string]any{
				"innerKey": "innerValue",
			},
			wantErr: false,
		},
		{
			name: "Error path - intermediate path is not a map",
			data: map[string]any{
				"a": "stringValue",
			},
			keys:    []string{"a", "b"},
			wantErr: true,
		},
		{
			name: "Error path - key not present at the end",
			data: map[string]any{
				"a": map[string]any{
					"b": map[string]any{},
				},
			},
			keys:    []string{"a", "b", "c"},
			wantErr: true,
		},
		{
			name: "Happy path - value is a slice",
			data: map[string]any{
				"list": []any{"item1", "item2"},
			},
			keys:      []string{"list"},
			wantValue: []any{"item1", "item2"},
			wantErr:   false,
		},
		{
			name: "Boundary case - key is empty string",
			data: map[string]any{
				"": "emptyKey",
			},
			keys:      []string{""},
			wantValue: "emptyKey",
			wantErr:   false,
		},
		{
			name:    "Error path - data is not a map",
			data:    "not a map",
			keys:    []string{"a"},
			wantErr: true,
		},
		{
			name: "Happy path - deep nested key",
			data: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": map[string]any{
							"level4": "deepValue",
						},
					},
				},
			},
			keys:      []string{"level1", "level2", "level3", "level4"},
			wantValue: "deepValue",
			wantErr:   false,
		},
		{
			name: "Error path - nil key in middle",
			data: map[string]any{
				"a": map[string]any{
					"b": "value",
				},
			},
			keys:    []string{"a", "", "c"},
			wantErr: true,
		},
		{
			name: "Happy path - keys with special characters",
			data: map[string]any{
				"a.b": map[string]any{
					"c.d": "specialValue",
				},
			},
			keys:      []string{"a.b", "c.d"},
			wantValue: "specialValue",
			wantErr:   false,
		},
		{
			name:    "Boundary case - data is empty map",
			data:    map[string]any{},
			keys:    []string{"a"},
			wantErr: true,
		},
		{
			name:      "Boundary case - data is nil",
			data:      nil,
			keys:      []string{},
			wantValue: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, err := GetValueAtPath(tt.data, tt.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValueAtPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("GetValueAtPath() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

// TestSetValueAtPath tests the SetValueAtPath function.
func TestSetValueAtPath(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		keys     []string
		value    any
		wantData any
		wantErr  bool
	}{
		{
			name: "Happy path - set new value at valid path",
			data: map[string]any{
				"a": map[string]any{
					"b": "oldValue",
				},
			},
			keys:  []string{"a", "b"},
			value: "newValue",
			wantData: map[string]any{
				"a": map[string]any{
					"b": "newValue",
				},
			},
			wantErr: false,
		},
		{
			name: "Happy path - create missing nested maps",
			data: map[string]any{
				"a": map[string]any{},
			},
			keys:  []string{"a", "b", "c"},
			value: "nestedValue",
			wantData: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "nestedValue",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Error path - intermediate key is not a map",
			data: map[string]any{
				"a": "not a map",
			},
			keys:    []string{"a", "b"},
			value:   "value",
			wantErr: true,
		},
		{
			name: "Boundary case - empty keys",
			data: map[string]any{
				"a": "value",
			},
			keys:     []string{},
			value:    "newRootValue",
			wantData: "newRootValue",
			wantErr:  true,
		},
		{
			name:    "Boundary case - data is nil",
			data:    nil,
			keys:    []string{"a"},
			value:   "value",
			wantErr: true,
		},
		{
			name: "Happy path - set value at root",
			data: map[string]any{},
			keys: []string{},
			value: map[string]any{
				"new": "rootValue",
			},
			wantData: nil,
			wantErr:  true,
		},
		{
			name: "Happy path - overwrite existing map",
			data: map[string]any{
				"a": map[string]any{
					"b": "oldValue",
				},
			},
			keys:  []string{"a"},
			value: "newAValue",
			wantData: map[string]any{
				"a": "newAValue",
			},
			wantErr: false,
		},
		{
			name:    "Error path - data is not a map",
			data:    "not a map",
			keys:    []string{"a"},
			value:   "value",
			wantErr: true,
		},
		{
			name:  "Happy path - keys with special characters",
			data:  map[string]any{},
			keys:  []string{"key.with.dots"},
			value: "specialKeyValue",
			wantData: map[string]any{
				"key.with.dots": "specialKeyValue",
			},
			wantErr: false,
		},
		{
			name: "Boundary case - set value at non-existent root",
			data: nil,
			keys: []string{},
			value: map[string]any{
				"root": "rootValue",
			},
			wantData: nil,
			wantErr:  true,
		},
		{
			name: "Error path - intermediate path is nil",
			data: map[string]any{
				"a": nil,
			},
			keys:    []string{"a", "b"},
			value:   "value",
			wantErr: true,
		},
		{
			name: "Error path - empty string as key",
			data: map[string]any{
				"a": map[string]any{},
			},
			keys:    []string{"a", ""},
			value:   "emptyKey",
			wantErr: true,
		},
		{
			name: "Error path - cannot set value in a non-map and non-slice type",
			data: map[string]any{
				"a": "stringValue",
			},
			keys:    []string{"a", "b"},
			value:   "value",
			wantErr: true,
		},
		{
			name:  "Happy path - set value at deep nested key",
			data:  map[string]any{},
			keys:  []string{"level1", "level2", "level3", "level4"},
			value: "deepValue",
			wantData: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": map[string]any{
							"level4": "deepValue",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Error path - nil key in middle",
			data: map[string]any{
				"a": map[string]any{},
			},
			keys:    []string{"a", "", "c"},
			value:   "value",
			wantErr: true,
		},
		{
			name:     "Boundary case - empty data and keys",
			data:     nil,
			keys:     []string{},
			value:    "rootValue",
			wantData: "rootValue",
			wantErr:  true,
		},
		{
			name:  "Happy path - set value in an empty map",
			data:  map[string]any{},
			keys:  []string{"a"},
			value: "value",
			wantData: map[string]any{
				"a": "value",
			},
			wantErr: false,
		},
		{
			name: "Error path - set value at invalid index in slice",
			data: map[string]any{
				"list": []any{"item1", "item2"},
			},
			keys:    []string{"list", "2"},
			value:   "item3",
			wantErr: true,
		},
		{
			name:     "Error path - setting value with nil data and keys",
			data:     nil,
			keys:     nil,
			value:    "value",
			wantData: "value",
			wantErr:  true,
		},
		{
			name: "Error path - keys is nil",
			data: map[string]any{
				"a": "value",
			},
			keys:     nil,
			value:    "newRootValue",
			wantData: "newRootValue",
			wantErr:  true,
		},
		{
			name:    "Error path - setting value in nil map",
			data:    nil,
			keys:    []string{"a", "b"},
			value:   "value",
			wantErr: true,
		},
		{
			name: "Boundary case - setting nil value",
			data: map[string]any{
				"a": "value",
			},
			keys:  []string{"a"},
			value: nil,
			wantData: map[string]any{
				"a": nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		// Make a deep copy of the original data to avoid side effects.
		dataCopy := DeepCopyValue(tt.data)
		t.Run(tt.name, func(t *testing.T) {
			err := SetValueAtPath(dataCopy, tt.keys, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetValueAtPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Compare the modified data with the expected data.
			if !reflect.DeepEqual(dataCopy, tt.wantData) {
				t.Errorf("After SetValueAtPath(), data = %v, want %v", dataCopy, tt.wantData)
			}
		})
	}
}

// TestDeleteValueAtPath tests the DeleteValueAtPath function.
func TestDeleteValueAtPath(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		keys     []string
		wantData any
		wantErr  bool
	}{
		{
			name: "Happy path - delete existing key",
			data: map[string]any{
				"a": map[string]any{
					"b": "value",
				},
			},
			keys: []string{"a", "b"},
			wantData: map[string]any{
				"a": map[string]any{},
			},
			wantErr: false,
		},
		{
			name: "Happy path - delete non-existent key",
			data: map[string]any{
				"a": map[string]any{
					"b": "value",
				},
			},
			wantData: map[string]any{
				"a": map[string]any{
					"b": "value",
				},
			},
			keys:    []string{"a", "c"},
			wantErr: false,
		},
		{
			name: "Error path - intermediate path is not a map",
			data: map[string]any{
				"a": "not a map",
			},
			keys:    []string{"a", "b"},
			wantErr: true,
		},
		{
			name: "Boundary case - empty keys",
			data: map[string]any{
				"a": "value",
			},
			keys: []string{},
			wantData: map[string]any{
				"a": "value",
			},
			wantErr: true,
		},
		{
			name:    "Boundary case - data is nil",
			data:    nil,
			keys:    []string{"a"},
			wantErr: true,
		},
		{
			name: "Happy path - delete root key",
			data: map[string]any{
				"root":  "value",
				"other": "otherValue",
			},
			keys: []string{"root"},
			wantData: map[string]any{
				"other": "otherValue",
			},
			wantErr: false,
		},
		{
			name:    "Error path - data is not a map",
			data:    "not a map",
			keys:    []string{"a"},
			wantErr: true,
		},
		{
			name: "Happy path - delete key with empty string",
			data: map[string]any{
				"": "emptyKey",
			},
			keys:     []string{""},
			wantData: map[string]any{},
			wantErr:  false,
		},
		{
			name: "Happy path - delete key in slice",
			data: map[string]any{
				"list": []any{"item1", "item2", "item3"},
			},
			keys: []string{"list", "1"},
			// Deleting an index in a slice is not supported.
			wantErr: true,
		},
		{
			name: "Happy path - delete deep nested key",
			data: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": "deepValue",
					},
				},
			},
			keys: []string{"level1", "level2", "level3"},
			wantData: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{},
				},
			},
			wantErr: false,
		},
		{
			name: "Happy path - nil key in keys",
			data: map[string]any{
				"a": map[string]any{
					"b": "value",
				},
			},
			wantData: map[string]any{
				"a": map[string]any{
					"b": "value",
				},
			},
			keys:    []string{"a", "", "c"},
			wantErr: false,
		},
		{
			name:     "Happy path - deleting from an empty map",
			data:     map[string]any{},
			wantData: map[string]any{},
			keys:     []string{"a"},
			wantErr:  false,
		},
		{
			name: "Happy path - delete key with special characters",
			data: map[string]any{
				"a.b":   "specialKeyValue",
				"other": "otherValue",
			},
			keys: []string{"a.b"},
			wantData: map[string]any{
				"other": "otherValue",
			},
			wantErr: false,
		},
		{
			name:    "Error path - delete from nil data",
			data:    nil,
			keys:    []string{"a"},
			wantErr: true,
		},
		{
			name: "Error path - keys is nil",
			data: map[string]any{
				"a": "value",
			},
			keys:    nil,
			wantErr: true,
		},
		{
			name: "Happy path - deleting non-existent nested key",
			data: map[string]any{
				"a": map[string]any{
					"b": "value",
				},
			},
			wantData: map[string]any{
				"a": map[string]any{
					"b": "value",
				},
			},
			keys:    []string{"a", "c"},
			wantErr: false,
		},
		{
			name: "Boundary case - deleting from non-map type",
			data: map[string]any{
				"a": "stringValue",
			},
			keys:    []string{"a", "b"},
			wantErr: true,
		},
		{
			name: "Happy path - delete entire data",
			data: map[string]any{
				"a": "value",
			},
			keys: []string{},
			wantData: map[string]any{
				"a": "value",
			},
			// Cannot delete root data.
			wantErr: true,
		},
	}

	for _, tt := range tests {
		// Make a deep copy of the original data to avoid side effects.
		dataCopy := DeepCopyValue(tt.data)
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteValueAtPath(dataCopy, tt.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteValueAtPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Compare the modified data with the expected data.
			if !reflect.DeepEqual(dataCopy, tt.wantData) {
				t.Errorf("After DeleteValueAtPath(), data = %v, want %v", dataCopy, tt.wantData)
			}
		})
	}
}

// TestDeepCopyValue tests the DeepCopyValue function.
func TestDeepCopyValue(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		wantValue any
	}{
		{
			name: "Copy map[string]any",
			value: map[string]any{
				"a": "value",
				"b": map[string]any{
					"c": "nestedValue",
				},
			},
			wantValue: map[string]any{
				"a": "value",
				"b": map[string]any{
					"c": "nestedValue",
				},
			},
		},
		{
			name: "Copy []any",
			value: []any{
				"value1",
				[]any{"nestedValue1", "nestedValue2"},
				map[string]any{"a": "value"},
			},
			wantValue: []any{
				"value1",
				[]any{"nestedValue1", "nestedValue2"},
				map[string]any{"a": "value"},
			},
		},
		{
			name:      "Copy string value",
			value:     "stringValue",
			wantValue: "stringValue",
		},
		{
			name:      "Copy integer value",
			value:     42,
			wantValue: 42,
		},
		{
			name:      "Copy nil value",
			value:     nil,
			wantValue: nil,
		},
		{
			name:      "Copy float value",
			value:     3.14,
			wantValue: 3.14,
		},
		{
			name:      "Copy boolean value",
			value:     true,
			wantValue: true,
		},
		{
			name: "Copy complex nested structure",
			value: map[string]any{
				"a": []any{
					map[string]any{
						"b": "value",
					},
				},
				"c": 123,
			},
			wantValue: map[string]any{
				"a": []any{
					map[string]any{
						"b": "value",
					},
				},
				"c": 123,
			},
		},
		{
			name:      "Copy empty map",
			value:     map[string]any{},
			wantValue: map[string]any{},
		},
		{
			name:      "Copy empty slice",
			value:     []any{},
			wantValue: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue := DeepCopyValue(tt.value)

			// Modify the original value to ensure deep copy.
			modifyValue(tt.value)

			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("DeepCopyValue() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

// modifyValue modifies the input value in place for testing deep copy.
func modifyValue(value any) {
	const m = "modified"
	switch v := value.(type) {
	case map[string]any:
		for key := range v {
			v[key] = m
		}
	case []any:
		for i := range v {
			v[i] = m
		}
	}
}

// TestNavigateToParentMap tests the NavigateToParentMap function.
func TestNavigateToParentMap(t *testing.T) {
	tests := []struct {
		name          string
		data          any
		keys          []string
		createMissing bool
		wantMap       map[string]any
		wantLastKey   string
		wantErr       bool
	}{
		{
			name: "Happy path - existing path",
			data: map[string]any{
				"a": map[string]any{
					"b": map[string]any{},
				},
			},
			keys:          []string{"a", "b", "c"},
			createMissing: true,
			wantMap:       map[string]any{},
			wantLastKey:   "c",
			wantErr:       false,
		},
		{
			name: "Error path - path is not a map",
			data: map[string]any{
				"a": "not a map",
			},
			keys:          []string{"a", "b"},
			createMissing: false,
			wantErr:       true,
		},
		{
			name: "Happy path - key not found and not creating missing but parent present",
			data: map[string]any{
				"a": map[string]any{},
			},
			keys:          []string{"a", "b"},
			createMissing: false,
			wantMap:       map[string]any{},
			wantErr:       false,
			wantLastKey:   "b",
		},
		{
			name: "Happy path - create missing maps",
			data: map[string]any{
				"a": map[string]any{},
			},
			keys:          []string{"a", "b", "c"},
			createMissing: true,
			wantMap:       map[string]any{},
			wantLastKey:   "c",
			wantErr:       false,
		},
		{
			name:          "Error path - data is not a map",
			data:          "not a map",
			keys:          []string{"a", "b"},
			createMissing: true,
			wantErr:       true,
		},
		{
			name: "Boundary case - empty keys",
			data: map[string]any{
				"a": "value",
			},
			keys:          []string{},
			createMissing: false,
			wantLastKey:   "",
			wantErr:       true,
		},
		{
			name: "Happy path - keys with special characters",
			data: map[string]any{
				"a.b": map[string]any{
					"c.d": "value",
				},
			},
			keys:          []string{"a.b", "c.d"},
			createMissing: false,
			wantMap: map[string]any{
				"c.d": "value",
			},
			wantLastKey: "c.d",
			wantErr:     false,
		},
		{
			name: "Error path - intermediate map is nil",
			data: map[string]any{
				"a": nil,
			},
			keys:          []string{"a", "b"},
			createMissing: false,
			wantErr:       true,
		},
		{
			name: "Happy path - createMissing false but path exists",
			data: map[string]any{
				"a": map[string]any{
					"b": map[string]any{},
				},
			},
			keys:          []string{"a", "b", "c"},
			createMissing: false,
			wantMap:       map[string]any{},
			wantLastKey:   "c",
			wantErr:       false,
		},
		{
			name: "Error path - non-string key in keys",
			data: map[string]any{
				"a": map[string]any{
					"b": "value",
				},
			},
			keys:          []string{"a", "", "c"},
			createMissing: false,
			wantErr:       true,
		},
		{
			name:          "Boundary case - data is nil",
			data:          nil,
			keys:          []string{"a"},
			createMissing: true,
			wantErr:       true,
		},
		{
			name:          "Happy path - empty data and create missing",
			data:          map[string]any{},
			keys:          []string{"a", "b"},
			createMissing: true,
			wantMap:       map[string]any{},
			wantLastKey:   "b",
			wantErr:       false,
		},
		{
			name:          "Error path - data is not a map and createMissing false",
			data:          "not a map",
			keys:          []string{},
			createMissing: false,
			wantErr:       true,
		},
		{
			name: "Error path - create missing with nil intermediate map",
			data: map[string]any{
				"a": nil,
			},
			keys:          []string{"a", "b"},
			createMissing: true,
			wantLastKey:   "",
			wantErr:       true,
		},
		{
			name: "Happy path - empty keys",
			data: map[string]any{
				"a": "value",
			},
			keys:          []string{},
			createMissing: true,
			wantLastKey:   "",
			wantErr:       true,
		},
		{
			name:          "Error path - data is nil and keys are empty",
			data:          nil,
			keys:          []string{},
			createMissing: false,
			wantErr:       true,
		},
		{
			name: "Happy path - deep nested create missing",
			data: map[string]any{
				"a": map[string]any{},
			},
			keys:          []string{"a", "b", "c", "d"},
			createMissing: true,
			wantMap:       map[string]any{},
			wantLastKey:   "d",
			wantErr:       false,
		},
		{
			name: "Error path - path is not a map and createMissing true",
			data: map[string]any{
				"a": "value",
			},
			keys:          []string{"a", "b"},
			createMissing: true,
			wantErr:       true,
		},
		{
			name: "Error path - intermediate value is not a map",
			data: map[string]any{
				"a": 123,
			},
			keys:          []string{"a", "b"},
			createMissing: true,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		// Make a deep copy of the original data to avoid side effects.
		dataCopy := DeepCopyValue(tt.data)
		t.Run(tt.name, func(t *testing.T) {
			gotMap, gotLastKey, err := NavigateToParentMap(dataCopy, tt.keys, tt.createMissing)
			if (err != nil) != tt.wantErr {
				t.Errorf("NavigateToParentMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			// Compare gotMap with expected map.
			if !reflect.DeepEqual(gotMap, tt.wantMap) {
				t.Errorf("NavigateToParentMap() gotMap = %v, want %v", gotMap, tt.wantMap)
			}
			// Compare gotLastKey with expected last key.
			if gotLastKey != tt.wantLastKey {
				t.Errorf(
					"NavigateToParentMap() gotLastKey = %v, want %v",
					gotLastKey,
					tt.wantLastKey,
				)
			}
		})
	}
}
