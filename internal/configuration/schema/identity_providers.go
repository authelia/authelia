package schema

import (
	"crypto/rsa"
	"net"
	"net/url"
	"time"
)

// IdentityProviders represents the Identity Providers configuration for Authelia.
type IdentityProviders struct {
	OIDC *IdentityProvidersOpenIDConnect `koanf:"oidc" json:"oidc"`
}

// IdentityProvidersOpenIDConnect represents the configuration for OpenID Connect 1.0.
type IdentityProvidersOpenIDConnect struct {
	HMACSecret  string `koanf:"hmac_secret" json:"hmac_secret" jsonschema:"title=HMAC Secret" jsonschema_description:"The HMAC Secret used to sign Access Tokens."`
	JSONWebKeys []JWK  `koanf:"jwks" json:"jwks" jsonschema:"title=Issuer JSON Web Keys" jsonschema_description:"The JWK's which are to be used to sign various objects like ID Tokens."`

	EnableClientDebugMessages bool `koanf:"enable_client_debug_messages" json:"enable_client_debug_messages" jsonschema:"default=false,title=Enable Client Debug Messages" jsonschema_description:"Enables additional debug messages for clients."`
	MinimumParameterEntropy   int  `koanf:"minimum_parameter_entropy" json:"minimum_parameter_entropy" jsonschema:"default=8,minimum=-1,title=Minimum Parameter Entropy" jsonschema_description:"The minimum entropy of the nonce parameter."`

	EnforcePKCE              string `koanf:"enforce_pkce" json:"enforce_pkce" jsonschema:"default=public_clients_only,enum=public_clients_only,enum=never,enum=always,title=Enforce PKCE" jsonschema_description:"Controls enforcement of the use of Proof Key for Code Exchange on all clients."`
	EnablePKCEPlainChallenge bool   `koanf:"enable_pkce_plain_challenge" json:"enable_pkce_plain_challenge" jsonschema:"default=false,title=Enable PKCE Plain Challenge" jsonschema_description:"Enables use of the discouraged plain Proof Key for Code Exchange challenges."`

	EnableJWTAccessTokenStatelessIntrospection bool `koanf:"enable_jwt_access_token_stateless_introspection" json:"enable_jwt_access_token_stateless_introspection" jsonschema:"title=Enable JWT Access Token Stateless Introspection" jsonschema_description:"Allows the use of stateless introspection of JWT Access Tokens which is not recommended."`

	DiscoverySignedResponseAlg   string `koanf:"discovery_signed_response_alg" json:"discovery_signed_response_alg" jsonschema:"default=none,enum=none,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Discovery Response Signing Algorithm" jsonschema_description:"The Algorithm this provider uses to sign the Discovery and Metadata Document responses."`
	DiscoverySignedResponseKeyID string `koanf:"discovery_signed_response_key_id" json:"discovery_signed_response_key_id" jsonschema:"title=Discovery Response Signing Key ID" jsonschema_description:"The Key ID this provider uses to sign the Discovery and Metadata Document responses (overrides the 'discovery_signed_response_alg')."`

	RequirePushedAuthorizationRequests bool `koanf:"require_pushed_authorization_requests" json:"require_pushed_authorization_requests" jsonschema:"title=Require Pushed Authorization Requests" jsonschema_description:"Requires Pushed Authorization Requests for all clients for this Issuer."`

	CORS IdentityProvidersOpenIDConnectCORS `koanf:"cors" json:"cors" jsonschema:"title=CORS" jsonschema_description:"Configuration options for Cross-Origin Request Sharing."`

	Clients []IdentityProvidersOpenIDConnectClient `koanf:"clients" json:"clients" jsonschema:"title=Clients" jsonschema_description:"OpenID Connect 1.0 clients registry."`

	AuthorizationPolicies map[string]IdentityProvidersOpenIDConnectPolicy       `koanf:"authorization_policies" json:"authorization_policies" jsonschema:"title=Authorization Policies" jsonschema_description:"Custom client authorization policies."`
	Lifespans             IdentityProvidersOpenIDConnectLifespans               `koanf:"lifespans" json:"lifespans" jsonschema:"title=Lifespans" jsonschema_description:"Token lifespans configuration."`
	ClaimsPolicies        map[string]IdentityProvidersOpenIDConnectClaimsPolicy `koanf:"claims_policies" json:"claims_policies" jsonschema:"title=Claims Policies" jsonschema_description:"The dictionary of claims policies which can be applied to clients."`
	Scopes                map[string]IdentityProvidersOpenIDConnectScope        `koanf:"scopes" json:"scopes" jsonschema:"title=Scopes" jsonschema_description:"List of custom scopes."`

	Discovery IdentityProvidersOpenIDConnectDiscovery `json:"-"` // MetaData value. Not configurable by users.

	IssuerCertificateChain X509CertificateChain `koanf:"issuer_certificate_chain" json:"issuer_certificate_chain" jsonschema:"title=Issuer Certificate Chain,deprecated" jsonschema_description:"The Issuer Certificate Chain with an RSA Public Key used to sign ID Tokens."`
	IssuerPrivateKey       *rsa.PrivateKey      `koanf:"issuer_private_key" json:"issuer_private_key" jsonschema:"title=Issuer Private Key,deprecated" jsonschema_description:"The Issuer Private Key with an RSA Private Key used to sign ID Tokens."`
}

