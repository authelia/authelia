package oidc

import (
	"context"
	"net/http"
	"net/url"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	fjwt "authelia.com/provider/oauth2/token/jwt"
	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/expression"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/storage"
)

// OpenIDConnectProvider for OpenID Connect.
type OpenIDConnectProvider struct {
	*Store
	*Config

	Issuer *Issuer

	discovery OpenIDConnectWellKnownConfiguration

	oauthelia2.Provider
}

// Store is Authelia's internal representation of the oauthelia2.Storage interface. It maps the following
// interfaces to the storage.Provider interface:
// oauthelia2.Storage, oauthelia2.ClientManager, storage.Transactional, oauth2.AuthorizeCodeStorage, oauth2.AccessTokenStorage,
// oauth2.RefreshTokenStorage, oauth2.TokenRevocationStorage, pkce.PKCERequestStorage,
// openid.OpenIDConnectRequestStorage, and partially implements rfc7523.RFC7523KeyStorage.
type Store struct {
	ClientStore

	provider storage.Provider
}

// ClientStore is an abstraction used for the Store struct which stores clients.
type ClientStore interface {
	// GetRegisteredClient returns a Client matching the provided id.
	GetRegisteredClient(ctx context.Context, id string) (client Client, err error)
}

// MemoryClientStore is an implementation of the ClientStore which just stores the clients in memory.
type MemoryClientStore struct {
	clients map[string]Client
}

// RegisteredClient represents a registered client.
type RegisteredClient struct {
	ID                   string
	Name                 string
	ClientSecret         *ClientSecretDigest
	RotatedClientSecrets []*ClientSecretDigest
	SectorIdentifierURI  *url.URL
	Public               bool

	RequirePushedAuthorizationRequests bool

	RequirePKCE                bool
	RequirePKCEChallengeMethod bool
	PKCEChallengeMethod        string

	Audience      []string
	Scopes        []string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	ResponseModes []oauthelia2.ResponseModeType

	Lifespans      schema.IdentityProvidersOpenIDConnectLifespan
	ClaimsStrategy ClaimsStrategy

	AuthorizationSignedResponseAlg      string
	AuthorizationSignedResponseKeyID    string
	AuthorizationEncryptedResponseAlg   string
	AuthorizationEncryptedResponseEnc   string
	AuthorizationEncryptedResponseKeyID string

	IDTokenSignedResponseAlg      string
	IDTokenSignedResponseKeyID    string
	IDTokenEncryptedResponseAlg   string
	IDTokenEncryptedResponseEnc   string
	IDTokenEncryptedResponseKeyID string

	AccessTokenSignedResponseAlg      string
	AccessTokenSignedResponseKeyID    string
	AccessTokenEncryptedResponseAlg   string
	AccessTokenEncryptedResponseEnc   string
	AccessTokenEncryptedResponseKeyID string

	UserinfoSignedResponseAlg      string
	UserinfoSignedResponseKeyID    string
	UserinfoEncryptedResponseAlg   string
	UserinfoEncryptedResponseEnc   string
	UserinfoEncryptedResponseKeyID string

	IntrospectionSignedResponseAlg      string
	IntrospectionSignedResponseKeyID    string
	IntrospectionEncryptedResponseAlg   string
	IntrospectionEncryptedResponseEnc   string
	IntrospectionEncryptedResponseKeyID string

	RequestObjectSigningAlg    string
	RequestObjectEncryptionAlg string
	RequestObjectEncryptionEnc string

	TokenEndpointAuthMethod     string
	TokenEndpointAuthSigningAlg string

	RevocationEndpointAuthMethod     string
	RevocationEndpointAuthSigningAlg string

	IntrospectionEndpointAuthMethod     string
	IntrospectionEndpointAuthSigningAlg string

	PushedAuthorizationRequestEndpointAuthMethod     string
	PushedAuthorizationRequestEndpointAuthSigningAlg string

	RefreshFlowIgnoreOriginalGrantedScopes  bool
	AllowMultipleAuthenticationMethods      bool
	ClientCredentialsFlowAllowImplicitScope bool

	AuthorizationPolicy ClientAuthorizationPolicy

	ConsentPolicy         ClientConsentPolicy
	RequestedAudienceMode ClientRequestedAudienceMode

	RequestURIs    []string
	JSONWebKeys    *jose.JSONWebKeySet
	JSONWebKeysURI *url.URL

	ScopeDescriptions map[string]string
}

// Client represents the internal client definitions.
type Client interface {
	oauthelia2.Client
	oauthelia2.ResponseModeClient
	RefreshFlowScopeClient

	GetName() (name string)
	GetSectorIdentifierURI() (sector string)

	GetClaimsStrategy() (strategy ClaimsStrategy)

	GetAuthorizationSignedResponseKeyID() (kid string)
	GetAuthorizationSignedResponseAlg() (alg string)
	GetAuthorizationEncryptedResponseKeyID() (kid string)
	GetAuthorizationEncryptedResponseAlg() (alg string)
	GetAuthorizationEncryptedResponseEnc() (enc string)

	GetIDTokenSignedResponseKeyID() (kid string)
	GetIDTokenSignedResponseAlg() (alg string)
	GetIDTokenEncryptedResponseKeyID() (kid string)
	GetIDTokenEncryptedResponseAlg() (kid string)
	GetIDTokenEncryptedResponseEnc() (kid string)

	GetAccessTokenSignedResponseKeyID() (kid string)
	GetAccessTokenSignedResponseAlg() (alg string)
	GetAccessTokenEncryptedResponseKeyID() (kid string)
	GetAccessTokenEncryptedResponseAlg() (alg string)
	GetAccessTokenEncryptedResponseEnc() (enc string)
	GetEnableJWTProfileOAuthAccessTokens() bool

	GetUserinfoSignedResponseKeyID() (kid string)
	GetUserinfoSignedResponseAlg() (alg string)
	GetUserinfoEncryptedResponseKeyID() (kid string)
	GetUserinfoEncryptedResponseAlg() (alg string)
	GetUserinfoEncryptedResponseEnc() (enc string)

	GetIntrospectionSignedResponseKeyID() (kid string)
	GetIntrospectionSignedResponseAlg() (alg string)
	GetIntrospectionEncryptedResponseKeyID() (kid string)
	GetIntrospectionEncryptedResponseAlg() (kid string)
	GetIntrospectionEncryptedResponseEnc() (kid string)

	GetRequirePushedAuthorizationRequests() (enforce bool)

	GetJSONWebKeys() (jwks *jose.JSONWebKeySet)
	GetJSONWebKeysURI() (uri string)

	GetEnforcePKCE() (enforce bool)
	GetEnforcePKCEChallengeMethod() (enforce bool)
	GetPKCEChallengeMethod() (method string)

	ValidateResponseModePolicy(r oauthelia2.AuthorizeRequester) (err error)

	GetConsentResponseBody(session RequesterFormSession, form url.Values, authTime time.Time, disablePreConf bool) (body ConsentGetResponseBody)
	GetConsentPolicy() ClientConsentPolicy
	IsAuthenticationLevelSufficient(level authentication.Level, subject authorization.Subject) (sufficient bool)
	GetAuthorizationPolicyRequiredLevel(subject authorization.Subject) (level authorization.Level)
	GetAuthorizationPolicy() (policy ClientAuthorizationPolicy)

	GetEffectiveLifespan(gt oauthelia2.GrantType, tt oauthelia2.TokenType, fallback time.Duration) (lifespan time.Duration)
}

