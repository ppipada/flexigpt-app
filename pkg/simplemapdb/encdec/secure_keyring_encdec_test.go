package encdec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

// Mock implementations to simulate error conditions and control key retrieval in tests.
// Since we cannot use external packages, we'll simulate key operations.

func TestEncodeDecode(t *testing.T) {
	testCases := []struct {
		desc        string
		input       string
		expectError bool
		tamperData  bool
	}{
		{"Empty string", "", false, false},
		{"Short string", "a", false, false},
		{"Normal string", "This is a test string.", false, false},
		{"Long string", strings.Repeat("a", 1000), false, false},
		{"Unicode string", "こんにちは世界", false, false},
		{"Special characters", "!@#$%^&*()_+-=[]{}|;':\",./<>?", false, false},
		{"Control characters", "\x00\x01\x02", false, false},
		{"Tampered data", "This should fail", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			encoderDecoder := EncryptedStringValueEncoderDecoder{}

			buffer := &bytes.Buffer{}
			err := encoderDecoder.Encode(buffer, tc.input)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			if tc.tamperData {
				data := buffer.Bytes()
				if len(data) > 0 {
					// Tamper with the last byte of the data.
					data[len(data)-1] ^= 0xFF
				}
				buffer = bytes.NewBuffer(data)
			}
			var decodedValue string
			err = encoderDecoder.Decode(buffer, &decodedValue)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Decode failed: %v", err)
				}
				if decodedValue != tc.input {
					t.Errorf("Decoded value does not match the original. Got %q, want %q", decodedValue, tc.input)
				}
			}
		})
	}
}

func TestDecodeInvalidData(t *testing.T) {
	encoderDecoder := EncryptedStringValueEncoderDecoder{}

	testCases := []struct {
		desc         string
		invalidInput string
	}{
		{"Invalid base64 string", "!!! not base64 !!!"},
		{"Empty input", ""},
		{"Whitespace input", "    "},
		{"Non-base64 characters", "@@@###$$$"},
		{"Incomplete base64 padding", "abcd==="},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			var v string
			buffer := bytes.NewBufferString(tc.invalidInput)
			err := encoderDecoder.Decode(buffer, &v)
			if err == nil {
				t.Errorf(
					"Expected error when decoding invalid data '%s', but got none",
					tc.invalidInput,
				)
			}
		})
	}
}

type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestEncodeWithErrorWriter(t *testing.T) {
	encoderDecoder := EncryptedStringValueEncoderDecoder{}

	w := &errorWriter{}
	err := encoderDecoder.Encode(w, "some data")
	if err == nil {
		t.Errorf("Expected error when encoding with erroring writer, but got none")
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestDecodeWithErrorReader(t *testing.T) {
	encoderDecoder := EncryptedStringValueEncoderDecoder{}
	var v string
	r := &errorReader{}
	err := encoderDecoder.Decode(r, &v)
	if err == nil {
		t.Errorf("Expected error when decoding with erroring reader, but got none")
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	invalidBase64Inputs := []string{
		"!!! not base64 !!!",
		"",
		"    ",
		"@@@###$$$",
		"abcd===",
	}

	for _, input := range invalidBase64Inputs {
		t.Run("Invalid base64 input", func(t *testing.T) {
			_, err := decryptString(input)
			if err == nil {
				t.Errorf(
					"Expected error when decrypting invalid base64 input '%s', but got none",
					input,
				)
			}
		})
	}
}

func TestDecryptCiphertextTooShort(t *testing.T) {
	key, err := getKey()
	if err != nil {
		t.Fatalf("getKey failed: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes.NewCipher failed: %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("cipher.NewGCM failed: %v", err)
	}

	nonceSize := aesGCM.NonceSize()
	// Create a ciphertext shorter than nonceSize.
	shortCiphertext := make([]byte, nonceSize-1)

	encoded := base64.StdEncoding.EncodeToString(shortCiphertext)
	_, err = decryptString(encoded)
	if err == nil || err.Error() != "ciphertext too short" {
		t.Errorf("Expected 'ciphertext too short' error, but got %v", err)
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	originalValue := "This is a test"
	encrypted, err := encryptString(originalValue)
	if err != nil {
		t.Fatalf("encryptString failed: %v", err)
	}

	// Modify the encrypted data slightly.
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	// Flip a bit in the ciphertext (after the nonce).
	if len(data) > 0 {
		data[len(data)-1] ^= 0xFF
	}

	tamperedEncrypted := base64.StdEncoding.EncodeToString(data)
	_, err = decryptString(tamperedEncrypted)
	if err == nil {
		t.Errorf("Expected error when decrypting invalid ciphertext, but got none")
	}
}

func TestEncryptDecryptConsistency(t *testing.T) {
	// Ensure that encrypting the same plaintext produces different ciphertexts due to randomness in nonce.
	plaintext := "Consistent plaintext"

	encrypted1, err := encryptString(plaintext)
	if err != nil {
		t.Fatalf("encryptString failed: %v", err)
	}

	encrypted2, err := encryptString(plaintext)
	if err != nil {
		t.Fatalf("encryptString failed: %v", err)
	}

	if encrypted1 == encrypted2 {
		t.Errorf(
			"Expected encrypted outputs to differ due to different nonces, but they were the same",
		)
	}
}

func TestDecodeWithInterface(t *testing.T) {
	encoderDecoder := EncryptedStringValueEncoderDecoder{}

	// Prepare a valid encoded string.
	originalValue := "Test string for interface"
	encodedBuffer := &bytes.Buffer{}
	err := encoderDecoder.Encode(encodedBuffer, originalValue)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode into an any.
	var decodedValue any
	err = encoderDecoder.Decode(encodedBuffer, &decodedValue)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Assert the type and value.
	strValue, ok := decodedValue.(string)
	if !ok {
		t.Errorf("Decoded value is not of type string")
	}
	if strValue != originalValue {
		t.Errorf(
			"Decoded value does not match the original. Got %q, want %q",
			strValue,
			originalValue,
		)
	}
}

func TestDecodeWithNonStringInterface(t *testing.T) {
	encoderDecoder := EncryptedStringValueEncoderDecoder{}

	// Prepare a valid encoded string.
	originalValue := "Test string for non-string interface"
	encodedBuffer := &bytes.Buffer{}
	err := encoderDecoder.Encode(encodedBuffer, originalValue)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode dest that is not a string.
	// Initialize with a non-string type.
	decodedValue := 123
	err = encoderDecoder.Decode(encodedBuffer, &decodedValue)
	if err == nil {
		t.Errorf("Expected error when decoding into a non-string interface, but got none")
	}
}

func TestDecodeWithNilInterface(t *testing.T) {
	encoderDecoder := EncryptedStringValueEncoderDecoder{}

	// Prepare a valid encoded string.
	originalValue := "Test string for nil interface"
	encodedBuffer := &bytes.Buffer{}
	err := encoderDecoder.Encode(encodedBuffer, originalValue)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode into a nil interface.
	var decodedValue any
	err = encoderDecoder.Decode(encodedBuffer, decodedValue)
	if err == nil {
		t.Errorf("Expected error when decoding into a nil interface, but got none")
	}
}
