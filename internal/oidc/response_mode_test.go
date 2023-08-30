package oidc_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/ory/fosite"
	"github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestOpenIDConnectProvider_ResponseModeHandler(t *testing.T) {
	testCases := []struct {
		name     string
		have     fosite.ResponseModeHandler
		expected any
	}{
		{
			"ShouldReturnDefaultFosite",
			nil,
			&fosite.DefaultResponseModeHandler{},
		},
		{
			"ShouldReturnInternal",
			&oidc.ResponseModeHandler{},
			&oidc.ResponseModeHandler{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{Handlers: oidc.HandlersConfig{ResponseMode: tc.have}}

			provider := &oidc.OpenIDConnectProvider{Config: config}

			actual := provider.ResponseModeHandler(context.TODO())

			assert.IsType(t, tc.expected, actual)
		})
	}
}

func TestOpenIDConnectProvider_WriteAuthorizeResponse(t *testing.T) {
	testCases := []struct {
		name       string
		requester  fosite.AuthorizeRequester
		responder  fosite.AuthorizeResponder
		setup      func(t *testing.T, config *oidc.Config)
		code       int
		header     http.Header
		headerFunc func(t *testing.T, header http.Header)
		body       string
		bodyRegexp *regexp.Regexp
	}{
		{
			"ShouldHandleResponseModeQuery",
			&fosite.AuthorizeRequest{
				ResponseMode: oidc.ResponseModeQuery,
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			nil,
			fasthttp.StatusSeeOther,
			http.Header{fasthttp.HeaderLocation: []string{"https://app.example.com/callback?code=1234&iss=https%3A%2F%2Fauth.example.com"}},
			nil,
			"",
			nil,
		},
		{
			"ShouldWriteHeaders",
			&fosite.AuthorizeRequest{
				ResponseMode: oidc.ResponseModeQuery,
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Header: http.Header{
					fasthttp.HeaderAccept: []string{"123"},
				},
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			nil,
			fasthttp.StatusSeeOther,
			http.Header{
				fasthttp.HeaderLocation: []string{"https://app.example.com/callback?code=1234&iss=https%3A%2F%2Fauth.example.com"},
				fasthttp.HeaderAccept:   []string{"123"},
			},
			nil,
			"",
			nil,
		},
		{
			"ShouldHandleBadClient",
			&fosite.AuthorizeRequest{
				ResponseMode: oidc.ResponseModeQuery,
				Request: fosite.Request{
					Client: &fosite.DefaultClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Header: http.Header{
					fasthttp.HeaderAccept: []string{"123"},
				},
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			nil,
			fasthttp.StatusInternalServerError,
			http.Header{
				fasthttp.HeaderContentType: []string{"application/json; charset=utf-8"},
				fasthttp.HeaderAccept:      []string{"123"},
			},
			nil,
			"{\"error\":\"server_error\",\"error_description\":\"The authorization server encountered an unexpected condition that prevented it from fulfilling the request.\"}",
			nil,
		},
		{
			"ShouldHandleResponseModeQueryWithExistingQuery",
			&fosite.AuthorizeRequest{
				ResponseMode: oidc.ResponseModeQuery,
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback?abc=true",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback", RawQuery: "abc=true"},
			},
			&fosite.AuthorizeResponse{
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			nil,
			fasthttp.StatusSeeOther,
			http.Header{fasthttp.HeaderLocation: []string{"https://app.example.com/callback?abc=true&code=1234&iss=https%3A%2F%2Fauth.example.com"}},
			nil,
			"",
			nil,
		},
		{
			"ShouldHandleResponseModeFragment",
			&fosite.AuthorizeRequest{
				ResponseMode: oidc.ResponseModeFragment,
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			nil,
			fasthttp.StatusSeeOther,
			http.Header{fasthttp.HeaderLocation: []string{"https://app.example.com/callback#code=1234&iss=https%3A%2F%2Fauth.example.com"}},
			nil,
			"",
			nil,
		},
		{
			"ShouldHandleResponseModeFormPost",
			&fosite.AuthorizeRequest{
				ResponseMode: oidc.ResponseModeFormPost,
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			nil,
			fasthttp.StatusOK,
			http.Header{fasthttp.HeaderContentType: []string{"text/html; charset=utf-8"}},
			nil,
			"<!DOCTYPE html>\n<html lang=\"en\">\n\t<head>\n\t\t<title>Submit This Form</title>\n\t\t<script type=\"text/javascript\">\n\t\t\twindow.onload = function() {\n\t\t\t\tdocument.forms[0].submit();\n\t\t\t};\n\t\t</script>\n\t</head>\n\t<body>\n\t\t<form method=\"post\" action=\"https://app.example.com/callback\">\n\t\t\t\n\t\t\t\n\t\t\t<input type=\"hidden\" name=\"code\" value=\"1234\"/>\n\t\t\t\n\t\t\t\n\t\t\t\n\t\t\t<input type=\"hidden\" name=\"iss\" value=\"https://auth.example.com\"/>\n\t\t\t\n\t\t\t\n\t\t</form>\n\t</body>\n</html>\n",
			nil,
		},
		{
			"ShouldReturnEncoderErrorResponseModeJWT",
			&fosite.AuthorizeRequest{
				ResponseMode:  oidc.ResponseModeJWT,
				ResponseTypes: fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			func(t *testing.T, config *oidc.Config) {
				config.Signer = nil
			},
			fasthttp.StatusInternalServerError,
			http.Header{fasthttp.HeaderContentType: []string{"application/json; charset=utf-8"}},
			nil,
			"{\"error\":\"server_error\",\"error_description\":\"The authorization server encountered an unexpected condition that prevented it from fulfilling the request.\"}",
			nil,
		},
		{
			"ShouldReturnEncoderErrorResponseModeFormPostJWT",
			&fosite.AuthorizeRequest{
				ResponseMode:  oidc.ResponseModeFormPostJWT,
				ResponseTypes: fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			func(t *testing.T, config *oidc.Config) {
				config.Signer = nil
			},
			fasthttp.StatusInternalServerError,
			http.Header{fasthttp.HeaderContentType: []string{"application/json; charset=utf-8"}},
			nil,
			"{\"error\":\"server_error\",\"error_description\":\"The authorization server encountered an unexpected condition that prevented it from fulfilling the request.\"}",
			nil,
		},
		{
			"ShouldReturnEncoderErrorResponseModeFragmentJWT",
			&fosite.AuthorizeRequest{
				ResponseMode:  oidc.ResponseModeFragmentJWT,
				ResponseTypes: fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			func(t *testing.T, config *oidc.Config) {
				config.Signer = nil
			},
			fasthttp.StatusInternalServerError,
			http.Header{fasthttp.HeaderContentType: []string{"application/json; charset=utf-8"}},
			nil,
			"{\"error\":\"server_error\",\"error_description\":\"The authorization server encountered an unexpected condition that prevented it from fulfilling the request.\"}",
			nil,
		},
		{
			"ShouldEncodeJWTResponseModeJWTResponseTypesCode",
			&fosite.AuthorizeRequest{
				ResponseMode:  oidc.ResponseModeJWT,
				ResponseTypes: fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			nil,
			fasthttp.StatusSeeOther,
			nil,
			func(t *testing.T, header http.Header) {
				uri, err := url.ParseRequestURI(header.Get(fasthttp.HeaderLocation))
				assert.NoError(t, err)

				require.NotNil(t, uri)

				assert.Equal(t, "https", uri.Scheme)
				assert.Equal(t, "app.example.com", uri.Host)
				assert.Equal(t, "/callback", uri.Path)
				assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+$`), uri.Query().Get(oidc.FormParameterResponse))
			},
			"",
			nil,
		},
		{
			"ShouldEncodeJWTResponseModeJWTResponseTypesNotCode",
			&fosite.AuthorizeRequest{
				ResponseMode:  oidc.ResponseModeJWT,
				ResponseTypes: fosite.Arguments{oidc.ResponseTypeImplicitFlowBoth},
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			nil,
			fasthttp.StatusSeeOther,
			nil,
			func(t *testing.T, header http.Header) {
				uri, err := url.Parse(header.Get(fasthttp.HeaderLocation))
				assert.NoError(t, err)

				require.NotNil(t, uri)

				assert.Equal(t, "https", uri.Scheme)
				assert.Equal(t, "app.example.com", uri.Host)
				assert.Equal(t, "/callback", uri.Path)
				assert.Regexp(t, regexp.MustCompile(`^response=[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+$`), uri.Fragment)
			},
			"",
			nil,
		},
		{
			"ShouldEncodeJWTResponseModeFormPost",
			&fosite.AuthorizeRequest{
				ResponseMode:  oidc.ResponseModeFormPostJWT,
				ResponseTypes: fosite.Arguments{oidc.ResponseTypeAuthorizationCodeFlow},
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			&fosite.AuthorizeResponse{
				Parameters: url.Values{
					oidc.FormParameterAuthorizationCode: []string{"1234"},
					oidc.FormParameterIssuer:            []string{"https://auth.example.com"},
				},
			},
			nil,
			fasthttp.StatusOK,
			http.Header{fasthttp.HeaderContentType: []string{"text/html; charset=utf-8"}},
			nil,
			"",
			regexp.MustCompile(`<input type="hidden" name="response" value="[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+"/>`),
		},
	}

	tp, err := templates.New(templates.Config{})

	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{
				Signer: &jwt.DefaultSigner{
					GetPrivateKey: func(ctx context.Context) (interface{}, error) {
						return keyRSA2048, nil
					},
				},
				Templates: tp,
				Issuers: oidc.IssuersConfig{
					AuthorizationServerIssuerIdentification: "https://auth.example.com",
					JWTSecuredResponseMode:                  "https://auth.example.com",
				},
			}

			mock := httptest.NewRecorder()

			handler := &oidc.ResponseModeHandler{
				Config: config,
			}

			config.Handlers.ResponseMode = handler

			provider := &oidc.OpenIDConnectProvider{Config: config}

			if tc.setup != nil {
				tc.setup(t, config)
			}

			provider.WriteAuthorizeResponse(context.TODO(), mock, tc.requester, tc.responder)

			result := mock.Result()

			assert.Equal(t, tc.code, result.StatusCode)

			if tc.header != nil {
				assert.Equal(t, tc.header, result.Header)
			}

			if tc.headerFunc != nil {
				tc.headerFunc(t, result.Header)
			}

			data, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			if tc.bodyRegexp == nil {
				assert.Equal(t, tc.body, string(data))
			} else {
				assert.Regexp(t, tc.bodyRegexp, string(data))
			}
		})
	}
}

func TestOpenIDConnectProvider_WriteAuthorizeError(t *testing.T) {
	testCases := []struct {
		name       string
		requester  fosite.AuthorizeRequester
		setup      func(t *testing.T, config *oidc.Config)
		error      error
		code       int
		header     http.Header
		headerFunc func(t *testing.T, header http.Header)
		body       string
		bodyRegexp *regexp.Regexp
	}{
		{
			"ShouldHandleErrorResponse",
			&fosite.AuthorizeRequest{
				ResponseMode: oidc.ResponseModeQuery,
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			nil,
			fosite.ErrServerError.WithDebug("The Debug."),
			fasthttp.StatusSeeOther,
			http.Header{
				fasthttp.HeaderLocation: []string{"https://app.example.com/callback?error=server_error&error_description=The+authorization+server+encountered+an+unexpected+condition+that+prevented+it+from+fulfilling+the+request.&iss=https%3A%2F%2Fauth.example.com"},
			},
			nil,
			"",
			nil,
		},
		{
			"ShouldHandleErrorResponseWithState",
			&fosite.AuthorizeRequest{
				ResponseMode: oidc.ResponseModeQuery,
				State:        "abc123state",
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/callback"},
			},
			nil,
			fosite.ErrServerError.WithDebug("The Debug."),
			fasthttp.StatusSeeOther,
			http.Header{
				fasthttp.HeaderLocation: []string{"https://app.example.com/callback?error=server_error&error_description=The+authorization+server+encountered+an+unexpected+condition+that+prevented+it+from+fulfilling+the+request.&iss=https%3A%2F%2Fauth.example.com&state=abc123state"},
			},
			nil,
			"",
			nil,
		},
		{
			"ShouldHandleErrorResponseWithInvalidRedirectURI",
			&fosite.AuthorizeRequest{
				ResponseMode: oidc.ResponseModeQuery,
				State:        "abc123state",
				Request: fosite.Request{
					Client: &oidc.BaseClient{
						ID: "example",
						RedirectURIs: []string{
							"https://app.example.com/callback",
						},
					},
				},
				RedirectURI: &url.URL{Scheme: "https", Host: "app.example.com", Path: "/invalid"},
			},
			nil,
			fosite.ErrServerError.WithDebug("The Debug."),
			fasthttp.StatusInternalServerError,
			http.Header{
				fasthttp.HeaderContentType: []string{"application/json; charset=utf-8"},
			},
			nil,
			"{\"error\":\"server_error\",\"error_description\":\"The authorization server encountered an unexpected condition that prevented it from fulfilling the request.\"}",
			nil,
		},
	}

	tp, err := templates.New(templates.Config{})

	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &oidc.Config{
				Signer: &jwt.DefaultSigner{
					GetPrivateKey: func(ctx context.Context) (interface{}, error) {
						return keyRSA2048, nil
					},
				},
				Templates: tp,
				Issuers: oidc.IssuersConfig{
					AuthorizationServerIssuerIdentification: "https://auth.example.com",
					JWTSecuredResponseMode:                  "https://auth.example.com",
				},
			}

			mock := httptest.NewRecorder()

			handler := &oidc.ResponseModeHandler{
				Config: config,
			}

			config.Handlers.ResponseMode = handler

			if tc.setup != nil {
				tc.setup(t, config)
			}

			handler.WriteAuthorizeError(context.TODO(), mock, tc.requester, tc.error)

			result := mock.Result()

			assert.Equal(t, tc.code, result.StatusCode)

			if tc.header != nil {
				assert.Equal(t, tc.header, result.Header)
			}

			if tc.headerFunc != nil {
				tc.headerFunc(t, result.Header)
			}

			data, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			if tc.bodyRegexp == nil {
				assert.Equal(t, tc.body, string(data))
			} else {
				assert.Regexp(t, tc.bodyRegexp, string(data))
			}
		})
	}
}
