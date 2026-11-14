// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	_ "embed" // Required for the embedded Apple App Attestation root.
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

const (
	appleMaxChainLength = 5

	appleEnvironmentProduction  = "production"
	appleEnvironmentDevelopment = "development"

	appleAuthDataFlagAttested   = 0x40
	appleAuthDataFlagExtensions = 0x80

	appleAuthDataHeaderLength     = 37
	appleAuthDataCredentialHeader = 18
)

var (
	oidAppleAppAttestNonce = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 2}

	appleAAGUIDProduction  = []byte("appattest\x00\x00\x00\x00\x00\x00\x00")
	appleAAGUIDDevelopment = []byte("appattestdevelop")
)

//go:embed roots/apple_app_attestation_root_ca.pem
var appleRootPEM []byte

var appleRoot = sync.OnceValue(func() *x509.Certificate {
	block, _ := pem.Decode(appleRootPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		panic("error decoding the apple app attestation root certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("error parsing the apple app attestation root certificate: %w", err))
	}

	return cert
})

var appleDecMode = sync.OnceValue(func() cbor.DecMode {
	dm, err := cbor.DecOptions{
		DupMapKey:   cbor.DupMapKeyEnforcedAPF,
		IndefLength: cbor.IndefLengthForbidden,
		TagsMd:      cbor.TagsForbidden,
	}.DecMode()
	if err != nil {
		panic(err)
	}

	return dm
})

// AppleAppAttestationRoot returns the pinned Apple App Attestation root certificate.
func AppleAppAttestationRoot() *x509.Certificate {
	return appleRoot()
}

type appleAttestationObject struct {
	Format   string                    `cbor:"fmt"`
	AttStmt  appleAttestationStatement `cbor:"attStmt"`
	AuthData []byte                    `cbor:"authData"`
}

type appleAttestationStatement struct {
	X5C     [][]byte `cbor:"x5c"`
	Receipt []byte   `cbor:"receipt"`
}

type appleNonceExtension struct {
	Nonce []byte `asn1:"explicit,tag:1"`
}

type appleAuthData struct {
	rpIDHash     []byte
	flags        byte
	signCount    uint32
	aaguid       []byte
	credentialID []byte
}

func (v *Verifier) verifyAppleAppAttest(_ context.Context, in AttestationInput) (AttestationResult, error) {
	keyID, err := base64.RawURLEncoding.DecodeString(in.KeyID)
	if err != nil {
		return AttestationResult{}, errors.New("key id is not valid base64url")
	}

	if len(keyID) != sha256.Size {
		return AttestationResult{}, errors.New("key id is not 32 bytes")
	}

	object, err := base64.RawURLEncoding.DecodeString(in.AttestationObject)
	if err != nil {
		return AttestationResult{}, errors.New("attestation object is not valid base64url")
	}

	clientDataHash := sha256.Sum256([]byte(in.RegistrationString))

	identity, err := v.verifyAppleAttestationObject(v.clock.Now(), keyID, object, clientDataHash[:])
	if err != nil {
		return AttestationResult{}, err
	}

	return AttestationResult{
		Format:   AttestationFormatAppleAppAttest,
		Identity: identity.TeamID + "." + identity.BundleID,
		Detail:   identity.Environment,
		KeyID:    keyID,
		Attested: true,
	}, nil
}

func (v *Verifier) verifyAppleAttestationObject(now time.Time, keyID, raw, clientDataHash []byte) (identity schema.AppAuthenticatorAppleIdentity, err error) {
	var object appleAttestationObject

	if err = appleDecMode().Unmarshal(raw, &object); err != nil {
		return identity, fmt.Errorf("error decoding attestation object: %w", err)
	}

	if object.Format != AttestationFormatAppleAppAttest {
		return identity, fmt.Errorf("attestation object format '%s' is not '%s'", object.Format, AttestationFormatAppleAppAttest)
	}

	certs, err := decodeAppleChain(object.AttStmt.X5C)
	if err != nil {
		return identity, err
	}

	if err = v.verifyAppleChainTrust(now, certs); err != nil {
		return identity, err
	}

	nonce := sha256.Sum256(slices.Concat(object.AuthData, clientDataHash))

	if err = checkAppleNonce(certs[0], nonce[:]); err != nil {
		return identity, err
	}

	if err = checkAppleCredentialKey(certs[0], keyID); err != nil {
		return identity, err
	}

	authData, err := parseAppleAuthData(object.AuthData)
	if err != nil {
		return identity, err
	}

	return v.matchAppleIdentity(authData, keyID)
}

func decodeAppleChain(x5c [][]byte) ([]*x509.Certificate, error) {
	switch {
	case len(x5c) == 0:
		return nil, errors.New("certificate chain is empty")
	case len(x5c) == 1:
		return nil, errors.New("certificate chain is too short")
	case len(x5c) > appleMaxChainLength:
		return nil, errors.New("certificate chain is too long")
	}

	certs := make([]*x509.Certificate, len(x5c))

	for i, der := range x5c {
		var err error

		if certs[i], err = x509.ParseCertificate(der); err != nil {
			return nil, fmt.Errorf("error parsing certificate %d: %w", i, err)
		}
	}

	return certs, nil
}

func (v *Verifier) verifyAppleChainTrust(now time.Time, certs []*x509.Certificate) error {
	if v.appleRoot == nil {
		return errors.New("no trusted root is configured")
	}

	for i := 0; i < len(certs)-1; i++ {
		if err := certs[i+1].CheckSignature(certs[i].SignatureAlgorithm, certs[i].RawTBSCertificate, certs[i].Signature); err != nil {
			return fmt.Errorf("certificate %d signature is not valid: %w", i, err)
		}
	}

	last := len(certs) - 1

	if err := v.appleRoot.CheckSignature(certs[last].SignatureAlgorithm, certs[last].RawTBSCertificate, certs[last].Signature); err != nil {
		return fmt.Errorf("certificate %d is not signed by the trusted root", last)
	}

	for i, cert := range certs {
		if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
			return fmt.Errorf("certificate %d is not valid at the current time", i)
		}
	}

	for i, cert := range certs[1:] {
		if !cert.BasicConstraintsValid || !cert.IsCA || cert.KeyUsage&x509.KeyUsageCertSign == 0 {
			return fmt.Errorf("certificate %d is not a certificate authority", i+1)
		}
	}

	return nil
}

