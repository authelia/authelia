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

	EnableClientDebugMessages bool `koanf:"enable_client_debug_messages"`
	MinimumParameterEntropy   int  `koanf:"minimum_parameter_entropy"`

	EnforcePKCE              string `koanf:"enforce_pkce"`
	EnablePKCEPlainChallenge bool   `koanf:"enable_pkce_plain_challenge"`

	PAR  OpenIDConnectPAR  `koanf:"pushed_authorizations"`
	CORS OpenIDConnectCORS `koanf:"cors"`

	Clients []OpenIDConnectClient `koanf:"clients"`

	AuthorizationPolicies map[string]OpenIDConnectPolicy `koanf:"authorization_policies"`
	Lifespans             OpenIDConnectLifespans         `koanf:"lifespans"`

	Discovery OpenIDConnectDiscovery // MetaData value. Not configurable by users.
}

// OpenIDConnectPolicy configuration for OpenID Connect 1.0 authorization policies.
type OpenIDConnectPolicy struct {
	DefaultPolicy string `koanf:"default_policy"`

	Rules []OpenIDConnectPolicyRule `koanf:"rules"`
}

// OpenIDConnectPolicyRule configuration for OpenID Connect 1.0 authorization policies rules.
type OpenIDConnectPolicyRule struct {
	Policy   string     `koanf:"policy"`
	Subjects [][]string `koanf:"subject"`
}

// OpenIDConnectDiscovery is information discovered during validation reused for the discovery handlers.
type OpenIDConnectDiscovery struct {
	AuthorizationPolicies       []string
	Lifespans                   []string
	DefaultKeyIDs               map[string]string
	DefaultKeyID                string
	ResponseObjectSigningKeyIDs []string
	ResponseObjectSigningAlgs   []string
	RequestObjectSigningAlgs    []string
}

type OpenIDConnectLifespans struct {
	OpenIDConnectLifespanToken `koanf:",squash"`
	Custom                     map[string]OpenIDConnectLifespan `koanf:"custom"`
}

type OpenIDConnectLifespan struct {
	OpenIDConnectLifespanToken `koanf:",squash"`

	Grants OpenIDConnectLifespanGrants `koanf:"grants"`
}

type OpenIDConnectLifespanGrants struct {
	AuthorizeCode     OpenIDConnectLifespanToken `koanf:"authorize_code"`
	Implicit          OpenIDConnectLifespanToken `koanf:"implicit"`
	ClientCredentials OpenIDConnectLifespanToken `koanf:"client_credentials"`
	RefreshToken      OpenIDConnectLifespanToken `koanf:"refresh_token"`
	JWTBearer         OpenIDConnectLifespanToken `koanf:"jwt_bearer"`
}

type OpenIDConnectLifespanToken struct {
	AccessToken   time.Duration `koanf:"access_token"`
	AuthorizeCode time.Duration `koanf:"authorize_code"`
	IDToken       time.Duration `koanf:"id_token"`
	RefreshToken  time.Duration `koanf:"refresh_token"`
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

	AuthorizationPolicy string `koanf:"authorization_policy"`
	Lifespan            string `koanf:"lifespan"`

	ConsentMode                  string         `koanf:"consent_mode"`
	ConsentPreConfiguredDuration *time.Duration `koanf:"pre_configured_consent_duration"`

	ClientCredentialsFlowGrantAllScopesWhenOmitted bool `koanf:"client_credentials_flow_grant_all_scopes_when_omitted"`

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
	Lifespans: OpenIDConnectLifespans{
		OpenIDConnectLifespanToken: OpenIDConnectLifespanToken{
			AccessToken:   time.Hour,
			AuthorizeCode: time.Minute,
			IDToken:       time.Hour,
			RefreshToken:  time.Minute * 90,
		},
	},
	EnforcePKCE: "public_clients_only",
}

var DefaultOpenIDConnectPolicyConfiguration = OpenIDConnectPolicy{
	DefaultPolicy: policyTwoFactor,
}

var defaultOIDCClientConsentPreConfiguredDuration = time.Hour * 24 * 7

// DefaultOpenIDConnectClientConfiguration contains defaults for OIDC Clients.
var DefaultOpenIDConnectClientConfiguration = OpenIDConnectClient{
	AuthorizationPolicy:          policyTwoFactor,
	Scopes:                       []string{"openid", "groups", "profile", "email"},
	ResponseTypes:                []string{"code"},
	ResponseModes:                []string{"form_post"},
	IDTokenSigningAlg:            "RS256",
	UserinfoSigningAlg:           "none",
	ConsentMode:                  "auto",
	ConsentPreConfiguredDuration: &defaultOIDCClientConsentPreConfiguredDuration,
}
