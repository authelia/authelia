package utils

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNormalizedRequestURI(t *testing.T) {
	testCases := []struct {
		name      string
		have      string
		expected  string
		path      string
		rawPath   string
		pathQuery string
		err       string
	}{
		{"ShouldPreserveAlreadyNormalizedURI", "https://example.com/path?query=value", "https://example.com/path?query=value", "/path", "", "/path?query=value", ""},
		{"ShouldNormalizeRFC3986Section622Example", "eXAMPLE://a/./b/../b/%63/%7bfoo%7d", "example://a/b/c/%7Bfoo%7D", "/b/c/{foo}", "", "/b/c/%7Bfoo%7D", ""},
		{"ShouldBeIdempotentForRFC3986Section622NormalForm", "example://a/b/c/%7Bfoo%7D", "example://a/b/c/%7Bfoo%7D", "/b/c/{foo}", "", "/b/c/%7Bfoo%7D", ""},

		{"ShouldLowercaseScheme", "HTTP://example.com/", "http://example.com/", "/", "", "/", ""},
		{"ShouldLowercaseSchemeWithUpperCaseHost", "HTTP://EXAMPLE.COM/", "http://EXAMPLE.COM/", "/", "", "/", ""},
		{"ShouldUppercasePercentEncodingHexDigitsInPath", "http://example.com/%7bfoo%7d", "http://example.com/%7Bfoo%7D", "/{foo}", "", "/%7Bfoo%7D", ""},
		{"ShouldUppercasePercentEncodingHexDigitsInQuery", "http://example.com/?a=%2f", "http://example.com/?a=%2F", "/", "", "/?a=%2F", ""},

		{"ShouldDecodeUnreservedTilde", "http://example.com/%7euser", "http://example.com/~user", "/~user", "", "/~user", ""},
		{"ShouldDecodeUnreservedAlpha", "http://example.com/%41%42", "http://example.com/AB", "/AB", "", "/AB", ""},
		{"ShouldDecodeUnreservedDigit", "http://example.com/%31%32", "http://example.com/12", "/12", "", "/12", ""},
		{"ShouldDecodeUnreservedHyphenPeriodUnderscoreTilde", "http://example.com/%2D%5F%7E", "http://example.com/-_~", "/-_~", "", "/-_~", ""},
		{"ShouldRetainReservedEncodedSlashInPath", "http://example.com/foo%2Fbar", "http://example.com/foo%2Fbar", "/foo/bar", "/foo%2Fbar", "/foo%2Fbar", ""},
		{"ShouldRetainReservedEncodedQuestionMarkInPath", "http://example.com/foo%3Fbar", "http://example.com/foo%3Fbar", "/foo?bar", "", "/foo%3Fbar", ""},
		{"ShouldDecodeUnreservedInQuery", "http://example.com/foo?a=%7e", "http://example.com/foo?a=~", "/foo", "", "/foo?a=~", ""},
		{"ShouldRetainReservedEncodedInQuery", "http://example.com/foo?b=%2f", "http://example.com/foo?b=%2F", "/foo", "", "/foo?b=%2F", ""},
		{"ShouldNormalizeQueryUnreservedAndReserved", "http://example.com/foo?a=%7e&b=%2F", "http://example.com/foo?a=~&b=%2F", "/foo", "", "/foo?a=~&b=%2F", ""},
		{"ShouldRetainHostCaseAndNormalizeQuery", "https://EXAMPLE.com/p?A=%7e&B=%2f", "https://EXAMPLE.com/p?A=~&B=%2F", "/p", "", "/p?A=~&B=%2F", ""},

		{"ShouldRemoveSingleDotSegment", "http://example.com/a/./b", "http://example.com/a/b", "/a/b", "", "/a/b", ""},
		{"ShouldRemoveDoubleDotSegment", "http://example.com/a/b/../c", "http://example.com/a/c", "/a/c", "", "/a/c", ""},
		{"ShouldRemoveMultipleDotSegments", "http://example.com/a/b/c/./../../g", "http://example.com/a/g", "/a/g", "", "/a/g", ""},
		{"ShouldRemoveLeadingDoubleDotSegment", "http://example.com/../foo", "http://example.com/foo", "/foo", "", "/foo", ""},
		{"ShouldResolveTrailingDotToSlash", "http://example.com/foo/.", "http://example.com/foo/", "/foo/", "", "/foo/", ""},
		{"ShouldResolveTrailingDoubleDotToRoot", "http://example.com/foo/..", "http://example.com/", "/", "", "/", ""},
		{"ShouldResolveDotToRoot", "http://example.com/.", "http://example.com/", "/", "", "/", ""},
		{"ShouldResolveDoubleDotToRoot", "http://example.com/..", "http://example.com/", "/", "", "/", ""},
		{"ShouldMergeEmptyPathSegments", "http://example.com//a///b", "http://example.com/a/b", "/a/b", "", "/a/b", ""},
		{"ShouldMergeSlashesSurroundingDotSegment", "http://example.com/a//../b", "http://example.com/b", "/b", "", "/b", ""},
		{"ShouldMergeLeadingAndTrailingSlashes", "http://example.com///a///", "http://example.com/a/", "/a/", "", "/a/", ""},
		{"ShouldMergeSlashesWithoutAuthority", "//x//y", "/x/y", "/x/y", "", "/x/y", ""},
		{"ShouldNotMergeEncodedSlashes", "http://example.com/foo%2F%2Fbar", "http://example.com/foo%2F%2Fbar", "/foo//bar", "/foo%2F%2Fbar", "/foo%2F%2Fbar", ""},
		{"ShouldMergeUnencodedSlashesButRetainEncodedSlash", "http://example.com/a//%2f//b", "http://example.com/a/%2F/b", "/a///b", "/a/%2F/b", "/a/%2F/b", ""},
		{"ShouldNotTreatEncodedSlashAsSegmentSeparator", "http://example.com/a%2fb/../c", "http://example.com/c", "/c", "", "/c", ""},
		{"ShouldDecodeEncodedDotSegmentsThenRemove", "http://example.com/%2e%2e/foo", "http://example.com/foo", "/foo", "", "/foo", ""},

		{"ShouldRetainMultipleReservedEncodedSlashesInPath", "http://example.com/foo%2Fbar%2Fbaz", "http://example.com/foo%2Fbar%2Fbaz", "/foo/bar/baz", "/foo%2Fbar%2Fbaz", "/foo%2Fbar%2Fbaz", ""},
		{"ShouldUppercaseEncodedSlashHexDigits", "http://example.com/a%2fb", "http://example.com/a%2Fb", "/a/b", "/a%2Fb", "/a%2Fb", ""},
		{"ShouldNotResolveDotSegmentInsideEncodedSlashes", "http://example.com/a%2F..%2Fb", "http://example.com/a%2F..%2Fb", "/a/../b", "/a%2F..%2Fb", "/a%2F..%2Fb", ""},
		{"ShouldResolveLiteralDotSegmentButRetainEncodedSlash", "http://example.com/a%2Fb/c/../d", "http://example.com/a%2Fb/d", "/a/b/d", "/a%2Fb/d", "/a%2Fb/d", ""},

		{"ShouldAppendSlashToEmptyPathWithAuthority", "http://example.com", "http://example.com/", "/", "", "/", ""},
		{"ShouldRemoveDefaultHTTPPort", "http://example.com:80/", "http://example.com/", "/", "", "/", ""},
		{"ShouldRemoveDefaultHTTPSPort", "https://example.com:443/", "https://example.com/", "/", "", "/", ""},
		{"ShouldRemoveDefaultWSPort", "ws://example.com:80/", "ws://example.com/", "/", "", "/", ""},
		{"ShouldRemoveDefaultWSSPort", "wss://example.com:443/", "wss://example.com/", "/", "", "/", ""},
		{"ShouldRemoveEmptyPort", "http://example.com:/", "http://example.com/", "/", "", "/", ""},
		{"ShouldRetainNonDefaultPort", "http://example.com:8080/", "http://example.com:8080/", "/", "", "/", ""},
		{"ShouldRetainDefaultPortOfDifferentScheme", "https://example.com:80/", "https://example.com:80/", "/", "", "/", ""},
		{"ShouldNotRemovePortForUnknownScheme", "example://a:80/b", "example://a:80/b", "/b", "", "/b", ""},

		{"ShouldNormalizeAbsolutePathWithoutAuthority", "/a/./b/../c", "/a/c", "/a/c", "", "/a/c", ""},
		{"ShouldNormalizeAllComponentsCombined", "HTTPS://example.com:443/Path/%7bx%7d?Q=%7e", "https://example.com/Path/%7Bx%7D?Q=~", "/Path/{x}", "", "/Path/%7Bx%7D?Q=~", ""},

		{"ShouldErrorOnOpaqueURI", "mailto:foo@example.com", "", "", "", "", "cannot normalize opaque URI \"mailto:foo@example.com\""},
		{"ShouldErrorOnRelativeReference", "foo/bar", "", "", "", "", "parse \"foo/bar\": invalid URI for request"},
		{"ShouldErrorOnEmptyInput", "", "", "", "", "", "parse \"\": empty url"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uri, err := ParseNormalizedRequestURI(tc.have)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, uri)
			} else {
				require.NoError(t, err)
				require.NotNil(t, uri)
				assert.Equal(t, tc.expected, uri.String())
				assert.Equal(t, tc.path, uri.Path)
				assert.Equal(t, tc.rawPath, uri.RawPath)
				assert.Equal(t, tc.pathQuery, PathWithQueryFromURI(uri))
			}
		})
	}
}

