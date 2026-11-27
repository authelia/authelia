// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

func (s *AuthzSuite) TestShouldNotSendIdentityHeadersForAHeldSessionOnBypass() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	authz := s.Builder().WithConfig(&mock.Ctx.Configuration).Build()

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.SetPasswordChangeRequired(mock.Clock.Now(), testUsername)

	s.Require().NoError(mock.Ctx.SaveSession(&userSession))
	s.Require().Equal(authentication.NotAuthenticated, userSession.AuthenticationLevel(false))
	s.Require().Empty(userSession.Username)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Empty(mock.Ctx.Response.Header.PeekBytes(headerRemoteUser))
	s.Empty(mock.Ctx.Response.Header.PeekBytes(headerRemoteGroups))
	s.Empty(mock.Ctx.Response.Header.PeekBytes(headerRemoteEmail))
	s.Empty(mock.Ctx.Response.Header.PeekBytes(headerRemoteName))
}

func (s *AuthzSuite) TestShouldRefuseBasicAuthorizationForAUserWhoMustChangeTheirPassword() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuilderWithProxyAuthorizationBasicSchemeCached().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Providers.UserAttributeResolver = &expression.UserAttributes{}
	mock.Ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute = "pwd_reset"

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.EXPECT().GetDetailsCached(gomock.Eq("john")).Return(&authentication.UserDetails{Username: "john"}, nil).AnyTimes()
	mock.UserProviderMock.EXPECT().CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).Return(true, nil).AnyTimes()
	mock.StorageMock.EXPECT().LoadBannedIP(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mock.StorageMock.EXPECT().LoadBannedUser(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mock.StorageMock.EXPECT().AppendAuthenticationLog(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mock.UserProviderMock.EXPECT().GetDetailsExtendedCached(gomock.Eq("john")).
		Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: "john"}, Extra: map[string]any{"pwd_reset": true}}, nil)

	authz.Handler(mock.Ctx)

	s.NotEqual(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Empty(mock.Ctx.Response.Header.PeekBytes(headerRemoteUser))
}

func (s *AuthzSuite) TestShouldAllowBasicAuthorizationForAUserWhoNeedNotChangeTheirPassword() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuilderWithProxyAuthorizationBasicSchemeCached().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Providers.UserAttributeResolver = &expression.UserAttributes{}
	mock.Ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute = "pwd_reset"

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.EXPECT().GetDetailsCached(gomock.Eq("john")).Return(&authentication.UserDetails{Username: "john"}, nil).AnyTimes()
	mock.UserProviderMock.EXPECT().CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).Return(true, nil).AnyTimes()
	mock.StorageMock.EXPECT().LoadBannedIP(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mock.StorageMock.EXPECT().LoadBannedUser(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mock.StorageMock.EXPECT().AppendAuthenticationLog(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mock.UserProviderMock.EXPECT().GetDetailsExtendedCached(gomock.Eq("john")).
		Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: "john"}, Extra: map[string]any{"pwd_reset": false}}, nil)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func TestCookieSessionAuthnStrategyPasswordChangeRequiredOnRefresh(t *testing.T) {
	testCases := []struct {
		name     string
		refresh  schema.RefreshIntervalDuration
		expired  bool
		flagged  bool
		lookup   bool
		expected string
	}{
		{"ShouldEndTheSessionWhenFlaggedAtTheRefreshInterval", schema.NewRefreshIntervalDuration(time.Minute), true, true, true, anonymous},
		{"ShouldKeepTheSessionWhenNotFlaggedAtTheRefreshInterval", schema.NewRefreshIntervalDuration(time.Minute), true, false, true, testUsername},
		{"ShouldNotCheckBeforeTheRefreshInterval", schema.NewRefreshIntervalDuration(time.Minute), false, true, false, testUsername},
		{"ShouldEndTheSessionWhenFlaggedWithRefreshAlways", schema.NewRefreshIntervalDurationAlways(), false, true, true, anonymous},
		{"ShouldNotCheckWithRefreshNever", schema.NewRefreshIntervalDurationNever(), true, true, false, testUsername},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Providers.UserAttributeResolver = &expression.UserAttributes{}
			mock.Ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute = "pwd_reset"

			targetURI := &url.URL{Scheme: "https", Host: "app.example.com"}

			primary, err := mock.Ctx.GetSessionManagerByTargetURI(targetURI)
			require.NoError(t, err)

			userSession := session.NewUserSession(testUsername)
			userSession.CookieDomain = "example.com"
			userSession.SetOneFactorPassword(mock.Ctx.GetClock().Now(), false)

			if tc.expired {
				userSession.RefreshTTL = mock.Ctx.GetClock().Now().Add(-time.Minute)
			} else {
				userSession.RefreshTTL = mock.Ctx.GetClock().Now().Add(time.Minute)
			}

			require.NoError(t, primary.SaveSession(&userSession))

			mock.UserProviderMock.EXPECT().
				GetDetailsCached(gomock.Eq(testUsername)).
				Return(&authentication.UserDetails{Username: testUsername}, nil).
				AnyTimes()

			if tc.lookup {
				mock.UserProviderMock.EXPECT().
					GetDetailsExtendedCached(gomock.Eq(testUsername)).
					Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: testUsername}, Extra: map[string]any{"pwd_reset": tc.flagged}}, nil)
			}

			manager, err := mock.Ctx.GetAuthzSessionManagerByTargetURI(targetURI)
			require.NoError(t, err)

			var authn *Authn

			withAuthzRequest(mock.Ctx, func() {
				authn, err = NewCookieSessionAuthnStrategy(tc.refresh).Get(mock.Ctx, manager, nil)
			})

			require.NoError(t, err)
			assert.Equal(t, tc.expected, authn.Username)
		})
	}
}

