package oidc_test

import (
	"testing"

	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"

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
				Extra: map[string]any{
					"a": 1,
				},
			},
			map[string]any{
				"a": 1,
			},
		},
		{
			"ShouldReturnIDTokenClaimsExtra",
			&oidc.Session{
				DefaultSession: &openid.DefaultSession{
					Claims: &jwt.IDTokenClaims{
						Extra: map[string]any{
							"b": 2,
						},
					},
				},
				Extra: map[string]any{
					"a": 1,
				},
			},
			map[string]any{
				"b": 2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.GetExtraClaims())
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
