package oidc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestToStringSlice(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		expected []string
	}{
		{
			"ShouldParseStringSlice",
			[]string{"abc", "123"},
			[]string{"abc", "123"},
		},
		{
			"ShouldParseAnySlice",
			[]any{"abc", "123"},
			[]string{"abc", "123"},
		},
		{
			"ShouldParseAnySlice",
			"abc",
			[]string{"abc"},
		},
		{
			"ShouldParseNil",
			nil,
			nil,
		},
		{
			"ShouldParseInt",
			5,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, toStringSlice(tc.have))
		})
	}
}

func TestToTime(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		def      time.Time
		expected time.Time
	}{
		{
			"ShouldParseFloat64",
			float64(123),
			time.Unix(0, 0).UTC(),
			time.Unix(123, 0).UTC(),
		},
		{
			"ShouldParseInt64",
			int64(123),
			time.Unix(0, 0).UTC(),
			time.Unix(123, 0).UTC(),
		},
		{
			"ShouldParseInt",
			123,
			time.Unix(0, 0).UTC(),
			time.Unix(123, 0).UTC(),
		},
		{
			"ShouldParseTime",
			time.Unix(1235, 0).UTC(),
			time.Unix(0, 0).UTC(),
			time.Unix(1235, 0).UTC(),
		},
		{
			"ShouldReturnDefault",
			"abc",
			time.Unix(0, 0).UTC(),
			time.Unix(0, 0).UTC(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, toTime(tc.have, tc.def))
		})
	}
}
