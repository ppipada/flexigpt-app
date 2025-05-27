package filestore

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync"

	simplemapdbEncdec "github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
)

// KeyEncDecGetter: given the path so far, if applicable, returns a StringEncoderDecoder
// It encodes decodes: The key at the path i.e last part of the path array.
type KeyEncDecGetter func(pathSoFar []string) simplemapdbEncdec.StringEncoderDecoder

// ValueEncDecGetter: given the path so far, if applicable, returns a EncoderDecoder
// It encodes decodes: Value at the key i.e value at last part of the path array.
type ValueEncDecGetter func(pathSoFar []string) simplemapdbEncdec.EncoderDecoder

// MapFileStore is a file-backed implementation of a thread-safe key-value store.
type MapFileStore struct {
	data              map[string]any
	defaultData       map[string]any
	mu                sync.RWMutex
	encdec            simplemapdbEncdec.EncoderDecoder
	filename          string
	autoFlush         bool
	createIfNotExists bool
	getValueEncDec    ValueEncDecGetter
	getKeyEncDec      KeyEncDecGetter
}

// Option defines a function type that applies a configuration option to the MapFileStore.
type Option func(*MapFileStore)

// WithEncoderDecoder sets a custom encoder/decoder for the store.
func WithEncoderDecoder(encoder simplemapdbEncdec.EncoderDecoder) Option {
	return func(store *MapFileStore) {
		store.encdec = encoder
	}
}

// WithAutoFlush sets the AutoFlush option.
func WithAutoFlush(autoFlush bool) Option {
	return func(store *MapFileStore) {
		store.autoFlush = autoFlush
	}
}

// WithValueEncDecsGetter registers the user’s value encoding decoding handler callback.
func WithValueEncDecGetter(valueEncDecGetter ValueEncDecGetter) Option {
	return func(store *MapFileStore) {
		store.getValueEncDec = valueEncDecGetter
	}
}

// WithKeyEncDecsGetter registers the user’s key encoding decoding handler callback.
func WithKeyEncDecGetter(getter KeyEncDecGetter) Option {
	return func(store *MapFileStore) {
		store.getKeyEncDec = getter
	}
}

// WithCreateIfNotExists sets the option to create the file if it does not exist.
func WithCreateIfNotExists(createIfNotExists bool) Option {
	return func(store *MapFileStore) {
		store.createIfNotExists = createIfNotExists
	}
}

// NewMapFileStore initializes a new MapFileStore.
// If the file does not exist and createIfNotExists is false, it returns an error.
func NewMapFileStore(
	filename string,
	defaultData map[string]any,
	opts ...Option,
) (*MapFileStore, error) {
	store := &MapFileStore{
		data:        make(map[string]any),
		defaultData: defaultData,
		filename:    filename,
		autoFlush:   true,                                   // Default to true
		encdec:      simplemapdbEncdec.JSONEncoderDecoder{}, // Default to JSON encoder/decoder
	}

	// Apply options
	for _, opt := range opts {
		opt(store)
	}

	// create file if not exists
	err := store.createFileIfNotExists(filename)
	if err != nil {
		return nil, err
	}

	err = store.load()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// createFileIfNotExists checks if a file exists and creates it if it doesn't.
func (store *MapFileStore) createFileIfNotExists(filename string) error {
	// Check if the file exists
	if _, err := os.Stat(filename); err == nil {
		return nil // File exists, nothing to do
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat file %s: %w", filename, err)
	}

	// File does not exist
	if !store.createIfNotExists {
		return fmt.Errorf("file %s does not exist", filename)
	}

	// Copy default data to store
	store.data = make(map[string]any)
	maps.Copy(store.data, store.defaultData)

	// Create directories if needed
	if err := os.MkdirAll(filepath.Dir(filename), os.FileMode(0o770)); err != nil {
		return fmt.Errorf("failed to create directories for file %s: %w", filename, err)
	}

	// Create the file
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer f.Close()

	// Flush the store data to the file
	if err := store.flush(); err != nil {
		return fmt.Errorf("failed to flush file %s: %w", filename, err)
	}

	return nil
}

// load the data from the file into the in-memory store.
func (store *MapFileStore) load() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Open the file
	f, err := os.Open(store.filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", store.filename, err)
	}
	defer f.Close()

	// Decode the data from the file
	store.data = make(map[string]any)
	if err := store.encdec.Decode(f, &store.data); err != nil {
		return fmt.Errorf("failed to decode data from file %s: %w", store.filename, err)
	}

	// Do processing in place for load as you want laoded data to be non encoded decoded
	// First process keys in decode mode
	encodeMode := false
	err = encodeDecodeAllKeysRecursively(store.data, []string{}, store.getKeyEncDec, encodeMode)
	if err != nil {
		return err
	}

	// Then process values in decode mode
	newObj, err := encodeDecodeAllValuesRecursively(
		store.data,
		[]string{},
		store.getValueEncDec,
		encodeMode,
	)
	if err != nil {
		return err
	}
	store.data, _ = newObj.(map[string]any)

	return nil
}

