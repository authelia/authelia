package schema

// IdentityProvidersConfiguration represents the IdentityProviders 2.0 configuration for Authelia.
type IdentityProvidersConfiguration struct {
	OIDCServer *OpenIDConnectConfiguration `mapstructure:"oidc"`
}

// OpenIDConnectConfiguration configuration for OpenID Connect.
type OpenIDConnectConfiguration struct {
	// This secret must be 32 bytes long
	HMACSecret       string `mapstructure:"hmac_secret"`
	IssuerPrivateKey string `mapstructure:"issuer_private_key"`

	Clients []OpenIDConnectClientConfiguration `mapstructure:"clients"`
}

// OpenIDConnectClientConfiguration configuration for an OpenID Connect client.
type OpenIDConnectClientConfiguration struct {
	ID            string   `mapstructure:"id"`
	Secret        string   `mapstructure:"secret"`
	RedirectURIs  []string `mapstructure:"redirect_uris"`
	Policy        string   `mapstructure:"policy"`
	Scopes        []string `mapstructure:"scopes"`
	GrantTypes    []string `mapstructure:"grant_types"`
	ResponseTypes []string `mapstructure:"response_types"`
}

// DefaultOpenIDConnectClientConfiguration contains defaults for OIDC Clients.
var DefaultOpenIDConnectClientConfiguration = OpenIDConnectClientConfiguration{
	Scopes:        []string{"openid"},
	ResponseTypes: []string{"code"},
	GrantTypes:    []string{"refresh_token", "authorization_code"},
}
