package utils

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLPathFullClean(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldReturnFullPathNormal", "https://example.com/test", "/test"},
		{"ShouldReturnFullPathWithQuery", "https://example.com/test?query=1", "/test?query=1"},
		{"ShouldReturnCleanedPath", "https://example.com/five/../test?query=1", "/test?query=1"},
		{"ShouldReturnCleanedPathEscaped", "https://example.com/five/..%2ftest?query=1", "/test?query=1"},
		{"ShouldReturnCleanedPathEscapedExtra", "https://example.com/five/..%2ftest?query=1", "/test?query=1"},
		{"ShouldReturnCleanedPathEscapedExtraSurrounding", "https://example.com/five/%2f..%2f/test?query=1", "/test?query=1"},
		{"ShouldReturnCleanedPathEscapedPeriods", "https://example.com/five/%2f%2e%2e%2f/test?query=1", "/test?query=1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.have)
			require.NoError(t, err)

			actual := URLPathFullClean(u)

			assert.Equal(t, tc.expected, actual)
		})
	}
}
