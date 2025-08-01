package encdec

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Base64StringEncoderDecoder is a simple KeyEncoderDecoder that uses base64.
type Base64StringEncoderDecoder struct{}

func (b Base64StringEncoderDecoder) Encode(plain string) string {
	return base64.StdEncoding.EncodeToString([]byte(plain))
}

func (b Base64StringEncoderDecoder) Decode(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to base64-decode %q: %w", encoded, err)
	}
	return string(raw), nil
}

func Base64JSONEncode[T any](t T) string {
	raw, _ := json.Marshal(t)
	return base64.StdEncoding.EncodeToString(raw)
}

func Base64JSONDecode[T any](s string) (T, error) {
	var t T
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(raw, &t)
	return t, err
}

// ComputeSHA returns the hex SHA-256 of the given string.
func ComputeSHA(in string) string {
	sum := sha256.Sum256([]byte(in))
	return hex.EncodeToString(sum[:])
}
