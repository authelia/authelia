package schema

import (
	"crypto/rsa"
	"net"
	"net/url"
	"time"
)

// IdentityProviders represents the Identity Providers configuration for Authelia.
type IdentityProviders struct {
	OIDC *IdentityProvidersOpenIDConnect `koanf:"oidc" yaml:"oidc,omitempty" toml:"oidc,omitempty" json:"oidc,omitempty"`
}

// IdentityProvidersOpenIDConnect represents the configuration for OpenID Connect 1.0.
type IdentityProvidersOpenIDConnect struct {
	HMACSecret  string `koanf:"hmac_secret" yaml:"hmac_secret,omitempty" toml:"hmac_secret,omitempty" json:"hmac_secret,omitempty" jsonschema:"title=HMAC Secret" jsonschema_description:"The HMAC Secret used to sign Access Tokens."`
	JSONWebKeys []JWK  `koanf:"jwks" yaml:"jwks,omitempty" toml:"jwks,omitempty" json:"jwks,omitempty" jsonschema:"title=Issuer JSON Web Keys" jsonschema_description:"The JWK's which are to be used to sign various objects like ID Tokens."`

	EnableClientDebugMessages bool `koanf:"enable_client_debug_messages" yaml:"enable_client_debug_messages" toml:"enable_client_debug_messages" json:"enable_client_debug_messages" jsonschema:"default=false,title=Enable Client Debug Messages" jsonschema_description:"Enables additional debug messages for clients."`
	MinimumParameterEntropy   int  `koanf:"minimum_parameter_entropy" yaml:"minimum_parameter_entropy" toml:"minimum_parameter_entropy" json:"minimum_parameter_entropy" jsonschema:"default=8,minimum=-1,title=Minimum Parameter Entropy" jsonschema_description:"The minimum entropy of the nonce parameter."`

	EnforcePKCE              string `koanf:"enforce_pkce" yaml:"enforce_pkce,omitempty" toml:"enforce_pkce,omitempty" json:"enforce_pkce,omitempty" jsonschema:"default=public_clients_only,enum=public_clients_only,enum=never,enum=always,title=Enforce PKCE" jsonschema_description:"Controls enforcement of the use of Proof Key for Code Exchange on all clients."`
	EnablePKCEPlainChallenge bool   `koanf:"enable_pkce_plain_challenge" yaml:"enable_pkce_plain_challenge" toml:"enable_pkce_plain_challenge" json:"enable_pkce_plain_challenge" jsonschema:"default=false,title=Enable PKCE Plain Challenge" jsonschema_description:"Enables use of the discouraged plain Proof Key for Code Exchange challenges."`

	EnableJWTAccessTokenStatelessIntrospection bool `koanf:"enable_jwt_access_token_stateless_introspection" yaml:"enable_jwt_access_token_stateless_introspection" toml:"enable_jwt_access_token_stateless_introspection" json:"enable_jwt_access_token_stateless_introspection" jsonschema:"title=Enable JWT Access Token Stateless Introspection" jsonschema_description:"Allows the use of stateless introspection of JWT Access Tokens which is not recommended."`

	DiscoverySignedResponseAlg   string `koanf:"discovery_signed_response_alg" yaml:"discovery_signed_response_alg,omitempty" toml:"discovery_signed_response_alg,omitempty" json:"discovery_signed_response_alg,omitempty" jsonschema:"default=none,enum=,enum=none,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Discovery Response Signing Algorithm" jsonschema_description:"The Algorithm this provider uses to sign the Discovery and Metadata Document responses."`
	DiscoverySignedResponseKeyID string `koanf:"discovery_signed_response_key_id" yaml:"discovery_signed_response_key_id,omitempty" toml:"discovery_signed_response_key_id,omitempty" json:"discovery_signed_response_key_id,omitempty" jsonschema:"title=Discovery Response Signing Key ID" jsonschema_description:"The Key ID this provider uses to sign the Discovery and Metadata Document responses (overrides the 'discovery_signed_response_alg')."`

	RequirePushedAuthorizationRequests bool `koanf:"require_pushed_authorization_requests" yaml:"require_pushed_authorization_requests" toml:"require_pushed_authorization_requests" json:"require_pushed_authorization_requests" jsonschema:"title=Require Pushed Authorization Requests" jsonschema_description:"Requires Pushed Authorization Requests for all clients for this Issuer."`

	CORS IdentityProvidersOpenIDConnectCORS `koanf:"cors" yaml:"cors,omitempty" toml:"cors,omitempty" json:"cors,omitempty" jsonschema:"title=CORS" jsonschema_description:"Configuration options for Cross-Origin Request Sharing."`

	Clients []IdentityProvidersOpenIDConnectClient `koanf:"clients" yaml:"clients,omitempty" toml:"clients,omitempty" json:"clients,omitempty" jsonschema:"title=Clients" jsonschema_description:"OpenID Connect 1.0 clients registry."`

	AuthorizationPolicies map[string]IdentityProvidersOpenIDConnectPolicy       `koanf:"authorization_policies" yaml:"authorization_policies,omitempty" toml:"authorization_policies,omitempty" json:"authorization_policies,omitempty" jsonschema:"title=Authorization Policies" jsonschema_description:"Custom client authorization policies."`
	Lifespans             IdentityProvidersOpenIDConnectLifespans               `koanf:"lifespans" yaml:"lifespans,omitempty" toml:"lifespans,omitempty" json:"lifespans,omitempty" jsonschema:"title=Lifespans" jsonschema_description:"Token lifespans configuration."`
	ClaimsPolicies        map[string]IdentityProvidersOpenIDConnectClaimsPolicy `koanf:"claims_policies" yaml:"claims_policies,omitempty" toml:"claims_policies,omitempty" json:"claims_policies,omitempty" jsonschema:"title=Claims Policies" jsonschema_description:"The dictionary of claims policies which can be applied to clients."`
	Scopes                map[string]IdentityProvidersOpenIDConnectScope        `koanf:"scopes" yaml:"scopes,omitempty" toml:"scopes,omitempty" json:"scopes,omitempty" jsonschema:"title=Scopes" jsonschema_description:"List of custom scopes."`

	Discovery IdentityProvidersOpenIDConnectDiscovery `json:"-"` // MetaData value. Not configurable by users.

	IssuerCertificateChain X509CertificateChain `koanf:"issuer_certificate_chain" yaml:"issuer_certificate_chain,omitempty" toml:"issuer_certificate_chain,omitempty" json:"issuer_certificate_chain,omitempty" jsonschema:"title=Issuer Certificate Chain,deprecated" jsonschema_description:"The Issuer Certificate Chain with an RSA Public Key used to sign ID Tokens."`
	IssuerPrivateKey       *rsa.PrivateKey      `koanf:"issuer_private_key" yaml:"issuer_private_key,omitempty" toml:"issuer_private_key,omitempty" json:"issuer_private_key,omitempty" jsonschema:"title=Issuer Private Key,deprecated" jsonschema_description:"The Issuer Private Key with an RSA Private Key used to sign ID Tokens."`
}

