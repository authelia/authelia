package authorization

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestShouldAppendQueryParamToURL(t *testing.T) {
	targetURL, err := url.ParseRequestURI("https://domain.example.com/api?type=none")

	require.NoError(t, err)

	object := NewObject(targetURL, fasthttp.MethodGet)

	assert.Equal(t, "https", object.URL.Scheme)
	assert.Equal(t, "domain.example.com", object.Domain)
	assert.Equal(t, "/api?type=none", object.Path)
	assert.Equal(t, fasthttp.MethodGet, object.Method)
}

func TestShouldCreateNewObjectFromRaw(t *testing.T) {
	targetURL, err := url.ParseRequestURI("https://domain.example.com/api")

	require.NoError(t, err)

	object := NewObjectRaw(targetURL, []byte(fasthttp.MethodGet))

	assert.Equal(t, "https", object.URL.Scheme)
	assert.Equal(t, "domain.example.com", object.Domain)
	assert.Equal(t, "/api", object.URL.Path)
	assert.Equal(t, "/api", object.Path)
	assert.Equal(t, fasthttp.MethodGet, object.Method)
}

func TestRuleMatchResult_IsPotentialMatch(t *testing.T) {
	testCases := []struct {
		name     string
		have     RuleMatchResult
		expected bool
	}{
		{
			"ShouldNotMatch",
			RuleMatchResult{},
			false,
		},
		{
			"ShouldMatch",
			RuleMatchResult{nil, true, true, true, true, true, true, true, false},
			true,
		},
		{
			"ShouldMatchExact",
			RuleMatchResult{nil, true, true, true, true, true, true, true, true},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.IsPotentialMatch())
		})
	}
}

func TestTypesMisc(t *testing.T) {
	object := &Object{URL: nil}

	assert.Equal(t, "", object.String())
}

func TestNewObjectMethodURL(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		rawURL         string
		expectedMethod string
		expectedDomain string
		expectedPath   string
		err            string
	}{
		{
			"ShouldAllowEmptyMethod",
			"",
			"https://app.example.com/",
			"",
			"app.example.com",
			"/",
			"",
		},
		{
			"ShouldNormalizeLowercaseMethod",
			"get",
			"https://app.example.com/",
			fasthttp.MethodGet,
			"app.example.com",
			"/",
			"",
		},
		{
			"ShouldNormalizeMixedCaseMethod",
			"GeT",
			"https://app.example.com/",
			fasthttp.MethodGet,
			"app.example.com",
			"/",
			"",
		},
		{
			"ShouldAllowUppercaseMethod",
			fasthttp.MethodPost,
			"https://app.example.com/",
			fasthttp.MethodPost,
			"app.example.com",
			"/",
			"",
		},
		{
			"ShouldNormalizeSchemeAndHost",
			fasthttp.MethodGet,
			"HTTPS://APP.EXAMPLE.COM/Path",
			fasthttp.MethodGet,
			"app.example.com",
			"/Path",
			"",
		},
		{
			"ShouldRejectMethodWithDigit",
			"GET1",
			"https://app.example.com/",
			"",
			"",
			"",
			"method header with value 'GET1' has invalid characters",
		},
		{
			"ShouldRejectMethodWithSpace",
			"GET ",
			"https://app.example.com/",
			"",
			"",
			"",
			"method header with value 'GET ' has invalid characters",
		},
		{
			"ShouldRejectInvalidURL",
			fasthttp.MethodGet,
			"notaurl",
			"",
			"",
			"",
			"error occurred parsing object url: parse \"notaurl\": invalid URI for request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			object, err := NewObjectMethodURL([]byte(tc.method), []byte(tc.rawURL))

			if tc.err == "" {
				require.NoError(t, err)
				require.NotNil(t, object)

				assert.Equal(t, tc.expectedMethod, object.Method)
				assert.Equal(t, tc.expectedDomain, object.Domain)
				assert.Equal(t, tc.expectedPath, object.Path)
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, object)
			}
		})
	}
}

