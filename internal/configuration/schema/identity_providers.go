package schema

// IdentityProvidersConfiguration represents the IdentityProviders 2.0 configuration for Authelia.
type IdentityProvidersConfiguration struct {
	OIDC *OpenIDConnectConfiguration `koanf:"oidc"`
}

// OpenIDConnectConfiguration configuration for OpenID Connect.
type OpenIDConnectConfiguration struct {
	// This secret must be 32 bytes long
	HMACSecret       string `koanf:"hmac_secret"`
	IssuerPrivateKey string `koanf:"issuer_private_key"`

	Clients []OpenIDConnectClientConfiguration `koanf:"clients"`
}

// OpenIDConnectClientConfiguration configuration for an OpenID Connect client.
type OpenIDConnectClientConfiguration struct {
	ID            string   `koanf:"id"`
	Description   string   `koanf:"description"`
	Secret        string   `koanf:"secret"`
	RedirectURIs  []string `koanf:"redirect_uris"`
	Policy        string   `koanf:"authorization_policy"`
	Scopes        []string `koanf:"scopes"`
	GrantTypes    []string `koanf:"grant_types"`
	ResponseTypes []string `koanf:"response_types"`
}

// DefaultOpenIDConnectClientConfiguration contains defaults for OIDC AutheliaClients.
var DefaultOpenIDConnectClientConfiguration = OpenIDConnectClientConfiguration{
	Scopes:        []string{"openid", "groups", "profile", "email"},
	ResponseTypes: []string{"code"},
	GrantTypes:    []string{"refresh_token", "authorization_code"},
	Policy:        "two_factor",
}
