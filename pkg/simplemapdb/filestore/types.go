package filestore

import (
	"time"

	simplemapdbEncdec "github.com/ppipada/flexigpt-app/pkg/simplemapdb/encdec"
)

// Operation is the kind of mutation that happened on a file or a key.
type Operation string

const (
	OpSetFile   Operation = "setFile"   // whole file created / replaced
	OpResetFile Operation = "resetFile" // Reset()
	OpSetKey    Operation = "setKey"    // SetKey()
	OpDeleteKey Operation = "deleteKey" // DeleteKey()
)

// Event is delivered *after* a mutation has been written to disk.
type Event struct {
	Op        Operation
	File      string         // absolute path of the backing JSON file
	Keys      []string       // nil for file-level ops
	OldValue  any            // nil for OpSetFile / OpResetFile
	NewValue  any            // nil for delete
	Data      map[string]any // deep-copy of the *entire* map after the change
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
