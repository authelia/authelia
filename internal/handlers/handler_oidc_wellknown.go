package handlers

import (
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// OpenIDConnectConfigurationWellKnownGET handles requests to a .well-known endpoint (RFC5785) which returns the
// OpenID Connect Discovery 1.0 metadata.
//
// https://datatracker.ietf.org/doc/html/rfc5785
//
// https://openid.net/specs/openid-connect-discovery-1_0.html
func OpenIDConnectConfigurationWellKnownGET(ctx *middlewares.AutheliaCtx) {
	var (
		issuer *url.URL
		err    error
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.Logger.Errorf("Error occurred determining OpenID Connect issuer details: %+v", err)

		ctx.ReplyStatusCode(fasthttp.StatusBadRequest)

		return
	}

	wellKnown := ctx.Providers.OpenIDConnect.GetOpenIDConnectWellKnownConfiguration(issuer.String())

	if err = ctx.ReplyJSON(wellKnown, fasthttp.StatusOK); err != nil {
		ctx.Logger.Errorf("Error occurred in JSON encode: %+v", err)

		// TODO: Determine if this is the appropriate error code here.
		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}

// OAuthAuthorizationServerWellKnownGET handles requests to a .well-known endpoint (RFC5785) which returns the
// OAuth 2.0 Authorization Server Metadata (RFC8414).
//
// https://datatracker.ietf.org/doc/html/rfc5785
//
// https://datatracker.ietf.org/doc/html/rfc8414
func OAuthAuthorizationServerWellKnownGET(ctx *middlewares.AutheliaCtx) {
	var (
		issuer *url.URL
		err    error
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.Logger.Errorf("Error occurred determining OpenID Connect issuer details: %+v", err)

		ctx.ReplyStatusCode(fasthttp.StatusBadRequest)

		return
	}

	wellKnown := ctx.Providers.OpenIDConnect.GetOAuth2WellKnownConfiguration(issuer.String())

	if err = ctx.ReplyJSON(wellKnown, fasthttp.StatusOK); err != nil {
		ctx.Logger.Errorf("Error occurred in JSON encode: %+v", err)

		// TODO: Determine if this is the appropriate error code here.
		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
