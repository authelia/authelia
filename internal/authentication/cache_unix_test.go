// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

//go:build unix && (amd64 || arm64)

package authentication

import (
	"crypto/sha256"
	"math"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialCacheHMAC_WriteDigestShouldRejectValuesTooLong(t *testing.T) {
	length := int(math.MaxUint32) + 1

	data, err := syscall.Mmap(-1, 0, length, syscall.PROT_READ, syscall.MAP_ANON|syscall.MAP_PRIVATE)
	if err != nil {
		t.Skipf("unable to reserve address space for the test: %v", err)
	}

	t.Cleanup(func() {
		require.NoError(t, syscall.Munmap(data))
	})

	value := unsafe.String(unsafe.SliceData(data), len(data))

	cache := NewCredentialCacheHMAC(sha256.New, 5*time.Minute)

	assert.EqualError(t, cache.writeDigest(sha256.New(), value), "error occurred calculating cache hmac: value is too long: 4294967296 > 4294967295")
}
