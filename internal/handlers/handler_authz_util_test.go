package handlers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestParseAuthzPortalURL(t *testing.T) {
	testCases := []struct {
		Name     string
		Have     []byte
		Expected string
		Error    string
	}{
		{
			Name: "ShouldReturnNilForNilValue",
			Have: nil,
		},
		{
			Name:  "ShouldReturnErrorForEmptyValue",
			Have:  []byte(""),
			Error: "parse \"\": empty url",
		},
		{
			Name:     "ShouldParseAbsoluteURL",
			Have:     []byte("https://auth.example.com"),
			Expected: "https://auth.example.com",
		},
		{
			Name:     "ShouldParseAbsoluteURLWithPath",
			Have:     []byte("https://auth.example.com/path"),
			Expected: "https://auth.example.com/path",
		},
		{
			Name:     "ShouldParseAbsoluteURLWithPort",
			Have:     []byte("https://auth.example.com:9091/"),
			Expected: "https://auth.example.com:9091/",
		},
		{
			Name:     "ShouldParseAbsoluteURLWithQuery",
			Have:     []byte("https://auth.example.com/?rd=https%3A%2F%2Fapp.example.com"),
			Expected: "https://auth.example.com/?rd=https%3A%2F%2Fapp.example.com",
		},
		{
			Name:     "ShouldParseAbsolutePath",
			Have:     []byte("/path"),
			Expected: "/path",
		},
		{
			Name:  "ShouldReturnErrorForRelativePath",
			Have:  []byte("path"),
			Error: "parse \"path\": invalid URI for request",
		},
		{
			Name:  "ShouldReturnErrorForInvalidControlCharacter",
			Have:  []byte("https://auth.example.com/\x7f"),
			Error: "parse \"https://auth.example.com/\\x7f\": net/url: invalid control character in URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := parseAuthzPortalURL(tc.Have)

			switch {
			case tc.Error != "":
				assert.EqualError(t, err, tc.Error)
				assert.Nil(t, actual)
			case tc.Expected == "":
				assert.NoError(t, err)
				assert.Nil(t, actual)
			default:
				assert.NoError(t, err)
				require.NotNil(t, actual)
				assert.Equal(t, tc.Expected, actual.String())
			}
		})
	}
}

func TestGetAuthzRedirectStatusCode(t *testing.T) {
	testCases := []struct {
		Name     string
		Method   string
		Accept   string
		XHR      bool
		Expected int
	}{
		{
			Name:     "ShouldReturnFoundForGet",
			Method:   fasthttp.MethodGet,
			Accept:   "text/html",
			Expected: fasthttp.StatusFound,
		},
		{
			Name:     "ShouldReturnFoundForOptions",
			Method:   fasthttp.MethodOptions,
			Accept:   "text/html",
			Expected: fasthttp.StatusFound,
		},
		{
			Name:     "ShouldReturnFoundForHead",
			Method:   fasthttp.MethodHead,
			Accept:   "text/html",
			Expected: fasthttp.StatusFound,
		},
		{
			Name:     "ShouldReturnFoundForEmptyMethod",
			Method:   "",
			Accept:   "text/html",
			Expected: fasthttp.StatusFound,
		},
		{
			Name:     "ShouldReturnSeeOtherForPost",
			Method:   fasthttp.MethodPost,
			Accept:   "text/html",
			Expected: fasthttp.StatusSeeOther,
		},
		{
			Name:     "ShouldReturnSeeOtherForPut",
			Method:   fasthttp.MethodPut,
			Accept:   "text/html",
			Expected: fasthttp.StatusSeeOther,
		},
		{
			Name:     "ShouldReturnSeeOtherForDelete",
			Method:   fasthttp.MethodDelete,
			Accept:   "text/html",
			Expected: fasthttp.StatusSeeOther,
		},
		{
			Name:     "ShouldReturnFoundForAcceptWildcard",
			Method:   fasthttp.MethodGet,
			Accept:   "*/*",
			Expected: fasthttp.StatusFound,
		},
		{
			Name:     "ShouldReturnFoundForAcceptListIncludingHTML",
			Method:   fasthttp.MethodGet,
			Accept:   "application/json, text/html;q=0.9",
			Expected: fasthttp.StatusFound,
		},
		{
			Name:     "ShouldReturnUnauthorizedForXHR",
			Method:   fasthttp.MethodGet,
			Accept:   "text/html",
			XHR:      true,
			Expected: fasthttp.StatusUnauthorized,
		},
		{
			Name:     "ShouldReturnUnauthorizedForXHRWithPost",
			Method:   fasthttp.MethodPost,
			Accept:   "text/html",
			XHR:      true,
			Expected: fasthttp.StatusUnauthorized,
		},
		{
			Name:     "ShouldReturnUnauthorizedWhenHTMLNotAccepted",
			Method:   fasthttp.MethodGet,
			Accept:   "application/json",
			Expected: fasthttp.StatusUnauthorized,
		},
		{
			Name:     "ShouldReturnUnauthorizedWhenAcceptEmpty",
			Method:   fasthttp.MethodGet,
			Expected: fasthttp.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			if tc.Accept != "" {
				mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, tc.Accept)
			}

			if tc.XHR {
				mock.Ctx.Request.Header.Set("X-Requested-With", "XMLHttpRequest")
			}

			assert.Equal(t, tc.Expected, getAuthzRedirectStatusCode(mock.Ctx, tc.Method))
		})
	}
}

