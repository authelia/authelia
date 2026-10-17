// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestHandleOAuth2AuthorizationBounceShouldNotBounce(t *testing.T) {
	marker := map[string]string{cookieNameOAuth2AuthorizationBounce: cookieValueOAuth2AuthorizationBounce}

	testCases := []struct {
		name    string
		site    string
		mode    string
		dest    string
		query   string
		cookies map[string]string
		cleared bool
	}{
		{"ShouldHandleNoHeadersWithSessionCookie", "", "", "", "client_id=abc", map[string]string{"authelia_session": "value"}, false},
		{"ShouldHandleSameOrigin", "same-origin", "navigate", "document", "client_id=abc", nil, false},
		{"ShouldHandleSameSite", "same-site", "navigate", "document", "client_id=abc", nil, false},
		{"ShouldHandleNone", "none", "navigate", "document", "client_id=abc", nil, false},
		{"ShouldHandleModeCORS", "cross-site", "cors", "document", "client_id=abc", nil, false},
		{"ShouldHandleModeNoCORS", "cross-site", "no-cors", "document", "client_id=abc", nil, false},
		{"ShouldHandleModeAbsent", "cross-site", "", "document", "client_id=abc", nil, false},
		{"ShouldHandleDestIFrame", "cross-site", "navigate", "iframe", "client_id=abc", nil, false},
		{"ShouldHandleDestEmpty", "cross-site", "navigate", "empty", "client_id=abc", nil, false},
		{"ShouldHandleDestAbsent", "cross-site", "navigate", "", "client_id=abc", nil, false},
		{"ShouldHandleQueryWithSemicolon", "cross-site", "navigate", "document", "client_id=abc;url=https://evil.example.net/", nil, false},
		{"ShouldHandleQueryWithQuotedSemicolon", "cross-site", "navigate", "document", "';url=https://evil.example.net/", nil, false},
		{"ShouldHandleQueryWithBadEscape", "cross-site", "navigate", "document", "client_id=abc&state=%zz", nil, false},
		{"ShouldHandleSessionCookiePresent", "cross-site", "navigate", "document", "client_id=abc", map[string]string{"authelia_session": "value"}, false},
		{"ShouldHandleMarkerPresentSameOrigin", "same-origin", "navigate", "document", "client_id=abc", marker, true},
		{"ShouldHandleMarkerPresentSameSite", "same-site", "navigate", "document", "client_id=abc", marker, true},
		{"ShouldHandleMarkerPresentNone", "none", "navigate", "document", "client_id=abc", marker, true},
		{"ShouldHandleMarkerPresentWithoutHeaders", "", "", "", "client_id=abc", marker, true},
		{"ShouldHandleMarkerPresentWithSessionCookie", "same-origin", "navigate", "document", "client_id=abc", map[string]string{cookieNameOAuth2AuthorizationBounce: cookieValueOAuth2AuthorizationBounce, "authelia_session": "value"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			headers := map[string]string{}

			if tc.site != "" {
				headers[headerSecFetchSite] = tc.site
			}

			if tc.mode != "" {
				headers[headerSecFetchMode] = tc.mode
			}

			if tc.dest != "" {
				headers[headerSecFetchDest] = tc.dest
			}

			rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodGet, "", nil, headers, tc.cookies)

			r.URL.RawQuery = tc.query

			assert.False(t, handleOAuth2AuthorizationBounce(mock.Ctx, rw, r))

			assert.Empty(t, rw.Body.String())
			assert.Empty(t, rw.Header().Get(fasthttp.HeaderContentSecurityPolicy))
			assert.Empty(t, rw.Header().Get(fasthttp.HeaderContentType))
			assert.Equal(t, tc.query, r.URL.RawQuery)

			cookies := rw.Result().Cookies()

			if !tc.cleared {
				assert.Empty(t, rw.Header())
				assert.Empty(t, cookies)

				return
			}

			require.Len(t, cookies, 1)

			assert.Equal(t, cookieNameOAuth2AuthorizationBounce, cookies[0].Name)
			assert.Equal(t, oidc.EndpointPathAuthorization, cookies[0].Path)
			assert.Equal(t, -1, cookies[0].MaxAge)
			assert.Contains(t, rw.Header().Get(fasthttp.HeaderSetCookie), "Max-Age=0")
		})
	}
}

