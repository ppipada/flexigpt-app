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
	"sync"
	"testing"
	"time"

	"github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
)

type operation interface {
	Execute(store *MapFileStore, t *testing.T)
}

type setKeyOperation struct {
	key   string
	value any
}

func (op setKeyOperation) Execute(store *MapFileStore, t *testing.T) {
	if err := store.SetKey(strings.Split(op.key, "."), op.value); err != nil {
		t.Errorf("failed to set key %s: %v", op.key, err)
	}
}

type getKeyOperation struct {
	key           string
	expectedValue any
}

func (op getKeyOperation) Execute(store *MapFileStore, t *testing.T) {
	val, err := store.GetKey(strings.Split(op.key, "."))
	if err != nil {
		t.Errorf("failed to get key %s: %v", op.key, err)
		return
	}
	if !reflect.DeepEqual(val, op.expectedValue) {
		t.Errorf(
			"value for key %s does not match expected.\ngot: %v\nwant:%v",
			op.key,
			val,
			op.expectedValue,
		)
	}
}

func TestMapFileStore(t *testing.T) {
	tests := []struct {
		name              string
		initialData       map[string]any
		keyEncDecs        map[string]encdec.EncoderDecoder
		operations        []operation
		expectedFinalData map[string]any
	}{
		{
			name: "test with per-key encoders",
			initialData: map[string]any{
				"foo":    "hello",
				"bar":    "world",
				"parent": map[string]any{"child": "secret"},
			},
			keyEncDecs: map[string]encdec.EncoderDecoder{
				// Example: "foo" => reverseStringEncoderDecoder{}, etc.
				"foo":          reverseStringEncoderDecoder{},
				"parent.child": reverseStringEncoderDecoder{},
			},
			operations: []operation{
				setKeyOperation{key: "foo", value: "new value for foo"},
				getKeyOperation{key: "foo", expectedValue: "new value for foo"},
				setKeyOperation{key: "bar", value: "new value for bar"},
				getKeyOperation{key: "bar", expectedValue: "new value for bar"},
			},
			expectedFinalData: map[string]any{
				"foo": "new value for foo",
				"bar": "new value for bar",
				"parent": map[string]any{
					"child": "secret",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file.
			tempDir := t.TempDir()
			filename := filepath.Join(tempDir, "simplemapdb_test.json")

			// Create store with initial data, using the new WithValueEncDecGetter.
			store, err := NewMapFileStore(
				filename,
				tt.initialData,
				WithCreateIfNotExists(true),
				WithValueEncDecGetter(func(pathSoFar []string) encdec.EncoderDecoder {
					joined := strings.Join(pathSoFar, ".")
					if ed, ok := tt.keyEncDecs[joined]; ok {
						return ed
					}
					return nil
				}),
			)
			if err != nil {
				t.Fatalf("failed to create store: %v", err)
			}

			// Flush data.
			if err := store.Flush(); err != nil {
				t.Fatalf("failed to flush data: %v", err)
			}

			// Read raw data from file.
			rawData, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("failed to read raw data from file: %v", err)
			}
			t.Logf("Raw data in file: %s", string(rawData))

			// Unmarshal raw data.
			var fileData map[string]any
			if err := json.Unmarshal(rawData, &fileData); err != nil {
				t.Fatalf("failed to unmarshal raw data: %v", err)
			}

			// Check that the values for the "encoded" keys are properly base64-encoded reversed strings.
			for key := range tt.keyEncDecs {
				keys := strings.Split(key, ".")
				val, err := GetValueAtPath(fileData, keys)
				if err != nil {
					t.Errorf("failed to get value at key %s in file data: %v", key, err)
					continue
				}

				strVal, ok := val.(string)
				if !ok {
					t.Errorf("expected string value at key %s in file data, got %T", key, val)
					continue
				}

				// The value should be a base64-encoded string.
				decodedBytes, err := base64.StdEncoding.DecodeString(strVal)
				if err != nil {
					t.Errorf("failed to base64-decode value at key %s: %v", key, err)
					continue
				}

				// The decoded bytes should be the reversed original string.
				reversedValue := string(decodedBytes)

				// Get the original value from initialData at the same key.
				originalVal, err := GetValueAtPath(tt.initialData, keys)
				if err != nil {
					t.Errorf("failed to get original value at key %s: %v", key, err)
					continue
				}

				origStrVal, ok := originalVal.(string)
				if !ok {
					t.Errorf(
						"expected string value at key %s in initial data, got %T",
						key,
						originalVal,
					)
					continue
				}

				expectedEncodedValue := reverseString(origStrVal)
				if reversedValue != expectedEncodedValue {
					t.Errorf(
						"encoded value at key %s does not match expected reversed value.\ngot: %s\nwant: %s",
						key,
						reversedValue,
						expectedEncodedValue,
					)
				}
			}

			// Now, create a new store from the file (using the same approach).
			newStore, err := NewMapFileStore(
				filename,
				tt.initialData,
				WithCreateIfNotExists(false),
				WithValueEncDecGetter(func(pathSoFar []string) encdec.EncoderDecoder {
					joined := strings.Join(pathSoFar, ".")
					if ed, ok := tt.keyEncDecs[joined]; ok {
						return ed
					}
					return nil
				}),
			)
			if err != nil {
				t.Fatalf("failed to create store from file: %v", err)
			}

			// Perform the user-defined operations.
			for _, op := range tt.operations {
				op.Execute(newStore, t)
			}

			// Flush store after operations.
			if err := newStore.Flush(); err != nil {
				t.Fatalf("failed to flush data after operations: %v", err)
			}

			// Check final in-memory data.
			finalData, err := newStore.GetAll(false)
			if err != nil {
				t.Errorf("Failed to get data err: %v", err)
			}
			if !reflect.DeepEqual(finalData, tt.expectedFinalData) {
				t.Errorf(
					"final data does not match expected.\ngot: %v\nwant:%v",
					finalData,
					tt.expectedFinalData,
				)
			}

			// Verify file contents again after operations.
			rawDataAfterOps, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("failed to read raw data from file after operations: %v", err)
			}
			var fileDataAfterOps map[string]any
			if err := json.Unmarshal(rawDataAfterOps, &fileDataAfterOps); err != nil {
				t.Fatalf("failed to unmarshal raw data after operations: %v", err)
			}

			// Check that the values for the "encoded" keys remain properly encoded.
			for key := range tt.keyEncDecs {
				keys := strings.Split(key, ".")
				val, err := GetValueAtPath(fileDataAfterOps, keys)
				if err != nil {
					t.Errorf(
						"failed to get value at key %s in file data after operations: %v",
						key,
						err,
					)
					continue
				}

				strVal, ok := val.(string)
				if !ok {
					t.Errorf(
						"expected string value at key %s in file data after operations, got %T",
						key,
						val,
					)
					continue
				}

				decodedBytes, err := base64.StdEncoding.DecodeString(strVal)
				if err != nil {
					t.Errorf(
						"failed to base64-decode value at key %s after operations: %v",
						key,
						err,
					)
					continue
				}

				reversedValue := string(decodedBytes)

				// Compare to finalData’s in-memory value.
				finalVal, err := GetValueAtPath(tt.expectedFinalData, keys)
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
					t.Errorf(
						"encoded value at key %s after operations does not match expected reversed value.\ngot: %s\nwant: %s",
						key,
						reversedValue,
						expectedEncodedValue,
					)
				}
			}
		})
	}
}