// RefreshFlowScopeClient is a client which can be customized to ignore scopes that were not originally granted.
type RefreshFlowScopeClient interface {
	oauthelia2.Client

	GetRefreshFlowIgnoreOriginalGrantedScopes(ctx context.Context) (ignoreOriginalGrantedScopes bool)
}

// Context represents the context implementation that is used by some OpenID Connect 1.0 implementations.
type Context interface {
	RootURL() (issuerURL *url.URL)
	IssuerURL() (issuerURL *url.URL, err error)
	GetClock() (clock clock.Provider)
	GetRandom() (random random.Provider)
	GetConfiguration() (config schema.Configuration)
	GetJWTWithTimeFuncOption() (option jwt.ParserOption)
	GetProviderUserAttributeResolver() expression.UserAttributeResolver

	context.Context
}

// ClaimsStrategyContext is a context used for the CustomClaimsStrategy implementation.
type ClaimsStrategyContext interface {
	GetProviderUserAttributeResolver() expression.UserAttributeResolver

	context.Context
}

type ClientContext interface {
	GetHTTPClient() *http.Client

	context.Context
}

// ClientRequesterResponder is a oauthelia2.Requster or fosite.Responder with a GetClient method.
type ClientRequesterResponder interface {
	GetClient() oauthelia2.Client
}

// IDTokenClaimsSession is a session which can return the IDTokenClaims type.
type IDTokenClaimsSession interface {
	GetIDTokenClaims() *fjwt.IDTokenClaims
}

// Configurator is an internal extension to the oauthelia2.Configurator.
type Configurator interface {
	oauthelia2.Configurator

	AuthorizationServerIssuerIdentificationProvider
}

// AuthorizationServerIssuerIdentificationProvider provides OAuth 2.0 Authorization Server Issuer Identification related methods.
type AuthorizationServerIssuerIdentificationProvider interface {
	GetAuthorizationServerIdentificationIssuer(ctx context.Context) (issuer string)
}

// JWTSecuredResponseModeProvider provides JARM related methods.
type JWTSecuredResponseModeProvider interface {
	GetJWTSecuredAuthorizeResponseModeLifespan(ctx context.Context) (lifespan time.Duration)
	GetJWTSecuredAuthorizeResponseModeStrategy(ctx context.Context) (strategy fjwt.Strategy)
	GetJWTSecuredAuthorizeResponseModeIssuer(ctx context.Context) (issuer string)
}

// IDTokenSessionContainer is similar to the oauth2.JWTSessionContainer to facilitate obtaining the headers as appropriate.
type IDTokenSessionContainer interface {
	IDTokenHeaders() *fjwt.Headers
	IDTokenClaims() *fjwt.IDTokenClaims
}

type UserDetailer interface {
	GetUsername() (username string)
	GetGroups() (groups []string)
	GetDisplayName() (name string)
	GetEmails() (emails []string)
	GetGivenName() (given string)
	GetFamilyName() (family string)
	GetMiddleName() (middle string)
	GetNickname() (nickname string)
	GetProfile() (profile string)
	GetPicture() (picture string)
	GetWebsite() (website string)
	GetGender() (gender string)
	GetBirthdate() (birthdate string)
	GetZoneInfo() (info string)
	GetLocale() (locale string)
	GetPhoneNumber() (number string)
	GetPhoneExtension() (extension string)
	GetPhoneNumberRFC3966() (number string)
	GetStreetAddress() (address string)
	GetLocality() (locality string)
	GetRegion() (region string)
	GetPostalCode() (postcode string)
	GetCountry() (country string)
	GetExtra() (extra map[string]any)
}

// ConsentGetResponseBody schema of the response body of the consent GET endpoint.
type ConsentGetResponseBody struct {
	ClientID          string            `json:"client_id"`
	ClientDescription string            `json:"client_description"`
	Scopes            []string          `json:"scopes"`
	ScopeDescriptions map[string]string `json:"scope_descriptions"`
	Audience          []string          `json:"audience"`
	PreConfiguration  bool              `json:"pre_configuration"`
	Claims            []string          `json:"claims"`
	EssentialClaims   []string          `json:"essential_claims"`
	RequireLogin      bool              `json:"require_login"`
}

// ConsentPostRequestBody schema of the request body of the consent POST endpoint.
type ConsentPostRequestBody struct {
	FlowID       *string  `json:"flow_id"`
	ClientID     string   `json:"client_id"`
	Consent      bool     `json:"consent"`
	PreConfigure bool     `json:"pre_configure"`
	Claims       []string `json:"claims"`
	SubFlow      *string  `json:"subflow"`
	UserCode     *string  `json:"user_code"`
}

// ConsentPostResponseBody schema of the response body of the consent POST endpoint.
type ConsentPostResponseBody struct {
	RedirectURI string `json:"redirect_uri,omitempty"`
	FlowID      string `json:"flow_id,omitempty"`
}

