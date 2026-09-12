// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
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

	if req.Method == http.MethodPost {
		oidcEndSessionRedirectPost(ctx, rw, req, issuer)

		return
	}

	var requester oauthelia2.RPInitiatedLogoutRequester

	if requester, err = ctx.Providers.OpenIDConnect.NewRPInitiatedLogoutRequest(ctx, req); err != nil {
		oidcEndSessionRedirectError(ctx, rw, issuer, err)

		return
	}

	var flowID string

	if flowID, err = oidcEndSessionStore(ctx, requester); err != nil {
		oidcEndSessionRedirectError(ctx, rw, issuer, oauthelia2.ErrServerError.WithHint("Could not store the logout request.").WithWrap(err).WithDebugError(err))

		return
	}

	location := issuer.JoinPath(oidcFrontendEndpointPathLogout)

	location.RawQuery = url.Values{queryArgFlowID: []string{flowID}}.Encode()

	oidcEndSessionRedirect(rw, location, http.StatusFound)
}

func oidcEndSessionRedirectPost(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request, issuer *url.URL) {
	req.Body = http.MaxBytesReader(rw, req.Body, oidcEndSessionMaxBodySize)

	//nolint:gosec // G120 FALSE positive: bounded by MaxBytesReader on req.Body above.
	if err := req.ParseMultipartForm(oidcEndSessionMaxBodySize); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		oidcEndSessionRedirectError(ctx, rw, issuer, oauthelia2.ErrInvalidRequest.WithHint("Unable to parse HTTP body, make sure to send a properly formatted form request body.").WithWrap(err).WithDebugError(err))

		return
	}

	location := issuer.JoinPath(oidc.EndpointPathEndSession)

	location.RawQuery = req.Form.Encode()

	oidcEndSessionRedirect(rw, location, http.StatusSeeOther)
}

func oidcEndSessionStore(ctx *middlewares.AutheliaCtx, requester oauthelia2.RPInitiatedLogoutRequester) (flowID string, err error) {
	var (
		userSession session.UserSession
		id          uuid.UUID
	)

	if userSession, err = ctx.GetSession(); err != nil {
		return "", err
	}

	if id, err = uuid.NewRandom(); err != nil {
		return "", err
	}

	logout := &session.OpenIDConnectLogout{
		FlowID:  id.String(),
		Expires: ctx.GetClock().Now().Add(oidcEndSessionLifespan),
	}

	if client := requester.GetClient(); client != nil {
		logout.ClientID = client.GetID()
	}

	if redirect := requester.GetPostLogoutRedirectURI(); redirect != nil {
		logout.RedirectURI = redirect.String()
		logout.State = requester.GetState()
	}

	userSession.OpenIDConnectLogout = logout

	if err = ctx.SaveSession(&userSession); err != nil {
		return "", err
	}

	return logout.FlowID, nil
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
