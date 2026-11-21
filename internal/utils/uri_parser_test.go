// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURIParserCanonical(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		opts     []URIParserOption
		expected string
	}{
		{"ShouldLowercaseSchemeHostAndRemoveDefaultPortHTTPS", "HTTPS://EXAMPLE.com:443", nil, "https://example.com/"},
		{"ShouldRetainNonDefaultPort", "https://example.com:8443/a", nil, "https://example.com:8443/a"},
		{"ShouldRemoveDefaultPortHTTP", "http://example.com:80/", nil, "http://example.com/"},
		{"ShouldRemoveDefaultPortIPv6", "https://[::1]:443", nil, "https://[::1]/"},
		{"ShouldRetainHTTPPortOnHTTPS", "https://example.com:80/", nil, "https://example.com:80/"},
		{"ShouldRemoveDotSegments", "/a/../b/./c", nil, "/b/c"},
		{"ShouldRetainTrailingSlashAfterDotDot", "/a/b/..", nil, "/a/"},
		{"ShouldRetainTrailingSlashAfterDot", "/a/b/.", nil, "/a/b/"},
		{"ShouldRemoveDotSegmentsAboveRoot", "/../../a", nil, "/a"},
		{"ShouldRetainEmptySegments", "/a//b/", nil, "/a//b/"},
		{"ShouldRemoveOnlyEmptySegmentWithDotDot", "/a/b//../c", []URIParserOption{WithURLAllowAmbiguity(URLAmbiguityEmptyDotSegment)}, "/a/b/c"},
		{"ShouldRemoveOneOfRunOfEmptySegmentsWithDotDot", "/a/b///../c", []URIParserOption{WithURLAllowAmbiguity(URLAmbiguityEmptyDotSegment)}, "/a/b//c"},
		{"ShouldRemoveOnlyEmptySegmentWithTrailingDotDot", "/a//..", []URIParserOption{WithURLAllowAmbiguity(URLAmbiguityEmptyDotSegment)}, "/a/"},
		{"ShouldRemoveEmptySegmentsWithDotDotAboveRoot", "//../a", nil, "/a"},
		{"ShouldRetainEmptySegmentsAfterDotDot", "/a/..//b", nil, "//b"},
		{"ShouldRetainEncodedSeparator", "/a%2fb", nil, "/a%2Fb"},
		{"ShouldRetainEncodedSeparatorRuns", "/a%2f%2fb", nil, "/a%2F%2Fb"},
		{"ShouldDecodeUnreserved", "/%7euser/%41", nil, "/~user/A"},
		{"ShouldRetainEncodedSubDelim", "/a%3bb", nil, "/a%3Bb"},
		{"ShouldUppercaseHex", "/caf%c3%a9", nil, "/caf%C3%A9"},
		{"ShouldNormalizeQuery", "/a?b=%7e&c=%2f", nil, "/a?b=~&c=%2F"},
		{"ShouldDropEmptyQuery", "/a?", nil, "/a"},
		{"ShouldReEncodeIPv6Zone", "https://[fe80::1%25en0]/a", nil, "https://[fe80::1%25en0]/a"},
		{"ShouldRetainIPv6ZoneCase", "https://[FE80::1%25EN0]/a", nil, "https://[fe80::1%25EN0]/a"},
		{"ShouldDropEmptyHostWithPort", "https://:443/a", nil, "/a"},
		{"ShouldTreatEmptyPathAsSlash", "https://example.com", nil, "https://example.com/"},
		{"ShouldDropUserInfo", "https://u:p@example.com/", []URIParserOption{WithURLAllowAmbiguity(URLAmbiguityUserInfo)}, "https://example.com/"},
		{"ShouldRemoveEncodedDotSegments", "/a/%2e%2e/b", []URIParserOption{WithURLAllowAmbiguity(URLAmbiguityEncodedDotSegment)}, "/b"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseRequestURI(tc.have, tc.opts...)
			require.NoError(t, err)
			require.NotNil(t, parsed)

			assert.Equal(t, tc.expected, parsed.Canonical)

			t.Run("ShouldBeIdempotent", func(t *testing.T) {
				again, err := ParseRequestURI(parsed.Canonical, tc.opts...)
				require.NoError(t, err)

				assert.Equal(t, parsed.Canonical, again.Canonical)
			})
		})
	}
}

