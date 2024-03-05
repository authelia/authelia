package oidc_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestOpenIDConnectProvider_NewIntrospectionRequest(t *testing.T) {
	testCases := []struct {
		name       string
		clients    []schema.IdentityProvidersOpenIDConnectClient
		setup      func(ctx gomock.Matcher, provider *oidc.OpenIDConnectProvider, mock *mocks.MockTokenIntrospector)
		req        *http.Request
		expected   fosite.TokenUse
		expectedtt string
		err        string
	}{
		{
			"ShouldNotIntrospectTokenWithoutCredentials",
			nil,
			nil,
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-introspection-token"},
				},
			},
			fosite.AccessToken,
			fosite.BearerAccessToken,
			"The request could not be authorized. HTTP Authorization header missing.",
		},
		{
			"ShouldIntrospectAccessTokenWithBearerCredentials",
			nil,
			func(ctx gomock.Matcher, provider *oidc.OpenIDConnectProvider, mock *mocks.MockTokenIntrospector) {
				mock.EXPECT().IntrospectToken(ctx, "arbitrary-bearer-token", gomock.Any(), gomock.Any(), gomock.Any()).Return(fosite.AccessToken, nil)
				mock.EXPECT().IntrospectToken(ctx, "arbitrary-introspection-token", gomock.Any(), gomock.Any(), gomock.Any()).Return(fosite.AccessToken, nil)
			},
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Bearer arbitrary-bearer-token"},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-introspection-token"},
				},
			},
			fosite.AccessToken,
			fosite.BearerAccessToken,
			"",
		},
		{
			"ShouldIntrospectAccessTokenWithBasicCredentials",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:     "client_id",
					Secret: MustDecodeSecret("$plaintext$client_secret"),
				},
			},
			func(ctx gomock.Matcher, provider *oidc.OpenIDConnectProvider, mock *mocks.MockTokenIntrospector) {
				mock.EXPECT().IntrospectToken(ctx, "arbitrary-introspection-token", gomock.Any(), gomock.Any(), gomock.Any()).Return(fosite.AccessToken, nil)
			},
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Basic Y2xpZW50X2lkOmNsaWVudF9zZWNyZXQ="},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-introspection-token"},
				},
			},
			fosite.AccessToken,
			fosite.BearerAccessToken,
			"",
		},
		{
			"ShouldNotIntrospectAccessTokenWithBasicCredentialsInvalidClientSecret",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:     "client_id",
					Secret: MustDecodeSecret("$plaintext$client_secret2"),
				},
			},
			nil,
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Basic Y2xpZW50X2lkOmNsaWVudF9zZWNyZXQ="},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-introspection-token"},
				},
			},
			fosite.AccessToken,
			fosite.BearerAccessToken,
			"The request could not be authorized. OAuth 2.0 Client credentials are invalid.",
		},
		{
			"ShouldNotIntrospectAccessTokenWithBasicCredentialsInvalidClientID",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:     "client_id2",
					Secret: MustDecodeSecret("$plaintext$client_secret"),
				},
			},
			nil,
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Basic Y2xpZW50X2lkOmNsaWVudF9zZWNyZXQ="},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-introspection-token"},
				},
			},
			fosite.AccessToken,
			fosite.BearerAccessToken,
			"The request could not be authorized. Unable to find OAuth 2.0 Client from HTTP basic authorization header. Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method). Client with id 'client_id' does not appear to be a registered client.",
		},
		{
			"ShouldNotIntrospectAccessTokenWithBasicCredentialsInvalidAuthorizationHeader",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:     "client_id2",
					Secret: MustDecodeSecret("$plaintext$client_secret"),
				},
			},
			nil,
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"x"},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-introspection-token"},
				},
			},
			fosite.AccessToken,
			fosite.BearerAccessToken,
			"The request could not be authorized. HTTP Authorization header malformed.",
		},
		{
			"ShouldNotIntrospectAccessTokenWithBearerCredentialsHasError",
			nil,
			func(ctx gomock.Matcher, provider *oidc.OpenIDConnectProvider, mock *mocks.MockTokenIntrospector) {
				mock.EXPECT().IntrospectToken(ctx, "arbitrary-bearer-token", gomock.Any(), gomock.Any(), gomock.Any()).Return(fosite.TokenType(""), fosite.ErrNotFound)
			},
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Bearer arbitrary-bearer-token"},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-introspection-token"},
				},
			},
			fosite.AccessToken,
			fosite.BearerAccessToken,
			"The request could not be authorized. HTTP Authorization header missing, malformed, or credentials used are invalid. Could not find the requested resource(s).",
		},
		{
			"ShouldIntrospectRefreshToken",
			nil,
			func(ctx gomock.Matcher, provider *oidc.OpenIDConnectProvider, mock *mocks.MockTokenIntrospector) {
				mock.EXPECT().IntrospectToken(ctx, "arbitrary-bearer-token", gomock.Any(), gomock.Any(), gomock.Any()).Return(fosite.AccessToken, nil)
				mock.EXPECT().IntrospectToken(ctx, "arbitrary-introspection-token", gomock.Any(), gomock.Any(), gomock.Any()).Return(fosite.RefreshToken, nil)
			},
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Bearer arbitrary-bearer-token"},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-introspection-token"},
				},
			},
			fosite.RefreshToken,
			"",
			"",
		},
		{
			"ShouldNotIntrospectWhenBearerRefreshToken",
			nil,
			func(ctx gomock.Matcher, provider *oidc.OpenIDConnectProvider, mock *mocks.MockTokenIntrospector) {
				mock.EXPECT().IntrospectToken(ctx, "arbitrary-bearer-token", gomock.Any(), gomock.Any(), gomock.Any()).Return(fosite.RefreshToken, nil)
			},
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Bearer arbitrary-bearer-token"},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-introspection-token"},
				},
			},
			fosite.RefreshToken,
			"",
			"The request could not be authorized. HTTP Authorization header did not provide a token of type 'access_token', got type 'refresh_token'.",
		},
		{
			"ShouldNotIntrospectWhenIdenticalTokens",
			nil,
			nil,
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Bearer arbitrary-bearer-token"},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-bearer-token"},
				},
			},
			fosite.RefreshToken,
			"",
			"The request could not be authorized. Bearer and introspection token are identical.",
		},
		{
			"ShouldNotIntrospectInvalidMethodVerb",
			nil,
			nil,
			&http.Request{
				Method: fasthttp.MethodGet,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Bearer arbitrary-bearer-token"},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-bearer-token"},
				},
			},
			fosite.RefreshToken,
			"",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. HTTP method is 'GET' but expected 'POST'.",
		},
		{
			"ShouldNotIntrospectEmptyPost",
			nil,
			nil,
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Bearer arbitrary-bearer-token"},
				},
				PostForm: nil,
			},
			fosite.RefreshToken,
			"",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. The POST body can not be empty.",
		},
		{
			"ShouldNotIntrospectCorruptMultiPartFormDataPost",
			nil,
			nil,
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Bearer arbitrary-bearer-token"},
					fasthttp.HeaderContentType:   []string{"multipart/form-data"},
					fasthttp.HeaderContentLength: []string{"3"},
				},
				Body:     io.NopCloser(bytes.NewReader([]byte{0x01, 0x00, 0x02})),
				PostForm: nil,
			},
			fosite.RefreshToken,
			"",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Unable to parse HTTP body, make sure to send a properly formatted form request body. no multipart boundary param in Content-Type",
		},
		{
			"ShouldReturnIntrospectionTokenError",
			nil,
			func(ctx gomock.Matcher, provider *oidc.OpenIDConnectProvider, mock *mocks.MockTokenIntrospector) {
				mock.EXPECT().IntrospectToken(ctx, "arbitrary-bearer-token", gomock.Any(), gomock.Any(), gomock.Any()).Return(fosite.AccessToken, nil)
				mock.EXPECT().IntrospectToken(ctx, "arbitrary-introspection-token", gomock.Any(), gomock.Any(), gomock.Any()).Return(fosite.TokenType(""), fosite.ErrNotFound)
			},
			&http.Request{
				Method: fasthttp.MethodPost,
				Header: http.Header{
					fasthttp.HeaderAuthorization: []string{"Bearer arbitrary-bearer-token"},
				},
				PostForm: url.Values{
					oidc.FormParameterToken: []string{"arbitrary-introspection-token"},
				},
			},
			fosite.AccessToken,
			fosite.BearerAccessToken,
			"Token is inactive because it is malformed, expired or otherwise invalid. An introspection strategy indicated that the token is inactive. Could not find the requested resource(s).",
		},
	}

	tp, err := templates.New(templates.Config{})

	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			defer ctrl.Finish()

			mock := mocks.NewMockTokenIntrospector(ctrl)

			store := mocks.NewMockStorage(ctrl)

			ctx := gomock.AssignableToTypeOf(context.WithValue(context.TODO(), fosite.ContextKey("test"), nil))

			provider := oidc.NewOpenIDConnectProvider(&schema.IdentityProvidersOpenIDConnect{HMACSecret: badhmac, Clients: tc.clients}, store, tp)

			provider.Config.Handlers.TokenIntrospection = fosite.TokenIntrospectionHandlers{mock}

			if tc.setup != nil {
				tc.setup(ctx, provider, mock)
			}

			responder, err := provider.NewIntrospectionRequest(context.TODO(), tc.req, oidc.NewSession())

			if len(tc.err) == 0 {
				assert.NoError(t, oidc.ErrorToDebugRFC6749Error(err))
				require.NotNil(t, responder)

				assert.Equal(t, tc.expected, responder.GetTokenUse())
				assert.Equal(t, tc.expectedtt, responder.GetAccessTokenType())
			} else {
				assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.err)
			}
		})
	}
}

func TestIntrospectionResponse(t *testing.T) {
	testCases := []struct {
		name            string
		have            *oidc.IntrospectionResponse
		clientID        string
		active          bool
		accessRequestID string
	}{
		{
			"ShouldTestActiveToken",
			&oidc.IntrospectionResponse{
				Client: &oidc.BaseClient{ID: "client1"},
				Active: true,
				AccessRequester: &fosite.AccessRequest{
					Request: fosite.Request{
						ID: abc,
					},
				},
			},
			"client1",
			true,
			abc,
		},
		{
			"ShouldReturnNilBadResponse",
			&oidc.IntrospectionResponse{
				Client:          nil,
				Active:          false,
				AccessRequester: nil,
			},
			"",
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.clientID == "" {
				assert.Nil(t, tc.have.GetClient())
			} else {
				assert.Equal(t, tc.clientID, tc.have.GetClient().GetID())
			}

			if tc.accessRequestID == "" {
				assert.Nil(t, tc.have.GetAccessRequester())
			} else {
				assert.Equal(t, tc.accessRequestID, tc.have.GetAccessRequester().GetID())
			}

			assert.Equal(t, tc.active, tc.have.IsActive())
		})
	}
}
