package filestore

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	simplemapdbEncdec "github.com/flexigpt/flexiui/pkggo/simplemapdb/encdec"
)

func TestNewMapFileStore(t *testing.T) {
	tempDir := t.TempDir()
	tests := []struct {
		name              string
		filename          string
		defaultData       map[string]interface{}
		createFile        bool
		fileContent       string
		options           []Option
		expectError       bool
		expectedErrorText string
	}{
		{
			name:        "File does not exist, createIfNotExists true",
			filename:    filepath.Join(tempDir, "store1.json"),
			defaultData: map[string]interface{}{"k": "v"},
			options:     []Option{WithCreateIfNotExists(true)},
			expectError: false,
		},
		{
			name:              "File does not exist, createIfNotExists false",
			filename:          filepath.Join(tempDir, "store2.json"),
			defaultData:       map[string]interface{}{"k": "v"},
			options:           []Option{WithCreateIfNotExists(false)},
			expectError:       true,
			expectedErrorText: "does not exist",
		},
		{
			name:        "File exists with valid content",
			filename:    filepath.Join(tempDir, "store3.json"),
			defaultData: map[string]interface{}{"k": "v"},
			createFile:  true,
			fileContent: `{"foo":"bar"}`,
			options:     []Option{},
			expectError: false,
		},
		{
			name:        "File exists with invalid content",
			filename:    filepath.Join(tempDir, "store4.json"),
			defaultData: map[string]interface{}{"k": "v"},
			createFile:  true,
			fileContent: `{invalid json}`,
			options:     []Option{},
			expectError: true,
		},
		{
			name:        "File exists but cannot open",
			filename:    filepath.Join(tempDir, "store5.json"),
			defaultData: map[string]interface{}{"k": "v"},
			createFile:  true,
			fileContent: `{"foo":"bar"}`,
			options:     []Option{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		if tt.createFile {
			err := os.WriteFile(tt.filename, []byte(tt.fileContent), 0o644)
			if err != nil {
				t.Fatalf("[%s] Failed to create test file: %v", tt.name, err)
			}
		}

		if tt.name == "File exists but cannot open" {
			// Create a file with no read permissions
			err := os.Chmod(tt.filename, 0o000)
			if err != nil {
				t.Fatalf("[%s] Failed to change file permissions: %v", tt.name, err)
			}
			ch := func() {
				_ = os.Chmod(tt.filename, 0o644) // Ensure we can clean up later}
			}
			defer ch()
		}

		_, err := NewMapFileStore(tt.filename, tt.defaultData, tt.options...)
		if tt.expectError {
			if err == nil {
				t.Errorf("[%s] Expected error but got nil", tt.name)
			} else if tt.expectedErrorText != "" && !strings.Contains(err.Error(), tt.expectedErrorText) {
				t.Errorf("[%s] Expected error containing '%s' but got '%v'", tt.name, tt.expectedErrorText, err)
			}
		} else {
			if err != nil {
				t.Errorf("[%s] Unexpected error: %v", tt.name, err)
			}
		}
	}
}

func TestMapFileStore_SetKey_GetKey(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore.json")
	defaultData := map[string]interface{}{"foo": "bar"}
	keyEncDecs := map[string]simplemapdbEncdec.EncoderDecoder{
		"foo":          simplemapdbEncdec.EncryptedStringValueEncoderDecoder{},
		"parent.child": simplemapdbEncdec.EncryptedStringValueEncoderDecoder{},
	}
	store, err := NewMapFileStore(
		filename,
		defaultData,
		WithCreateIfNotExists(true),
		WithKeyEncoders(keyEncDecs),
	)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	tests := []struct {
		name       string
		key        string
		value      interface{}
		wantErrSet bool
		wantErrGet bool
	}{
		{
			name:  "Set and get simple key",
			key:   "foo",
			value: "bar",
		},
		{
			name:  "Set and get nested key",
			key:   "parent.child",
			value: "grandson",
		},
		{
			name:  "Set and get deep nested key",
			key:   "grand.parent.child.key",
			value: true,
		},
		{
			name:       "Set empty key",
			key:        "",
			value:      "value",
			wantErrSet: true,
		},
		{
			name:       "Set key with empty segment",
			key:        "parent..child",
			value:      "value",
			wantErrSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.SetKey(tt.key, tt.value)
			if tt.wantErrSet {
				if err == nil {
					t.Errorf("[%s] Expected error in SetKey but got nil", tt.name)
				}
				return
			} else if err != nil {
				t.Errorf("[%s] Unexpected error in SetKey: %v", tt.name, err)
				return
			}

			got, err := store.GetKey(tt.key)
			if tt.wantErrGet {
				if err == nil {
					t.Errorf("[%s] Expected error in GetKey but got nil", tt.name)
				}
				return
			} else {
				if err != nil {
					t.Errorf("[%s] Unexpected error in GetKey: %v", tt.name, err)
				} else if got != tt.value {
					t.Errorf("[%s] GetKey returned %v, expected %v", tt.name, got, tt.value)
				}
			}
		})
	}
}

func TestMapFileStore_DeleteKey(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore.json")
	// Pre-populate the store
	initialData := map[string]interface{}{
		"foo":    "bar",
		"parent": map[string]interface{}{"child": "value"},
		"grand": map[string]interface{}{
			"parent": map[string]interface{}{"child": map[string]interface{}{"key": "deep"}},
		},
		"nondeletable": "persist",
		"empty":        map[string]interface{}{"parent": map[string]interface{}{}},
		"list":         []interface{}{1, 2, 3},
		"nonexistent":  nil,
		"another": map[string]interface{}{
			"parent": map[string]interface{}{"child1": "val1", "child2": "val2"},
		},
	}

	store, err := NewMapFileStore(filename, initialData, WithCreateIfNotExists(true))
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	tests := []struct {
		name       string
		key        string
		wantErr    bool
		checkExist bool
	}{
		{
			name:       "Delete simple key",
			key:        "foo",
			checkExist: true,
		},
		{
			name:       "Delete nested key",
			key:        "parent.child",
			checkExist: true,
		},
		{
			name:       "Delete deep nested key",
			key:        "grand.parent.child.key",
			checkExist: true,
		},
		{
			name:    "Delete non-existent key",
			key:     "does.not.exist",
			wantErr: false,
		},
		{
			name:    "Delete key with empty segment",
			key:     "parent..child",
			wantErr: false,
		},
		{
			name:       "Delete empty map",
			key:        "empty.parent",
			checkExist: true,
		},
		{
			name:       "Delete from map with multiple keys",
			key:        "another.parent.child1",
			checkExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteKey(tt.key)
			if tt.wantErr {
				if err == nil {
					t.Errorf("[%s] Expected error in DeleteKey but got nil", tt.name)
				}
				return
			} else if err != nil {
				t.Errorf("[%s] Unexpected error in DeleteKey: %v", tt.name, err)
				return
			}

			if tt.checkExist {
				_, err := store.GetKey(tt.key)
				if err == nil {
					t.Errorf(
						"[%s] Expected key %s to be deleted, but it still exists",
						tt.name,
						tt.key,
					)
				}
			}
		})
	}
}