func TestURIParserResult(t *testing.T) {
	parsed, err := ParseRequestURI("https://example.com/five/../a%2Fb?query=1&alt=2")
	require.NoError(t, err)

	require.NotNil(t, parsed.URL)

	assert.Equal(t, "example.com", parsed.URL.Host)
	assert.Equal(t, "/five/../a%2Fb", parsed.URL.RawPath)
	assert.Equal(t, "/a/b?query=1&alt=2", parsed.AccessPath)
	assert.Equal(t, "https://example.com/a%2Fb?query=1&alt=2", parsed.Canonical)
}

func TestURIParserAmbiguity(t *testing.T) {
	testCases := []struct {
		name     string
		have     []string
		expected []URLAmbiguity
	}{
		{"ShouldDetectUserInfo", []string{"https://user@example.com/", "https://@example.com/"}, []URLAmbiguity{URLAmbiguityUserInfo}},
		{"ShouldDetectEncodedDotSegment", []string{"/foo/%2e%2e/bar", "/%2e./x"}, []URLAmbiguity{URLAmbiguityEncodedDotSegment}},
		{"ShouldDetectEncodedDotSegmentSeparator", []string{"/foo/..%2fbar", "/foo%2f..%2fbar", "/%2e.%2f"}, []URLAmbiguity{URLAmbiguityEncodedDotSegment}},
		{"ShouldDetectParameterDotSegment", []string{"/public/..;/admin", "/.;/a", "/;/a", "/%2e%2e;x/"}, []URLAmbiguity{URLAmbiguityParameterDotSegment}},
		{"ShouldDetectDoubleEncodedDotSegment", []string{"/a/%252e%252e%252fb", "/%252f..%252f"}, []URLAmbiguity{URLAmbiguityDoubleEncodedDotSegment, URLAmbiguityEncodedPercent}},
		{"ShouldDetectBackslash", []string{`/a\b`, "/a%5Cb"}, []URLAmbiguity{URLAmbiguityBackslash}},
		{"ShouldDetectControlCharacter", []string{"/a%00b", "/a?b=%0d%0a"}, []URLAmbiguity{URLAmbiguityControlCharacter}},
		{"ShouldDetectInvalidUTF8", []string{"/a/%c0%ae", "/a%ff"}, []URLAmbiguity{URLAmbiguityInvalidUTF8}},
		{"ShouldDetectFragment", []string{"/a#b"}, []URLAmbiguity{URLAmbiguityFragment}},
		{"ShouldDetectEncodedSeparator", []string{"/a%2Fb"}, []URLAmbiguity{URLAmbiguityEncodedSeparator, URLAmbiguityEncodedNonUnreserved}},
		{"ShouldDetectEncodedPercent", []string{"/a%25b"}, []URLAmbiguity{URLAmbiguityEncodedPercent, URLAmbiguityEncodedNonUnreserved}},
		{"ShouldDetectDotSegment", []string{"/a/../b"}, []URLAmbiguity{URLAmbiguityDotSegment}},
		{"ShouldDetectEmptySegment", []string{"/a//b"}, []URLAmbiguity{URLAmbiguityEmptySegment}},
		{"ShouldDetectPathParameter", []string{"/foo;x=1"}, []URLAmbiguity{URLAmbiguityPathParameter}},
		{"ShouldDetectInvalidQueryEncoding", []string{"/a?b=%zz"}, []URLAmbiguity{URLAmbiguityInvalidQueryEncoding}},
		{"ShouldDetectEncodedNonUnreserved", []string{"/a%20b", "/caf%C3%A9"}, []URLAmbiguity{URLAmbiguityEncodedNonUnreserved}},
		{"ShouldDetectEncodedUnreserved", []string{"/%61dmin", "/file%2ejson"}, []URLAmbiguity{URLAmbiguityEncodedUnreserved}},
		{"ShouldDetectEmptyDotSegment", []string{"/public//../admin", "/a//./../b", "/a//..", "/a/b///../c", "/a//%2e%2e/b", "/five/%2f..%2f/test"}, []URLAmbiguity{URLAmbiguityEmptyDotSegment}},
		{"ShouldNotDetectQueryEncoding", []string{"/a?b=%2F%20", "/a/b/?c=d", "/", "/a/.../..b/b.."}, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, have := range tc.have {
				t.Run(have, func(t *testing.T) {
					u, err := url.ParseRequestURI(have)
					require.NoError(t, err)

					actual := urlAmbiguity(have, u)

					if len(tc.expected) == 0 {
						assert.Equal(t, URLAmbiguity(0), actual)
					}

					for _, expected := range tc.expected {
						assert.True(t, actual.Has(expected), "expected %q to include: %s, actual: %s", have, expected, actual)

						parsed, err := ParseRequestURI(have, WithURLAllowAmbiguity(URLAmbiguityAll), WithURLRejectAmbiguity(expected))

						assert.Nil(t, parsed)
						assert.ErrorIs(t, err, ErrURLAmbiguous)

						parsed, err = ParseRequestURI(have, WithURLAllowAmbiguity(URLAmbiguityAll))

						assert.NoError(t, err)
						assert.NotNil(t, parsed)
					}
				})
			}
		})
	}
}

