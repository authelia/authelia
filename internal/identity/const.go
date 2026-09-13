// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

import "errors"

var supportedIDTokenSigningAlgs = []string{"ES256", "ES384", "ES512", "PS256", "PS384", "PS512", "RS256", "RS384", "RS512"}

const (
	defaultIDTokenSigningAlg = "RS256"

	pathWellKnownOpenIDConfiguration = "/.well-known/openid-configuration"

	headerAccept      = "Accept"
	headerContentType = "Content-Type"

	mimeApplicationJSON = "application/json"

	claimIssuer            = "iss"
	claimSubject           = "sub"
	claimAudience          = "aud"
	claimAuthorizedParty   = "azp"
	claimNonce             = "nonce"
	claimAccessTokenHash   = "at_hash"
	claimPreferredUsername = "preferred_username"
	claimName              = "name"
	claimEmail             = "email"
	claimAuthnMethodRefs   = "amr"
)

const (
	// ProviderTypeOpenIDConnect is the type of a provider which is an OpenID Connect 1.0 Provider.
	ProviderTypeOpenIDConnect = "openid_connect"

	// ProviderTypeDiscord is the type of a provider which is Discord.
	ProviderTypeDiscord = "discord"

	// ProviderTypePlex is the type of a provider which is Plex.
	ProviderTypePlex = "plex"

	// ProviderTypeGitHub is the type of a provider which is GitHub.
	ProviderTypeGitHub = "github"

	// ResponseModeQuery is the response mode which delivers the authorization response in the query of a GET request
	// to the redirect URI.
	ResponseModeQuery = "query"

	// ResponseModeFormPost is the response mode which delivers the authorization response in the form body of a POST
	// request to the redirect URI.
	ResponseModeFormPost = "form_post"
)

const providerErrorValueLimit = 256

const (
	parameterClientID     = "client_id"
	parameterClientSecret = "client_secret"
	parameterRequestURI   = "request_uri"

	mimeApplicationXWWWFormURLEncoded = "application/x-www-form-urlencoded"

	pushedAuthorizationResponseLimit = 1024 * 64
)

const (
	discordIssuer                = "https://discord.com"
	discordAuthorizationEndpoint = "https://discord.com/oauth2/authorize"
	discordTokenEndpoint         = "https://discord.com/api/oauth2/token" //nolint:gosec // This is a URL, not a credential.
	discordUserEndpoint          = "https://discord.com/api/v10/users/@me"
)

const (
	githubIssuer                = "https://github.com/login/oauth"
	githubAuthorizationEndpoint = "https://github.com/login/oauth/authorize"
	githubTokenEndpoint         = "https://github.com/login/oauth/access_token" //nolint:gosec // This is a URL, not a credential.
	githubUserEndpoint          = "https://api.github.com/user"
	githubEmailsEndpoint        = "https://api.github.com/user/emails"

	headerGitHubAPIVersion    = "X-GitHub-Api-Version"
	githubAPIVersion          = "2022-11-28"
	mimeApplicationGitHubJSON = "application/vnd.github+json"
)

const (
	plexIssuer                = "https://plex.tv"
	plexProduct               = "Authelia"
	plexAuthorizationEndpoint = "https://app.plex.tv/auth"
	plexPINsEndpoint          = "https://plex.tv/api/v2/pins"
	plexUserEndpoint          = "https://plex.tv/api/v2/user"

	headerPlexProduct          = "X-Plex-Product"
	headerPlexVersion          = "X-Plex-Version"
	headerPlexClientIdentifier = "X-Plex-Client-Identifier"
	headerPlexDeviceName       = "X-Plex-Device-Name"
	headerPlexLanguage         = "X-Plex-Language"
	headerPlexToken            = "X-Plex-Token" //nolint:gosec // This is a header name, not a credential.
)

const (
	headerAuthorization = "Authorization"

	mimeApplicationJWT = "application/jwt"

	userinfoResponseLimit = 1024 * 512
)

var (
	// ErrDiscoveryIssuerMismatch is returned when the discovery document issuer does not match the configured issuer.
	ErrDiscoveryIssuerMismatch = errors.New("the discovery document issuer does not match the configured issuer")

	// ErrDiscoveryEndpointMissing is returned when the discovery document omits a required endpoint.
	ErrDiscoveryEndpointMissing = errors.New("the discovery document is missing a required endpoint")

	// ErrDiscoveryUnsupported is returned when the discovery document advertises capabilities which do not include a
	// value the provider is configured to use.
	ErrDiscoveryUnsupported = errors.New("the discovery document does not advertise support for the configured value")

	// ErrDiscoveryNoSupportedAlg is returned when no ID Token signing algorithm is configured and the discovery document
	// advertises none which Authelia supports.
	ErrDiscoveryNoSupportedAlg = errors.New("the discovery document does not advertise an id token signing algorithm which is supported")

	// ErrPushedAuthorizationRequestEndpointMissing is returned when pushed authorization requests are required but no
	// pushed authorization request endpoint is known for the provider.
	ErrPushedAuthorizationRequestEndpointMissing = errors.New("the pushed authorization request endpoint is not known")

	// ErrPushedAuthorizationResponseInvalid is returned when a successful pushed authorization response could not be
	// used.
	ErrPushedAuthorizationResponseInvalid = errors.New("the pushed authorization response is invalid")

	// ErrTokenSignatureInvalid is returned when the ID Token signature could not be verified.
	ErrTokenSignatureInvalid = errors.New("the id token signature could not be verified")

	// ErrTokenClaimInvalid is returned when an ID Token claim failed validation.
	ErrTokenClaimInvalid = errors.New("the id token contains an invalid claim")

	// ErrTokenNoKey is returned when no key in the JSON Web Key Set matched the ID Token.
	ErrTokenNoKey = errors.New("no key in the json web key set matched the id token")

	// ErrValidationOptionsInvalid is returned when the ID Token validation options themselves are unusable, which
	// indicates a caller error rather than a defective ID Token.
	ErrValidationOptionsInvalid = errors.New("the id token validation options are invalid")

	// ErrUserInfoAccessTokenMissing is returned when the UserInfo Endpoint must be requested but the token response
	// did not contain an access token to request it with.
	ErrUserInfoAccessTokenMissing = errors.New("the token response did not contain an access token")

	// ErrUserInfoResponseInvalid is returned when the UserInfo Response could not be used.
	ErrUserInfoResponseInvalid = errors.New("the userinfo response is invalid")

	// ErrUserInfoSubjectMismatch is returned when the 'sub' claim of the UserInfo Response is not the 'sub' claim of
	// the ID Token.
	ErrUserInfoSubjectMismatch = errors.New("the userinfo response 'sub' claim does not match the id token 'sub' claim")

	// ErrDiscordUserInvalid is returned when the Discord current user response could not be used.
	ErrDiscordUserInvalid = errors.New("the discord user response is invalid")

	// ErrGitHubResponseInvalid is returned when a GitHub response could not be used.
	ErrGitHubResponseInvalid = errors.New("the github response is invalid")

	// ErrPlexResponseInvalid is returned when a Plex response could not be used.
	ErrPlexResponseInvalid = errors.New("the plex response is invalid")

	// ErrPlexPINMismatch is returned when the Plex PIN the browser returned is not the Plex PIN the flow created.
	ErrPlexPINMismatch = errors.New("the plex pin returned does not match the plex pin requested")

	// ErrPlexPINUnauthorized is returned when the Plex PIN has not been approved by the user.
	ErrPlexPINUnauthorized = errors.New("the plex pin has not been authorized")
)
