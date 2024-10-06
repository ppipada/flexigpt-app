package encdec

import (
	"encoding/json"
	"io"
)

// jsonEncoderDecoder uses JSON encoding/decoding.
type JSONEncoderDecoder struct{}

// Encode encodes the given value into JSON format and writes it to the writer.
func (d JSONEncoderDecoder) Encode(w io.Writer, value interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // For pretty output
	return encoder.Encode(value)
}

// Decode decodes JSON data from the reader into the given value.
func (d JSONEncoderDecoder) Decode(r io.Reader, value interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(value)
}
