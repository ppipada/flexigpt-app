package filestore

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"sync"
	"time"

	simplemapdbEncdec "github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
)

const maxSetAllRetries = 3

// isSameFileInfo compares inode+device, size and ModTime.
func isSameFileInfo(a, b os.FileInfo) bool {
	return a != nil && b != nil &&
		os.SameFile(a, b) &&
		a.Size() == b.Size() && a.ModTime().Equal(b.ModTime())
}

// MapFileStore is a file-backed implementation of a thread-safe key-value store.
type MapFileStore struct {
	filename    string
	data        map[string]any
	defaultData map[string]any
	mu          sync.RWMutex

	// Snapshot for optimistic CAS (nil = unknown).
	lastStat          os.FileInfo
	encdec            simplemapdbEncdec.EncoderDecoder
	autoFlush         bool
	createIfNotExists bool

	getValueEncDec ValueEncDecGetter
	getKeyEncDec   KeyEncDecGetter
	listeners      []Listener
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

// WithListeners registers one or more listeners during store creation.
func WithListeners(ls ...Listener) Option {
	return func(s *MapFileStore) { s.listeners = append(s.listeners, ls...) }
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
		autoFlush:   true,
		encdec:      simplemapdbEncdec.JSONEncoderDecoder{},
	}

	// Apply options.
	for _, opt := range opts {
		opt(store)
	}

	// Create file if not exists.
	err := store.createFileIfNotExists(filename)
	if err != nil {
		return nil, err
	}

	err = store.load()
	if err != nil {
		return nil, err
	}

	if err := store.rememberStat(); err != nil {
		// File disappeared between load and stat, extremely unlikely.
		return nil, err
	}

	return store, nil
}

// fireEvent delivers e to all listeners, recovering from panics so that a faulty
// observer cannot crash the store.
func (s *MapFileStore) fireEvent(e Event) {
	for _, l := range s.listeners {
		if l == nil {
			continue
		}
		func(cb Listener) {
			defer func() {
				if r := recover(); r != nil {
					// Log.Printf("filestore: listener panic: %v", r).
					slog.Error(
						"Filestore listener panic",
						"err",
						r,
						"event",
						e,
						"stack",
						string(debug.Stack()),
					)
				}
			}()
			cb(e)
		}(l)
	}
}

func (s *MapFileStore) rememberStat() error {
	st, err := os.Stat(s.filename)
	if err != nil {
		// Caller decides whether ENOENT is fatal.
		return err
	}
	s.lastStat = st
	return nil
}

// createFileIfNotExists checks if a file exists and creates it if it doesn't.
func (store *MapFileStore) createFileIfNotExists(filename string) error {
	// Check if the file exists.
	if _, err := os.Stat(filename); err == nil {
		// File exists, nothing to do.
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat file %s: %w", filename, err)
	}

	if !store.createIfNotExists {
		return fmt.Errorf("file %s does not exist", filename)
	}

	// Try to create the file atomically.
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o666)
	if err != nil {
		if os.IsExist(err) {
			// Someone else created it first, nothing to do.
			return nil
		}
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	// We just wanted to create the file, not write to it directly.
	f.Close()

	// Copy default data to store.
	store.data = make(map[string]any)
	maps.Copy(store.data, store.defaultData)

	// Flush the store data to the file.
	if err := store.flushUnlocked(); err != nil {
		return fmt.Errorf("failed to flush file %s: %w", filename, err)
	}

	return nil
}

// load the data from the file into the in-memory store.
func (store *MapFileStore) load() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Open the file.
	f, err := os.Open(store.filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", store.filename, err)
	}
	defer f.Close()

	// Decode the data from the file.
	store.data = make(map[string]any)
	if err := store.encdec.Decode(f, &store.data); err != nil {
		return fmt.Errorf("failed to decode data from file %s: %w", store.filename, err)
	}

	// Do processing in place for load as you want loaded data to be non encoded decoded
	// First process keys in decode mode.
	encodeMode := false
	err = encodeDecodeAllKeysRecursively(store.data, []string{}, store.getKeyEncDec, encodeMode)
	if err != nil {
		return err
	}

	// Then process values in decode mode.
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

	return store.rememberStat()
}

