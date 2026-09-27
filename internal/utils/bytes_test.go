// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesJoin(t *testing.T) {
	a := []byte("a")
	b := []byte("b")

	assert.Equal(t, "ab", string(BytesJoin(a, b)))
	assert.Equal(t, "a", string(BytesJoin(a)))
	assert.Equal(t, "", string(BytesJoin()))
}
