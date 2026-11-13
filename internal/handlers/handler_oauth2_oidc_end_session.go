// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OpenIDConnectEndSession handles requests made by resource owners when the relying-party redirects them
// requesting they logout.
//
// OpenID Connect RP-Initiated Logout 1.0 (https://openid.net/specs/openid-connect-rpinitiated-1_0.html)
func OpenIDConnectEndSession(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request) {
	var (
		issuer *url.URL
		err    error
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred determining issuer")

		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	code := http.StatusFound

	if req.Method == http.MethodPost {
		// The End-User is redirected to the logout page with a GET request rather than the POST being repeated.
		code = http.StatusSeeOther

		req.Body = http.MaxBytesReader(rw, req.Body, oidcEndSessionMaxBodySize)
	}

	var requester oauthelia2.RPInitiatedLogoutRequester

	if requester, err = ctx.Providers.OpenIDConnect.NewRPInitiatedLogoutRequest(ctx, req); err != nil {
		oidcEndSessionRedirectError(ctx, rw, issuer, err)

		return
	}

	var flowID uuid.UUID

	if flowID, err = oidcEndSessionStore(ctx, requester); err != nil {
		oidcEndSessionRedirectError(ctx, rw, issuer, oauthelia2.ErrServerError.WithHint("Could not store the logout request.").WithWrap(err).WithDebugError(err))

		return
	}

	location := issuer.JoinPath(oidcFrontendEndpointPathLogout)

	location.RawQuery = url.Values{queryArgFlowID: []string{flowID.String()}}.Encode()

	oidcEndSessionRedirect(rw, location, code)
}

func oidcEndSessionStore(ctx *middlewares.AutheliaCtx, requester oauthelia2.RPInitiatedLogoutRequester) (flowID uuid.UUID, err error) {
	if flowID, err = uuid.NewRandom(); err != nil {
		return uuid.Nil, err
	}

	now := ctx.GetClock().Now()

	logout := model.OAuth2SessionLogout{
		FlowID:    flowID,
		CreatedAt: now,
		ExpiresAt: now.Add(oidcEndSessionLifespan),
	}

	if oidcEndSessionHintTrusted(ctx, requester) {
		logout.Subject, logout.SessionID = requester.GetSubject(), requester.GetSessionID()
	} else if requester.GetIDTokenHintClaims() != nil {
		ctx.Logger.Debug("RP-Initiated Logout request 'id_token_hint' is not trusted to identify the End-User or session and is treated as if it was absent")
	}

	if client := requester.GetClient(); client != nil {
		logout.ClientID = client.GetID()
	}

	if redirect := requester.GetPostLogoutRedirectURI(); redirect != nil {
		logout.RedirectURI = redirect.String()
		logout.State = requester.GetState()
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2SessionLogout(ctx, logout); err != nil {
		return uuid.Nil, err
	}

	return flowID, nil
}

func oidcEndSessionHintTrusted(ctx *middlewares.AutheliaCtx, requester oauthelia2.RPInitiatedLogoutRequester) (trusted bool) {
	claims := requester.GetIDTokenHintClaims()
	if claims == nil {
		return false
	}

	client, ok := requester.GetClient().(oidc.Client)
	if !ok {
		return false
	}

	if jwt.IsSignedJWTClientSecretAlgStr(client.GetIDTokenSignedResponseAlg()) {
		return false
	}

	now := ctx.GetClock().Now()

	if exp, err := claims.GetExpirationTime(); err == nil && exp != nil && now.Before(exp.Time) {
		return true
	}

	if oidcEndSessionHintSessionCurrent(ctx, requester.GetSessionID()) {
		return true
	}

	iat, err := claims.GetIssuedAt()
	if err != nil || iat == nil {
		return false
	}

	lifespan := oidcEndSessionHintRecentLifespan(ctx, client)

	return lifespan < 0 || now.Before(iat.Add(lifespan))
}

func oidcEndSessionHintSessionCurrent(ctx *middlewares.AutheliaCtx, sid string) (current bool) {
	if sid == "" {
		return false
	}

	provider, err := ctx.GetSessionProvider()
	if err != nil {
		ctx.Logger.WithError(err).Debug("Error occurred obtaining the session provider to determine if the 'id_token_hint' session is current")

		return false
	}

	record, err := ctx.Providers.StorageProvider.LoadOAuth2SessionIDBySessionID(ctx, provider.GetIssuer(), sid)
	if err != nil {
		ctx.Logger.WithError(err).Debug("Error occurred loading the session identified by the 'id_token_hint' to determine if it is current")

		return false
	}

	return record != nil
}

func oidcEndSessionHintRecentLifespan(ctx *middlewares.AutheliaCtx, client oidc.Client) (lifespan time.Duration) {
	fallback := ctx.Providers.OpenIDConnect.GetRefreshTokenLifespan(ctx)

	for _, gt := range []oauthelia2.GrantType{oauthelia2.GrantTypeAuthorizationCode, oauthelia2.GrantTypeDeviceCode, oauthelia2.GrantTypeRefreshToken} {
		value := oauthelia2.GetEffectiveLifespan(client, gt, oauthelia2.RefreshToken, fallback)

		if value < 0 {
			return value
		}

		lifespan = max(lifespan, value)
	}

	return lifespan
}

func oidcEndSessionRedirectError(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, issuer *url.URL, err error) {
	rfc := oauthelia2.ErrorToRFC6749Error(err)

	ctx.Logger.WithError(err).Errorf("RP-Initiated Logout request could not be processed: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

	location := oidc.NewConsentCompletionErrorURI(issuer, rfc, ctx.Providers.OpenIDConnect.GetSendDebugMessagesToClients(ctx))

	oidcEndSessionRedirect(rw, location, http.StatusFound)
}

func oidcEndSessionRedirect(rw http.ResponseWriter, location *url.URL, code int) {
	rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
	rw.Header().Set(fasthttp.HeaderPragma, "no-cache")
	rw.Header().Set(fasthttp.HeaderLocation, location.String())
	rw.WriteHeader(code)
}
