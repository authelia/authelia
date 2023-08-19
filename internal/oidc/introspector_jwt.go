package oidc

import (
	"context"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/token/jwt"
	"github.com/ory/x/errorsx"
)

type StatelessJWTValidator struct {
	jwt.Signer
	Config interface {
		fosite.ScopeStrategyProvider
	}
}

func (v *StatelessJWTValidator) IntrospectToken(ctx context.Context, token string, tokenUse fosite.TokenUse, accessRequest fosite.AccessRequester, scopes []string) (fosite.TokenUse, error) {
	if !isJWT(token) {
		return "", fosite.ErrUnknownRequest.WithDebug("The provided token appears to be an opaque token not a JWT.")
	}

	t, err := jwtValidate(ctx, v.Signer, token)
	if err != nil {
		return "", err
	}

	if !IsJWTProfileAccessToken(t) {
		return "", errorsx.WithStack(fosite.ErrRequestUnauthorized.WithDebug("The provided token is not a valid RFC9068 JWT Profile Access Token as it is missing the header 'typ' value of 'at+jwt'."))
	}

	requester := oauth2.AccessTokenJWTToRequest(t)

	if err = MatchScopes(v.Config.GetScopeStrategy(ctx), requester.GetGrantedScopes(), scopes); err != nil {
		return fosite.AccessToken, err
	}

	accessRequest.Merge(requester)

	return fosite.AccessToken, nil
}
