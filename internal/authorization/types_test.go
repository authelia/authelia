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

func TestShouldCleanURL(t *testing.T) {
	testCases := []struct {
		have     string
		havePath string
		method   string

		expectedScheme, expectedDomain, expectedPath, expectedPathClean string
	}{
		{"https://a.com", "/a/../t", fasthttp.MethodGet, "https", "a.com", "/a/../t", "/t"},
		{"https://a.com", "/a/..%2f/t", fasthttp.MethodGet, "https", "a.com", "/a/..//t", "/t"},
		{"https://a.com", "/a/..%2ft", fasthttp.MethodGet, "https", "a.com", "/a/../t", "/t"},
		{"https://a.com", "/a/..%2F/t", fasthttp.MethodGet, "https", "a.com", "/a/..//t", "/t"},
		{"https://a.com", "/a/..%2Ft", fasthttp.MethodGet, "https", "a.com", "/a/../t", "/t"},
		{"https://a.com", "/a/..%2Ft", fasthttp.MethodGet, "https", "a.com", "/a/../t", "/t"},
		{"https://a.com", "/a/%2F..%2Ft", fasthttp.MethodGet, "https", "a.com", "/a//../t", "/t"},
		{"https://a.com", "/a/%2F%2e%2e%2Ft", fasthttp.MethodGet, "https", "a.com", "/a//../t", "/t"},
	}

	for _, tc := range testCases {
		t.Run(tc.have, func(t *testing.T) {
			have, err := url.ParseRequestURI(tc.have + tc.havePath)
			require.NoError(t, err)

			object := NewObject(have, tc.method)

			assert.Equal(t, tc.expectedScheme, object.URL.Scheme)
			assert.Equal(t, tc.expectedDomain, object.Domain)
			assert.Equal(t, tc.expectedPath, object.URL.Path)
			assert.Equal(t, tc.expectedPathClean, object.Path)
			assert.Equal(t, tc.method, object.Method)

			have, err = url.ParseRequestURI(tc.have)
			require.NoError(t, err)

			path, err := url.ParseRequestURI(tc.havePath)
			require.NoError(t, err)

			have.Path, have.RawQuery = path.Path, path.RawQuery

			object = NewObject(have, tc.method)

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
