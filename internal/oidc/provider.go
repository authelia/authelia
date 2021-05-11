package oidc

import (
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/herodot"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// OpenIDConnectProvider for OpenID Connect.
type OpenIDConnectProvider struct {
	Fosite  fosite.OAuth2Provider
	Store   *OpenIDConnectStore
	Herodot *herodot.JSONWriter
}

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(configuration *schema.OpenIDConnectConfiguration) (provider OpenIDConnectProvider, err error) {
	provider = OpenIDConnectProvider{
		Fosite: nil,
	}

	if configuration == nil {
		return provider, nil
	}

	provider.Store, err = NewOpenIDConnectStore(configuration)
	if err != nil {
		return provider, err
	}

	composeConfiguration := &compose.Config{
		AccessTokenLifespan:        configuration.AccessTokenLifespan,
		AuthorizeCodeLifespan:      configuration.AuthorizeCodeLifespan,
		IDTokenLifespan:            configuration.IDTokenLifespan,
		RefreshTokenLifespan:       configuration.RefreshTokenLifespan,
		SendDebugMessagesToClients: configuration.EnableClientDebugMessages,
	}

	key, err := provider.Store.KeyManager.GetActivePrivateKey()
	if err != nil {
		return provider, err
	}

	strategy := &compose.CommonStrategy{
		CoreStrategy: compose.NewOAuth2HMACStrategy(
			composeConfiguration,
			[]byte(utils.HashSHA256FromString(configuration.HMACSecret)),
			nil,
		),
		OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(
			composeConfiguration,
			key,
		),
		JWTStrategy: provider.Store.KeyManager.Strategy(),
	}

	provider.Herodot = herodot.NewJSONWriter(nil)
	provider.Fosite = compose.Compose(
		composeConfiguration,
		provider.Store,
		strategy,
		AutheliaHasher{},

		/*
			These are the OAuth2 and OpenIDConnect factories. Order is important (the OAuth2 factories at the top must
			be before the OpenIDConnect factories) and taken directly from fosite.compose.ComposeAllEnabled. The
			commented factories are not enabled as we don't yet use them but are still here for reference purposes.
		*/
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2AuthorizeImplicitFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		compose.OAuth2ResourceOwnerPasswordCredentialsFactory,
		// compose.RFC7523AssertionGrantFactory,

		compose.OpenIDConnectExplicitFactory,
		compose.OpenIDConnectImplicitFactory,
		compose.OpenIDConnectHybridFactory,
		compose.OpenIDConnectRefreshFactory,

		compose.OAuth2TokenIntrospectionFactory,
		compose.OAuth2TokenRevocationFactory,

		// compose.OAuth2PKCEFactory,
	)

	return provider, nil
}
