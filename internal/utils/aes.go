package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// The implementation of Encrypt and Decrypt methods comes from:
// https://github.com/gtank/cryptopasta.

// Encrypt encrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Output takes the
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Encrypt(plaintext, aad []byte, key *[32]byte) (ciphertext []byte, err error) {
	var (
		block cipher.Block
		gcm   cipher.AEAD
	)

	if block, err = aes.NewCipher(key[:]); err != nil {
		return nil, err
	}

	if gcm, err = cipher.NewGCM(block); err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, aad), nil
}

// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Decrypt(ciphertext, aad []byte, key *[32]byte) (plaintext []byte, err error) {
	var (
		block cipher.Block
		gcm   cipher.AEAD
	)

	if block, err = aes.NewCipher(key[:]); err != nil {
		return nil, err
	}

	if gcm, err = cipher.NewGCM(block); err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil, ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():], aad)
}