type IdentityProvidersOpenIDConnectClaimsPolicy struct {
	IDToken     []string `koanf:"id_token" yaml:"id_token,omitempty" toml:"id_token,omitempty" json:"id_token,omitempty" jsonschema:"title=ID Token" jsonschema_description:"The list of claims to automatically apply to an ID Token in addition to the specified ID Token Claims."`
	AccessToken []string `koanf:"access_token" yaml:"access_token,omitempty" toml:"access_token,omitempty" json:"access_token,omitempty" jsonschema:"title=Access Token" jsonschema_description:"The list of claims to automatically apply to an Access Token in addition to the specified Access Token Claims."`

	IDTokenAudienceMode string `koanf:"id_token_audience_mode" yaml:"id_token_audience_mode,omitempty" toml:"id_token_audience_mode,omitempty" json:"id_token_audience_mode,omitempty" jsonschema:"default=specification,title=ID Token Audience Mode,enum=specification,enum=experimental-merged" jsonschema_description:"Sets the mode for ID Token audience derivation for clients that use this policy."`

	CustomClaims IdentityProvidersOpenIDConnectCustomClaims `koanf:"custom_claims" yaml:"custom_claims,omitempty" toml:"custom_claims,omitempty" json:"custom_claims,omitempty" jsonschema:"title=Custom Claims" jsonschema_description:"The custom claims available in this policy in addition to the Standard Claims."`
}

type IdentityProvidersOpenIDConnectCustomClaims map[string]IdentityProvidersOpenIDConnectCustomClaim

func (c IdentityProvidersOpenIDConnectCustomClaims) GetCustomClaimByName(name string) IdentityProvidersOpenIDConnectCustomClaim {
	for _, properties := range c {
		if properties.Name == name {
			return properties
		}
	}

	return IdentityProvidersOpenIDConnectCustomClaim{}
}

type IdentityProvidersOpenIDConnectCustomClaim struct {
	Name      string `koanf:"name" yaml:"name" toml:"name,omitempty" json:"name,omitempty" jsonschema:"title=Name" jsonschema_description:"The name of claim."`
	Attribute string `koanf:"attribute" yaml:"attribute,omitempty" toml:"attribute,omitempty" json:"attribute,omitempty" jsonschema:"title=Attribute" jsonschema_description:"The attribute that populates this claim."`
}

type IdentityProvidersOpenIDConnectScope struct {
	Claims []string `koanf:"claims" yaml:"claims,omitempty" toml:"claims,omitempty" json:"claims,omitempty" jsonschema:"title=Claims" jsonschema_description:"The list of claims that this scope includes. When this scope is used by a client the clients claim policy must satisfy every claim."`
}

// IdentityProvidersOpenIDConnectPolicy configuration for OpenID Connect 1.0 authorization policies.
type IdentityProvidersOpenIDConnectPolicy struct {
	DefaultPolicy string `koanf:"default_policy" yaml:"default_policy,omitempty" toml:"default_policy,omitempty" json:"default_policy,omitempty" jsonschema:"enum=one_factor,enum=two_factor,enum=deny,title=Default Policy" jsonschema_description:"The default policy action for this policy."`

	Rules []IdentityProvidersOpenIDConnectPolicyRule `koanf:"rules" yaml:"rules,omitempty" toml:"rules,omitempty" json:"rules,omitempty" jsonschema:"title=Rules" jsonschema_description:"The list of rules for this policy."`
}

// IdentityProvidersOpenIDConnectPolicyRule configuration for OpenID Connect 1.0 authorization policies rules.
type IdentityProvidersOpenIDConnectPolicyRule struct {
	Policy   string                    `koanf:"policy" yaml:"policy,omitempty" toml:"policy,omitempty" json:"policy,omitempty" jsonschema:"enum=one_factor,enum=two_factor,enum=deny,title=Policy" jsonschema_description:"The policy to apply to this rule."`
	Subjects AccessControlRuleSubjects `koanf:"subject" yaml:"subject,omitempty" toml:"subject,omitempty" json:"subject,omitempty" jsonschema:"title=Subject" jsonschema_description:"Subject criteria of the Authorization for this rule to be a match."`
	Networks []*net.IPNet              `koanf:"networks" yaml:"networks,omitempty" toml:"networks,omitempty" json:"networks,omitempty" jsonschema:"title=Networks" jsonschema_description:"Networks criteria of the Authorization for this rule to be a match."`
}

