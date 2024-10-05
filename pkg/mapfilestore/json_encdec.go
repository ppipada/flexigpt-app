package mapfilestore

import (
	"encoding/json"
	"io"
)

// jsonEncoderDecoder uses JSON encoding/decoding.
type jsonEncoderDecoder struct{}

// Encode encodes the given value into JSON format and writes it to the writer.
func (d jsonEncoderDecoder) Encode(w io.Writer, value interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // For pretty output
	return encoder.Encode(value)
}

// Decode decodes JSON data from the reader into the given value.
func (d jsonEncoderDecoder) Decode(r io.Reader, value interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(value)
}
