// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"

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

// oidcEndSessionStore stores the logout request until the End-User confirms it. The request is identified by the flow
// id rather than stored in the session, as the session cookie isn't sent with a cross-site POST, nor with any cross-site
// request when it has the SameSite attribute set to Strict, and a session saved without it would replace the cookie of
// the End-User. The logout page requests the logout using the flow id, which is same-site and so carries the cookie.
func oidcEndSessionStore(ctx *middlewares.AutheliaCtx, requester oauthelia2.RPInitiatedLogoutRequester) (flowID uuid.UUID, err error) {
	if flowID, err = uuid.NewRandom(); err != nil {
		return uuid.Nil, err
	}

	now := ctx.GetClock().Now()

	logout := model.OAuth2LogoutSession{
		FlowID:    flowID,
		CreatedAt: now,
		ExpiresAt: now.Add(oidcEndSessionLifespan),
		Subject:   requester.GetSubject(),
		SessionID: requester.GetSessionID(),
	}

	if client := requester.GetClient(); client != nil {
		logout.ClientID = client.GetID()
	}

	if redirect := requester.GetPostLogoutRedirectURI(); redirect != nil {
		logout.RedirectURI = redirect.String()
		logout.State = requester.GetState()
	}

	if err = ctx.Providers.StorageProvider.SaveOAuth2LogoutSession(ctx, logout); err != nil {
		return uuid.Nil, err
	}

	return flowID, nil
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
