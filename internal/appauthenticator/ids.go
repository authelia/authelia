// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"regexp"
)

var (
	reID       = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)
	reVerifier = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
	reBase64   = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

// IsValidID returns true if the string is a valid server issued id.
func IsValidID(s string) bool {
	return reID.MatchString(s)
}

// IsValidVerifier returns true if the string is a valid verifier (43 base64url characters).
func IsValidVerifier(s string) bool {
	return reVerifier.MatchString(s)
}

// IsValidCodeChallenge returns true if the string is a valid code challenge (43 base64url characters).
func IsValidCodeChallenge(s string) bool {
	return reVerifier.MatchString(s)
}

// IsBase64URL returns true if the string is non-empty unpadded base64url text.
func IsBase64URL(s string) bool {
	return reBase64.MatchString(s)
}

// NewID generates 32 random bytes encoded as base64url.
func NewID() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("error generating random id: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// NewNumber generates a uniformly random zero padded decimal number of the given length (2 to 4).
func NewNumber(length int) (string, error) {
	if length < 2 || length > 4 {
		return "", fmt.Errorf("number length must be between 2 and 4 but was %d", length)
	}

	max := big.NewInt(1)

	for i := 0; i < length; i++ {
		max.Mul(max, big.NewInt(10))
	}

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("error generating random number: %w", err)
	}

	return fmt.Sprintf("%0*d", length, n.Int64()), nil
}

// HashID returns the SHA-256 of the UTF-8 bytes of the string.
func HashID(s string) []byte {
	h := sha256.Sum256([]byte(s))

	return h[:]
}
