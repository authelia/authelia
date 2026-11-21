// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authorization

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
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
			"ShouldAllowHyphenatedMethod",
			"version-control",
			"https://app.example.com/",
			"VERSION-CONTROL",
			"app.example.com",
			"/",
			"",
		},
		{
			"ShouldAllowMaxLengthMethod",
			"ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEF",
			"https://app.example.com/",
			"ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEF",
			"app.example.com",
			"/",
			"",
		},
		{
			"ShouldRejectMethodOverMaxLength",
			"ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFG",
			"https://app.example.com/",
			"",
			"",
			"",
			"method header with length 33 exceeds the maximum length of 32",
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
			object, err := NewObjectMethodURL(utils.DefaultURIParser(), []byte(tc.method), []byte(tc.rawURL))

			switch tc.err {
			case "":
				require.NoError(t, err)
				require.NotNil(t, object)

				assert.Equal(t, tc.expectedMethod, object.Method)
				assert.Equal(t, tc.expectedDomain, object.Domain)
				assert.Equal(t, tc.expectedPath, object.Path)
			case errAmbiguous:
				assert.ErrorIs(t, err, utils.ErrURLAmbiguous)
				assert.Nil(t, object)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, object)
			}
		})
	}
}

const errAmbiguous = "ambiguous"

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
			"https://example.com", "", "/", "",
		},
		{
			"ShouldParseQueryWithoutPath",
			[]byte("https"), []byte("example.com"), []byte("?query=value"),
			"https://example.com?query=value", "", "/?query=value", "",
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
			"ShouldRejectEncodedDotSegments",
			[]byte("https"), []byte("example.com"), []byte("/%2e%2e/secret"),
			"", "", "", errAmbiguous,
		},
		{
			"ShouldRejectPathParameter",
			[]byte("https"), []byte("example.com"), []byte("/admin;x/secret"),
			"", "", "", errAmbiguous,
		},
		{
			"ShouldRejectEncodedDotSegmentsMidPath",
			[]byte("https"), []byte("example.com"), []byte("/foo/%2e%2e/bar"),
			"", "", "", errAmbiguous,
		},
		{
			"ShouldRejectEncodedSlashInTraversal",
			[]byte("https"), []byte("example.com"), []byte("/foo/..%2fbar"),
			"", "", "", errAmbiguous,
		},
		{
			"ShouldRejectDotSegmentBetweenEncodedSlashes",
			[]byte("https"), []byte("example.com"), []byte("/foo%2f..%2fbar"),
			"", "", "", errAmbiguous,
		},
		{
			"ShouldRejectDotSegmentBetweenUppercaseEncodedSlashes",
			[]byte("https"), []byte("example.com"), []byte("/foo%2F..%2Fbar"),
			"", "", "", errAmbiguous,
		},
		{
			"ShouldRejectEncodedDotSegmentBetweenEncodedSlashes",
			[]byte("https"), []byte("example.com"), []byte("/foo%2f%2e%2e%2fbar"),
			"", "", "", errAmbiguous,
		},
		{
			"ShouldRejectFullyEncodedTraversal",
			[]byte("https"), []byte("example.com"), []byte("/%2e%2e%2f%2e%2e%2fetc%2fpasswd"),
			"", "", "", errAmbiguous,
		},
		{
			"ShouldRejectEncodedSlashesAroundDotSegment",
			[]byte("https"), []byte("example.com"), []byte("/%2f..%2f"),
			"", "", "", errAmbiguous,
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
			"ShouldErrorOnMissingSchemeAndHost",
			nil, nil, []byte("/"),
			"", "", "", "missing scheme value",
		},
		{
			"ShouldErrorOnEmptySchemeAndHost",
			[]byte(""), []byte(""), []byte("/"),
			"", "", "", "missing scheme value",
		},
		{
			"ShouldErrorOnMissingSchemeWithHost",
			nil, []byte("example.com"), []byte("/"),
			"", "", "", "missing scheme value",
		},
		{
			"ShouldErrorOnEmptySchemeWithHost",
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
			object, err := NewObjectMethodSchemeHostPath(utils.DefaultURIParser(), []byte(fasthttp.MethodGet), tc.scheme, tc.host, tc.path)

			switch tc.err {
			case "":
				require.NoError(t, err)
				require.NotNil(t, object)

				assert.Equal(t, tc.expected, object.URL.String())
				assert.Equal(t, tc.expectedClean, object.Path)

				if tc.expectedPath != "" {
					assert.Equal(t, tc.expectedPath, object.URL.Path)
				}
			case errAmbiguous:
				assert.ErrorIs(t, err, utils.ErrURLAmbiguous)
				assert.Nil(t, object)
			default:
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, object)
			}
		})
	}
}

