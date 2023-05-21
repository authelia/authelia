package oidc

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	jose "gopkg.in/square/go-jose.v2"
)

func TestIsSigningAlgLess(t *testing.T) {
	assert.False(t, isSigningAlgLess(SigningAlgRSAUsingSHA256, SigningAlgRSAUsingSHA256))
	assert.False(t, isSigningAlgLess(SigningAlgRSAUsingSHA256, SigningAlgHMACUsingSHA256))
	assert.True(t, isSigningAlgLess(SigningAlgHMACUsingSHA256, SigningAlgNone))
	assert.True(t, isSigningAlgLess(SigningAlgHMACUsingSHA256, SigningAlgRSAUsingSHA512))
	assert.True(t, isSigningAlgLess(SigningAlgHMACUsingSHA256, SigningAlgRSAPSSUsingSHA256))
	assert.True(t, isSigningAlgLess(SigningAlgHMACUsingSHA256, SigningAlgECDSAUsingP521AndSHA512))
	assert.True(t, isSigningAlgLess(SigningAlgRSAUsingSHA256, SigningAlgECDSAUsingP521AndSHA512))
	assert.True(t, isSigningAlgLess(SigningAlgECDSAUsingP521AndSHA512, "JS121"))
	assert.False(t, isSigningAlgLess("JS121", SigningAlgECDSAUsingP521AndSHA512))
	assert.False(t, isSigningAlgLess("JS121", "TS512"))
}

func TestSortedJSONWebKey(t *testing.T) {
	testCases := []struct {
		name     string
		have     []jose.JSONWebKey
		expected []jose.JSONWebKey
	}{
		{
			"ShouldOrderByKID",
			[]jose.JSONWebKey{
				{KeyID: "abc"},
				{KeyID: "123"},
			},
			[]jose.JSONWebKey{
				{KeyID: "123"},
				{KeyID: "abc"},
			},
		},
		{
			"ShouldOrderByAlg",
			[]jose.JSONWebKey{
				{Algorithm: "RS256"},
				{Algorithm: "HS256"},
			},
			[]jose.JSONWebKey{
				{Algorithm: "HS256"},
				{Algorithm: "RS256"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(SortedJSONWebKey(tc.have))

			assert.Equal(t, tc.expected, tc.have)
		})
	}
}