func TestHandleOAuth2AuthorizationBounceShouldNotBounceOtherMethods(t *testing.T) {
	testCases := []struct {
		name   string
		method string
	}{
		{"ShouldHandlePOST", fasthttp.MethodPost},
		{"ShouldHandlePUT", fasthttp.MethodPut},
		{"ShouldHandleHEAD", fasthttp.MethodHead},
		{"ShouldHandleDELETE", fasthttp.MethodDelete},
		{"ShouldHandleLowerCaseGET", "get"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			form := url.Values{oidc.FormParameterClientID: []string{"abc"}}

			rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodPost, "client_id=abc", form, newTestOAuth2AuthorizationBounceHeaders(), nil)

			r.Method = tc.method

			assert.False(t, handleOAuth2AuthorizationBounce(mock.Ctx, rw, r))

			assert.Empty(t, rw.Body.String())
			assert.Empty(t, rw.Header())
			assert.Empty(t, rw.Result().Cookies())
			assert.Equal(t, "client_id=abc", r.URL.RawQuery)

			assertTestOAuth2AuthorizationBounceRequestIntact(t, r, form.Encode())
		})
	}
}

func TestHandleOAuth2AuthorizationBounceShouldNotBounceWithoutASessionProvider(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")

	_, err := mock.Ctx.GetSessionProvider()

	require.EqualError(t, err, "unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")

	rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodGet, "client_id=abc", nil, newTestOAuth2AuthorizationBounceHeaders(), nil)

	assert.False(t, handleOAuth2AuthorizationBounce(mock.Ctx, rw, r))

	assert.Empty(t, rw.Body.String())
	assert.Empty(t, rw.Header())
	assert.Empty(t, rw.Result().Cookies())
	assert.Equal(t, "client_id=abc", r.URL.RawQuery)
}

func TestHandleOAuth2AuthorizationBounceShouldBounceGET(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodGet, "client_id=abc&prompt=none&state=%3Cscript%3E%22%27&redirect_uri=https%3A%2F%2Fapp.example2.com%2Fcb", nil, newTestOAuth2AuthorizationBounceHeaders(), nil)

	require.True(t, handleOAuth2AuthorizationBounce(mock.Ctx, rw, r))

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.True(t, strings.HasPrefix(rw.Header().Get(fasthttp.HeaderContentType), "text/html"))
	assert.Equal(t, "no-store", rw.Header().Get(fasthttp.HeaderCacheControl))
	assert.Equal(t, "no-referrer", rw.Header().Get(fasthttp.HeaderReferrerPolicy))
	assert.Empty(t, rw.Header().Get(fasthttp.HeaderContentSecurityPolicy))

	assertTestOAuth2AuthorizationBounceCookie(t, rw, "/api/oidc/authorization")

	body := rw.Body.String()

	assert.Contains(t, body, `<meta http-equiv="refresh"`)
	assert.NotContains(t, body, "<script")
	assert.NotContains(t, body, "<form")
	assert.NotContains(t, body, testOAuth2AuthorizationBounceState)
	assert.NotContains(t, body, "state=<")
	assert.Contains(t, body, `<a href="/api/oidc/authorization?client_id=abc&amp;prompt=none&amp;redirect_uri=https%3A%2F%2Fapp.example2.com%2Fcb&amp;state=%3Cscript%3E%22%27">`)

	match := regexpOAuth2AuthorizationBounceLink.FindStringSubmatch(body)

	require.Len(t, match, 2)

	target, err := url.Parse(html.UnescapeString(match[1]))

	require.NoError(t, err)

	assert.Empty(t, target.Scheme)
	assert.Empty(t, target.Host)
	assert.Equal(t, "/api/oidc/authorization", target.Path)
	assert.Equal(t, url.Values{
		"client_id":    []string{"abc"},
		"prompt":       []string{"none"},
		"state":        []string{testOAuth2AuthorizationBounceState},
		"redirect_uri": []string{"https://app.example2.com/cb"},
	}, target.Query())

	match = regexpOAuth2AuthorizationBounceRefresh.FindStringSubmatch(body)

	require.Len(t, match, 2)

	assert.Equal(t, "0;url="+target.String(), html.UnescapeString(match[1]))
}