func checkAppleNonce(cert *x509.Certificate, nonce []byte) error {
	ext := findExtension(cert, oidAppleAppAttestNonce)
	if ext == nil {
		return errors.New("credential certificate does not contain the nonce extension")
	}

	var value appleNonceExtension

	if err := unmarshalStrict(ext, &value, ""); err != nil {
		return fmt.Errorf("error parsing nonce extension: %w", err)
	}

	if canonical, err := asn1.Marshal(value); err != nil || !bytes.Equal(canonical, ext) {
		return errors.New("error parsing nonce extension: extension is not a sequence of exactly one nonce")
	}

	if !bytes.Equal(value.Nonce, nonce) {
		return errors.New("nonce does not match")
	}

	return nil
}

func checkAppleCredentialKey(cert *x509.Certificate, keyID []byte) error {
	key, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok || key.Curve != elliptic.P256() {
		return errors.New("credential certificate public key is not a P-256 key")
	}

	point, err := key.ECDH()
	if err != nil {
		return fmt.Errorf("error encoding credential certificate public key: %w", err)
	}

	sum := sha256.Sum256(point.Bytes())

	if !bytes.Equal(sum[:], keyID) {
		return errors.New("credential certificate public key does not match the key id")
	}

	return nil
}

func parseAppleAuthData(data []byte) (authData appleAuthData, err error) {
	if len(data) < appleAuthDataHeaderLength {
		return authData, errors.New("authenticator data is too short")
	}

	authData.rpIDHash = data[:32]
	authData.flags = data[32]
	authData.signCount = binary.BigEndian.Uint32(data[33:37])

	if authData.flags&appleAuthDataFlagAttested == 0 {
		return authData, errors.New("authenticator data does not contain attested credential data")
	}

	rest := data[appleAuthDataHeaderLength:]

	if len(rest) < appleAuthDataCredentialHeader {
		return authData, errors.New("authenticator data is too short")
	}

	authData.aaguid = rest[:16]

	n := int(binary.BigEndian.Uint16(rest[16:18]))

	rest = rest[appleAuthDataCredentialHeader:]

	if len(rest) < n {
		return authData, errors.New("authenticator data is too short")
	}

	authData.credentialID, rest = rest[:n], rest[n:]

	var key map[int]any

	if rest, err = appleDecMode().UnmarshalFirst(rest, &key); err != nil {
		return authData, fmt.Errorf("error parsing credential public key: %w", err)
	}

	if authData.flags&appleAuthDataFlagExtensions != 0 {
		var extensions map[string]any

		if rest, err = appleDecMode().UnmarshalFirst(rest, &extensions); err != nil {
			return authData, fmt.Errorf("error parsing authenticator data extensions: %w", err)
		}
	}

	if len(rest) != 0 {
		return authData, errors.New("authenticator data contains trailing data")
	}

	return authData, nil
}

func (v *Verifier) matchAppleIdentity(authData appleAuthData, keyID []byte) (identity schema.AppAuthenticatorAppleIdentity, err error) {
	var candidates []schema.AppAuthenticatorAppleIdentity

	for _, candidate := range v.config.Apple {
		sum := sha256.Sum256([]byte(candidate.TeamID + "." + candidate.BundleID))

		if bytes.Equal(sum[:], authData.rpIDHash) {
			candidates = append(candidates, candidate)
		}
	}

	if len(candidates) == 0 {
		return identity, errors.New("relying party id hash does not match any configured apple identity")
	}

	if authData.signCount != 0 {
		return identity, fmt.Errorf("sign count %d is not zero", authData.signCount)
	}

	i := slices.IndexFunc(candidates, func(candidate schema.AppAuthenticatorAppleIdentity) bool {
		return bytes.Equal(appleAAGUID(candidate.Environment), authData.aaguid)
	})

	if i == -1 {
		return identity, errors.New("aaguid does not match the environment of any matching apple identity")
	}

	if !bytes.Equal(authData.credentialID, keyID) {
		return identity, errors.New("credential id does not match the key id")
	}

	return candidates[i], nil
}

func appleAAGUID(environment string) []byte {
	switch environment {
	case appleEnvironmentProduction:
		return appleAAGUIDProduction
	case appleEnvironmentDevelopment:
		return appleAAGUIDDevelopment
	default:
		return nil
	}
}
