package oidc_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/openid"
	fjwt "github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestClientCredentialsGrantHandler_CanSkipClientAuth(t *testing.T) {
	handler := oidc.ClientCredentialsGrantHandler{}

	assert.False(t, handler.CanSkipClientAuth(context.TODO(), &fosite.AccessRequest{}))
}

func TestClientCredentialsGrantHandler_HandleTokenEndpointRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockClientCredentialsGrantStorage(ctrl)
	chgen := mocks.NewMockAccessTokenStrategy(ctrl)

	defer ctrl.Finish()

	handler := oidc.ClientCredentialsGrantHandler{
		HandleHelper: &oauth2.HandleHelper{
			AccessTokenStorage:  store,
			AccessTokenStrategy: chgen,
			Config: &fosite.Config{
				AccessTokenLifespan: time.Hour,
			},
		},
		Config: &fosite.Config{
			ScopeStrategy:            fosite.HierarchicScopeStrategy,
			AudienceMatchingStrategy: fosite.DefaultAudienceMatchingStrategy,
		},
	}

	testCases := []struct {
		name     string
		setup    func(mock *mocks.MockAccessRequester)
		err      error
		expected string
	}{
		{
			name: "ShouldFailNotResponsible",
			setup: func(mock *mocks.MockAccessRequester) {
				mock.EXPECT().GetGrantTypes().Return(fosite.Arguments{""})
			},
			err:      fosite.ErrUnknownRequest,
			expected: "The handler is not responsible for this request.",
		},
		{
			name: "ShouldFailInvalidAudience",
			setup: func(mock *mocks.MockAccessRequester) {
				mock.EXPECT().GetGrantTypes().Return(fosite.Arguments{oidc.GrantTypeClientCredentials})
				mock.EXPECT().GetRequestedScopes().Return([]string{})
				mock.EXPECT().GetRequestedAudience().Return([]string{"https://www.ory.sh/not-api"})
				mock.EXPECT().GetClient().Return(&fosite.DefaultClient{
					GrantTypes: fosite.Arguments{oidc.GrantTypeClientCredentials},
					Audience:   []string{"https://www.ory.sh/api"},
				})
			},
			err:      fosite.ErrInvalidRequest,
			expected: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Requested audience 'https://www.ory.sh/not-api' has not been whitelisted by the OAuth 2.0 Client.",
		},
		{
			name: "ShouldFailInvalidScope",
			setup: func(mock *mocks.MockAccessRequester) {
				mock.EXPECT().GetGrantTypes().Return(fosite.Arguments{oidc.GrantTypeClientCredentials})
				mock.EXPECT().GetRequestedScopes().Return([]string{"foo", "bar", "baz.bar"})
				mock.EXPECT().GetClient().Return(&fosite.DefaultClient{
					GrantTypes: fosite.Arguments{oidc.GrantTypeClientCredentials},
					Scopes:     []string{"foo"},
				})
				mock.EXPECT().GrantScope("foo")
			},
			err:      fosite.ErrInvalidScope,
			expected: "The requested scope is invalid, unknown, or malformed. The OAuth 2.0 Client is not allowed to request scope 'bar'.",
		},
		{
			name: "ShouldPass",
			setup: func(mock *mocks.MockAccessRequester) {
				mock.EXPECT().GetSession().Return(new(fosite.DefaultSession))
				mock.EXPECT().GetGrantTypes().Return(fosite.Arguments{oidc.GrantTypeClientCredentials})
				mock.EXPECT().GetRequestedScopes().Return([]string{"foo", "bar", "baz.bar"})
				mock.EXPECT().GetRequestedAudience().Return([]string{})
				mock.EXPECT().GetClient().Return(&fosite.DefaultClient{
					GrantTypes: fosite.Arguments{oidc.GrantTypeClientCredentials},
					Scopes:     []string{"foo", "bar", "baz"},
				})

				mock.EXPECT().GrantScope("foo")
				mock.EXPECT().GrantScope("bar")
				mock.EXPECT().GrantScope("baz.bar")
			},
		},
		{
			name: "ShouldFailPublicClient",
			setup: func(mock *mocks.MockAccessRequester) {
				mock.EXPECT().GetGrantTypes().Return(fosite.Arguments{oidc.GrantTypeClientCredentials})
				mock.EXPECT().GetClient().Return(&fosite.DefaultClient{
					GrantTypes: fosite.Arguments{oidc.GrantTypeClientCredentials},
					Scopes:     []string{"foo", "bar", "baz"},
					Public:     true,
				})
			},
			err:      fosite.ErrInvalidGrant,
			expected: "The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client. The OAuth 2.0 Client is marked as public and is thus not allowed to use authorization grant 'client_credentials'.",
		},
		{
			name: "ShouldPassBaseClient",
			setup: func(mock *mocks.MockAccessRequester) {
				mock.EXPECT().GetSession().Return(new(fosite.DefaultSession))
				mock.EXPECT().GetGrantTypes().Return(fosite.Arguments{oidc.GrantTypeClientCredentials})
				mock.EXPECT().GetRequestedScopes().Return([]string{"foo", "bar", "baz.bar"})
				mock.EXPECT().GetRequestedAudience().Return([]string{})
				mock.EXPECT().GetClient().Return(&oidc.BaseClient{
					GrantTypes: fosite.Arguments{oidc.GrantTypeClientCredentials},
					Scopes:     []string{"foo", "bar", "baz"},
				})

				mock.EXPECT().GrantScope("foo")
				mock.EXPECT().GrantScope("bar")
				mock.EXPECT().GrantScope("baz.bar")
			},
		},
		{
			name: "ShouldPassBaseNoScopes",
			setup: func(mock *mocks.MockAccessRequester) {
				mock.EXPECT().GetSession().Return(new(fosite.DefaultSession))
				mock.EXPECT().GetGrantTypes().Return(fosite.Arguments{oidc.GrantTypeClientCredentials})
				mock.EXPECT().GetRequestedScopes().Return([]string{})
				mock.EXPECT().GetRequestedAudience().Return([]string{})
				mock.EXPECT().GetClient().Return(&oidc.BaseClient{
					GrantTypes: fosite.Arguments{oidc.GrantTypeClientCredentials},
					Scopes:     []string{"foo", "bar", "baz"},
				})

				mock.EXPECT().GrantScope("foo")
				mock.EXPECT().GrantScope("bar")
				mock.EXPECT().GrantScope("baz")
			},
		},
		{
			name: "ShouldPassBaseNoScopesWithAuto",
			setup: func(mock *mocks.MockAccessRequester) {
				mock.EXPECT().GetSession().Return(new(fosite.DefaultSession))
				mock.EXPECT().GetGrantTypes().Return(fosite.Arguments{oidc.GrantTypeClientCredentials})
				mock.EXPECT().GetRequestedScopes().Return([]string{})
				mock.EXPECT().GetRequestedAudience().Return([]string{})
				mock.EXPECT().GetClient().Return(&oidc.BaseClient{
					GrantTypes: fosite.Arguments{oidc.GrantTypeClientCredentials},
					Scopes:     []string{"foo", "bar", "baz"},
				})

				mock.EXPECT().GrantScope("foo")
				mock.EXPECT().GrantScope("bar")
				mock.EXPECT().GrantScope("baz")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAccessRequester(ctrl)

			tc.setup(mock)

			err := handler.HandleTokenEndpointRequest(context.Background(), mock)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.expected)
			} else {
				assert.NoError(t, oidc.ErrorToDebugRFC6749Error(err))
			}
		})
	}
}