func TestNewObjectAccessPathUnchanged(t *testing.T) {
	testCases := []struct {
		name      string
		have      string
		path      string
		canonical string
	}{
		{"ShouldDecodeSpace", "https://e.com/a%20b", "/a b", "https://e.com/a%20b"},
		{"ShouldDecodeUTF8", "https://e.com/caf%C3%A9", "/café", "https://e.com/caf%C3%A9"},
		{"ShouldDecodeEncodedSlash", "https://e.com/a%2Fb", "/a/b", "https://e.com/a%2Fb"},
		{"ShouldKeepQueryEncoded", "https://e.com/a/?x=%2f", "/a/?x=%2f", "https://e.com/a/?x=%2F"},
		{"ShouldHandleEmptyPath", "https://e.com", "/", "https://e.com/"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			object, err := NewObjectMethodURL(utils.DefaultURIParser(), []byte(fasthttp.MethodGet), []byte(tc.have))

			require.NoError(t, err)
			require.NotNil(t, object)

			assert.Equal(t, tc.path, object.Path)
			assert.Equal(t, tc.canonical, object.CanonicalURI)
		})
	}
}

func TestNewObjectACLRegexStillMatches(t *testing.T) {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(bypass).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"e.com"},
			Policy:    deny,
			Resources: createSliceRegexRule(t, []string{`^/café`, `^/a b`}),
		}).
		Build()

	for _, have := range []string{"https://e.com/caf%C3%A9", "https://e.com/a%20b"} {
		t.Run(have, func(t *testing.T) {
			tester.CheckAuthorizations(t, Subject{}, have, fasthttp.MethodGet, Denied)
		})
	}

	tester.CheckAuthorizations(t, Subject{}, "https://e.com/other", fasthttp.MethodGet, Bypass)
}

func TestNewObjectMethodURLParserOptions(t *testing.T) {
	parser := utils.NewURIParser(utils.WithURLAllowAmbiguity(utils.URLAmbiguityEncodedDotSegment))

	object, err := NewObjectMethodURL(parser, []byte(fasthttp.MethodGet), []byte("https://e.com/foo/%2e%2e/bar"))

	require.NoError(t, err)
	require.NotNil(t, object)

	assert.Equal(t, "/bar", object.Path)

	object, err = NewObjectMethodURL(nil, []byte(fasthttp.MethodGet), []byte("https://e.com/foo/%2e%2e/bar"))

	assert.ErrorIs(t, err, utils.ErrURLAmbiguous)
	assert.Nil(t, object)
}

func TestObjectString(t *testing.T) {
	testCases := []struct {
		name     string
		have     Object
		expected string
	}{
		{"ShouldReturnEmptyWithoutURL", Object{}, ""},
		{"ShouldReturnURL", Object{URL: &url.URL{Scheme: "https", Host: "e.com", Path: "/a"}}, "https://e.com/a"},
		{"ShouldRedactPassword", Object{URL: &url.URL{Scheme: "https", User: url.UserPassword("u", "pw"), Host: "e.com", Path: "/a"}}, "https://u:xxxxx@e.com/a"},
		{"ShouldRetainUsername", Object{URL: &url.URL{Scheme: "https", User: url.User("u"), Host: "e.com", Path: "/a"}}, "https://u@e.com/a"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.String())
		})
	}
}

