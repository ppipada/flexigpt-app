package encdec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type JSONEncoderDecoder struct{}

// Encode encodes the given value into JSON format and writes it to the writer.
func (d JSONEncoderDecoder) Encode(w io.Writer, value any) error {
	if w == nil {
		return errors.New("writer cannot be nil")
	}

	encoder := json.NewEncoder(w)
	// For pretty output.
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	return nil
}

// Decode decodes JSON data from the reader into the given value.
func (d JSONEncoderDecoder) Decode(r io.Reader, value any) error {
	if r == nil {
		return errors.New("reader cannot be nil")
	}

	if _, err := requireNonNilPointer(value, "value"); err != nil {
		return err
	}

	decoder := newDecoder(r, true) // Disallow unknown fields
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

func StructWithJSONTagsToMap(data any) (map[string]any, error) {
	if data == nil {
		return nil, errors.New("input data cannot be nil")
	}
	// Marshal the struct to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct to JSON: %w", err)
	}

	// Unmarshal the JSON into a map.
	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	return result, nil
}

func MapToStructWithJSONTags(data map[string]any, out any) error {
	if data == nil {
		return errors.New("input data cannot be nil")
	}

	if _, err := requirePointerToStruct(out, "output parameter"); err != nil {
		return err
	}

	// Marshal the map to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	// Disallow unknown fields, don't require trailing-data check (same behavior as before).
	if err := decodeBytes(jsonData, out, true, false); err != nil {
		return fmt.Errorf("failed to unmarshal JSON to struct: %w", err)
	}

	return nil
}

// EncodeToJSONRaw encodes any value to json.RawMessage.
// No typed method here as value being of a type doesnt really affect its functionality.
func EncodeToJSONRaw(value any) (json.RawMessage, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode JSON: %w", err)
	}
	return json.RawMessage(data), nil
}

// DecodeJSONRaw decodes a json.RawMessage into a typed value T, disallowing unknown fields and rejecting trailing data.
// If raw is empty, or only whitespace, it returns the zero value of T.
func DecodeJSONRaw[T any](raw json.RawMessage) (T, error) {
	var zero T
	if isBlankJSON(raw) {
		return zero, nil
	}

	var v T
	if err := decodeBytes(raw, &v, true, true); err != nil {
		return zero, err
	}
	return v, nil
}

func requirePointerToStruct(p any, name string) (reflect.Value, error) {
	rv, err := requireNonNilPointer(p, name)
	if err != nil {
		return reflect.Value{}, err
	}
	if rv.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("%s must be a pointer to a struct", name)
	}
	return rv, nil
}

func requireNonNilPointer(p any, name string) (reflect.Value, error) {
	if p == nil {
		return reflect.Value{}, fmt.Errorf("%s cannot be nil", name)
	}
	rv := reflect.ValueOf(p)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return reflect.Value{}, fmt.Errorf("%s must be a non-nil pointer", name)
	}
	return rv, nil
}

// decodeBytes decodes JSON bytes into out with options:
// - disallowUnknown: Disallow unknown fields if true.
// - requireEOF: Reject trailing JSON after the first value if true.
func decodeBytes(data []byte, out any, disallowUnknown, requireEOF bool) error {
	dec := newDecoder(bytes.NewReader(data), disallowUnknown)
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	if requireEOF {
		if err := requireNoTrailing(dec); err != nil {
			return err
		}
	}
	return nil
}

func newDecoder(r io.Reader, disallowUnknown bool) *json.Decoder {
	dec := json.NewDecoder(r)
	if disallowUnknown {
		dec.DisallowUnknownFields()
	}
	return dec
}

// requireNoTrailing ensures there is no trailing data after the first JSON value.
func requireNoTrailing(dec *json.Decoder) error {
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("unexpected trailing data after JSON value")
		}
		return fmt.Errorf("trailing data validation: %w", err)
	}
	return nil
}

func isBlankJSON(b []byte) bool {
	return len(bytes.TrimSpace(b)) == 0
}
