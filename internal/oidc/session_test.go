// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"authelia.com/provider/oauth2/handler/openid"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOpenIDSession(t *testing.T) {
	session := &oidc.Session{
		DefaultSession: &openid.DefaultSession{},
	}

	assert.Nil(t, session.GetIDTokenClaims())
	assert.NotNil(t, session.Clone())

	session = nil

	assert.Nil(t, session.Clone())
}

func TestSession_Clone(t *testing.T) {
	session := &oidc.Session{
		DefaultSession: &openid.DefaultSession{
			Claims: &jwt.IDTokenClaims{
				Subject: "abc",
				Extra:   map[string]any{"a": "b"},
			},
			Headers:  &jwt.Headers{Extra: map[string]any{"kid": "123"}},
			Username: "john",
		},
		AccessToken: &oidc.AccessTokenSession{
			Headers: map[string]any{"typ": "at+jwt"},
			Claims:  map[string]any{"iss": "https://auth.example.com"},
		},
		ChallengeID:           uuid.NullUUID{UUID: uuid.MustParse("0b8a1b7e-1f1e-4c5c-9f4c-3c6f1f8b2d5a"), Valid: true},
		ClientID:              "client",
		ClientCredentials:     true,
		ExcludeNotBeforeClaim: true,
		AllowedTopLevelClaims: []string{"email"},
		ClaimRequests: &oidc.ClaimsRequests{
			IDToken:  map[string]*oidc.ClaimRequest{"email": {Essential: true, Values: []any{"a"}}, "nil": nil},
			UserInfo: map[string]*oidc.ClaimRequest{"name": {Value: "john"}},
		},
		GrantedClaims: []string{"email"},
		Extra:         map[string]any{"x": "y"},
	}

	clone, ok := session.Clone().(*oidc.Session)
	require.True(t, ok)

	assert.Equal(t, session, clone)

	clone.Claims.Extra["a"] = "changed"
	clone.Headers.Extra["kid"] = "changed"
	clone.AccessToken.Headers["typ"] = "changed"
	clone.AccessToken.Claims["iss"] = "changed"
	clone.AllowedTopLevelClaims[0] = "changed"
	clone.ClaimRequests.IDToken["email"].Values[0] = "changed"
	clone.ClaimRequests.UserInfo["name"].Value = "changed"
	clone.GrantedClaims[0] = "changed"
	clone.Extra["x"] = "changed"

	assert.Equal(t, "b", session.Claims.Extra["a"])
	assert.Equal(t, "123", session.Headers.Extra["kid"])
	assert.Equal(t, "at+jwt", session.AccessToken.Headers["typ"])
	assert.Equal(t, "https://auth.example.com", session.AccessToken.Claims["iss"])
	assert.Equal(t, "email", session.AllowedTopLevelClaims[0])
	assert.Equal(t, "a", session.ClaimRequests.IDToken["email"].Values[0])
	assert.Equal(t, "john", session.ClaimRequests.UserInfo["name"].Value)
	assert.Equal(t, "email", session.GrantedClaims[0])
	assert.Equal(t, "y", session.Extra["x"])

	clone, ok = (&oidc.Session{}).Clone().(*oidc.Session)
	require.True(t, ok)

	assert.Equal(t, &oidc.Session{}, clone)
}