type IdentityProvidersOpenIDConnectClaimsPolicy struct {
	IDToken     []string `koanf:"id_token" json:"id_token" jsonschema:"title=ID Token" jsonschema_description:"The list of claims to automatically apply to an ID Token in addition to the specified ID Token Claims."`
	AccessToken []string `koanf:"access_token" json:"access_token" jsonschema:"title=Access Token" jsonschema_description:"The list of claims to automatically apply to an Access Token in addition to the specified Access Token Claims."`

	CustomClaims map[string]IdentityProvidersOpenIDConnectCustomClaim `koanf:"custom_claims" json:"custom_claims" jsonschema:"title=Custom Claims" jsonschema_description:"The custom claims available in this policy in addition to the Standard Claims."`
}

type IdentityProvidersOpenIDConnectCustomClaim struct {
	Attribute string `koanf:"attribute" json:"attribute" jsonschema:"title=Attribute" jsonschema_description:"The attribute that populates this claim."`
}

type IdentityProvidersOpenIDConnectScope struct {
	Claims []string `koanf:"claims" json:"claims" jsonschema:"title=Claims" jsonschema_description:"The list of claims that this scope includes. When this scope is used by a client the clients claim policy must satisfy every claim."`
}

// IdentityProvidersOpenIDConnectPolicy configuration for OpenID Connect 1.0 authorization policies.
type IdentityProvidersOpenIDConnectPolicy struct {
	DefaultPolicy string `koanf:"default_policy" json:"default_policy" jsonschema:"enum=one_factor,enum=two_factor,enum=deny,title=Default Policy" jsonschema_description:"The default policy action for this policy."`

	Rules []IdentityProvidersOpenIDConnectPolicyRule `koanf:"rules" json:"rules" jsonschema:"title=Rules" jsonschema_description:"The list of rules for this policy."`
}

// IdentityProvidersOpenIDConnectPolicyRule configuration for OpenID Connect 1.0 authorization policies rules.
type IdentityProvidersOpenIDConnectPolicyRule struct {
	Policy   string                    `koanf:"policy" json:"policy" jsonschema:"enum=one_factor,enum=two_factor,enum=deny,title=Policy" jsonschema_description:"The policy to apply to this rule."`
	Subjects AccessControlRuleSubjects `koanf:"subject" json:"subject" jsonschema:"title=Subject" jsonschema_description:"Subject criteria of the Authorization for this rule to be a match."`
	Networks []*net.IPNet              `koanf:"networks" json:"networks" jsonschema:"title=Networks" jsonschema_description:"Networks criteria of the Authorization for this rule to be a match."`
}

// IdentityProvidersOpenIDConnectDiscovery is information discovered during validation reused for the discovery handlers.
type IdentityProvidersOpenIDConnectDiscovery struct {
	Claims                       []string
	Scopes                       []string
	AuthorizationPolicies        []string
	Lifespans                    []string
	DefaultSigKeyIDs             map[string]string
	DefaultEncKeyIDs             map[string]string
	DefaultKeyID                 string
	ResponseObjectSigningKeyIDs  []string
	ResponseObjectSigningAlgs    []string
	ResponseObjectEncryptionAlgs []string
	RequestObjectSigningAlgs     []string
	JWTResponseAccessTokens      bool
	BearerAuthorization          bool
	ClientSecretPlainText        bool
	ClientAuthMethodSymmetric    bool
}