func TestClientCredentialsGrantHandler_PopulateTokenEndpointResponse(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(request *fosite.AccessRequest, store *mocks.MockClientCredentialsGrantStorage, strategy *mocks.MockAccessTokenStrategy)
		req      *http.Request
		err      error
		expected string
	}{
		{
			name: "ShouldFailNotResponsible",
			setup: func(request *fosite.AccessRequest, store *mocks.MockClientCredentialsGrantStorage, strategy *mocks.MockAccessTokenStrategy) {
				request.GrantTypes = fosite.Arguments{""}
			},
			err:      fosite.ErrUnknownRequest,
			expected: "The handler is not responsible for this request.",
		},
		{
			name: "ShouldFailGrantTypeNotAllowed",
			setup: func(request *fosite.AccessRequest, store *mocks.MockClientCredentialsGrantStorage, strategy *mocks.MockAccessTokenStrategy) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeClientCredentials}
				request.Client = &fosite.DefaultClient{GrantTypes: fosite.Arguments{oidc.GrantTypeAuthorizationCode}}
			},
			err:      fosite.ErrUnauthorizedClient,
			expected: "The client is not authorized to request a token using this method. The OAuth 2.0 Client is not allowed to use authorization grant 'client_credentials'.",
		},
		{
			name: "ShouldPass",
			setup: func(request *fosite.AccessRequest, store *mocks.MockClientCredentialsGrantStorage, strategy *mocks.MockAccessTokenStrategy) {
				request.GrantTypes = fosite.Arguments{oidc.GrantTypeClientCredentials}
				request.Session = &fosite.DefaultSession{}
				request.Client = &fosite.DefaultClient{GrantTypes: fosite.Arguments{oidc.GrantTypeClientCredentials}}
				strategy.EXPECT().GenerateAccessToken(gomock.Any(), request).Return("tokenfoo.bar", "bar", nil)
				store.EXPECT().CreateAccessTokenSession(gomock.Any(), "bar", gomock.Eq(request.Sanitize([]string{}))).Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mocks.NewMockClientCredentialsGrantStorage(ctrl)
			strategy := mocks.NewMockAccessTokenStrategy(ctrl)
			request := fosite.NewAccessRequest(new(fosite.DefaultSession))
			response := fosite.NewAccessResponse()
			defer ctrl.Finish()

			handler := oidc.ClientCredentialsGrantHandler{
				HandleHelper: &oauth2.HandleHelper{
					AccessTokenStorage:  store,
					AccessTokenStrategy: strategy,
					Config: &fosite.Config{
						AccessTokenLifespan: time.Hour,
					},
				},
				Config: &fosite.Config{
					ScopeStrategy: fosite.HierarchicScopeStrategy,
				},
			}

			tc.setup(request, store, strategy)

			err := handler.PopulateTokenEndpointResponse(context.Background(), request, response)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.expected)
			} else {
				assert.NoError(t, oidc.ErrorToDebugRFC6749Error(err))
			}
		})
	}
}

