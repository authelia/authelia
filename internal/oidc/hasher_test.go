package oidc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldNotRaiseErrorOnEqualPasswordsPlainText(t *testing.T) {
	hasher := AdaptiveHasher{}

	a := []byte("abc")
	b := []byte("abc")

	ctx := context.Background()

	err := hasher.Compare(ctx, a, b)

	assert.NoError(t, err)
}

func TestShouldRaiseErrorOnNonEqualPasswordsPlainText(t *testing.T) {
	hasher := AdaptiveHasher{}

	a := []byte("abc")
	b := []byte("abcd")

	ctx := context.Background()

	err := hasher.Compare(ctx, a, b)

	assert.Equal(t, errPasswordsDoNotMatch, err)
}

func TestShouldHashPassword(t *testing.T) {
	hasher := AdaptiveHasher{}

	data := []byte("abc")

	ctx := context.Background()

	hash, err := hasher.Hash(ctx, data)

	assert.NoError(t, err)
	assert.Equal(t, data, hash)
}
