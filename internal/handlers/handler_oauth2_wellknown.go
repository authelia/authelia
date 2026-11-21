// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// WellKnownOAuthAuthorizationServerGET handles requests to a .well-known endpoint (RFC5785) which returns the
// OAuth 2.0 Authorization Server Metadata (RFC8414).
//
// RFC5785: Defining Well-Known URIs (https://datatracker.ietf.org/doc/html/rfc5785)
//
// RFC8414: OAuth 2.0 Authorization Server Metadata (https://datatracker.ietf.org/doc/html/rfc8414)
func WellKnownOAuthAuthorizationServerGET(ctx *middlewares.AutheliaCtx) {
	var (
		metadata *oauthelia2.OAuth2AuthorizationServerMetadata
		err      error
	)

	if _, err = ctx.IssuerURL(); err != nil {
		rfc := oidc.ErrEffectiveIssuer.WithWrap(err)

		ctx.GetLogger().WithError(err).Errorf("OAuth 2.0 Discovery Request could not be processed: %s", oauthelia2.ErrorToDebugRFC6749Error(rfc))

		ctx.ReplyStatusCode(rfc.StatusCode())

		return
	}

	if metadata, err = ctx.Providers.OpenIDConnect.NewOAuth2AuthorizationServerMetadata(ctx); err != nil {
		ctx.GetLogger().WithError(err).Errorf("OAuth 2.0 Discovery Request could not be processed: error occurred building the metadata")

		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	if err = ctx.ReplyJSON(metadata, fasthttp.StatusOK); err != nil {
		ctx.GetLogger().WithError(err).Error("Error occurred encoding JSON response")

		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
