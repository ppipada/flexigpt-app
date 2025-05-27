package filestore

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
)

// deepEqual is a simple helper function to compare two any values for equality.
func deepEqual(a, b any) bool {
	switch aVal := a.(type) {
	case map[string]any:
		bVal, ok := b.(map[string]any)
		if !ok || len(aVal) != len(bVal) {
			return false
		}
		for k, v := range aVal {
			if !deepEqual(v, bVal[k]) {
				return false
			}
		}
		return true
	case []any:
		bVal, ok := b.([]any)
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

func TestNewMapFileStore(t *testing.T) {
	tempDir := t.TempDir()
	type testType struct {
		name              string
		filename          string
		defaultData       map[string]any
		createFile        bool
		fileContent       string
		options           []Option
		expectError       bool
		expectedErrorText string
	}
	tests := []testType{
		{
			name:        "File does not exist, createIfNotExists true",
			filename:    filepath.Join(tempDir, "store1.json"),
			defaultData: map[string]any{"k": "v"},
			options:     []Option{WithCreateIfNotExists(true)},
			expectError: false,
		},
		{
			name:              "File does not exist, createIfNotExists false",
			filename:          filepath.Join(tempDir, "store2.json"),
			defaultData:       map[string]any{"k": "v"},
			options:           []Option{WithCreateIfNotExists(false)},
			expectError:       true,
			expectedErrorText: "does not exist",
		},
		{
			name:        "File exists with valid content",
			filename:    filepath.Join(tempDir, "store3.json"),
			defaultData: map[string]any{"k": "v"},
			createFile:  true,
			fileContent: `{"foo":"bar"}`,
			options:     []Option{},
			expectError: false,
		},
		{
			name:        "File exists with invalid content",
			filename:    filepath.Join(tempDir, "store4.json"),
			defaultData: map[string]any{"k": "v"},
			createFile:  true,
			fileContent: `{invalid json}`,
			options:     []Option{},
			expectError: true,
		},
		{
			name:        "File exists but cannot open",
			filename:    filepath.Join(tempDir, "store5.json"),
			defaultData: map[string]any{"k": "v"},
			createFile:  true,
			fileContent: `{"foo":"bar"}`,
			options:     []Option{},
			expectError: true,
		},
	}

	runtNewMapFileStoreTestCase := func(t *testing.T, tt testType) {
		if tt.createFile {
			err := os.WriteFile(tt.filename, []byte(tt.fileContent), 0o600)
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

			defer func() {
				_ = os.Chmod(tt.filename, 0o644) // Ensure we can clean up later
			}()
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

	for _, tt := range tests {
		runtNewMapFileStoreTestCase(t, tt)
	}
}

func TestMapFileStore_SetKey_GetKey(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore.json")
	defaultData := map[string]any{"foo": "bar"}

	// Example: We'll return an EncoderDecoder for the paths "foo" and "parent.child".
	valueEncDecs := map[string]encdec.EncoderDecoder{
		"foo":          encdec.EncryptedStringValueEncoderDecoder{},
		"parent.child": encdec.EncryptedStringValueEncoderDecoder{},
	}

	store, err := NewMapFileStore(
		filename,
		defaultData,
		WithCreateIfNotExists(true),
		// New approach: We pass a function that returns an EncoderDecoder depending on pathSoFar.
		WithValueEncDecGetter(func(pathSoFar []string) encdec.EncoderDecoder {
			joined := strings.Join(pathSoFar, ".")
			if ed, ok := valueEncDecs[joined]; ok {
				return ed
			}
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	tests := []struct {
		name       string
		keys       []string
		value      any
		wantErrSet bool
		wantErrGet bool
	}{
		{
			name:  "Set and get simple key",
			keys:  []string{"foo"},
			value: "bar",
		},
		{
			name:  "Set and get nested key",
			keys:  []string{"parent", "child"},
			value: "grandson",
		},
		{
			name:  "Set and get deep nested key",
			keys:  []string{"grand", "parent", "child", "key"},
			value: true,
		},
		{
			name:       "Set empty key slice",
			keys:       []string{},
			value:      "value",
			wantErrSet: true,
		},
		{
			name:       "Set key with empty segment (like parent..child in dotted form)",
			keys:       []string{"parent", "", "child"},
			value:      "value",
			wantErrSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.SetKey(tt.keys, tt.value)
			if tt.wantErrSet {
				if err == nil {
					t.Errorf("[%s] Expected error in SetKey but got nil", tt.name)
				}
				return
			} else if err != nil {
				t.Errorf("[%s] Unexpected error in SetKey: %v", tt.name, err)
				return
			}

			got, err := store.GetKey(tt.keys)
			if tt.wantErrGet {
				if err == nil {
					t.Errorf("[%s] Expected error in GetKey but got nil", tt.name)
				}
				return
			} else {
				if err != nil {
					t.Errorf("[%s] Unexpected error in GetKey: %v", tt.name, err)
				} else if !reflect.DeepEqual(got, tt.value) {
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
	initialData := map[string]any{
		"foo":    "bar",
		"parent": map[string]any{"child": "value"},
		"grand": map[string]any{
			"parent": map[string]any{"child": map[string]any{"key": "deep"}},
		},
		"nondeletable": "persist",
		"empty":        map[string]any{"parent": map[string]any{}},
		"list":         []any{1, 2, 3},
		"nonexistent":  nil,
		"another": map[string]any{
			"parent": map[string]any{"child1": "val1", "child2": "val2"},
		},
	}

	store, err := NewMapFileStore(filename, initialData, WithCreateIfNotExists(true))
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	tests := []struct {
		name       string
		keys       []string
		wantErr    bool
		checkExist bool
	}{
		{
			name:       "Delete simple key",
			keys:       []string{"foo"},
			checkExist: true,
		},
		{
			name:       "Delete nested key",
			keys:       []string{"parent", "child"},
			checkExist: true,
		},
		{
			name:       "Delete deep nested key",
			keys:       []string{"grand", "parent", "child", "key"},
			checkExist: true,
		},
		{
			name:    "Delete non-existent key",
			keys:    []string{"does", "not", "exist"},
			wantErr: false,
		},
		{
			name:    "Delete key with empty segment",
			keys:    []string{"parent", "", "child"},
			wantErr: false,
		},
		{
			name:       "Delete empty map",
			keys:       []string{"empty", "parent"},
			checkExist: true,
		},
		{
			name:       "Delete from map with multiple keys",
			keys:       []string{"another", "parent", "child1"},
			checkExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteKey(tt.keys)
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
				_, err := store.GetKey(tt.keys)
				if err == nil {
					t.Errorf(
						"[%s] Expected key %v to be deleted, but it still exists",
						tt.name,
						tt.keys,
					)
				}
			}
		})
	}
}

func TestMapFileStore_SetAll_GetAll(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore.json")
	defaultData := map[string]any{"foo": "bar"}
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
		data          map[string]any
		expectedData  map[string]any
		modifyData    bool
		expectedAfter map[string]any
	}{
		{
			name: "Set and get all simple data",
			data: map[string]any{
				"key1": "value1",
				"key2": 2,
			},
			expectedData: map[string]any{
				"key1": "value1",
				"key2": 2,
			},
		},
		{
			name: "Set and get all nested data",
			data: map[string]any{
				"parent": map[string]any{
					"child": "value",
				},
			},
			expectedData: map[string]any{
				"parent": map[string]any{
					"child": "value",
				},
			},
		},
		{
			name: "Set and get all empty data",
			data: map[string]any{},
		},
		{
			name: "Modify returned data should not affect store",
			data: map[string]any{
				"original": "value",
			},
			expectedData: map[string]any{
				"original": "value",
			},
			modifyData: true,
			expectedAfter: map[string]any{
				"original": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset store to default before each subtest
			err := store.Reset()
			if err != nil {
				t.Fatalf("[%s] Failed to reset store: %v", tt.name, err)
			}

			err = store.SetAll(tt.data)
			if err != nil {
				t.Errorf("[%s] SetAll failed: %v", tt.name, err)
				return
			}

			got, err := store.GetAll(false)
			if err != nil {
				t.Errorf("[%s] Failed to get data: %v", tt.name, err)
				return
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
	defaultData := map[string]any{"foo": "bar"}
	store, err := NewMapFileStore(filename, defaultData, WithCreateIfNotExists(true))
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Re-populate the store
	initialData := map[string]any{
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
	defaultData := map[string]any{"k": "v"}
	store, err := NewMapFileStore(
		filename,
		defaultData,
		WithCreateIfNotExists(true),
		WithAutoFlush(true),
	)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	err = store.SetKey([]string{"foo"}, "bar")
	if err != nil {
		t.Fatalf("SetKey failed: %v", err)
	}

	// Reopen the store
	store2, err := NewMapFileStore(filename, defaultData)
	if err != nil {
		t.Fatalf("Failed to reopen store: %v", err)
	}

	val, err := store2.GetKey([]string{"foo"})
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
	defaultData := map[string]any{"k": "v"}
	store, err := NewMapFileStore(
		filename,
		defaultData,
		WithCreateIfNotExists(true),
		WithAutoFlush(false),
	)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	err = store.SetKey([]string{"foo"}, "bar")
	if err != nil {
		t.Fatalf("SetKey failed: %v", err)
	}

	// Reopen the store
	store2, err := NewMapFileStore(filename, defaultData)
	if err != nil {
		t.Fatalf("Failed to reopen store: %v", err)
	}

	_, err = store2.GetKey([]string{"foo"})
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

	val, err := store3.GetKey([]string{"foo"})
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
	defaultData := map[string]any{"k": "v"}
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

	err = store.SetKey([]string{"foo"}, "bar")
	if err == nil {
		t.Errorf("Expected error in SetKey due to unwritable file, but got nil")
	}

	// Simulate Decode error by writing invalid data into the file
	err = os.Chmod(filename, 0o666)
	if err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	err = os.WriteFile(filename, []byte(`invalid json`), 0o600)
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
	defaultData := map[string]any{"k": "v"}
	store, err := NewMapFileStore(filename, defaultData, WithCreateIfNotExists(true))
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Set nested data
	tests := []struct {
		name      string
		keys      []string
		value     any
		expectErr bool
	}{
		{
			name:  "Set nested map",
			keys:  []string{"parent", "child"},
			value: "value",
		},
		{
			name:  "Set nested map with existing parent",
			keys:  []string{"parent", "anotherChild"},
			value: 123,
		},
		{
			name:  "Set deep nested map",
			keys:  []string{"grand", "parent", "child"},
			value: true,
		},
		{
			name:      "Set value where intermediate is not a map",
			keys:      []string{"parent", "child", "key"},
			value:     "invalid",
			expectErr: true, // 'parent.child' is a string, cannot set 'key' under it
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.SetKey(tt.keys, tt.value)
			if tt.expectErr {
				if err == nil {
					t.Errorf("[%s] Expected error in SetKey but got nil", tt.name)
				}
				return
			} else if err != nil {
				t.Errorf("[%s] SetKey failed: %v", tt.name, err)
				return
			}

			got, err := store.GetKey(tt.keys)
			if err != nil {
				t.Errorf("[%s] GetKey failed: %v", tt.name, err)
				return
			}
			if !reflect.DeepEqual(got, tt.value) {
				t.Errorf("[%s] GetKey returned %v, expected %v", tt.name, got, tt.value)
			}
		})
	}
}

// TestMapFileStore_KeyEncodingDecoding demonstrates how keys are encoded/decoded via WithKeyEncDecGetter.
// We run table-driven sub-tests that verify:
//  1. We can set nested keys, then upon store reload, retrieve them with the *original* plain path.
//  2. If the on-disk data is corrupted (invalid base64 in a key), we get an error when loading.
//  3. Partial path coverage vs. all path coverage, etc.
func TestMapFileStore_KeyEncodingDecoding(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore_key_encdec.json")

	// We'll define a function that returns mockB64KeyEncDec for *all* paths
	// i.e. it encodes/decodes every key in the store.
	keyEncDecGetter := func(pathSoFar []string) encdec.StringEncoderDecoder {
		// Always return the base64 encoder/decoder
		return encdec.Base64StringEncoderDecoder{}
	}

	// Create store with default data and our KeyEncDec
	defaultData := map[string]any{"plainKey": "plainVal"}
	store, err := NewMapFileStore(
		filename,
		defaultData,
		WithCreateIfNotExists(true),
		WithKeyEncDecGetter(keyEncDecGetter),
	)
	if err != nil {
		t.Fatalf("Failed to create store with keyEncDec: %v", err)
	}

	t.Run("Set and Get with Key Encoding - Nested Paths", func(t *testing.T) {
		testCases := []struct {
			name       string
			keys       []string
			value      any
			wantErrSet bool
			wantErrGet bool
		}{
			{
				name:  "simple top-level key",
				keys:  []string{"hello"},
				value: "world",
			},
			{
				name:  "nested key with two levels",
				keys:  []string{"level1", "level2"},
				value: 123,
			},
			{
				name:  "deep nested key",
				keys:  []string{"grandlevel", "parentlevel", "childlevel"},
				value: true,
			},
			{
				name:       "empty key slice",
				keys:       []string{},
				value:      "value",
				wantErrSet: true, // cannot set root
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := store.SetKey(tc.keys, tc.value)
				if tc.wantErrSet {
					if err == nil {
						t.Errorf("[%s] Expected error in SetKey but got nil", tc.name)
					}
					return
				} else if err != nil {
					t.Errorf("[%s] Unexpected error in SetKey: %v", tc.name, err)
					return
				}

				gotVal, err := store.GetKey(tc.keys)
				if tc.wantErrGet {
					if err == nil {
						t.Errorf("[%s] Expected error in GetKey but got nil", tc.name)
					}
					return
				}

				if err != nil {
					t.Errorf("[%s] Unexpected error in GetKey: %v", tc.name, err)
					return
				}
				if !reflect.DeepEqual(gotVal, tc.value) {
					t.Errorf("[%s] GetKey returned %v, expected %v", tc.name, gotVal, tc.value)
				}
			})
		}
	})

	t.Run("Reloading the Store Should Decode Keys Correctly", func(t *testing.T) {
		// We set some keys in the previous sub-test. Let's close the store,
		// reopen it, and ensure we get the same data with the same key paths.
		err := store.Save()
		if err != nil {
			t.Fatalf("Unexpected error while saving: %v", err)
		}
		store2, err := NewMapFileStore(
			filename,
			defaultData,
			WithKeyEncDecGetter(keyEncDecGetter),
		)
		if err != nil {
			t.Fatalf("Failed to reopen store after saving: %v", err)
		}

		// Try retrieving some keys we set in the previous sub-test:
		keysToCheck := [][]string{
			{"hello"},
			{"level1", "level2"},
			{"grandlevel", "parentlevel", "childlevel"},
		}
		for _, k := range keysToCheck {
			val, err := store2.GetKey(k)
			if err != nil {
				t.Errorf("GetKey(%v) failed after reopen: %v", k, err)
			} else {
				t.Logf("After reload, got key %v => %v (ok)", k, val)
			}
		}
	})

	t.Run("Invalid Base64 Key on Disk - Load Should Fail", func(t *testing.T) {
		// We'll forcibly write a key that is not valid base64 into the JSON file
		// to mimic a corrupt on-disk scenario.
		// Then, a new store load should fail because decode will error out on that key.
		err := store.Save()
		if err != nil {
			t.Fatalf("Unexpected error while saving: %v", err)
		}

		// Overwrite the store file with one invalid key (like: {"bad-base64??===": "someVal"})
		invalidJSON := `{"bad-base64??===": "someVal"}`
		err = os.WriteFile(filename, []byte(invalidJSON), 0o600)
		if err != nil {
			t.Fatalf("Failed to write invalid base64 key to file: %v", err)
		}

		// Now attempt to open a new store that uses the same KeyEncDec
		_, err = NewMapFileStore(
			filename,
			defaultData,
			WithKeyEncDecGetter(keyEncDecGetter),
		)
		if err == nil {
			t.Errorf("Expected error when loading invalid base64 key from disk, but got nil")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})
}

// Below is an example test enhancement for SetAll/GetAll with key encoding.
// We'll ensure that keys are also re-encoded/de-encoded properly when using SetAll.
func TestMapFileStore_SetAll_KeyEncDec(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "teststore_setall_keyencdec.json")

	keyEncDecGetter := func(pathSoFar []string) encdec.StringEncoderDecoder {
		return encdec.Base64StringEncoderDecoder{}
	}
	defaultData := map[string]any{"default": "val"}

	store, err := NewMapFileStore(
		filename,
		defaultData,
		WithCreateIfNotExists(true),
		WithKeyEncDecGetter(keyEncDecGetter),
	)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	tests := []struct {
		name        string
		data        map[string]any
		keysToCheck [][]string
	}{
		{
			name: "simple set of keys",
			data: map[string]any{
				"alpha":   "bravo",
				"charlie": "delta",
			},
			keysToCheck: [][]string{{"alpha"}, {"charlie"}},
		},
		{
			name: "nested maps inside setAll",
			data: map[string]any{
				"root": map[string]any{
					"inner": "val1",
					"deep": map[string]any{
						"level": 42,
					},
				},
			},
			keysToCheck: [][]string{
				{"root", "inner"},
				{"root", "deep", "level"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := store.SetAll(tc.data)
			if err != nil {
				t.Errorf("[%s] SetAll error: %v", tc.name, err)
				return
			}
			// Now retrieve them to ensure they're set properly
			for _, k := range tc.keysToCheck {
				got, err := store.GetKey(k)
				if err != nil {
					t.Errorf("[%s] GetKey(%v) error: %v", tc.name, k, err)
					continue
				}
				wantVal := getValueAtPath(tc.data, k)
				if !reflect.DeepEqual(got, wantVal) {
					t.Errorf("[%s] mismatch: got %v, want %v for path %v", tc.name, got, wantVal, k)
				}
			}
		})
	}
}

// getValueAtPath is a helper to retrieve the value from a map at the given nested path.
// This is only used in our test expansions for quick checking inside the table-driven tests.
func getValueAtPath(m map[string]any, path []string) any {
	current := any(m)
	for _, p := range path {
		subMap, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = subMap[p]
	}
	return current
}
