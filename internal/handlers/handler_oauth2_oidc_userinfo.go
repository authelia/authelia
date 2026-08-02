package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/oauth2"
	"authelia.com/provider/oauth2/handler/rfc9449"
	"authelia.com/provider/oauth2/token/jwt"
	"authelia.com/provider/oauth2/x/errorsx"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// schemeOpenIDConnectUserinfoBearer is the RFC6750 authentication scheme an Access Token which is not bound to a
// proof-of-possession key is presented with.
const schemeOpenIDConnectUserinfoBearer = "Bearer"

// challengeSchemeOpenIDConnectUserinfo returns the authentication scheme the WWW-Authenticate challenge for a rejected
// request must be issued under. RFC9449 Section 7.1 requires a protected resource to challenge with the DPoP scheme
// when the client presented the Access Token with it.
func challengeSchemeOpenIDConnectUserinfo(dpop bool) (scheme string) {
	if dpop {
		return oidc.SchemeDPoP
	}

	return schemeOpenIDConnectUserinfoBearer
}

// writeOpenIDConnectUserinfoDPoPError rejects a request which presented the Access Token using the DPoP authentication
// scheme. The oauthelia2 errors which describe a DPoP failure carry the HTTP 400 status code of the token endpoint,
// whereas RFC9449 Section 7.1 requires a protected resource to respond with HTTP 401 and a DPoP scheme challenge, so
// neither errorsx.WriteJSONError nor errorsx.WriteRFC6750Error can produce this response.
func writeOpenIDConnectUserinfoDPoPError(rw http.ResponseWriter, r *http.Request, rfc *oauthelia2.RFC6749Error) {
	rw.Header().Set(fasthttp.HeaderWWWAuthenticate, fmt.Sprintf("%s %s", oidc.SchemeDPoP, oidc.RFC6750Header("", "", rfc)))

	errorsx.WriteJSONErrorCode(rw, r, http.StatusUnauthorized, rfc)
}

