package oidc

import (
	"github.com/ory/fosite"
	"github.com/ory/x/errorsx"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// NewClient creates a new Client.
func NewClient(config schema.OpenIDConnectClientConfiguration) (client *Client) {
	client = &Client{
		ID:               config.ID,
		Description:      config.Description,
		Secret:           config.Secret,
		SectorIdentifier: config.SectorIdentifier.String(),
		Public:           config.Public,

		EnforcePKCE:                config.EnforcePKCE || config.PKCEChallengeMethod != "",
		EnforcePKCEChallengeMethod: config.PKCEChallengeMethod != "",
		PKCEChallengeMethod:        config.PKCEChallengeMethod,

		Audience:      config.Audience,
		Scopes:        config.Scopes,
		RedirectURIs:  config.RedirectURIs,
		GrantTypes:    config.GrantTypes,
		ResponseTypes: config.ResponseTypes,
		ResponseModes: []fosite.ResponseModeType{},

		EnforcePAR: config.EnforcePAR,

		UserinfoSigningAlgorithm: config.UserinfoSigningAlgorithm,

		Policy: authorization.NewLevel(config.Policy),

		Consent: NewClientConsent(config.ConsentMode, config.ConsentPreConfiguredDuration),
	}

	for _, mode := range config.ResponseModes {
		client.ResponseModes = append(client.ResponseModes, fosite.ResponseModeType(mode))
	}

	return client
}

// ValidatePKCEPolicy is a helper function to validate PKCE policy constraints on a per-client basis.
func (c *Client) ValidatePKCEPolicy(r fosite.Requester) (err error) {
	form := r.GetRequestForm()

	if c.EnforcePKCE {
		if form.Get(FormParameterCodeChallenge) == "" {
			return errorsx.WithStack(fosite.ErrInvalidRequest.
				WithHint("Clients must include a code_challenge when performing the authorize code flow, but it is missing.").
				WithDebug("The server is configured in a way that enforces PKCE for this client."))
		}

		if c.EnforcePKCEChallengeMethod {
			if method := form.Get(FormParameterCodeChallengeMethod); method != c.PKCEChallengeMethod {
				return errorsx.WithStack(fosite.ErrInvalidRequest.
					WithHintf("Client must use code_challenge_method=%s, %s is not allowed.", c.PKCEChallengeMethod, method).
					WithDebugf("The server is configured in a way that enforces PKCE %s as challenge method for this client.", c.PKCEChallengeMethod))
			}
		}
	}

	return nil
}

// ValidatePARPolicy is a helper function to validate additional policy constraints on a per-client basis.
func (c *Client) ValidatePARPolicy(r fosite.Requester, prefix string) (err error) {
	if c.EnforcePAR {
		if !IsPushedAuthorizedRequest(r, prefix) {
			switch requestURI := r.GetRequestForm().Get(FormParameterRequestURI); requestURI {
			case "":
				return errorsx.WithStack(ErrPAREnforcedClientMissingPAR.WithDebug("The request_uri parameter was empty."))
			default:
				return errorsx.WithStack(ErrPAREnforcedClientMissingPAR.WithDebugf("The request_uri parameter '%s' is malformed.", requestURI))
			}
		}
	}

	return nil
}

// ValidateResponseModePolicy is an additional check to the response mode parameter to ensure if it's omitted that the
// default response mode for the fosite.AuthorizeRequester is permitted.
func (c *Client) ValidateResponseModePolicy(r fosite.AuthorizeRequester) (err error) {
	if r.GetResponseMode() != fosite.ResponseModeDefault {
		return nil
	}

	mode := r.GetDefaultResponseMode()

	for _, m := range c.GetResponseModes() {
		if m == mode {
			return nil
		}
	}

	return errorsx.WithStack(fosite.ErrUnsupportedResponseMode.WithHintf(`The request omitted the response_mode making the default response_mode "%s" based on the other authorization request parameters but registered OAuth 2.0 client doesn't support this response_mode`, mode))
}

// IsAuthenticationLevelSufficient returns if the provided authentication.Level is sufficient for the client of the AutheliaClient.
func (c *Client) IsAuthenticationLevelSufficient(level authentication.Level) bool {
	if level == authentication.NotAuthenticated {
		return false
	}

	return authorization.IsAuthLevelSufficient(level, c.Policy)
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

// GetID returns the ID.
func (c *Client) GetID() string {
	return c.ID
}

// GetHashedSecret returns the Secret.
func (c *Client) GetHashedSecret() (secret []byte) {
	if c.Secret == nil {
		return []byte(nil)
	}

	return []byte(c.Secret.Encode())
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
