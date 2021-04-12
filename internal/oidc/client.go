package oidc

import (
	"github.com/ory/fosite"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
)

// InternalClient represents the client internally.
type InternalClient struct {
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
func (c InternalClient) IsAuthenticationLevelSufficient(level authentication.Level) bool {
	return authorization.IsAuthLevelSufficient(level, c.Policy)
}

// GetID returns the ID.
func (c InternalClient) GetID() string {
	return c.ID
}

// GetHashedSecret returns the Secret.
func (c InternalClient) GetHashedSecret() []byte {
	return c.Secret
}

// GetRedirectURIs returns the RedirectURIs.
func (c InternalClient) GetRedirectURIs() []string {
	return c.RedirectURIs
}

// GetGrantTypes returns the GrantTypes.
func (c InternalClient) GetGrantTypes() fosite.Arguments {
	if len(c.GrantTypes) == 0 {
		return fosite.Arguments{"authorization_code"}
	}

	return c.GrantTypes
}

// GetResponseTypes returns the ResponseTypes.
func (c InternalClient) GetResponseTypes() fosite.Arguments {
	if len(c.ResponseTypes) == 0 {
		return fosite.Arguments{"code"}
	}

	return c.ResponseTypes
}

// GetScopes returns the Scopes.
func (c InternalClient) GetScopes() fosite.Arguments {
	return c.Scopes
}

// IsPublic returns the value of the Public property.
func (c InternalClient) IsPublic() bool {
	return c.Public
}

// GetAudience returns the Audience.
func (c InternalClient) GetAudience() fosite.Arguments {
	return c.Audience
}