func (store *MapFileStore) flushUnlocked() error {
	// We'll make a deep copy so we don't mutate in-memory.
	// No error as store.data is always a map.
	encodeMode := true
	dataCopy, _ := DeepCopyValue(store.data).(map[string]any)

	// First encode values so that all keys from in mem are non mutated.
	// Encode KEYS next, so that on disk, the providers/modelnames become base64, etc.
	tmpd, err := encodeDecodeAllValuesRecursively(
		dataCopy,
		[]string{},
		store.getValueEncDec,
		encodeMode,
	)
	if err != nil {
		return err
	}
	dataCopy, _ = tmpd.(map[string]any)

	// Encode KEYS next, so that on disk, the providers/modelnames become base64, etc.
	err = encodeDecodeAllKeysRecursively(dataCopy, []string{}, store.getKeyEncDec, encodeMode)
	if err != nil {
		return err
	}

	if store.lastStat != nil {
		// Optimistic CAS check.
		if cur, err := os.Stat(store.filename); err == nil {
			if !isSameFileInfo(cur, store.lastStat) {
				return ErrConflict
			}
			f, permErr := os.OpenFile(store.filename, os.O_WRONLY, 0)
			if permErr != nil {
				return permErr
			}
			f.Close()
		} else if !os.IsNotExist(err) {
			return err
		} else {
			// File vanished, treat as conflict.
			return ErrConflict
		}
	}

	if err := os.MkdirAll(filepath.Dir(store.filename), 0o770); err != nil {
		return fmt.Errorf(
			"failed to ensure directory for file %s for flush: %w",
			store.filename,
			err,
		)
	}
	tmpName := fmt.Sprintf("%s.tmp-%d", store.filename, time.Now().UnixNano())
	tmpFile, err := os.Create(tmpName)
	if err != nil {
		return fmt.Errorf("failed to open file %s for flush: %w", store.filename, err)
	}
	if err := store.encdec.Encode(tmpFile, dataCopy); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to encode data to file %s: %w", store.filename, err)
	}
	tmpFile.Close()
	if store.lastStat != nil {
		_ = os.Chmod(tmpName, store.lastStat.Mode().Perm())
	}

	if err := os.Rename(tmpName, store.filename); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	return store.rememberStat()
}

// Flush writes the current data to the file. No event is emitted for flush.
func (store *MapFileStore) Flush() error {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.flushUnlocked()
}

func (store *MapFileStore) reset() (copyAfter map[string]any, err error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.data = make(map[string]any)
	maps.Copy(store.data, store.defaultData)
	copyAfter, _ = DeepCopyValue(store.data).(map[string]any)

	if err = store.flushUnlocked(); err != nil {
		return nil, fmt.Errorf("failed to save data after Reset: %w", err)
	}
	return copyAfter, nil
}

// Reset removes all data from the store.
func (store *MapFileStore) Reset() error {
	copyAfter, err := store.reset()
	if err != nil {
		return err
	}
	store.fireEvent(Event{
		Op:        OpResetFile,
		File:      store.filename,
		Data:      copyAfter,
		Timestamp: time.Now(),
	})

	return nil
}

// GetAll returns a copy of all data in the store, refreshing from the file first.
func (store *MapFileStore) GetAll(forceFetch bool) (map[string]any, error) {
	if forceFetch {
		stat, err := os.Stat(store.filename)
		if err != nil {
			return nil, fmt.Errorf("failed to stat file: %w", err)
		}
		if !isSameFileInfo(stat, store.lastStat) {
			if err := store.load(); err != nil {
				return nil, fmt.Errorf("failed to reload file: %w", err)
			}
		}
	}
	store.mu.RLock()
	defer store.mu.RUnlock()

	// Return a copy of the in-memory data.
	dataCopy := make(map[string]any)
	maps.Copy(dataCopy, store.data)
	return dataCopy, nil
}