func (store *MapFileStore) flush() error {
	// We'll make a deep copy so we don't mutate in-memory. no error as store.data is always a map
	encodeMode := true
	dataCopy, _ := DeepCopyValue(store.data).(map[string]any)

	// First encode values so that all keys from in mem are non mutated
	// Encode KEYS next, so that on disk, the providers/modelnames become base64, etc.
	d, err := encodeDecodeAllValuesRecursively(
		dataCopy,
		[]string{},
		store.getValueEncDec,
		encodeMode,
	)
	if err != nil {
		return err
	}
	dataCopy, _ = d.(map[string]any)

	// Encode KEYS next, so that on disk, the providers/modelnames become base64, etc.
	err = encodeDecodeAllKeysRecursively(dataCopy, []string{}, store.getKeyEncDec, encodeMode)
	if err != nil {
		return err
	}

	f, err := os.Create(store.filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s for flush: %w", store.filename, err)
	}
	defer f.Close()

	if err := store.encdec.Encode(f, dataCopy); err != nil {
		return fmt.Errorf("failed to encode data to file %s: %w", store.filename, err)
	}
	return nil
}

// Reset removes all data from the store.
func (store *MapFileStore) Reset() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.data = make(map[string]any)
	maps.Copy(store.data, store.defaultData)

	if err := store.flush(); err != nil {
		return fmt.Errorf("failed to save data after Reset: %w", err)
	}
	return nil
}

// Save writes the current data to the file.
func (store *MapFileStore) Save() error {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.flush()
}

// GetAll returns a copy of all data in the store, refreshing from the file first.
func (store *MapFileStore) GetAll(forceFetch bool) (map[string]any, error) {
	if forceFetch {
		// Refresh data from the file
		if err := store.load(); err != nil {
			return nil, fmt.Errorf("failed to refresh data from file: %w", err)
		}
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	// Return a copy of the in-memory data
	dataCopy := make(map[string]any)
	maps.Copy(dataCopy, store.data)
	return dataCopy, nil
}

// SetAll overwrites all data in the store with the provided data.
func (store *MapFileStore) SetAll(data map[string]any) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	// Deep copy the input data to prevent external modifications after setting.
	store.data = make(map[string]any)
	maps.Copy(store.data, data)

	if store.autoFlush {
		if err := store.flush(); err != nil {
			return fmt.Errorf("failed to save data after SetAll: %w", err)
		}
	}
	return nil
}

// GetKey retrieves the value associated with the given key.
// The key can be a dot-separated path to a nested value.
func (store *MapFileStore) GetKey(keys []string) (any, error) {
	if len(keys) == 0 {
		return nil, errors.New("cannot get value at root")
	}
	store.mu.RLock()
	defer store.mu.RUnlock()

	val, err := GetValueAtPath(store.data, keys)
	if err != nil {
		return nil, err
	}
	return DeepCopyValue(val), nil
}

