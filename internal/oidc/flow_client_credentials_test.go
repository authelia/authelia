package oidc_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

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