func TestHandleOAuth2AuthorizationBounceShouldBounceWithoutFetchMetadata(t *testing.T) {
	testCases := []struct {
		name    string
		headers map[string]string
	}{
		{"ShouldHandleNoHeaders", nil},
		{"ShouldHandleNoSite", map[string]string{headerSecFetchMode: valueSecFetchModeNavigate, headerSecFetchDest: valueSecFetchDestDocument}},
		{"ShouldHandleNoSiteWithOtherModeAndDest", map[string]string{headerSecFetchMode: "cors", headerSecFetchDest: "empty"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodGet, "client_id=abc&prompt=none", nil, tc.headers, nil)

			require.True(t, handleOAuth2AuthorizationBounce(mock.Ctx, rw, r))

			assert.Equal(t, http.StatusOK, rw.Code)

			assertTestOAuth2AuthorizationBounceCookie(t, rw, "/api/oidc/authorization")

			assert.Contains(t, rw.Body.String(), `<meta http-equiv="refresh" content="0;url=/api/oidc/authorization?client_id=abc&amp;prompt=none">`)
		})
	}
}

func TestHandleOAuth2AuthorizationBounceShouldBounceCrossSiteWithMarker(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodGet, "client_id=abc", nil, newTestOAuth2AuthorizationBounceHeaders(), map[string]string{cookieNameOAuth2AuthorizationBounce: cookieValueOAuth2AuthorizationBounce})

	require.True(t, handleOAuth2AuthorizationBounce(mock.Ctx, rw, r))

	assertTestOAuth2AuthorizationBounceCookie(t, rw, "/api/oidc/authorization")

	assert.Contains(t, rw.Body.String(), `<meta http-equiv="refresh" content="0;url=/api/oidc/authorization?client_id=abc">`)
}

func TestHandleOAuth2AuthorizationBounceShouldBounceGETWithoutQuery(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodGet, "", nil, map[string]string{headerSecFetchSite: "Cross-Site", headerSecFetchMode: "Navigate", headerSecFetchDest: "Document"}, map[string]string{"other": "value"})

	require.True(t, handleOAuth2AuthorizationBounce(mock.Ctx, rw, r))

	body := rw.Body.String()

	assert.Contains(t, body, `<meta http-equiv="refresh" content="0;url=/api/oidc/authorization">`)
	assert.Contains(t, body, `<a href="/api/oidc/authorization">`)
}

func TestHandleOAuth2AuthorizationBounceShouldUseTheBasePath(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.SetUserValue(middlewares.UserValueKeyBaseURL, testOAuth2AuthorizationBounceBasePath)

	rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodGet, "client_id=abc", nil, newTestOAuth2AuthorizationBounceHeaders(), nil)

	require.True(t, handleOAuth2AuthorizationBounce(mock.Ctx, rw, r))

	assertTestOAuth2AuthorizationBounceCookie(t, rw, "/authelia/api/oidc/authorization")

	body := rw.Body.String()

	assert.Contains(t, body, `<a href="/authelia/api/oidc/authorization?client_id=abc">`)
	assert.Contains(t, body, `<meta http-equiv="refresh" content="0;url=/authelia/api/oidc/authorization?client_id=abc">`)
}

