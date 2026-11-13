// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	fjwt "authelia.com/provider/oauth2/token/jwt"
)

// IDTokenSubjectValidationStrategy wraps an openid.TokenValidationStrategy and additionally requires the claims that
// Authelia issues in ID Tokens and never in the other JWTs it signs with the same keys (such as signed UserInfo
// responses), so that only genuine ID Tokens are accepted as RFC 8693 subject tokens.
type IDTokenSubjectValidationStrategy struct {
	openid.TokenValidationStrategy
}

// ValidateIDToken implements openid.TokenValidationStrategy.
func (s *IDTokenSubjectValidationStrategy) ValidateIDToken(ctx context.Context, request oauthelia2.Requester, token string, opts ...oauthelia2.IDTokenValidationOpt) (claims fjwt.MapClaims, err error) {
	if claims, err = s.TokenValidationStrategy.ValidateIDToken(ctx, request, token, opts...); err != nil {
		return nil, err
	}

	if _, ok := claims[ClaimAuthenticationTime]; !ok {
		return nil, oauthelia2.ErrInvalidRequest.WithHint("The token is not an ID Token issued by this authorization server.")
	}

	return claims, nil
}

var _ openid.TokenValidationStrategy = (*IDTokenSubjectValidationStrategy)(nil)
