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

func TestNewObjectMethodURLCleanURL(t *testing.T) {
	testCases := []struct {
		name                                                            string
		have                                                            string
		havePath                                                        string
		method                                                          string
		error                                                           string
		expectedScheme, expectedDomain, expectedPath, expectedPathClean string
	}{
		{"ShouldNormalizePathSegments", "https://example.com", "/a/../t", fasthttp.MethodGet, "", "https", "example.com", "/t", "/t"},
		{"ShouldNormalizeSinglePathSegments", "https://example.com", "/a/./t", fasthttp.MethodGet, "", "https", "example.com", "/a/t", "/a/t"},
		{"ShouldDecodePathElementsPerSpecification", "https://example.com", "/a/..%2f/t", fasthttp.MethodGet, "", "https", "example.com", "/a/..//t", "/a/..%2F/t"},
		{"ShouldDecodePathElementsPerSpecificationRegardlessOfCase", "https://example.com", "/a/..%2F/t", fasthttp.MethodGet, "", "https", "example.com", "/a/..//t", "/a/..%2F/t"},
		{"ShouldDecodeNonReservedCharactersPerSpecification", "https://example.com", "/%4C/..%2f/%74", fasthttp.MethodGet, "", "https", "example.com", "/L/..//t", "/L/..%2F/t"},
		{"ShouldFailToParseEscapeInDomain", "https://exa%6Dple.com", "/%4C/..%2f/%74", fasthttp.MethodGet, "error occurred parsing object url: parse \"https://exa%6Dple.com/%4C/..%2f/%74\": invalid URL escape \"%6D\"", "", "", "", ""},
		{"ShouldNotDecodePathSegmentEncodedCharacters", "https://example.com", "/a/..%2Ft", fasthttp.MethodGet, "", "https", "example.com", "/a/../t", "/a/..%2Ft"},
		{"ShouldNotDecodePathSegmentEncodedCharactersOnBothSides", "https://example.com", "/a/%2F..%2Ft", fasthttp.MethodGet, "", "https", "example.com", "/a//../t", "/a/%2F..%2Ft"},
		{"ShouldNotDecodeFullyEncodedReservedCharacters", "https://example.com", "/a/%2F%2e%2e%2Ft", fasthttp.MethodGet, "", "https", "example.com", "/a//../t", "/a/%2F..%2Ft"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			object, err := NewObjectMethodURL([]byte(tc.method), []byte(tc.have+tc.havePath))
			if tc.error != "" {
				assert.EqualError(t, err, tc.error)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tc.expectedScheme, object.URL.Scheme)
			assert.Equal(t, tc.expectedDomain, object.Domain)
			assert.Equal(t, tc.expectedPath, object.URL.Path)
			assert.Equal(t, tc.expectedPathClean, object.Path)
			assert.Equal(t, tc.method, object.Method)
		})
	}
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
