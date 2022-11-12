package oidc

import (
	"context"

	"github.com/go-crypt/crypt"
)

// Compare compares the hash with the data and returns an error if they don't match.
func (h AdaptiveHasher) Compare(_ context.Context, hash, data []byte) (err error) {
	var digest crypt.Digest

	if digest, err = crypt.DecodeWithPlainText(string(hash)); err != nil {
		return err
	}

	if digest.MatchBytes(data) {
		return nil
	}

	return errPasswordsDoNotMatch
}

// Hash creates a new hash from data.
func (h AdaptiveHasher) Hash(_ context.Context, data []byte) (hash []byte, err error) {
	return data, nil
}