func TestCookieSessionAuthnStrategyPasswordChangeRequiredOnRefreshLookupFailure(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Providers.UserAttributeResolver = &expression.UserAttributes{}
	mock.Ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute = "pwd_reset"

	targetURI := &url.URL{Scheme: "https", Host: "app.example.com"}

	primary, err := mock.Ctx.GetSessionManagerByTargetURI(targetURI)
	require.NoError(t, err)

	userSession := session.NewUserSession(testUsername)
	userSession.CookieDomain = "example.com"
	userSession.SetOneFactorPassword(mock.Ctx.GetClock().Now(), false)

	require.NoError(t, primary.SaveSession(&userSession))

	mock.UserProviderMock.EXPECT().
		GetDetailsCached(gomock.Eq(testUsername)).
		Return(&authentication.UserDetails{Username: testUsername}, nil).
		AnyTimes()

	mock.UserProviderMock.EXPECT().
		GetDetailsExtendedCached(gomock.Eq(testUsername)).
		Return(nil, fmt.Errorf("failed to mock the details"))

	manager, err := mock.Ctx.GetAuthzSessionManagerByTargetURI(targetURI)
	require.NoError(t, err)

	var authn *Authn

	withAuthzRequest(mock.Ctx, func() {
		authn, err = NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()).Get(mock.Ctx, manager, nil)
	})

	require.NoError(t, err)
	assert.Equal(t, testUsername, authn.Username)
	mock.AssertLastLogMessage(t, "Error occurred determining if the user must change their password", "error occurred retrieving extended user details to determine if a password change is required: failed to mock the details")
}

