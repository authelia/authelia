package oidc

import (
	"context"
	"crypto/subtle"
)

// AutheliaHasher implements the fosite.Hasher interface without an actual hashing algo.
type AutheliaHasher struct {
}

// Compare compares the hash with the data and returns an error if they don't match.
func (h AutheliaHasher) Compare(ctx context.Context, hash, data []byte) (err error) {
	if subtle.ConstantTimeCompare(hash, data) == 0 {
		return errPasswordsDoNotMatch
	}

	return nil
}

// Hash creates a new hash from data.
func (h AutheliaHasher) Hash(ctx context.Context, data []byte) (hash []byte, err error) {
	return data, nil
}