func TestMapFileStore_SetAll_GetAll(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore.json")
	defaultData := map[string]interface{}{"foo": "bar"}
	store, err := NewMapFileStore(
		filename,
		defaultData,
		WithCreateIfNotExists(true),
		WithAutoFlush(true),
	)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	tests := []struct {
		name          string
		data          map[string]interface{}
		expectedData  map[string]interface{}
		modifyData    bool
		expectedAfter map[string]interface{}
	}{
		{
			name: "Set and get all simple data",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": 2,
			},
			expectedData: map[string]interface{}{
				"key1": "value1",
				"key2": 2,
			},
		},
		{
			name: "Set and get all nested data",
			data: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "value",
				},
			},
			expectedData: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "value",
				},
			},
		},
		{
			name: "Set and get all empty data",
			data: map[string]interface{}{},
		},
		{
			name: "Modify returned data should not affect store",
			data: map[string]interface{}{
				"original": "value",
			},
			expectedData: map[string]interface{}{
				"original": "value",
			},
			modifyData: true,
			expectedAfter: map[string]interface{}{
				"original": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Reset()
			if err != nil {
				t.Fatalf("[%s] Failed to clear store: %v", tt.name, err)
			}

			err = store.SetAll(tt.data)
			if err != nil {
				t.Errorf("[%s] SetAll failed: %v", tt.name, err)
				return
			}

			got, err := store.GetAll(false)
			if err != nil {
				t.Errorf("Failed to get data err: %v", err)
			}
			if !deepEqual(got, tt.expectedData) {
				t.Errorf("[%s] GetAll returned %v, expected %v", tt.name, got, tt.expectedData)
			}

			if tt.modifyData {
				got["original"] = "modified"
				gotAfterModification, err := store.GetAll(false)
				if err != nil {
					t.Errorf("Failed to get data err: %v", err)
				}
				if !deepEqual(gotAfterModification, tt.expectedAfter) {
					t.Errorf(
						"[%s] Store data modified after external change: got %v, expected %v",
						tt.name,
						gotAfterModification,
						tt.expectedAfter,
					)
				}
			}
		})
	}
}

func TestMapFileStore_DeleteAll(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore.json")
	defaultData := map[string]interface{}{"foo": "bar"}
	store, err := NewMapFileStore(filename, defaultData, WithCreateIfNotExists(true))
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Re-populate the store
	initialData := map[string]interface{}{
		"key1": "value1",
		"key2": 2,
		"key3": true,
	}
	err = store.SetAll(initialData)
	if err != nil {
		t.Fatalf("Failed to set initial data: %v", err)
	}

	err = store.Reset()
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	got, err := store.GetAll(false)
	if err != nil {
		t.Errorf("Failed to get data err: %v", err)
	}
	if !reflect.DeepEqual(got, defaultData) {
		t.Errorf("Expected store to reset to defaultData %v, but got %v", defaultData, got)
	}
}