// IdentityProvidersOpenIDConnectDiscovery is information discovered during validation reused for the discovery handlers.
type IdentityProvidersOpenIDConnectDiscovery struct {
	Claims                           []string
	Scopes                           []string
	AuthorizationPolicies            []string
	Lifespans                        []string
	DefaultSigKeyIDs                 map[string]string
	DefaultEncKeyIDs                 map[string]string
	DefaultKeyID                     string
	ResponseObjectSigningKeyIDs      []string
	ResponseObjectEncryptionKeyIDs   []string
	ResponseObjectSigningAlgs        []string
	ResponseObjectEncryptionAlgs     []string
	RequestObjectSigningAlgs         []string
	JWTResponseAccessTokens          bool
	BearerAuthorization              bool
	ClientSecretPlainText            bool
	ResponseObjectSymmetricSigEncAlg bool
	RequestObjectSymmetricSigEncAlg  bool
}

type IdentityProvidersOpenIDConnectLifespans struct {
	IdentityProvidersOpenIDConnectLifespanToken `koanf:",squash"`

	DeviceCode              time.Duration `koanf:"device_code" yaml:"device_code,omitempty" toml:"device_code,omitempty" json:"device_code,omitempty" jsonschema:"default=10 minutes,title=Device Code Lifespan" jsonschema_description:"The duration an Device Code is valid for."`
	JWTSecuredAuthorization time.Duration `koanf:"jwt_secured_authorization" yaml:"jwt_secured_authorization,omitempty" toml:"jwt_secured_authorization,omitempty" json:"jwt_secured_authorization,omitempty" jsonschema:"default=5 minutes,title=JARM" jsonschema_description:"Allows tuning the token lifespan for the JWT Secured Authorization Response Modes (JARM)."`

	Custom map[string]IdentityProvidersOpenIDConnectLifespan `koanf:"custom" yaml:"custom,omitempty" toml:"custom,omitempty" json:"custom,omitempty" jsonschema:"title=Custom Lifespans" jsonschema_description:"Allows creating custom lifespans to be used by individual clients."`
}

// IdentityProvidersOpenIDConnectLifespan allows tuning the lifespans for OpenID Connect 1.0 issued tokens.
type IdentityProvidersOpenIDConnectLifespan struct {
	IdentityProvidersOpenIDConnectLifespanToken `koanf:",squash"`

	DeviceCode time.Duration `koanf:"device_code" yaml:"device_code,omitempty" toml:"device_code,omitempty" json:"device_code,omitempty" jsonschema:"default=10 minutes,title=Device Code Lifespan" jsonschema_description:"The duration an Device Code is valid for."`

	Grants IdentityProvidersOpenIDConnectLifespanGrants `koanf:"grants" yaml:"grants,omitempty" toml:"grants,omitempty" json:"grants,omitempty" jsonschema:"title=Grant Types" jsonschema_description:"Allows tuning the token lifespans for individual grant types."`
}

// IdentityProvidersOpenIDConnectLifespanGrants allows tuning the lifespans for each grant type.
type IdentityProvidersOpenIDConnectLifespanGrants struct {
	AuthorizeCode     IdentityProvidersOpenIDConnectLifespanToken `koanf:"authorize_code" yaml:"authorize_code,omitempty" toml:"authorize_code,omitempty" json:"authorize_code,omitempty" jsonschema:"title=Authorize Code Grant" jsonschema_description:"Allows tuning the token lifespans for the authorize code grant."`
	DeviceCode        IdentityProvidersOpenIDConnectLifespanToken `koanf:"device_code" yaml:"device_code,omitempty" toml:"device_code,omitempty" json:"device_code,omitempty" jsonschema:"title=Device Code Grant" jsonschema_description:"Allows tuning the token lifespans for the device code grant."`
	Implicit          IdentityProvidersOpenIDConnectLifespanToken `koanf:"implicit" yaml:"implicit,omitempty" toml:"implicit,omitempty" json:"implicit,omitempty" jsonschema:"title=Implicit Grant" jsonschema_description:"Allows tuning the token lifespans for the implicit flow and grant."`
	ClientCredentials IdentityProvidersOpenIDConnectLifespanToken `koanf:"client_credentials" yaml:"client_credentials,omitempty" toml:"client_credentials,omitempty" json:"client_credentials,omitempty" jsonschema:"title=Client Credentials Grant" jsonschema_description:"Allows tuning the token lifespans for the client credentials grant."`
	RefreshToken      IdentityProvidersOpenIDConnectLifespanToken `koanf:"refresh_token" yaml:"refresh_token,omitempty" toml:"refresh_token,omitempty" json:"refresh_token,omitempty" jsonschema:"title=Refresh Token Grant" jsonschema_description:"Allows tuning the token lifespans for the refresh token grant."`
	JWTBearer         IdentityProvidersOpenIDConnectLifespanToken `koanf:"jwt_bearer" yaml:"jwt_bearer,omitempty" toml:"jwt_bearer,omitempty" json:"jwt_bearer,omitempty" jsonschema:"title=JWT Bearer Grant" jsonschema_description:"Allows tuning the token lifespans for the JWT bearer grant."`
}

// IdentityProvidersOpenIDConnectLifespanToken allows tuning the lifespans for each token type.
type IdentityProvidersOpenIDConnectLifespanToken struct {
	AccessToken   time.Duration `koanf:"access_token" yaml:"access_token,omitempty" toml:"access_token,omitempty" json:"access_token,omitempty" jsonschema:"default=60 minutes,title=Access Token Lifespan" jsonschema_description:"The duration an Access Token is valid for."`
	RefreshToken  time.Duration `koanf:"refresh_token" yaml:"refresh_token,omitempty" toml:"refresh_token,omitempty" json:"refresh_token,omitempty" jsonschema:"default=90 minutes,title=Refresh Token Lifespan" jsonschema_description:"The duration a Refresh Token is valid for."`
	IDToken       time.Duration `koanf:"id_token" yaml:"id_token,omitempty" toml:"id_token,omitempty" json:"id_token,omitempty" jsonschema:"default=60 minutes,title=ID Token Lifespan" jsonschema_description:"The duration an ID Token is valid for."`
	AuthorizeCode time.Duration `koanf:"authorize_code" yaml:"authorize_code,omitempty" toml:"authorize_code,omitempty" json:"authorize_code,omitempty" jsonschema:"default=1 minute,title=Authorize Code Lifespan" jsonschema_description:"The duration an Authorization Code is valid for."`
}