func TestPathWithQueryFromURI(t *testing.T) {
	testCases := []struct {
		name     string
		have     *url.URL
		expected string
	}{
		{"ShouldReturnRootPath", &url.URL{Path: "/"}, "/"},
		{"ShouldReturnPathWithoutQuery", &url.URL{Path: "/test"}, "/test"},
		{"ShouldReturnPathWithTrailingSlash", &url.URL{Path: "/test/"}, "/test/"},
		{"ShouldReturnPathWithQuery", &url.URL{Path: "/test", RawQuery: "query=1&alt=2"}, "/test?query=1&alt=2"},
		{"ShouldReturnPathWithTrailingSlashAndQuery", &url.URL{Path: "/test/", RawQuery: "query=1&alt=2"}, "/test/?query=1&alt=2"},
		{"ShouldReturnEmptyPathWithQuery", &url.URL{Path: "", RawQuery: "query=1"}, "?query=1"},
		{"ShouldReturnEmptyPathWithoutQuery", &url.URL{Path: ""}, ""},
		{"ShouldRetainReservedEncodedSlashInPath", &url.URL{Path: "/foo/bar", RawPath: "/foo%2Fbar"}, "/foo%2Fbar"},
		{"ShouldRetainReservedEncodedSlashInPathWithQuery", &url.URL{Path: "/foo/bar", RawPath: "/foo%2Fbar", RawQuery: "a=b"}, "/foo%2Fbar?a=b"},
		{"ShouldEscapeReservedQuestionMarkInPath", &url.URL{Path: "/foo?bar"}, "/foo%3Fbar"},
		{"ShouldEscapeReservedNumberSignInPath", &url.URL{Path: "/foo#bar"}, "/foo%23bar"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, PathWithQueryFromURI(tc.have))
		})
	}
}

