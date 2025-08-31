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
	ScopePhone         = "phone"
	ScopeAddress       = "address"
	ScopeGroups        = "groups"

	ScopeAutheliaBearerAuthz = "authelia.bearer.authz"
)

const (
	fmtAutheliaOpaqueOAuth2Token = "authelia_%s_" //nolint:gosec
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
	ClaimNonce                               = valueNonce
	ClaimAudience                            = "aud"
	ClaimGroups                              = "groups"
	ClaimAuthorizedParty                     = "azp"
	ClaimAuthenticationContextClassReference = "acr"
	ClaimAuthenticationMethodsReference      = "amr"
	ClaimClientIdentifier                    = valueClientID
	ClaimScope                               = valueScope
	ClaimScopeNonStandard                    = "scp"
	ClaimExtra                               = "ext"
	ClaimSubject                             = "sub"
	ClaimFullName                            = "name"
	ClaimGivenName                           = "given_name"
	ClaimFamilyName                          = "family_name"
	ClaimMiddleName                          = "middle_name"
	ClaimNickname                            = "nickname"
	ClaimPreferredUsername                   = "preferred_username"
	ClaimProfile                             = "profile"
	ClaimPicture                             = "picture"
	ClaimWebsite                             = "website"
	ClaimEmail                               = "email"
	ClaimEmailVerified                       = "email_verified"
	ClaimGender                              = "gender"
	ClaimBirthdate                           = "birthdate"
	ClaimZoneinfo                            = "zoneinfo"
	ClaimLocale                              = "locale"
	ClaimPhoneNumber                         = "phone_number"
	ClaimPhoneNumberVerified                 = "phone_number_verified"
	ClaimAddress                             = "address"
	ClaimUpdatedAt                           = "updated_at"
	ClaimActive                              = "active"
	ClaimUsername                            = "username"
	ClaimTokenIntrospection                  = "token_introspection"
)

