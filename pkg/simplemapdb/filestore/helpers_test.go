package filestore

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// Below is your simple "reverse string" EncoderDecoder for demonstration.
type reverseStringEncoderDecoder struct{}

func (e reverseStringEncoderDecoder) Encode(w io.Writer, v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("expected string value, got %T", v)
	}
	reversed := reverseString(s)
	_, err := w.Write([]byte(reversed))
	return err
}

func (e reverseStringEncoderDecoder) Decode(r io.Reader, v any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	reversed := reverseString(string(data))
	ptr, ok := v.(*any)
	if !ok {
		return fmt.Errorf("expected *any pointer, got %T", v)
	}
	*ptr = reversed
	return nil
}

type operation interface {
	Execute(t *testing.T, store *MapFileStore)
}

type setKeyOperation struct {
	key   string
	value any
}

func (op setKeyOperation) Execute(t *testing.T, store *MapFileStore) {
	t.Helper()
	if err := store.SetKey(strings.Split(op.key, "."), op.value); err != nil {
		t.Errorf("failed to set key %s: %v", op.key, err)
	}
}

type getKeyOperation struct {
	key           string
	expectedValue any
}

func (op getKeyOperation) Execute(t *testing.T, store *MapFileStore) {
	t.Helper()
	val, err := store.GetKey(strings.Split(op.key, "."))
	if err != nil {
		t.Errorf("failed to get key %s: %v", op.key, err)
		return
	}
	if !reflect.DeepEqual(val, op.expectedValue) {
		t.Errorf(
			"value for key %s does not match expected.\ngot: %v\nwant:%v",
			op.key,
			val,
			op.expectedValue,
		)
	}
}

func reverseString(s string) string {
	runes := []rune(s)
	n := len(runes)
	for i := range n / 2 {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}
