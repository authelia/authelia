package oidcrp

import "errors"

var (
	// ErrDiscoveryIssuerMismatch is returned when the discovery document issuer does not match the configured issuer.
	ErrDiscoveryIssuerMismatch = errors.New("the discovery document issuer does not match the configured issuer")

	// ErrDiscoveryEndpointMissing is returned when the discovery document omits a required endpoint.
	ErrDiscoveryEndpointMissing = errors.New("the discovery document is missing a required endpoint")

	// ErrTokenSignatureInvalid is returned when the ID Token signature could not be verified.
	ErrTokenSignatureInvalid = errors.New("the id token signature could not be verified")

	// ErrTokenClaimInvalid is returned when an ID Token claim failed validation.
	ErrTokenClaimInvalid = errors.New("the id token contains an invalid claim")

	// ErrTokenNoKey is returned when no key in the JSON Web Key Set matched the ID Token.
	ErrTokenNoKey = errors.New("no key in the json web key set matched the id token")

	// ErrValidationOptionsInvalid is returned when the ID Token validation options themselves are unusable, which
	// indicates a caller error rather than a defective ID Token.
	ErrValidationOptionsInvalid = errors.New("the id token validation options are invalid")
)
