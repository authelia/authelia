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
			u, err := url.ParseRequestURI(tc.have)
			require.NoError(t, err)

			actual := URLPathFullClean(u)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func isURLSafe(requestURI string, domain string) bool { //nolint:unparam
	u, _ := url.ParseRequestURI(requestURI)
	return IsURISafeRedirection(u, domain)
}

func TestIsRedirectionSafe_ShouldReturnTrueOnExactDomain(t *testing.T) {
	assert.True(t, isURLSafe("https://example.com", "example.com"))
}

func TestIsRedirectionSafe_ShouldReturnFalseOnBadScheme(t *testing.T) {
	assert.False(t, isURLSafe("http://secure.example.com", "example.com"))
	assert.False(t, isURLSafe("ftp://secure.example.com", "example.com"))
	assert.True(t, isURLSafe("https://secure.example.com", "example.com"))
}

func TestIsRedirectionSafe_ShouldReturnFalseOnBadDomain(t *testing.T) {
	assert.False(t, isURLSafe("https://secure.example.com.c", "example.com"))
	assert.False(t, isURLSafe("https://secure.example.comc", "example.com"))
	assert.False(t, isURLSafe("https://secure.example.co", "example.com"))
}

func TestHasDomainSuffix(t *testing.T) {
	assert.False(t, HasDomainSuffix("abc", ""))
	assert.False(t, HasDomainSuffix("", ""))
}

func TestEqualURLs(t *testing.T) {
	assert.False(t, EqualURLs(MustParseURL(url.Parse("https://google.com/abc#frag")), MustParseURL(url.Parse("https://google.com/abc"))))
	assert.False(t, EqualURLs(&url.URL{Scheme: "https", Host: "example.com", RawFragment: "example"}, &url.URL{Scheme: "https", Host: "example.com"}))

	assert.True(t, EqualURLs(MustParseURL(url.Parse("https://google.com")), MustParseURL(url.Parse("https://google.com"))))
	assert.True(t, EqualURLs(MustParseURL(url.Parse("https://google.com")), MustParseURL(url.Parse("https://Google.com"))))
	assert.True(t, EqualURLs(MustParseURL(url.Parse("https://google.com/abc")), MustParseURL(url.Parse("https://Google.com/abc"))))
	assert.False(t, EqualURLs(MustParseURL(url.Parse("https://google.com/abc")), MustParseURL(url.Parse("https://Google.com/ABC"))))
	assert.False(t, EqualURLs(MustParseURL(url.Parse("https://google.com/abc?abc=1")), MustParseURL(url.Parse("https://Google.com/abc"))))
	assert.False(t, EqualURLs(MustParseURL(url.Parse("https://google2.com/abc")), MustParseURL(url.Parse("https://Google.com/abc"))))
	assert.False(t, EqualURLs(MustParseURL(url.Parse("http://google.com/abc")), MustParseURL(url.Parse("https://Google.com/abc"))))
	assert.True(t, EqualURLs(nil, nil))
	assert.False(t, EqualURLs(nil, MustParseURL(url.Parse("http://google.com/abc"))))
}

func MustParseURL(uri *url.URL, err error) *url.URL {
	if err != nil {
		panic(err)
	}

	return uri
}

func TestIsURLInSlice(t *testing.T) {
	urls := URLsFromStringSlice([]string{"https://google.com", "https://example.com", "https://www.authelia.com/docs"})

	google, err := url.ParseRequestURI("https://google.com")
	assert.NoError(t, err)

	microsoft, err := url.ParseRequestURI("https://microsoft.com")
	assert.NoError(t, err)

	example, err := url.ParseRequestURI("https://example.com")
	assert.NoError(t, err)

	autheliaOne, err := url.ParseRequestURI("https://www.aUthelia.com/docs")
	assert.NoError(t, err)

	autheliaTwo, err := url.ParseRequestURI("https://www.authelia.com/docs")
	assert.NoError(t, err)

	autheliaThree, err := url.ParseRequestURI("https://www.authelia.com/")
	assert.NoError(t, err)

	autheliaFour, err := url.ParseRequestURI("httpS://www.autHelia.com/docs")
	assert.NoError(t, err)

	assert.True(t, IsURLInSlice(google, urls))
	assert.False(t, IsURLInSlice(microsoft, urls))
	assert.True(t, IsURLInSlice(example, urls))

	assert.True(t, IsURLInSlice(autheliaOne, urls))
	assert.True(t, IsURLInSlice(autheliaTwo, urls))
	assert.False(t, IsURLInSlice(autheliaThree, urls))
	assert.True(t, IsURLInSlice(autheliaFour, urls))
}
