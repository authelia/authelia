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
	headerAuthorization = "Authorization"

	mimeApplicationJWT = "application/jwt"

	userinfoResponseLimit = 1024 * 512
)

var (
	// ErrDiscoveryIssuerMismatch is returned when the discovery document issuer does not match the configured issuer.
	ErrDiscoveryIssuerMismatch = errors.New("the discovery document issuer does not match the configured issuer")

	// ErrDiscoveryEndpointMissing is returned when the discovery document omits a required endpoint.
	ErrDiscoveryEndpointMissing = errors.New("the discovery document is missing a required endpoint")

	// ErrDiscoveryURLInsecure is returned when a URL in the discovery document does not use the https scheme.
	ErrDiscoveryURLInsecure = errors.New("the discovery document includes a url which does not use the https scheme")

	// ErrRedirectInsecure is returned when a provider responds with a redirect to a URL which does not use the https
	// scheme.
	ErrRedirectInsecure = errors.New("the provider redirected to a url which does not use the https scheme")

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
)