func TestSession_ValidIssuer(t *testing.T) {
	issuer := &url.URL{Scheme: "https", Host: "auth.example.com"}

	testCases := []struct {
		name     string
		have     *oidc.Session
		issuer   *url.URL
		expected bool
	}{
		{
			"ShouldReturnTrueWhenSessionNil",
			nil,
			issuer,
			true,
		},
		{
			"ShouldReturnTrueWhenDefaultSessionNil",
			&oidc.Session{},
			issuer,
			true,
		},
		{
			"ShouldReturnTrueWhenClaimsNil",
			&oidc.Session{DefaultSession: &openid.DefaultSession{}},
			issuer,
			true,
		},
		{
			"ShouldReturnTrueWhenClaimsIssuerEmpty",
			&oidc.Session{DefaultSession: &openid.DefaultSession{Claims: &jwt.IDTokenClaims{}}},
			issuer,
			true,
		},
		{
			"ShouldReturnTrueWhenClaimsIssuerMatches",
			&oidc.Session{DefaultSession: &openid.DefaultSession{Claims: &jwt.IDTokenClaims{Issuer: "https://auth.example.com"}}},
			issuer,
			true,
		},
		{
			"ShouldReturnFalseWhenClaimsIssuerMismatches",
			&oidc.Session{DefaultSession: &openid.DefaultSession{Claims: &jwt.IDTokenClaims{Issuer: "https://other.example.com"}}},
			issuer,
			false,
		},
		{
			"ShouldReturnFalseWhenClaimsIssuerTrailingSlashDiffers",
			&oidc.Session{DefaultSession: &openid.DefaultSession{Claims: &jwt.IDTokenClaims{Issuer: "https://auth.example.com/"}}},
			issuer,
			false,
		},
		{
			"ShouldReturnFalseWhenClaimsIssuerSchemeDiffers",
			&oidc.Session{DefaultSession: &openid.DefaultSession{Claims: &jwt.IDTokenClaims{Issuer: "http://auth.example.com"}}},
			issuer,
			false,
		},
		{
			"ShouldReturnFalseWhenIssuerNil",
			&oidc.Session{DefaultSession: &openid.DefaultSession{Claims: &jwt.IDTokenClaims{Issuer: "https://auth.example.com"}}},
			nil,
			false,
		},
		{
			"ShouldReturnFalseWhenIssuerNilAndClaimsIssuerEmpty",
			&oidc.Session{DefaultSession: &openid.DefaultSession{Claims: &jwt.IDTokenClaims{}}},
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.ValidIssuer(tc.issuer))
		})
	}
}

func TestOpenIDSession_GetExtraClaims(t *testing.T) {
	testCases := []struct {
		name     string
		have     *oidc.Session
		expected map[string]any
	}{
		{
			"ShouldReturnEmptyMap",
			&oidc.Session{},
			map[string]any{},
		},
		{
			"ShouldReturnExtra",
			&oidc.Session{
				AccessToken: &oidc.AccessTokenSession{
					Claims: map[string]any{
						"a": 1,
					},
				},
			},
			map[string]any{
				"a": 1,
			},
		},
		{
			"ShouldNotReturnIDTokenClaimsExtra",
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{
					Claims: &jwt.IDTokenClaims{
						Extra: map[string]any{
							"b": 2,
						},
					},
				},
				AccessToken: &oidc.AccessTokenSession{
					Claims: map[string]any{
						"a": 1,
					},
				},
			},
			map[string]any{
				"a": 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.GetExtraClaims())
		})
	}
}

func TestSession_GetJWTHeader(t *testing.T) {
	testCases := []struct {
		name     string
		have     *oidc.Session
		expected *jwt.Headers
	}{
		{
			"ShouldReturnDefaults",
			&oidc.Session{DefaultSession: openid.NewDefaultSession()},
			&jwt.Headers{Extra: map[string]any{oidc.JWTHeaderKeyType: oidc.JWTHeaderTypeValueAccessTokenJWT}},
		},
		{
			"ShouldMergeAccessTokenHeaders",
			&oidc.Session{
				DefaultSession: openid.NewDefaultSession(),
				AccessToken: &oidc.AccessTokenSession{
					Headers: map[string]any{"kid": "my-key-id", "custom": "value"},
				},
			},
			&jwt.Headers{Extra: map[string]any{oidc.JWTHeaderKeyType: oidc.JWTHeaderTypeValueAccessTokenJWT, "kid": "my-key-id", "custom": "value"}},
		},
		{
			"ShouldOverrideDefaultWithAccessTokenHeaders",
			&oidc.Session{
				DefaultSession: openid.NewDefaultSession(),
				AccessToken: &oidc.AccessTokenSession{
					Headers: map[string]any{oidc.JWTHeaderKeyType: "custom-type"},
				},
			},
			&jwt.Headers{Extra: map[string]any{oidc.JWTHeaderKeyType: "custom-type"}},
		},
		{
			"ShouldHandleEmptyAccessTokenHeaders",
			&oidc.Session{
				DefaultSession: openid.NewDefaultSession(),
				AccessToken:    &oidc.AccessTokenSession{Headers: map[string]any{}},
			},
			&jwt.Headers{Extra: map[string]any{oidc.JWTHeaderKeyType: oidc.JWTHeaderTypeValueAccessTokenJWT}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.GetJWTHeader())
		})
	}
}

