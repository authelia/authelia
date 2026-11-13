// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"
	fjwt "authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/oidc"
)

type testIDTokenValidationStrategy struct {
	claims fjwt.MapClaims
	err    error
}

func (s *testIDTokenValidationStrategy) ValidateIDToken(_ context.Context, _ oauthelia2.Requester, _ string, _ ...oauthelia2.IDTokenValidationOpt) (claims fjwt.MapClaims, err error) {
	return s.claims, s.err
}

func TestIDTokenSubjectValidationStrategy(t *testing.T) {
	errUnderlying := errors.New("underlying failure")

	testCases := []struct {
		Name     string
		Claims   fjwt.MapClaims
		Err      error
		Expected string
		Hint     string
	}{
		{Name: "ShouldAcceptIDTokenClaims", Claims: fjwt.MapClaims{"sub": "a", "iat": 1, "exp": 2, "auth_time": 1}},
		{Name: "ShouldRejectUserInfoClaims", Claims: fjwt.MapClaims{"sub": "a", "iat": 1, "exp": 2, "jti": "b"}, Hint: "The token is not an ID Token issued by this authorization server."},
		{Name: "ShouldRejectNilClaims", Hint: "The token is not an ID Token issued by this authorization server."},
		{Name: "ShouldPropagateUnderlyingError", Err: errUnderlying, Expected: "underlying failure"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			strategy := &oidc.IDTokenSubjectValidationStrategy{TokenValidationStrategy: &testIDTokenValidationStrategy{claims: tc.Claims, err: tc.Err}}

			claims, err := strategy.ValidateIDToken(context.Background(), nil, "token")

			switch {
			case tc.Expected != "":
				assert.EqualError(t, err, tc.Expected)
				assert.Nil(t, claims)
			case tc.Hint != "":
				require.Error(t, err)

				var rfc *oauthelia2.RFC6749Error

				require.True(t, errors.As(err, &rfc))

				assert.Equal(t, "invalid_request", rfc.ErrorField)
				assert.Equal(t, tc.Hint, rfc.HintField)
				assert.Nil(t, claims)
			default:
				assert.NoError(t, err)
				assert.Equal(t, tc.Claims, claims)
			}
		})
	}
}
