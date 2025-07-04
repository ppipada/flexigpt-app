package filestore

import (
	"errors"
	"time"

	simplemapdbEncdec "github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
)

// ErrConflict is returned when flush/delete detects that somebody modified the file since we last read or wrote it.
var ErrConflict = errors.New("filestore: concurrent modification detected")

// Operation is the kind of mutation that happened on a file or a key.
type Operation string

const (
	OpSetFile    Operation = "setFile"
	OpResetFile  Operation = "resetFile"
	OpDeleteFile Operation = "deleteFile"
	OpSetKey     Operation = "setKey"
	OpDeleteKey  Operation = "deleteKey"
)

// Event is delivered *after* a mutation has been written to disk.
type Event struct {
	Op Operation
	// Absolute path of the backing JSON file.
	File string
	// Nil for file-level ops.
	Keys []string
	// Nil for OpSetFile / OpResetFile.
	OldValue any
	// Nil for delete.
	NewValue any
	// Deep-copy of the entire map after the change.
	Data      map[string]any
	Timestamp time.Time
}

// Listener is a callback that observes mutations.
type Listener func(Event)

// KeyEncDecGetter: given the path so far, if applicable, returns a StringEncoderDecoder
// It encodes decodes: The key at the path i.e last part of the path array.
type KeyEncDecGetter func(pathSoFar []string) simplemapdbEncdec.StringEncoderDecoder

// ValueEncDecGetter: given the path so far, if applicable, returns a EncoderDecoder
// It encodes decodes: Value at the key i.e value at last part of the path array.
type ValueEncDecGetter func(pathSoFar []string) simplemapdbEncdec.EncoderDecoder