func TestURIParserDefaults(t *testing.T) {
	for _, have := range []string{"/group%2Fproject", "/a%2f%2fb", "/a%25b", "/file%2ejson", "/a/../b", "/a//b"} {
		t.Run("ShouldAllow"+have, func(t *testing.T) {
			parsed, err := DefaultURIParser().Parse(have)

			assert.NoError(t, err)
			assert.NotNil(t, parsed)
		})
	}

	for _, have := range []string{"/foo/%2e%2e/bar", "/foo/..%2fbar", "/foo%2f..%2fbar", "/public/..;/admin", "/a/%252e%252e%252fb", "/a%5Cb", "https://u@example.com/", "/foo;jsessionid=1", "/admin;x/secret", "/public//../admin", "/a//./../b", "/a//..", "/a/b///../c"} {
		t.Run("ShouldReject"+have, func(t *testing.T) {
			parsed, err := DefaultURIParser().Parse(have)

			assert.Nil(t, parsed)
			assert.ErrorIs(t, err, ErrURLAmbiguous)
		})
	}
}

func TestURIParserOptions(t *testing.T) {
	const have = "/a%2Fb"

	testCases := []struct {
		name     string
		opts     []URIParserOption
		rejected bool
	}{
		{"ShouldAllowByDefault", nil, false},
		{"ShouldRejectWhenAdded", []URIParserOption{WithURLRejectAmbiguity(URLAmbiguityEncodedSeparator)}, true},
		{"ShouldAllowWhenRejectedThenAllowed", []URIParserOption{WithURLRejectAmbiguity(URLAmbiguityEncodedSeparator), WithURLAllowAmbiguity(URLAmbiguityEncodedSeparator)}, false},
		{"ShouldRejectWhenAllowedThenRejected", []URIParserOption{WithURLAllowAmbiguity(URLAmbiguityEncodedSeparator), WithURLRejectAmbiguity(URLAmbiguityEncodedSeparator)}, true},
		{"ShouldRejectWhenAllRejected", []URIParserOption{WithURLRejectAmbiguity(URLAmbiguityAll)}, true},
		{"ShouldAllowWhenAllAllowed", []URIParserOption{WithURLAllowAmbiguity(URLAmbiguityAll)}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := NewURIParser(tc.opts...).Parse(have)

			if tc.rejected {
				assert.Nil(t, parsed)
				assert.ErrorIs(t, err, ErrURLAmbiguous)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, parsed)
			}
		})
	}
}

