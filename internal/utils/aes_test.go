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

func TestShouldFailDecryptOnInvalidCypherText(t *testing.T) {
	var key = sha256.Sum256([]byte("the key"))

	encryptedSecret := []byte("abc123")

	_, err := Decrypt(encryptedSecret, &key)

	assert.Error(t, err, "message authentication failed")
}
