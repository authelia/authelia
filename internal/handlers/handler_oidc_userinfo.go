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
		client    *oidc.Client
		err       error
	)

	oidcSession := oidc.NewSession()

	if tokenType, requester, err = ctx.Providers.OpenIDConnect.Fosite.IntrospectToken(
		req.Context(), fosite.AccessTokenFromRequest(req), fosite.AccessToken, oidcSession); err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)

		ctx.Logger.Errorf("UserInfo Request failed with error: %+v", rfc)

		if rfc.StatusCode() == http.StatusUnauthorized {
			rw.Header().Set(fasthttp.HeaderWWWAuthenticate, fmt.Sprintf(`Bearer error="%s",error_description="%s"`, rfc.ErrorField, rfc.GetDescription()))
		}

		ctx.Providers.OpenIDConnect.WriteError(rw, req, err)

		return
	}

	clientID := requester.GetClient().GetID()

	if tokenType != fosite.AccessToken {
		ctx.Logger.Errorf("UserInfo Request with id '%s' on client with id '%s' failed with error: bearer authorization failed as the token is not an access_token", requester.GetID(), client.GetID())

		errStr := "Only access tokens are allowed in the authorization header."
		rw.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer error="invalid_token",error_description="%s"`, errStr))
		ctx.Providers.OpenIDConnect.WriteErrorCode(rw, req, http.StatusUnauthorized, errors.New(errStr))

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.Store.GetFullClient(clientID); err != nil {
		ctx.Providers.OpenIDConnect.WriteError(rw, req, errors.WithStack(fosite.ErrServerError.WithHint("Unable to assert type of client")))

		return
	}

	claims := requester.GetSession().(*model.OpenIDSession).IDTokenClaims().ToMap()
	delete(claims, "jti")
	delete(claims, "sid")
	delete(claims, "at_hash")
	delete(claims, "c_hash")
	delete(claims, "exp")
	delete(claims, "nonce")

	audience, ok := claims["aud"].([]string)

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

	claims["aud"] = audience

	var token string

	ctx.Logger.Tracef("UserInfo Response with id '%s' on client with id '%s' is being sent with the following claims: %+v", requester.GetID(), clientID, claims)

	switch client.UserinfoSigningAlgorithm {
	case "RS256":
		var jti uuid.UUID

		if jti, err = uuid.NewRandom(); err != nil {
			ctx.Providers.OpenIDConnect.WriteError(rw, req, fosite.ErrServerError.WithHintf("Could not generate JTI."))

			return
		}

		claims["jti"] = jti.String()
		claims["iat"] = time.Now().Unix()

		headers := &jwt.Headers{
			Extra: map[string]any{"kid": ctx.Providers.OpenIDConnect.KeyManager.GetActiveKeyID()},
		}

		if token, _, err = ctx.Providers.OpenIDConnect.KeyManager.Strategy().Generate(req.Context(), claims, headers); err != nil {
			ctx.Providers.OpenIDConnect.WriteError(rw, req, err)

			return
		}

		rw.Header().Set("Content-Type", "application/jwt")
		_, _ = rw.Write([]byte(token))
	case "none", "":
		ctx.Providers.OpenIDConnect.Write(rw, req, claims)
	default:
		ctx.Providers.OpenIDConnect.WriteError(rw, req, errors.WithStack(fosite.ErrServerError.WithHintf("Unsupported UserInfo signing algorithm '%s'.", client.UserinfoSigningAlgorithm)))
	}
}