func TestFirstFactorPasskeyPOSTShouldHoldAUserWhoMustChangeTheirPassword(t *testing.T) {
	clientJSON := `{"type":"webauthn.get","challenge":"in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE","origin":"https://login.example.com:8080","crossOrigin":false,"other_keys_can_be_added_here":"do not compare clientDataJSON against a template. See https://goo.gl/yabPex"}`
	body := fmt.Sprintf(`{"response":{"id":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","rawId":"rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU","response":{"authenticatorData":"DGygg5w6VoNVeDP2GKJVZmXfKgiJZHh9U4ULStTTvtwFAAAAAw","userHandle":"ZXhhbXBsZQ","clientDataJSON":"%s","signature":"MEQCIBlJ2Fxf6ZwLNTCQglz0AW0pD4HlU8W5Yk696jjfxVxhAiAhAMkLh8iKyhW6zSmzwfQDjMF2nKjVHzEs7jLHRPDZ2A"},"type":"public-key","clientExtensionResults":{},"authenticatorAttachment":"cross-platform"},"targetURL":null}`,
		base64.RawURLEncoding.EncodeToString([]byte(clientJSON)))

	testCases := []struct {
		name    string
		flagged bool
	}{
		{"ShouldHoldTheUserWhenFlagged", true},
		{"ShouldAuthenticateTheUserWhenNotFlagged", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			setUpMockClock(mock)

			mock.Ctx.Configuration.WebAuthn = schema.DefaultWebAuthnConfiguration
			mock.Ctx.Providers.UserAttributeResolver = &expression.UserAttributes{}
			mock.Ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute = "pwd_reset"
			mock.Ctx.Request.SetBodyString(body)
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "login.example.com:8080")

			us, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			us.WebAuthn = &session.WebAuthn{SessionData: &webauthn.SessionData{Challenge: "in1cL-oWfSjSd7uuwUvv2ndOAmRXb0cOAbUoTtAqvGE", Expires: time.Now().Add(time.Minute), UserVerification: "preferred"}}

			require.NoError(t, mock.Ctx.SaveSession(&us))

			credential := model.WebAuthnCredential{
				ID: 1, CreatedAt: time.Now(), LastUsedAt: sql.NullTime{Time: mock.Clock.Now().UTC().Add(time.Second * -10), Valid: true},
				RPID: "login.example.com", Username: testUsername, Description: "test",
				KID:             model.NewBase64(tDecodeBase64StringStdEncoding(t, "rwOwV8WCh1hrE0M6mvaoRGpGHidqK6IlhkDJ2xERhPU=")),
				AAGUID:          uuid.NullUUID{UUID: uuid.Must(uuid.Parse("01020304-0506-0708-0102-030405060708")), Valid: true},
				AttestationType: "packed", Attachment: "cross-platform", Transport: "usb", SignCount: 2,
				Discoverable: true, Present: true, Verified: true,
				PublicKey: []byte{165, 1, 2, 3, 38, 32, 1, 33, 88, 32, 184, 17, 198, 170, 14, 81, 23, 237, 100, 218, 123, 122, 48, 76, 56, 148, 23, 111, 173, 245, 67, 239, 176, 229, 199, 205, 213, 46, 239, 91, 222, 183, 34, 88, 32, 171, 141, 116, 74, 68, 180, 81, 66, 81, 127, 81, 41, 236, 173, 38, 7, 9, 34, 128, 167, 101, 51, 25, 84, 239, 100, 10, 124, 117, 165, 178, 179},
			}

			mock.StorageMock.EXPECT().LoadWebAuthnUserByUserID(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq("example")).Return(&model.WebAuthnUser{UserID: "example", Username: testUsername}, nil)
			mock.StorageMock.EXPECT().LoadWebAuthnPasskeyCredentialsByUsername(mock.Ctx, gomock.Eq("login.example.com"), gomock.Eq(testUsername)).Return([]model.WebAuthnCredential{credential}, nil)
			mock.StorageMock.EXPECT().UpdateWebAuthnCredentialSignIn(mock.Ctx, gomock.Any()).Return(nil)
			mock.UserProviderMock.EXPECT().GetDetailsExtended(gomock.Eq(testUsername)).
				Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: testUsername}, Extra: map[string]any{"pwd_reset": tc.flagged}}, nil)
			mock.StorageMock.EXPECT().LoadBannedIP(gomock.Any(), gomock.Any()).Return(nil, nil)
			mock.StorageMock.EXPECT().LoadBannedUser(gomock.Any(), gomock.Any()).Return(nil, nil)
			mock.StorageMock.EXPECT().AppendAuthenticationLog(gomock.Any(), gomock.Any()).Return(nil)

			FirstFactorPasskeyPOST(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

			us, err = mock.Ctx.GetSession()
			require.NoError(t, err)

			if tc.flagged {
				assert.Equal(t, authentication.NotAuthenticated, us.AuthenticationLevel(false))
				assert.True(t, us.IsPasswordChangeRequired())
				assert.Equal(t, testUsername, *us.PasswordChangeRequiredUsername)
				assert.Empty(t, us.Username)
			} else {
				assert.Equal(t, authentication.OneFactor, us.AuthenticationLevel(false))
				assert.False(t, us.IsPasswordChangeRequired())
				assert.Equal(t, testUsername, us.Username)
			}
		})
	}
}

