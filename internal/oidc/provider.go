package oidc

import (
	"crypto/rsa"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/storage"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// New new-ups a OpenIDConnectProvider.
func New(configuration *schema.OpenIDConnectConfiguration) (provider OpenIDConnectProvider) {
	provider = OpenIDConnectProvider{
		Fosite: nil,
	}

	if configuration == nil {
		return provider
	}

	var err error

	clients := make(map[string]fosite.Client)
	provider.Clients = make(map[string]*AutheliaClient)

	for _, client := range configuration.Clients {
		provider.Clients[client.ID] = &AutheliaClient{
			ID:            client.ID,
			Description:   client.Description,
			Policy:        authorization.PolicyToLevel(client.Policy),
			Secret:        []byte(client.Secret),
			RedirectURIs:  client.RedirectURIs,
			GrantTypes:    client.GrantTypes,
			ResponseTypes: client.ResponseTypes,
			Scopes:        client.Scopes,
		}
		clients[client.ID] = provider.Clients[client.ID]
	}

	provider.Storage = &storage.MemoryStore{
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

	provider.ComposeConfiguration = new(compose.Config)

	key, err := utils.ParseRsaPrivateKeyFromPemStr(configuration.IssuerPrivateKey)
	if err != nil {
		panic(err)
	}

	provider.PrivateKeys = make(map[string]*rsa.PrivateKey)
	provider.PrivateKeys["main-key"] = key

	provider.Fosite = compose.ComposeAllEnabled(
		provider.ComposeConfiguration,
		provider.Storage,
		[]byte(configuration.HMACSecret),
		provider.PrivateKeys["main-key"])

	return provider
}

// OpenIDConnectProvider for OpenID Connect.
type OpenIDConnectProvider struct {
	Clients     map[string]*AutheliaClient
	PrivateKeys map[string]*rsa.PrivateKey

	ComposeConfiguration *compose.Config
	Fosite               fosite.OAuth2Provider
	Storage              fosite.Storage
}

// GetKeySet returns the jose.JSONWebKeySet for the OpenIDConnectProvider.
func (p OpenIDConnectProvider) GetKeySet() (webKeySet jose.JSONWebKeySet) {
	for keyID, key := range p.PrivateKeys {
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
func (p OpenIDConnectProvider) GetClient(id string) (config *AutheliaClient) {
	if p.IsValidClientID(id) {
		return p.Clients[id]
	}

	return nil
}

// IsValidClientID returns true if the provided id exists in the OpenIDConnectProvider.Clients map.
func (p OpenIDConnectProvider) IsValidClientID(id string) (valid bool) {
	if _, ok := p.Clients[id]; ok {
		return true
	}

	return false
}

// IsAuthenticationLevelSufficient returns a bool provided a clientID and authentication.Level.
func (p OpenIDConnectProvider) IsAuthenticationLevelSufficient(clientID string, level authentication.Level) bool {
	if client, ok := p.Clients[clientID]; ok {
		return client.IsAuthenticationLevelSufficient(level)
	}

	return false
}
