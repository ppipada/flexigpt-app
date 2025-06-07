package filestore

import (
	"time"

	simplemapdbEncdec "github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
)

// Operation is the kind of mutation that happened on a file or a key.
type Operation string

const (
	OpSetFile   Operation = "setFile"
	OpResetFile Operation = "resetFile"
	OpSetKey    Operation = "setKey"
	OpDeleteKey Operation = "deleteKey"
)

// Event is delivered *after* a mutation has been written to disk.
type Event struct {
	Op Operation
	// absolute path of the backing JSON file
	File string
	// nil for file-level ops
	Keys []string
	// nil for OpSetFile / OpResetFile
	OldValue any
	// nil for delete
	NewValue any
	// deep-copy of the *entire* map after the change
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
