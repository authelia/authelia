package oidc

import (
	"time"
)

// Scope strings.
const (
	ScopeOfflineAccess = "offline_access"
	ScopeOffline       = "offline"
	ScopeOpenID        = "openid"
	ScopeProfile       = "profile"
	ScopeEmail         = "email"
	ScopeGroups        = "groups"

	ScopeAutheliaBearerAuthz = "authelia.bearer.authz"
)

// Registered Claim strings. See https://www.iana.org/assignments/jwt/jwt.xhtml.
const (
	ClaimJWTID                               = "jti"
	ClaimSessionID                           = "sid"
	ClaimAccessTokenHash                     = "at_hash"
	ClaimCodeHash                            = "c_hash"
	ClaimStateHash                           = "s_hash"
	ClaimIssuedAt                            = "iat"
	ClaimNotBefore                           = "nbf"
	ClaimRequestedAt                         = "rat"
	ClaimExpirationTime                      = "exp"
	ClaimAuthenticationTime                  = "auth_time"
	ClaimIssuer                              = valueIss
	ClaimSubject                             = "sub"
	ClaimNonce                               = "nonce"
	ClaimAudience                            = "aud"
	ClaimGroups                              = "groups"
	ClaimFullName                            = "name"
	ClaimPreferredUsername                   = "preferred_username"
	ClaimPreferredEmail                      = "email"
	ClaimEmailVerified                       = "email_verified"
	ClaimAuthorizedParty                     = "azp"
	ClaimAuthenticationContextClassReference = "acr"
	ClaimAuthenticationMethodsReference      = "amr"
	ClaimClientIdentifier                    = valueClientID
	ClaimScope                               = valueScope
	ClaimScopeNonStandard                    = "scp"
	ClaimExtra                               = "ext"
	ClaimActive                              = "active"
	ClaimUsername                            = "username"
	ClaimTokenIntrospection                  = "token_introspection"
)

const (
	lifespanTokenDefault                      = time.Hour
	lifespanRefreshTokenDefault               = time.Hour * 24 * 30
	lifespanAuthorizeCodeDefault              = time.Minute * 15
	lifespanJWTSecuredAuthorizationDefault    = time.Minute * 5
	lifespanPARContextDefault                 = time.Minute * 5
	lifespanRFC8628CodeDefault                = time.Minute * 10
	lifespanRFC8628PollingIntervalDefault     = time.Second * 10
	lifespanVerifiableCredentialsNonceDefault = time.Hour
)

const (
	RedirectURIPrefixPushedAuthorizationRequestURN = "urn:ietf:params:oauth:request_uri:"
)

const (
	// ClaimEmailAlts is an unregistered/custom claim.
	// It represents the emails which are not considered primary.
	ClaimEmailAlts = "alt_emails"
)

// Response Mode strings.
const (
	ResponseModeFormPost    = "form_post"
	ResponseModeQuery       = "query"
	ResponseModeFragment    = "fragment"
	ResponseModeJWT         = "jwt"
	ResponseModeFormPostJWT = "form_post.jwt"
	ResponseModeQueryJWT    = "query.jwt"
	ResponseModeFragmentJWT = "fragment.jwt"
)

// Grant Type strings.
const (
	GrantTypeImplicit          = valueImplicit
	GrantTypeRefreshToken      = valueRefreshToken
	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypeClientCredentials = "client_credentials"
)

// Client Auth Method strings.
const (
	ClientAuthMethodClientSecretBasic = "client_secret_basic"
	ClientAuthMethodClientSecretPost  = "client_secret_post"
	ClientAuthMethodClientSecretJWT   = "client_secret_jwt"
	ClientAuthMethodPrivateKeyJWT     = "private_key_jwt"
	ClientAuthMethodNone              = "none"
)

// Response Type strings.
const (
	ResponseTypeAuthorizationCodeFlow = "code"
	ResponseTypeImplicitFlowIDToken   = "id_token"
	ResponseTypeImplicitFlowToken     = "token"
	ResponseTypeImplicitFlowBoth      = "id_token token"
	ResponseTypeHybridFlowIDToken     = "code id_token"
	ResponseTypeHybridFlowToken       = "code token"
	ResponseTypeHybridFlowBoth        = "code id_token token"
)