func TestNewObjectNormalization(t *testing.T) {
	testCases := []struct {
		name      string
		have      string
		domain    string
		path      string
		canonical string
	}{
		{"ShouldLowercaseSchemeAndHost", "HTTPS://E.COM/Path", "e.com", "/Path", "https://e.com/Path"},
		{"ShouldExcludePortFromDomain", "https://e.com:8443/a", "e.com", "/a", "https://e.com:8443/a"},
		{"ShouldRemoveDefaultPortFromCanonicalOnly", "https://e.com:443/a", "e.com", "/a", "https://e.com/a"},
		{"ShouldCleanDotSegments", "https://e.com/a/../b/./c", "e.com", "/b/c", "https://e.com/b/c"},
		{"ShouldMergeSlashesInPathOnly", "https://e.com/a//b", "e.com", "/a/b", "https://e.com/a//b"},
		{"ShouldRetainTrailingSlash", "https://e.com/a/", "e.com", "/a/", "https://e.com/a/"},
		{"ShouldRetainTrailingSlashWithQuery", "https://e.com/a/?x=1", "e.com", "/a/?x=1", "https://e.com/a/?x=1"},
		{"ShouldNotCleanQuery", "https://e.com/a?x=/../b", "e.com", "/a?x=/../b", "https://e.com/a?x=/../b"},
		{"ShouldDecodeUnreservedInPathAndCanonical", "https://e.com/%61", "e.com", "/a", "https://e.com/a"},
		{"ShouldHandleEmptyPathWithQuery", "https://e.com?x=1", "e.com", "/?x=1", "https://e.com/?x=1"},
		{"ShouldHandleRootWithQuery", "https://e.com/?x=1", "e.com", "/?x=1", "https://e.com/?x=1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			object, err := NewObjectMethodURL(nil, []byte(fasthttp.MethodGet), []byte(tc.have))

			require.NoError(t, err)

			assert.Equal(t, tc.domain, object.Domain)
			assert.Equal(t, tc.path, object.Path)
			assert.Equal(t, tc.canonical, object.CanonicalURI)
		})
	}
}

func TestNewObjectMethodURLOrSchemeHostPathParser(t *testing.T) {
	const (
		rawURL = "https://e.com/foo/%2e%2e/bar"
		scheme = "https"
		host   = "e.com"
		path   = "/foo/%2e%2e/bar"
	)

	allow := utils.NewURIParser(utils.WithURLAllowAmbiguity(utils.URLAmbiguityEncodedDotSegment))

	testCases := []struct {
		name     string
		parser   *utils.URIParser
		rawURL   string
		rejected bool
	}{
		{"ShouldRejectFromURLWithNilParser", nil, rawURL, true},
		{"ShouldRejectFromURLWithDefaultParser", utils.DefaultURIParser(), rawURL, true},
		{"ShouldAllowFromURLWithParser", allow, rawURL, false},
		{"ShouldRejectFromPartsWithNilParser", nil, "", true},
		{"ShouldRejectFromPartsWithDefaultParser", utils.DefaultURIParser(), "", true},
		{"ShouldAllowFromPartsWithParser", allow, "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			object, err := NewObjectMethodURLOrSchemeHostPath(tc.parser, []byte(fasthttp.MethodGet), []byte(tc.rawURL), []byte(scheme), []byte(host), []byte(path))

			if tc.rejected {
				assert.ErrorIs(t, err, utils.ErrURLAmbiguous)
				assert.Nil(t, object)

				return
			}

			require.NoError(t, err)

			assert.Equal(t, "/bar", object.Path)
			assert.Equal(t, "https://e.com/bar", object.CanonicalURI)
		})
	}
}

func TestNewObjectMethodURLErrorRedactsPassword(t *testing.T) {
	testCases := []struct {
		name string
		have string
	}{
		{"ShouldRedactRejectedUserInfo", "https://u:pw@e.com/"},
		{"ShouldRedactParseError", "https://u:pw@e.com/%zz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			object, err := NewObjectMethodURL(nil, []byte(fasthttp.MethodGet), []byte(tc.have))

			require.Error(t, err)

			assert.Nil(t, object)
			assert.NotContains(t, err.Error(), "pw")
		})
	}
}

