package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/ory/fosite"
	fositejwt "github.com/ory/fosite/token/jwt"
	"github.com/pkg/errors"

	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/oidc"
)

func oidcUserinfo(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	session := newOpenIDSession("")

	tokenType, ar, err := ctx.Providers.OpenIDConnect.Fosite.IntrospectToken(req.Context(), fosite.AccessTokenFromRequest(req), fosite.AccessToken, session)
	if err != nil {
		rfc := fosite.ErrorToRFC6749Error(err)
		if rfc.StatusCode() == http.StatusUnauthorized {
			rw.Header().Set("WWW-Authenticate", fmt.Sprintf("error=%s,error_description=%s", rfc.ErrorField, rfc.GetDescription()))
		}

		ctx.Providers.OpenIDConnect.Herodot.WriteError(rw, req, err)

		return
	}

	if tokenType != fosite.AccessToken {
		errStr := "Authorization header must contain an OAuth access token."
		rw.Header().Set("WWW-Authenticate", fmt.Sprintf("error_description=\"%s\"", errStr))
		ctx.Providers.OpenIDConnect.Herodot.WriteErrorCode(rw, req, http.StatusUnauthorized, errors.New(errStr))

		return
	}

	client, ok := ar.GetClient().(*oidc.InternalClient)
	if !ok {
		ctx.Providers.OpenIDConnect.Herodot.WriteError(rw, req, errors.WithStack(fosite.ErrServerError.WithHint("Unable to assert type of client")))

		return
	}

	claims := ar.GetSession().(*oidc.OpenIDSession).IDTokenClaims().ToMap()
	delete(claims, "jti")
	delete(claims, "sid")
	delete(claims, "at_hash")
	delete(claims, "c_hash")
	delete(claims, "exp")
	delete(claims, "nonce")

	if audience, ok := claims["aud"].([]string); !ok || len(audience) == 0 {
		claims["aud"] = []string{client.GetID()}
	}

	if client.UserinfoAlgorithm == "RS256" {
		claims["jti"] = uuid.New()
		claims["iat"] = time.Now().Unix()

		keyID, err := ctx.Providers.OpenIDConnect.Store.KeyManager.Strategy().GetPublicKeyID(req.Context())
		if err != nil {
			ctx.Providers.OpenIDConnect.Herodot.WriteError(rw, req, err)

			return
		}

		token, _, err := ctx.Providers.OpenIDConnect.Store.KeyManager.Strategy().Generate(req.Context(), jwt.MapClaims(claims),
			&fositejwt.Headers{
				Extra: map[string]interface{}{"kid": keyID},
			})
		if err != nil {
			ctx.Providers.OpenIDConnect.Herodot.WriteError(rw, req, err)

			return
		}

		rw.Header().Set("Content-Type", "application/jwt")
		_, _ = rw.Write([]byte(token))
	} else if client.UserinfoAlgorithm == "" || client.UserinfoAlgorithm == "none" {
		ctx.Providers.OpenIDConnect.Herodot.Write(rw, req, claims)
	} else {
		ctx.Providers.OpenIDConnect.Herodot.WriteError(rw, req, errors.WithStack(fosite.ErrServerError.WithHintf("Unsupported userinfo signing algorithm '%s'.", client.UserinfoAlgorithm)))
	}
}