/*
CommonDiscoveryOptions represents the discovery options used in both OAuth 2.0 and OpenID Connect.
See Also:

	OpenID Connect Discovery: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
	OAuth 2.0 Discovery: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-discovery-10#section-2
*/
type CommonDiscoveryOptions struct {
	/*
		REQUIRED. URL using the https scheme with no query or fragment component that the OP asserts as its Issuer
		Identifier. If Issuer discovery is supported (see Section 2), this value MUST be identical to the issuer value
		returned by WebFinger. This also MUST be identical to the iss Claim value in ID Tokens issued from this Issuer.
	*/
	Issuer string `json:"issuer"`

	/*
		REQUIRED. URL of the OP's JSON Web Key Set [JWK] document. This contains the signing key(s) the RP uses to
		validate signatures from the OP. The JWK Set MAY also contain the Server's encryption key(s), which are used by
		RPs to encrypt requests to the Server. When both signing and encryption keys are made available, a use (Key Use)
		parameter value is REQUIRED for all keys in the referenced JWK Set to indicate each key's intended usage.
		Although some algorithms allow the same key to be used for both signatures and encryption, doing so is NOT
		RECOMMENDED, as it is less secure. The JWK x5c parameter MAY be used to provide X.509 representations of keys
		provided. When used, the bare key values MUST still be present and MUST match those in the certificate.
	*/
	JWKSURI string `json:"jwks_uri,omitempty"`

	/*
		REQUIRED. URL of the OP's OAuth 2.0 Authorization Endpoint [OpenID.Core].
		See Also:
			OpenID.Core: https://openid.net/specs/openid-connect-core-1_0.html
	*/
	AuthorizationEndpoint string `json:"authorization_endpoint"`

	/*
		URL of the OP's OAuth 2.0 Token Endpoint [OpenID.Core]. This is REQUIRED unless only the Implicit Flow is used.
		See Also:
			OpenID.Core: https://openid.net/specs/openid-connect-core-1_0.html
	*/
	TokenEndpoint string `json:"token_endpoint,omitempty"`

	/*
		REQUIRED. JSON array containing a list of the Subject Identifier types that this OP supports. Valid types
		include pairwise and public.
	*/
	SubjectTypesSupported []string `json:"subject_types_supported"`

	/*
		REQUIRED. JSON array containing a list of the OAuth 2.0 response_type values that this OP supports. Dynamic
		OpenID Providers MUST support the code, id_token, and the token id_token Response Type values.
	*/
	ResponseTypesSupported []string `json:"response_types_supported"`

	/*
		OPTIONAL. JSON array containing a list of the OAuth 2.0 Grant Type values that this OP supports. Dynamic OpenID
		Providers MUST support the authorization_code and implicit Grant Type values and MAY support other Grant Types.
		If omitted, the default value is ["authorization_code", "implicit"].
	*/
	GrantTypesSupported []string `json:"grant_types_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the OAuth 2.0 response_mode values that this OP supports, as specified
		in OAuth 2.0 Multiple Response Type Encoding Practices [OAuth.Responses]. If omitted, the default for Dynamic
		OpenID Providers is ["query", "fragment"].
	*/
	ResponseModesSupported []string `json:"response_modes_supported,omitempty"`

	/*
		RECOMMENDED. JSON array containing a list of the OAuth 2.0 [RFC6749] scope values that this server supports.
		The server MUST support the openid scope value. Servers MAY choose not to advertise some supported scope values
		even when this parameter is used, although those defined in [OpenID.Core] SHOULD be listed, if supported.
		See Also:
			OAuth 2.0: https://datatracker.ietf.org/doc/html/rfc6749
			OpenID.Core: https://openid.net/specs/openid-connect-core-1_0.html
	*/
	ScopesSupported []string `json:"scopes_supported,omitempty"`

	/*
		RECOMMENDED. JSON array containing a list of the Claim Names of the Claims that the OpenID Provider MAY be able
		to supply values for. Note that for privacy or other reasons, this might not be an exhaustive list.
	*/
	ClaimsSupported []string `json:"claims_supported,omitempty"`

	/*
		OPTIONAL. Languages and scripts supported for the user interface, represented as a JSON array of BCP47 [RFC5646]
		language tag values.
		See Also:
			BCP47: https://datatracker.ietf.org/doc/html/rfc5646
	*/
	UILocalesSupported []string `json:"ui_locales_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of Client Authentication methods supported by this Token Endpoint. The
		options are client_secret_post, client_secret_basic, client_secret_jwt, and private_key_jwt, as described in
		Section 9 of OpenID Connect Core 1.0 [OpenID.Core]. Other authentication methods MAY be defined by extensions.
		If omitted, the default is client_secret_basic -- the HTTP Basic Authentication Scheme specified in Section
		2.3.1 of OAuth 2.0 [RFC6749].
		See Also:
			OAuth 2.0: https://datatracker.ietf.org/doc/html/rfc6749
			OpenID.Core Section 9: https://openid.net/specs/openid-connect-core-1_0.html#ClientAuthentication
	*/
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWS signing algorithms (alg values) supported by the Token Endpoint
		for the signature on the JWT [JWT] used to authenticate the Client at the Token Endpoint for the private_key_jwt
		and client_secret_jwt authentication methods. Servers SHOULD support RS256. The value none MUST NOT be used.
		See Also:
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL. URL of a page containing human-readable information that developers might want or need to know when
		using the OpenID Provider. In particular, if the OpenID Provider does not support Dynamic Client Registration,
		then information on how to register Clients needs to be provided in this documentation.
	*/
	ServiceDocumentation string `json:"service_documentation,omitempty"`

	/*
		OPTIONAL. URL that the OpenID Provider provides to the person registering the Client to read about the OP's
		requirements on how the Relying Party can use the data provided by the OP. The registration process SHOULD
		display this URL to the person registering the Client if it is given.
	*/
	OPPolicyURI string `json:"op_policy_uri,omitempty"`

	/*
		OPTIONAL. URL that the OpenID Provider provides to the person registering the Client to read about OpenID
		Provider's terms of service. The registration process SHOULD display this URL to the person registering the
		Client if it is given.
	*/
	OPTOSURI string `json:"op_tos_uri,omitempty"`
}

