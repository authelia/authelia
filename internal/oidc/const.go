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
)

// Registered Claim strings. See https://www.iana.org/assignments/jwt/jwt.xhtml.
const (
	ClaimJWTID                               = "jti"
	ClaimSessionID                           = "sid"
	ClaimAccessTokenHash                     = "at_hash"
	ClaimCodeHash                            = "c_hash"
	ClaimIssuedAt                            = "iat"
	ClaimNotBefore                           = "nbf"
	ClaimRequestedAt                         = "rat"
	ClaimExpirationTime                      = "exp"
	ClaimAuthenticationTime                  = "auth_time"
	ClaimIssuer                              = "iss"
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
	ClaimClientIdentifier                    = "client_id"
)

const (
	lifespanTokenDefault         = time.Hour
	lifespanRefreshTokenDefault  = time.Hour * 24 * 30
	lifespanAuthorizeCodeDefault = time.Minute * 15
	lifespanPARContextDefault    = time.Minute * 5
)

const (
	urnPARPrefix = "urn:ietf:params:oauth:request_uri:"
)

const (
	// ClaimEmailAlts is an unregistered/custom claim.
	// It represents the emails which are not considered primary.
	ClaimEmailAlts = "alt_emails"
)

// Response Mode strings.
const (
	ResponseModeQuery    = "query"
	ResponseModeFormPost = "form_post"
	ResponseModeFragment = "fragment"
)

// Grant Type strings.
const (
	GrantTypeImplicit          = implicit
	GrantTypeRefreshToken      = "refresh_token"
	GrantTypeAuthorizationCode = "authorization_code"
	GrantTypePassword          = "password"
	GrantTypeClientCredentials = "client_credentials"
)

// Client Auth Method strings.
const (
	ClientAuthMethodClientSecretBasic = "client_secret_basic"
	ClientAuthMethodClientSecretPost  = "client_secret_post"
	ClientAuthMethodClientSecretJWT   = "client_secret_jwt"
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

// Signing Algorithm strings.
const (
	SigningAlgorithmNone          = none
	SigningAlgorithmRSAWithSHA256 = "RS256"
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
	FormParameterRequestURI          = "request_uri"
	FormParameterCodeChallenge       = "code_challenge"
	FormParameterCodeChallengeMethod = "code_challenge_method"
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
)

const (
	tokenPrefixOrgAutheliaFmt = "authelia_%s_" //nolint:gosec
	tokenPrefixOrgOryFmt      = "ory_%s_"      //nolint:gosec

	tokenPrefixPartAccessToken   = "at"
	tokenPrefixPartRefreshToken  = "rt"
	tokenPrefixPartAuthorizeCode = "ac"
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
	// one-time password as per RFC4949. One-time password specifications that this authentication method applies to
	// include RFC4226 and RFC6238.
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
	implicit      = "implicit"
	explicit      = "explicit"
	preconfigured = "pre-configured"
	none          = "none"
)