func (store *MapFileStore) setAll(data map[string]any) (copyAfter map[string]any, err error) {
	if data == nil {
		return nil, errors.New("SetAll: nil data")
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	// Deep copy the input data to prevent external modifications after setting.
	store.data = make(map[string]any)
	maps.Copy(store.data, data)
	copyAfter, _ = DeepCopyValue(store.data).(map[string]any)

	if store.autoFlush {
		if err = store.flushUnlocked(); err != nil {
			return nil, fmt.Errorf("failed to save data after SetAll: %w", err)
		}
	}
	return copyAfter, nil
}

// SetAll overwrites all data in the store with the provided data.
// It retries automatically if another writer wins the race and flushUnlocked returns ErrConflict.
func (store *MapFileStore) SetAll(data map[string]any) error {
	if data == nil {
		return errors.New("SetAll: nil data")
	}

	var (
		copyAfter map[string]any
		err       error
	)

	for range maxSetAllRetries {
		copyAfter, err = store.setAll(data)
		if err == nil {
			store.fireEvent(Event{
				Op:        OpSetFile,
				File:      store.filename,
				Data:      copyAfter,
				Timestamp: time.Now(),
			})
			return nil
		}

		// Any error that isn't ErrConflict is fatal.
		if !errors.Is(err, ErrConflict) {
			return err
		}

		// ErrConflict - reload latest on-disk state so that store.lastStat is refreshed, then retry.
		if loadErr := store.load(); loadErr != nil {
			return fmt.Errorf("SetAll conflict reload failed: %w", loadErr)
		}
	}

	return fmt.Errorf("SetAll: %w after %d retries", ErrConflict, maxSetAllRetries)
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

func (store *MapFileStore) setKey(
	keys []string,
	value any,
) (oldVal any, copyAfter map[string]any, err error) {
	if len(keys) == 0 {
		return nil, nil, errors.New("cannot set value at root")
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	oldVal, _ = GetValueAtPath(store.data, keys)
	if err := SetValueAtPath(store.data, keys, value); err != nil {
		return nil, nil, fmt.Errorf("failed to set value at key %v: %w", keys, err)
	}
	copyAfter, _ = DeepCopyValue(store.data).(map[string]any)
	if store.autoFlush {
		if err := store.flushUnlocked(); err != nil {
			return nil, nil, fmt.Errorf(
				"failed to save data after SetKey for keys %v: %w",
				keys,
				err,
			)
		}
	}
	return oldVal, copyAfter, nil
}

// SetKey sets the value for the given key.
// The key can be a dot-separated path to a nested value.
func (store *MapFileStore) SetKey(keys []string, value any) error {
	oldVal, copyAfter, err := store.setKey(keys, value)
	if err != nil {
		return err
	}
	store.fireEvent(Event{
		Op:        OpSetKey,
		File:      store.filename,
		Keys:      slices.Clone(keys),
		OldValue:  DeepCopyValue(oldVal),
		NewValue:  DeepCopyValue(value),
		Data:      copyAfter,
		Timestamp: time.Now(),
	})
	return nil
}

func (store *MapFileStore) deleteKey(
	keys []string,
) (oldVal any, copyAfter map[string]any, err error) {
	if len(keys) == 0 {
		return nil, nil, errors.New("cannot delete value at root")
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	oldVal, _ = GetValueAtPath(store.data, keys)

	if err := DeleteValueAtPath(store.data, keys); err != nil {
		return nil, nil, fmt.Errorf("failed to delete key %v: %w", keys, err)
	}
	copyAfter, _ = DeepCopyValue(store.data).(map[string]any)

	if store.autoFlush {
		if err := store.flushUnlocked(); err != nil {
			return nil, nil, fmt.Errorf(
				"failed to save data after DeleteKey for key %v: %w",
				keys,
				err,
			)
		}
	}
	return oldVal, copyAfter, nil
}

// DeleteKey deletes the value associated with the given key.
// The key can be a dot-separated path to a nested value.
func (store *MapFileStore) DeleteKey(keys []string) error {
	oldVal, copyAfter, err := store.deleteKey(keys)
	if err != nil {
		return err
	}
	store.fireEvent(Event{
		Op:        OpDeleteKey,
		File:      store.filename,
		Keys:      slices.Clone(keys),
		OldValue:  DeepCopyValue(oldVal),
		NewValue:  nil,
		Data:      copyAfter,
		Timestamp: time.Now(),
	})
	return nil
}

// DeleteFile removes the backing file atomically, emits an OpDeleteFile event and clears lastStat.
// Returns ErrConflict if the file changed since we last observed it.
func (store *MapFileStore) DeleteFile() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if store.lastStat != nil {
		if cur, err := os.Stat(store.filename); err == nil {
			if !isSameFileInfo(cur, store.lastStat) {
				return ErrConflict
			}
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	if err := os.Remove(store.filename); err != nil && !os.IsNotExist(err) {
		return err
	}

	store.lastStat = nil
	store.data = make(map[string]any)

	store.fireEvent(Event{
		Op:        OpDeleteFile,
		File:      store.filename,
		Timestamp: time.Now(),
	})
	return nil
}

func (store *MapFileStore) Close() error {
	// Should not flush here as file may be deleted.
	return nil
}

// If "KeyEncDecGetter(pathSoFar)" returns a StringEncoderDecoder, it renames all immediate sub-keys using Encode() or Decode() depending on the mode.
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
		newPath := slices.Clone(pathSoFar)
		newPath = append(newPath, k)
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
		newPath := slices.Clone(pathSoFar)
		newPath = append(newPath, k)
		// If the child's value is a map, keep going.
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
	// If the user has a value-encoder for this path, encode/decode the entire obj here.
	if getValueEncDec != nil {
		valEncDec := getValueEncDec(pathSoFar)
		if valEncDec != nil {
			var (
				buf       bytes.Buffer
				finalVal  any
				base64Str string
			)
			if encodeMode {
				if err := valEncDec.Encode(&buf, obj); err != nil {
					return obj, fmt.Errorf("failed encoding at path %v: %w", pathSoFar, err)
				}
				base64Str = base64.StdEncoding.EncodeToString(buf.Bytes())
				return base64Str, nil
			}

			// Decode mode obj should be a base64-encoded string.
			strVal, ok := obj.(string)
			if !ok {
				// We expected it to be string but found something else, either error or just skip.
				return obj, nil
			}
			rawBytes, err := base64.StdEncoding.DecodeString(strVal)
			if err != nil {
				// Move on or return an error.
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
		// Not a map, nothing left to do.
		return obj, nil
	}

	for k, v := range m {
		newPath := slices.Clone(pathSoFar)
		newPath = append(newPath, k)
		newChild, err := encodeDecodeAllValuesRecursively(
			v,
			newPath,
			getValueEncDec,
			encodeMode,
		)
		if err != nil {
			return obj, err
		}
		// Store the possibly-encoded child back.
		m[k] = newChild
	}
	return m, nil
}
