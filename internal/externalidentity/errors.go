// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import "errors"

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
