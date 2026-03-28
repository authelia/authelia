package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPBKDF2VariantDefaultIterations(t *testing.T) {
	testCases := []struct {
		name     string
		variant  string
		expected int
	}{
		{"ShouldReturnSHA512ForExplicit", SHA512Lower, 310000},
		{"ShouldReturnSHA512ForEmpty", "", 310000},
		{"ShouldReturnSHA384", SHA384Lower, 280000},
		{"ShouldReturnSHA256", SHA256Lower, 700000},
		{"ShouldReturnSHA224", SHA224Lower, 900000},
		{"ShouldReturnSHA1ForSHA1", "sha1", 1600000},
		{"ShouldReturnSHA1ForUnknown", "unknown", 1600000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, PBKDF2VariantDefaultIterations(tc.variant))
		})
	}
}