func TestMapFileStore_AutoFlush(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore_autoflush.json")
	defaultData := map[string]interface{}{"k": "v"}
	store, err := NewMapFileStore(
		filename,
		defaultData,
		WithCreateIfNotExists(true),
		WithAutoFlush(true),
	)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	err = store.SetKey("foo", "bar")
	if err != nil {
		t.Fatalf("SetKey failed: %v", err)
	}

	// Reopen the store
	store2, err := NewMapFileStore(filename, defaultData)
	if err != nil {
		t.Fatalf("Failed to reopen store: %v", err)
	}

	val, err := store2.GetKey("foo")
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}
	if val != "bar" {
		t.Errorf("Expected 'bar', got %v", val)
	}
}

func TestMapFileStore_NoAutoFlush(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore_noautoflush.json")
	defaultData := map[string]interface{}{"k": "v"}
	store, err := NewMapFileStore(
		filename,
		defaultData,
		WithCreateIfNotExists(true),
		WithAutoFlush(false),
	)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	err = store.SetKey("foo", "bar")
	if err != nil {
		t.Fatalf("SetKey failed: %v", err)
	}

	// Reopen the store
	store2, err := NewMapFileStore(filename, defaultData)
	if err != nil {
		t.Fatalf("Failed to reopen store: %v", err)
	}

	_, err = store2.GetKey("foo")
	if err == nil {
		t.Errorf("Expected error getting 'foo' from store2 as it should not be saved yet")
	}

	// Now save and reopen
	err = store.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	store3, err := NewMapFileStore(filename, defaultData)
	if err != nil {
		t.Fatalf("Failed to reopen store after save: %v", err)
	}

	val, err := store3.GetKey("foo")
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}
	if val != "bar" {
		t.Errorf("Expected 'bar', got %v", val)
	}
}

func TestMapFileStorePermissionErrorCases(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore_errors.json")
	defaultData := map[string]interface{}{"k": "v"}
	store, err := NewMapFileStore(filename, defaultData, WithCreateIfNotExists(true))
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Simulate Save error by making the file unwritable
	err = os.Chmod(filename, 0o444)
	if err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	ch := func() { _ = os.Chmod(filename, 0o644) }
	defer ch()

	err = store.SetKey("foo", "bar")
	if err == nil {
		t.Errorf("Expected error in SetKey due to unwritable file, but got nil")
	}

	// Simulate Decode error by writing invalid data into the file
	err = os.Chmod(filename, 0o666)
	if err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	err = os.WriteFile(filename, []byte(`invalid json`), 0o666)
	if err != nil {
		t.Fatalf("Failed to write invalid data to file: %v", err)
	}

	_, err = NewMapFileStore(filename, defaultData)
	if err == nil {
		t.Errorf("Expected error when loading store from invalid data, but got nil")
	}
}

func TestMapFileStore_NestedStructures(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore_nested.json")
	defaultData := map[string]interface{}{"k": "v"}
	store, err := NewMapFileStore(filename, defaultData, WithCreateIfNotExists(true))
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Set nested data
	tests := []struct {
		name      string
		key       string
		value     interface{}
		expectErr bool
	}{
		{
			name:  "Set nested map",
			key:   "parent.child",
			value: "value",
		},
		{
			name:  "Set nested map with existing parent",
			key:   "parent.anotherChild",
			value: 123,
		},
		{
			name:  "Set deep nested map",
			key:   "grand.parent.child",
			value: true,
		},
		{
			name:      "Set value where intermediate is not a map",
			key:       "parent.child.key",
			value:     "invalid",
			expectErr: true, // 'parent.child' is a string, cannot set 'key' under it
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.SetKey(tt.key, tt.value)
			if tt.expectErr {
				if err == nil {
					t.Errorf("[%s] Expected error in SetKey but got nil", tt.name)
				}
				return
			} else if err != nil {
				t.Errorf("[%s] SetKey failed: %v", tt.name, err)
				return
			}

			got, err := store.GetKey(tt.key)
			if err != nil {
				t.Errorf("[%s] GetKey failed: %v", tt.name, err)
				return
			}
			if got != tt.value {
				t.Errorf("[%s] GetKey returned %v, expected %v", tt.name, got, tt.value)
			}
		})
	}
}

// deepEqual is a simple helper function to compare two interface{} values for equality.
// It handles comparison of maps and basic types.
func deepEqual(a, b interface{}) bool {
	switch aVal := a.(type) {
	case map[string]interface{}:
		bVal, ok := b.(map[string]interface{})
		if !ok || len(aVal) != len(bVal) {
			return false
		}
		for k, v := range aVal {
			if !deepEqual(v, bVal[k]) {
				return false
			}
		}
		return true
	case []interface{}:
		bVal, ok := b.([]interface{})
		if !ok || len(aVal) != len(bVal) {
			return false
		}
		for i := range aVal {
			if !deepEqual(aVal[i], bVal[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