type IdentityProvidersOpenIDConnectLifespans struct {
	IdentityProvidersOpenIDConnectLifespanToken `koanf:",squash"`
	JWTSecuredAuthorization                     time.Duration `koanf:"jwt_secured_authorization" json:"jwt_secured_authorization" jsonschema:"default=5 minutes,title=JARM" jsonschema_description:"Allows tuning the token lifespan for the JWT Secured Authorization Response Mode (JARM)."`

	Custom map[string]IdentityProvidersOpenIDConnectLifespan `koanf:"custom" json:"custom" jsonschema:"title=Custom Lifespans" jsonschema_description:"Allows creating custom lifespans to be used by individual clients."`
}

// IdentityProvidersOpenIDConnectLifespan allows tuning the lifespans for OpenID Connect 1.0 issued tokens.
type IdentityProvidersOpenIDConnectLifespan struct {
	IdentityProvidersOpenIDConnectLifespanToken `koanf:",squash"`

	Grants IdentityProvidersOpenIDConnectLifespanGrants `koanf:"grants" json:"grants" jsonschema:"title=Grant Types" jsonschema_description:"Allows tuning the token lifespans for individual grant types."`
}

// IdentityProvidersOpenIDConnectLifespanGrants allows tuning the lifespans for each grant type.
type IdentityProvidersOpenIDConnectLifespanGrants struct {
	AuthorizeCode     IdentityProvidersOpenIDConnectLifespanToken `koanf:"authorize_code" json:"authorize_code" jsonschema:"title=Authorize Code Grant" jsonschema_description:"Allows tuning the token lifespans for the authorize code grant."`
	Implicit          IdentityProvidersOpenIDConnectLifespanToken `koanf:"implicit" json:"implicit" jsonschema:"title=Implicit Grant" jsonschema_description:"Allows tuning the token lifespans for the implicit flow and grant."`
	ClientCredentials IdentityProvidersOpenIDConnectLifespanToken `koanf:"client_credentials" json:"client_credentials" jsonschema:"title=Client Credentials Grant" jsonschema_description:"Allows tuning the token lifespans for the client credentials grant."`
	RefreshToken      IdentityProvidersOpenIDConnectLifespanToken `koanf:"refresh_token" json:"refresh_token" jsonschema:"title=Refresh Token Grant" jsonschema_description:"Allows tuning the token lifespans for the refresh token grant."`
	JWTBearer         IdentityProvidersOpenIDConnectLifespanToken `koanf:"jwt_bearer" json:"jwt_bearer" jsonschema:"title=JWT Bearer Grant" jsonschema_description:"Allows tuning the token lifespans for the JWT bearer grant."`
}

// IdentityProvidersOpenIDConnectLifespanToken allows tuning the lifespans for each token type.
type IdentityProvidersOpenIDConnectLifespanToken struct {
	AccessToken   time.Duration `koanf:"access_token" json:"access_token" jsonschema:"default=60 minutes,title=Access Token Lifespan" jsonschema_description:"The duration an Access Token is valid for."`
	AuthorizeCode time.Duration `koanf:"authorize_code" json:"authorize_code" jsonschema:"default=1 minute,title=Authorize Code Lifespan" jsonschema_description:"The duration an Authorization Code is valid for."`
	IDToken       time.Duration `koanf:"id_token" json:"id_token" jsonschema:"default=60 minutes,title=ID Token Lifespan" jsonschema_description:"The duration an ID Token is valid for."`
	RefreshToken  time.Duration `koanf:"refresh_token" json:"refresh_token" jsonschema:"default=90 minutes,title=Refresh Token Lifespan" jsonschema_description:"The duration a Refresh Token is valid for."`
}

