package authorization

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldAppendQueryParamToURL(t *testing.T) {
	targetURL, err := url.Parse("https://domain.example.com/api?type=none")

	require.NoError(t, err)

	object := NewObject(targetURL, "GET")

	assert.Equal(t, "https", object.URL.Scheme)
	assert.Equal(t, "domain.example.com", object.Domain)
	assert.Equal(t, "/api?type=none", object.Path)
	assert.Equal(t, "GET", object.Method)
}

func TestShouldCreateNewObjectFromRaw(t *testing.T) {
	targetURL, err := url.Parse("https://domain.example.com/api")

	require.NoError(t, err)

	object := NewObjectRaw(targetURL, []byte("GET"))

	assert.Equal(t, "https", object.URL.Scheme)
	assert.Equal(t, "domain.example.com", object.Domain)
	assert.Equal(t, "/api", object.URL.Path)
	assert.Equal(t, "/api", object.Path)
	assert.Equal(t, "GET", object.Method)
}

func TestShouldCleanURL(t *testing.T) {
	testCases := []struct {
		have     string
		havePath string
		method   string

		expectedScheme, expectedDomain, expectedPath, expectedPathClean string
	}{
		{"https://a.com", "/a/../t", "GET", "https", "a.com", "/a/../t", "/t"},
		{"https://a.com", "/a/..%2f/t", "GET", "https", "a.com", "/a/..//t", "/t"},
		{"https://a.com", "/a/..%2ft", "GET", "https", "a.com", "/a/../t", "/t"},
		{"https://a.com", "/a/..%2F/t", "GET", "https", "a.com", "/a/..//t", "/t"},
		{"https://a.com", "/a/..%2Ft", "GET", "https", "a.com", "/a/../t", "/t"},
		{"https://a.com", "/a/..%2Ft", "GET", "https", "a.com", "/a/../t", "/t"},
		{"https://a.com", "/a/%2F..%2Ft", "GET", "https", "a.com", "/a//../t", "/t"},
		{"https://a.com", "/a/%2F%2e%2e%2Ft", "GET", "https", "a.com", "/a//../t", "/t"},
	}

	for _, tc := range testCases {
		t.Run(tc.have, func(t *testing.T) {
			have, err := url.Parse(tc.have + tc.havePath)
			require.NoError(t, err)

			object := NewObject(have, tc.method)

			assert.Equal(t, tc.expectedScheme, object.URL.Scheme)
			assert.Equal(t, tc.expectedDomain, object.Domain)
			assert.Equal(t, tc.expectedPath, object.URL.Path)
			assert.Equal(t, tc.expectedPathClean, object.Path)
			assert.Equal(t, tc.method, object.Method)

			have, err = url.Parse(tc.have)
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