func TestNewObjectMethodURLRejectsConfiguredAmbiguity(t *testing.T) {
	testCases := []struct {
		name      string
		ambiguity utils.URLAmbiguity
		have      string
	}{
		{"ShouldRejectEncodedSeparator", utils.URLAmbiguityEncodedSeparator, "https://e.com/a%2Fb"},
		{"ShouldRejectEncodedPercent", utils.URLAmbiguityEncodedPercent, "https://e.com/a%25b"},
		{"ShouldRejectDotSegment", utils.URLAmbiguityDotSegment, "https://e.com/a/../b"},
		{"ShouldRejectEmptySegment", utils.URLAmbiguityEmptySegment, "https://e.com/a//b"},
		{"ShouldRejectInvalidQueryEncoding", utils.URLAmbiguityInvalidQueryEncoding, "https://e.com/a?x=%zz"},
		{"ShouldRejectEncodedNonUnreserved", utils.URLAmbiguityEncodedNonUnreserved, "https://e.com/a%20b"},
		{"ShouldRejectEncodedUnreserved", utils.URLAmbiguityEncodedUnreserved, "https://e.com/%61"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			object, err := NewObjectMethodURL(nil, []byte(fasthttp.MethodGet), []byte(tc.have))

			require.NoError(t, err)
			require.NotNil(t, object)

			object, err = NewObjectMethodURL(utils.NewURIParser(utils.WithURLRejectAmbiguity(tc.ambiguity)), []byte(fasthttp.MethodGet), []byte(tc.have))

			assert.ErrorIs(t, err, utils.ErrURLAmbiguous)
			assert.Nil(t, object)
		})
	}
}

func TestNewObjectACLWithURIParser(t *testing.T) {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains: []string{"bypass.example.com"},
			Policy:  bypass,
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"resource.example.com"},
			Policy:    bypass,
			Resources: createSliceRegexRule(t, []string{"^/bypass/.*$"}),
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"resource.example.com"},
			Policy:    oneFactor,
			Resources: createSliceRegexRule(t, []string{"^/one_factor/.*$"}),
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"resource.example.com"},
			Policy:    twoFactor,
			Resources: createSliceRegexRule(t, []string{"^/a/longer/rule/.*$"}),
		}).
		Build()

	allowAll := utils.NewURIParser(utils.WithURLAllowAmbiguity(utils.URLAmbiguityAll))

	testCases := []struct {
		name     string
		parser   *utils.URIParser
		have     string
		expected Level
	}{
		{"ShouldDecodeEncodedUnreservedByDefault", nil, "https://resource.example.com/%62ypass/abc", Bypass},
		{"ShouldDecodeEncodedSeparatorByDefault", nil, "https://resource.example.com/a/longer%2Frule/abc", TwoFactor},
		{"ShouldResolveDotSegmentByDefault", nil, "https://resource.example.com/bypass/../a/longer/rule/abc", TwoFactor},
		{"ShouldMergeEmptySegmentByDefault", nil, "https://resource.example.com/bypass/..//a/longer/rule/abc", TwoFactor},
		{"ShouldNotDecodeEncodedPercentTwiceByDefault", nil, "https://resource.example.com/%2562ypass/abc", Denied},
		{"ShouldIgnoreInvalidQueryEncodingByDefault", nil, "https://resource.example.com/one_factor/abc?x=%zz", OneFactor},
		{"ShouldResolveEmptyDotSegmentPerRFC3986WhenAllowed", allowAll, "https://resource.example.com/bypass//../a/longer/rule/abc", Bypass},
		{"ShouldNotStripPathParameterWhenAllowed", allowAll, "https://resource.example.com/one_factor;x/abc", Denied},
		{"ShouldTraverseEncodedDotDotLower", allowAll, "https://resource.example.com/bypass/%2e%2e/a/longer/rule/abc", TwoFactor},
		{"ShouldTraverseEncodedDotDotLowerDoubleSlash", allowAll, "https://resource.example.com/bypass/%2e%2e//a/longer/rule/abc", TwoFactor},
		{"ShouldTraverseEncodedDotDotUpper", allowAll, "https://resource.example.com/bypass/%2E%2E/a/longer/rule/abc", TwoFactor},
		{"ShouldTraverseEncodedDotDotUpperDoubleSlash", allowAll, "https://resource.example.com/bypass/%2E%2E//a/longer/rule/abc", TwoFactor},
		{"ShouldResolveEncodedSingleDotLower", allowAll, "https://resource.example.com/one_factor/%2e/abc", OneFactor},
		{"ShouldResolveEncodedSingleDotUpper", allowAll, "https://resource.example.com/one_factor/%2E/abc", OneFactor},
		{"ShouldTraverseEncodedSeparatorDotDot", allowAll, "https://resource.example.com/bypass/..%2fa/longer/rule/abc", TwoFactor},
		{"ShouldNotTraverseDoubleEncodedDotDot", allowAll, "https://resource.example.com/bypass/%252e%252e/a/longer/rule/abc", Bypass},
		{"ShouldNotTraverseOverlongEncodedDotDot", allowAll, "https://resource.example.com/bypass/%c0%ae%c0%ae/a/longer/rule/abc", Bypass},
		{"ShouldNotTraverseEncodedBackslash", allowAll, "https://resource.example.com/bypass/..%5c/a/longer/rule/abc", Bypass},
		{"ShouldNotTraverseLiteralBackslash", allowAll, "https://resource.example.com/bypass/..\\/a/longer/rule/abc", Bypass},
		{"ShouldNotTraverseEncodedDotDotEncodedBackslash", allowAll, "https://resource.example.com/bypass/%2e%2e%5c/a/longer/rule/abc", Bypass},
		{"ShouldNotTraverseNullByte", allowAll, "https://resource.example.com/bypass/%00/a/longer/rule/abc", Bypass},
		{"ShouldNotTraverseSemicolonParameter", allowAll, "https://resource.example.com/bypass/..;/a/longer/rule/abc", Bypass},
		{"ShouldMatchHostNotUserInfo", allowAll, "https://bypass.example.com@resource.example.com/one_factor/abc", OneFactor},
		{"ShouldMatchHostNotUserInfoDenied", allowAll, "https://bypass.example.com:pw@other.example.com/", Denied},
		{"ShouldNotTreatQueryLikeFragmentAsPath", allowAll, "https://resource.example.com/one_factor/abc?x=#/../../bypass/", OneFactor},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			object, err := NewObjectMethodURL(tc.parser, []byte(fasthttp.MethodGet), []byte(tc.have))

			require.NoError(t, err)

			_, level := tester.GetRequiredLevel(Subject{Username: "john"}, *object)

			assert.Equal(t, tc.expected, level)
		})
	}
}

