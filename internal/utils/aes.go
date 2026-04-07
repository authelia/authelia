package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// The implementation of Encrypt and Decrypt methods comes from:
// https://github.com/gtank/cryptopasta.

// Encrypt encrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Output takes the
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Encrypt(plaintext []byte, key []byte) (ciphertext []byte, err error) {
	if len(key) != 32 {
		return nil, errors.New("error encrypting data: key must be 256 bits (32 bytes) long")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error encrypting data: error occurred creating the AES cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error encrypting data: error occurred creating the GCM cipher: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("error encrypting data: error occurred generating random nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Decrypt(ciphertext []byte, key []byte) (plaintext []byte, err error) {
	if len(key) != 32 {
		return nil, errors.New("error decrypting data: key must be 256 bits (32 bytes) long")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data: error occurred creating the AES cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data: error occurred creating the GCM cipher: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("error decrypting data: the ciphertext is too short and is malformed")
	}

	if plaintext, err = gcm.Open(nil, ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():], nil); err != nil {
		return nil, fmt.Errorf("error decrypting data: error occurred decrypting the ciphertext: %w", err)
	}

	return plaintext, nil
}

// DeriveCryptographyKey256 derives an encryption or HMAC key from a raw byte slice. It uses the HKDF-SHA256 algorithm to do
// this, which is suitable for AES-256-GCM encryption and HMAC-SHA256.
func DeriveCryptographyKey256(raw []byte, info string) (key []byte, err error) {
	key = make([]byte, 32)

	reader := hkdf.New(sha256.New, raw, nil, []byte(info))

	if _, err = io.ReadFull(reader, key); err != nil {
		return nil, err
	}

	return key, nil
}

// DeriveLegacyEncryptionKey derives am encryption or HMAC key from a raw byte slice. This function is intended only
// for backwards compatibility and should not be used for new keys. Use DeriveCryptographyKey256 instead.
func DeriveLegacyEncryptionKey(raw []byte) (key []byte) {
	sum := sha256.Sum256(raw)

	return sum[:]
}