// IdentityProvidersOpenIDConnectCORS represents an OpenID Connect 1.0 CORS config.
type IdentityProvidersOpenIDConnectCORS struct {
	Endpoints      []string   `koanf:"endpoints" yaml:"endpoints,omitempty" toml:"endpoints,omitempty" json:"endpoints,omitempty" jsonschema:"uniqueItems,enum=authorization,enum=device-authorization,enum=pushed-authorization-request,enum=token,enum=introspection,enum=revocation,enum=userinfo,title=Endpoints" jsonschema_description:"List of endpoints to enable CORS handling for."`
	AllowedOrigins []*url.URL `koanf:"allowed_origins" yaml:"allowed_origins,omitempty" toml:"allowed_origins,omitempty" json:"allowed_origins,omitempty" jsonschema:"format=uri,title=Allowed Origins" jsonschema_description:"List of arbitrary allowed origins for CORS requests."`

	AllowedOriginsFromClientRedirectURIs bool `koanf:"allowed_origins_from_client_redirect_uris" yaml:"allowed_origins_from_client_redirect_uris" toml:"allowed_origins_from_client_redirect_uris" json:"allowed_origins_from_client_redirect_uris" jsonschema:"default=false,title=Allowed Origins From Client Redirect URIs" jsonschema_description:"Automatically include the redirect URIs from the registered clients."`
}

