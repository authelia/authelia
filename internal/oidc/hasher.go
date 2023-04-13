package oidc

import (
	"context"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-crypt/crypt/algorithm/plaintext"
)

// NewHasher returns a new Hasher.
func NewHasher() (hasher *Hasher, err error) {
	hasher = &Hasher{}

	if hasher.decoder, err = crypt.NewDefaultDecoder(); err != nil {
		return nil, err
	}

	if err = plaintext.RegisterDecoderPlainText(hasher.decoder); err != nil {
		return nil, err
	}

	return hasher, nil
}

// Hasher implements the fosite.Hasher interface and adaptively compares hashes.
type Hasher struct {
	decoder algorithm.DecoderRegister
}

// Compare compares the hash with the data and returns an error if they don't match.
func (h Hasher) Compare(_ context.Context, hash, data []byte) (err error) {
	var digest algorithm.Digest

	if digest, err = h.decoder.Decode(string(hash)); err != nil {
		return err
	}

	if digest.MatchBytes(data) {
		return nil
	}

	return errPasswordsDoNotMatch
}

// Hash creates a new hash from data.
func (h Hasher) Hash(_ context.Context, data []byte) (hash []byte, err error) {
	return data, nil
}
