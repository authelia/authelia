package utils

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldEncryptAndDecriptUsingAES(t *testing.T) {
	var key = sha256.Sum256([]byte("the key"))

	var secret = "the secret"

	encryptedSecret, err := Encrypt([]byte(secret), &key)
	assert.NoError(t, err, "")

	decryptedSecret, err := Decrypt(encryptedSecret, &key)

	assert.NoError(t, err, "")
	assert.Equal(t, secret, string(decryptedSecret))
}

func TestShouldFailDecryptOnInvalidKey(t *testing.T) {
	var key = sha256.Sum256([]byte("the key"))

	var secret = "the secret"

	encryptedSecret, err := Encrypt([]byte(secret), &key)
	assert.NoError(t, err, "")

	key = sha256.Sum256([]byte("the key 2"))

	_, err = Decrypt(encryptedSecret, &key)

	assert.Error(t, err, "message authentication failed")
}

func TestDecrypt(t *testing.T) {
	key := sha256.Sum256([]byte("the key"))

	testCases := []struct {
		name       string
		ciphertext []byte
		err        string
	}{
		{
			"ShouldReturnErrorForTooShortCiphertext",
			[]byte("abc123"),
			"malformed ciphertext",
		},
		{
			"ShouldReturnErrorForEmptyCiphertext",
			[]byte{},
			"malformed ciphertext",
		},
		{
			"ShouldReturnErrorForNilCiphertext",
			nil,
			"malformed ciphertext",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decrypt(tc.ciphertext, &key)

			if len(tc.err) != 0 {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