func TestHandleOAuth2AuthorizationBounceShouldEscapeTheTarget(t *testing.T) {
	testCases := []struct {
		name     string
		query    string
		expected string
	}{
		{
			"ShouldHandleEntityLikeNames",
			"a=1&copy=2&lt=3&amp=4&quot=5&reg=6",
			"/api/oidc/authorization?a=1&amp=4&copy=2&lt=3&quot=5&reg=6",
		},
		{
			"ShouldHandleURLParameter",
			"url=https://evil.example.net/&x=1",
			"/api/oidc/authorization?url=https%3A%2F%2Fevil.example.net%2F&x=1",
		},
		{
			"ShouldHandleRefreshInjectionEncoded",
			"x=1%3Burl%3Dhttps%3A%2F%2Fevil.example.net%2F",
			"/api/oidc/authorization?x=1%3Burl%3Dhttps%3A%2F%2Fevil.example.net%2F",
		},
		{
			"ShouldHandleAttributeBreakout",
			`x="><a href="https://evil.example.net/"><form id="bounce" method="post" action="https://evil.example.net/">`,
			"/api/oidc/authorization?x=%22%3E%3Ca+href%3D%22https%3A%2F%2Fevil.example.net%2F%22%3E%3Cform+id%3D%22bounce%22+method%3D%22post%22+action%3D%22https%3A%2F%2Fevil.example.net%2F%22%3E",
		},
		{
			"ShouldHandleAttributeBreakoutEncoded",
			"x=%22%3E%3Cmeta%20http-equiv%3D%22refresh%22%20content%3D%220%3Burl%3Dhttps%3A%2F%2Fevil.example.net%2F%22%3E",
			"/api/oidc/authorization?x=%22%3E%3Cmeta+http-equiv%3D%22refresh%22+content%3D%220%3Burl%3Dhttps%3A%2F%2Fevil.example.net%2F%22%3E",
		},
		{
			"ShouldHandleScriptBreakout",
			`x=</title><script>alert(1)</script>&y='onmouseover='alert(1)`,
			"/api/oidc/authorization?x=%3C%2Ftitle%3E%3Cscript%3Ealert%281%29%3C%2Fscript%3E&y=%27onmouseover%3D%27alert%281%29",
		},
		{
			"ShouldHandleSchemeLikeValues",
			"javascript:alert(1)=javascript:alert(1)&//evil.example.net=//evil.example.net",
			"/api/oidc/authorization?%2F%2Fevil.example.net=%2F%2Fevil.example.net&javascript%3Aalert%281%29=javascript%3Aalert%281%29",
		},
		{
			"ShouldHandleFragment",
			"x=1%23https://evil.example.net/",
			"/api/oidc/authorization?x=1%23https%3A%2F%2Fevil.example.net%2F",
		},
		{
			"ShouldHandleWhitespace",
			"x=%0A%0D%09%20//evil.example.net/&y=a+b",
			"/api/oidc/authorization?x=%0A%0D%09+%2F%2Fevil.example.net%2F&y=a+b",
		},
		{
			"ShouldHandleBackslashes",
			`x=\\evil.example.net\&y=%5C%5Cevil.example.net`,
			"/api/oidc/authorization?x=%5C%5Cevil.example.net%5C&y=%5C%5Cevil.example.net",
		},
		{
			"ShouldHandleTemplateSyntax",
			"x={{.Nonce}}&y=%7B%7B.Target%7D%7D",
			"/api/oidc/authorization?x=%7B%7B.Nonce%7D%7D&y=%7B%7B.Target%7D%7D",
		},
		{
			"ShouldHandleCraftedNames",
			"%22%3E%3Cscript%3Ealert(1)%3C%2Fscript%3E=%22%3E%3Cimg+src%3Dx+onerror%3Dalert(1)%3E%26amp%3B",
			"/api/oidc/authorization?%22%3E%3Cscript%3Ealert%281%29%3C%2Fscript%3E=%22%3E%3Cimg+src%3Dx+onerror%3Dalert%281%29%3E%26amp%3B",
		},
		{
			"ShouldHandleRepeatedNames",
			"resource=b&resource=a&client_id=abc",
			"/api/oidc/authorization?client_id=abc&resource=b&resource=a",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodGet, "", nil, newTestOAuth2AuthorizationBounceHeaders(), nil)

			r.URL.RawQuery = tc.query

			require.True(t, handleOAuth2AuthorizationBounce(mock.Ctx, rw, r))

			body := rw.Body.String()

			assert.NotContains(t, body, "<meta http-equiv=\"refresh\" content=\"0;url=https")
			assert.NotContains(t, body, "alert(1)<")
			assert.NotContains(t, body, "evil.example.net/\"")
			assert.NotContains(t, body, "<img")

			assert.Equal(t, 0, strings.Count(body, "<script"))
			assert.Equal(t, 0, strings.Count(body, "<form"))
			assert.Equal(t, 1, strings.Count(body, "<meta http-equiv"))

			link := regexpOAuth2AuthorizationBounceLink.FindAllStringSubmatch(body, -1)

			require.Len(t, link, 1)

			refresh := regexpOAuth2AuthorizationBounceRefresh.FindAllStringSubmatch(body, -1)

			require.Len(t, refresh, 1)

			content := html.UnescapeString(refresh[0][1])

			require.True(t, strings.HasPrefix(content, "0;url="))

			for _, target := range []string{html.UnescapeString(link[0][1]), strings.TrimPrefix(content, "0;url=")} {
				assert.Equal(t, tc.expected, target)
				assert.True(t, target == "/api/oidc/authorization" || strings.HasPrefix(target, "/api/oidc/authorization?"))
				assert.NotContains(t, target, ";")
				assert.NotContains(t, target, "'")
				assert.NotContains(t, target, `"`)
				assert.NotContains(t, target, " ")
				assert.NotContains(t, target, "#")

				parsed, err := url.Parse(target)

				require.NoError(t, err)

				assert.Empty(t, parsed.Scheme)
				assert.Empty(t, parsed.Host)
				assert.Empty(t, parsed.Fragment)
				assert.Equal(t, "/api/oidc/authorization", parsed.Path)
				assert.Equal(t, r.URL.Query(), parsed.Query())
			}
		})
	}
}