func TestNewObjectPathAndCanonicalURIAgree(t *testing.T) {
	segments := []string{"a", "b", "", ".", "..", "%2e", "%2E%2E"}

	parser := utils.NewURIParser(utils.WithURLAllowAmbiguity(utils.URLAmbiguityAll))

	flatten := func(value string) (flat string) {
		parts := make([]string, 0)

		for _, segment := range strings.Split(value, "/") {
			if segment != "" {
				parts = append(parts, segment)
			}
		}

		return "/" + strings.Join(parts, "/")
	}

	var walk func(prefix string, depth int)

	walk = func(prefix string, depth int) {
		if prefix != "" {
			object, err := NewObjectMethodURL(parser, []byte(fasthttp.MethodGet), []byte("https://e.com"+prefix))

			require.NoError(t, err)

			assert.Equal(t, flatten(object.Path), flatten(strings.TrimPrefix(object.CanonicalURI, "https://e.com")), "request path '%s'", prefix)
		}

		if depth == 0 {
			return
		}

		for _, segment := range segments {
			walk(prefix+"/"+segment, depth-1)
		}
	}

	walk("", 5)
}

func TestNewObjectEmptyPath(t *testing.T) {
	testCases := []struct {
		name      string
		have      string
		path      string
		canonical string
	}{
		{"ShouldTreatEmptyPathAsSlash", "https://app.example.com", "/", "https://app.example.com/"},
		{"ShouldTreatEmptyPathAsSlashWithQuery", "https://app.example.com?x=1", "/?x=1", "https://app.example.com/?x=1"},
	}

	authorizer := NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: deny,
			Rules: []schema.AccessControlRule{
				{
					Domains:   []string{"app.example.com"},
					Resources: createSliceRegexRule(t, []string{"^/$", "^/\\?.*$"}),
					Policy:    bypass,
				},
			},
		},
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			object, err := NewObjectMethodURL(utils.DefaultURIParser(), []byte(fasthttp.MethodGet), []byte(tc.have))

			require.NoError(t, err)

			assert.Equal(t, tc.path, object.Path)
			assert.Equal(t, tc.canonical, object.CanonicalURI)

			_, level := authorizer.GetRequiredLevel(Subject{}, *object)

			assert.Equal(t, Bypass, level)
		})
	}
}
