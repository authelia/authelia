package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateToken(t *testing.T) {
	sig, err := validateToken("none", nil)
	assert.Equal(t, "", sig)
	assert.EqualError(t, err, "go-jose/go-jose: compact JWS format must have three parts")
}

func TestGetTokenSignature(t *testing.T) {
	sig, err := getTokenSignature("abc.123")
	assert.Equal(t, "", sig)
	assert.EqualError(t, err, "header, body and signature must all be set")
}

func TestAssign(t *testing.T) {
	a := map[string]any{
		"a": "valuea",
		"c": "valuea",
	}

	b := map[string]any{
		"b": "valueb",
		"c": "valueb",
	}

	c := assign(a, b)

	assert.Equal(t, "valuea", c["a"])
	assert.Equal(t, "valueb", c["b"])
	assert.Equal(t, "valuea", c["c"])
}
