package reqresp

import (
	"bytes"
	"encoding/json"
)

// checkForEmptyOrNullData checks if the data is zero or 'null' and returns a standardized error.
func checkForEmptyOrNullData(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return &JSONRPCError{
			Code:    ParseError,
			Message: GetDefaultErrorMessage(ParseError) + ": Received empty data",
		}
	}
	if bytes.Equal(data, []byte("null")) {
		return &JSONRPCError{
			Code:    ParseError,
			Message: GetDefaultErrorMessage(ParseError) + ": Received null data",
		}
	}
	return nil
}

// Generic function to unmarshal BatchItem structures.
func unmarshalBatchItem[T any](data []byte, isBatch *bool, items *[]T) error {
	if err := checkForEmptyOrNullData(data); err != nil {
		return err
	}

	data = bytes.TrimSpace(data)
	// Try to unmarshal into []json.RawMessage to detect if it's a batch.
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(data, &rawMessages); err == nil {
		// Data is a batch.
		*isBatch = true
		// Process each message in the batch, empty slice input is also ok and valid.
		for _, msg := range rawMessages {
			// Empty or null single item also should not be present.
			if err := checkForEmptyOrNullData(msg); err != nil {
				return err
			}
			var item T
			if err := json.Unmarshal(msg, &item); err != nil {
				return &JSONRPCError{
					Code: ParseError,
					Message: GetDefaultErrorMessage(
						ParseError,
					) + ": Failed to unmarshal batch item: " + err.Error(),
				}
			}
			*items = append(*items, item)
		}
	} else {
		var item T
		if err := json.Unmarshal(data, &item); err != nil {
			return &JSONRPCError{
				Code:    ParseError,
				Message: GetDefaultErrorMessage(ParseError) + ": Failed to unmarshal single item: " + err.Error(),
			}
		}
		*isBatch = false
		*items = append(*items, item)
	}
	return nil
}

// Generic function to marshal BatchItem structures.
func marshalBatchItem[T any](isBatch bool, items []T) ([]byte, error) {
	if isBatch {
		return json.Marshal(items)
	}
	if len(items) > 0 {
		return json.Marshal(items[0])
	}
	return json.Marshal(nil)
}

// BatchItem is a generic struct to detect and handle batch Items of any type.
type BatchItem[T any] struct {
	IsBatch bool `json:"-"`
	Items   []T
}

// UnmarshalJSON implements json.Unmarshaler for BatchItem[T].
func (m *BatchItem[T]) UnmarshalJSON(data []byte) error {
	m.Items = make([]T, 0)
	err := unmarshalBatchItem(data, &m.IsBatch, &m.Items)
	return err
}

// MarshalJSON implements json.Marshaler for BatchItem[T].
func (m BatchItem[T]) MarshalJSON() ([]byte, error) {
	return marshalBatchItem(m.IsBatch, m.Items)
}