func TestNewObjectMethodSchemeHostPath(t *testing.T) {
	testCases := []struct {
		name          string
		scheme        []byte
		host          []byte
		path          []byte
		expected      string
		expectedPath  string
		expectedClean string
		err           string
	}{
		{
			"ShouldParseFullURL",
			[]byte("https"), []byte("example.com"), []byte("/path?query=value"),
			"https://example.com/path?query=value", "", "/path?query=value", "",
		},
		{
			"ShouldParseWithoutURI",
			[]byte("https"), []byte("example.com"), nil,
			"https://example.com", "", ".", "",
		},
		{
			"ShouldParseHTTP",
			[]byte("http"), []byte("example.com:8080"), []byte("/"),
			"http://example.com:8080/", "", "/", "",
		},
		{
			"ShouldPreserveUnnormalizedDotDotSegments",
			[]byte("https"), []byte("example.com"), []byte("/foo/../bar"),
			"https://example.com/foo/../bar", "/foo/../bar", "/bar", "",
		},
		{
			"ShouldPreserveUnnormalizedSingleDotSegments",
			[]byte("https"), []byte("example.com"), []byte("/foo/./bar"),
			"https://example.com/foo/./bar", "/foo/./bar", "/foo/bar", "",
		},
		{
			"ShouldPreserveTraversalToRoot",
			[]byte("https"), []byte("example.com"), []byte("/../../etc/passwd"),
			"https://example.com/../../etc/passwd", "/../../etc/passwd", "/etc/passwd", "",
		},
		{
			"ShouldPreserveEncodedDotSegments",
			[]byte("https"), []byte("example.com"), []byte("/%2e%2e/secret"),
			"https://example.com/%2e%2e/secret", "/../secret", "/secret", "",
		},
		{
			"ShouldPreserveEncodedDotSegmentsMidPath",
			[]byte("https"), []byte("example.com"), []byte("/foo/%2e%2e/bar"),
			"https://example.com/foo/%2e%2e/bar", "/foo/../bar", "/bar", "",
		},
		{
			"ShouldPreserveEncodedSlashInTraversal",
			[]byte("https"), []byte("example.com"), []byte("/foo/..%2fbar"),
			"https://example.com/foo/..%2fbar", "/foo/../bar", "/bar", "",
		},
		{
			"ShouldPreserveDotSegmentBetweenEncodedSlashes",
			[]byte("https"), []byte("example.com"), []byte("/foo%2f..%2fbar"),
			"https://example.com/foo%2f..%2fbar", "/foo/../bar", "/bar", "",
		},
		{
			"ShouldPreserveDotSegmentBetweenUppercaseEncodedSlashes",
			[]byte("https"), []byte("example.com"), []byte("/foo%2F..%2Fbar"),
			"https://example.com/foo%2F..%2Fbar", "/foo/../bar", "/bar", "",
		},
		{
			"ShouldPreserveEncodedDotSegmentBetweenEncodedSlashes",
			[]byte("https"), []byte("example.com"), []byte("/foo%2f%2e%2e%2fbar"),
			"https://example.com/foo%2f%2e%2e%2fbar", "/foo/../bar", "/bar", "",
		},
		{
			"ShouldPreserveFullyEncodedTraversal",
			[]byte("https"), []byte("example.com"), []byte("/%2e%2e%2f%2e%2e%2fetc%2fpasswd"),
			"https://example.com/%2e%2e%2f%2e%2e%2fetc%2fpasswd", "/../../etc/passwd", "/etc/passwd", "",
		},
		{
			"ShouldPreserveEncodedSlashesAroundDotSegment",
			[]byte("https"), []byte("example.com"), []byte("/%2f..%2f"),
			"https://example.com/%2f..%2f", "//../", "//", "",
		},
		{
			"ShouldPreserveEncodedSpace",
			[]byte("https"), []byte("example.com"), []byte("/path%20with%20space"),
			"https://example.com/path%20with%20space", "/path with space", "/path with space", "",
		},
		{
			"ShouldPreserveDoubleSlash",
			[]byte("https"), []byte("example.com"), []byte("/foo//bar"),
			"https://example.com/foo//bar", "/foo//bar", "/foo/bar", "",
		},
		{
			"ShouldErrorOnMissingScheme",
			nil, []byte("example.com"), []byte("/"),
			"", "", "", "missing scheme value",
		},
		{
			"ShouldErrorOnEmptyScheme",
			[]byte(""), []byte("example.com"), []byte("/"),
			"", "", "", "missing scheme value",
		},
		{
			"ShouldErrorOnMissingHost",
			[]byte("https"), nil, []byte("/"),
			"", "", "", "missing host value",
		},
		{
			"ShouldErrorOnEmptyHost",
			[]byte("https"), []byte(""), []byte("/"),
			"", "", "", "missing host value",
		},
		{
			"ShouldErrorOnInvalidControlCharacter",
			[]byte("https"), []byte("example.com"), []byte("/path\x00"),
			"", "", "", "error occurred parsing object url: parse \"https://example.com/path\\x00\": net/url: invalid control character in URL",
		},
		{
			"ShouldErrorOnInvalidCharacterInHost",
			[]byte("https"), []byte("exa mple.com"), []byte("/"),
			"", "", "", "error occurred parsing object url: parse \"https://exa mple.com/\": invalid character \" \" in host name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			object, err := NewObjectMethodSchemeHostPath([]byte(fasthttp.MethodGet), tc.scheme, tc.host, tc.path)

			if tc.err == "" {
				require.NoError(t, err)
				require.NotNil(t, object)

				assert.Equal(t, tc.expected, object.URL.String())
				assert.Equal(t, tc.expectedClean, object.Path)

				if tc.expectedPath != "" {
					assert.Equal(t, tc.expectedPath, object.URL.Path)
				}
			} else {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, object)
			}
		})
	}
}
