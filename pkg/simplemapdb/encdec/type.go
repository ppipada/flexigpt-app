package encdec

import "io"

// EncoderDecoder is an interface that defines methods for encoding and decoding data.
type EncoderDecoder interface {
	Encode(w io.Writer, value any) error
	Decode(r io.Reader, value any) error
}

type StringEncoderDecoder interface {
	Encode(plain string) string
	Decode(encoded string) (string, error)
}