func TestFirstFactorExternalIdentityCallbackGETShouldHoldAUserWhoMustChangeTheirPassword(t *testing.T) {
	key, server, _ := newTestOpenIDConnectUpstream(t)

	defer server.Close()

	mock := newTestOpenIDConnectCallbackMock(t)

	defer mock.Close()

	mock.Ctx.Providers.UserAttributeResolver = &expression.UserAttributes{}
	mock.Ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute = "pwd_reset"
	mock.Ctx.Providers.ExternalIdentity = newTestRelyingPartyProvidersWithUpstream(server.URL+"/token", key, false)
	mock.Ctx.SetUserValue("provider", "example")
	mock.Ctx.Request.SetRequestURI("/api/identity/example/callback?code=the-code&state=the-state")
	mock.Ctx.Request.SetHost("login.example.com:8080")

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	userSession.ExternalIdentity = &session.ExternalIdentityFlow{
		Provider:      "example",
		State:         "the-state",
		Nonce:         "the-nonce",
		CodeVerifier:  "the-verifier",
		TargetURL:     "https://app.example.com",
		RequestMethod: fasthttp.MethodGet,
		Expires:       mock.Ctx.GetClock().Now().Add(time.Minute * 3),
	}

	require.NoError(t, mock.Ctx.SaveSession(&userSession))

	mock.StorageMock.EXPECT().
		LoadExternalIdentityLinkBySubject(mock.Ctx, "openid_connect", "https://op.example.com", "abc123").
		Return(&model.ExternalIdentityLink{ID: 1, Provider: "example", Issuer: "https://op.example.com", Subject: "abc123", Username: "john"}, nil)
	mock.StorageMock.EXPECT().LoadBannedIP(mock.Ctx, model.NewIP(mock.Ctx.RemoteIP())).Return(nil, nil)
	mock.StorageMock.EXPECT().LoadBannedUser(mock.Ctx, "john").Return(nil, nil)
	mock.UserProviderMock.EXPECT().GetDetailsExtended("john").
		Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: "john"}, Extra: map[string]any{"pwd_reset": true}}, nil)
	mock.StorageMock.EXPECT().AppendAuthenticationLog(mock.Ctx, gomock.Any()).Return(nil)

	FirstFactorExternalIdentityCallbackGET(mock.Ctx)

	assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
	assert.NotEqual(t, "https://app.example.com/", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

	userSession, err = mock.Ctx.GetSession()
	require.NoError(t, err)

	assert.Equal(t, authentication.NotAuthenticated, userSession.AuthenticationLevel(false))
	assert.True(t, userSession.IsPasswordChangeRequired())
	assert.Equal(t, "john", *userSession.PasswordChangeRequiredUsername)
	assert.Empty(t, userSession.Username)
}

func TestFirstFactorPasswordPOSTShouldLookUpTheUserOnce(t *testing.T) {
	testCases := []struct {
		name     string
		flagged  bool
		expected authentication.Level
	}{
		{"ShouldAuthenticateWhenNotFlagged", false, authentication.OneFactor},
		{"ShouldHoldWhenFlagged", true, authentication.NotAuthenticated},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.UserAttributeResolver = &expression.UserAttributes{}
			mock.Ctx.Configuration.AuthenticationBackend.PasswordChange.RequiredAttribute = "pwd_reset"

			mock.UserProviderMock.EXPECT().CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).Return(true, nil)
			mock.UserProviderMock.EXPECT().GetDetailsExtended(gomock.Eq(testValue)).
				Return(&authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{Username: testValue, Emails: []string{"test@example.com"}}, Extra: map[string]any{"pwd_reset": tc.flagged}}, nil).
				Times(1)
			mock.StorageMock.EXPECT().LoadBannedIP(gomock.Any(), gomock.Any()).Return(nil, nil)
			mock.StorageMock.EXPECT().LoadBannedUser(gomock.Any(), gomock.Any()).Return(nil, nil)
			mock.StorageMock.EXPECT().AppendAuthenticationLog(gomock.Any(), gomock.Any()).Return(nil)

			mock.Ctx.Request.SetBodyString(`{"username": "test", "password": "hello"}`)

			FirstFactorPasswordPOST(nil)(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			assert.Equal(t, tc.expected, userSession.AuthenticationLevel(false))
			assert.Equal(t, tc.flagged, userSession.IsPasswordChangeRequired())
		})
	}
}

func TestStateGETShouldReportTheHeldUsername(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	userSession.SetPasswordChangeRequired(mock.Clock.Now(), testUsername)

	require.NoError(t, mock.Ctx.SaveSession(&userSession))

	StateGET(mock.Ctx)

	assert.Contains(t, string(mock.Ctx.Response.Body()), `"username":"`+testUsername+`"`)
	assert.Contains(t, string(mock.Ctx.Response.Body()), `"password_change_required":true`)
}
