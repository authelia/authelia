package schema

// OAuthConfiguration represents the OAuth 2.0 configuration for Authelia.
type OAuthConfiguration struct {
	OIDCServer *OpenIDConnectServerConfiguration `mapstructure:"oidc_server"`
}

// OpenIDConnectServerConfiguration configuration for OpenID Connect.
type OpenIDConnectServerConfiguration struct {
	// This secret must be 32 bytes long
	HMACSecret string `mapstructure:"hmac_secret"`
	// This is a path because viper strip new lines of the private key preventing the crypto lib to parse it properly.
	// TODO: find a way to not strip the new lines
	IssuerPrivateKeyPath string `mapstructure:"issuer_private_key_path"`

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
