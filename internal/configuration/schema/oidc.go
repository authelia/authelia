package schema

// OpenIDConnectConfiguration configuration for OpenID Connect.
type OpenIDConnectConfiguration struct {
	// This secret must be 32 bytes long
	OAuth2HMACSecret string `mapstructure:"oauth2_hmac_secret"`
	// This is a path because viper strip new lines of the private key preventing the crypto lib to parse it properly.
	// TODO: find a way to not strip the new lines
	OIDCIssuerPrivateKeyPath string `mapstructure:"oidc_issuer_private_key_path"`

	Clients []OpenIDConnectClientConfiguration `mapstructure:"clients"`
}

// OpenIDConnectClientConfiguration configuration for an OpenID Connect client.
type OpenIDConnectClientConfiguration struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURIs []string `mapstructure:"redirect_uris"`
	Policy       string   `mapstructure:"policy"`
}
