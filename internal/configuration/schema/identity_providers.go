package schema

import (
	"crypto/rsa"
	"net/url"
	"time"
)

// IdentityProviders represents the Identity Providers configuration for Authelia.
type IdentityProviders struct {
	OIDC *OpenIDConnect `koanf:"oidc"`
}

// OpenIDConnect configuration for OpenID Connect 1.0.
type OpenIDConnect struct {
	HMACSecret        string `koanf:"hmac_secret"`
	IssuerPrivateKeys []JWK  `koanf:"issuer_private_keys"`

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

	PAR  OpenIDConnectPAR  `koanf:"pushed_authorizations"`
	CORS OpenIDConnectCORS `koanf:"cors"`

	Clients []OpenIDConnectClient `koanf:"clients"`

	Discovery OpenIDConnectDiscovery // MetaData value. Not configurable by users.
}

// OpenIDConnectDiscovery is information discovered during validation reused for the discovery handlers.
type OpenIDConnectDiscovery struct {
	DefaultKeyIDs               map[string]string
	DefaultKeyID                string
	ResponseObjectSigningKeyIDs []string
	ResponseObjectSigningAlgs   []string
	RequestObjectSigningAlgs    []string
}

// OpenIDConnectPAR represents an OpenID Connect 1.0 PAR config.
type OpenIDConnectPAR struct {
	Enforce         bool          `koanf:"enforce"`
	ContextLifespan time.Duration `koanf:"context_lifespan"`
}

// OpenIDConnectCORS represents an OpenID Connect 1.0 CORS config.
type OpenIDConnectCORS struct {
	Endpoints      []string  `koanf:"endpoints"`
	AllowedOrigins []url.URL `koanf:"allowed_origins"`

	AllowedOriginsFromClientRedirectURIs bool `koanf:"allowed_origins_from_client_redirect_uris"`
}

// OpenIDConnectClient represents a configuration for an OpenID Connect 1.0 client.
type OpenIDConnectClient struct {
	ID               string          `koanf:"id"`
	Description      string          `koanf:"description"`
	Secret           *PasswordDigest `koanf:"secret"`
	SectorIdentifier url.URL         `koanf:"sector_identifier"`
	Public           bool            `koanf:"public"`

	RedirectURIs []string `koanf:"redirect_uris"`

	Audience      []string `koanf:"audience"`
	Scopes        []string `koanf:"scopes"`
	GrantTypes    []string `koanf:"grant_types"`
	ResponseTypes []string `koanf:"response_types"`
	ResponseModes []string `koanf:"response_modes"`

	Policy string `koanf:"authorization_policy"`

	ConsentMode                  string         `koanf:"consent_mode"`
	ConsentPreConfiguredDuration *time.Duration `koanf:"pre_configured_consent_duration"`

	EnforcePAR  bool `koanf:"enforce_par"`
	EnforcePKCE bool `koanf:"enforce_pkce"`

	PKCEChallengeMethod string `koanf:"pkce_challenge_method"`

	IDTokenSigningAlg           string `koanf:"id_token_signing_alg"`
	IDTokenSigningKeyID         string `koanf:"id_token_signing_key_id"`
	UserinfoSigningAlg          string `koanf:"userinfo_signing_alg"`
	UserinfoSigningKeyID        string `koanf:"userinfo_signing_key_id"`
	RequestObjectSigningAlg     string `koanf:"request_object_signing_alg"`
	TokenEndpointAuthSigningAlg string `koanf:"token_endpoint_auth_signing_alg"`

	TokenEndpointAuthMethod string `koanf:"token_endpoint_auth_method"`

	PublicKeys OpenIDConnectClientPublicKeys `koanf:"public_keys"`

	Discovery OpenIDConnectDiscovery
}

// OpenIDConnectClientPublicKeys represents the Client Public Keys configuration for an OpenID Connect 1.0 client.
type OpenIDConnectClientPublicKeys struct {
	URI    *url.URL `koanf:"uri"`
	Values []JWK    `koanf:"values"`
}

// DefaultOpenIDConnectConfiguration contains defaults for OIDC.
var DefaultOpenIDConnectConfiguration = OpenIDConnect{
	AccessTokenLifespan:   time.Hour,
	AuthorizeCodeLifespan: time.Minute,
	IDTokenLifespan:       time.Hour,
	RefreshTokenLifespan:  time.Minute * 90,
	EnforcePKCE:           "public_clients_only",
}

var defaultOIDCClientConsentPreConfiguredDuration = time.Hour * 24 * 7

// DefaultOpenIDConnectClientConfiguration contains defaults for OIDC Clients.
var DefaultOpenIDConnectClientConfiguration = OpenIDConnectClient{
	Policy:                       "two_factor",
	Scopes:                       []string{"openid", "groups", "profile", "email"},
	ResponseTypes:                []string{"code"},
	ResponseModes:                []string{"form_post"},
	IDTokenSigningAlg:            "RS256",
	UserinfoSigningAlg:           "none",
	ConsentMode:                  "auto",
	ConsentPreConfiguredDuration: &defaultOIDCClientConsentPreConfiguredDuration,
}
