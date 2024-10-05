package mapfilestore

import "io"

// EncoderDecoder is an interface that defines methods for encoding and decoding data.
type EncoderDecoder interface {
	Encode(io.Writer, interface{}) error
	Decode(io.Reader, interface{}) error
}

// IMapFileStore defines the interface for the key-value store.
type IMapFileStore interface {
	Save() error
	GetAll() map[string]interface{}
	SetAll(map[string]interface{}) error
	DeleteAll() error
	GetKey(key string) (interface{}, error)
	SetKey(key string, value interface{}) error
	DeleteKey(key string) error
}
