package filestore

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"

	simplemapdbEncdec "github.com/flexigpt/flexiui/pkggo/simplemapdb/encdec"
)

// MapFileStore is a file-backed implementation of a thread-safe key-value store.
type MapFileStore struct {
	data              map[string]interface{}
	defaultData       map[string]interface{}
	mu                sync.RWMutex
	encdec            simplemapdbEncdec.EncoderDecoder
	filename          string
	autoFlush         bool
	keyEncDecs        map[string]simplemapdbEncdec.EncoderDecoder
	createIfNotExists bool
}

// Option defines a function type that applies a configuration option to the MapFileStore.
type Option func(*MapFileStore)

// WithEncoder sets a custom encoder/decoder for the store.
func WithEncoder(encoder simplemapdbEncdec.EncoderDecoder) Option {
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

// WithKeyEncoders sets per-key encoder/decoders.
func WithKeyEncoders(keyEncDecs map[string]simplemapdbEncdec.EncoderDecoder) Option {
	return func(store *MapFileStore) {
		store.keyEncDecs = keyEncDecs
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
	defaultData map[string]interface{},
	opts ...Option,
) (*MapFileStore, error) {
	store := &MapFileStore{
		data:        make(map[string]interface{}),
		defaultData: defaultData,
		filename:    filename,
		autoFlush:   true, // Default to true
		keyEncDecs:  make(map[string]simplemapdbEncdec.EncoderDecoder),
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
		return fmt.Errorf("failed to stat file %s: %v", filename, err)
	}

	// File does not exist
	if !store.createIfNotExists {
		return fmt.Errorf("file %s does not exist", filename)
	}

	// Copy default data to store
	store.data = make(map[string]interface{})
	maps.Copy(store.data, store.defaultData)

	// Create directories if needed
	if err := os.MkdirAll(filepath.Dir(filename), os.FileMode(0o770)); err != nil {
		return fmt.Errorf("failed to create directories for file %s: %v", filename, err)
	}

	// Create the file
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filename, err)
	}
	defer f.Close()

	// Flush the store data to the file
	if err := store.flush(); err != nil {
		return fmt.Errorf("failed to flush file %s: %v", filename, err)
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
		return fmt.Errorf("failed to open file %s: %v", store.filename, err)
	}
	defer f.Close()

	// Decode the data from the file
	store.data = make(map[string]interface{})
	if err := store.encdec.Decode(f, &store.data); err != nil {
		return fmt.Errorf("failed to decode data from file %s: %v", store.filename, err)
	}

	// Apply per-key decoders
	if len(store.keyEncDecs) > 0 {
		for key, encDec := range store.keyEncDecs {
			keys := strings.Split(key, ".")
			if err := decodeValueAtPath(store.data, keys, encDec); err != nil {
				if _, ok := err.(*KeyNotFoundError); ok {
					// Key doesn't exist
					continue
				}
				return fmt.Errorf("failed to decode value at key %s: %v", key, err)
			}
		}
	}

	return nil
}

func (store *MapFileStore) flush() error {
	var dataToSave interface{}
	if len(store.keyEncDecs) > 0 {
		// Need to make a copy of store.data and apply per-key encodings
		dataCopy := deepCopyValue(store.data)
		for key, encDec := range store.keyEncDecs {
			keys := strings.Split(key, ".")
			if err := encodeValueAtPath(dataCopy, keys, encDec); err != nil {
				if _, ok := err.(*KeyNotFoundError); ok {
					// Key doesnt exist
					continue
				}
				return fmt.Errorf("failed to encode value at key %s: %v", key, err)
			}
		}
		dataToSave = dataCopy
	} else {
		// No per-key encodings, can use store.data directly
		dataToSave = store.data
	}

	f, err := os.Create(store.filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s for flush: %v", store.filename, err)
	}
	defer f.Close()

	if err := store.encdec.Encode(f, dataToSave); err != nil {
		return fmt.Errorf("failed to encode data to file %s: %v", store.filename, err)
	}
	return nil
}