func TestPopulateClientCredentialsFlowSessionWithAccessRequest(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(ctx oidc.Context)
		ctx      oidc.Context
		client   fosite.Client
		have     *oidc.Session
		expected *oidc.Session
		err      string
	}{
		{
			"ShouldHandleIssuerError",
			nil,
			&TestContext{
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return nil, errors.New("an error")
				},
			},
			nil,
			oidc.NewSession(),
			nil,
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Failed to determine the issuer with error: an error.",
		},
		{
			"ShouldHandleClientError",
			nil,
			&TestContext{
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return &url.URL{Scheme: "https", Host: "example.com"}, nil
				},
			},
			nil,
			oidc.NewSession(),
			nil,
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Failed to get the client for the request.",
		},
		{
			"ShouldUpdateValues",
			func(ctx oidc.Context) {
				c := ctx.(*TestContext)

				c.Clock = clock.NewFixed(time.Unix(10000000000, 0))
			},
			&TestContext{
				IssuerURLFunc: func() (issuerURL *url.URL, err error) {
					return &url.URL{Scheme: "https", Host: "example.com"}, nil
				},
			},
			&oidc.BaseClient{
				ID: abc,
			},
			oidc.NewSession(),
			&oidc.Session{
				Extra: map[string]any{},
				DefaultSession: &openid.DefaultSession{
					Headers: &fjwt.Headers{
						Extra: map[string]any{},
					},
					Claims: &fjwt.IDTokenClaims{
						Issuer:      "https://example.com",
						IssuedAt:    time.Unix(10000000000, 0).UTC(),
						RequestedAt: time.Unix(10000000000, 0).UTC(),
						Subject:     abc,
						Extra:       map[string]any{},
					},
				},
				ClientID:          abc,
				ClientCredentials: true,
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(tc.ctx)
			}

			err := oidc.PopulateClientCredentialsFlowSessionWithAccessRequest(tc.ctx, tc.client, tc.have)

			assert.Equal(t, "", tc.have.GetSubject())
			if len(tc.err) == 0 {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.expected, tc.have)
			} else {
				assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.err)
			}
		})
	}
}

