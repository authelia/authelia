package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jwt"
	"authelia.com/provider/oauth2/x/errorsx"
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
		errorsx.WriteJSONError(rw, req, oauthelia2.ErrServerError)

		return
	}

	oidcSession := oidc.NewSession()

	ctx.Logger.Debugf("UserInfo Request with id '%s' is being processed", requestID)

	if tokenType, requester, err = ctx.Providers.OpenIDConnect.IntrospectToken(req.Context(), oauthelia2.AccessTokenFromRequest(req), oauthelia2.AccessToken, oidcSession); err != nil {
		ctx.Logger.Errorf("UserInfo Request with id '%s' failed with error: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(err))

		if rfc := oauthelia2.ErrorToRFC6749Error(err); rfc.StatusCode() == http.StatusUnauthorized {
			rw.Header().Set(fasthttp.HeaderWWWAuthenticate, fmt.Sprintf(`Bearer %s`, oidc.RFC6750Header("", "", rfc)))
		}

		errorsx.WriteJSONError(rw, req, err)

		return
	}

	if tokenType != oauthelia2.AccessToken {
		ctx.Logger.Errorf("UserInfo Request with id '%s' on client with id '%s' failed with error: bearer authorization failed as the token is not an access token", requestID, client.GetID())

		errStr := "Only access tokens are allowed in the authorization header."
		rw.Header().Set(fasthttp.HeaderWWWAuthenticate, fmt.Sprintf(`Bearer error="invalid_token",error_description="%s"`, errStr))
		errorsx.WriteJSONErrorCode(rw, req, http.StatusUnauthorized, errors.New(errStr))

		return
	}

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, requester.GetClient().GetID()); err != nil {
		ctx.Logger.Errorf("UserInfo Request with id '%s' on client with id '%s' failed to retrieve client configuration with error: %s", requestID, client.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		errorsx.WriteJSONError(rw, req, err)

		return
	}

	var (
		original map[string]any
		requests map[string]*oidc.ClaimRequest
		userinfo bool
	)

	switch session := requester.GetSession().(type) {
	case *oidc.Session:
		original = session.IDTokenClaims().ToMap()
		requests = session.ClaimRequests.GetUserInfoRequests()
		userinfo = !session.ClientCredentials
	default:
		ctx.Logger.Errorf("UserInfo Request with id '%s' on client with id '%s' failed to handle session with type '%T'", requestID, client.GetID(), session)

		errorsx.WriteJSONError(rw, req, oauthelia2.ErrServerError.WithDebugf("Failed to handle session with type '%T'.", session))

		return
	}

	claims := map[string]any{}

	ctx.Logger.WithFields(map[string]any{"claims": original, "requested_scopes": requester.GetRequestedScopes(), "granted_scopes": requester.GetGrantedScopes()}).Debug("Original Values")

	strategy := ctx.Providers.OpenIDConnect.GetScopeStrategy(ctx)

	var detailer oidc.UserDetailer

	if detailer, err = oidcDetailerFromClaims(ctx, original); err != nil {
		client.GetClaimsStrategy().PopulateClientCredentialsUserInfoClaims(ctx, client, original, claims)

		if userinfo {
			ctx.Logger.WithError(err).Errorf("UserInfo Request with id '%s' on client with id '%s' error occurred loading user information", requestID, client.GetID())
		}
	} else {
		client.GetClaimsStrategy().PopulateUserInfoClaims(ctx, strategy, client, requester.GetGrantedScopes(), requests, detailer, ctx.Clock.Now(), original, claims)
	}

	/*
		client.GetClaimsStrategy().PopulateIDTokenClaims(ctx, strategy, client, requester.GetGrantedScopes(), requests.GetIDTokenRequests(), detailer, ctx.Clock.Now(), nil, extra)

		if detailer, err = oidcDetailerFromClaims(ctx, original); err != nil {
			oidc.GrantClaimsScoped(ctx, strategy, client, requester.GetGrantedScopes(), claims)

			if userinfo {
				ctx.Logger.WithError(err).Errorf("UserInfo Request with id '%s' on client with id '%s' error occurred loading user information", requestID, client.GetID())
			}
		} else {
			oidc.GrantClaimsScoped(ctx, strategy, client, requester.GetGrantedScopes(), detailer, original, claims)
			oidc.GrantClaimsRequested(ctx, strategy, client, requests, detailer, claims)
		}
	*/

	var token string

	ctx.Logger.Tracef("UserInfo Response with id '%s' on client with id '%s' is being sent with the following claims: %+v", requestID, requester.GetClient().GetID(), claims)

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
			errorsx.WriteJSONError(rw, req, errors.WithStack(oauthelia2.ErrServerError.WithHintf("Unsupported UserInfo signing algorithm '%s'.", alg)))

			return
		}

		ctx.Logger.Debugf("UserInfo Request with id '%s' on client with id '%s' is being returned signed as per the registered client configuration with key id '%s' using the '%s' algorithm", requestID, client.GetID(), jwk.KeyID(), jwk.JWK().Algorithm)

		var jti uuid.UUID

		if jti, err = uuid.NewRandom(); err != nil {
			errorsx.WriteJSONError(rw, req, oauthelia2.ErrServerError.WithHint("Could not generate JTI."))

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
			errorsx.WriteJSONError(rw, req, err)

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
