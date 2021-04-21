package oidc

import (
	"crypto/rsa"
	"fmt"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

// OpenIDConnectProvider for OpenID Connect.
type OpenIDConnectProvider struct {
	clients     map[string]*InternalClient
	privateKeys map[string]*rsa.PrivateKey

	Fosite  fosite.OAuth2Provider
	storage fosite.Storage
}

// NewOpenIDConnectProvider new-ups a OpenIDConnectProvider.
func NewOpenIDConnectProvider(configuration *schema.OpenIDConnectConfiguration) (provider OpenIDConnectProvider, err error) {
	provider = OpenIDConnectProvider{
		Fosite: nil,
	}

	if configuration == nil {
		return provider, nil
	}

	clients := make(map[string]fosite.Client)
	provider.clients = make(map[string]*InternalClient)

	for _, client := range configuration.Clients {
		policy := authorization.PolicyToLevel(client.Policy)
		logging.Logger().Debugf("Registering client %s with policy %s (%v)", client.ID, client.Policy, policy)

		provider.clients[client.ID] = &InternalClient{
			ID:            client.ID,
			Description:   client.Description,
			Policy:        authorization.PolicyToLevel(client.Policy),
			Secret:        []byte(client.Secret),
			RedirectURIs:  client.RedirectURIs,
			GrantTypes:    client.GrantTypes,
			ResponseTypes: client.ResponseTypes,
			Scopes:        client.Scopes,
		}
		clients[client.ID] = provider.clients[client.ID]
	}

	// TODO: Implement our own storage mapping.
	provider.storage = &storage.MemoryStore{
		IDSessions:             make(map[string]fosite.Requester),
		Clients:                clients,
		Users:                  map[string]storage.MemoryUserRelation{},
		AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
		AccessTokens:           map[string]fosite.Requester{},
		RefreshTokens:          map[string]storage.StoreRefreshToken{},
		PKCES:                  map[string]fosite.Requester{},
		AccessTokenRequestIDs:  map[string]string{},
		RefreshTokenRequestIDs: map[string]string{},
	}

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
		provider.storage,
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

// GetClient returns the AutheliaClient matching the id provided if it exists.
func (p OpenIDConnectProvider) GetClient(id string) (config *InternalClient) {
	if p.IsValidClientID(id) {
		return p.clients[id]
	}

	return nil
}

// IsValidClientID returns true if the provided id exists in the OpenIDConnectProvider.Clients map.
func (p OpenIDConnectProvider) IsValidClientID(id string) (valid bool) {
	if _, ok := p.clients[id]; ok {
		return true
	}

	return false
}
