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
