package schema

// OpenIDConnectConfiguration configuration for OpenID Connect.
type OpenIDConnectConfiguration struct {
	OAuth2HMACSecret     string `mapstructure:"oauth2_hmac_secret"`
	OIDCIssuerPrivateKey string `mapstructure:"oidc_issuer_private_key"`
}