func TestOAuth2AuthorizationShouldBounceBeforeProcessing(t *testing.T) {
	testCases := []struct {
		name    string
		method  string
		handler middlewares.AutheliaHandlerFunc
		bounced bool
	}{
		{"ShouldBounceGET", fasthttp.MethodGet, OAuth2AuthorizationGET, true},
		{"ShouldNotBouncePOST", fasthttp.MethodPost, OAuth2AuthorizationPOST, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			setupTestOIDCProvider(t, mock, nil)

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "http")

			_, err := mock.Ctx.IssuerURL()

			require.EqualError(t, err, "invalid X-Forwarded-Proto header value 'http'")

			values := newTestOIDCAuthorizationValues()

			rw, r := newTestOAuth2AuthorizationBounceRequest(tc.method, values.Encode(), values, newTestOAuth2AuthorizationBounceHeaders(), nil)

			tc.handler(mock.Ctx, rw, r)

			if !tc.bounced {
				assert.NotContains(t, rw.Body.String(), "Continue")
				assert.NotContains(t, rw.Body.String(), `<meta http-equiv="refresh"`)
				assert.Empty(t, rw.Result().Cookies())

				require.NotNil(t, mock.Hook.LastEntry())

				assert.Contains(t, mock.Hook.LastEntry().Message, "could not be processed")

				return
			}

			assert.Equal(t, http.StatusOK, rw.Code)
			assert.Empty(t, rw.Header().Get(fasthttp.HeaderLocation))
			assert.Contains(t, rw.Body.String(), `/api/oidc/authorization?`)
			assert.Contains(t, rw.Body.String(), "Continue")
			assert.Nil(t, mock.Hook.LastEntry())

			assertTestOAuth2AuthorizationBounceCookie(t, rw, "/api/oidc/authorization")

			rw, r = newTestOAuth2AuthorizationBounceRequest(tc.method, values.Encode(), values, newTestOAuth2AuthorizationBounceFollowUpHeaders(), map[string]string{cookieNameOAuth2AuthorizationBounce: cookieValueOAuth2AuthorizationBounce})

			tc.handler(mock.Ctx, rw, r)

			assert.NotContains(t, rw.Body.String(), "Continue")

			require.NotNil(t, mock.Hook.LastEntry())

			assert.Contains(t, mock.Hook.LastEntry().Message, "could not be processed")

			cookies := rw.Result().Cookies()

			require.Len(t, cookies, 1)

			assert.Equal(t, cookieNameOAuth2AuthorizationBounce, cookies[0].Name)
			assert.Equal(t, -1, cookies[0].MaxAge)
		})
	}
}

func TestOAuth2AuthorizationPOSTShouldRedirectWhenCrossSite(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	setupTestOIDCProvider(t, mock, nil)

	rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodPost, "prompt=none", newTestOIDCAuthorizationValues(), newTestOAuth2AuthorizationBounceHeaders(), nil)

	OAuth2AuthorizationPOST(mock.Ctx, rw, r)

	require.Equal(t, http.StatusSeeOther, rw.Code)

	assert.Empty(t, rw.Body.String())
	assert.Empty(t, rw.Header().Get(fasthttp.HeaderContentType))
	assert.Empty(t, rw.Header().Get(fasthttp.HeaderSetCookie))
	assert.Empty(t, rw.Result().Cookies())

	location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

	require.NoError(t, err)

	assert.Equal(t, "https", location.Scheme)
	assert.Equal(t, "/api/oidc/authorization", location.Path)
	assert.Equal(t, url.Values{
		"client_id":     []string{"authorization-code"},
		"response_type": []string{"code"},
		"redirect_uri":  []string{"https://app.example.com/oidc/callback"},
		"scope":         []string{"openid"},
		"state":         []string{"abcdefghijklmnopqrstuvwxyz"},
		"prompt":        []string{"none"},
	}, location.Query())
}

