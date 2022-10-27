package oidc

import (
	"github.com/ory/fosite"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// NewClient creates a new Client.
func NewClient(config schema.OpenIDConnectClientConfiguration) (client *Client) {
	var secret []byte

	if config.Secret != nil {
		secret = []byte(config.Secret.Encode())
	}

	client = &Client{
		ID:               config.ID,
		Description:      config.Description,
		Secret:           secret,
		SectorIdentifier: config.SectorIdentifier.String(),
		Public:           config.Public,

		Audience:      config.Audience,
		Scopes:        config.Scopes,
		RedirectURIs:  config.RedirectURIs,
		GrantTypes:    config.GrantTypes,
		ResponseTypes: config.ResponseTypes,
		ResponseModes: []fosite.ResponseModeType{fosite.ResponseModeDefault},

		UserinfoSigningAlgorithm: config.UserinfoSigningAlgorithm,

		Policy: authorization.StringToLevel(config.Policy),

		Consent: NewClientConsent(config.ConsentMode, config.ConsentPreConfiguredDuration),
	}

	for _, mode := range config.ResponseModes {
		client.ResponseModes = append(client.ResponseModes, fosite.ResponseModeType(mode))
	}

	return client
}

// IsAuthenticationLevelSufficient returns if the provided authentication.Level is sufficient for the client of the AutheliaClient.
func (c *Client) IsAuthenticationLevelSufficient(level authentication.Level) bool {
	if level == authentication.NotAuthenticated {
		return false
	}

	return authorization.IsAuthLevelSufficient(level, c.Policy)
}

// GetID returns the ID.
func (c *Client) GetID() string {
	return c.ID
}

// GetSectorIdentifier returns the SectorIdentifier for this client.
func (c *Client) GetSectorIdentifier() string {
	return c.SectorIdentifier
}

// GetConsentResponseBody returns the proper consent response body for this session.OIDCWorkflowSession.
func (c *Client) GetConsentResponseBody(consent *model.OAuth2ConsentSession) ConsentGetResponseBody {
	body := ConsentGetResponseBody{
		ClientID:          c.ID,
		ClientDescription: c.Description,
		PreConfiguration:  c.Consent.Mode == ClientConsentModePreConfigured,
	}

	if consent != nil {
		body.Scopes = consent.RequestedScopes
		body.Audience = consent.RequestedAudience
	}

	return body
}

// GetHashedSecret returns the Secret.
func (c *Client) GetHashedSecret() (secret []byte) {
	return c.Secret
}

// GetRedirectURIs returns the RedirectURIs.
func (c *Client) GetRedirectURIs() (redirectURIs []string) {
	return c.RedirectURIs
}

// GetGrantTypes returns the GrantTypes.
func (c *Client) GetGrantTypes() fosite.Arguments {
	if len(c.GrantTypes) == 0 {
		return fosite.Arguments{"authorization_code"}
	}

	return c.GrantTypes
}

// GetResponseTypes returns the ResponseTypes.
func (c *Client) GetResponseTypes() fosite.Arguments {
	if len(c.ResponseTypes) == 0 {
		return fosite.Arguments{"code"}
	}

	return c.ResponseTypes
}

// GetScopes returns the Scopes.
func (c *Client) GetScopes() fosite.Arguments {
	return c.Scopes
}

// IsPublic returns the value of the Public property.
func (c *Client) IsPublic() bool {
	return c.Public
}

// GetAudience returns the Audience.
func (c *Client) GetAudience() fosite.Arguments {
	return c.Audience
}

// GetResponseModes returns the valid response modes for this client.
//
// Implements the fosite.ResponseModeClient.
func (c *Client) GetResponseModes() []fosite.ResponseModeType {
	return c.ResponseModes
}