func TestDoAuthzRedirect(t *testing.T) {
	testCases := []struct {
		Name         string
		Method       string
		StatusCode   int
		Expected     int
		ExpectedBody string
	}{
		{
			Name:         "ShouldRedirectWithBodyForGet",
			Method:       fasthttp.MethodGet,
			StatusCode:   fasthttp.StatusFound,
			Expected:     fasthttp.StatusFound,
			ExpectedBody: "<a href=\"https://auth.example.com/?rd=https%3A%2F%2Fapp.example.com\">302 Found</a>",
		},
		{
			Name:         "ShouldRedirectWithBodyForPost",
			Method:       fasthttp.MethodPost,
			StatusCode:   fasthttp.StatusSeeOther,
			Expected:     fasthttp.StatusSeeOther,
			ExpectedBody: "<a href=\"https://auth.example.com/?rd=https%3A%2F%2Fapp.example.com\">303 See Other</a>",
		},
		{
			Name:       "ShouldRedirectWithoutBodyForHead",
			Method:     fasthttp.MethodHead,
			StatusCode: fasthttp.StatusFound,
			Expected:   fasthttp.StatusFound,
		},
		{
			Name:       "ShouldRedirectWithoutBodyForHeadWithUnauthorized",
			Method:     fasthttp.MethodHead,
			StatusCode: fasthttp.StatusUnauthorized,
			Expected:   fasthttp.StatusUnauthorized,
		},
		{
			Name:         "ShouldRedirectWithBodyForUnauthorized",
			Method:       fasthttp.MethodGet,
			StatusCode:   fasthttp.StatusUnauthorized,
			Expected:     fasthttp.StatusUnauthorized,
			ExpectedBody: "<a href=\"https://auth.example.com/?rd=https%3A%2F%2Fapp.example.com\">401 Unauthorized</a>",
		},
	}

	redirectionURL, err := url.ParseRequestURI("https://auth.example.com/?rd=https%3A%2F%2Fapp.example.com")
	require.NoError(t, err)

	targetURL, err := url.ParseRequestURI("https://app.example.com")
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			authn := &Authn{
				Username: "john",
				Method:   friendlyMethod(tc.Method),
				Object:   authorization.NewObject(targetURL, tc.Method),
			}

			doAuthzRedirect(mock.Ctx, authn, redirectionURL, tc.StatusCode)

			assert.Equal(t, tc.Expected, mock.Ctx.Response.StatusCode())
			assert.Equal(t, redirectionURL.String(), string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
			assert.Equal(t, tc.ExpectedBody, string(mock.Ctx.Response.Body()))
		})
	}
}