func TestOAuth2AuthorizationPOSTShouldReportMalformedMultipartFormWhenCrossSite(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	setupTestOIDCProvider(t, mock, nil)

	rw, r := newTestOAuth2AuthorizationBounceRequest(fasthttp.MethodPost, "", nil, newTestOAuth2AuthorizationBounceHeaders(), nil)

	r.Header.Set(fasthttp.HeaderContentType, "multipart/form-data; boundary=abc")
	r.Body = io.NopCloser(strings.NewReader("not a valid multipart body"))
	r.ContentLength = 26

	OAuth2AuthorizationPOST(mock.Ctx, rw, r)

	require.Equal(t, http.StatusSeeOther, rw.Code)

	location, err := url.Parse(rw.Header().Get(fasthttp.HeaderLocation))

	require.NoError(t, err)

	assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
	assert.NotEmpty(t, location.Query().Get("error"))
	assert.Empty(t, rw.Result().Cookies())

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAuthorizationMultipartError, regexpAnyError)
}

func TestOAuth2AuthorizationBounceShouldSurviveTheServerMiddlewares(t *testing.T) {
	testCases := []struct {
		name string
		base string
	}{
		{"ShouldHandleGET", ""},
		{"ShouldHandleGETWithBasePath", testOAuth2AuthorizationBounceBasePath},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			config := newTestOIDCConfig(t)
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

			setupTestOIDCProvider(t, mock, config)

			bridge := middlewares.NewBridgeBuilder(mock.Ctx.Configuration, mock.Ctx.Providers).WithPreMiddlewares(
				middlewares.SecurityHeadersBase, middlewares.SecurityHeadersCSPNoneOpenIDConnect, middlewares.SecurityHeadersNoStore,
			).Build()

			policy := middlewares.NewCORSPolicyBuilder().
				WithAllowedMethods(fasthttp.MethodOptions, fasthttp.MethodGet, fasthttp.MethodPost).
				WithEnabled(false).
				Build()

			handler := middlewares.StripPath(tc.base)(policy.Middleware(bridge(middlewares.NewHTTPToAutheliaHandlerAdaptor(OAuth2AuthorizationGET))))

			values := newTestOIDCAuthorizationValues()

			path := tc.base + oidc.EndpointPathAuthorization

			ctx := newTestOAuth2AuthorizationBounceRequestCtx(path, values, "")

			handler(ctx)

			assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())

			for _, name := range []string{fasthttp.HeaderContentSecurityPolicy, fasthttp.HeaderXFrameOptions, fasthttp.HeaderCacheControl, fasthttp.HeaderReferrerPolicy, fasthttp.HeaderContentType} {
				assert.Len(t, ctx.Response.Header.PeekAll(name), 1, name)
			}

			assert.Equal(t, [][]byte{[]byte("default-src 'none'")}, ctx.Response.Header.PeekAll(fasthttp.HeaderContentSecurityPolicy))
			assert.Equal(t, "DENY", string(ctx.Response.Header.Peek(fasthttp.HeaderXFrameOptions)))
			assert.Equal(t, "no-store", string(ctx.Response.Header.Peek(fasthttp.HeaderCacheControl)))
			assert.Equal(t, "no-referrer", string(ctx.Response.Header.Peek(fasthttp.HeaderReferrerPolicy)))
			assert.Equal(t, "text/html; charset=utf-8", string(ctx.Response.Header.Peek(fasthttp.HeaderContentType)))

			cookies := ctx.Response.Header.PeekAll(fasthttp.HeaderSetCookie)

			require.Len(t, cookies, 1)

			assert.Equal(t, "authelia_authorization_bounce=1; Path="+path+"; Max-Age=60; HttpOnly; Secure; SameSite=Lax", string(cookies[0]))

			body := string(ctx.Response.Body())

			assert.Contains(t, body, `<meta http-equiv="refresh" content="0;url=`+path+`?`)
			assert.NotContains(t, body, "<script")
			assert.NotContains(t, body, "<style")
			assert.NotContains(t, body, "<form")
			assert.NotContains(t, body, "<link")
			assert.NotContains(t, body, "<img")
			assert.NotContains(t, body, "style=")

			setupTestOIDCConsentStore(t, mock)

			ctx = newTestOAuth2AuthorizationBounceRequestCtx(path, values, cookieNameOAuth2AuthorizationBounce+"="+cookieValueOAuth2AuthorizationBounce)

			handler(ctx)

			assert.Equal(t, fasthttp.StatusSeeOther, ctx.Response.StatusCode())
			assert.NotContains(t, string(ctx.Response.Body()), "Continue")
			assert.Equal(t, [][]byte{[]byte("default-src 'none'")}, ctx.Response.Header.PeekAll(fasthttp.HeaderContentSecurityPolicy))

			cookies = ctx.Response.Header.PeekAll(fasthttp.HeaderSetCookie)

			require.Len(t, cookies, 1)

			assert.Equal(t, "authelia_authorization_bounce=1; Path="+path+"; Max-Age=0; HttpOnly; Secure; SameSite=Lax", string(cookies[0]))
		})
	}
}