// OAuth2DiscoveryOptions represents the discovery options specific to OAuth 2.0.
type OAuth2DiscoveryOptions struct {
	/*
		 	OPTIONAL. URL of the authorization server's OAuth 2.0 introspection endpoint [RFC7662].
			See Also:
				OAuth 2.0 Token Introspection: https://datatracker.ietf.org/doc/html/rfc7662
	*/
	IntrospectionEndpoint string `json:"introspection_endpoint,omitempty"`

	/*
		OPTIONAL. URL of the authorization server's OAuth 2.0 revocation endpoint [RFC7009].
		See Also:
			OAuth 2.0 Token Revocation: https://datatracker.ietf.org/doc/html/rfc7009
	*/
	RevocationEndpoint string `json:"revocation_endpoint,omitempty"`

	/*
		OPTIONAL. URL of the authorization server's OAuth 2.0 Dynamic Client Registration endpoint [RFC7591].
		See Also:
			OAuth 2.0 Dynamic Client Registration Protocol: https://datatracker.ietf.org/doc/html/rfc7591
	*/
	RegistrationEndpoint string `json:"registration_endpoint,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of client authentication methods supported by this introspection endpoint.
		The valid client authentication method values are those registered in the IANA "OAuth Token Endpoint
		Authentication Methods" registry [IANA.OAuth.Parameters] or those registered in the IANA "OAuth Access Token Types"
		registry [IANA.OAuth.Parameters]. (These values are and will remain distinct, due to Section 7.2.) If omitted,
		the set of supported authentication methods MUST be determined by other means.
		See Also:
			IANA.OAuth.Parameters: https://www.iana.org/assignments/oauth-parameters/oauth-parameters.xhtml
			OAuth 2.0 Authorization Server Metadata - Updated Registration Instructions: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-discovery-10#section-7.2
	*/
	IntrospectionEndpointAuthMethodsSupported []string `json:"introspection_endpoint_auth_methods_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWS signing algorithms ("alg" values) supported by the
		introspection endpoint for the signature on the JWT [JWT] used to authenticate the client at the introspection
		endpoint for the "private_key_jwt" and "client_secret_jwt" authentication methods. This metadata entry MUST be
		present if either of these authentication methods are specified in the
		"introspection_endpoint_auth_methods_supported" entry. No default algorithms are implied if this entry is omitted.
		The value "none" MUST NOT be used.
		See Also:
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	IntrospectionEndpointAuthSigningAlgValuesSupported []string `json:"introspection_endpoint_auth_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of client authentication methods supported by this revocation endpoint.
		The valid client authentication method values are those registered in the IANA "OAuth Token Endpoint
		Authentication Methods" registry [IANA.OAuth.Parameters]. If omitted, the default is "client_secret_basic" --
		the HTTP Basic Authentication Scheme specified in Section 2.3.1 of OAuth 2.0 [RFC6749].
		See Also:
			IANA.OAuth.Parameters: https://www.iana.org/assignments/oauth-parameters/oauth-parameters.xhtml
			OAuth 2.0 - Client Password: https://datatracker.ietf.org/doc/html/rfc6749#section-2.3.1
	*/
	RevocationEndpointAuthMethodsSupported []string `json:"revocation_endpoint_auth_methods_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWS signing algorithms ("alg" values) supported by the revocation
		endpoint for the signature on the JWT [JWT] used to authenticate the client at the revocation endpoint for the
		"private_key_jwt" and "client_secret_jwt" authentication methods. This metadata entry MUST be present if either
		of these authentication methods are specified in the "revocation_endpoint_auth_methods_supported" entry. No
		default algorithms are implied if this entry is omitted. The value "none" MUST NOT be used.
		See Also:
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	RevocationEndpointAuthSigningAlgValuesSupported []string `json:"revocation_endpoint_auth_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of PKCE [RFC7636] code challenge methods supported by this authorization
		server. Code challenge method values are used in the "code_challenge_method" parameter defined in Section 4.3 of
		[RFC7636]. The valid code challenge method values are those registered in the IANA "PKCE Code Challenge Methods"
		registry [IANA.OAuth.Parameters]. If omitted, the authorization server does not support PKCE.
		See Also:
			PKCE: https://datatracker.ietf.org/doc/html/rfc7636
			IANA.OAuth.Parameters: https://www.iana.org/assignments/oauth-parameters/oauth-parameters.xhtml
	*/
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported,omitempty"`
}

