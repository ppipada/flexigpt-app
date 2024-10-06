package filestore

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	simplemapdbEncdec "github.com/flexigpt/flexiui/pkg/simplemapdb/encdec"
)

type operation interface {
	Execute(store *MapFileStore, t *testing.T)
}

type setKeyOperation struct {
	key   string
	value interface{}
}

func (op setKeyOperation) Execute(store *MapFileStore, t *testing.T) {
	if err := store.SetKey(op.key, op.value); err != nil {
		t.Errorf("failed to set key %s: %v", op.key, err)
	}
}

type getKeyOperation struct {
	key           string
	expectedValue interface{}
}

func (op getKeyOperation) Execute(store *MapFileStore, t *testing.T) {
	val, err := store.GetKey(op.key)
	if err != nil {
		t.Errorf("failed to get key %s: %v", op.key, err)
		return
	}
	if !reflect.DeepEqual(val, op.expectedValue) {
		t.Errorf("value for key %s does not match expected.\ngot: %v\nwant:%v", op.key, val, op.expectedValue)
	}
}

func TestMapFileStore(t *testing.T) {

	tests := []struct {
		name              string
		initialData       map[string]interface{}
		keyEncDecs        map[string]simplemapdbEncdec.EncoderDecoder
		operations        []operation
		expectedFinalData map[string]interface{}
	}{
		{
			name:        "test with per-key encoders",
			initialData: map[string]interface{}{"foo": "hello", "bar": "world", "parent": map[string]interface{}{"child": "secret"}},
			keyEncDecs: map[string]simplemapdbEncdec.EncoderDecoder{
				// "foo":          encryptedStringValueEncoderDecoder{},
				"foo":          reverseStringEncoderDecoder{},
				"parent.child": reverseStringEncoderDecoder{},
			},
			operations: []operation{
				setKeyOperation{key: "foo", value: "new value for foo"},
				getKeyOperation{key: "foo", expectedValue: "new value for foo"},
				setKeyOperation{key: "bar", value: "new value for bar"},
				getKeyOperation{key: "bar", expectedValue: "new value for bar"},
			},
			expectedFinalData: map[string]interface{}{
				"foo": "new value for foo",
				"bar": "new value for bar",
				"parent": map[string]interface{}{
					"child": "secret",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tempDir := t.TempDir()
			filename := filepath.Join(tempDir, "simplemapdb_test.json")

			// Create store with initial data
			store, err := NewMapFileStore(
				filename,
				WithCreateIfNotExists(true),
				WithKeyEncoders(tt.keyEncDecs),
			)
			if err != nil {
				t.Fatalf("failed to create store: %v", err)
			}

			// Set all data
			if err := store.SetAll(tt.initialData); err != nil {
				t.Fatalf("failed to set initial data: %v", err)
			}

			// Save data
			if err := store.Save(); err != nil {
				t.Fatalf("failed to save data: %v", err)
			}

			// Read raw data from file
			rawData, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("failed to read raw data from file: %v", err)
			}

			// Print raw data
			t.Logf("Raw data in file: %s", string(rawData))

			// Unmarshal raw data to map
			var fileData map[string]interface{}
			if err := json.Unmarshal(rawData, &fileData); err != nil {
				t.Fatalf("failed to unmarshal raw data: %v", err)
			}

			// Check that the values for the encoded keys are properly encoded
			for key := range tt.keyEncDecs {
				keys := strings.Split(key, ".")
				val, err := getValueAtPath(fileData, keys)
				if err != nil {
					t.Errorf("failed to get value at key %s in file data: %v", key, err)
					continue
				}

				strVal, ok := val.(string)
				if !ok {
					t.Errorf("expected string value at key %s in file data, got %T", key, val)
					continue
				}

				// The value should be a base64 encoded string
				decodedBytes, err := base64.StdEncoding.DecodeString(strVal)
				if err != nil {
					t.Errorf("failed to base64-decode value at key %s: %v", key, err)
					continue
				}

				// The decoded bytes should be the reversed string
				reversedValue := string(decodedBytes)

				// Get the initial value from tt.initialData at the same key
				originalVal, err := getValueAtPath(tt.initialData, keys)
				if err != nil {
					t.Errorf("failed to get original value at key %s: %v", key, err)
					continue
				}

				origStrVal, ok := originalVal.(string)
				if !ok {
					t.Errorf("expected string value at key %s in initial data, got %T", key, originalVal)
					continue
				}

				expectedEncodedValue := reverseString(origStrVal)
				if reversedValue != expectedEncodedValue {
					t.Errorf("encoded value at key %s does not match expected reversed value.\ngot: %s\nwant: %s", key, reversedValue, expectedEncodedValue)
				}
			}

			// Now, create a new store by reading from the file
			newStore, err := NewMapFileStore(filename, WithCreateIfNotExists(false), WithKeyEncoders(tt.keyEncDecs))
			if err != nil {
				t.Fatalf("failed to create store from file: %v", err)
			}

			// Perform operations
			for _, op := range tt.operations {
				op.Execute(newStore, t)
			}

			// Save store after operations
			if err := newStore.Save(); err != nil {
				t.Fatalf("failed to save data after operations: %v", err)
			}

			// Get the final data and verify
			finalData := newStore.GetAll()
			if !reflect.DeepEqual(finalData, tt.expectedFinalData) {
				t.Errorf("final data does not match expected.\ngot: %v\nwant:%v", finalData, tt.expectedFinalData)
			}

			// Read raw data from file again
			rawDataAfterOps, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("failed to read raw data from file after operations: %v", err)
			}

			// Unmarshal raw data to map
			var fileDataAfterOps map[string]interface{}
			if err := json.Unmarshal(rawDataAfterOps, &fileDataAfterOps); err != nil {
				t.Fatalf("failed to unmarshal raw data after operations: %v", err)
			}

			// Check that the values for the encoded keys are properly encoded after operations
			for key := range tt.keyEncDecs {
				keys := strings.Split(key, ".")
				val, err := getValueAtPath(fileDataAfterOps, keys)
				if err != nil {
					t.Errorf("failed to get value at key %s in file data after operations: %v", key, err)
					continue
				}

				strVal, ok := val.(string)
				if !ok {
					t.Errorf("expected string value at key %s in file data after operations, got %T", key, val)
					continue
				}

				// The value should be a base64 encoded string
				decodedBytes, err := base64.StdEncoding.DecodeString(strVal)
				if err != nil {
					t.Errorf("failed to base64-decode value at key %s after operations: %v", key, err)
					continue
				}

				// The decoded bytes should be the reversed string
				reversedValue := string(decodedBytes)

				// Get the value from finalData at the same key
				finalVal, err := getValueAtPath(tt.expectedFinalData, keys)
				if err != nil {
					t.Errorf("failed to get final value at key %s: %v", key, err)
					continue
				}

				finalStrVal, ok := finalVal.(string)
				if !ok {
					t.Errorf("expected string value at key %s in final data, got %T", key, finalVal)
					continue
				}

				expectedEncodedValue := reverseString(finalStrVal)
				if reversedValue != expectedEncodedValue {
					t.Errorf("encoded value at key %s after operations does not match expected reversed value.\ngot: %s\nwant: %s", key, reversedValue, expectedEncodedValue)
				}
			}
		})
	}
}

// Helper functions

type reverseStringEncoderDecoder struct{}

func (e reverseStringEncoderDecoder) Encode(w io.Writer, v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("expected string value, got %T", v)
	}
	reversed := reverseString(s)
	_, err := w.Write([]byte(reversed))
	return err
}

func (e reverseStringEncoderDecoder) Decode(r io.Reader, v interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	reversed := reverseString(string(data))
	ptr, ok := v.(*interface{})
	if !ok {
		return fmt.Errorf("expected *interface{} pointer, got %T", v)
	}
	*ptr = reversed
	return nil
}

func reverseString(s string) string {
	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}