const (
	ClaimTypeNormal = "normal"
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

// Response Modes strings.
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
	GrantTypeDeviceCode        = "urn:ietf:params:oauth:grant-type:device_code"
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

const (
	EncryptionAlgNone             = "none"
	EncryptionAlgRSA15            = "RSA1_5"
	EncryptionAlgRSAOAEP          = "RSA-OAEP"
	EncryptionAlgRSAOAEP256       = "RSA-OAEP-256"
	EncryptionAlgA128KW           = "A128KW"
	EncryptionAlgA192KW           = "A192KW"
	EncryptionAlgA256KW           = "A256KW"
	EncryptionAlgDirect           = "dir"
	EncryptionAlgECDHES           = "ECDH-ES"
	EncryptionAlgECDHESA128KW     = "ECDH-ES+A128KW"
	EncryptionAlgECDHESA192KW     = "ECDH-ES+A192KW"
	EncryptionAlgECDHESA256KW     = "ECDH-ES+A256KW"
	EncryptionAlgA128GCMKW        = "A128GCMKW"
	EncryptionAlgA192GCMKW        = "A192GCMKW"
	EncryptionAlgA256GCMKW        = "A256GCMKW"
	EncryptionAlgPBES2HS256A128KW = "PBES2-HS256+A128KW"
	EncryptionAlgPBES2HS284A192KW = "PBES2-HS384+A192KW"
	EncryptionAlgPBES2HS512A256KW = "PBES2-HS512+A256KW"
)

const (
	EncryptionEncA128CBCHS256 = "A128CBC-HS256"
	EncryptionEncA192CBCHS384 = "A192CBC-HS384"
	EncryptionEncA256CBCHS512 = "A256CBC-HS512"
	EncryptionEncA128GCM      = "A128GCM"
	EncryptionEncA192GCM      = "A192GCM"
	EncryptionEncA256GCM      = "A256GCM"
)

// JWS Algorithm Prefixes.
const (
	SigningAlgPrefixRSA    = "RS"
	SigningAlgPrefixHMAC   = "HS"
	SigningAlgPrefixRSAPSS = "PS"
	SigningAlgPrefixECDSA  = "ES"
)

const (
	KeyUseSignature  = "sig"
	KeyUseEncryption = "enc"
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
	RedirectURISpecialOAuth2InstalledApp = "urn:ietf:wg:oauth:2.0:oob"
)

const (
	FormParameterState        = "state"
	FormParameterClientID     = valueClientID
	FormParameterRequestURI   = "request_uri"
	FormParameterRedirectURI  = "redirect_uri"
	FormParameterResponseMode = "response_mode"
	FormParameterResponseType = "response_type"
	FormParameterScope        = valueScope
	FormParameterPrompt       = "prompt"
	FormParameterMaximumAge   = "max_age"
	FormParameterClaims       = "claims"
	FormParameterUserCode     = "user_code"
	FormParameterFlowID       = "flow_id"
	FormParameterNonce        = valueNonce
)

const (
	PromptConsent       = "consent"
	PromptLogin         = "login"
	PromptNone          = valueNone
	PromptSelectAccount = "select_account"
	// PromptCreate  = "create" // This prompt value is currently unused.
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
	JWTHeaderTypeValueAccessTokenJWT = "at+jwt"
)

const (
	IDTokenAudienceModeSpecification      = "specification"
	IDTokenAudienceModeExperimentalMerged = "experimental-merged"
)

// Endpoints.
const (
	EndpointConsent                    = "consent"
	EndpointAuthorization              = "authorization"
	EndpointDeviceAuthorization        = "device-authorization"
	EndpointToken                      = "token"
	EndpointUserinfo                   = "userinfo"
	EndpointIntrospection              = "introspection"
	EndpointRevocation                 = "revocation"
	EndpointPushedAuthorizationRequest = "pushed-authorization-request"
)

// Paths.
const (
	FrontendEndpointPathConsentCompletion = "/consent/completion"

	FrontendEndpointPathConsent                    = "/consent/openid"
	FrontendEndpointPathConsentDecision            = FrontendEndpointPathConsent + "/decision"
	FrontendEndpointPathConsentDeviceAuthorization = FrontendEndpointPathConsent + "/" + EndpointDeviceAuthorization

	EndpointPathWellKnownOpenIDConfiguration      = "/.well-known/openid-configuration"
	EndpointPathWellKnownOAuthAuthorizationServer = "/.well-known/oauth-authorization-server"
	EndpointPathJWKs                              = "/jwks.json"

	EndpointPathRoot = "/api/oidc"

	EndpointPathConsent                    = EndpointPathRoot + "/" + EndpointConsent
	EndpointPathAuthorization              = EndpointPathRoot + "/" + EndpointAuthorization
	EndpointPathToken                      = EndpointPathRoot + "/" + EndpointToken
	EndpointPathUserinfo                   = EndpointPathRoot + "/" + EndpointUserinfo
	EndpointPathIntrospection              = EndpointPathRoot + "/" + EndpointIntrospection
	EndpointPathRevocation                 = EndpointPathRoot + "/" + EndpointRevocation
	EndpointPathDeviceAuthorization        = EndpointPathRoot + "/" + EndpointDeviceAuthorization
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

	// AMRProofOfPossession is an Authentication Method Reference Value that
	// represents authentication via a proof-of-Possession (PoP) of a software-secured (swk) or hardware-secured (hwk)
	// key.
	AMRProofOfPossession = "pop"

	// AMRHardwareSecuredKey is an RFC8176 Authentication Method Reference Value that
	// represents authentication via a proof-of-Possession (PoP) of a hardware-secured key.
	//
	// Authelia utilizes this when a user has used WebAuthn to authenticate. Factor: Have, Channel: Browser.
	//
	// RFC8176: https://datatracker.ietf.org/doc/html/rfc8176
	AMRHardwareSecuredKey = "hwk"

	// AMRSoftwareSecuredKey is an RFC8176 Authentication Method Reference Value that
	// represents authentication via a proof-of-Possession (PoP) of a software-secured key.
	//
	// Authelia utilizes this when a user has used WebAuthn to authenticate. Factor: Have, Channel: Browser.
	//
	// RFC8176: https://datatracker.ietf.org/doc/html/rfc8176
	AMRSoftwareSecuredKey = "swk"

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
	valueNonce         = "nonce"
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