type OAuth2JWTIntrospectionResponseDiscoveryOptions struct {
	/*
		OPTIONAL.  JSON array containing a list of the JWS [RFC7515] signing algorithms ("alg" values) as defined in JWA
		[RFC7518] supported by the introspection endpoint to sign the response.
	*/
	IntrospectionSigningAlgValuesSupported []string `json:"introspection_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL.  JSON array containing a list of the JWE [RFC7516] encryption algorithms ("alg" values) as defined in
		JWA [RFC7518] supported by the introspection endpoint to encrypt the content encryption key for introspection
		responses (content key encryption).
	*/
	IntrospectionEncryptionAlgValuesSupported []string `json:"introspection_encryption_alg_values_supported,omitempty"`

	/*
		OPTIONAL.  JSON array containing a list of the JWE [RFC7516] encryption algorithms ("enc" values) as defined in
		JWA [RFC7518] supported by the introspection endpoint to encrypt the response (content encryption).
	*/
	IntrospectionEncryptionEncValuesSupported []string `json:"introspection_encryption_enc_values_supported,omitempty"`
}

type OAuth2DeviceAuthorizationGrantDiscoveryOptions struct {
	/*
		OPTIONAL.  URL of the authorization server's device authorization endpoint, as defined in Section 3.1.
	*/
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
}

type OAuth2MutualTLSClientAuthenticationDiscoveryOptions struct {
	/*
		OPTIONAL. Boolean value indicating server support for mutual-TLS client certificate-bound access tokens. If
		omitted, the default value is false.
	*/
	TLSClientCertificateBoundAccessTokens bool `json:"tls_client_certificate_bound_access_tokens"`

	/*
		OPTIONAL. A JSON object containing alternative authorization server endpoints that, when present, an OAuth
		client intending to do mutual TLS uses in preference to the conventional endpoints. The parameter value itself
		consists of one or more endpoint parameters, such as token_endpoint, revocation_endpoint,
		introspection_endpoint, etc., conventionally defined for the top level of authorization server metadata. An
		OAuth client intending to do mutual TLS (for OAuth client authentication and/or to acquire or use
		certificate-bound tokens) when making a request directly to the authorization server MUST use the alias URL of
		the endpoint within the mtls_endpoint_aliases, when present, in preference to the endpoint URL of the same name
		at the top level of metadata. When an endpoint is not present in mtls_endpoint_aliases, then the client uses the
		conventional endpoint URL defined at the top level of the authorization server metadata. Metadata parameters
		within mtls_endpoint_aliases that do not define endpoints to which an OAuth client makes a direct request have
		no meaning and SHOULD be ignored.
	*/
	MutualTLSEndpointAliases OAuth2MutualTLSClientAuthenticationAliasesDiscoveryOptions `json:"mtls_endpoint_aliases"`
}

type OAuth2MutualTLSClientAuthenticationAliasesDiscoveryOptions struct {
	AuthorizationEndpoint              string `json:"authorization_endpoint,omitempty"`
	TokenEndpoint                      string `json:"token_endpoint,omitempty"`
	IntrospectionEndpoint              string `json:"introspection_endpoint,omitempty"`
	RevocationEndpoint                 string `json:"revocation_endpoint,omitempty"`
	EndSessionEndpoint                 string `json:"end_session_endpoint,omitempty"`
	UserinfoEndpoint                   string `json:"userinfo_endpoint,omitempty"`
	BackChannelAuthenticationEndpoint  string `json:"backchannel_authentication_endpoint,omitempty"`
	FederationRegistrationEndpoint     string `json:"federation_registration_endpoint,omitempty"`
	PushedAuthorizationRequestEndpoint string `json:"pushed_authorization_request_endpoint,omitempty"`
	RegistrationEndpoint               string `json:"registration_endpoint,omitempty"`
}

type OAuth2JWTSecuredAuthorizationRequestDiscoveryOptions struct {
	/*
		Indicates where authorization request needs to be protected as Request Object and provided through either
		request or request_uri parameter.
	*/
	RequireSignedRequestObject bool `json:"require_signed_request_object"`
}

type OAuth2IssuerIdentificationDiscoveryOptions struct {
	AuthorizationResponseIssuerParameterSupported bool `json:"authorization_response_iss_parameter_supported"`
}

// OAuth2PushedAuthorizationDiscoveryOptions represents the well known discovery document specific to the
// OAuth 2.0 Pushed Authorization Requests (RFC9126) implementation.
//
// OAuth 2.0 Pushed Authorization Requests: https://datatracker.ietf.org/doc/html/rfc9126#section-5
type OAuth2PushedAuthorizationDiscoveryOptions struct {
	/*
	   The URL of the pushed authorization request endpoint at which a client can post an authorization request to
	   exchange for a "request_uri" value usable at the authorization server.
	*/
	PushedAuthorizationRequestEndpoint string `json:"pushed_authorization_request_endpoint"`

	/*
		Boolean parameter indicating whether the authorization server accepts authorization request data only via PAR.
		If omitted, the default value is "false".
	*/
	RequirePushedAuthorizationRequests bool `json:"require_pushed_authorization_requests"`
}

// OpenIDConnectDiscoveryOptions represents the discovery options specific to OpenID Connect.
type OpenIDConnectDiscoveryOptions struct {
	/*
		REQUIRED. JSON array containing a list of the JWS signing algorithms (alg values) supported by the OP for the ID
		Token to encode the Claims in a JWT [JWT]. The algorithm RS256 MUST be included. The value none MAY be supported,
		but MUST NOT be used unless the Response Type used returns no ID Token from the Authorization Endpoint (such as
		when using the Authorization Code Flow).
		See Also:
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWE encryption algorithms (alg values) supported by the OP for the
		ID Token to encode the Claims in a JWT [JWT].
		See Also:
			JWE: https://datatracker.ietf.org/doc/html/rfc7516
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	IDTokenEncryptionAlgValuesSupported []string `json:"id_token_encryption_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWE encryption algorithms (enc values) supported by the OP for the
		ID Token to encode the Claims in a JWT [JWT].
		See Also:
			JWE: https://datatracker.ietf.org/doc/html/rfc7516
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	IDTokenEncryptionEncValuesSupported []string `json:"id_token_encryption_enc_values_supported,omitempty"`

	/*
		RECOMMENDED. URL of the OP's UserInfo Endpoint [OpenID.Core]. This URL MUST use the https scheme and MAY contain
		port, path, and query parameter components.
		See Also:
			OpenID.Core: https://openid.net/specs/openid-connect-core-1_0.html
	*/
	UserinfoEndpoint string `json:"userinfo_endpoint,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWS [JWS] signing algorithms (alg values) [JWA] supported by the
		UserInfo Endpoint to encode the Claims in a JWT [JWT]. The value none MAY be included.
		See Also:
			JWS: https://datatracker.ietf.org/doc/html/rfc7515
			JWA: https://datatracker.ietf.org/doc/html/rfc7518
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	UserinfoSigningAlgValuesSupported []string `json:"userinfo_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWE [JWE] encryption algorithms (alg values) [JWA] supported by
		the UserInfo Endpoint to encode the Claims in a JWT [JWT].
		See Also:
			JWE: https://datatracker.ietf.org/doc/html/rfc7516
			JWA: https://datatracker.ietf.org/doc/html/rfc7518
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	UserinfoEncryptionAlgValuesSupported []string `json:"userinfo_encryption_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWE encryption algorithms (enc values) [JWA] supported by the
		UserInfo Endpoint to encode the Claims in a JWT [JWT].
		See Also:
			JWE: https://datatracker.ietf.org/doc/html/rfc7516
			JWA: https://datatracker.ietf.org/doc/html/rfc7518
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	UserinfoEncryptionEncValuesSupported []string `json:"userinfo_encryption_enc_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWS signing algorithms (alg values) supported by the OP for Request
		Objects, which are described in Section 6.1 of OpenID Connect Core 1.0 [OpenID.Core]. These algorithms are used
		both when the Request Object is passed by value (using the request parameter) and when it is passed by reference
		(using the request_uri parameter). Servers SHOULD support none and RS256.
	*/
	RequestObjectSigningAlgValuesSupported []string `json:"request_object_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWE encryption algorithms (alg values) supported by the OP for
		Request Objects. These algorithms are used both when the Request Object is passed by value and when it is passed
		by reference.
		See Also:
			JWE: https://datatracker.ietf.org/doc/html/rfc7516
	*/
	RequestObjectEncryptionAlgValuesSupported []string `json:"request_object_encryption_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWE encryption algorithms (enc values) supported by the OP for
		Request Objects. These algorithms are used both when the Request Object is passed by value and when it is passed
		by reference.
		See Also:
			JWE: https://datatracker.ietf.org/doc/html/rfc7516
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	RequestObjectEncryptionEncValuesSupported []string `json:"request_object_encryption_enc_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the Authentication Context Class References that this OP supports.
	*/
	ACRValuesSupported []string `json:"acr_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the display parameter values that the OpenID Provider supports. These
		values are described in Section 3.1.2.1 of OpenID Connect Core 1.0 [OpenID.Core].
		See Also:
			OpenID.Core Section 3.1.2.1: https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest
	*/
	DisplayValuesSupported []string `json:"display_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the Claim Types that the OpenID Provider supports. These Claim Types
		are described in Section 5.6 of OpenID Connect Core 1.0 [OpenID.Core]. Values defined by this specification are
		normal, aggregated, and distributed. If omitted, the implementation supports only normal Claims.
		See Also:
			OpenID.Core Section 5.6: https://openid.net/specs/openid-connect-core-1_0.html#ClaimTypes
	*/
	ClaimTypesSupported []string `json:"claim_types_supported,omitempty"`

	/*
		OPTIONAL. Languages and scripts supported for values in Claims being returned, represented as a JSON array of
		BCP47 [RFC5646] language tag values. Not all languages and scripts are necessarily supported for all Claim values.
		See Also:
			BCP47: https://datatracker.ietf.org/doc/html/rfc5646
	*/
	ClaimLocalesSupported []string `json:"claims_locales_supported,omitempty"`

	/*
		OPTIONAL. Boolean value specifying whether the OP supports use of the request parameter, with true indicating
		support. If omitted, the default value is false.
	*/
	RequestParameterSupported bool `json:"request_parameter_supported"`

	/*
		OPTIONAL. Boolean value specifying whether the OP supports use of the request_uri parameter, with true indicating
		support. If omitted, the default value is true.
	*/
	RequestURIParameterSupported bool `json:"request_uri_parameter_supported"`

	/*
		OPTIONAL. Boolean value specifying whether the OP requires any request_uri values used to be pre-registered using
		the request_uris registration parameter. Pre-registration is REQUIRED when the value is true. If omitted, the
		default value is false.
	*/
	RequireRequestURIRegistration bool `json:"require_request_uri_registration"`

	/*
		OPTIONAL. Boolean value specifying whether the OP supports use of the claims parameter, with true indicating
		support. If omitted, the default value is false.
	*/
	ClaimsParameterSupported bool `json:"claims_parameter_supported"`
}

// OpenIDConnectFrontChannelLogoutDiscoveryOptions represents the discovery options specific to
// OpenID Connect Front-Channel Logout functionality.
// See Also:
//
//	OpenID Connect Front-Channel Logout: https://openid.net/specs/openid-connect-frontchannel-1_0.html#OPLogout
type OpenIDConnectFrontChannelLogoutDiscoveryOptions struct {
	/*
		OPTIONAL. Boolean value specifying whether the OP supports HTTP-based logout, with true indicating support. If
		omitted, the default value is false.
	*/
	FrontChannelLogoutSupported bool `json:"frontchannel_logout_supported"`

	/*
		OPTIONAL. Boolean value specifying whether the OP can pass iss (issuer) and sid (session ID) query parameters to
		identify the RP session with the OP when the frontchannel_logout_uri is used. If supported, the sid Claim is also
		included in ID Tokens issued by the OP. If omitted, the default value is false.
	*/
	FrontChannelLogoutSessionSupported bool `json:"frontchannel_logout_session_supported"`
}

// OpenIDConnectBackChannelLogoutDiscoveryOptions represents the discovery options specific to
// OpenID Connect Back-Channel Logout functionality.
// See Also:
//
//	OpenID Connect Back-Channel Logout: https://openid.net/specs/openid-connect-backchannel-1_0.html#BCSupport
type OpenIDConnectBackChannelLogoutDiscoveryOptions struct {
	/*
		OPTIONAL. Boolean value specifying whether the OP supports back-channel logout, with true indicating support.
		If omitted, the default value is false.
	*/
	BackChannelLogoutSupported bool `json:"backchannel_logout_supported"`

	/*
		OPTIONAL. Boolean value specifying whether the OP can pass a sid (session ID) Claim in the Logout Token to
		identify the RP session with the OP. If supported, the sid Claim is also included in ID Tokens issued by the OP.
		If omitted, the default value is false.
	*/
	BackChannelLogoutSessionSupported bool `json:"backchannel_logout_session_supported"`
}

// OpenIDConnectSessionManagementDiscoveryOptions represents the discovery options specific to OpenID Connect 1.0
// Session Management.
//
// To support OpenID Connect Session Management, the RP needs to obtain the Session Management related OP metadata. This
// OP metadata is normally obtained via the OP's Discovery response, as described in OpenID Connect Discovery 1.0, or
// MAY be learned via other mechanisms. This OpenID Provider Metadata parameter MUST be included in the Server's
// discovery responses when Session Management and Discovery are supported.
//
// See Also:
//
// OpenID Connect 1.0 Session Management: https://openid.net/specs/openid-connect-session-1_0.html
type OpenIDConnectSessionManagementDiscoveryOptions struct {
	/*
		REQUIRED. URL of an OP iframe that supports cross-origin communications for session state information with the
		RP Client, using the HTML5 postMessage API. This URL MUST use the https scheme and MAY contain port, path, and
		query parameter components. The page is loaded from an invisible iframe embedded in an RP page so that it can
		run in the OP's security context. It accepts postMessage requests from the relevant RP iframe and uses
		postMessage to post back the login status of the End-User at the OP.
	*/
	CheckSessionIFrame string `json:"check_session_iframe"`
}

// OpenIDConnectRPInitiatedLogoutDiscoveryOptions represents the discovery options specific to
// OpenID Connect RP-Initiated Logout 1.0.
//
// To support OpenID Connect RP-Initiated Logout, the RP needs to obtain the RP-Initiated Logout related OP metadata.
// This OP metadata is normally obtained via the OP's Discovery response, as described in OpenID Connect Discovery 1.0,
// or MAY be learned via other mechanisms. This OpenID Provider Metadata parameter MUST be included in the Server's
// discovery responses when RP-Initiated Logout and Discovery are supported.
//
// See Also:
//
// OpenID Connect RP-Initiated Logout 1.0: https://openid.net/specs/openid-connect-rpinitiated-1_0.html
type OpenIDConnectRPInitiatedLogoutDiscoveryOptions struct {
	/*
		REQUIRED. URL at the OP to which an RP can perform a redirect to request that the End-User be logged out at the
		OP. This URL MUST use the https scheme and MAY contain port, path, and query parameter components.
	*/
	EndSessionEndpoint string `json:"end_session_endpoint"`
}

// OpenIDConnectPromptCreateDiscoveryOptions represents the discovery options specific to Initiating User Registration
// via OpenID Connect 1.0 functionality.
//
// This specification extends the OpenID Connect Discovery Metadata Section 3.
//
// See Also:
//
//	Initiating User Registration via OpenID Connect 1.0: https://openid.net/specs/openid-connect-prompt-create-1_0.html
type OpenIDConnectPromptCreateDiscoveryOptions struct {
	/*
		OPTIONAL. JSON array containing the list of prompt values that this OP supports.

		This metadata element is OPTIONAL in the context of the OpenID Provider not supporting the create value. If
		omitted, the Relying Party should assume that this specification is not supported. The OpenID Provider MAY
		provide this metadata element even if it doesn't support the create value.
		Specific to this specification, a value of create in the array indicates to the Relying party that this OpenID
		Provider supports this specification. If an OpenID Provider supports this specification it MUST define this metadata
		element in the openid-configuration file. Additionally, if this metadata element is defined by the OpenID
		Provider, the OP must also specify all other prompt values which it supports.
		See Also:
			OpenID.PromptCreate: https://openid.net/specs/openid-connect-prompt-create-1_0.html
	*/
	PromptValuesSupported []string `json:"prompt_values_supported,omitempty"`
}

// OpenIDConnectClientInitiatedBackChannelAuthFlowDiscoveryOptions represents the discovery options specific to
// OpenID Connect Client-Initiated Backchannel Authentication Flow - Core 1.0
//
// The following authorization server metadata parameters are introduced by this specification for OPs publishing their
// support of the CIBA flow and details thereof.
//
// See Also:
//
// OpenID Connect Client-Initiated Backchannel Authentication Flow - Core 1.0:
// https://openid.net/specs/openid-client-initiated-backchannel-authentication-core-1_0.html#rfc.section.4
type OpenIDConnectClientInitiatedBackChannelAuthFlowDiscoveryOptions struct {
	/*
		REQUIRED. URL of the OP's Backchannel Authentication Endpoint as defined in Section 7.
	*/
	BackChannelAuthenticationEndpoint string `json:"backchannel_authentication_endpoint"`

	/*
		REQUIRED. JSON array containing one or more of the following values: poll, ping, and push.
	*/
	BackChannelTokenDeliveryModesSupported []string `json:"backchannel_token_delivery_modes_supported"`

	/*
		OPTIONAL. JSON array containing a list of the JWS signing algorithms (alg values) supported by the OP for signed
		authentication requests, which are described in Section 7.1.1. If omitted, signed authentication requests are
		not supported by the OP.
	*/
	BackChannelAuthRequestSigningAlgValuesSupported []string `json:"backchannel_authentication_request_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL. Boolean value specifying whether the OP supports the use of the user_code parameter, with true
		indicating support. If omitted, the default value is false.
	*/
	BackChannelUserCodeParameterSupported bool `json:"backchannel_user_code_parameter_supported"`
}

// OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions represents the discovery options specific to
// JWT Secured Authorization Response Modes for OAuth 2.0 (JARM).
//
// Authorization servers SHOULD publish the supported algorithms for signing and encrypting the JWT of an authorization
// response by utilizing OAuth 2.0 Authorization Server Metadata [RFC8414] parameters. The following parameters are
// introduced by this specification.
//
// See Also:
//
// JWT Secured Authorization Response Modes for OAuth 2.0 (JARM):
// https://openid.net/specs/oauth-v2-jarm.html#name-authorization-server-metada
type OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions struct {
	/*
		OPTIONAL. A JSON array containing a list of the JWS [RFC7515] signing algorithms (alg values) supported by the
		authorization endpoint to sign the response.
	*/
	AuthorizationSigningAlgValuesSupported []string `json:"authorization_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL. A JSON array containing a list of the JWE [RFC7516] encryption algorithms (alg values) supported by
		the authorization endpoint to encrypt the response.
	*/
	AuthorizationEncryptionAlgValuesSupported []string `json:"authorization_encryption_alg_values_supported,omitempty"`

	/*
		OPTIONAL. A JSON array containing a list of the JWE [RFC7516] encryption algorithms (enc values) supported by
		the authorization endpoint to encrypt the response.
	*/
	AuthorizationEncryptionEncValuesSupported []string `json:"authorization_encryption_enc_values_supported,omitempty"`
}

type OpenIDFederationDiscoveryOptions struct {
	/*
		OPTIONAL. URL of the OP's federation-specific Dynamic Client Registration Endpoint. If the OP supports explicit
		client registration as described in Section 10.2, then this claim is REQUIRED.
	*/
	FederationRegistrationEndpoint string `json:"federation_registration_endpoint,omitempty"`

	/*
		REQUIRED. Array specifying the federation types supported. Federation-type values defined by this specification
		are automatic and explicit.
	*/
	ClientRegistrationTypesSupported []string `json:"client_registration_types_supported"`

	/*
		OPTIONAL. A JSON Object defining the client authentications supported for each endpoint. The endpoint names are
		defined in the IANA "OAuth Authorization Server Metadata" registry [IANA.OAuth.Parameters]. Other endpoints and
		authentication methods are possible if made recognizable according to established standards and not in conflict
		with the operating principles of this specification. In OpenID Connect Core, no client authentication is
		performed at the authentication endpoint. Instead, the request itself is authenticated. The OP maps information
		in the request (like the redirect_uri) to information it has gained on the client through static or dynamic
		registration. If the mapping is successful, the request can be processed. If the RP uses Automatic Registration,
		as defined in Section 10.1, the OP has no prior knowledge of the RP. Therefore, the OP must start by gathering
		information about the RP using the process outlined in Section 6. Once it has the RP's metadata, the OP can
		verify the request in the same way as if it had known the RP's metadata beforehand. To make the request
		verification more secure, we demand the use of a client authentication or verification method that proves that
		the RP is in possession of a key that appears in the RP's metadata.
	*/
	RequestAuthenticationMethodsSupported []string `json:"request_authentication_methods_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWS signing algorithms (alg values) supported for the signature on
		the JWT [RFC7519] used in the request_object contained in the request parameter of an authorization request or
		in the private_key_jwt of a pushed authorization request. This entry MUST be present if either of these
		authentication methods are specified in the request_authentication_methods_supported entry. No default
		algorithms are implied if this entry is omitted. Servers SHOULD support RS256. The value none MUST NOT be used.
	*/
	RequestAuthenticationSigningAlgValuesSupported []string `json:"request_authentication_signing_alg_values_supported,omitempty"`
}