func TestPopulateClientCredentialsFlowRequester(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(ctx oidc.Context)
		ctx      oidc.Context
		config   fosite.Configurator
		client   fosite.Client
		have     *fosite.Request
		expected *fosite.Request
		err      string
	}{
		{
			"ShouldHandleBasic",
			nil,
			&TestContext{},
			&oidc.Config{},
			&oidc.BaseClient{},
			&fosite.Request{},
			&fosite.Request{},
			"",
		},
		{
			"ShouldHandleNilErrorClient",
			nil,
			&TestContext{},
			&oidc.Config{},
			nil,
			&fosite.Request{},
			&fosite.Request{},
			"The authorization server encountered an unexpected condition that prevented it from fulfilling the request. Failed to get the client, configuration, or requester for the request.",
		},
		{
			"ShouldHandleBadScopeCombinationAuthz",
			nil,
			&TestContext{},
			&oidc.Config{},
			&oidc.BaseClient{ID: "abc", Scopes: []string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOpenID}},
			&fosite.Request{RequestedScope: fosite.Arguments{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOpenID}},
			&fosite.Request{},
			"The requested scope is invalid, unknown, or malformed. The scope 'authelia.bearer.authz' must only be requested by itself or with the 'offline_access' scope, no other scopes are permitted.",
		},
		{
			"ShouldHandleScopeNotPermitted",
			nil,
			&TestContext{},
			&oidc.Config{},
			&oidc.BaseClient{ID: "abc", Scopes: []string{oidc.ScopeAutheliaBearerAuthz}},
			&fosite.Request{RequestedScope: fosite.Arguments{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess}},
			&fosite.Request{},
			"The requested scope is invalid, unknown, or malformed. The scope 'offline_access' is not authorized on client with id 'abc'.",
		},
		{
			"ShouldHandleGoodScopesWithoutAudience",
			nil,
			&TestContext{},
			&oidc.Config{},
			&oidc.BaseClient{ID: "abc", Scopes: []string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess}},
			&fosite.Request{RequestedScope: fosite.Arguments{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess}},
			&fosite.Request{},
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified. The scope 'authelia.bearer.authz' requires the request also include an audience.",
		},
		{
			"ShouldHandleGoodScopesAndBadAudience",
			nil,
			&TestContext{},
			&oidc.Config{},
			&oidc.BaseClient{ID: "abc", Scopes: []string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess}},
			&fosite.Request{RequestedScope: fosite.Arguments{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess}, RequestedAudience: fosite.Arguments{"https://example.com"}},
			&fosite.Request{},
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Requested audience 'https://example.com' has not been whitelisted by the OAuth 2.0 Client.",
		},
		{
			"ShouldHandleGoodScopesAndAudience",
			nil,
			&TestContext{},
			&oidc.Config{},
			&oidc.BaseClient{ID: "abc", Scopes: []string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess}, Audience: fosite.Arguments{"https://example.com"}},
			&fosite.Request{
				RequestedScope:    fosite.Arguments{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess},
				RequestedAudience: fosite.Arguments{"https://example.com"},
			},
			&fosite.Request{
				RequestedScope:    fosite.Arguments{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess},
				GrantedScope:      fosite.Arguments{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess},
				RequestedAudience: fosite.Arguments{"https://example.com"},
				GrantedAudience:   fosite.Arguments{"https://example.com"},
			},
			"",
		},
		{
			"ShouldHandleGoodScopesAndAudienceSubSet",
			nil,
			&TestContext{},
			&oidc.Config{},
			&oidc.BaseClient{ID: "abc", Scopes: []string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess}, Audience: fosite.Arguments{"https://example.com", "https://app.example.com"}},
			&fosite.Request{
				RequestedScope:    fosite.Arguments{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess},
				RequestedAudience: fosite.Arguments{"https://example.com"},
			},
			&fosite.Request{
				RequestedScope:    fosite.Arguments{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess},
				GrantedScope:      fosite.Arguments{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess},
				RequestedAudience: fosite.Arguments{"https://example.com"},
				GrantedAudience:   fosite.Arguments{"https://example.com"},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(tc.ctx)
			}

			err := oidc.PopulateClientCredentialsFlowRequester(tc.ctx, tc.config, tc.client, tc.have)

			if len(tc.err) == 0 {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.expected, tc.have)
			} else {
				assert.EqualError(t, oidc.ErrorToDebugRFC6749Error(err), tc.err)
			}
		})
	}
}