// JWS Algorithm strings.
// See: https://datatracker.ietf.org/doc/html/rfc7518#section-3.1
const (
	SigningAlgNone = valueNone

	SigningAlgRSAUsingSHA256 = "RS256"
	SigningAlgRSAUsingSHA384 = "RS384"
	SigningAlgRSAUsingSHA512 = "RS512"

	SigningAlgRSAPSSUsingSHA256 = "PS256"
	SigningAlgRSAPSSUsingSHA384 = "PS384"
	SigningAlgRSAPSSUsingSHA512 = "PS512"

	SigningAlgECDSAUsingP256AndSHA256 = "ES256"
	SigningAlgECDSAUsingP384AndSHA384 = "ES384"
	SigningAlgECDSAUsingP521AndSHA512 = "ES512"

	SigningAlgHMACUsingSHA256 = "HS256"
	SigningAlgHMACUsingSHA384 = "HS384"
	SigningAlgHMACUsingSHA512 = "HS512"
)

// JWS Algorithm Prefixes.
const (
	SigningAlgPrefixRSA    = "RS"
	SigningAlgPrefixHMAC   = "HS"
	SigningAlgPrefixRSAPSS = "PS"
	SigningAlgPrefixECDSA  = "ES"
)

const (
	KeyUseSignature = "sig"
)

// Subject Type strings.
const (
	SubjectTypePublic   = "public"
	SubjectTypePairwise = "pairwise"
)

// Proof Key Code Exchange Challenge Method strings.
const (
	PKCEChallengeMethodPlain  = "plain"
	PKCEChallengeMethodSHA256 = "S256"
)

const (
	FormParameterState        = "state"
	FormParameterClientID     = valueClientID
	FormParameterRequestURI   = "request_uri"
	FormParameterRedirectURI  = "redirect_uri"
	FormParameterResponseMode = "response_mode"
	FormParameterResponseType = "response_type"
	FormParameterScope        = valueScope
	FormParameterIssuer       = valueIss
	FormParameterPrompt       = "prompt"
)

const (
	PromptNone    = valueNone
	PromptLogin   = "login"
	PromptConsent = "consent"
	// PromptCreate  = "create" // This prompt value is currently unused.
)

// Endpoints.
const (
	EndpointAuthorization              = "authorization"
	EndpointToken                      = "token"
	EndpointUserinfo                   = "userinfo"
	EndpointIntrospection              = "introspection"
	EndpointRevocation                 = "revocation"
	EndpointPushedAuthorizationRequest = "pushed-authorization-request"
)

// JWT Headers.
const (
	// JWTHeaderKeyIdentifier is the JWT Header referencing the JWS Key Identifier used to sign a token.
	JWTHeaderKeyIdentifier = "kid"

	// JWTHeaderKeyAlgorithm is the JWT Header referencing the JWS Key algorithm used to sign a token.
	JWTHeaderKeyAlgorithm = "alg"

	// JWTHeaderKeyType is the JWT Header referencing the JWT type.
	JWTHeaderKeyType = "typ"
)

const (
	JWTHeaderTypeValueTokenIntrospectionJWT = "token-introspection+jwt"
	JWTHeaderTypeValueAccessTokenJWT        = "at+jwt"
)

// Paths.
const (
	EndpointPathConsent                           = "/consent"
	EndpointPathWellKnownOpenIDConfiguration      = "/.well-known/openid-configuration"
	EndpointPathWellKnownOAuthAuthorizationServer = "/.well-known/oauth-authorization-server"
	EndpointPathJWKs                              = "/jwks.json"

	EndpointPathRoot = "/api/oidc"

	EndpointPathAuthorization = EndpointPathRoot + "/" + EndpointAuthorization
	EndpointPathToken         = EndpointPathRoot + "/" + EndpointToken
	EndpointPathUserinfo      = EndpointPathRoot + "/" + EndpointUserinfo
	EndpointPathIntrospection = EndpointPathRoot + "/" + EndpointIntrospection
	EndpointPathRevocation    = EndpointPathRoot + "/" + EndpointRevocation

	EndpointPathPushedAuthorizationRequest = EndpointPathRoot + "/" + EndpointPushedAuthorizationRequest

	EndpointPathRFC8628UserVerificationURL = EndpointPathRoot + "/device-code/user-verification"
)

