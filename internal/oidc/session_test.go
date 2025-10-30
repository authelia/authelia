package oidc_test

import (
	"testing"
	"time"

	"authelia.com/provider/oauth2/handler/openid"
	"authelia.com/provider/oauth2/token/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