func TestSession_GetJWTClaims(t *testing.T) {
	testCases := []struct {
		name     string
		have     *oidc.Session
		expected *jwt.JWTClaims
	}{
		{
			"ShouldReturnDefaults",
			&oidc.Session{DefaultSession: openid.NewDefaultSession()},
			&jwt.JWTClaims{Extra: map[string]any{}},
		},
		{
			"ShouldIncludeClientID",
			&oidc.Session{DefaultSession: openid.NewDefaultSession(), ClientID: abc},
			&jwt.JWTClaims{Extra: map[string]any{oidc.ClaimClientIdentifier: abc}},
		},
		{
			"ShouldAllowTopLevelClaims",
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{
					Claims: &jwt.IDTokenClaims{
						Extra: map[string]any{"test": 1},
					},
					RequestedAt: time.Now(),
					Headers:     &jwt.Headers{},
				},
				AccessToken: &oidc.AccessTokenSession{
					Claims: map[string]any{oidc.ClaimClientIdentifier: "x", "test": 1},
				},
				ClientID:              abc,
				AllowedTopLevelClaims: []string{oidc.ClaimClientIdentifier},
			},
			&jwt.JWTClaims{
				Extra: map[string]any{
					oidc.ClaimClientIdentifier: abc,
					"test":                     1,
				},
			},
		},
		{
			"ShouldAllowTopLevelClaimsAlt",
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{
					Claims: &jwt.IDTokenClaims{
						Extra: map[string]any{"test": 2},
					},
					RequestedAt: time.Now(),
					Headers:     &jwt.Headers{},
				},
				AccessToken: &oidc.AccessTokenSession{
					Claims: map[string]any{oidc.ClaimClientIdentifier: "x"},
				},
				ClientID:              abc,
				AllowedTopLevelClaims: []string{oidc.ClaimClientIdentifier, "test"},
			},
			&jwt.JWTClaims{Extra: map[string]any{oidc.ClaimClientIdentifier: abc, "test": 2}},
		},
		{
			"ShouldNotIncludeAMR",
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{
					Claims: &jwt.IDTokenClaims{
						Extra: map[string]any{oidc.ClaimAuthenticationMethodsReference: []string{oidc.AMRMultiFactorAuthentication}},
					},
					RequestedAt: time.Now(),
					Headers:     &jwt.Headers{},
				},
				AccessToken: &oidc.AccessTokenSession{
					Claims: map[string]any{},
				},
				ClientID:              abc,
				AllowedTopLevelClaims: []string{oidc.ClaimClientIdentifier},
			},
			&jwt.JWTClaims{Extra: map[string]any{oidc.ClaimClientIdentifier: abc}},
		},
		{
			"ShouldNotIncludeAMRAbsent",
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{
					Claims: &jwt.IDTokenClaims{
						Extra: map[string]any{},
					},
					RequestedAt: time.Now(),
					Headers:     &jwt.Headers{},
				},
				AccessToken: &oidc.AccessTokenSession{
					Claims: map[string]any{},
				},
				ClientID:              abc,
				AllowedTopLevelClaims: []string{oidc.ClaimClientIdentifier, oidc.ClaimAuthenticationMethodsReference},
			},
			&jwt.JWTClaims{Extra: map[string]any{oidc.ClaimClientIdentifier: abc}},
		},
		{
			"ShouldIncludeAMR",
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{
					Claims: &jwt.IDTokenClaims{
						AuthenticationMethodsReferences: []string{oidc.AMRMultiFactorAuthentication},
						Extra:                           map[string]any{},
					},
					RequestedAt: time.Now(),
					Headers:     &jwt.Headers{},
				},
				AccessToken: &oidc.AccessTokenSession{
					Claims: map[string]any{},
				},
				ClientID:              abc,
				AllowedTopLevelClaims: []string{oidc.ClaimClientIdentifier, oidc.ClaimAuthenticationMethodsReference},
			},
			&jwt.JWTClaims{Extra: map[string]any{oidc.ClaimAuthenticationMethodsReference: []string{oidc.AMRMultiFactorAuthentication}, oidc.ClaimClientIdentifier: abc}},
		},

		{
			"ShouldSetClientCredentialsSubjectFromClientID",
			&oidc.Session{DefaultSession: openid.NewDefaultSession(), ClientID: abc, ClientCredentials: true},
			&jwt.JWTClaims{Subject: abc, Extra: map[string]any{oidc.ClaimClientIdentifier: abc}},
		},
		{
			"ShouldNotOverrideClientCredentialsSubject",
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{
					Subject:     "existing",
					Claims:      &jwt.IDTokenClaims{Extra: map[string]any{}},
					Headers:     &jwt.Headers{},
					RequestedAt: time.Now(),
				},
				ClientID:          abc,
				ClientCredentials: true,
			},
			&jwt.JWTClaims{Subject: "existing", Extra: map[string]any{oidc.ClaimClientIdentifier: abc}},
		},
		{
			"ShouldNotSetSubjectFromClientIDWhenNotClientCredentials",
			&oidc.Session{DefaultSession: openid.NewDefaultSession(), ClientID: abc},
			&jwt.JWTClaims{Extra: map[string]any{oidc.ClaimClientIdentifier: abc}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			i := tc.have.GetJWTClaims()

			actual, ok := i.(*jwt.JWTClaims)

			require.True(t, ok)

			assert.Equal(t, tc.expected.Subject, actual.Subject)
			assert.Equal(t, tc.expected.JTI, actual.JTI)
			assert.Equal(t, tc.expected.Audience, actual.Audience)
			assert.Equal(t, tc.expected.Issuer, actual.Issuer)
			assert.Equal(t, tc.expected.ExpiresAt, actual.ExpiresAt)
			assert.Equal(t, tc.expected.Extra, actual.Extra)
			assert.Equal(t, tc.expected.Scope, actual.Scope)
			assert.Equal(t, tc.expected.ScopeField, actual.ScopeField)
		})
	}
}

