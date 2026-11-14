// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import "time"

const (
	// PurposeRegister is the label prefixing the registration signed string.
	PurposeRegister = "authelia-app-v1:register"

	// PurposeApprove is the label prefixing the approval signed string.
	PurposeApprove = "authelia-app-v1:approve"

	// PurposeRequest is the label prefixing the request signed string.
	PurposeRequest = "authelia-app-v1:request"

	// TimestampSkew is the maximum accepted difference between server and request time.
	TimestampSkew = 60 * time.Second

	// ReplayRetention is how long request nonces are retained.
	ReplayRetention = 120 * time.Second

	// MaxBodySize is the maximum accepted request body size in bytes.
	MaxBodySize = 64 * 1024

	// AuthorizationScheme is the Authorization header scheme.
	AuthorizationScheme = "AutheliaApp"

	// EmptyBodyHash is the base64url SHA-256 of zero bytes.
	EmptyBodyHash = "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU"
)

const (
	separator = "::"

	qrScheme = "authelia-app://"
	qrPath   = "/api/register/app/"
)
