// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"net/url"
	"strings"
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

func TestHasURIDomainSuffix(t *testing.T) {
	assert.True(t, HasURIDomainSuffix(&url.URL{Scheme: "https", Host: "example.com"}, "example.com"))
	assert.True(t, HasURIDomainSuffix(&url.URL{Scheme: "https", Host: "auth.example.com"}, "example.com"))
	assert.False(t, HasURIDomainSuffix(&url.URL{Scheme: "https", Host: "auth.xexample.com"}, "example.com"))
}

func TestHasDomainSuffix(t *testing.T) {
	testCases := []struct {
		Name         string
		Domain       string
		DomainSuffix string
		Expected     bool
	}{
		{"ShouldNotMatchEmptySuffix", "abc", "", false},
		{"ShouldNotMatchEmptyDomainAndEmptySuffix", "", "", false},
		{"ShouldNotMatchEmptyDomain", "", "example.com", false},
		{"ShouldMatchExactEqual", "example.com", "example.com", true},
		{"ShouldMatchExactEqualDifferentCase", "Example.COM", "example.com", true},
		{"ShouldMatchSubdomain", "auth.example.com", "example.com", true},
		{"ShouldMatchSubdomainDifferentCase", "Auth.Example.COM", "example.com", true},
		{"ShouldMatchSubdomainSuffixDifferentCase", "auth.example.com", "Example.COM", true},
		{"ShouldMatchPeriodPrefixedSuffix", "auth.example.com", ".example.com", true},
		{"ShouldMatchPeriodPrefixedSuffixDifferentCase", "Auth.Example.COM", ".example.com", true},
		{"ShouldMatchPeriodPrefixedSuffixEqualToDomain", ".example.com", ".example.com", true},
		{"ShouldNotMatchUnrelatedDomain", "example.org", "example.com", false},
		{"ShouldNotMatchPartialLabel", "xexample.com", "example.com", false},
		{"ShouldNotMatchPartialLabelWithPeriodPrefix", "xexample.com", ".example.com", false},
		{"ShouldNotMatchSuffixLongerThanDomain", "com", "example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, HasDomainSuffix(tc.Domain, tc.DomainSuffix))
		})
	}
}

func TestIsHostnameValid(t *testing.T) {
	testCases := []struct {
		Name     string
		Hostname string
		Expected bool
	}{
		{"ShouldAllowDomain", "example.com", true},
		{"ShouldAllowSubdomain", "auth.example.com", true},
		{"ShouldAllowSingleLabel", "localhost", true},
		{"ShouldAllowUpperCase", "Auth.Example.COM", true},
		{"ShouldAllowHyphen", "my-app.example.com", true},
		{"ShouldAllowDigits", "1app.example123.com", true},
		{"ShouldAllowMaxLengthLabel", strings.Repeat("a", 63) + ".example.com", true},
		{"ShouldAllowMaxLength", strings.Repeat("a.", 126) + "a", true},
		{"ShouldAllowIPv4", "192.168.1.1", true},
		{"ShouldAllowIPv6", "2001:db8::1", true},
		{"ShouldNotAllowEmpty", "", false},
		{"ShouldNotAllowUnderscore", "foo_bar.example.com", false},
		{"ShouldNotAllowEmptyLabel", "auth..example.com", false},
		{"ShouldNotAllowLeadingPeriod", ".example.com", false},
		{"ShouldNotAllowTrailingPeriod", "example.com.", false},
		{"ShouldNotAllowLeadingHyphen", "-auth.example.com", false},
		{"ShouldNotAllowTrailingHyphen", "auth-.example.com", false},
		{"ShouldNotAllowLabelTooLong", strings.Repeat("a", 64) + ".example.com", false},
		{"ShouldNotAllowTooLong", strings.Repeat("a.", 126) + "ab", false},
		{"ShouldNotAllowSpace", "auth example.com", false},
		{"ShouldNotAllowWildcard", "*.example.com", false},
		{"ShouldNotAllowPercentEncoding", "auth%2eexample.com", false},
		{"ShouldNotAllowNonASCII", "authé.example.com", false},
		{"ShouldNotAllowPort", "example.com:443", false},
		{"ShouldNotAllowBracketedIPv6", "[2001:db8::1]", false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, IsHostnameValid(tc.Hostname))
		})
	}
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