func TestGetSafeAutheliaURL(t *testing.T) {
	testCases := []struct {
		Name     string
		Have     string
		Domain   string
		Expected string
		Error    string
	}{
		{
			Name:     "ShouldAllowExactDomainMatch",
			Have:     "https://example.com",
			Domain:   "example.com",
			Expected: "https://example.com",
		},
		{
			Name:     "ShouldAllowSubdomain",
			Have:     "https://auth.example.com",
			Domain:   "example.com",
			Expected: "https://auth.example.com",
		},
		{
			Name:     "ShouldAllowDeepSubdomain",
			Have:     "https://auth.internal.example.com/path",
			Domain:   "example.com",
			Expected: "https://auth.internal.example.com/path",
		},
		{
			Name:     "ShouldAllowDomainWithLeadingPeriod",
			Have:     "https://auth.example.com",
			Domain:   ".example.com",
			Expected: "https://auth.example.com",
		},
		{
			Name:     "ShouldAllowMismatchedCase",
			Have:     "https://AUTH.Example.COM",
			Domain:   "example.com",
			Expected: "https://AUTH.Example.COM",
		},
		{
			Name:     "ShouldAllowPort",
			Have:     "https://auth.example.com:9091/",
			Domain:   "example.com",
			Expected: "https://auth.example.com:9091/",
		},
		{
			Name:     "ShouldAllowInsecureScheme",
			Have:     "http://auth.example.com",
			Domain:   "example.com",
			Expected: "http://auth.example.com",
		},
		{
			Name:   "ShouldNotAllowDifferentDomain",
			Have:   "https://auth.example.org",
			Domain: "example.com",
			Error:  "authelia url 'https://auth.example.org' is not valid for detected domain 'example.com' as the url does not have the domain as a suffix",
		},
		{
			Name:   "ShouldNotAllowSuffixWithoutPeriodBoundary",
			Have:   "https://notexample.com",
			Domain: "example.com",
			Error:  "authelia url 'https://notexample.com' is not valid for detected domain 'example.com' as the url does not have the domain as a suffix",
		},
		{
			Name:   "ShouldNotAllowParentDomain",
			Have:   "https://example.com",
			Domain: "auth.example.com",
			Error:  "authelia url 'https://example.com' is not valid for detected domain 'auth.example.com' as the url does not have the domain as a suffix",
		},
		{
			Name:   "ShouldNotAllowDomainInPath",
			Have:   "https://auth.example.org/example.com",
			Domain: "example.com",
			Error:  "authelia url 'https://auth.example.org/example.com' is not valid for detected domain 'example.com' as the url does not have the domain as a suffix",
		},
		{
			Name:   "ShouldNotAllowEmptyDomain",
			Have:   "https://auth.example.com",
			Domain: "",
			Error:  "authelia url 'https://auth.example.com' is not valid for detected domain '' as the url does not have the domain as a suffix",
		},
		{
			Name:   "ShouldNotAllowEmptyHost",
			Have:   "https:///path",
			Domain: "example.com",
			Error:  "authelia url 'https:///path' is not valid for detected domain 'example.com' as the url does not have the domain as a suffix",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			have, err := url.Parse(tc.Have)
			require.NoError(t, err)

			actual, err := getSafeAutheliaURL(have, tc.Domain)

			if tc.Error == "" {
				assert.NoError(t, err)
				require.NotNil(t, actual)
				assert.Equal(t, tc.Expected, actual.String())
				assert.Same(t, have, actual)
			} else {
				assert.EqualError(t, err, tc.Error)
				assert.Nil(t, actual)
			}
		})
	}
}
