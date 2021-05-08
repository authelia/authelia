package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/ory/fosite"
	fositejwt "github.com/ory/fosite/token/jwt"

	"github.com/authelia/authelia/internal/middlewares"
)

func oidcUserinfo(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	session := newOpenIDSession("")

	tokenType, ar, err := ctx.Providers.OpenIDConnect.Fosite.IntrospectToken(req.Context(), fosite.AccessTokenFromRequest(req), fosite.AccessToken, session)

	if err != nil {
		rfc6749Err := fosite.ErrorToRFC6749Error(err)
		if rfc6749Err.StatusCode() == http.StatusUnauthorized {
			rw.Header().Set("WWW-Authenticate", fmt.Sprintf("error=%s,error_description=%s", rfc6749Err.ErrorField, rfc6749Err.GetDescription()))
		}

		ctx.Providers.OpenIDConnect.Writer.WriteError(rw, req, err)
		return
	}

	if tokenType != fosite.AccessToken {
		errorDescription := "Only access tokens are allowed in the authorization header."
		rw.Header().Set("WWW-Authenticate", fmt.Sprintf("error_description=\"%s\"", errorDescription))
		ctx.Providers.OpenIDConnect.Writer.WriteErrorCode(rw, req, http.StatusUnauthorized, errors.New(errorDescription))
		return
	}

	claims := ar.GetSession().(*OpenIDSession).IDTokenClaims().ToMap()

	delete(claims, "nonce")
	delete(claims, "at_hash")
	delete(claims, "c_hash")
	delete(claims, "exp")
	delete(claims, "sid")
	delete(claims, "jti")

	client, err := ctx.Providers.OpenIDConnect.Store.GetInternalClient(ar.GetClient().GetID())
	if err != nil {
		ctx.Providers.OpenIDConnect.Writer.WriteError(rw, req, err)
		return
	}

	if aud, ok := claims["aud"].([]string); !ok || len(aud) == 0 {
		claims["aud"] = []string{client.GetID()}
	}

	claims["jti"] = uuid.New()
	claims["iat"] = time.Now().Unix()

	keyID := ctx.Providers.OpenIDConnect.Store.KeyManager.GetActiveKeyID()

	token, _, err := ctx.Providers.OpenIDConnect.Store.KeyManager.Strategy().Generate(req.Context(), jwt.MapClaims(claims), &fositejwt.Headers{
		Extra: map[string]interface{}{"kid": keyID},
	})
	if err != nil {
		ctx.Providers.OpenIDConnect.Writer.WriteError(rw, req, err)
		return
	}

	rw.Header().Set("Content-Type", "application/jwt")
	_, _ = rw.Write([]byte(token))
}
