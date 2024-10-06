package filestore

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"maps"
	"os"
	"strings"
	"sync"

	simplemapdbEncdec "github.com/flexigpt/flexiui/pkg/simplemapdb/encdec"
)

// MapFileStore is a file-backed implementation of a thread-safe key-value store.
type MapFileStore struct {
	data              map[string]interface{}
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
func NewMapFileStore(filename string, opts ...Option) (*MapFileStore, error) {
	store := &MapFileStore{
		data:       make(map[string]interface{}),
		filename:   filename,
		autoFlush:  true, // Default to true
		keyEncDecs: make(map[string]simplemapdbEncdec.EncoderDecoder),
		encdec:     simplemapdbEncdec.JSONEncoderDecoder{}, // Default to JSON encoder/decoder
	}

	// Apply options
	for _, opt := range opts {
		opt(store)
	}

	// Open file
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			if !store.createIfNotExists {
				return nil, fmt.Errorf("file %s does not exist", filename)
			}
			// File does not exist but createIfNotExists is true, so save the empty data to create the file.
			if err := store.flush(); err != nil {
				return nil, fmt.Errorf("failed to create file %s: %v", filename, err)
			}
			return store, nil
		}
		return nil, fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer f.Close()

	// Decode data
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := store.encdec.Decode(f, &store.data); err != nil {
		return nil, fmt.Errorf("failed to decode data from file %s: %v", filename, err)
	}

	// Apply per-key decoders
	if len(store.keyEncDecs) > 0 {
		for key, encDec := range store.keyEncDecs {
			keys := strings.Split(key, ".")
			if err := decodeValueAtPath(store.data, keys, encDec); err != nil {
				if _, ok := err.(*KeyNotFoundError); ok {
					// Key doesnt exist
					continue
				}
				return nil, fmt.Errorf("failed to decode value at key %s: %v", key, err)
			}
		}
	}

	return store, nil
}

// Save writes the current data to the file.
func (store *MapFileStore) Save() error {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.flush()
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
		return fmt.Errorf("failed to create file %s: %v", store.filename, err)
	}
	defer f.Close()

	if err := store.encdec.Encode(f, dataToSave); err != nil {
		return fmt.Errorf("failed to encode data to file %s: %v", store.filename, err)
	}
	return nil
}

// GetAll returns a copy of all data in the store.
func (store *MapFileStore) GetAll() map[string]interface{} {
	store.mu.RLock()
	defer store.mu.RUnlock()
	var dataCopy = make(map[string]interface{})
	maps.Copy(dataCopy, store.data)
	return dataCopy
}

// SetAll overwrites all data in the store with the provided data.
func (store *MapFileStore) SetAll(data map[string]interface{}) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	// Deep copy the input data to prevent external modifications after setting.
	maps.Copy(store.data, data)

	if store.autoFlush {
		if err := store.flush(); err != nil {
			return fmt.Errorf("failed to save data after SetAll: %v", err)
		}
	}
	return nil
}

// DeleteAll removes all data from the store.
func (store *MapFileStore) DeleteAll() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.data = make(map[string]interface{})
	if store.autoFlush {
		if err := store.flush(); err != nil {
			return fmt.Errorf("failed to save data after DeleteAll: %v", err)
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
func encodeValueAtPath(data interface{}, keys []string, encDec simplemapdbEncdec.EncoderDecoder) error {
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
func decodeValueAtPath(data interface{}, keys []string, encDec simplemapdbEncdec.EncoderDecoder) error {
	parentMap, lastKey, err := navigateToParentMap(data, keys, false)
	if err != nil {
		return err
	}
	val, ok := parentMap[lastKey]
	if !ok {
		return &KeyNotFoundError{Key: lastKey, Path: strings.Join(keys, ".")}
	}

	// val should be the encoded value. We need to decode it via encDec

	// We assume val is a string
	strVal, ok := val.(string)
	if !ok {
		return fmt.Errorf("value at key %s is not a string", strings.Join(keys, "."))
	}

	// Decode the base64 string to bytes
	decodedBytes, err := base64.StdEncoding.DecodeString(strVal)
	if err != nil {
		return fmt.Errorf("failed to base64-decode value at key %s: %v", strings.Join(keys, "."), err)
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
