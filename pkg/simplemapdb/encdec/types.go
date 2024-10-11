package encdec

import "io"

// EncoderDecoder is an interface that defines methods for encoding and decoding data.
type EncoderDecoder interface {
	Encode(io.Writer, interface{}) error
	Decode(io.Reader, interface{}) error
}