// SetKey sets the value for the given key.
// The key can be a dot-separated path to a nested value.
func (store *MapFileStore) SetKey(keys []string, value any) error {
	if len(keys) == 0 {
		return errors.New("cannot set value at root")
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	if err := SetValueAtPath(store.data, keys, value); err != nil {
		return fmt.Errorf("failed to set value at key %v: %w", keys, err)
	}

	if store.autoFlush {
		if err := store.flush(); err != nil {
			return fmt.Errorf("failed to save data after SetKey for keys %v: %w", keys, err)
		}
	}
	return nil
}

// DeleteKey deletes the value associated with the given key.
// The key can be a dot-separated path to a nested value.
func (store *MapFileStore) DeleteKey(keys []string) error {
	if len(keys) == 0 {
		return errors.New("cannot delete value at root")
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	if err := DeleteValueAtPath(store.data, keys); err != nil {
		return fmt.Errorf("failed to delete key %v: %w", keys, err)
	}

	if store.autoFlush {
		if err := store.flush(); err != nil {
			return fmt.Errorf("failed to save data after DeleteKey for key %v: %w", keys, err)
		}
	}
	return nil
}

// encodeDecodeAllKeysRecursively walks the data as a map.
// if "KeyEncDecGetter(pathSoFar)" returns a StringEncoderDecoder, it renames all immediate sub-keys using Encode() or Decode() depending on the mode.
// Then it recurses into each sub-value with an updated path.
// Here obj needs to be any as we may get non map objects in recursive traversal, dont do anything.
func encodeDecodeAllKeysRecursively(
	currentMap map[string]any,
	pathSoFar []string,
	getKeyEncDec KeyEncDecGetter,
	encodeMode bool,
) error {
	if getKeyEncDec == nil {
		return nil
	}

	// 1) Collect all needed renames for *this* level
	//    We don't mutate the map while iterating. We'll rename afterwards.
	var renames []struct {
		oldKey, newKey string
		val            any
	}

	for k, v := range currentMap {
		newPath := pathSoFar
		newPath = append(newPath, k)
		// Does the user want to rename THIS child's key?
		if keyEncDec := getKeyEncDec(newPath); keyEncDec != nil {
			if encodeMode {
				newK := keyEncDec.Encode(k)
				if newK != k {
					renames = append(renames, struct {
						oldKey, newKey string
						val            any
					}{k, newK, v})
				}
			} else {
				decodedK, err := keyEncDec.Decode(k)
				if err != nil {
					return fmt.Errorf("failed to decode key %q at path %v: %w", k, newPath, err)
				}
				if decodedK != k {
					renames = append(renames, struct {
						oldKey, newKey string
						val            any
					}{k, decodedK, v})
				}
			}
		}
	}

	// 2) Apply the renames so the map keys reflect the new names.
	//    After this, the child values have new keys in currentMap.
	for _, r := range renames {
		delete(currentMap, r.oldKey)
		currentMap[r.newKey] = r.val
	}

	// 3) Now recurse into each child to see if they also want to rename sub-keys.
	//    Note that if we changed a key from oldK -> newK, we pass newK in pathSoFar.
	for k, v := range currentMap {
		newPath := pathSoFar
		newPath = append(newPath, k)
		// If the child's value is a map, keep going
		if subMap, ok := v.(map[string]any); ok {
			if err := encodeDecodeAllKeysRecursively(subMap, newPath, getKeyEncDec, encodeMode); err != nil {
				return err
			}
		}
		// If it's not a map, there's no deeper "keys" to rename.
	}
	return nil
}

func encodeDecodeAllValuesRecursively(
	obj any,
	pathSoFar []string,
	getValueEncDec ValueEncDecGetter,
	encodeMode bool,
) (any, error) {
	// If the user has a value-encoder for this path, encode/decode the *entire* obj here.
	if getValueEncDec != nil {
		valEncDec := getValueEncDec(pathSoFar)
		if valEncDec != nil {
			var (
				buf       bytes.Buffer
				finalVal  any
				base64Str string
			)
			if encodeMode {
				// Encode obj to bytes -> base64 -> store as string
				if err := valEncDec.Encode(&buf, obj); err != nil {
					return obj, fmt.Errorf("failed encoding at path %v: %w", pathSoFar, err)
				}
				base64Str = base64.StdEncoding.EncodeToString(buf.Bytes())
				return base64Str, nil
			}

			// Decode mode: obj should be a base64-encoded string...
			strVal, ok := obj.(string)
			if !ok {
				// we expected it to be string but found something else, either error or just skip
				return obj, nil
			}
			rawBytes, err := base64.StdEncoding.DecodeString(strVal)
			if err != nil {
				// Move on or return an error
				return obj, fmt.Errorf("failed base64 decode at path %v: %w", pathSoFar, err)
			}
			if err := valEncDec.Decode(bytes.NewReader(rawBytes), &finalVal); err != nil {
				return obj, fmt.Errorf("failed decode at path %v: %w", pathSoFar, err)
			}
			return finalVal, nil
		}
	}

	// If we get here, no special (en/de)coding applies at this node.
	// If obj is a map, recurse its children.
	m, ok := obj.(map[string]any)
	if !ok {
		// Not a map => nothing left to do
		return obj, nil
	}

	for k, v := range m {
		newChild, err := encodeDecodeAllValuesRecursively(
			v,
			append(pathSoFar, k),
			getValueEncDec,
			encodeMode,
		)
		if err != nil {
			return obj, err
		}
		m[k] = newChild // store the possibly-encoded child back
	}
	return m, nil
}