func TestURLAmbiguityNames(t *testing.T) {
	var all URLAmbiguity

	names := URLAmbiguityNames()

	assert.Len(t, names, 17)

	for _, name := range names {
		ambiguity, ok := URLAmbiguityFromName(name)

		assert.True(t, ok, name)
		assert.NotZero(t, ambiguity, name)
		assert.Zero(t, all&ambiguity, name)

		all |= ambiguity
	}

	assert.Equal(t, URLAmbiguityAll, all)

	ambiguity, ok := URLAmbiguityFromName("unknown")

	assert.False(t, ok)
	assert.Zero(t, ambiguity)

	assert.Equal(t, "url contains userinfo, path contains a backslash", (URLAmbiguityUserInfo | URLAmbiguityBackslash).String())
}

func TestURIParserErrors(t *testing.T) {
	testCases := []struct {
		name string
		have string
		err  string
		is   error
	}{
		{"ShouldErrorEmpty", "", "parse \"\": empty url", nil},
		{"ShouldErrorRelativePath", "a/b", "parse \"a/b\": invalid URI for request", nil},
		{"ShouldErrorInvalidPathEncoding", "/a%zz", "parse \"/a%zz\": invalid URL escape \"%zz\"", nil},
		{"ShouldErrorOpaque", "https:example.com", "url does not have an absolute path: 'https:example.com'", ErrURLNotAbsolutePath},
		{"ShouldErrorAsterisk", "*", "url does not have an absolute path: '*'", ErrURLNotAbsolutePath},
		{"ShouldErrorAmbiguous", "/a%5Cb", "url is ambiguous: path contains a backslash", ErrURLAmbiguous},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseRequestURI(tc.have)

			assert.EqualError(t, err, tc.err)
			assert.Nil(t, parsed)

			if tc.is != nil {
				assert.ErrorIs(t, err, tc.is)
			}
		})
	}
}

func TestURIParserErrorRedactsPassword(t *testing.T) {
	testCases := []struct {
		name string
		have string
		opts []URIParserOption
	}{
		{"ShouldRedactParseError", "https://u:pw@example.com/%zz", nil},
		{"ShouldRedactRejection", "https://u:pw@example.com/", nil},
		{"ShouldRedactOpaque", "https:u:pw@example.com", nil},
		{"ShouldRedactFragmentInUserInfo", "https://u:pw#x@example.com/", nil},
		{"ShouldRedactSlashInPassword", "https://u:pw/x@example.com/", nil},
		{"ShouldRedactOpaqueWithLaterAuthorityMarker", "https:u:pw@example.com?x=http://a", nil},
		{"ShouldRedactOpaqueWithLaterAuthorityMarkerPath", "https:u:pw@example.com/x://y", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseRequestURI(tc.have, tc.opts...)
			require.Error(t, err)

			assert.NotContains(t, err.Error(), "pw")
		})
	}

	_, err := ParseRequestURI("https://u:pw@example.com/%zz")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "https://example.com/%zz")

	_, err = ParseRequestURI("https://u:pw@example.com/")
	require.ErrorIs(t, err, ErrURLAmbiguous)
}

func TestURIParserConcurrent(t *testing.T) {
	p := NewURIParser()

	wg := &sync.WaitGroup{}

	for i := 0; i < 16; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				parsed, err := p.Parse("https://EXAMPLE.com:443/a/../b%2f?c=%7e")

				assert.NoError(t, err)
				assert.Equal(t, "https://example.com/b%2F?c=~", parsed.Canonical)

				_, err = p.Parse("/a/%2e%2e/b")

				assert.ErrorIs(t, err, ErrURLAmbiguous)
			}
		}()
	}

	wg.Wait()
}

func TestURICanonicalRetainsPortWithoutDefaultForScheme(t *testing.T) {
	assert.Equal(t, "ftp://example.com:443/a", URICanonical(&url.URL{Scheme: "FTP", Host: "example.com:443", Path: "/a"}))
}

