package handlers

import (
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
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

	metadata := oidc.OpenIDConnectWellKnownSignedConfiguration{
		OpenIDConnectWellKnownConfiguration: ctx.Providers.OpenIDConnect.GetOpenIDConnectWellKnownConfiguration(issuer.String()),
	}

	if ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseKeyID != "" {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, &oidc.OpenIDConnectWellKnownClaims{
			OpenIDConnectWellKnownSignedConfiguration: metadata,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        uuid.New().String(),
				Issuer:    issuer.String(),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		})

		kid := ctx.Providers.OpenIDConnect.KeyManager.GetKeyID(ctx, ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseKeyID, ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseAlg)

		token.Header[oidc.JWTHeaderKeyIdentifier] = kid

		if metadata.SignedMetadata, err = token.SignedString(ctx.Providers.OpenIDConnect.KeyManager.Get(ctx, kid, ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseAlg).PrivateJWK().Key); err != nil {
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

	metadata := oidc.OAuth2WellKnownSignedConfiguration{
		OAuth2WellKnownConfiguration: ctx.Providers.OpenIDConnect.GetOAuth2WellKnownConfiguration(issuer.String()),
	}

	if ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseKeyID != "" {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, &oidc.OAuth2WellKnownClaims{
			OAuth2WellKnownSignedConfiguration: metadata,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        uuid.New().String(),
				Issuer:    issuer.String(),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		})

		kid := ctx.Providers.OpenIDConnect.KeyManager.GetKeyID(ctx, ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseKeyID, ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseAlg)

		token.Header[oidc.JWTHeaderKeyIdentifier] = kid

		if metadata.SignedMetadata, err = token.SignedString(ctx.Providers.OpenIDConnect.KeyManager.Get(ctx, kid, ctx.Configuration.IdentityProviders.OIDC.DiscoverySignedResponseAlg).PrivateJWK().Key); err != nil {
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