type OpenIDConnectIdentityAssurance struct {
	/*
		Required. JSON array containing all supported trust frameworks. This array shall have at least one member.
	*/
	TrustFrameworksSupported []string `json:"trust_frameworks_supported"`

	/*
			Required. JSON array containing all claims supported within verified_claims. claims that are not present in
		    this array shall not be returned within the verified_claims object. This array shall have at least one member.
	*/
	ClaimsInVerifiedClaimsSupported []string `json:"claims_in_verified_claims_supported"`

	/*
			Required when one or more type of evidence is supported. JSON array containing all types of identity evidence
		    the OP uses. This array shall have at least one member. Members of this array should only be the types of
		    evidence supported by the OP in the evidence element (see section 5.4.4 of [IDA-verified-claims]).
	*/
	EvidenceSupported []string `json:"evidence_supported,omitempty"`

	/*
		Required when evidence_supported contains "document". JSON array containing all identity document types
		utilized by the OP for identity verification. This array shall have at least one member.
	*/
	DocumentsSupported []string `json:"documents_supported,omitempty"`

	/*
		Optional. JSON array containing the verification methods the OP supports for evidences of type "document" (see
		[predefined_values_page]). When present this array shall have at least one member.
	*/
	DocumentsMethodsSupported []string `json:"documents_methods_supported,omitempty"`

	/*
		Optional. JSON array containing the check methods the OP supports for evidences of type "document" (see
		[predefined_values_page]). When present this array shall have at least one member.
	*/
	DocumentsCheckMethodsSupported []string `json:"documents_check_methods_supported,omitempty"`

	/*
		Required when evidence_supported contains "electronic_record". JSON array containing all electronic record types
		the OP supports (see [predefined_values_page]). When present this array shall have at least one member.
	*/
	ElectronicRecordsSupported []string `json:"electronic_records_supported,omitempty"`
}

