package authentication

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// SSHA256 type is used for ldap password storage.
type SSHA256 []byte

// Encode encodes the []byte of raw password.
func (pass SSHA256) Encode() ([]byte, error) {
	hash := makeSSHA256Hash(pass, makeSalt())
	b64 := base64.StdEncoding.EncodeToString(hash)
	return []byte(fmt.Sprintf("{SSHA256}%s", b64)), nil
}

// Matches matches the encoded password and the raw password.
func (pass SSHA256) Matches(encodedPassPhrase []byte) bool {
	// strip the {SSHA}.
	eppS := string(encodedPassPhrase)[6:]
	hash, err := base64.StdEncoding.DecodeString(eppS)
	if err != nil {
		return false
	}
	salt := hash[len(hash)-4:]

	sha := sha256.New()
	sha.Write(pass)
	sha.Write(salt)
	sum := sha.Sum(nil)

	return bytes.Equal(sum, hash[:len(hash)-4])
}

// makeSalt make a 4 byte array containing random bytes.
func makeSalt() []byte {
	sbytes := make([]byte, 4)
	if _, err := rand.Read(sbytes); err != nil {
		// this should never happen.
		return []byte("salt")
	}
	return sbytes
}

// makeSSHA256Hash make hasing using SHA-256 with salt. This is not the final output though. You need to append {SSHA} string with base64 of this hash.
func makeSSHA256Hash(passphrase, salt []byte) []byte {
	sha := sha256.New()
	sha.Write(passphrase)
	sha.Write(salt)

	h := sha.Sum(nil)
	return append(h, salt...)
}