func TestURLSchemePrefix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldReturnSchemeWithAuthorityMarker", "https://example.com", "https://"},
		{"ShouldReturnSchemeWithoutAuthorityMarker", "https:example.com", "https:"},
		{"ShouldAcceptDigitPlusHyphenPeriod", "a1+b-c.d://example.com", "a1+b-c.d://"},
		{"ShouldRejectLeadingDigit", "1a://example.com", ""},
		{"ShouldRejectLeadingColon", "://example.com", ""},
		{"ShouldRejectInvalidCharacter", "a_b://example.com", ""},
		{"ShouldReturnEmptyWithoutColon", "example", ""},
		{"ShouldReturnEmptyForEmpty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, urlSchemePrefix(tc.have))
		})
	}
}

func TestURLRemoveUserInfo(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldRemoveUserInfo", "https://u:pw@example.com/", "https://example.com/"},
		{"ShouldRemoveUpToLastAt", "https://u:p@w@example.com/", "https://example.com/"},
		{"ShouldRetainComplexScheme", "a1+b-c.d://u:pw@example.com/", "a1+b-c.d://example.com/"},
		{"ShouldRemoveEverythingWithoutValidScheme", "1a://u:pw@example.com/", "example.com/"},
		{"ShouldNotChangeWithoutAt", "https://example.com/", "https://example.com/"},
		{"ShouldNotChangeAbsolutePath", "/a@b", "/a@b"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, URIRemoveUserInfo(tc.have))
		})
	}
}

func TestURLAmbiguityQuery(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected URLAmbiguity
	}{
		{"ShouldNotDetectPlain", "a=b", 0},
		{"ShouldDetectLiteralControl", "a=\x01", URLAmbiguityControlCharacter},
		{"ShouldDetectLiteralDelete", "a=\x7f", URLAmbiguityControlCharacter},
		{"ShouldDetectEncodedControl", "a=%0a", URLAmbiguityControlCharacter},
		{"ShouldDetectEncodedDelete", "a=%7F", URLAmbiguityControlCharacter},
		{"ShouldDetectMalformed", "a=%zz", URLAmbiguityInvalidQueryEncoding},
		{"ShouldDetectTruncated", "a=%2", URLAmbiguityInvalidQueryEncoding},
		{"ShouldNotDetectEncodedPrintable", "a=%2f", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, urlAmbiguityQuery(tc.have))
		})
	}
}

func TestURLAmbiguityFromNameScopePrefix(t *testing.T) {
	testCases := []struct {
		name     string
		expected URLAmbiguity
		have     string
	}{
		{"user_info", URLAmbiguityUserInfo, "https://user@example.com/"},
		{"control_character", URLAmbiguityControlCharacter, "/a?b=%0a"},
		{"fragment", URLAmbiguityFragment, "/a?b#c"},
		{"path_encoded_dot_segment", URLAmbiguityEncodedDotSegment, "/a/%2e%2e/b"},
		{"path_parameter_dot_segment", URLAmbiguityParameterDotSegment, "/a/..;/b"},
		{"path_double_encoded_dot_segment", URLAmbiguityDoubleEncodedDotSegment, "/a/%252e%252e/b"},
		{"path_backslash", URLAmbiguityBackslash, "/a%5Cb"},
		{"path_invalid_utf8", URLAmbiguityInvalidUTF8, "/%C0%AE"},
		{"path_encoded_separator", URLAmbiguityEncodedSeparator, "/a%2Fb"},
		{"path_encoded_percent", URLAmbiguityEncodedPercent, "/a%25b"},
		{"path_dot_segment", URLAmbiguityDotSegment, "/a/../b"},
		{"path_empty_segment", URLAmbiguityEmptySegment, "/a//b"},
		{"path_parameter", URLAmbiguityPathParameter, "/a;x"},
		{"path_encoded_non_unreserved", URLAmbiguityEncodedNonUnreserved, "/a%20b"},
		{"path_encoded_unreserved", URLAmbiguityEncodedUnreserved, "/%61"},
		{"query_invalid_encoding", URLAmbiguityInvalidQueryEncoding, "/a?b=%zz"},
		{"path_empty_dot_segment", URLAmbiguityEmptyDotSegment, "/a//../b"},
	}

	assert.Len(t, testCases, len(URLAmbiguityNames()))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ambiguity, ok := URLAmbiguityFromName(tc.name)

			require.True(t, ok)
			assert.Equal(t, tc.expected, ambiguity)

			_, err := ParseRequestURI(tc.have, WithURLAllowAmbiguity(URLAmbiguityAll), WithURLRejectAmbiguity(ambiguity))

			assert.ErrorIs(t, err, ErrURLAmbiguous)
		})
	}

	for _, name := range []string{
		"encoded_dot_segment", "parameter_dot_segment", "double_encoded_dot_segment", "backslash", "invalid_utf8",
		"encoded_separator", "encoded_percent", "dot_segment", "empty_segment", "invalid_query_encoding",
		"encoded_non_unreserved", "encoded_unreserved",
	} {
		_, ok := URLAmbiguityFromName(name)

		assert.False(t, ok, name)
	}
}