// IdentityProvidersOpenIDConnectClient represents a configuration for an OpenID Connect 1.0 client.
type IdentityProvidersOpenIDConnectClient struct {
	ID                  string          `koanf:"client_id" yaml:"client_id" toml:"client_id" json:"client_id" jsonschema:"required,minLength=1,title=Client ID" jsonschema_description:"The Client ID."`
	Name                string          `koanf:"client_name" yaml:"client_name,omitempty" toml:"client_name,omitempty" json:"client_name" jsonschema:"title=Client Name" jsonschema_description:"The Client Name displayed to End-Users."`
	Secret              *PasswordDigest `koanf:"client_secret" yaml:"client_secret,omitempty" toml:"client_secret,omitempty" json:"client_secret" jsonschema:"title=Client Secret" jsonschema_description:"The Client Secret for Client Authentication."`
	SectorIdentifierURI *url.URL        `koanf:"sector_identifier_uri" yaml:"sector_identifier_uri,omitempty" toml:"sector_identifier_uri,omitempty" json:"sector_identifier_uri" jsonschema:"title=Sector Identifier URI" jsonschema_description:"The Client Sector Identifier URI for Privacy Isolation via Pairwise subject types."`
	Public              bool            `koanf:"public" yaml:"public" toml:"public" json:"public" jsonschema:"default=false,title=Public" jsonschema_description:"Enables the Public Client Type."`

	RedirectURIs IdentityProvidersOpenIDConnectClientURIs `koanf:"redirect_uris" yaml:"redirect_uris,omitempty" toml:"redirect_uris,omitempty" json:"redirect_uris" jsonschema:"title=Redirect URIs" jsonschema_description:"List of whitelisted redirect URIs."`
	RequestURIs  IdentityProvidersOpenIDConnectClientURIs `koanf:"request_uris" yaml:"request_uris,omitempty" toml:"request_uris,omitempty" json:"request_uris" jsonschema:"title=Request URIs" jsonschema_description:"List of whitelisted request URIs."`

	Audience      []string `koanf:"audience" yaml:"audience,omitempty" toml:"audience,omitempty" json:"audience" jsonschema:"uniqueItems,title=Audience" jsonschema_description:"List of authorized audiences."`
	Scopes        []string `koanf:"scopes" yaml:"scopes,omitempty" toml:"scopes,omitempty" json:"scopes" jsonschema:"required,enum=openid,enum=offline_access,enum=profile,enum=email,enum=address,enum=phone,enum=groups,enum=authelia.bearer.authz,uniqueItems,title=Scopes" jsonschema_description:"The Scopes this client is allowed request and be granted."`
	GrantTypes    []string `koanf:"grant_types" yaml:"grant_types,omitempty" toml:"grant_types,omitempty" json:"grant_types" jsonschema:"enum=authorization_code,enum=implicit,enum=refresh_token,enum=client_credentials,enum=urn:ietf:params:oauth:grant-type:device_code,uniqueItems,title=Grant Types" jsonschema_description:"The Grant Types this client is allowed to use for the protected endpoints."`
	ResponseTypes []string `koanf:"response_types" yaml:"response_types,omitempty" toml:"response_types,omitempty" json:"response_types" jsonschema:"enum=code,enum=id_token token,enum=id_token,enum=token,enum=code token,enum=code id_token,enum=code id_token token,uniqueItems,title=Response Types" jsonschema_description:"The Response Types the client is authorized to request."`
	ResponseModes []string `koanf:"response_modes" yaml:"response_modes,omitempty" toml:"response_modes,omitempty" json:"response_modes" jsonschema:"enum=form_post,enum=form_post.jwt,enum=query,enum=query.jwt,enum=fragment,enum=fragment.jwt,enum=jwt,uniqueItems,title=Response Modes" jsonschema_description:"The Response Modes this client is authorized request."`

	AuthorizationPolicy string `koanf:"authorization_policy" yaml:"authorization_policy,omitempty" toml:"authorization_policy,omitempty" json:"authorization_policy" jsonschema:"title=Authorization Policy" jsonschema_description:"The Authorization Policy to apply to this client."`
	Lifespan            string `koanf:"lifespan" yaml:"lifespan,omitempty" toml:"lifespan,omitempty" json:"lifespan" jsonschema:"title=Lifespan Name" jsonschema_description:"The name of the custom lifespan to utilize for this client."`
	ClaimsPolicy        string `koanf:"claims_policy" yaml:"claims_policy,omitempty" toml:"claims_policy,omitempty" json:"claims_policy" jsonschema:"title=Claims Policy" jsonschema_description:"The claims policy to apply to this client."`

	RequestedAudienceMode        string         `koanf:"requested_audience_mode" yaml:"requested_audience_mode,omitempty" toml:"requested_audience_mode,omitempty" json:"requested_audience_mode" jsonschema:"enum=explicit,enum=implicit,title=Requested Audience Modes" jsonschema_description:"The Requested Audience Modes used for this client."`
	ConsentMode                  string         `koanf:"consent_mode" yaml:"consent_mode,omitempty" toml:"consent_mode,omitempty" json:"consent_mode" jsonschema:"enum=auto,enum=explicit,enum=implicit,enum=pre-configured,title=Consent Modes" jsonschema_description:"The Consent Modes used for this client."`
	ConsentPreConfiguredDuration *time.Duration `koanf:"pre_configured_consent_duration" yaml:"pre_configured_consent_duration,omitempty" toml:"pre_configured_consent_duration,omitempty" json:"pre_configured_consent_duration" jsonschema:"default=7 days,title=Pre-Configured Consent Duration" jsonschema_description:"The Pre-Configured Consent Duration when using Consent Modes pre-configured for this client."`

	RequirePushedAuthorizationRequests bool `koanf:"require_pushed_authorization_requests" yaml:"require_pushed_authorization_requests,omitempty" toml:"require_pushed_authorization_requests,omitempty" json:"require_pushed_authorization_requests" jsonschema:"default=false,title=Require Pushed Authorization Requests" jsonschema_description:"Requires Pushed Authorization Requests for this client to perform an authorization."`
	RequirePKCE                        bool `koanf:"require_pkce" yaml:"require_pkce,omitempty" toml:"require_pkce,omitempty" json:"require_pkce" jsonschema:"default=false,title=Require PKCE" jsonschema_description:"Requires a Proof Key for this client to perform Code Exchange."`

	PKCEChallengeMethod string `koanf:"pkce_challenge_method" yaml:"pkce_challenge_method,omitempty" toml:"pkce_challenge_method,omitempty" json:"pkce_challenge_method" jsonschema:"enum=,enum=plain,enum=S256,title=PKCE Challenge Method" jsonschema_description:"The PKCE Challenge Method enforced on this client."`

	AuthorizationSignedResponseAlg      string `koanf:"authorization_signed_response_alg" yaml:"authorization_signed_response_alg,omitempty" toml:"authorization_signed_response_alg,omitempty" json:"authorization_signed_response_alg" jsonschema:"default=RS256,enum=,enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Authorization Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client uses to sign the Authorization objects that it generates and responds with. i.e. the JWS 'alg' value."`
	AuthorizationSignedResponseKeyID    string `koanf:"authorization_signed_response_key_id" yaml:"authorization_signed_response_key_id,omitempty" toml:"authorization_signed_response_key_id,omitempty" json:"authorization_signed_response_key_id" jsonschema:"title=Authorization Signing Key ID" jsonschema_description:"The Key ID of a JOSE signing key (JWS) this client uses to sign the Authorization objects that it generates and responds with. This value overrides the 'authorization_signed_response_alg'. i.e. the JWS 'kid' value."`
	AuthorizationEncryptedResponseAlg   string `koanf:"authorization_encrypted_response_alg" yaml:"authorization_encrypted_response_alg,omitempty" toml:"authorization_encrypted_response_alg,omitempty" json:"authorization_encrypted_response_alg" jsonschema:"default=none,enum=,enum=none,enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=Authorization Encryption Algorithm (CEK)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Authorization objects CEK that it generates and responds with. i.e. the JWE 'alg' value."`
	AuthorizationEncryptedResponseEnc   string `koanf:"authorization_encrypted_response_enc" yaml:"authorization_encrypted_response_enc,omitempty" toml:"authorization_encrypted_response_enc,omitempty" json:"authorization_encrypted_response_enc" jsonschema:"default=A128CBC-HS256,enum=,enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=Authorization Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Authorization objects content that it generates and responds with. i.e. the JWE 'enc' value."`
	AuthorizationEncryptedResponseKeyID string `koanf:"authorization_encrypted_response_key_id" yaml:"authorization_encrypted_response_key_id,omitempty" toml:"authorization_encrypted_response_key_id,omitempty" json:"authorization_encrypted_response_key_id" jsonschema:"title=Authorization Signing Key ID" jsonschema_description:"The Key ID of a JOSE encryption key (JWE) this client uses to encrypt the Authorization objects that it generates and responds with. This value overrides the 'authorization_encrypted_response_alg' and 'authorization_encrypted_response_enc'. i.e. the JWE 'kid' value."`

	IDTokenSignedResponseAlg      string `koanf:"id_token_signed_response_alg" yaml:"id_token_signed_response_alg,omitempty" toml:"id_token_signed_response_alg,omitempty" json:"id_token_signed_response_alg" jsonschema:"default=RS256,enum=,enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=ID Token Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client uses to sign the ID Token objects that it generates and responds with. i.e. the JWS 'alg' value."`
	IDTokenSignedResponseKeyID    string `koanf:"id_token_signed_response_key_id" yaml:"id_token_signed_response_key_id,omitempty" toml:"id_token_signed_response_key_id,omitempty" json:"id_token_signed_response_key_id" jsonschema:"title=ID Token Signing Key ID" jsonschema_description:"The Key ID of a JOSE signing key (JWS) this client uses to sign the ID Token objects that it generates and responds with. This value overrides the 'id_token_signed_response_alg'. i.e. the JWS 'kid' value."`
	IDTokenEncryptedResponseAlg   string `koanf:"id_token_encrypted_response_alg" yaml:"id_token_encrypted_response_alg,omitempty" toml:"id_token_encrypted_response_alg,omitempty" json:"id_token_encrypted_response_alg" jsonschema:"default=none,enum=,enum=none,enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=ID Token Encryption Algorithm (CEK)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the ID Token objects CEK that it generates and responds with. i.e. the JWE 'alg' value."`
	IDTokenEncryptedResponseEnc   string `koanf:"id_token_encrypted_response_enc" yaml:"id_token_encrypted_response_enc,omitempty" toml:"id_token_encrypted_response_enc,omitempty" json:"id_token_encrypted_response_enc" jsonschema:"default=A128CBC-HS256,enum=,enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=ID Token Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the ID Token objects content that it generates and responds with. i.e. the JWE 'enc' value."`
	IDTokenEncryptedResponseKeyID string `koanf:"id_token_encrypted_response_key_id" yaml:"id_token_encrypted_response_key_id,omitempty" toml:"id_token_encrypted_response_key_id,omitempty" json:"id_token_encrypted_response_key_id" jsonschema:"title=ID Token Signing Key ID" jsonschema_description:"The Key ID of a JOSE encryption key (JWE) this client uses to encrypt the ID Token objects that it generates and responds with. This value overrides the 'id_token_encrypted_response_alg' and 'id_token_encrypted_response_enc'. i.e. the JWE 'kid' value."`

	AccessTokenSignedResponseAlg      string `koanf:"access_token_signed_response_alg" yaml:"access_token_signed_response_alg,omitempty" toml:"access_token_signed_response_alg,omitempty" json:"access_token_signed_response_alg" jsonschema:"default=none,enum=,enum=none,enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Access Token Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client uses to sign the Access Token objects that it generates and responds with. i.e. the JWS 'alg' value."`
	AccessTokenSignedResponseKeyID    string `koanf:"access_token_signed_response_key_id" yaml:"access_token_signed_response_key_id,omitempty" toml:"access_token_signed_response_key_id,omitempty" json:"access_token_signed_response_key_id" jsonschema:"title=Access Token Signing Key ID" jsonschema_description:"The Key ID of a JOSE signing key (JWS) this client uses to sign the Access Token objects that it generates and responds with. This value overrides the 'access_token_signed_response_alg'. i.e. the JWS 'kid' value."`
	AccessTokenEncryptedResponseAlg   string `koanf:"access_token_encrypted_response_alg" yaml:"access_token_encrypted_response_alg,omitempty" toml:"access_token_encrypted_response_alg,omitempty" json:"access_token_encrypted_response_alg" jsonschema:"default=none,enum=,enum=none,enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=Access Token Encryption Algorithm (CEK)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Access Token objects CEK that it generates and responds with. i.e. the JWE 'alg' value."`
	AccessTokenEncryptedResponseEnc   string `koanf:"access_token_encrypted_response_enc" yaml:"access_token_encrypted_response_enc,omitempty" toml:"access_token_encrypted_response_enc,omitempty" json:"access_token_encrypted_response_enc" jsonschema:"default=A128CBC-HS256,enum=,enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=Access Token Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Access Token objects content that it generates and responds with. i.e. the JWE 'enc' value."`
	AccessTokenEncryptedResponseKeyID string `koanf:"access_token_encrypted_response_key_id" yaml:"access_token_encrypted_response_key_id,omitempty" toml:"access_token_encrypted_response_key_id,omitempty" json:"access_token_encrypted_response_key_id" jsonschema:"title=Access Token Signing Key ID" jsonschema_description:"The Key ID of a JOSE encryption key (JWE) this client uses to encrypt the Access Token objects that it generates and responds with. This value overrides the 'access_token_encrypted_response_alg' and 'access_token_encrypted_response_enc'. i.e. the JWE 'kid' value."`

	UserinfoSignedResponseAlg      string `koanf:"userinfo_signed_response_alg" yaml:"userinfo_signed_response_alg,omitempty" toml:"userinfo_signed_response_alg,omitempty" json:"userinfo_signed_response_alg" jsonschema:"default=none,enum=,enum=none,enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Userinfo Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client uses to sign the Userinfo objects that it generates and responds with. i.e. the JWS 'alg' value."`
	UserinfoSignedResponseKeyID    string `koanf:"userinfo_signed_response_key_id" yaml:"userinfo_signed_response_key_id,omitempty" toml:"userinfo_signed_response_key_id,omitempty" json:"userinfo_signed_response_key_id" jsonschema:"title=Userinfo Signing Key ID" jsonschema_description:"The Key ID of a JOSE signing key (JWS) this client uses to sign the Userinfo objects that it generates and responds with. This value overrides the 'userinfo_signed_response_alg'. i.e. the JWS 'kid' value."`
	UserinfoEncryptedResponseAlg   string `koanf:"userinfo_encrypted_response_alg" yaml:"userinfo_encrypted_response_alg,omitempty" toml:"userinfo_encrypted_response_alg,omitempty" json:"userinfo_encrypted_response_alg" jsonschema:"default=none,enum=,enum=none,enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=Userinfo Encryption Algorithm (CEK)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Userinfo objects CEK that it generates and responds with. i.e. the JWE 'alg' value."`
	UserinfoEncryptedResponseEnc   string `koanf:"userinfo_encrypted_response_enc" yaml:"userinfo_encrypted_response_enc,omitempty" toml:"userinfo_encrypted_response_enc,omitempty" json:"userinfo_encrypted_response_enc" jsonschema:"default=A128CBC-HS256,enum=,enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=Userinfo Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Userinfo objects content that it generates and responds with. i.e. the JWE 'enc' value."`
	UserinfoEncryptedResponseKeyID string `koanf:"userinfo_encrypted_response_key_id" yaml:"userinfo_encrypted_response_key_id,omitempty" toml:"userinfo_encrypted_response_key_id,omitempty" json:"userinfo_encrypted_response_key_id" jsonschema:"title=Userinfo Signing Key ID" jsonschema_description:"The Key ID of a JOSE encryption key (JWE) this client uses to encrypt the Userinfo objects that it generates and responds with. This value overrides the 'userinfo_encrypted_response_alg' and 'userinfo_encrypted_response_enc'. i.e. the JWE 'kid' value."`

	IntrospectionSignedResponseAlg      string `koanf:"introspection_signed_response_alg" yaml:"introspection_signed_response_alg,omitempty" toml:"introspection_signed_response_alg,omitempty" json:"introspection_signed_response_alg" jsonschema:"default=none,enum=,enum=none,enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Introspection Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client uses to sign the Introspection objects that it generates and responds with. i.e. the JWS 'alg' value."`
	IntrospectionSignedResponseKeyID    string `koanf:"introspection_signed_response_key_id" yaml:"introspection_signed_response_key_id,omitempty" toml:"introspection_signed_response_key_id,omitempty" json:"introspection_signed_response_key_id" jsonschema:"title=Introspection Signing Key ID" jsonschema_description:"The Key ID of a JOSE signing key (JWS) this client uses to sign the Introspection objects that it generates and responds with. This value overrides the 'introspection_signed_response_alg'. i.e. the JWS 'kid' value."`
	IntrospectionEncryptedResponseAlg   string `koanf:"introspection_encrypted_response_alg" yaml:"introspection_encrypted_response_alg,omitempty" toml:"introspection_encrypted_response_alg,omitempty" json:"introspection_encrypted_response_alg" jsonschema:"default=none,enum=,enum=none,enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=Introspection Encryption Algorithm (CEK)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Introspection objects CEK that it generates and responds with. i.e. the JWE 'alg' value."`
	IntrospectionEncryptedResponseEnc   string `koanf:"introspection_encrypted_response_enc" yaml:"introspection_encrypted_response_enc,omitempty" toml:"introspection_encrypted_response_enc,omitempty" json:"introspection_encrypted_response_enc" jsonschema:"default=A128CBC-HS256,enum=,enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=Introspection Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Introspection objects content that it generates and responds with. i.e. the JWE 'enc' value."`
	IntrospectionEncryptedResponseKeyID string `koanf:"introspection_encrypted_response_key_id" yaml:"introspection_encrypted_response_key_id,omitempty" toml:"introspection_encrypted_response_key_id,omitempty" json:"introspection_encrypted_response_key_id" jsonschema:"title=Introspection Signing Key ID" jsonschema_description:"The Key ID of a JOSE encryption key (JWE) this client uses to encrypt the Introspection objects that it generates and responds with. This value overrides the 'introspection_encrypted_response_alg' and 'introspection_encrypted_response_enc'. i.e. the JWE 'kid' value."`

	RequestObjectSigningAlg    string `koanf:"request_object_signing_alg" yaml:"request_object_signing_alg,omitempty" toml:"request_object_signing_alg,omitempty" json:"request_object_signing_alg" jsonschema:"enum=,enum=none,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Request Object Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client must use to sign Request Objects that it uses. i.e. the JWS 'alg' value."`
	RequestObjectEncryptionAlg string `koanf:"request_object_encryption_alg" yaml:"request_object_encryption_alg,omitempty" toml:"request_object_encryption_alg,omitempty" json:"request_object_encryption_alg" jsonschema:"default=none,enum=,enum=none,enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=Request Object Encryption Algorithm (CEK)"  jsonschema_description:"The JOSE encryption algorithm (JWE) this client must use to encrypt the Request Object CEK. i.e. the JWE 'alg' value."`
	RequestObjectEncryptionEnc string `koanf:"request_object_encryption_enc" yaml:"request_object_encryption_enc,omitempty" toml:"request_object_encryption_enc,omitempty" json:"request_object_encryption_enc" jsonschema:"default=A128CBC-HS256,enum=,enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=Request Object Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client must use to encrypt the Request Object content. i.e. the JWE 'enc' value."`

	TokenEndpointAuthMethod     string `koanf:"token_endpoint_auth_method" yaml:"token_endpoint_auth_method,omitempty" toml:"token_endpoint_auth_method,omitempty" json:"token_endpoint_auth_method" jsonschema:"default=client_secret_basic,enum=,enum=none,enum=client_secret_post,enum=client_secret_basic,enum=private_key_jwt,enum=client_secret_jwt,title=Token Endpoint Auth Method" jsonschema_description:"The Token Endpoint Auth Method enforced by the provider for this client."`
	TokenEndpointAuthSigningAlg string `koanf:"token_endpoint_auth_signing_alg" yaml:"token_endpoint_auth_signing_alg,omitempty" toml:"token_endpoint_auth_signing_alg,omitempty" json:"token_endpoint_auth_signing_alg" jsonschema:"enum=,enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Token Endpoint Auth Signing Algorithm" jsonschema_description:"The Token Endpoint Auth Signing Algorithm the provider accepts for this client."`

	RevocationEndpointAuthMethod     string `koanf:"revocation_endpoint_auth_method" yaml:"revocation_endpoint_auth_method,omitempty" toml:"revocation_endpoint_auth_method,omitempty" json:"revocation_endpoint_auth_method" jsonschema:"default=client_secret_basic,enum=,enum=none,enum=client_secret_post,enum=client_secret_basic,enum=private_key_jwt,enum=client_secret_jwt,title=Revocation Endpoint Auth Method" jsonschema_description:"The Revocation Endpoint Auth Method enforced by the provider for this client."`
	RevocationEndpointAuthSigningAlg string `koanf:"revocation_endpoint_auth_signing_alg" yaml:"revocation_endpoint_auth_signing_alg,omitempty" toml:"revocation_endpoint_auth_signing_alg,omitempty" json:"revocation_endpoint_auth_signing_alg" jsonschema:"enum=,enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Revocation Endpoint Auth Signing Algorithm" jsonschema_description:"The Revocation Endpoint Auth Signing Algorithm the provider accepts for this client."`

	IntrospectionEndpointAuthMethod     string `koanf:"introspection_endpoint_auth_method" yaml:"introspection_endpoint_auth_method,omitempty" toml:"introspection_endpoint_auth_method,omitempty" json:"introspection_endpoint_auth_method" jsonschema:"default=client_secret_basic,enum=,enum=none,enum=client_secret_post,enum=client_secret_basic,enum=private_key_jwt,enum=client_secret_jwt,title=Introspection Endpoint Auth Method" jsonschema_description:"The Introspection Endpoint Auth Method enforced by the provider for this client."`
	IntrospectionEndpointAuthSigningAlg string `koanf:"introspection_endpoint_auth_signing_alg" yaml:"introspection_endpoint_auth_signing_alg,omitempty" toml:"introspection_endpoint_auth_signing_alg,omitempty" json:"introspection_endpoint_auth_signing_alg" jsonschema:"enum=,enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Introspection Endpoint Auth Signing Algorithm" jsonschema_description:"The Introspection Endpoint Auth Signing Algorithm the provider accepts for this client."`

	PushedAuthorizationRequestEndpointAuthMethod string `koanf:"pushed_authorization_request_endpoint_auth_method" yaml:"pushed_authorization_request_endpoint_auth_method,omitempty" toml:"pushed_authorization_request_endpoint_auth_method,omitempty" json:"pushed_authorization_request_endpoint_auth_method" jsonschema:"default=client_secret_basic,enum=,enum=none,enum=client_secret_post,enum=client_secret_basic,enum=private_key_jwt,enum=client_secret_jwt,title=Pushed Authorization Request Endpoint Auth Method" jsonschema_description:"The Pushed Authorization Request Endpoint Auth Method enforced by the provider for this client."`
	PushedAuthorizationRequestAuthSigningAlg     string `koanf:"pushed_authorization_request_endpoint_auth_signing_alg" yaml:"pushed_authorization_request_endpoint_auth_signing_alg,omitempty" toml:"pushed_authorization_request_endpoint_auth_signing_alg,omitempty" json:"pushed_authorization_request_endpoint_auth_signing_alg" jsonschema:"enum=,enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Pushed Authorization Request Endpoint Auth Signing Algorithm" jsonschema_description:"The Pushed Authorization Request Endpoint Auth Signing Algorithm the provider accepts for this client."`

	AllowMultipleAuthenticationMethods bool `koanf:"allow_multiple_auth_methods" yaml:"allow_multiple_auth_methods,omitempty" toml:"allow_multiple_auth_methods,omitempty" json:"allow_multiple_auth_methods" jsonschema:"title=Allow Multiple Authentication Methods" jsonschema_description:"Permits this registered client to accept misbehaving clients which use a broad authentication approach. This is not standards complaint, use at your own security risk."`

	JSONWebKeysURI *url.URL `koanf:"jwks_uri" yaml:"jwks_uri,omitempty" toml:"jwks_uri,omitempty" json:"jwks_uri" jsonschema:"title=JSON Web Keys URI" jsonschema_description:"URI of the JWKS endpoint which contains the Public Keys used to validate request objects and the 'private_key_jwt' client authentication method for this client."`
	JSONWebKeys    []JWK    `koanf:"jwks" yaml:"jwks,omitempty" toml:"jwks,omitempty" json:"jwks" jsonschema:"title=JSON Web Keys" jsonschema_description:"List of arbitrary Public Keys used to validate request objects and the 'private_key_jwt' client authentication method for this client."`

	Discovery IdentityProvidersOpenIDConnectDiscovery `yaml:"-" json:"-"` // MetaData value. Not configurable by users.
}

