package schema

import (
	"crypto/rsa"
	"net/url"
	"time"
)

// IdentityProvidersConfiguration represents the IdentityProviders 2.0 configuration for Authelia.
type IdentityProvidersConfiguration struct {
	OIDC *OpenIDConnectConfiguration `koanf:"oidc"`
}

// OpenIDConnectConfiguration configuration for OpenID Connect.
type OpenIDConnectConfiguration struct {
	HMACSecret             string               `koanf:"hmac_secret"`
	IssuerCertificateChain X509CertificateChain `koanf:"issuer_certificate_chain"`
	IssuerPrivateKey       *rsa.PrivateKey      `koanf:"issuer_private_key"`

	AccessTokenLifespan   time.Duration `koanf:"access_token_lifespan"`
	AuthorizeCodeLifespan time.Duration `koanf:"authorize_code_lifespan"`
	IDTokenLifespan       time.Duration `koanf:"id_token_lifespan"`
	RefreshTokenLifespan  time.Duration `koanf:"refresh_token_lifespan"`

	EnableClientDebugMessages bool `koanf:"enable_client_debug_messages"`
	MinimumParameterEntropy   int  `koanf:"minimum_parameter_entropy"`

	EnforcePKCE              string `koanf:"enforce_pkce"`
	EnablePKCEPlainChallenge bool   `koanf:"enable_pkce_plain_challenge"`

	CORS OpenIDConnectCORSConfiguration `koanf:"cors"`

	Clients []OpenIDConnectClientConfiguration `koanf:"clients"`
}

// OpenIDConnectCORSConfiguration represents an OpenID Connect CORS config.
type OpenIDConnectCORSConfiguration struct {
	Endpoints      []string  `koanf:"endpoints"`
	AllowedOrigins []url.URL `koanf:"allowed_origins"`

	AllowedOriginsFromClientRedirectURIs bool `koanf:"allowed_origins_from_client_redirect_uris"`
}

// OpenIDConnectClientConfiguration configuration for an OpenID Connect client.
type OpenIDConnectClientConfiguration struct {
	ID               string  `koanf:"id"`
	Description      string  `koanf:"description"`
	Secret           string  `koanf:"secret"`
	SectorIdentifier url.URL `koanf:"sector_identifier"`
	Public           bool    `koanf:"public"`

	RedirectURIs []string `koanf:"redirect_uris"`

	Audience      []string `koanf:"audience"`
	Scopes        []string `koanf:"scopes"`
	GrantTypes    []string `koanf:"grant_types"`
	ResponseTypes []string `koanf:"response_types"`
	ResponseModes []string `koanf:"response_modes"`

	UserinfoSigningAlgorithm string `koanf:"userinfo_signing_algorithm"`

	Policy string `koanf:"authorization_policy"`

	ConsentMode                  string         `koanf:"consent_mode"`
	ConsentPreConfiguredDuration *time.Duration `koanf:"pre_configured_consent_duration"`
}

// DefaultOpenIDConnectConfiguration contains defaults for OIDC.
var DefaultOpenIDConnectConfiguration = OpenIDConnectConfiguration{
	AccessTokenLifespan:   time.Hour,
	AuthorizeCodeLifespan: time.Minute,
	IDTokenLifespan:       time.Hour,
	RefreshTokenLifespan:  time.Minute * 90,
	EnforcePKCE:           "public_clients_only",
}

var defaultOIDCClientConsentPreConfiguredDuration = time.Hour * 24 * 7

// DefaultOpenIDConnectClientConfiguration contains defaults for OIDC Clients.
var DefaultOpenIDConnectClientConfiguration = OpenIDConnectClientConfiguration{
	Policy:        "two_factor",
	Scopes:        []string{"openid", "groups", "profile", "email"},
	GrantTypes:    []string{"refresh_token", "authorization_code"},
	ResponseTypes: []string{"code"},
	ResponseModes: []string{"form_post", "query", "fragment"},

	UserinfoSigningAlgorithm:     "none",
	ConsentMode:                  "auto",
	ConsentPreConfiguredDuration: &defaultOIDCClientConsentPreConfiguredDuration,
}
