package oidc

import (
	"crypto/sha512"
	"fmt"

	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/hmac"
	"github.com/ory/herodot"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(config *schema.OpenIDConnectConfiguration, store storage.Provider) (provider *OpenIDConnectProvider, err error) {
	if config == nil {
		return nil, nil
	}

	provider = &OpenIDConnectProvider{
		JSONWriter:  herodot.NewJSONWriter(nil),
		Store:       NewOpenIDConnectStore(config, store),
		KeyStrategy: NewKeyStrategy(config),
	}

	cconfig := &compose.Config{
		AccessTokenLifespan:            config.AccessTokenLifespan,
		AuthorizeCodeLifespan:          config.AuthorizeCodeLifespan,
		IDTokenLifespan:                config.IDTokenLifespan,
		RefreshTokenLifespan:           config.RefreshTokenLifespan,
		SendDebugMessagesToClients:     config.EnableClientDebugMessages,
		MinParameterEntropy:            config.MinimumParameterEntropy,
		EnforcePKCE:                    config.EnforcePKCE == "always",
		EnforcePKCEForPublicClients:    config.EnforcePKCE != "never",
		EnablePKCEPlainChallengeMethod: config.EnablePKCEPlainChallenge,
	}

	strategy := &compose.CommonStrategy{
		CoreStrategy: &oauth2.HMACSHAStrategy{
			Enigma: &hmac.HMACStrategy{
				GlobalSecret:         []byte(utils.HashSHA256FromString(config.HMACSecret)),
				RotatedGlobalSecrets: nil,
				TokenEntropy:         cconfig.GetTokenEntropy(),
				Hash:                 sha512.New512_256,
			},
			AccessTokenLifespan:   cconfig.GetAccessTokenLifespan(),
			AuthorizeCodeLifespan: cconfig.GetAuthorizeCodeLifespan(),
			RefreshTokenLifespan:  cconfig.GetRefreshTokenLifespan(),
		},
		OpenIDConnectTokenStrategy: &openid.DefaultStrategy{
			JWTStrategy:         provider.KeyStrategy,
			Expiry:              cconfig.GetIDTokenLifespan(),
			Issuer:              cconfig.IDTokenIssuer,
			MinParameterEntropy: cconfig.GetMinParameterEntropy(),
		},
		JWTStrategy: provider.KeyStrategy,
	}

	provider.OAuth2Provider = compose.Compose(
		cconfig,
		provider.Store,
		strategy,
		PlainTextHasher{},

		/*
			These are the OAuth2 and OpenIDConnect factories. Order is important (the OAuth2 factories at the top must
			be before the OpenIDConnect factories) and taken directly from fosite.compose.ComposeAllEnabled. The
			commented factories are not enabled as we don't yet use them but are still here for reference purposes.
		*/
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2AuthorizeImplicitFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		// compose.OAuth2ResourceOwnerPasswordCredentialsFactory,
		// compose.RFC7523AssertionGrantFactory,.

		compose.OpenIDConnectExplicitFactory,
		compose.OpenIDConnectImplicitFactory,
		compose.OpenIDConnectHybridFactory,
		compose.OpenIDConnectRefreshFactory,

		compose.OAuth2TokenIntrospectionFactory,
		compose.OAuth2TokenRevocationFactory,

		compose.OAuth2PKCEFactory,
	)

	provider.discovery = NewOpenIDConnectWellKnownConfiguration(config.EnablePKCEPlainChallenge, provider.Store.clients, provider.KeyStrategy.SigningAlgValues())

	return provider, nil
}

// GetOAuth2WellKnownConfiguration returns the discovery document for the OAuth Configuration.
func (p *OpenIDConnectProvider) GetOAuth2WellKnownConfiguration(issuer string) OAuth2WellKnownConfiguration {
	options := OAuth2WellKnownConfiguration{
		CommonDiscoveryOptions: p.discovery.CommonDiscoveryOptions,
		OAuth2DiscoveryOptions: p.discovery.OAuth2DiscoveryOptions,
	}

	options.Issuer = issuer
	options.JWKSURI = fmt.Sprintf("%s%s", issuer, EndpointPathJWKs)

	options.IntrospectionEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathIntrospection)
	options.TokenEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathToken)

	options.AuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathAuthorization)
	options.RevocationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathRevocation)

	return options
}

// GetOpenIDConnectWellKnownConfiguration returns the discovery document for the OpenID Configuration.
func (p *OpenIDConnectProvider) GetOpenIDConnectWellKnownConfiguration(issuer string) OpenIDConnectWellKnownConfiguration {
	options := OpenIDConnectWellKnownConfiguration{
		CommonDiscoveryOptions:                          p.discovery.CommonDiscoveryOptions,
		OAuth2DiscoveryOptions:                          p.discovery.OAuth2DiscoveryOptions,
		OpenIDConnectDiscoveryOptions:                   p.discovery.OpenIDConnectDiscoveryOptions,
		OpenIDConnectFrontChannelLogoutDiscoveryOptions: p.discovery.OpenIDConnectFrontChannelLogoutDiscoveryOptions,
		OpenIDConnectBackChannelLogoutDiscoveryOptions:  p.discovery.OpenIDConnectBackChannelLogoutDiscoveryOptions,
	}

	options.Issuer = issuer
	options.JWKSURI = fmt.Sprintf("%s%s", issuer, EndpointPathJWKs)

	options.IntrospectionEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathIntrospection)
	options.TokenEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathToken)

	options.AuthorizationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathAuthorization)
	options.RevocationEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathRevocation)
	options.UserinfoEndpoint = fmt.Sprintf("%s%s", issuer, EndpointPathUserinfo)

	return options
}