// DefaultOpenIDConnectConfiguration contains defaults for OIDC.
var DefaultOpenIDConnectConfiguration = IdentityProvidersOpenIDConnect{
	Lifespans: IdentityProvidersOpenIDConnectLifespans{
		IdentityProvidersOpenIDConnectLifespanToken: IdentityProvidersOpenIDConnectLifespanToken{
			AccessToken:   time.Hour,
			AuthorizeCode: time.Minute,
			IDToken:       time.Hour,
			RefreshToken:  time.Minute * 90,
		},
		DeviceCode: time.Minute * 10,
	},
	EnforcePKCE: "public_clients_only",
}

var DefaultOpenIDConnectPolicyConfiguration = IdentityProvidersOpenIDConnectPolicy{
	DefaultPolicy: policyTwoFactor,
}

var defaultOIDCClientConsentPreConfiguredDuration = time.Hour * 24 * 7

// DefaultOpenIDConnectClientConfiguration contains defaults for OIDC Clients.
var DefaultOpenIDConnectClientConfiguration = IdentityProvidersOpenIDConnectClient{
	AuthorizationPolicy:            policyTwoFactor,
	Scopes:                         []string{"openid", "groups", "profile", "email"},
	ResponseTypes:                  []string{"code"},
	ResponseModes:                  []string{"form_post"},
	AuthorizationSignedResponseAlg: "RS256",
	IDTokenSignedResponseAlg:       "RS256",
	AccessTokenSignedResponseAlg:   "none",
	UserinfoSignedResponseAlg:      "none",
	IntrospectionSignedResponseAlg: "none",
	RequestedAudienceMode:          "explicit",
	ConsentMode:                    "auto",
	ConsentPreConfiguredDuration:   &defaultOIDCClientConsentPreConfiguredDuration,
}