func TestURLAmbiguityScopeDoesNotCrossComponents(t *testing.T) {
	testCases := []struct {
		name      string
		ambiguity URLAmbiguity
		have      string
	}{
		{"ShouldNotApplyBackslashToQuery", URLAmbiguityBackslash, "/a?b=%5C"},
		{"ShouldNotApplyEncodedSeparatorToQuery", URLAmbiguityEncodedSeparator, "/a?b=%2F"},
		{"ShouldNotApplyEncodedPercentToQuery", URLAmbiguityEncodedPercent, "/a?b=%25"},
		{"ShouldNotApplyEncodedNonUnreservedToQuery", URLAmbiguityEncodedNonUnreserved, "/a?b=%20"},
		{"ShouldNotApplyEncodedUnreservedToQuery", URLAmbiguityEncodedUnreserved, "/a?b=%61"},
		{"ShouldNotApplyInvalidUTF8ToQuery", URLAmbiguityInvalidUTF8, "/a?b=%C0%AE"},
		{"ShouldNotApplyDotSegmentToQuery", URLAmbiguityDotSegment, "/a?b=/../c"},
		{"ShouldNotApplyEmptySegmentToQuery", URLAmbiguityEmptySegment, "/a?b=//c"},
		{"ShouldNotApplyPathParameterToQuery", URLAmbiguityPathParameter, "/a?b=;c"},
		{"ShouldNotApplyInvalidQueryEncodingToPath", URLAmbiguityInvalidQueryEncoding, "/a%2Fb"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseRequestURI(tc.have, WithURLAllowAmbiguity(URLAmbiguityAll), WithURLRejectAmbiguity(tc.ambiguity))

			assert.NoError(t, err)
		})
	}
}

func TestNewURIParserFromNames(t *testing.T) {
	testCases := []struct {
		name     string
		allow    []string
		reject   []string
		have     string
		rejected bool
	}{
		{"ShouldUseDefaults", nil, nil, "/a;x", true},
		{"ShouldAllowNamedClass", []string{"path_parameter"}, nil, "/a;x", false},
		{"ShouldRejectNamedClass", nil, []string{"path_encoded_percent"}, "/a%25b", true},
		{"ShouldIgnoreUnknownNames", []string{"nope"}, []string{"nope"}, "/a%25b", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewURIParserFromNames(tc.allow, tc.reject).Parse(tc.have)

			if tc.rejected {
				assert.ErrorIs(t, err, ErrURLAmbiguous)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestURIParserEmptyDotSegmentNotDetected(t *testing.T) {
	for _, have := range []string{"/a//b", "//admin", "/a//./b", "//../a", "/a/..//b", "/a//b/../c", "/a/b?c=//../d"} {
		t.Run(have, func(t *testing.T) {
			u, err := url.ParseRequestURI(have)
			require.NoError(t, err)

			assert.False(t, urlAmbiguity(have, u).Has(URLAmbiguityEmptyDotSegment))
		})
	}
}