func newTestOAuth2AuthorizationBounceRequest(method, query string, form url.Values, headers, cookies map[string]string) (rw *httptest.ResponseRecorder, r *http.Request) {
	var body io.Reader

	if method == fasthttp.MethodPost && form != nil {
		body = strings.NewReader(form.Encode())
	}

	uri := testOIDCAuthorizationEndpoint

	if query != "" {
		uri += "?" + query
	}

	r = httptest.NewRequest(method, uri, body)

	if body != nil {
		r.Header.Set(fasthttp.HeaderContentType, "application/x-www-form-urlencoded")
	}

	for name, value := range headers {
		r.Header.Set(name, value)
	}

	pairs := make([]string, 0, len(cookies))

	for name, value := range cookies {
		pairs = append(pairs, name+"="+value)
	}

	if len(pairs) != 0 {
		r.Header.Set(fasthttp.HeaderCookie, strings.Join(pairs, "; "))
	}

	return httptest.NewRecorder(), r
}

func newTestOAuth2AuthorizationBounceRequestCtx(path string, values url.Values, cookie string) (ctx *fasthttp.RequestCtx) {
	ctx = &fasthttp.RequestCtx{}

	ctx.Request.Header.SetMethod(fasthttp.MethodGet)
	ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "login.example.com:8080")
	ctx.Request.Header.Set(headerSecFetchSite, valueSecFetchSiteCrossSite)
	ctx.Request.Header.Set(headerSecFetchMode, valueSecFetchModeNavigate)
	ctx.Request.Header.Set(headerSecFetchDest, valueSecFetchDestDocument)

	if cookie != "" {
		// A request carrying the marker models the request made by the bounce page, which is same-origin.
		ctx.Request.Header.Set(headerSecFetchSite, "same-origin")
		ctx.Request.Header.Set(fasthttp.HeaderCookie, cookie)
	}

	ctx.Request.SetRequestURI(path + "?" + values.Encode())

	return ctx
}

func newTestOAuth2AuthorizationBounceHeaders() (headers map[string]string) {
	return map[string]string{
		headerSecFetchSite: valueSecFetchSiteCrossSite,
		headerSecFetchMode: valueSecFetchModeNavigate,
		headerSecFetchDest: valueSecFetchDestDocument,
	}
}

func newTestOAuth2AuthorizationBounceFollowUpHeaders() (headers map[string]string) {
	return map[string]string{
		headerSecFetchSite: "same-origin",
		headerSecFetchMode: valueSecFetchModeNavigate,
		headerSecFetchDest: valueSecFetchDestDocument,
	}
}

func assertTestOAuth2AuthorizationBounceCookie(t *testing.T, rw *httptest.ResponseRecorder, path string) {
	t.Helper()

	cookies := rw.Result().Cookies()

	require.Len(t, cookies, 1)

	assert.Equal(t, "authelia_authorization_bounce", cookies[0].Name)
	assert.Equal(t, "1", cookies[0].Value)
	assert.Equal(t, path, cookies[0].Path)
	assert.Empty(t, cookies[0].Domain)
	assert.Equal(t, 60, cookies[0].MaxAge)
	assert.True(t, cookies[0].Secure)
	assert.True(t, cookies[0].HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
}

func assertTestOAuth2AuthorizationBounceRequestIntact(t *testing.T, r *http.Request, body string) {
	t.Helper()

	assert.Nil(t, r.Form)
	assert.Nil(t, r.PostForm)
	assert.Nil(t, r.MultipartForm)

	remaining, err := io.ReadAll(r.Body)

	require.NoError(t, err)

	assert.Equal(t, body, string(remaining))
}
