package oidc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldNotRaiseErrorOnEqualPasswordsPlainText(t *testing.T) {
	hasher, err := NewHasher()

	require.NoError(t, err)

	a := []byte("$plaintext$abc")
	b := []byte("abc")

	ctx := context.Background()

	assert.NoError(t, hasher.Compare(ctx, a, b))
}

func TestShouldNotRaiseErrorOnEqualPasswordsPlainTextWithSeparator(t *testing.T) {
	hasher, err := NewHasher()

	require.NoError(t, err)

	a := []byte("$plaintext$abc$123")
	b := []byte("abc$123")

	ctx := context.Background()

	assert.NoError(t, hasher.Compare(ctx, a, b))
}

func TestShouldRaiseErrorOnNonEqualPasswordsPlainText(t *testing.T) {
	hasher, err := NewHasher()

	require.NoError(t, err)

	a := []byte("$plaintext$abc")
	b := []byte("abcd")

	ctx := context.Background()

	assert.Equal(t, errPasswordsDoNotMatch, hasher.Compare(ctx, a, b))
}

func TestShouldHashPassword(t *testing.T) {
	hasher := Hasher{}

	data := []byte("abc")

	ctx := context.Background()

	hash, err := hasher.Hash(ctx, data)

	assert.NoError(t, err)
	assert.Equal(t, data, hash)
}
