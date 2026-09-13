// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

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
