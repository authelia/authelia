// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// OpenIDConnectConfigurationWellKnownGET handles requests to a .well-known endpoint (RFC5785) which returns the
// OpenID Connect Discovery 1.0 metadata.
//
// RFC5785: Defining Well-Known URIs (https://datatracker.ietf.org/doc/html/rfc5785)
//
// OpenID Connect Discovery 1.0 (https://openid.net/specs/openid-connect-discovery-1_0.html)
func OpenIDConnectConfigurationWellKnownGET(ctx *middlewares.AutheliaCtx) {
	if err := ctx.ReplyJSON(ctx.Providers.OpenIDConnect.GetOpenIDConnectWellKnownConfiguration(ctx.RootURL().String()), fasthttp.StatusOK); err != nil {
		ctx.Logger.Errorf("Error occurred in JSON encode: %+v", err)

		// TODO: Determine if this is the appropriate error code here.
		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}

// OAuthAuthorizationServerWellKnownGET handles requests to a .well-known endpoint (RFC5785) which returns the
// OAuth 2.0 Authorization Server Metadata (RFC8414).
//
// RFC5785: Defining Well-Known URIs (https://datatracker.ietf.org/doc/html/rfc5785)
//
// RFC8414: OAuth 2.0 Authorization Server Metadata (https://datatracker.ietf.org/doc/html/rfc8414)
func OAuthAuthorizationServerWellKnownGET(ctx *middlewares.AutheliaCtx) {
	if err := ctx.ReplyJSON(ctx.Providers.OpenIDConnect.GetOAuth2WellKnownConfiguration(ctx.RootURL().String()), fasthttp.StatusOK); err != nil {
		ctx.Logger.Errorf("Error occurred in JSON encode: %+v", err)

		// TODO: Determine if this is the appropriate error code here.
		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