// handleOpenIDConnectUserinfoDPoP performs the RFC9449 Section 7.2 checks for a request to the OpenID Connect 1.0
// UserInfo endpoint. It returns true when the request has been rejected and the response has been written.
//
// The checks are performed when the Access Token is bound to a proof-of-possession key or when it was presented using
// the DPoP authentication scheme. An unbound Access Token presented using the bearer scheme is not subject to RFC9449.
func handleOpenIDConnectUserinfoDPoP(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request, requester oauthelia2.AccessRequester, requestID uuid.UUID, token string, dpop bool) (handled bool) {
	session, ok := requester.GetSession().(*oidc.Session)
	if !ok {
		return false
	}

	jkt := session.GetDPoPJWKThumbprint()

	if jkt == "" && !dpop {
		return false
	}

	var (
		nonce string
		err   error
	)

	if nonce, err = oidc.ValidateDPoPResourceAccess(r.Context(), ctx.Providers.OpenIDConnect.GetDPoPStrategy(ctx), r, token, jkt, ctx.Providers.OpenIDConnect.GetDPoPNonceRequired(ctx)); err == nil {
		return false
	}

	ctx.GetLogger().Errorf("User Info Request with id '%s' failed with error: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(err))

	if nonce != "" {
		rw.Header().Set(oidc.HeaderDPoPNonce, nonce)
	}

	writeOpenIDConnectUserinfoDPoPError(rw, r, oauthelia2.ErrorToRFC6749Error(err))

	return true
}

// OpenIDConnectUserinfo handles GET/POST requests to the OpenID Connect 1.0 UserInfo endpoint.
//
// https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
//
//nolint:gocyclo
func OpenIDConnectUserinfo(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, r *http.Request) {
	var (
		issuer    *url.URL
		requestID uuid.UUID
		tokenType oauthelia2.TokenType
		requester oauthelia2.AccessRequester
		client    oidc.Client
		err       error
	)

	token, dpop := rfc9449.AccessTokenFromRequest(r)

	if requestID, err = uuid.NewRandom(); err != nil {
		errorsx.WriteJSONError(rw, r, oauthelia2.ErrServerError)

		return
	}

	ctx.GetLogger().Debugf("User Info Request with id '%s' is being processed", requestID)

	if issuer, err = ctx.IssuerURL(); err != nil {
		rfc := oidc.ErrEffectiveIssuer.WithWrap(err)

		ctx.GetLogger().WithError(err).Errorf("User Info Request with id '%s' could not be processed: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(rfc))

		errorsx.WriteJSONError(rw, r, rfc)

		return
	}

	if tokenType, requester, err = ctx.Providers.OpenIDConnect.IntrospectToken(oauth2.SetSkipStatelessIntrospection(r.Context()), token, oauthelia2.AccessToken, oidc.NewSessionWithRequestedAt(ctx.GetClock().Now())); err != nil {
		ctx.GetLogger().Errorf("User Info Request with id '%s' failed with error: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(err))

		if rfc := oauthelia2.ErrorToRFC6749Error(err); rfc.StatusCode() == http.StatusUnauthorized {
			rw.Header().Set(fasthttp.HeaderWWWAuthenticate, fmt.Sprintf("%s %s", challengeSchemeOpenIDConnectUserinfo(dpop), oidc.RFC6750Header("", "", rfc)))
		}

		errorsx.WriteJSONError(rw, r, err)

		return
	}

	if tokenType != oauthelia2.AccessToken {
		ctx.GetLogger().Errorf("User Info Request with id '%s' on client with id '%s' failed with error: bearer authorization failed as the token is not an Access Token", requestID, requester.GetClient().GetID())

		rfc := oauthelia2.ErrInvalidTokenFormat.WithDescription("Only OpenID Connect 1.0 Access Tokens are allowed in the authorization header.")

		if dpop {
			writeOpenIDConnectUserinfoDPoPError(rw, r, rfc)
		} else {
			errorsx.WriteRFC6750Error(rw, rfc, nil)
		}

		return
	}

	if handleOpenIDConnectUserinfoDPoP(ctx, rw, r, requester, requestID, token, dpop) {
		return
	}

	var (
		original      map[string]any
		requests      map[string]*oidc.ClaimRequest
		claimsGranted oauthelia2.Arguments
		requested     time.Time
		userinfo      bool
	)

	if client, err = ctx.Providers.OpenIDConnect.GetRegisteredClient(ctx, requester.GetClient().GetID()); err != nil {
		ctx.GetLogger().Errorf("User Info Request with id '%s' on client with id '%s' failed to retrieve client configuration with error: %s", requestID, client.GetID(), oauthelia2.ErrorToDebugRFC6749Error(err))

		errorsx.WriteRFC6750Error(
			rw,
			oauthelia2.ErrInvalidRequest.WithHint("The client the access token was issued to is no longer registered with the authorization server."),
			nil,
		)

		errorsx.WriteJSONError(rw, r, err)

		return
	}

	switch session := requester.GetSession().(type) {
	case *oidc.Session:
		if !session.ValidIssuer(issuer) {
			err = oauthelia2.ErrInvalidRequest.WithDebug("The original request and the userinfo request occurred at endpoints where the origin or effective issuer did not match.")

			ctx.GetLogger().Errorf("User Info Request with id '%s' could not be processed: %s", requestID, oauthelia2.ErrorToDebugRFC6749Error(err))

			errorsx.WriteRFC6750Error(rw, err, nil)

			return
		}

		original = session.IDTokenClaims().ToMap()
		requests = session.ClaimRequests.GetUserInfoRequests()
		requested = session.GetRequestedAt()
		userinfo = !session.ClientCredentials
		claimsGranted = session.GrantedClaims
	default:
		ctx.GetLogger().Errorf("User Info Request with id '%s' on client with id '%s' failed to handle session with type '%T'", requestID, requester.GetClient().GetID(), session)

		errorsx.WriteJSONError(rw, r, oauthelia2.ErrServerError.WithDebugf("Failed to handle session with type '%T'.", session))

		return
	}

	if !requester.GetGrantedScopes().Has(oidc.ScopeOpenID) {
		ctx.GetLogger().Errorf("User Info Request with id '%s' on client with id '%s' failed with error: bearer authorization failed as the Access Token was not granted the appropriate scope", requestID, client.GetID())

		errorsx.WriteRFC6750Error(
			rw,
			oauthelia2.ErrInsufficientScope.WithHint("The granted scope was missing the 'openid' scope."),
			map[string]string{oidc.FormParameterScope: oidc.ScopeOpenID},
		)

		return
	}

	claims := jwt.MapClaims{}

	var detailer oidc.UserDetailer

	if detailer, err = oidc.UserDetailerFromClaims(ctx, original); err != nil {
		if err = client.GetClaimsStrategy().HydrateClientCredentialsUserInfoClaims(ctx, client, original, claims); err != nil {
			ctx.GetLogger().WithError(err).Errorf("User Info Request with id '%s' on client with id '%s' failed due to an error populating claims for the client credentials flow", requestID, client.GetID())

			errorsx.WriteJSONError(rw, r, oauthelia2.ErrServerError.WithDebugf("Error occurred populating claims for the client credentials flow: %v.", err))

			return
		}

		if userinfo {
			ctx.GetLogger().WithError(err).Errorf("User Info Request with id '%s' on client with id '%s' error occurred loading user information", requestID, client.GetID())
		}
	} else if err = client.GetClaimsStrategy().HydrateUserInfoClaims(ctx, ctx.Providers.OpenIDConnect.GetScopeStrategy(ctx), client, requester.GetGrantedScopes(), claimsGranted, requests, detailer, requested, ctx.GetClock().Now(), original, claims); err != nil {
		ctx.GetLogger().WithError(err).Errorf("User Info Request with id '%s' on client with id '%s' failed due to an error populating claims for the standard flow", requestID, client.GetID())

		errorsx.WriteJSONError(rw, r, oauthelia2.ErrServerError.WithDebugf("Error occurred populating claims for the standard flow: %v.", err))

		return
	}

	ctx.GetLogger().Tracef("User Info Response with id '%s' on client with id '%s' is being sent with the following claims: %+v", requestID, requester.GetClient().GetID(), claims)

	switch alg := client.GetUserinfoSignedResponseAlg(); alg {
	case oidc.SigningAlgNone:
		ctx.GetLogger().Debugf("User Info Request with id '%s' on client with id '%s' is being returned unsigned as per the registered client configuration", requestID, client.GetID())

		rw.Header().Set(fasthttp.HeaderContentType, middlewares.ContentTypeApplicationJSON)

		_ = json.NewEncoder(rw).Encode(claims)
	default:
		var (
			jti   uuid.UUID
			token string
		)

		jwtClient := oidc.NewUserinfoClient(client)

		ctx.GetLogger().Debugf("User Info Request with id '%s' on client with id '%s' is being returned signed as per the registered client configuration with key id '%s' using the '%s' algorithm", requestID, client.GetID(), jwtClient.GetSigningKeyID(), jwtClient.GetSigningAlg())

		if jti, err = uuid.NewRandom(); err != nil {
			ctx.GetLogger().WithError(err).Errorf("User Info Request with id '%s' on client with id '%s' failed due to an error generating a JTI for the JWT response", requestID, client.GetID())

			errorsx.WriteJSONError(rw, r, oauthelia2.ErrServerError.WithHint("Could not generate JTI."))

			return
		}

		claims[oidc.ClaimJWTID] = jti.String()
		claims[oidc.ClaimIssuedAt] = time.Now().UTC().Unix()

		strategy := ctx.Providers.OpenIDConnect.GetJWTStrategy(ctx)

		if token, _, err = strategy.Encode(ctx, claims, jwt.WithClient(jwtClient)); err != nil {
			errorsx.WriteJSONError(rw, r, err)

			return
		}

		rw.Header().Set(fasthttp.HeaderContentType, middlewares.ContentTypeApplicationJWT)

		_, _ = rw.Write([]byte(token)) //nolint:gosec // TODO: Run this line through taint analysis.
	}

	rw.Header().Set(fasthttp.HeaderCacheControl, middlewares.HeaderCacheControlNotStore)
	rw.Header().Set(fasthttp.HeaderPragma, middlewares.HeaderPragmaNoCache)

	ctx.GetLogger().Debugf("User Info Request with id '%s' on client with id '%s' was successfully processed", requestID, client.GetID())
}
