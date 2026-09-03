package oidcrp

const (
	pathWellKnownOpenIDConfiguration = "/.well-known/openid-configuration"

	headerAccept      = "Accept"
	headerContentType = "Content-Type"

	mimeApplicationJSON = "application/json"

	claimIssuer            = "iss"
	claimSubject           = "sub"
	claimAudience          = "aud"
	claimAuthorizedParty   = "azp"
	claimNonce             = "nonce"
	claimPreferredUsername = "preferred_username"
	claimName              = "name"
	claimEmail             = "email"
	claimAuthnMethodRefs   = "amr"
)
