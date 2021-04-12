package oidc

import (
	"crypto/rsa"
	"fmt"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/storage"
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

	composeConfiguration *compose.Config
	Fosite               fosite.OAuth2Provider
	storage              fosite.Storage
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

	provider.composeConfiguration = new(compose.Config)

	key, err := utils.ParseRsaPrivateKeyFromPemStr(configuration.IssuerPrivateKey)
	if err != nil {
		return provider, fmt.Errorf("unable to parse the private key of the OpenID issuer: %w", err)
	}

	provider.privateKeys = make(map[string]*rsa.PrivateKey)
	provider.privateKeys["main-key"] = key

	provider.Fosite = compose.ComposeAllEnabled(
		provider.composeConfiguration,
		provider.storage,
		[]byte(utils.HashSHA256FromString(configuration.HMACSecret)),
		provider.privateKeys["main-key"])

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
