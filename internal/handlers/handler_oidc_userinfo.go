package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/oauth2"
	"authelia.com/provider/oauth2/token/jwt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OpenIDConnectUserinfo handles GET/POST requests to the OpenID Connect 1.0 UserInfo endpoint.
//
// https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
func OpenIDConnectUserinfo(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		requestID uuid.UUID
		tokenType oauthelia2.TokenType
		requester oauthelia2.AccessRequester
		client    oidc.Client
		err       error
	)

	if requestID, err = uuid.NewRandom(); err != nil {
		ctx.Providers.OpenIDConnect.WriteError(rw, req, oauthelia2.ErrServerError)

		return
	}

	oidcSession := oidc.NewSession()

	ctx.Logger.Debugf("UserInfo Request with id '%s' is being processed", requestID)

	if tokenType, requester, err = ctx.Providers.OpenIDConnect.IntrospectToken(req.Context(), oauthelia2.AccessTokenFromRequest(req), oauthelia2.AccessToken, oidcSession); err != nil {
		ctx.Logger.Errorf("UserInfo Request with id '%s' failed with error: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(err))

		if rfc := oauthelia2.ErrorToRFC6749Error(err); rfc.StatusCode() == http.StatusUnauthorized {
			rw.Header().Set(fasthttp.HeaderWWWAuthenticate, fmt.Sprintf(`Bearer %s`, oidc.RFC6750Header("", "", rfc)))
		}

		ctx.Providers.OpenIDConnect.WriteError(rw, req, err)

		return
	}

	clientID := requester.GetClient().GetID()

	if tokenType != oauthelia2.AccessToken {
		ctx.Logger.Errorf("UserInfo Request with id '%s' on client with id '%s' failed with error: bearer authorization failed as the token is not an access token", requestID, client.GetID())

		errStr := "Only access tokens are allowed in the authorization header."
		rw.Header().Set(fasthttp.HeaderWWWAuthenticate, fmt.Sprintf(`Bearer error="invalid_token",error_description="%s"`, errStr))
		ctx.Providers.OpenIDConnect.WriteErrorCode(rw, req, http.StatusUnauthorized, errors.New(errStr))

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, clientID); err != nil {
		ctx.Logger.Errorf("UserInfo Request with id '%s' on client with id '%s' failed to retrieve client configuration with error: %s", requestID, client.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		ctx.Providers.OpenIDConnect.WriteError(rw, req, err)

		return
	}

	var (
		original map[string]any
	)

	switch session := requester.GetSession().(type) {
	case *oidc.Session:
		original = session.IDTokenClaims().ToMap()
	case *oauth2.JWTSession:
		original = session.JWTClaims.ToMap()
	default:
		ctx.Logger.Errorf("UserInfo Request with id '%s' on client with id '%s' failed to handle session with type '%T'", requestID, client.GetID(), session)

		ctx.Providers.OpenIDConnect.WriteError(rw, req, oauthelia2.ErrServerError.WithDebugf("Failed to handle session with type '%T'.", session))

		return
	}

	claims := map[string]any{}

	oidcApplyUserInfoClaims(clientID, requester.GetGrantedScopes(), original, claims, oidcCtxDetailResolver(ctx))

	var token string

	ctx.Logger.Tracef("UserInfo Response with id '%s' on client with id '%s' is being sent with the following claims: %+v", requestID, clientID, claims)

	switch alg := client.GetUserinfoSignedResponseAlg(); alg {
	case oidc.SigningAlgNone:
		ctx.Logger.Debugf("UserInfo Request with id '%s' on client with id '%s' is being returned unsigned as per the registered client configuration", requestID, client.GetID())

		rw.Header().Set(fasthttp.HeaderContentType, "application/json; charset=utf-8")
		rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
		rw.Header().Set(fasthttp.HeaderPragma, "no-cache")
		rw.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(rw).Encode(claims)
	default:
		var jwk *oidc.JWK

		if jwk = ctx.Providers.OpenIDConnect.KeyManager.Get(ctx, client.GetUserinfoSignedResponseKeyID(), alg); jwk == nil {
			ctx.Providers.OpenIDConnect.WriteError(rw, req, errors.WithStack(oauthelia2.ErrServerError.WithHintf("Unsupported UserInfo signing algorithm '%s'.", alg)))

			return
		}

		ctx.Logger.Debugf("UserInfo Request with id '%s' on client with id '%s' is being returned signed as per the registered client configuration with key id '%s' using the '%s' algorithm", requestID, client.GetID(), jwk.KeyID(), jwk.JWK().Algorithm)

		var jti uuid.UUID

		if jti, err = uuid.NewRandom(); err != nil {
			ctx.Providers.OpenIDConnect.WriteError(rw, req, oauthelia2.ErrServerError.WithHint("Could not generate JTI."))

			return
		}

		claims[oidc.ClaimJWTID] = jti.String()
		claims[oidc.ClaimIssuedAt] = time.Now().UTC().Unix()

		headers := &jwt.Headers{
			Extra: map[string]any{
				oidc.JWTHeaderKeyIdentifier: jwk.KeyID(),
			},
		}

		if token, _, err = jwk.Strategy().Generate(req.Context(), claims, headers); err != nil {
			ctx.Providers.OpenIDConnect.WriteError(rw, req, err)

			return
		}

		rw.Header().Set(fasthttp.HeaderContentType, "application/jwt; charset=utf-8")
		rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
		rw.Header().Set(fasthttp.HeaderPragma, "no-cache")
		rw.WriteHeader(http.StatusOK)

		_, _ = rw.Write([]byte(token))
	}

	ctx.Logger.Debugf("UserInfo Request with id '%s' on client with id '%s' was successfully processed", requestID, client.GetID())
}
