package encdec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/zalando/go-keyring"
)

// EncryptedStringValueEncoderDecoder uses your encryption for encoding/decoding.
type EncryptedStringValueEncoderDecoder struct{}

func (e EncryptedStringValueEncoderDecoder) Encode(w io.Writer, value any) error {
	v, ok := value.(string)
	if !ok {
		return errors.New("got non string encode input")
	}
	encryptedData, err := encryptString(v)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(encryptedData))
	return err
}

func (e EncryptedStringValueEncoderDecoder) Decode(r io.Reader, value any) error {
	encryptedData, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	decryptedData, err := decryptString(string(encryptedData))
	if err != nil {
		return err
	}

	// Use reflection to handle the value.
	valuePtr := reflect.ValueOf(value)

	// Check if value is a pointer.
	if valuePtr.Kind() != reflect.Ptr {
		return fmt.Errorf("value must be a pointer. Kind: %v", valuePtr.Kind())
	}

	// Dereference the pointer to get the underlying value.
	valueElem := valuePtr.Elem()

	// If the underlying value is an interface, set the decrypted data directly.
	if valueElem.Kind() == reflect.Interface {
		valueElem.Set(reflect.ValueOf(decryptedData))
		return nil
	}

	// Otherwise, check if the underlying value is a string.
	if valueElem.Kind() != reflect.String {
		return fmt.Errorf(
			"value must be a pointer to a string or interface. Kind: %v",
			valueElem.Kind(),
		)
	}

	// Set the decrypted data to the dereferenced value.
	valueElem.SetString(decryptedData)

	return nil
}

// encryptString encrypts a string using AES-256-GCM and returns a base64-encoded string.
func encryptString(plaintext string) (string, error) {
	key, err := getKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM block cipher mode: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return encoded, nil
}

// decryptString decrypts a base64-encoded string that was encrypted using AES-256-GCM.
func decryptString(encodedCiphertext string) (string, error) {
	key, err := getKey()
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM block cipher mode: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}
	return string(plaintext), nil
}

// getKey retrieves or generates an AES-256 encryption key from the keyring.
// If the key does not exist, it generates a new one, stores it, and returns it.
func getKey() ([]byte, error) {
	const (
		service = "FlexiGPTKeyRingEncDec"
		user    = "user"
		// AES-256 requires a 32-byte key.
		keySize = 32
	)

	// Attempt to retrieve the key from the keyring.
	keyStr, err := keyring.Get(service, user)
	switch {
	case err == nil:
		// Decode the base64-encoded key.
		key, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode key: %w", err)
		}
		return key, nil
	case errors.Is(err, keyring.ErrNotFound):
		// Generate a new 32-byte key if not found.
		key := make([]byte, keySize)
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			return nil, fmt.Errorf("failed to generate key: %w", err)
		}
		// Store the key in the keyring.
		keyStr := base64.StdEncoding.EncodeToString(key)
		if err := keyring.Set(service, user, keyStr); err != nil {
			return nil, fmt.Errorf("failed to store key in keyring: %w", err)
		}
		return key, nil
	default:
		return nil, fmt.Errorf("failed to retrieve key from keyring: %w", err)
	}
}
