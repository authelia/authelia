// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"net/url"

	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// OpenIDConnectEndSession handles requests made by resource owners when the relying-party redirects them
// requesting they logout.
//
// The request is parsed and validated by the provider, which resolves the client from the client_id and/or
// id_token_hint parameters and matches any post_logout_redirect_uri against those registered for that client. This
// handler ends no session itself: it redirects the resource owner to the front-end, which always asks them to
// confirm the logout before it occurs. The specification permits skipping that confirmation when the id_token_hint
// validates, however Authelia is deliberately stricter.
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

	var requester oauthelia2.RPInitiatedLogoutRequester

	if requester, err = ctx.Providers.OpenIDConnect.NewRPInitiatedLogoutRequest(ctx, req); err != nil {
		oidcEndSessionRedirectError(ctx, rw, req, issuer, err)

		return
	}

	oidcEndSessionRedirectSuccess(rw, req, issuer, requester)
}

func oidcEndSessionRedirectSuccess(rw http.ResponseWriter, req *http.Request, issuer *url.URL, requester oauthelia2.RPInitiatedLogoutRequester) {
	location := issuer.JoinPath(oidcFrontendEndpointPathLogout)

	query := location.Query()

	query.Set(oidcFrontendQueryParameterConfirm, "true")

	if redirect := requester.GetPostLogoutRedirectURI(); redirect != nil {
		query.Set(oidcFrontendQueryParameterRedirectionURL, redirect.String())

		if state := requester.GetState(); len(state) != 0 {
			query.Set(oidc.FormParameterState, state)
		}
	}

	location.RawQuery = query.Encode()

	oidcEndSessionRedirect(rw, req, location)
}

func oidcEndSessionRedirectError(ctx *middlewares.AutheliaCtx, rw http.ResponseWriter, req *http.Request, issuer *url.URL, err error) {
	rfc := oauthelia2.ErrorToRFC6749Error(err)

	ctx.Logger.WithError(err).Errorf("RP-Initiated Logout request could not be processed: %s", oauthelia2.ErrorToDebugRFC6749Error(err))

	location := issuer.JoinPath(oidc.FrontendEndpointPathConsentCompletion)

	query := location.Query()

	if len(rfc.ErrorField) != 0 {
		query.Set(oidcFrontendQueryParameterError, rfc.ErrorField)
	}

	if len(rfc.DescriptionField) != 0 {
		query.Set(oidcFrontendQueryParameterErrorDescription, rfc.DescriptionField)
	}

	if len(rfc.HintField) != 0 {
		query.Set(oidcFrontendQueryParameterErrorHint, rfc.HintField)
	}

	if ctx.Providers.OpenIDConnect != nil && ctx.Providers.OpenIDConnect.GetSendDebugMessagesToClients(ctx) && len(rfc.DebugField) != 0 {
		query.Set(oidcFrontendQueryParameterErrorDebug, rfc.DebugField)
	}

	location.RawQuery = query.Encode()

	oidcEndSessionRedirect(rw, req, location)
}

func oidcEndSessionRedirect(rw http.ResponseWriter, req *http.Request, location *url.URL) {
	rw.Header().Set(fasthttp.HeaderCacheControl, "no-store")
	rw.Header().Set(fasthttp.HeaderPragma, "no-cache")

	http.Redirect(rw, req, location.String(), http.StatusFound)
}

const (
	oidcFrontendEndpointPathLogout = "/logout"

	oidcFrontendQueryParameterConfirm        = "confirm"
	oidcFrontendQueryParameterRedirectionURL = "rd"

	oidcFrontendQueryParameterError            = "error"
	oidcFrontendQueryParameterErrorDescription = "error_description"
	oidcFrontendQueryParameterErrorHint        = "error_hint"
	oidcFrontendQueryParameterErrorDebug       = "error_debug"
)