// Authentication Method Reference Values https://datatracker.ietf.org/doc/html/rfc8176
const (
	// AMRMultiFactorAuthentication is an RFC8176 Authentication Method Reference Value that represents multiple-factor
	// authentication as per NIST.800-63-2 and ISO29115. When this is present, specific authentication methods used may
	// also be included.
	//
	// Authelia utilizes this when a user has performed any 2 AMR's with different factor values (excluding meta).
	// Factor: Meta, Channel: Meta.
	//
	// RFC8176: https://datatracker.ietf.org/doc/html/rfc8176
	//
	// NIST.800-63-2: http://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-63-2.pdf
	//
	// ISO29115: https://www.iso.org/standard/45138.html
	AMRMultiFactorAuthentication = "mfa"

	// AMRMultiChannelAuthentication is an RFC8176 Authentication Method Reference Value that represents
	// multiple-channel authentication. The authentication involves communication over more than one distinct
	// communication channel. For instance, a multiple-channel authentication might involve both entering information
	// into a workstation's browser and providing information on a telephone call to a pre-registered number.
	//
	// Authelia utilizes this when a user has performed any 2 AMR's with different channel values (excluding meta).
	// Factor: Meta, Channel: Meta.
	//
	// RFC8176: https://datatracker.ietf.org/doc/html/rfc8176
	AMRMultiChannelAuthentication = "mca"

	// AMRUserPresence is an RFC8176 Authentication Method Reference Value that represents authentication that included
	// a user presence test. Evidence that the end user is present and interacting with the device. This is sometimes
	// also referred to as "test of user presence" as per W3C.WD-webauthn-20170216.
	//
	// Authelia utilizes this when a user has used WebAuthn to authenticate and the user presence flag was set.
	// Factor: Meta, Channel: Meta.
	//
	// RFC8176: https://datatracker.ietf.org/doc/html/rfc8176
	//
	// W3C.WD-webauthn-20170216: https://datatracker.ietf.org/doc/html/rfc8176#ref-W3C.WD-webauthn-20170216
	AMRUserPresence = "user"

	// AMRPersonalIdentificationNumber is an RFC8176 Authentication Method Reference Value that represents
	// authentication that included a personal Identification Number (PIN) as per RFC4949 or pattern (not restricted to
	// containing only numbers) that a user enters to unlock a key on the device. This mechanism should have a way to
	// deter an attacker from obtaining the PIN by trying repeated guesses.
	//
	// Authelia utilizes this when a user has used WebAuthn to authenticate and the user verified flag was set.
	//	Factor: Meta, Channel: Meta.
	//
	// RFC8176: https://datatracker.ietf.org/doc/html/rfc8176
	//
	// RFC4949: https://datatracker.ietf.org/doc/html/rfc4949
	AMRPersonalIdentificationNumber = "pin"

	// AMRPasswordBasedAuthentication is an RFC8176 Authentication Method Reference Value that represents password-based
	// authentication as per RFC4949.
	//
	// Authelia utilizes this when a user has performed 1FA. Factor: Know, Channel: Browser.
	//
	// RFC8176: https://datatracker.ietf.org/doc/html/rfc8176
	//
	// RFC4949: https://datatracker.ietf.org/doc/html/rfc4949
	AMRPasswordBasedAuthentication = "pwd"

	// AMROneTimePassword is an RFC8176 Authentication Method Reference Value that represents authentication via a
	// Time-based One-Time Password as per RFC4949. One-time password specifications that this authentication method
	// applies to include RFC4226 and RFC6238.
	//
	// Authelia utilizes this when a user has used TOTP to authenticate. Factor: Have, Channel: Browser.
	//
	// RFC8176: https://datatracker.ietf.org/doc/html/rfc8176
	//
	// RFC4949: https://datatracker.ietf.org/doc/html/rfc4949
	//
	// RFC4226: https://datatracker.ietf.org/doc/html/rfc4226
	//
	// RFC6238: https://datatracker.ietf.org/doc/html/rfc6238
	AMROneTimePassword = "otp"

	// AMRHardwareSecuredKey is an RFC8176 Authentication Method Reference Value that
	// represents authentication via a proof-of-Possession (PoP) of a hardware-secured key.
	//
	// Authelia utilizes this when a user has used WebAuthn to authenticate. Factor: Have, Channel: Browser.
	//
	// RFC8176: https://datatracker.ietf.org/doc/html/rfc8176
	AMRHardwareSecuredKey = "hwk"

	// AMRShortMessageService is an RFC8176 Authentication Method Reference Value that
	// represents authentication via confirmation using SMS text message to the user at a registered number.
	//
	// Authelia utilizes this when a user has used Duo to authenticate. Factor: Have, Channel: Browser.
	//
	// RFC8176: https://datatracker.ietf.org/doc/html/rfc8176
	AMRShortMessageService = "sms"
)

const (
	valueScope         = "scope"
	valueClientID      = "client_id"
	valueImplicit      = "implicit"
	valueExplicit      = "explicit"
	valuePreconfigured = "pre-configured"
	valueNone          = "none"
	valueRefreshToken  = "refresh_token"
	valueIss           = "iss"
)

const (
	durationZero = time.Duration(0)
)

const (
	fieldRFC6750Error            = "error"
	fieldRFC6750ErrorDescription = "error_description"
	fieldRFC6750Realm            = "realm"
	fieldRFC6750Scope            = valueScope
)
