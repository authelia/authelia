// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
)

// ParsePublicKey parses a SPKI DER encoded P-256 public key.
func ParsePublicKey(spki []byte) (*ecdsa.PublicKey, error) {
	key, err := x509.ParsePKIXPublicKey(spki)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}

	pub, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not an ECDSA key")
	}

	if pub.Curve != elliptic.P256() {
		return nil, errors.New("public key is not a P-256 key")
	}

	return pub, nil
}

// Thumbprint returns the base64url SHA-256 of the SPKI bytes.
func Thumbprint(spki []byte) string {
	h := sha256.Sum256(spki)

	return base64.RawURLEncoding.EncodeToString(h[:])
}

// VerifySignature verifies a base64url DER ECDSA signature over the SHA-256 of the message. Any decoding error returns false.
func VerifySignature(pub *ecdsa.PublicKey, message string, signature string) bool {
	if pub == nil {
		return false
	}

	der, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	h := sha256.Sum256([]byte(message))

	return ecdsa.VerifyASN1(pub, h[:], der)
}