func TestNormalizeRequestURIMergeSlashes(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldReturnEmpty", "", ""},
		{"ShouldReturnSingleSlash", "/", "/"},
		{"ShouldNotChangeWithoutRuns", "/a/b/c", "/a/b/c"},
		{"ShouldMergeDoubleSlash", "/a//b", "/a/b"},
		{"ShouldMergeLongRun", "/a/////b", "/a/b"},
		{"ShouldMergeLeadingSlashes", "//a/b", "/a/b"},
		{"ShouldMergeTrailingSlashes", "/a/b///", "/a/b/"},
		{"ShouldMergeOnlySlashes", "////", "/"},
		{"ShouldNotMergeEncodedSlash", "/a%2F%2Fb", "/a%2F%2Fb"},
		{"ShouldMergeUnencodedButRetainEncodedSlash", "/a//%2F//b", "/a/%2F/b"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, normalizeRequestURIMergeSlashes(tc.have))
		})
	}
}

func TestURLPathWithQueryClean(t *testing.T) {
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
		{"ShouldReturnCleanedPathEscaped", "https://example.com/five/..%2ftest?query=1&alt=2", "/five/..%2ftest?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscapedExtra", "https://example.com/five/..%2ftest?query=1&alt=2", "/five/..%2ftest?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscapedExtraSurrounding", "https://example.com/five/%2f..%2f/test?query=1&alt=2", "/five/%2f..%2f/test?query=1&alt=2"},
		{"ShouldReturnCleanedPathEscapedPeriods", "https://example.com/five/%2f%2e%2e%2f/test?query=1&alt=2", "/five/%2f%2e%2e%2f/test?query=1&alt=2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.ParseRequestURI(tc.have)
			require.NoError(t, err)

			actual := URLPathWithQueryClean(u)

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
