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
		{"ShouldReturnFullPathSingleSlash", "https://example.com/", "/"},
		{"ShouldReturnFullPathSingleSlashWithQuery", "https://example.com/?query=1&alt=2", "/?query=1&alt=2"},
		{"ShouldReturnFullPathNormal", "https://example.com/test", "/test"},
		{"ShouldReturnFullPathNormalWithSlashSuffix", "https://example.com/test/", "/test/"},
		{"ShouldReturnFullPathNormalWithSlashSuffixAndQuery", "https://example.com/test/?query=1&alt=2", "/test/?query=1&alt=2"},
		{"ShouldReturnFullPathWithQuery", "https://example.com/test?query=1&alt=2", "/test?query=1&alt=2"},
		{"ShouldReturnCleanedPath", "https://example.com/five/../test?query=1&alt=2", "/test?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscaped", "https://example.com/five/..%2ftest?query=1&alt=2", "/test?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscapedExtra", "https://example.com/five/..%2ftest?query=1&alt=2", "/test?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscapedExtraSurrounding", "https://example.com/five/%2f..%2f/test?query=1&alt=2", "/test?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscapedPeriods", "https://example.com/five/%2f%2e%2e%2f/test?query=1&alt=2", "/test?query=1&alt=2"},
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
