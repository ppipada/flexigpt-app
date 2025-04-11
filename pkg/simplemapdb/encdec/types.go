package encdec

import "io"

// EncoderDecoder is an interface that defines methods for encoding and decoding data.
type EncoderDecoder interface {
	Encode(io.Writer, any) error
	Decode(io.Reader, any) error
}

type StringEncoderDecoder interface {
	Encode(plain string) string
	Decode(encoded string) (string, error)
}