// Reset removes all data from the store.
func (store *MapFileStore) Reset() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.data = make(map[string]interface{})
	maps.Copy(store.data, store.defaultData)

	if err := store.flush(); err != nil {
		return fmt.Errorf("failed to save data after Reset: %v", err)
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
func (store *MapFileStore) GetAll(forceFetch bool) (map[string]interface{}, error) {
	if forceFetch {
		// Refresh data from the file
		if err := store.load(); err != nil {
			return nil, fmt.Errorf("failed to refresh data from file: %v", err)
		}
	}

	store.mu.RLock()
	defer store.mu.RUnlock()

	// Return a copy of the in-memory data
	dataCopy := make(map[string]interface{})
	maps.Copy(dataCopy, store.data)
	return dataCopy, nil
}

// SetAll overwrites all data in the store with the provided data.
func (store *MapFileStore) SetAll(data map[string]interface{}) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	// Deep copy the input data to prevent external modifications after setting.
	store.data = make(map[string]interface{})
	maps.Copy(store.data, data)

	if store.autoFlush {
		if err := store.flush(); err != nil {
			return fmt.Errorf("failed to save data after SetAll: %v", err)
		}
	}
	return nil
}

// GetKey retrieves the value associated with the given key.
// The key can be a dot-separated path to a nested value.
func (store *MapFileStore) GetKey(key string) (interface{}, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	keys := strings.Split(key, ".")
	val, err := getValueAtPath(store.data, keys)
	if err != nil {
		return nil, err
	}
	return deepCopyValue(val), nil
}

// SetKey sets the value for the given key.
// The key can be a dot-separated path to a nested value.
func (store *MapFileStore) SetKey(key string, value interface{}) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	keys := strings.Split(key, ".")
	if err := setValueAtPath(store.data, keys, value); err != nil {
		return fmt.Errorf("failed to set value at key %s: %v", key, err)
	}

	if store.autoFlush {
		if err := store.flush(); err != nil {
			return fmt.Errorf("failed to save data after SetKey for key %s: %v", key, err)
		}
	}
	return nil
}

// DeleteKey deletes the value associated with the given key.
// The key can be a dot-separated path to a nested value.
func (store *MapFileStore) DeleteKey(key string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	keys := strings.Split(key, ".")
	if err := deleteValueAtPath(store.data, keys); err != nil {
		return fmt.Errorf("failed to delete key %s: %v", key, err)
	}

	if store.autoFlush {
		if err := store.flush(); err != nil {
			return fmt.Errorf("failed to save data after DeleteKey for key %s: %v", key, err)
		}
	}
	return nil
}

// encodeValueAtPath encodes the value at the specified path using encDec and stores the encoded value.
func encodeValueAtPath(
	data interface{},
	keys []string,
	encDec simplemapdbEncdec.EncoderDecoder,
) error {
	parentMap, lastKey, err := navigateToParentMap(data, keys, false)
	if err != nil {
		if _, ok := err.(*KeyNotFoundError); ok {
			// Key doesnt exist so cannot encode, noop
			return nil
		}
		return err
	}
	val, ok := parentMap[lastKey]
	if !ok {
		// Key doesnt exist so cannot encode, noop
		return nil
	}

	// Encode val via encDec
	var buf bytes.Buffer
	if err := encDec.Encode(&buf, val); err != nil {
		return fmt.Errorf("failed to encode value: %v", err)
	}

	// Get the bytes
	encodedBytes := buf.Bytes()

	// Encode the bytes to a base64 string
	encodedStr := base64.StdEncoding.EncodeToString(encodedBytes)

	// Store the base64 string in parentMap[lastKey]
	parentMap[lastKey] = encodedStr

	return nil
}

// decodeValueAtPath decodes the value at the specified path using encDec and replaces it with the decoded value.
func decodeValueAtPath(
	data interface{},
	keys []string,
	encDec simplemapdbEncdec.EncoderDecoder,
) error {
	parentMap, lastKey, err := navigateToParentMap(data, keys, false)
	if err != nil {
		return err
	}
	val, ok := parentMap[lastKey]
	if !ok {
		return &KeyNotFoundError{Key: lastKey, Path: strings.Join(keys, ".")}
	}

	// val should be the encoded value. We need to decode it via encDec

	// We assume val is a string because we encoded and stored it as string in encode func
	strVal, ok := val.(string)
	if !ok {
		return fmt.Errorf("value at key %s is not a string", strings.Join(keys, "."))
	}

	// Decode the base64 string to bytes
	decodedBytes, err := base64.StdEncoding.DecodeString(strVal)
	if err != nil {
		return fmt.Errorf(
			"failed to base64-decode value at key %s: %v",
			strings.Join(keys, "."),
			err,
		)
	}

	// Decode bytes via encDec
	buf := bytes.NewReader(decodedBytes)
	var decodedVal interface{}
	if err := encDec.Decode(buf, &decodedVal); err != nil {
		return fmt.Errorf("failed to decode value at key %s: %v", strings.Join(keys, "."), err)
	}

	// Replace the value in parentMap[lastKey] with the decoded value
	parentMap[lastKey] = decodedVal

	return nil
}
