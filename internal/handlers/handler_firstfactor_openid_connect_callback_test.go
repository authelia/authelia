package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidcrp"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestFirstFactorOpenIDConnectCallbackGET(t *testing.T) {
	key, server, requests := newTestOpenIDConnectUpstream(t)

	defer server.Close()

	var observed int64

	testCases := []struct {
		Name          string
		Provider      string
		Trust         bool
		TokenPath     string
		Authenticated bool
		Pending       *session.OpenIDConnectPending
		Flow          *session.OpenIDConnectFlow
		Query         map[string]string
		Setup         func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Assert        func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Location      string
	}{
		{
			Name:          "ShouldAuthenticateLinkedUser",
			Provider:      "example",
			Authenticated: true,
			Pending: &session.OpenIDConnectPending{
				Provider:       "example",
				Issuer:         "https://op.example.com",
				Subject:        "planted",
				RemoteUsername: "attacker",
				DisplayName:    "Attacker",
				Email:          "attacker@example.org",
				Expires:        time.Unix(9999999999, 0),
			},
			Flow:  &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://app.example.com", RequestMethod: fasthttp.MethodGet},
			Query: map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setupTestOpenIDConnectCallbackLinked(mock)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Equal(t, "john", userSession.Username)
				assert.Equal(t, "John Smith", userSession.DisplayName)
				assert.Equal(t, []string{"admins"}, userSession.Groups)
				assert.Equal(t, []string{"john@example.com"}, userSession.Emails)
				assert.Equal(t, authorization.AuthenticationMethodsReferences{FederatedIdentity: true}, userSession.AuthenticationMethodRefs)
				assert.Equal(t, authentication.OneFactor, userSession.AuthenticationLevel(false))
				assert.Nil(t, userSession.Elevations.User)
				assert.Zero(t, userSession.SecondFactorAuthnTimestamp)
				assert.Nil(t, userSession.OpenIDConnect)
				assert.Nil(t, userSession.OpenIDConnectPending)
			},
			Location: "https://app.example.com/",
		},
		{
			Name:     "ShouldNotAdoptUntrustedAuthenticationMethodsReference",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://app.example.com", RequestMethod: fasthttp.MethodGet},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setupTestOpenIDConnectCallbackLinked(mock)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Equal(t, authorization.AuthenticationMethodsReferences{FederatedIdentity: true}, userSession.AuthenticationMethodRefs)
				assert.Empty(t, userSession.AuthenticationMethodRefs.MarshalRFC8176())
				assert.False(t, userSession.AuthenticationMethodRefs.FactorKnowledge())
				assert.False(t, userSession.AuthenticationMethodRefs.FactorPossession())
				assert.False(t, userSession.IsAnonymous())
				assert.Equal(t, authentication.OneFactor, userSession.AuthenticationLevel(false))
			},
			Location: "https://app.example.com/",
		},
		{
			Name:     "ShouldAdoptTrustedAuthenticationMethodsReference",
			Provider: "example",
			Trust:    true,
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://app.example.com", RequestMethod: fasthttp.MethodGet},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setupTestOpenIDConnectCallbackLinked(mock)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Equal(t, authorization.AuthenticationMethodsReferences{FederatedIdentity: true, UsernameAndPassword: true, TOTP: true}, userSession.AuthenticationMethodRefs)
				assert.Equal(t, []string{"pwd", "otp", "mfa"}, userSession.AuthenticationMethodRefs.MarshalRFC8176())
				assert.False(t, userSession.IsAnonymous())
				assert.Equal(t, authentication.TwoFactor, userSession.AuthenticationLevel(false))
			},
			Location: "https://app.example.com/",
		},
		{ //nolint:gosec // Test Values.
			Name:      "ShouldAdoptTrustedPossessionOnlyAuthenticationMethodsReference",
			Provider:  "example",
			Trust:     true,
			TokenPath: "/token-amr-otp",
			Flow:      &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://app.example.com", RequestMethod: fasthttp.MethodGet},
			Query:     map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setupTestOpenIDConnectCallbackLinked(mock)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Equal(t, authorization.AuthenticationMethodsReferences{FederatedIdentity: true, TOTP: true}, userSession.AuthenticationMethodRefs)
				assert.Equal(t, []string{"otp"}, userSession.AuthenticationMethodRefs.MarshalRFC8176())
				assert.True(t, userSession.AuthenticationMethodRefs.FactorPossession())
				assert.False(t, userSession.AuthenticationMethodRefs.FactorKnowledge())
				assert.Equal(t, authentication.TwoFactor, userSession.AuthenticationLevel(false))
			},
			Location: "https://app.example.com/",
		},
		{
			Name:     "ShouldNotRedirectToAnUnsafeTargetURL",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://evil.example.org/steal", RequestMethod: fasthttp.MethodGet},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setupTestOpenIDConnectCallbackLinked(mock)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Equal(t, "john", userSession.Username)
			},
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldNotRedirectToATargetURLRequiringTwoFactor",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://two-factor.example.com/", RequestMethod: fasthttp.MethodGet},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setupTestOpenIDConnectCallbackLinked(mock)
			},
			Assert:   nil,
			Location: "http://login.example.com:8080/",
		},
		{
			Name:          "ShouldStashPendingLinkForUnlinkedIdentityWhenAuthenticated",
			Provider:      "example",
			Authenticated: true,
			Flow:          &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://app.example.com", RequestMethod: fasthttp.MethodGet},
			Query:         map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinkBySubject(mock.Ctx, "https://op.example.com", "abc123").
					Return(nil, storage.ErrNoOpenIDConnectLink)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				require.NotNil(t, userSession.OpenIDConnectPending)

				assert.Equal(t, "example", userSession.OpenIDConnectPending.Provider)
				assert.Equal(t, "https://op.example.com", userSession.OpenIDConnectPending.Issuer)
				assert.Equal(t, "abc123", userSession.OpenIDConnectPending.Subject)
				assert.Equal(t, "john", userSession.OpenIDConnectPending.RemoteUsername)
				assert.Equal(t, "John Smith", userSession.OpenIDConnectPending.DisplayName)
				assert.Equal(t, "john@example.com", userSession.OpenIDConnectPending.Email)
				assert.False(t, userSession.OpenIDConnectPending.Expires.IsZero())
				assert.False(t, userSession.IsAnonymous())
				assert.Nil(t, userSession.OpenIDConnect)
			},
			Location: "http://login.example.com:8080/settings/openid-connect",
		},
		{
			Name:     "ShouldNotStashPendingLinkForUnlinkedIdentityWhenAnonymous",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://app.example.com", RequestMethod: fasthttp.MethodGet},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinkBySubject(mock.Ctx, "https://op.example.com", "abc123").
					Return(nil, storage.ErrNoOpenIDConnectLink)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
				assert.Nil(t, userSession.OpenIDConnect)
				assert.Nil(t, userSession.OpenIDConnectPending)

				raw, err := json.Marshal(userSession)
				require.NoError(t, err)

				for _, value := range []string{"https://op.example.com", "abc123", "john", "John Smith", "john@example.com"} {
					assert.NotContains(t, string(raw), value)
				}
			},
			Location: "http://login.example.com:8080/?link_provider=example",
		},
		{
			Name:     "ShouldDiscardAPreExistingPendingLinkWhenAnonymous",
			Provider: "example",
			Pending: &session.OpenIDConnectPending{
				Provider:       "example",
				Issuer:         "https://op.example.com",
				Subject:        "planted",
				RemoteUsername: "attacker",
				DisplayName:    "Attacker",
				Email:          "attacker@example.org",
				Expires:        time.Unix(9999999999, 0),
			},
			Flow:  &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://app.example.com", RequestMethod: fasthttp.MethodGet},
			Query: map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinkBySubject(mock.Ctx, "https://op.example.com", "abc123").
					Return(nil, storage.ErrNoOpenIDConnectLink)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
				assert.Nil(t, userSession.OpenIDConnectPending)
			},
			Location: "http://login.example.com:8080/?link_provider=example",
		},
		{
			Name:     "ShouldRejectMismatchedState",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"code": "the-code", "state": "wrong-state"},
			Setup:    nil,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
				assert.Nil(t, userSession.OpenIDConnect)
				assert.Nil(t, userSession.OpenIDConnectPending)
			},
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectAbsentState",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"code": "the-code"},
			Setup:    nil,
			Assert:   nil,
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectFlowStateWithoutAStateValue",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"code": "the-code"},
			Setup:    nil,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
				assert.Nil(t, userSession.OpenIDConnect)
			},
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectAbsentFlowState",
			Provider: "example",
			Flow:     nil,
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup:    nil,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
				assert.Nil(t, userSession.OpenIDConnectPending)
			},
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectExpiredFlowState",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", Expires: time.Unix(1, 0)},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup:    nil,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
				assert.Nil(t, userSession.OpenIDConnect)
			},
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectUnknownProvider",
			Provider: "missing",
			Flow:     &session.OpenIDConnectFlow{Provider: "missing", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup:    nil,
			Assert:   nil,
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectFlowStateForAnotherProvider",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "other", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup:    nil,
			Assert:   nil,
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectMismatchedIssuerParameter",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"code": "the-code", "state": "the-state", "iss": "https://evil.example.org"},
			Setup:    nil,
			Assert:   nil,
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldAcceptMatchingIssuerParameter",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"code": "the-code", "state": "the-state", "iss": "https://op.example.com"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinkBySubject(mock.Ctx, "https://op.example.com", "abc123").
					Return(nil, storage.ErrNoOpenIDConnectLink)
			},
			Assert:   nil,
			Location: "http://login.example.com:8080/?link_provider=example",
		},
		{
			Name:     "ShouldRejectProviderError",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"code": "the-code", "state": "the-state", "error": "access_denied", "error_description": "<script>alert(1)</script>"},
			Setup:    nil,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, `<a href="http://login.example.com:8080/">302 Found</a>`, string(mock.Ctx.Response.Body()))
			},
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectAbsentCodeWithoutContactingTheProvider",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				observed = requests.Load()
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, observed, requests.Load())
			},
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectMismatchedCodeVerifier",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "wrong-verifier"},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup:    nil,
			Assert:   nil,
			Location: "http://login.example.com:8080/",
		},
		{
			Name:      "ShouldRejectTokenEndpointError",
			Provider:  "example",
			TokenPath: "/token-error",
			Flow:      &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:     map[string]string{"code": "the-code", "state": "the-state"},
			Setup:     nil,
			Assert:    nil,
			Location:  "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectMismatchedNonce",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "wrong-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup:    nil,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
				assert.Nil(t, userSession.OpenIDConnectPending)
			},
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectBannedUser",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://app.example.com", RequestMethod: fasthttp.MethodGet},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinkBySubject(mock.Ctx, "https://op.example.com", "abc123").
					Return(&model.OpenIDConnectLink{ID: 1, Provider: "example", Issuer: "https://op.example.com", Subject: "abc123", Username: "john"}, nil)

				mock.StorageMock.EXPECT().
					LoadBannedIP(mock.Ctx, model.NewIP(mock.Ctx.RemoteIP())).
					Return(nil, nil)

				mock.StorageMock.EXPECT().
					LoadBannedUser(mock.Ctx, "john").
					Return([]model.BannedUser{{Username: "john", Expires: sql.NullTime{Valid: true, Time: time.Unix(20000, 0)}}}, nil)

				mock.StorageMock.EXPECT().
					AppendAuthenticationLog(mock.Ctx, model.AuthenticationAttempt{
						Time:          mock.Clock.Now(),
						Successful:    false,
						Banned:        true,
						Username:      "john",
						Type:          regulation.AuthType1FA,
						RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
						RequestURI:    "https://app.example.com",
						RequestMethod: fasthttp.MethodGet,
					}).
					Return(nil)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
				assert.Empty(t, userSession.Username)
			},
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectUnknownUser",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier", TargetURL: "https://app.example.com", RequestMethod: fasthttp.MethodGet},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinkBySubject(mock.Ctx, "https://op.example.com", "abc123").
					Return(&model.OpenIDConnectLink{ID: 1, Provider: "example", Issuer: "https://op.example.com", Subject: "abc123", Username: "john"}, nil)

				mock.StorageMock.EXPECT().
					LoadBannedIP(mock.Ctx, model.NewIP(mock.Ctx.RemoteIP())).
					Return(nil, nil)

				mock.StorageMock.EXPECT().
					LoadBannedUser(mock.Ctx, "john").
					Return(nil, nil)

				mock.UserProviderMock.EXPECT().
					GetDetails("john").
					Return(nil, authentication.ErrUserNotFound)

				mock.StorageMock.EXPECT().
					AppendAuthenticationLog(mock.Ctx, gomock.Any()).
					Return(nil)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
			},
			Location: "http://login.example.com:8080/",
		},
		{
			Name:     "ShouldRejectStorageFailure",
			Provider: "example",
			Flow:     &session.OpenIDConnectFlow{Provider: "example", State: "the-state", Nonce: "the-nonce", CodeVerifier: "the-verifier"},
			Query:    map[string]string{"code": "the-code", "state": "the-state"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinkBySubject(mock.Ctx, "https://op.example.com", "abc123").
					Return(nil, sql.ErrConnDone)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
				assert.Nil(t, userSession.OpenIDConnectPending)
			},
			Location: "http://login.example.com:8080/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := newTestOpenIDConnectCallbackMock(t)

			defer mock.Close()

			tokenPath := tc.TokenPath

			if tokenPath == "" {
				tokenPath = "/token"
			}

			mock.Ctx.Providers.OpenIDConnectRelyingParty = newTestRelyingPartyProvidersWithUpstream(server.URL+tokenPath, key, tc.Trust)
			mock.Ctx.SetUserValue("provider", tc.Provider)

			query := url.Values{}

			for k, v := range tc.Query {
				query.Set(k, v)
			}

			mock.Ctx.Request.SetRequestURI("/api/firstfactor/openid-connect/" + tc.Provider + "/callback?" + query.Encode())
			mock.Ctx.Request.SetHost("login.example.com:8080")

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			if tc.Flow != nil {
				if tc.Flow.Expires.IsZero() {
					tc.Flow.Expires = mock.Ctx.GetClock().Now().Add(time.Minute * 3)
				}

				userSession.OpenIDConnect = tc.Flow
			}

			if tc.Pending != nil {
				userSession.OpenIDConnectPending = tc.Pending
			}

			if tc.Authenticated {
				userSession.SetOneFactorPassword(mock.Ctx.GetClock().Now(), &authentication.UserDetails{Username: "harry", DisplayName: "Harry Potter", Emails: []string{"harry@example.com"}, Groups: []string{"dev"}}, false)
				userSession.Elevations.User = &session.Elevation{ID: 1, RemoteIP: net.ParseIP("127.0.0.1"), Expires: mock.Ctx.GetClock().Now().Add(time.Minute * 10)}
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			if tc.Setup != nil {
				tc.Setup(t, mock)
			}

			FirstFactorOpenIDConnectCallbackGET(mock.Ctx)

			assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.Location, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

			if tc.Assert != nil {
				tc.Assert(t, mock)
			}
		})
	}
}

