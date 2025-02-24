package authentication

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredentialCacheHMAC(t *testing.T) {
	cache := NewCredentialCacheHMAC(sha256.New, time.Second*2)

	require.NoError(t, cache.Put("abc", "123"))

	var valid, found bool

	valid, found = cache.Valid("abc", "123")

	assert.True(t, found)
	assert.True(t, valid)

	valid, found = cache.Valid("abc", "123")

	assert.True(t, found)
	assert.True(t, valid)

	time.Sleep(time.Second * 2)

	valid, found = cache.Valid("abc", "123")

	assert.False(t, found)
	assert.False(t, valid)
}
