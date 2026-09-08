package handlers

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestHandleOAuth2TokenHydration(t *testing.T) {
	subject := uuid.MustParse("e79b6494-8852-4439-860c-159f2cba83dc")

	testCases := []struct {
		name    string
		session *oidc.Session
		handled bool
		setup   func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy)
		expect  func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session)
	}{
		{
			name:    "ShouldHydrateClientCredentialsFlow",
			session: oidc.NewSession(),
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy) {
				requester.EXPECT().GetGrantTypes().Return(oauthelia2.Arguments{oidc.GrantTypeClientCredentials}).AnyTimes()
				requester.EXPECT().GetRequestedScopes().Return(oauthelia2.Arguments{testOIDCScopeBearerAuthz}).AnyTimes()
				client.EXPECT().GetID().Return(testValue).AnyTimes()
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session) {
				assert.True(t, session.ClientCredentials)
				assert.Equal(t, testValue, session.ClientID)
				assert.Equal(t, testValue, session.Claims.Subject)
				assert.Empty(t, session.Subject)
			},
		},
		{
			name:    "ShouldHandleClientCredentialsHydrationError",
			session: oidc.NewSession(),
			handled: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy) {
				clearForwardedHeaders(mock)

				requester.EXPECT().GetGrantTypes().Return(oauthelia2.Arguments{oidc.GrantTypeClientCredentials}).AnyTimes()
				requester.EXPECT().GetID().Return(testValue).AnyTimes()
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Access Request encountered an error while trying to hydrate the Client Credentials Flow claims", regexpAnyError)
			},
		},
		{
			name:    "ShouldHandleClientCredentialsPopulateError",
			session: oidc.NewSession(),
			handled: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy) {
				requester.EXPECT().GetGrantTypes().Return(oauthelia2.Arguments{oidc.GrantTypeClientCredentials}).AnyTimes()
				requester.EXPECT().GetID().Return(testValue).AnyTimes()

				requester.EXPECT().GetRequestedScopes().Return(oauthelia2.Arguments{testOIDCScopeBearerAuthz, oidc.ScopeProfile}).AnyTimes()

				client.EXPECT().GetID().Return(testValue).AnyTimes()
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Access Request encountered an error while trying to populate the Client Credentials Flow requester", regexpAnyError)
			},
		},
		{
			name:    "ShouldSkipWhenJWTProfileDisabled",
			session: oidc.NewSession(),
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy) {
				requester.EXPECT().GetGrantTypes().Return(oauthelia2.Arguments{oidc.GrantTypeAuthorizationCode}).AnyTimes()
				client.EXPECT().GetEnableJWTProfileOAuthAccessTokens().Return(false).AnyTimes()
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session) {
				assert.Nil(t, session.AccessToken)
			},
		},
		{
			name:    "ShouldSkipJWTProfileWithoutSubject",
			session: oidc.NewSession(),
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy) {
				requester.EXPECT().GetGrantTypes().Return(oauthelia2.Arguments{oidc.GrantTypeAuthorizationCode}).AnyTimes()
				requester.EXPECT().GetRequestedScopes().Return(oauthelia2.Arguments{oidc.ScopeOpenID}).AnyTimes()
				requester.EXPECT().GetID().Return(testValue).AnyTimes()

				client.EXPECT().GetEnableJWTProfileOAuthAccessTokens().Return(true).AnyTimes()
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session) {
				assert.Nil(t, session.AccessToken)
			},
		},
		{
			name:    "ShouldHandleDetailerError",
			session: newTestOIDCSessionWithSubject(subject.String()),
			handled: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy) {
				requester.EXPECT().GetGrantTypes().Return(oauthelia2.Arguments{oidc.GrantTypeAuthorizationCode}).AnyTimes()
				requester.EXPECT().GetRequestedScopes().Return(oauthelia2.Arguments{oidc.ScopeOpenID}).AnyTimes()
				requester.EXPECT().GetID().Return(testValue).AnyTimes()

				client.EXPECT().GetEnableJWTProfileOAuthAccessTokens().Return(true).AnyTimes()

				mock.StorageMock.EXPECT().
					LoadUserOpaqueIdentifier(gomock.Eq(mock.Ctx), gomock.Eq(subject)).
					Return(nil, fmt.Errorf("error in db"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Access Request encountered an error while trying to obtain the detailer to hydrate the JWT Profile Access Token claims", regexpAnyError)
			},
		},
		{
			name:    "ShouldHandleClaimsHydrationError",
			session: newTestOIDCSessionWithSubject(subject.String()),
			handled: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy) {
				requester.EXPECT().GetGrantTypes().Return(oauthelia2.Arguments{oidc.GrantTypeAuthorizationCode}).AnyTimes()
				requester.EXPECT().GetRequestedScopes().Return(oauthelia2.Arguments{oidc.ScopeOpenID}).AnyTimes()
				requester.EXPECT().GetRequestedAt().Return(time.Unix(1000000, 0)).AnyTimes()
				requester.EXPECT().GetID().Return(testValue).AnyTimes()

				client.EXPECT().GetEnableJWTProfileOAuthAccessTokens().Return(true).AnyTimes()
				client.EXPECT().GetClaimsStrategy().Return(strategy).AnyTimes()

				expectSubjectLookup(t, mock, subject)

				strategy.EXPECT().
					HydrateAccessTokenClaims(gomock.Any(), gomock.Any(), gomock.Eq(client), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("bad claims"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Access Request encountered an error while trying to hydrate the JWT Profile Access Token claims", regexpAnyError)
			},
		},
		{
			name:    "ShouldHydrateJWTProfileClaimsIntoNewAccessToken",
			session: newTestOIDCSessionWithSubject(subject.String()),
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy) {
				requester.EXPECT().GetGrantTypes().Return(oauthelia2.Arguments{oidc.GrantTypeAuthorizationCode}).AnyTimes()
				requester.EXPECT().GetRequestedScopes().Return(oauthelia2.Arguments{oidc.ScopeOpenID}).AnyTimes()
				requester.EXPECT().GetRequestedAt().Return(time.Unix(1000000, 0)).AnyTimes()
				requester.EXPECT().GetID().Return(testValue).AnyTimes()

				client.EXPECT().GetEnableJWTProfileOAuthAccessTokens().Return(true).AnyTimes()
				client.EXPECT().GetClaimsStrategy().Return(strategy).AnyTimes()

				expectSubjectLookup(t, mock, subject)

				strategy.EXPECT().
					HydrateAccessTokenClaims(gomock.Any(), gomock.Any(), gomock.Eq(client), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_, _, _, _, _, _, _, _, _, _ any, extra map[string]any) error {
						extra["custom"] = "value"

						return nil
					})
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session) {
				require.NotNil(t, session.AccessToken)

				assert.Equal(t, "value", session.AccessToken.Claims["custom"])
				assert.NotNil(t, session.AccessToken.Headers)
			},
		},
		{
			name:    "ShouldHydrateJWTProfileClaimsIntoExistingAccessToken",
			session: newTestOIDCSessionWithAccessToken(subject.String()),
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy) {
				requester.EXPECT().GetGrantTypes().Return(oauthelia2.Arguments{oidc.GrantTypeAuthorizationCode}).AnyTimes()
				requester.EXPECT().GetRequestedScopes().Return(oauthelia2.Arguments{oidc.ScopeOpenID}).AnyTimes()
				requester.EXPECT().GetRequestedAt().Return(time.Unix(1000000, 0)).AnyTimes()
				requester.EXPECT().GetID().Return(testValue).AnyTimes()

				client.EXPECT().GetEnableJWTProfileOAuthAccessTokens().Return(true).AnyTimes()
				client.EXPECT().GetClaimsStrategy().Return(strategy).AnyTimes()

				expectSubjectLookup(t, mock, subject)

				strategy.EXPECT().
					HydrateAccessTokenClaims(gomock.Any(), gomock.Any(), gomock.Eq(client), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_, _, _, _, _, _, _, _, _, _ any, extra map[string]any) error {
						extra["fresh"] = true

						return nil
					})
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session) {
				require.NotNil(t, session.AccessToken)

				assert.Equal(t, true, session.AccessToken.Claims["fresh"])
				assert.NotContains(t, session.AccessToken.Claims, "stale")
				assert.Equal(t, testOIDCKeyID, session.AccessToken.Headers[oidc.JWTHeaderKeyIdentifier])
			},
		},
		{
			name:    "ShouldNotSetAccessTokenWhenNoClaimsProduced",
			session: newTestOIDCSessionWithSubject(subject.String()),
			handled: false,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, requester *mocks.MockAccessRequester, client *mocks.MockOIDCClient, strategy *mocks.MockClaimsStrategy) {
				requester.EXPECT().GetGrantTypes().Return(oauthelia2.Arguments{oidc.GrantTypeAuthorizationCode}).AnyTimes()
				requester.EXPECT().GetRequestedScopes().Return(oauthelia2.Arguments{oidc.ScopeOpenID}).AnyTimes()
				requester.EXPECT().GetRequestedAt().Return(time.Unix(1000000, 0)).AnyTimes()
				requester.EXPECT().GetID().Return(testValue).AnyTimes()

				client.EXPECT().GetEnableJWTProfileOAuthAccessTokens().Return(true).AnyTimes()
				client.EXPECT().GetClaimsStrategy().Return(strategy).AnyTimes()

				expectSubjectLookup(t, mock, subject)

				strategy.EXPECT().
					HydrateAccessTokenClaims(gomock.Any(), gomock.Any(), gomock.Eq(client), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, session *oidc.Session) {
				assert.Nil(t, session.AccessToken)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			setupTestOIDCProvider(t, mock, nil)

			requester := mocks.NewMockAccessRequester(mock.Ctrl)
			client := mocks.NewMockOIDCClient(mock.Ctrl)
			strategy := mocks.NewMockClaimsStrategy(mock.Ctrl)

			rw := httptest.NewRecorder()

			if tc.setup != nil {
				tc.setup(t, mock, requester, client, strategy)
			}

			handled := handleOAuth2TokenHydration(mock.Ctx, rw, requester, client, tc.session)

			assert.Equal(t, tc.handled, handled)

			if tc.expect != nil {
				tc.expect(t, mock, tc.session)
			}
		})
	}
}

func expectSubjectLookup(t *testing.T, mock *mocks.MockAutheliaCtx, subject uuid.UUID) {
	t.Helper()

	mock.StorageMock.EXPECT().
		LoadUserOpaqueIdentifier(gomock.Eq(mock.Ctx), gomock.Eq(subject)).
		Return(&model.UserOpaqueIdentifier{Service: "openid", Username: testUsername, Identifier: subject}, nil)

	mock.UserProviderMock.EXPECT().
		GetDetailsExtended(gomock.Eq(testUsername)).
		Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}}}, nil)
}

func newTestOIDCSessionWithSubject(subject string) (session *oidc.Session) {
	session = oidc.NewSession()
	session.Subject = subject

	return session
}

func newTestOIDCSessionWithAccessToken(subject string) (session *oidc.Session) {
	session = newTestOIDCSessionWithSubject(subject)
	session.AccessToken = &oidc.AccessTokenSession{
		Headers: map[string]any{oidc.JWTHeaderKeyIdentifier: testOIDCKeyID},
		Claims:  map[string]any{"stale": true},
	}

	return session
}
