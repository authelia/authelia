package handlers

import (
	"net/url"

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
	var (
		issuer *url.URL
		err    error
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred determining issuer")

		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	if err = ctx.ReplyJSON(ctx.Providers.OpenIDConnect.GetOpenIDConnectWellKnownConfiguration(issuer.String()), fasthttp.StatusOK); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred encoding JSON response")

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
	var (
		issuer *url.URL
		err    error
	)

	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred determining issuer")

		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	if err = ctx.ReplyJSON(ctx.Providers.OpenIDConnect.GetOAuth2WellKnownConfiguration(issuer.String()), fasthttp.StatusOK); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred encoding JSON response")

		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
