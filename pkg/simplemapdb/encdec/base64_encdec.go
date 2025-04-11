package encdec

import (
	"encoding/base64"
	"fmt"
)

// Base64StringEncoderDecoder is a simple KeyEncoderDecoder that uses base64 (URL variant).
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
