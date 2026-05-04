package oidc_test

import (
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

func TestOpenIDSession_GetExtraClaims(t *testing.T) {
	testCases := []struct {
		name     string
		have     *oidc.Session
		expected map[string]any
	}{
		{
			"ShouldReturnNil",
			&oidc.Session{},
			nil,
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			i := tc.have.GetJWTClaims()

			actual, ok := i.(*jwt.JWTClaims)

			require.True(t, ok)

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
			}

			oidc.ConsentGrant(consent, tc.explicit, tc.claims)

			assert.Equal(t, model.StringSlicePipeDelimited(tc.expectedScopes), consent.GrantedScopes)
			assert.Equal(t, model.StringSlicePipeDelimited{"https://example.com"}, consent.GrantedAudience)

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
