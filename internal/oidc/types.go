package oidc

import (
	"net/url"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
	"github.com/ory/herodot"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewSession creates a new empty OpenIDSession struct.
func NewSession() (session *model.OpenIDSession) {
	return &model.OpenIDSession{
		DefaultSession: &openid.DefaultSession{
			Claims: &jwt.IDTokenClaims{
				Extra: map[string]any{},
			},
			Headers: &jwt.Headers{
				Extra: map[string]any{},
			},
		},
		Extra: map[string]any{},
	}
}

// NewSessionWithAuthorizeRequest uses details from an AuthorizeRequester to generate an OpenIDSession.
func NewSessionWithAuthorizeRequest(issuer *url.URL, kid, username string, amr []string, extra map[string]any,
	authTime time.Time, consent *model.OAuth2ConsentSession, requester fosite.AuthorizeRequester) (session *model.OpenIDSession) {
	if extra == nil {
		extra = map[string]any{}
	}

	session = &model.OpenIDSession{
		DefaultSession: &openid.DefaultSession{
			Claims: &jwt.IDTokenClaims{
				Subject:     consent.Subject.UUID.String(),
				Issuer:      issuer.String(),
				AuthTime:    authTime,
				RequestedAt: consent.RequestedAt,
				IssuedAt:    time.Now(),
				Nonce:       requester.GetRequestForm().Get(ClaimNonce),
				Audience:    requester.GetGrantedAudience(),
				Extra:       extra,

				AuthenticationMethodsReferences: amr,
			},
			Headers: &jwt.Headers{
				Extra: map[string]any{
					JWTHeaderKeyIdentifier: kid,
				},
			},
			Subject:  consent.Subject.UUID.String(),
			Username: username,
		},
		Extra:       map[string]any{},
		ClientID:    requester.GetClient().GetID(),
		ChallengeID: consent.ChallengeID,
	}

	// Ensure required audience value of the client_id exists.
	if !utils.IsStringInSlice(requester.GetClient().GetID(), session.Claims.Audience) {
		session.Claims.Audience = append(session.Claims.Audience, requester.GetClient().GetID())
	}

	session.Claims.Add(ClaimAuthorizedParty, session.ClientID)
	session.Claims.Add(ClaimClientIdentifier, session.ClientID)

	return session
}

// OpenIDConnectProvider for OpenID Connect.
type OpenIDConnectProvider struct {
	fosite.OAuth2Provider

	Store      *OpenIDConnectStore
	KeyManager *KeyManager

	herodot *herodot.JSONWriter

	discovery OpenIDConnectWellKnownConfiguration
}

// OpenIDConnectStore is Authelia's internal representation of the fosite.Storage interface. It maps the following
// interfaces to the storage.Provider interface:
// fosite.Storage, fosite.ClientManager, storage.Transactional, oauth2.AuthorizeCodeStorage, oauth2.AccessTokenStorage,
// oauth2.RefreshTokenStorage, oauth2.TokenRevocationStorage, pkce.PKCERequestStorage,
// openid.OpenIDConnectRequestStorage, and partially implements rfc7523.RFC7523KeyStorage.
type OpenIDConnectStore struct {
	provider storage.Provider
	clients  map[string]*Client
}

// Client represents the client internally.
type Client struct {
	ID               string
	Description      string
	Secret           []byte
	SectorIdentifier string
	Public           bool

	Audience      []string
	Scopes        []string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	ResponseModes []fosite.ResponseModeType

	UserinfoSigningAlgorithm string

	Policy authorization.Level

	Consent ClientConsent
}

// NewClientConsent converts the schema.OpenIDConnectClientConsentConfig into a oidc.ClientConsent.
func NewClientConsent(mode string, duration *time.Duration) ClientConsent {
	switch mode {
	case ClientConsentModeImplicit.String():
		return ClientConsent{Mode: ClientConsentModeImplicit}
	case ClientConsentModePreConfigured.String():
		return ClientConsent{Mode: ClientConsentModePreConfigured, Duration: *duration}
	case ClientConsentModeExplicit.String():
		return ClientConsent{Mode: ClientConsentModeExplicit}
	default:
		return ClientConsent{Mode: ClientConsentModeExplicit}
	}
}

// ClientConsent is the consent configuration for a client.
type ClientConsent struct {
	Mode     ClientConsentMode
	Duration time.Duration
}

// String returns the string representation of the ClientConsentMode.
func (c ClientConsent) String() string {
	return c.Mode.String()
}

// ClientConsentMode represents the consent mode for a client.
type ClientConsentMode int

const (
	// ClientConsentModeExplicit means the client does not implicitly assume consent, and does not allow pre-configured
	// consent sessions.
	ClientConsentModeExplicit ClientConsentMode = iota

	// ClientConsentModePreConfigured means the client does not implicitly assume consent, but does allow pre-configured
	// consent sessions.
	ClientConsentModePreConfigured

	// ClientConsentModeImplicit means the client does implicitly assume consent, and does not allow pre-configured
	// consent sessions.
	ClientConsentModeImplicit
)

// String returns the string representation of the ClientConsentMode.
func (c ClientConsentMode) String() string {
	switch c {
	case ClientConsentModeExplicit:
		return explicit
	case ClientConsentModeImplicit:
		return implicit
	case ClientConsentModePreConfigured:
		return preconfigured
	default:
		return ""
	}
}

// KeyManager keeps track of all of the active/inactive rsa keys and provides them to services requiring them.
// It additionally allows us to add keys for the purpose of key rotation in the future.
type KeyManager struct {
	jwk  *JWK
	jwks *jose.JSONWebKeySet
}

// PlainTextHasher implements the fosite.Hasher interface without an actual hashing algo.
type PlainTextHasher struct{}

// ConsentGetResponseBody schema of the response body of the consent GET endpoint.
type ConsentGetResponseBody struct {
	ClientID          string   `json:"client_id"`
	ClientDescription string   `json:"client_description"`
	Scopes            []string `json:"scopes"`
	Audience          []string `json:"audience"`
	PreConfiguration  bool     `json:"pre_configuration"`
}

// ConsentPostRequestBody schema of the request body of the consent POST endpoint.
type ConsentPostRequestBody struct {
	ConsentID    string `json:"id"`
	ClientID     string `json:"client_id"`
	Consent      bool   `json:"consent"`
	PreConfigure bool   `json:"pre_configure"`
}

// ConsentPostResponseBody schema of the response body of the consent POST endpoint.
type ConsentPostResponseBody struct {
	RedirectURI string `json:"redirect_uri"`
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

// OpenIDConnectDiscoveryOptions represents the discovery options specific to OpenID Connect.
type OpenIDConnectDiscoveryOptions struct {
	/*
		RECOMMENDED. URL of the OP's UserInfo Endpoint [OpenID.Core]. This URL MUST use the https scheme and MAY contain
		port, path, and query parameter components.
		See Also:
			OpenID.Core: https://openid.net/specs/openid-connect-core-1_0.html
	*/
	UserinfoEndpoint string `json:"userinfo_endpoint,omitempty"`

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
		OPTIONAL. JSON array containing a list of the JWS [JWS] signing algorithms (alg values) [JWA] supported by the
		UserInfo Endpoint to encode the Claims in a JWT [JWT]. The value none MAY be included.
		See Also:
			JWS: https://datatracker.ietf.org/doc/html/rfc7515
			JWA: https://datatracker.ietf.org/doc/html/rfc7518
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	UserinfoSigningAlgValuesSupported []string `json:"userinfo_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWS signing algorithms (alg values) supported by the OP for Request
		Objects, which are described in Section 6.1 of OpenID Connect Core 1.0 [OpenID.Core]. These algorithms are used
		both when the Request Object is passed by value (using the request parameter) and when it is passed by reference
		(using the request_uri parameter). Servers SHOULD support none and RS256.
	*/
	RequestObjectSigningAlgValuesSupported []string `json:"request_object_signing_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWE encryption algorithms (alg values) supported by the OP for the
		ID Token to encode the Claims in a JWT [JWT].
		See Also:
			JWE: https://datatracker.ietf.org/doc/html/rfc7516
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	IDTokenEncryptionAlgValuesSupported []string `json:"id_token_encryption_alg_values_supported,omitempty"`

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
		OPTIONAL. JSON array containing a list of the JWE encryption algorithms (alg values) supported by the OP for
		Request Objects. These algorithms are used both when the Request Object is passed by value and when it is passed
		by reference.
		See Also:
			JWE: https://datatracker.ietf.org/doc/html/rfc7516
	*/
	RequestObjectEncryptionAlgValuesSupported []string `json:"request_object_encryption_alg_values_supported,omitempty"`

	/*
		OPTIONAL. JSON array containing a list of the JWE encryption algorithms (enc values) supported by the OP for the
		ID Token to encode the Claims in a JWT [JWT].
		See Also:
			JWE: https://datatracker.ietf.org/doc/html/rfc7516
			JWT: https://datatracker.ietf.org/doc/html/rfc7519
	*/
	IDTokenEncryptionEncValuesSupported []string `json:"id_token_encryption_enc_values_supported,omitempty"`

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

// OAuth2WellKnownConfiguration represents the well known discovery document specific to OAuth 2.0.
type OAuth2WellKnownConfiguration struct {
	CommonDiscoveryOptions
	OAuth2DiscoveryOptions
}

// OpenIDConnectWellKnownConfiguration represents the well known discovery document specific to OpenID Connect.
type OpenIDConnectWellKnownConfiguration struct {
	CommonDiscoveryOptions
	OAuth2DiscoveryOptions
	OpenIDConnectDiscoveryOptions
	OpenIDConnectFrontChannelLogoutDiscoveryOptions
	OpenIDConnectBackChannelLogoutDiscoveryOptions
}