func TestFirstFactorOpenIDConnectCallbackGETShouldConsumeTheFlowStateExactlyOnce(t *testing.T) {
	key, server, _ := newTestOpenIDConnectUpstream(t)

	defer server.Close()

	mock := newTestOpenIDConnectCallbackMock(t)

	defer mock.Close()

	mock.Ctx.Providers.OpenIDConnectRelyingParty = newTestRelyingPartyProvidersWithUpstream(server.URL+"/token", key, false)
	mock.Ctx.SetUserValue("provider", "example")
	mock.Ctx.Request.SetRequestURI("/api/firstfactor/openid-connect/example/callback?code=the-code&state=the-state")
	mock.Ctx.Request.SetHost("login.example.com:8080")

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	userSession.OpenIDConnect = &session.OpenIDConnectFlow{
		Provider:     "example",
		State:        "the-state",
		Nonce:        "the-nonce",
		CodeVerifier: "the-verifier",
		TargetURL:    "https://app.example.com",
		Expires:      mock.Ctx.GetClock().Now().Add(time.Minute * 3),
	}

	require.NoError(t, mock.Ctx.SaveSession(userSession))

	mock.StorageMock.EXPECT().
		LoadOpenIDConnectLinkBySubject(mock.Ctx, "https://op.example.com", "abc123").
		Return(nil, storage.ErrNoOpenIDConnectLink).
		Times(1)

	FirstFactorOpenIDConnectCallbackGET(mock.Ctx)

	assert.Equal(t, "http://login.example.com:8080/?link_provider=example", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

	userSession, err = mock.Ctx.GetSession()
	require.NoError(t, err)

	assert.Nil(t, userSession.OpenIDConnect)

	mock.Ctx.Response.Reset()

	FirstFactorOpenIDConnectCallbackGET(mock.Ctx)

	assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
	assert.Equal(t, "http://login.example.com:8080/", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

	userSession, err = mock.Ctx.GetSession()
	require.NoError(t, err)

	assert.Nil(t, userSession.OpenIDConnectPending)
}

func TestFirstFactorOpenIDConnectCallbackGETShouldRegenerateTheSession(t *testing.T) {
	key, server, _ := newTestOpenIDConnectUpstream(t)

	defer server.Close()

	mock := newTestOpenIDConnectCallbackMock(t)

	defer mock.Close()

	mock.Ctx.Providers.OpenIDConnectRelyingParty = newTestRelyingPartyProvidersWithUpstream(server.URL+"/token", key, false)
	mock.Ctx.SetUserValue("provider", "example")
	mock.Ctx.Request.SetRequestURI("/api/firstfactor/openid-connect/example/callback?code=the-code&state=the-state")
	mock.Ctx.Request.SetHost("login.example.com:8080")

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	userSession.OpenIDConnect = &session.OpenIDConnectFlow{
		Provider:      "example",
		State:         "the-state",
		Nonce:         "the-nonce",
		CodeVerifier:  "the-verifier",
		TargetURL:     "https://app.example.com",
		RequestMethod: fasthttp.MethodGet,
		Expires:       mock.Ctx.GetClock().Now().Add(time.Minute * 3),
	}

	require.NoError(t, mock.Ctx.SaveSession(userSession))

	before := string(mock.Ctx.Response.Header.PeekCookie("authelia_session"))

	require.NotEmpty(t, before)

	setupTestOpenIDConnectCallbackLinked(mock)

	FirstFactorOpenIDConnectCallbackGET(mock.Ctx)

	after := string(mock.Ctx.Response.Header.PeekCookie("authelia_session"))

	require.NotEmpty(t, after)

	assert.NotEqual(t, before, after)

	userSession, err = mock.Ctx.GetSession()
	require.NoError(t, err)

	assert.Equal(t, "john", userSession.Username)
}

func TestFirstFactorOpenIDConnectCallbackGETRememberMe(t *testing.T) {
	key, server, _ := newTestOpenIDConnectUpstream(t)

	defer server.Close()

	testCases := []struct {
		Name               string
		DisableRememberMe  bool
		KeepMeLoggedIn     bool
		ExpectedSession    bool
		ExpectedExpiration time.Duration
	}{
		{
			Name:               "ShouldExtendTheSessionExpirationWhenRequested",
			DisableRememberMe:  false,
			KeepMeLoggedIn:     true,
			ExpectedSession:    true,
			ExpectedExpiration: schema.DefaultSessionConfiguration.RememberMe,
		},
		{
			Name:               "ShouldNotExtendTheSessionExpirationWhenRememberMeIsDisabled",
			DisableRememberMe:  true,
			KeepMeLoggedIn:     true,
			ExpectedSession:    false,
			ExpectedExpiration: schema.DefaultSessionConfiguration.Expiration,
		},
		{
			Name:               "ShouldNotExtendTheSessionExpirationWhenNotRequested",
			DisableRememberMe:  false,
			KeepMeLoggedIn:     false,
			ExpectedSession:    false,
			ExpectedExpiration: schema.DefaultSessionConfiguration.Expiration,
		},
		{
			Name:               "ShouldNotExtendTheSessionExpirationWhenNotRequestedAndRememberMeIsDisabled",
			DisableRememberMe:  true,
			KeepMeLoggedIn:     false,
			ExpectedSession:    false,
			ExpectedExpiration: schema.DefaultSessionConfiguration.Expiration,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := newTestOpenIDConnectCallbackMock(t)

			defer mock.Close()

			if tc.DisableRememberMe {
				configuration := mock.Ctx.Configuration.Session
				configuration.Cookies[0].DisableRememberMe = true

				mock.Ctx.Providers.SessionProvider = session.NewProvider(configuration, nil)
			}

			mock.Ctx.Providers.OpenIDConnectRelyingParty = newTestRelyingPartyProvidersWithUpstream(server.URL+"/token", key, false)
			mock.Ctx.SetUserValue("provider", "example")
			mock.Ctx.Request.SetRequestURI("/api/firstfactor/openid-connect/example/callback?code=the-code&state=the-state")
			mock.Ctx.Request.SetHost("login.example.com:8080")

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.OpenIDConnect = &session.OpenIDConnectFlow{
				Provider:       "example",
				State:          "the-state",
				Nonce:          "the-nonce",
				CodeVerifier:   "the-verifier",
				TargetURL:      "https://app.example.com",
				RequestMethod:  fasthttp.MethodGet,
				KeepMeLoggedIn: tc.KeepMeLoggedIn,
				Expires:        mock.Ctx.GetClock().Now().Add(time.Minute * 3),
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			sessionProvider, err := mock.Ctx.GetSessionProvider()
			require.NoError(t, err)

			expiration, err := sessionProvider.GetExpiration(mock.Ctx.RequestCtx)
			require.NoError(t, err)
			require.Equal(t, schema.DefaultSessionConfiguration.Expiration, expiration)

			setupTestOpenIDConnectCallbackLinked(mock)

			FirstFactorOpenIDConnectCallbackGET(mock.Ctx)

			assert.Equal(t, "https://app.example.com/", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

			userSession, err = mock.Ctx.GetSession()
			require.NoError(t, err)

			assert.Equal(t, "john", userSession.Username)
			assert.Equal(t, tc.ExpectedSession, userSession.KeepMeLoggedIn)

			expiration, err = sessionProvider.GetExpiration(mock.Ctx.RequestCtx)
			require.NoError(t, err)

			assert.Equal(t, tc.ExpectedExpiration, expiration)
		})
	}
}

func newTestOpenIDConnectCallbackMock(t *testing.T) *mocks.MockAutheliaCtx {
	mock := mocks.NewMockAutheliaCtx(t)

	mock.Ctx.Init2(nil, nil, true)

	return mock
}

func setupTestOpenIDConnectCallbackLinked(mock *mocks.MockAutheliaCtx) {
	mock.StorageMock.EXPECT().
		LoadOpenIDConnectLinkBySubject(mock.Ctx, "https://op.example.com", "abc123").
		Return(&model.OpenIDConnectLink{ID: 1, Provider: "example", Issuer: "https://op.example.com", Subject: "abc123", Username: "john"}, nil)

	mock.StorageMock.EXPECT().
		LoadBannedIP(mock.Ctx, model.NewIP(mock.Ctx.RemoteIP())).
		Return(nil, nil)

	mock.StorageMock.EXPECT().
		LoadBannedUser(mock.Ctx, "john").
		Return(nil, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails("john").
		Return(&authentication.UserDetails{Username: "john", DisplayName: "John Smith", Emails: []string{"john@example.com"}, Groups: []string{"admins"}}, nil)

	mock.StorageMock.EXPECT().
		UpdateOpenIDConnectLinkSignIn(mock.Ctx, 1, gomock.Any()).
		Return(nil)

	mock.StorageMock.EXPECT().
		AppendAuthenticationLog(mock.Ctx, gomock.Any()).
		Return(nil)
}

func newTestOpenIDConnectUpstream(t *testing.T) (key *rsa.PrivateKey, server *httptest.Server, requests *atomic.Int64) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	requests = &atomic.Int64{}

	mux := http.NewServeMux()

	mux.HandleFunc("/token", func(rw http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		rw.Header().Set("Content-Type", "application/json")

		if err = r.ParseForm(); err != nil || r.PostFormValue("code") != "the-code" || r.PostFormValue("code_verifier") != "the-verifier" || r.PostFormValue("grant_type") != "authorization_code" {
			rw.WriteHeader(http.StatusBadRequest)

			_ = json.NewEncoder(rw).Encode(map[string]any{"error": "invalid_grant"})

			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"},
			"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "nonce": "the-nonce",
			"preferred_username": "john", "name": "John Smith", "email": "john@example.com",
			"email_verified": false, "amr": []string{"pwd", "otp"},
		})

		token.Header["kid"] = "kid1"

		var raw string

		if raw, err = token.SignedString(key); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)

			return
		}

		_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "at", "token_type": "Bearer", "id_token": raw})
	})

	mux.HandleFunc("/token-amr-otp", func(rw http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		rw.Header().Set("Content-Type", "application/json")

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss": "https://op.example.com", "sub": "abc123", "aud": []string{"client"},
			"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "nonce": "the-nonce",
			"preferred_username": "john", "name": "John Smith", "email": "john@example.com",
			"email_verified": false, "amr": []string{"otp"},
		})

		token.Header["kid"] = "kid1"

		var raw string

		if raw, err = token.SignedString(key); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)

			return
		}

		_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "at", "token_type": "Bearer", "id_token": raw})
	})

	mux.HandleFunc("/token-error", func(rw http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)

		_ = json.NewEncoder(rw).Encode(map[string]any{"error": "invalid_client"})
	})

	return key, httptest.NewServer(mux), requests
}

func newTestRelyingPartyProvidersWithUpstream(tokenEndpoint string, key *rsa.PrivateKey, trust bool) *oidcrp.Providers {
	providers := oidcrp.NewProviders(&schema.AuthenticationBackendOpenIDConnect{
		Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
			{
				ID: "example", Name: "Example", Issuer: "https://op.example.com",
				ClientID: "client", ClientSecret: "secret",
				Scopes:                   []string{"openid", "email"},
				TokenEndpointAuthMethod:  "client_secret_basic",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "S256"},
				Discovery:                schema.AuthenticationBackendOpenIDConnectProviderDiscovery{Disable: true},
				AuthenticationMethodsReference: schema.AuthenticationBackendOpenIDConnectProviderAMR{
					Trust: trust,
				},
				Endpoints: schema.AuthenticationBackendOpenIDConnectProviderEndpoints{
					Authorization: "https://op.example.com/authorize",
					Token:         tokenEndpoint,
				},
				JSONWebKeys: []schema.JWK{
					{KeyID: "kid1", Use: "sig", Algorithm: "RS256", Key: key.Public()},
				},
			},
		},
	}, nil)

	return providers
}
