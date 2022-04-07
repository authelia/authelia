package oidc

import (
	"context"
	"crypto/subtle"
)

// Compare compares the hash with the data and returns an error if they don't match.
func (h PlainTextHasher) Compare(_ context.Context, hash, data []byte) (err error) {
	if subtle.ConstantTimeCompare(hash, data) == 0 {
		return errPasswordsDoNotMatch
	}

	return nil
}

// Hash creates a new hash from data.
func (h PlainTextHasher) Hash(_ context.Context, data []byte) (hash []byte, err error) {
	return data, nil
}
