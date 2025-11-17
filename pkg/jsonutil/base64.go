package jsonutil

import (
	"encoding/base64"
	"encoding/json"
)

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
