// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// EncodeURLField encodes a field which could contain the separator as base64url.
func EncodeURLField(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

// CodeChallengeS256 returns the S256 code challenge of a verifier.
func CodeChallengeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))

	return base64.RawURLEncoding.EncodeToString(h[:])
}

// BodyHash returns the base64url SHA-256 of a request body.
func BodyHash(body []byte) string {
	h := sha256.Sum256(body)

	return base64.RawURLEncoding.EncodeToString(h[:])
}

// RegistrationString builds the registration signed string.
func RegistrationString(baseURL, challenge, number, verifier, approvalThumbprint, transportThumbprint string) string {
	return strings.Join([]string{PurposeRegister, EncodeURLField(baseURL), challenge, number, verifier, approvalThumbprint, transportThumbprint}, separator)
}

// ApprovalString builds the approval signed string.
func ApprovalString(baseURL, deviceID, challenge, number, verifier string) string {
	return strings.Join([]string{PurposeApprove, EncodeURLField(baseURL), deviceID, challenge, number, verifier}, separator)
}

// RequestString builds the request signed string.
func RequestString(baseURL, deviceID, method, requestURL, ts, nonce, bodyHash string) string {
	return strings.Join([]string{PurposeRequest, EncodeURLField(baseURL), deviceID, method, EncodeURLField(requestURL), ts, nonce, bodyHash}, separator)
}