func TestSession_GetIDTokenClaims(t *testing.T) {
	testCases := []struct {
		name     string
		have     *oidc.Session
		expected *jwt.IDTokenClaims
	}{
		{
			"ShouldReturnNil",
			&oidc.Session{},
			nil,
		},
		{
			"ShouldReturnClaimsNil",
			&oidc.Session{DefaultSession: &openid.DefaultSession{}},
			nil,
		},
		{
			"ShouldReturnActualClaims",
			&oidc.Session{DefaultSession: &openid.DefaultSession{Claims: &jwt.IDTokenClaims{JTI: "example"}}},
			&jwt.IDTokenClaims{JTI: "example"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.GetIDTokenClaims())
		})
	}
}

func TestConsentGrant(t *testing.T) {
	testCases := []struct {
		name            string
		explicit        bool
		requestedScopes []string
		claims          []string
		expectedScopes  []string
		expectedClaims  []string
	}{
		{
			"ShouldGrantAllScopesWhenExplicit",
			true,
			[]string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, oidc.ScopeProfile},
			[]string{"email", "name"},
			[]string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, oidc.ScopeProfile},
			[]string{"email", "name"},
		},
		{
			"ShouldSkipOfflineScopesWhenImplicit",
			false,
			[]string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, oidc.ScopeProfile},
			nil,
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile},
			nil,
		},
		{
			"ShouldSkipOfflineScopeWhenImplicit",
			false,
			[]string{oidc.ScopeOpenID, "offline", oidc.ScopeProfile},
			nil,
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile},
			nil,
		},
		{
			"ShouldGrantAllNonOfflineScopesWhenImplicit",
			false,
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail},
			[]string{"custom"},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail},
			[]string{"custom"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			consent := &model.OAuth2ConsentSession{
				RequestedScopes:   tc.requestedScopes,
				RequestedAudience: model.StringSlicePipeDelimited{"https://example.com"},
				RequestedResource: model.StringSlicePipeDelimited{"https://api.example.com"},
			}

			oidc.ConsentGrant(consent, tc.explicit, tc.claims)

			assert.Equal(t, model.StringSlicePipeDelimited(tc.expectedScopes), consent.GrantedScopes)
			assert.Equal(t, model.StringSlicePipeDelimited{"https://example.com"}, consent.GrantedAudience)
			assert.Equal(t, model.StringSlicePipeDelimited{"https://api.example.com"}, consent.GrantedResource)

			if tc.expectedClaims != nil {
				assert.Equal(t, model.StringSlicePipeDelimited(tc.expectedClaims), consent.GrantedClaims)
			}
		})
	}
}