// OAuth2WellKnownConfiguration represents the well known discovery document specific to OAuth 2.0.
type OAuth2WellKnownConfiguration struct {
	CommonDiscoveryOptions
	OAuth2DiscoveryOptions
	*OAuth2DeviceAuthorizationGrantDiscoveryOptions
	*OAuth2MutualTLSClientAuthenticationDiscoveryOptions
	*OAuth2IssuerIdentificationDiscoveryOptions
	*OAuth2JWTIntrospectionResponseDiscoveryOptions
	*OAuth2JWTSecuredAuthorizationRequestDiscoveryOptions
	*OAuth2PushedAuthorizationDiscoveryOptions
}

type OAuth2WellKnownSignedConfiguration struct {
	OAuth2WellKnownConfiguration

	/*
			A JWT containing metadata values about the authorization server as claims. This is a string value consisting of
		    the entire signed JWT. A "signed_metadata" metadata value SHOULD NOT appear as a claim in the JWT.
	*/
	SignedMetadata string `json:"signed_metadata,omitempty"`
}

func (claims *OAuth2WellKnownSignedConfiguration) ToMap() (result fjwt.MapClaims) {
	return fjwt.NewMapClaims(claims)
}

// OpenIDConnectWellKnownConfiguration represents the well known discovery document specific to OpenID Connect.
type OpenIDConnectWellKnownConfiguration struct {
	OAuth2WellKnownConfiguration

	OpenIDConnectDiscoveryOptions
	*OpenIDConnectFrontChannelLogoutDiscoveryOptions
	*OpenIDConnectBackChannelLogoutDiscoveryOptions
	*OpenIDConnectSessionManagementDiscoveryOptions
	*OpenIDConnectRPInitiatedLogoutDiscoveryOptions
	*OpenIDConnectPromptCreateDiscoveryOptions
	*OpenIDConnectClientInitiatedBackChannelAuthFlowDiscoveryOptions
	*OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions
	*OpenIDFederationDiscoveryOptions
	*OpenIDConnectIdentityAssurance
}

type OpenIDConnectWellKnownSignedConfiguration struct {
	OpenIDConnectWellKnownConfiguration

	/*
			A JWT containing metadata values about the authorization server as claims. This is a string value consisting of
		    the entire signed JWT. A "signed_metadata" metadata value SHOULD NOT appear as a claim in the JWT.
	*/
	SignedMetadata string `json:"signed_metadata,omitempty"`
}

func (claims *OpenIDConnectWellKnownSignedConfiguration) ToMap() (result fjwt.MapClaims) {
	return fjwt.NewMapClaims(claims)
}

type FormSession interface {
	GetForm() (form url.Values, err error)
}

type RequesterFormSession interface {
	FormSession

	GetRequestedAt() time.Time

	GetRequestedScopes() []string
	GetRequestedAudience() []string

	GetGrantedScopes() []string
	GetGrantedAudience() []string
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

var (
	_ Client                                                       = (*RegisteredClient)(nil)
	_ oauthelia2.Client                                            = (*RegisteredClient)(nil)
	_ oauthelia2.UserInfoClient                                    = (*RegisteredClient)(nil)
	_ oauthelia2.RotatedClientSecretsClient                        = (*RegisteredClient)(nil)
	_ oauthelia2.ProofKeyCodeExchangeClient                        = (*RegisteredClient)(nil)
	_ oauthelia2.ClientAuthenticationPolicyClient                  = (*RegisteredClient)(nil)
	_ oauthelia2.JARClient                                         = (*RegisteredClient)(nil)
	_ oauthelia2.AuthenticationMethodClient                        = (*RegisteredClient)(nil)
	_ oauthelia2.RefreshFlowScopeClient                            = (*RegisteredClient)(nil)
	_ oauthelia2.RevokeFlowRevokeRefreshTokensExplicitClient       = (*RegisteredClient)(nil)
	_ oauthelia2.JARMClient                                        = (*RegisteredClient)(nil)
	_ oauthelia2.PushedAuthorizationRequestClient                  = (*RegisteredClient)(nil)
	_ oauthelia2.ResponseModeClient                                = (*RegisteredClient)(nil)
	_ oauthelia2.ClientCredentialsFlowRequestedScopeImplicitClient = (*RegisteredClient)(nil)
	_ oauthelia2.RequestedAudienceImplicitClient                   = (*RegisteredClient)(nil)
	_ oauthelia2.JWTProfileClient                                  = (*RegisteredClient)(nil)
	_ oauthelia2.IntrospectionJWTResponseClient                    = (*RegisteredClient)(nil)

	_ RequesterFormSession = (*model.OAuth2ConsentSession)(nil)
	_ RequesterFormSession = (*model.OAuth2DeviceCodeSession)(nil)
)
