package oidc

import (
	"context"

	"github.com/go-crypt/crypt"

	"github.com/authelia/authelia/v4/internal/logging"
)

// Compare compares the hash with the data and returns an error if they don't match.
func (h AdaptiveHasher) Compare(ctx context.Context, hash, data []byte) (err error) {
	if internal, ok := ctx.Value(ContextKeySecretInternal).(bool); ok && internal {
		logging.Logger().Debugf("skipping compare")
		return nil
	}

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
