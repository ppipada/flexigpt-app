package encdec

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// jsonEncoderDecoder uses JSON encoding/decoding.
type JSONEncoderDecoder struct{}

// Encode encodes the given value into JSON format and writes it to the writer.
func (d JSONEncoderDecoder) Encode(w io.Writer, value interface{}) error {
	if w == nil {
		return errors.New("writer cannot be nil")
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // For pretty output

	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	return nil
}

// Decode decodes JSON data from the reader into the given value.
func (d JSONEncoderDecoder) Decode(r io.Reader, value interface{}) error {
	if r == nil {
		return errors.New("reader cannot be nil")
	}

	if value == nil {
		return errors.New("value cannot be nil")
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("value must be a non-nil pointer")
	}

	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

func StructWithJSONTagsToMap(data interface{}) (map[string]interface{}, error) {
	// Check if the input is nil
	if data == nil {
		return nil, errors.New("input data cannot be nil")
	}
	// Marshal the struct to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct to JSON: %w", err)
	}

	// Unmarshal the JSON into a map
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	return result, nil
}

func MapToStructWithJSONTags(data map[string]interface{}, out interface{}) error {
	if data == nil {
		return errors.New("input data cannot be nil")
	}

	// Check if out is a pointer
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("output parameter must be a non-nil pointer to a struct")
	}

	// Check if out is pointing to a struct
	if v.Elem().Kind() != reflect.Struct {
		return errors.New("output parameter must be a pointer to a struct")
	}

	// Marshal the map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	// Use a JSON decoder with DisallowUnknownFields
	decoder := json.NewDecoder(strings.NewReader(string(jsonData)))
	decoder.DisallowUnknownFields() // This will cause an error if there are unknown fields

	// Unmarshal the JSON into the struct
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("failed to unmarshal JSON to struct: %w", err)
	}

	return nil
}
