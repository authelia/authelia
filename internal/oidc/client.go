package oidc

import (
	"github.com/ory/fosite"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
)

// AutheliaClient represents the client internally.
type AutheliaClient struct {
	ID            string              `json:"id"`
	Description   string              `json:"-"`
	Secret        []byte              `json:"client_secret,omitempty"`
	RedirectURIs  []string            `json:"redirect_uris"`
	GrantTypes    []string            `json:"grant_types"`
	ResponseTypes []string            `json:"response_types"`
	Scopes        []string            `json:"scopes"`
	Audience      []string            `json:"audience"`
	Public        bool                `json:"public"`
	Policy        authorization.Level `json:"-"`
}

// IsAuthenticationLevelSufficient returns if the provided authentication.Level is sufficient for the client of the AutheliaClient.
func (c AutheliaClient) IsAuthenticationLevelSufficient(level authentication.Level) bool {
	return authorization.IsAuthLevelSufficient(level, c.Policy)
}

// GetID returns the ID of the AutheliaClient.
func (c AutheliaClient) GetID() string {
	return c.ID
}

// GetHashedSecret returns the Secret of the AutheliaClient.
func (c AutheliaClient) GetHashedSecret() []byte {
	return c.Secret
}

// GetRedirectURIs returns the RedirectURIs of the AutheliaClient.
func (c AutheliaClient) GetRedirectURIs() []string {
	return c.RedirectURIs
}

// GetGrantTypes returns the GrantTypes of the AutheliaClient.
func (c AutheliaClient) GetGrantTypes() fosite.Arguments {
	if len(c.GrantTypes) == 0 {
		return fosite.Arguments{"authorization_code"}
	}

	return c.GrantTypes
}

// GetResponseTypes returns the ResponseTypes of the AutheliaClient.
func (c AutheliaClient) GetResponseTypes() fosite.Arguments {
	if len(c.ResponseTypes) == 0 {
		return fosite.Arguments{"code"}
	}

	return c.ResponseTypes
}

// GetScopes returns the Scopes of the AutheliaClient.
func (c AutheliaClient) GetScopes() fosite.Arguments {
	return c.Scopes
}

// IsPublic returns the value of the Public property of the AutheliaClient.
func (c AutheliaClient) IsPublic() bool {
	return c.Public
}

// GetAudience returns the Audience of the AutheliaClient.
func (c AutheliaClient) GetAudience() fosite.Arguments {
	return c.Audience
}
