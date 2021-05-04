package oidc

import (
	"crypto/rsa"
	"fmt"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/token/jwt"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// OpenIDConnectProvider for OpenID Connect.
type OpenIDConnectProvider struct {
	privateKeys map[string]*rsa.PrivateKey

	Fosite fosite.OAuth2Provider
	Store  *OpenIDConnectStore
}

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(configuration *schema.OpenIDConnectConfiguration) (provider OpenIDConnectProvider, err error) {
	provider = OpenIDConnectProvider{
		Fosite: nil,
	}

	if configuration == nil {
		return provider, nil
	}

	provider.Store = NewOpenIDConnectStore(configuration)

	composeConfiguration := new(compose.Config)

	key, err := utils.ParseRsaPrivateKeyFromPemStr(configuration.IssuerPrivateKey)
	if err != nil {
		return provider, fmt.Errorf("unable to parse the private key of the OpenID issuer: %w", err)
	}

	provider.privateKeys = make(map[string]*rsa.PrivateKey)
	provider.privateKeys["main-key"] = key

	// TODO: Consider implementing RS512 as well.
	jwtStrategy := &jwt.RS256JWTStrategy{PrivateKey: key}

	strategy := &compose.CommonStrategy{
		CoreStrategy: compose.NewOAuth2HMACStrategy(
			composeConfiguration,
			[]byte(utils.HashSHA256FromString(configuration.HMACSecret)),
			nil,
		),
		OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(
			composeConfiguration,
			provider.privateKeys["main-key"],
		),
		JWTStrategy: jwtStrategy,
	}

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

// GetKeySet returns the jose.JSONWebKeySet for the OpenIDConnectProvider.
func (p OpenIDConnectProvider) GetKeySet() (webKeySet jose.JSONWebKeySet) {
	for keyID, key := range p.privateKeys {
		webKey := jose.JSONWebKey{
			Key:       &key.PublicKey,
			KeyID:     keyID,
			Algorithm: "RS256",
			Use:       "sig",
		}

		webKeySet.Keys = append(webKeySet.Keys, webKey)
	}

	return webKeySet
}