func TestConsentGrantImplicit(t *testing.T) {
	testCases := []struct {
		name            string
		requestedScopes []string
		claims          []string
		expectedScopes  []string
	}{
		{
			"ShouldSetSubjectAndRespondedAtThenGrant",
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile},
			[]string{"email"},
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile},
		},
		{
			"ShouldSkipOfflineAccessScopes",
			[]string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, oidc.ScopeProfile},
			nil,
			[]string{oidc.ScopeOpenID, oidc.ScopeProfile},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subject := uuid.MustParse("fb1bdb5e-96b3-4c04-b7a3-3e532b4d2e70")
			respondedAt := time.Now()

			consent := &model.OAuth2ConsentSession{
				RequestedScopes:   tc.requestedScopes,
				RequestedAudience: model.StringSlicePipeDelimited{"https://example.com"},
			}

			oidc.ConsentGrantImplicit(consent, tc.claims, subject, respondedAt)

			assert.Equal(t, model.StringSlicePipeDelimited(tc.expectedScopes), consent.GrantedScopes)
			assert.Equal(t, model.StringSlicePipeDelimited{"https://example.com"}, consent.GrantedAudience)
			assert.True(t, consent.Subject.Valid)
			assert.Equal(t, subject, consent.Subject.UUID)
			assert.True(t, consent.RespondedAt.Valid)
		})
	}
}

func TestSession_GetStorageSubject(t *testing.T) {
	testCases := []struct {
		name     string
		have     *oidc.Session
		expected string
	}{
		{
			"ShouldReturnEmptyWhenSessionNil",
			nil,
			"",
		},
		{
			"ShouldReturnEmptyWhenDefaultSessionNil",
			&oidc.Session{},
			"",
		},
		{
			"ShouldReturnSubject",
			&oidc.Session{DefaultSession: &openid.DefaultSession{Subject: "john"}},
			"john",
		},
		{
			"ShouldNotReturnClientIDForClientCredentials",
			&oidc.Session{ClientID: "example", ClientCredentials: true, DefaultSession: &openid.DefaultSession{Claims: &jwt.IDTokenClaims{Subject: "example"}}},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.GetStorageSubject())
		})
	}
}

