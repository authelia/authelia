package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/token/jwt"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OpenIDConnectUserinfo handles GET/POST requests to the OpenID Connect 1.0 UserInfo endpoint.
//
// https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
func OpenIDConnectUserinfo(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		tokenType fosite.TokenType
		requester fosite.AccessRequester
		client    oidc.Client
		err       error
	)

	oidcSession := oidc.NewSession()

	if tokenType, requester, err = ctx.Providers.OpenIDConnect.IntrospectToken(
		req.Context(), fosite.AccessTokenFromRequest(req), fosite.AccessToken, oidcSession); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("UserInfo Request failed with error: %s", rfc.WithExposeDebug(true).GetDescription())

		if rfc.StatusCode() == http.StatusUnauthorized {
			rw.Header().Set(fasthttp.HeaderWWWAuthenticate, fmt.Sprintf(`Bearer error="%s",error_description="%s"`, rfc.ErrorField, rfc.GetDescription()))
		}

		ctx.Providers.OpenIDConnect.WriteError(rw, req, err)

		return
	}

	clientID := requester.GetClient().GetID()

	if tokenType != fosite.AccessToken {
		ctx.Logger.Errorf("UserInfo Request with id '%s' on client with id '%s' failed with error: bearer authorization failed as the token is not an access token", requester.GetID(), client.GetID())

		errStr := "Only access tokens are allowed in the authorization header."
		rw.Header().Set(fasthttp.HeaderWWWAuthenticate, fmt.Sprintf(`Bearer error="invalid_token",error_description="%s"`, errStr))
		ctx.Providers.OpenIDConnect.WriteErrorCode(rw, req, http.StatusUnauthorized, errors.New(errStr))

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetFullClient(clientID); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("UserInfo Request with id '%s' on client with id '%s' failed to retrieve client configuration with error: %s", requester.GetID(), client.GetID(), rfc.WithExposeDebug(true).GetDescription())

		ctx.Providers.OpenIDConnect.WriteError(rw, req, errors.WithStack(rfc))

		return
	}

	claims := requester.GetSession().(*model.OpenIDSession).IDTokenClaims().ToMap()
	delete(claims, oidc.ClaimJWTID)
	delete(claims, oidc.ClaimSessionID)
	delete(claims, oidc.ClaimAccessTokenHash)
	delete(claims, oidc.ClaimCodeHash)
	delete(claims, oidc.ClaimExpirationTime)
	delete(claims, oidc.ClaimNonce)

	audience, ok := claims[oidc.ClaimAudience].([]string)

	if !ok || len(audience) == 0 {
		audience = []string{client.GetID()}
	} else {
		found := false

		for _, aud := range audience {
			if aud == clientID {
				found = true
				break
			}
		}

		if found {
			audience = append(audience, clientID)
		}
	}

	claims[oidc.ClaimAudience] = audience

	var token string

	ctx.Logger.Tracef("UserInfo Response with id '%s' on client with id '%s' is being sent with the following claims: %+v", requester.GetID(), clientID, claims)

	switch client.GetUserinfoSigningAlgorithm() {
	case oidc.SigningAlgorithmRSAWithSHA256:
		var jti uuid.UUID

		if jti, err = uuid.NewRandom(); err != nil {
			ctx.Providers.OpenIDConnect.WriteError(rw, req, fosite.ErrServerError.WithHint("Could not generate JTI."))

			return
		}

		claims[oidc.ClaimJWTID] = jti.String()
		claims[oidc.ClaimIssuedAt] = time.Now().Unix()

		headers := &jwt.Headers{
			Extra: map[string]any{
				oidc.JWTHeaderKeyIdentifier: ctx.Providers.OpenIDConnect.KeyManager.GetActiveKeyID(),
			},
		}

		if token, _, err = ctx.Providers.OpenIDConnect.KeyManager.Strategy().Generate(req.Context(), claims, headers); err != nil {
			ctx.Providers.OpenIDConnect.WriteError(rw, req, err)

			return
		}

		rw.Header().Set(fasthttp.HeaderContentType, "application/jwt")
		_, _ = rw.Write([]byte(token))
	case oidc.SigningAlgorithmNone, "":
		ctx.Providers.OpenIDConnect.Write(rw, req, claims)
	default:
		ctx.Providers.OpenIDConnect.WriteError(rw, req, errors.WithStack(fosite.ErrServerError.WithHintf("Unsupported UserInfo signing algorithm '%s'.", client.GetUserinfoSigningAlgorithm())))
	}
}