// IdentityProvidersOpenIDConnectCORS represents an OpenID Connect 1.0 CORS config.
type IdentityProvidersOpenIDConnectCORS struct {
	Endpoints      []string   `koanf:"endpoints" json:"endpoints" jsonschema:"uniqueItems,enum=authorization,enum=device-authorization,enum=pushed-authorization-request,enum=token,enum=introspection,enum=revocation,enum=userinfo,title=Endpoints" jsonschema_description:"List of endpoints to enable CORS handling for."`
	AllowedOrigins []*url.URL `koanf:"allowed_origins" json:"allowed_origins" jsonschema:"format=uri,title=Allowed Origins" jsonschema_description:"List of arbitrary allowed origins for CORS requests."`

	AllowedOriginsFromClientRedirectURIs bool `koanf:"allowed_origins_from_client_redirect_uris" json:"allowed_origins_from_client_redirect_uris" jsonschema:"default=false,title=Allowed Origins From Client Redirect URIs" jsonschema_description:"Automatically include the redirect URIs from the registered clients."`
}

// IdentityProvidersOpenIDConnectClient represents a configuration for an OpenID Connect 1.0 client.
type IdentityProvidersOpenIDConnectClient struct {
	ID                  string          `koanf:"client_id" json:"client_id" jsonschema:"required,minLength=1,title=Client ID" jsonschema_description:"The Client ID."`
	Name                string          `koanf:"client_name" json:"client_name" jsonschema:"title=Client Name" jsonschema_description:"The Client Name displayed to End-Users."`
	Secret              *PasswordDigest `koanf:"client_secret" json:"client_secret" jsonschema:"title=Client Secret" jsonschema_description:"The Client Secret for Client Authentication."`
	SectorIdentifierURI *url.URL        `koanf:"sector_identifier_uri" json:"sector_identifier_uri" jsonschema:"title=Sector Identifier URI" jsonschema_description:"The Client Sector Identifier URI for Privacy Isolation via Pairwise subject types."`
	Public              bool            `koanf:"public" json:"public" jsonschema:"default=false,title=Public" jsonschema_description:"Enables the Public Client Type."`

	RedirectURIs IdentityProvidersOpenIDConnectClientURIs `koanf:"redirect_uris" json:"redirect_uris" jsonschema:"title=Redirect URIs" jsonschema_description:"List of whitelisted redirect URIs."`
	RequestURIs  IdentityProvidersOpenIDConnectClientURIs `koanf:"request_uris" json:"request_uris" jsonschema:"title=Request URIs" jsonschema_description:"List of whitelisted request URIs."`

	Audience      []string `koanf:"audience" json:"audience" jsonschema:"uniqueItems,title=Audience" jsonschema_description:"List of authorized audiences."`
	Scopes        []string `koanf:"scopes" json:"scopes" jsonschema:"required,enum=openid,enum=offline_access,enum=groups,enum=email,enum=profile,enum=authelia.bearer.authz,uniqueItems,title=Scopes" jsonschema_description:"The Scopes this client is allowed request and be granted."`
	GrantTypes    []string `koanf:"grant_types" json:"grant_types" jsonschema:"enum=authorization_code,enum=implicit,enum=refresh_token,enum=client_credentials,enum=urn:ietf:params:oauth:grant-type:device_code,uniqueItems,title=Grant Types" jsonschema_description:"The Grant Types this client is allowed to use for the protected endpoints."`
	ResponseTypes []string `koanf:"response_types" json:"response_types" jsonschema:"enum=code,enum=id_token token,enum=id_token,enum=token,enum=code token,enum=code id_token,enum=code id_token token,uniqueItems,title=Response Types" jsonschema_description:"The Response Types the client is authorized to request."`
	ResponseModes []string `koanf:"response_modes" json:"response_modes" jsonschema:"enum=form_post,enum=form_post.jwt,enum=query,enum=query.jwt,enum=fragment,enum=fragment.jwt,enum=jwt,uniqueItems,title=Response Modes" jsonschema_description:"The Response Modes this client is authorized request."`

	AuthorizationPolicy string `koanf:"authorization_policy" json:"authorization_policy" jsonschema:"title=Authorization Policy" jsonschema_description:"The Authorization Policy to apply to this client."`
	Lifespan            string `koanf:"lifespan" json:"lifespan" jsonschema:"title=Lifespan Name" jsonschema_description:"The name of the custom lifespan to utilize for this client."`
	ClaimsPolicy        string `koanf:"claims_policy" json:"claims_policy" jsonschema:"title=Claims Policy" jsonschema_description:"The claims policy to apply to this client."`

	RequestedAudienceMode        string         `koanf:"requested_audience_mode" json:"requested_audience_mode" jsonschema:"enum=explicit,enum=implicit,title=Requested Audience Mode" jsonschema_description:"The Requested Audience Mode used for this client."`
	ConsentMode                  string         `koanf:"consent_mode" json:"consent_mode" jsonschema:"enum=auto,enum=explicit,enum=implicit,enum=pre-configured,title=Consent Mode" jsonschema_description:"The Consent Mode used for this client."`
	ConsentPreConfiguredDuration *time.Duration `koanf:"pre_configured_consent_duration" json:"pre_configured_consent_duration" jsonschema:"default=7 days,title=Pre-Configured Consent Duration" jsonschema_description:"The Pre-Configured Consent Duration when using Consent Mode pre-configured for this client."`

	RequirePushedAuthorizationRequests bool `koanf:"require_pushed_authorization_requests" json:"require_pushed_authorization_requests" jsonschema:"default=false,title=Require Pushed Authorization Requests" jsonschema_description:"Requires Pushed Authorization Requests for this client to perform an authorization."`
	RequirePKCE                        bool `koanf:"require_pkce" json:"require_pkce" jsonschema:"default=false,title=Require PKCE" jsonschema_description:"Requires a Proof Key for this client to perform Code Exchange."`

	PKCEChallengeMethod string `koanf:"pkce_challenge_method" json:"pkce_challenge_method" jsonschema:"enum=plain,enum=S256,title=PKCE Challenge Method" jsonschema_description:"The PKCE Challenge Method enforced on this client."`

	AuthorizationSignedResponseAlg      string `koanf:"authorization_signed_response_alg" json:"authorization_signed_response_alg" jsonschema:"default=none,enum=none,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Authorization Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client uses to sign the Authorization objects that it generates and responds with. i.e. the JWS 'alg' value."`
	AuthorizationSignedResponseKeyID    string `koanf:"authorization_signed_response_key_id" json:"authorization_signed_response_key_id" jsonschema:"title=Authorization Signing Key ID" jsonschema_description:"The Key ID of a JOSE signing key (JWS) this client uses to sign the Authorization objects that it generates and responds with. This value overrides the 'authorization_signed_response_alg'. i.e. the JWS 'kid' value."`
	AuthorizationEncryptedResponseAlg   string `koanf:"authorization_encrypted_response_alg" json:"authorization_encrypted_response_alg" jsonschema:"enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=Authorization Encryption Algorithm (CEK)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Authorization objects CEK that it generates and responds with. i.e. the JWE 'alg' value."`
	AuthorizationEncryptedResponseEnc   string `koanf:"authorization_encrypted_response_enc" json:"authorization_encrypted_response_enc" jsonschema:"enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=Authorization Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Authorization objects content that it generates and responds with. i.e. the JWE 'enc' value."`
	AuthorizationEncryptedResponseKeyID string `koanf:"authorization_encrypted_response_key_id" json:"authorization_encrypted_response_key_id" jsonschema:"title=Authorization Signing Key ID" jsonschema_description:"The Key ID of a JOSE encryption key (JWE) this client uses to encrypt the Authorization objects that it generates and responds with. This value overrides the 'authorization_encrypted_response_alg' and '%!s(MISSING)_encrypted_response_enc'. i.e. the JWE 'kid' value."`

	IDTokenSignedResponseAlg      string `koanf:"id_token_signed_response_alg" json:"id_token_signed_response_alg" jsonschema:"default=RS256,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=ID Token Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client uses to sign the ID Token objects that it generates and responds with. i.e. the JWS 'alg' value."`
	IDTokenSignedResponseKeyID    string `koanf:"id_token_signed_response_key_id" json:"id_token_signed_response_key_id" jsonschema:"title=ID Token Signing Key ID" jsonschema_description:"The Key ID of a JOSE signing key (JWS) this client uses to sign the ID Token objects that it generates and responds with. This value overrides the 'id_token_signed_response_alg'. i.e. the JWS 'kid' value."`
	IDTokenEncryptedResponseAlg   string `koanf:"id_token_encrypted_response_alg" json:"id_token_encrypted_response_alg" jsonschema:"enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=ID Token Encryption Algorithm (CEK)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the ID Token objects CEK that it generates and responds with. i.e. the JWE 'alg' value."`
	IDTokenEncryptedResponseEnc   string `koanf:"id_token_encrypted_response_enc" json:"id_token_encrypted_response_enc" jsonschema:"enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=ID Token Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the ID Token objects content that it generates and responds with. i.e. the JWE 'enc' value."`
	IDTokenEncryptedResponseKeyID string `koanf:"id_token_encrypted_response_key_id" json:"id_token_encrypted_response_key_id" jsonschema:"title=ID Token Signing Key ID" jsonschema_description:"The Key ID of a JOSE encryption key (JWE) this client uses to encrypt the ID Token objects that it generates and responds with. This value overrides the 'id_token_encrypted_response_alg' and '%!s(MISSING)_encrypted_response_enc'. i.e. the JWE 'kid' value."`

	AccessTokenSignedResponseAlg      string `koanf:"access_token_signed_response_alg" json:"access_token_signed_response_alg" jsonschema:"default=none,enum=none,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Access Token Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client uses to sign the Access Token objects that it generates and responds with. i.e. the JWS 'alg' value."`
	AccessTokenSignedResponseKeyID    string `koanf:"access_token_signed_response_key_id" json:"access_token_signed_response_key_id" jsonschema:"title=Access Token Signing Key ID" jsonschema_description:"The Key ID of a JOSE signing key (JWS) this client uses to sign the Access Token objects that it generates and responds with. This value overrides the 'access_token_signed_response_alg'. i.e. the JWS 'kid' value."`
	AccessTokenEncryptedResponseAlg   string `koanf:"access_token_encrypted_response_alg" json:"access_token_encrypted_response_alg" jsonschema:"enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=Access Token Encryption Algorithm (CEK)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Access Token objects CEK that it generates and responds with. i.e. the JWE 'alg' value."`
	AccessTokenEncryptedResponseEnc   string `koanf:"access_token_encrypted_response_enc" json:"access_token_encrypted_response_enc" jsonschema:"enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=Access Token Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Access Token objects content that it generates and responds with. i.e. the JWE 'enc' value."`
	AccessTokenEncryptedResponseKeyID string `koanf:"access_token_encrypted_response_key_id" json:"access_token_encrypted_response_key_id" jsonschema:"title=Access Token Signing Key ID" jsonschema_description:"The Key ID of a JOSE encryption key (JWE) this client uses to encrypt the Access Token objects that it generates and responds with. This value overrides the 'access_token_encrypted_response_alg' and '%!s(MISSING)_encrypted_response_enc'. i.e. the JWE 'kid' value."`

	UserinfoSignedResponseAlg      string `koanf:"userinfo_signed_response_alg" json:"userinfo_signed_response_alg" jsonschema:"default=none,enum=none,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Userinfo Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client uses to sign the Userinfo objects that it generates and responds with. i.e. the JWS 'alg' value."`
	UserinfoSignedResponseKeyID    string `koanf:"userinfo_signed_response_key_id" json:"userinfo_signed_response_key_id" jsonschema:"title=Userinfo Signing Key ID" jsonschema_description:"The Key ID of a JOSE signing key (JWS) this client uses to sign the Userinfo objects that it generates and responds with. This value overrides the 'userinfo_signed_response_alg'. i.e. the JWS 'kid' value."`
	UserinfoEncryptedResponseAlg   string `koanf:"userinfo_encrypted_response_alg" json:"userinfo_encrypted_response_alg" jsonschema:"enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=Userinfo Encryption Algorithm (CEK)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Userinfo objects CEK that it generates and responds with. i.e. the JWE 'alg' value."`
	UserinfoEncryptedResponseEnc   string `koanf:"userinfo_encrypted_response_enc" json:"userinfo_encrypted_response_enc" jsonschema:"enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=Userinfo Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Userinfo objects content that it generates and responds with. i.e. the JWE 'enc' value."`
	UserinfoEncryptedResponseKeyID string `koanf:"userinfo_encrypted_response_key_id" json:"userinfo_encrypted_response_key_id" jsonschema:"title=Userinfo Signing Key ID" jsonschema_description:"The Key ID of a JOSE encryption key (JWE) this client uses to encrypt the Userinfo objects that it generates and responds with. This value overrides the 'userinfo_encrypted_response_alg' and '%!s(MISSING)_encrypted_response_enc'. i.e. the JWE 'kid' value."`

	IntrospectionSignedResponseAlg      string `koanf:"introspection_signed_response_alg" json:"introspection_signed_response_alg" jsonschema:"default=none,enum=none,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Introspection Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client uses to sign the Introspection objects that it generates and responds with. i.e. the JWS 'alg' value."`
	IntrospectionSignedResponseKeyID    string `koanf:"introspection_signed_response_key_id" json:"introspection_signed_response_key_id" jsonschema:"title=Introspection Signing Key ID" jsonschema_description:"The Key ID of a JOSE signing key (JWS) this client uses to sign the Introspection objects that it generates and responds with. This value overrides the 'introspection_signed_response_alg'. i.e. the JWS 'kid' value."`
	IntrospectionEncryptedResponseAlg   string `koanf:"introspection_encrypted_response_alg" json:"introspection_encrypted_response_alg" jsonschema:"enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=Introspection Encryption Algorithm (CEK)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Introspection objects CEK that it generates and responds with. i.e. the JWE 'alg' value."`
	IntrospectionEncryptedResponseEnc   string `koanf:"introspection_encrypted_response_enc" json:"introspection_encrypted_response_enc" jsonschema:"enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=Introspection Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client uses to encrypt the Introspection objects content that it generates and responds with. i.e. the JWE 'enc' value."`
	IntrospectionEncryptedResponseKeyID string `koanf:"introspection_encrypted_response_key_id" json:"introspection_encrypted_response_key_id" jsonschema:"title=Introspection Signing Key ID" jsonschema_description:"The Key ID of a JOSE encryption key (JWE) this client uses to encrypt the Introspection objects that it generates and responds with. This value overrides the 'introspection_encrypted_response_alg' and '%!s(MISSING)_encrypted_response_enc'. i.e. the JWE 'kid' value."`

	RequestObjectSigningAlg    string `koanf:"request_object_signing_alg" json:"request_object_signing_alg" jsonschema:"enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Request Object Signing Algorithm" jsonschema_description:"The JOSE signing algorithm (JWS) this client must use to sign Request Objects that it uses. i.e. the JWS 'alg' value."`
	RequestObjectEncryptionAlg string `koanf:"request_object_encryption_alg" json:"request_object_encryption_alg" jsonschema:"enum=RSA1_5,enum=RSA-OAEP,enum=RSA-OAEP-256,enum=A128KW,enum=A192KW,enum=A256KW,enum=dir,enum=ECDH-ES,enum=ECDH-ES+A128KW,enum=ECDH-ES+A192KW,enum=ECDH-ES+A256KW,enum=A128GCMKW,enum=A192GCMKW,enum=A256GCMKW,enum=PBES2-HS256+A128KW,enum=PBES2-HS384+A192KW,enum=PBES2-HS512+A256KW,title=Request Object Encryption Algorithm (CEK)"  jsonschema_description:"The JOSE encryption algorithm (JWE) this client must use to encrypt the Request Object CEK. i.e. the JWE 'alg' value."`
	RequestObjectEncryptionEnc string `koanf:"request_object_encryption_enc" json:"request_object_encryption_enc" jsonschema:"enum=A128CBC-HS256,enum=A192CBC-HS384,enum=A256CBC-HS512,enum=A128GCM,enum=A192GCM,enum=A256GCM,title=Request Object Encryption Algorithm (Content)" jsonschema_description:"The JOSE encryption algorithm (JWE) this client must use to encrypt the Request Object content. i.e. the JWE 'enc' value."`

	TokenEndpointAuthMethod     string `koanf:"token_endpoint_auth_method" json:"token_endpoint_auth_method" jsonschema:"enum=none,enum=client_secret_post,enum=client_secret_basic,enum=private_key_jwt,enum=client_secret_jwt,title=Token Endpoint Auth Method" jsonschema_description:"The Token Endpoint Auth Method enforced by the provider for this client."`
	TokenEndpointAuthSigningAlg string `koanf:"token_endpoint_auth_signing_alg" json:"token_endpoint_auth_signing_alg" jsonschema:"enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Token Endpoint Auth Signing Algorithm" jsonschema_description:"The Token Endpoint Auth Signing Algorithm the provider accepts for this client."`

	RevocationEndpointAuthMethod     string `koanf:"revocation_endpoint_auth_method" json:"revocation_endpoint_auth_method" jsonschema:"enum=none,enum=client_secret_post,enum=client_secret_basic,enum=private_key_jwt,enum=client_secret_jwt,title=Revocation Endpoint Auth Method" jsonschema_description:"The Revocation Endpoint Auth Method enforced by the provider for this client."`
	RevocationEndpointAuthSigningAlg string `koanf:"revocation_endpoint_auth_signing_alg" json:"revocation_endpoint_auth_signing_alg" jsonschema:"enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Revocation Endpoint Auth Signing Algorithm" jsonschema_description:"The Revocation Endpoint Auth Signing Algorithm the provider accepts for this client."`

	IntrospectionEndpointAuthMethod     string `koanf:"introspection_endpoint_auth_method" json:"introspection_endpoint_auth_method" jsonschema:"enum=none,enum=client_secret_post,enum=client_secret_basic,enum=private_key_jwt,enum=client_secret_jwt,title=Introspection Endpoint Auth Method" jsonschema_description:"The Introspection Endpoint Auth Method enforced by the provider for this client."`
	IntrospectionEndpointAuthSigningAlg string `koanf:"introspection_endpoint_auth_signing_alg" json:"introspection_endpoint_auth_signing_alg" jsonschema:"enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Introspection Endpoint Auth Signing Algorithm" jsonschema_description:"The Introspection Endpoint Auth Signing Algorithm the provider accepts for this client."`

	PushedAuthorizationRequestEndpointAuthMethod string `koanf:"pushed_authorization_request_endpoint_auth_method" json:"pushed_authorization_request_endpoint_auth_method" jsonschema:"enum=none,enum=client_secret_post,enum=client_secret_basic,enum=private_key_jwt,enum=client_secret_jwt,title=Pushed Authorization Request Endpoint Auth Method" jsonschema_description:"The Pushed Authorization Request Endpoint Auth Method enforced by the provider for this client."`
	PushedAuthorizationRequestAuthSigningAlg     string `koanf:"pushed_authorization_request_endpoint_auth_signing_alg" json:"pushed_authorization_request_endpoint_auth_signing_alg" jsonschema:"enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Pushed Authorization Request Endpoint Auth Signing Algorithm" jsonschema_description:"The Pushed Authorization Request Endpoint Auth Signing Algorithm the provider accepts for this client."`

	AllowMultipleAuthenticationMethods bool `koanf:"allow_multiple_auth_methods" json:"allow_multiple_auth_methods" jsonschema:"title=Allow Multiple Authentication Methods" jsonschema_description:"Permits this registered client to accept misbehaving clients which use a broad authentication approach. This is not standards complaint, use at your own security risk."`

	JSONWebKeysURI *url.URL `koanf:"jwks_uri" json:"jwks_uri" jsonschema:"title=JSON Web Keys URI" jsonschema_description:"URI of the JWKS endpoint which contains the Public Keys used to validate request objects and the 'private_key_jwt' client authentication method for this client."`
	JSONWebKeys    []JWK    `koanf:"jwks" json:"jwks" jsonschema:"title=JSON Web Keys" jsonschema_description:"List of arbitrary Public Keys used to validate request objects and the 'private_key_jwt' client authentication method for this client."`

	Discovery IdentityProvidersOpenIDConnectDiscovery `json:"-"` // MetaData value. Not configurable by users.
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
	AuthorizationSignedResponseAlg: "none",
	IDTokenSignedResponseAlg:       "RS256",
	AccessTokenSignedResponseAlg:   "none",
	UserinfoSignedResponseAlg:      "none",
	IntrospectionSignedResponseAlg: "none",
	RequestedAudienceMode:          "explicit",
	ConsentMode:                    "auto",
	ConsentPreConfiguredDuration:   &defaultOIDCClientConsentPreConfiguredDuration,
}