func TestSession_SetID(t *testing.T) {
	testCases := []struct {
		name string
		have *oidc.Session
	}{
		{"ShouldInitializeAnEmptySession", &oidc.Session{}},
		{"ShouldInitializeAbsentClaimsAndHeaders", &oidc.Session{DefaultSession: &openid.DefaultSession{}}},
		{"ShouldInitializeAbsentClaimsExtra", &oidc.Session{DefaultSession: &openid.DefaultSession{Claims: &jwt.IDTokenClaims{}, Headers: &jwt.Headers{}}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.have.SetID("session-sid-value")

			require.NotNil(t, tc.have.DefaultSession)
			require.NotNil(t, tc.have.Claims)
			require.NotNil(t, tc.have.Headers)

			assert.Equal(t, "session-sid-value", tc.have.Claims.SessionID)
			assert.NotNil(t, tc.have.Claims.Extra, "a nil claims map panics on the first extra claim written to it")
			assert.NotNil(t, tc.have.Headers.Extra, "a nil headers map panics on the first extra header written to it")

			assert.NotPanics(t, func() {
				tc.have.Claims.Extra["example"] = "value"
				tc.have.Headers.Extra["example"] = "value"
			})
		})
	}

	t.Run("ShouldPreserveExistingClaims", func(t *testing.T) {
		session := oidc.NewSession()

		session.Claims.Subject = "john"
		session.Claims.Extra["preferred_username"] = "john"
		session.Headers.Extra["kid"] = "abc"

		session.SetID("session-sid-value")

		assert.Equal(t, "session-sid-value", session.Claims.SessionID)
		assert.Equal(t, "john", session.Claims.Subject)
		assert.Equal(t, "john", session.Claims.Extra["preferred_username"])
		assert.Equal(t, "abc", session.Headers.Extra["kid"])
	})

	t.Run("ShouldNotPanicOnNilSession", func(t *testing.T) {
		var session *oidc.Session

		assert.NotPanics(t, func() { session.SetID("session-sid-value") })
	})
}

// TestSession_SidJSONRoundTrip asserts that the 'sid' claim set via Session.SetID survives being serialized into
// the persisted session and read back, which is the mechanism a refresh grant relies on to reuse the sid minted at
// authorization time instead of the token endpoint rebuilding the session from storage without it. It asserts the
// claim lands at the specific JSON path the storage layer persists ('id_token.id_token_claims.sid') so a future
// change to the embedding or struct tags that moved or dropped the field would fail this test, and it asserts the
// key is omitted entirely when unset, which is what keeps the sid claim out of tokens for non-openid flows.
func TestSession_SidJSONRoundTrip(t *testing.T) {
	t.Run("ShouldRoundTripSidThroughStorageSerialization", func(t *testing.T) {
		session := oidc.NewSession()
		session.SetID("session-sid-value")

		data, err := json.Marshal(session)
		require.NoError(t, err)

		var raw map[string]any

		require.NoError(t, json.Unmarshal(data, &raw))

		idToken, ok := raw["id_token"].(map[string]any)
		require.True(t, ok, "expected an 'id_token' object in the serialized session")

		claims, ok := idToken["id_token_claims"].(map[string]any)
		require.True(t, ok, "expected an 'id_token_claims' object nested under 'id_token'")

		assert.Equal(t, "session-sid-value", claims["sid"])

		restored := oidc.NewSession()

		require.NoError(t, json.Unmarshal(data, restored))

		assert.Equal(t, "session-sid-value", restored.GetID())
	})

	t.Run("ShouldNotEmitSidWhenUnset", func(t *testing.T) {
		session := oidc.NewSession()

		data, err := json.Marshal(session)
		require.NoError(t, err)

		var raw map[string]any

		require.NoError(t, json.Unmarshal(data, &raw))

		idToken, ok := raw["id_token"].(map[string]any)
		require.True(t, ok, "expected an 'id_token' object in the serialized session")

		claims, ok := idToken["id_token_claims"].(map[string]any)
		require.True(t, ok, "expected an 'id_token_claims' object nested under 'id_token'")

		_, present := claims["sid"]
		assert.False(t, present, "the 'sid' key must be omitted entirely when no session id is set")

		restored := oidc.NewSession()

		require.NoError(t, json.Unmarshal(data, restored))

		assert.Equal(t, "", restored.GetID())
	})
}

func TestNewSessionWithClientAndRequestedAt(t *testing.T) {
	requestedAt := time.Unix(1700000000, 0)

	session := oidc.NewSessionWithClientAndRequestedAt(&oidc.RegisteredClient{ID: "client-id"}, requestedAt)

	assert.Equal(t, "client-id", session.ClientID)
	assert.Equal(t, requestedAt.UTC(), session.RequestedAt)
	assert.Empty(t, session.GetID())

	session = oidc.NewSessionWithClientAndRequestedAt(nil, requestedAt)

	assert.Empty(t, session.ClientID)
}
