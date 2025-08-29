package handlers

import (
	"net/url"
	"time"

	"authelia.com/provider/oauth2/token/jwt"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// WellKnownOpenIDConfigurationGET handles requests to a .well-known endpoint (RFC5785) which returns the
// OpenID Connect Discovery 1.0 metadata.
//
// RFC5785: Defining Well-Known URIs (https://datatracker.ietf.org/doc/html/rfc5785)
//
// OpenID Connect Discovery 1.0 (https://openid.net/specs/openid-connect-discovery-1_0.html)
func WellKnownOpenIDConfigurationGET(ctx *middlewares.AutheliaCtx) {
	var (
		issuer *url.URL
		err    error
	)
	if issuer, err = ctx.IssuerURL(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred determining issuer")

		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	metadata := oidc.OpenIDConnectWellKnownSignedConfiguration{
		OpenIDConnectWellKnownConfiguration: ctx.Providers.OpenIDConnect.GetOpenIDConnectWellKnownConfiguration(issuer.String()),
	}

	headers := &jwt.Headers{}

	if ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseKeyID != "" {
		headers.Add(oidc.JWTHeaderKeyIdentifier, ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseKeyID)
	}

	if ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseAlg != "" {
		headers.Add(oidc.JWTHeaderKeyAlgorithm, ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseAlg)
	}

	if len(headers.Extra) != 0 {
		claims := metadata.ToMap()

		claims[oidc.ClaimJWTID] = uuid.New().String()
		claims[oidc.ClaimIssuer] = issuer.String()
		claims[oidc.ClaimIssuedAt] = ctx.GetClock().Now().UTC().Unix()
		claims[oidc.ClaimExpirationTime] = ctx.GetClock().Now().Add(time.Hour).UTC().Unix()

		strategy := ctx.Providers.OpenIDConnect.GetJWTStrategy(ctx)

		if metadata.SignedMetadata, _, err = strategy.Encode(ctx, claims, jwt.WithHeaders(headers)); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred signing metadata")

			ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

			return
		}
	}

	if err = ctx.ReplyJSON(metadata, fasthttp.StatusOK); err != nil {
		ctx.Logger.WithError(err).Error("Error occurred encoding JSON response")

		ctx.ReplyStatusCode(fasthttp.StatusInternalServerError)

		return
	}
}
