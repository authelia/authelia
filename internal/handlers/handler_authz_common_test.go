package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/utils"
)

func TestGetRequestURIFromForwardedHeaders(t *testing.T) {
	testCases := []struct {
		Name          string
		Protocol      []byte
		Host          []byte
		URI           []byte
		Expected      string
		ExpectedPath  string
		ExpectedClean string
		Error         string
	}{
		{
			Name:          "ShouldParseFullURL",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/path?query=value"),
			Expected:      "https://example.com/path?query=value",
			ExpectedClean: "/path?query=value",
		},
		{
			Name:          "ShouldParseWithoutURI",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           nil,
			Expected:      "https://example.com",
			ExpectedClean: ".",
		},
		{
			Name:          "ShouldParseHTTP",
			Protocol:      []byte("http"),
			Host:          []byte("example.com:8080"),
			URI:           []byte("/"),
			Expected:      "http://example.com:8080/",
			ExpectedClean: "/",
		},
		{
			Name:          "ShouldPreserveUnnormalizedDotDotSegments",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/foo/../bar"),
			Expected:      "https://example.com/foo/../bar",
			ExpectedPath:  "/foo/../bar",
			ExpectedClean: "/bar",
		},
		{
			Name:          "ShouldPreserveUnnormalizedSingleDotSegments",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/foo/./bar"),
			Expected:      "https://example.com/foo/./bar",
			ExpectedPath:  "/foo/./bar",
			ExpectedClean: "/foo/bar",
		},
		{
			Name:          "ShouldPreserveTraversalToRoot",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/../../etc/passwd"),
			Expected:      "https://example.com/../../etc/passwd",
			ExpectedPath:  "/../../etc/passwd",
			ExpectedClean: "/etc/passwd",
		},
		{
			Name:          "ShouldPreserveEncodedDotSegments",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/%2e%2e/secret"),
			Expected:      "https://example.com/%2e%2e/secret",
			ExpectedPath:  "/../secret",
			ExpectedClean: "/secret",
		},
		{
			Name:          "ShouldPreserveEncodedDotSegmentsMidPath",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/foo/%2e%2e/bar"),
			Expected:      "https://example.com/foo/%2e%2e/bar",
			ExpectedPath:  "/foo/../bar",
			ExpectedClean: "/bar",
		},
		{
			Name:          "ShouldPreserveEncodedSlashInTraversal",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/foo/..%2fbar"),
			Expected:      "https://example.com/foo/..%2fbar",
			ExpectedPath:  "/foo/../bar",
			ExpectedClean: "/bar",
		},
		{
			Name:          "ShouldPreserveDotSegmentBetweenEncodedSlashes",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/foo%2f..%2fbar"),
			Expected:      "https://example.com/foo%2f..%2fbar",
			ExpectedPath:  "/foo/../bar",
			ExpectedClean: "/bar",
		},
		{
			Name:          "ShouldPreserveDotSegmentBetweenUppercaseEncodedSlashes",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/foo%2F..%2Fbar"),
			Expected:      "https://example.com/foo%2F..%2Fbar",
			ExpectedPath:  "/foo/../bar",
			ExpectedClean: "/bar",
		},
		{
			Name:          "ShouldPreserveEncodedDotSegmentBetweenEncodedSlashes",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/foo%2f%2e%2e%2fbar"),
			Expected:      "https://example.com/foo%2f%2e%2e%2fbar",
			ExpectedPath:  "/foo/../bar",
			ExpectedClean: "/bar",
		},
		{
			Name:          "ShouldPreserveFullyEncodedTraversal",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/%2e%2e%2f%2e%2e%2fetc%2fpasswd"),
			Expected:      "https://example.com/%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			ExpectedPath:  "/../../etc/passwd",
			ExpectedClean: "/etc/passwd",
		},
		{
			Name:          "ShouldPreserveEncodedSlashesAroundDotSegment",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/%2f..%2f"),
			Expected:      "https://example.com/%2f..%2f",
			ExpectedPath:  "//../",
			ExpectedClean: "//",
		},
		{
			Name:          "ShouldPreserveEncodedSpace",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/path%20with%20space"),
			Expected:      "https://example.com/path%20with%20space",
			ExpectedPath:  "/path with space",
			ExpectedClean: "/path with space",
		},
		{
			Name:          "ShouldPreserveDoubleSlash",
			Protocol:      []byte("https"),
			Host:          []byte("example.com"),
			URI:           []byte("/foo//bar"),
			Expected:      "https://example.com/foo//bar",
			ExpectedPath:  "/foo//bar",
			ExpectedClean: "/foo/bar",
		},
		{
			Name:     "ShouldErrorOnMissingProtocol",
			Protocol: nil,
			Host:     []byte("example.com"),
			URI:      []byte("/"),
			Error:    "missing protocol value",
		},
		{
			Name:     "ShouldErrorOnEmptyProtocol",
			Protocol: []byte(""),
			Host:     []byte("example.com"),
			URI:      []byte("/"),
			Error:    "missing protocol value",
		},
		{
			Name:     "ShouldErrorOnMissingHost",
			Protocol: []byte("https"),
			Host:     nil,
			URI:      []byte("/"),
			Error:    "missing host value",
		},
		{
			Name:     "ShouldErrorOnEmptyHost",
			Protocol: []byte("https"),
			Host:     []byte(""),
			URI:      []byte("/"),
			Error:    "missing host value",
		},
		{
			Name:     "ShouldErrorOnInvalidControlCharacter",
			Protocol: []byte("https"),
			Host:     []byte("example.com"),
			URI:      []byte("/path\x00"),
			Error:    "failed to parse forwarded headers: parse \"https://example.com/path\\x00\": net/url: invalid control character in URL",
		},
		{
			Name:     "ShouldErrorOnInvalidCharacterInHost",
			Protocol: []byte("https"),
			Host:     []byte("exa mple.com"),
			URI:      []byte("/"),
			Error:    "failed to parse forwarded headers: parse \"https://exa mple.com/\": invalid character \" \" in host name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			requestURI, err := getRequestURIFromForwardedHeaders(tc.Protocol, tc.Host, tc.URI)

			if tc.Error != "" {
				assert.Nil(t, requestURI)
				require.EqualError(t, err, tc.Error)
			} else {
				require.NoError(t, err)
				require.NotNil(t, requestURI)
				assert.Equal(t, tc.Expected, requestURI.String())

				if tc.ExpectedPath != "" {
					assert.Equal(t, tc.ExpectedPath, requestURI.Path)
				}

				assert.Equal(t, tc.ExpectedClean, utils.URLPathFullClean(requestURI))
			}
		})
	}
}