// Below is your simple "reverse string" EncoderDecoder for demonstration.
type reverseStringEncoderDecoder struct{}

func (e reverseStringEncoderDecoder) Encode(w io.Writer, v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("expected string value, got %T", v)
	}
	reversed := reverseString(s)
	_, err := w.Write([]byte(reversed))
	return err
}

func (e reverseStringEncoderDecoder) Decode(r io.Reader, v any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	reversed := reverseString(string(data))
	ptr, ok := v.(*any)
	if !ok {
		return fmt.Errorf("expected *any pointer, got %T", v)
	}
	*ptr = reversed
	return nil
}

func reverseString(s string) string {
	runes := []rune(s)
	n := len(runes)
	for i := range n / 2 {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}

// Basic event flow
// Sets a couple of keys, deletes them, then resets the file.  We attach a single
// listener that records every event and print a short, deterministic summary.
func Example_events_basicFlow() {
	tmp, _ := os.MkdirTemp("", "fs_example1")
	defer os.RemoveAll(tmp)
	file := filepath.Join(tmp, "store.json")

	// Record every event we receive.
	var mu sync.Mutex
	var got []Event
	rec := func(e Event) {
		mu.Lock()
		defer mu.Unlock()

		// Strip volatile fields so that the output is deterministic.
		e.File = ""
		e.Timestamp = time.Time{}
		got = append(got, e)
	}

	store, _ := NewMapFileStore(
		file,
		// No default data.
		nil,
		WithCreateIfNotExists(true),
		WithListeners(rec),
	)

	_ = store.SetAll(map[string]any{"a": 1})
	_ = store.SetKey([]string{"a"}, 2)
	_ = store.DeleteKey([]string{"a"})
	_ = store.Reset()

	// Pretty-print the recorded events.
	mu.Lock()
	for _, ev := range got {
		switch ev.Op {
		case OpSetFile:
			fmt.Printf("%s → %v\n", ev.Op, ev.Data)
		case OpResetFile:
			fmt.Printf("%s\n", ev.Op)
		default:
			// OpSetKey / OpDeleteKey.
			fmt.Printf("%s %v  old=%v  new=%v\n",
				ev.Op, ev.Keys, ev.OldValue, ev.NewValue)
		}
	}
	mu.Unlock()

	// Output:
	// setFile → map[a:1]
	// setKey [a]  old=1  new=2
	// deleteKey [a]  old=2  new=<nil>
	// resetFile
}

// AutoFlush =false
// Shows that events are still delivered immediately, but the mutation only
// reaches disk after an explicit Flush().
func Example_events_autoFlush() {
	tmp, _ := os.MkdirTemp("", "fs_example2")
	defer os.RemoveAll(tmp)
	file := filepath.Join(tmp, "store.json")

	var last Event
	listener := func(e Event) { last = e }

	st, _ := NewMapFileStore(
		file,
		nil,
		WithCreateIfNotExists(true),
		WithAutoFlush(false),
		WithListeners(listener),
	)

	_ = st.SetKey([]string{"unsaved"}, 123)
	fmt.Println("event op:", last.Op)

	// Re-open the file – the key is not there yet.
	reopen1, _ := NewMapFileStore(file, nil)
	if _, err := reopen1.GetKey([]string{"unsaved"}); err != nil {
		fmt.Println("not on disk yet")
	}

	// Flush and try again.
	_ = st.Flush()
	reopen2, _ := NewMapFileStore(file, nil)
	v, _ := reopen2.GetKey([]string{"unsaved"})
	fmt.Println("on disk after flush:", v)

	// Output:
	// event op: setKey
	// not on disk yet
	// on disk after flush: 123
}

// Panic isolation between listeners
// One listener panics; the second one must still be called.
func Example_events_panicIsolation() {
	tmp, _ := os.MkdirTemp("", "fs_example3")
	defer os.RemoveAll(tmp)
	file := filepath.Join(tmp, "store.json")

	bad := func(Event) { panic("boom") }
	var goodCalled bool
	good := func(Event) { goodCalled = true }

	st, _ := NewMapFileStore(
		file,
		nil,
		WithCreateIfNotExists(true),
		WithListeners(bad, good),
	)

	_ = st.SetKey([]string{"x"}, 1)
	fmt.Println("good listener called:", goodCalled)

	// Output:
	// good listener called: true
}
