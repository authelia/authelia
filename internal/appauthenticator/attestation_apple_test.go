// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

const (
	testAppleTeamID   = "ABCDE12345"
	testAppleBundleID = "com.authelia.mobile"
	testAppleRegistry = "registration string"
)

type testAppleOpts struct {
	registrationString   string
	rpID                 string
	flags                byte
	signCount            uint32
	aaguid               []byte
	credentialID         []byte
	extensions           []byte
	trailing             []byte
	certificateKey       *ecdsa.PrivateKey
	omitNonce            bool
	nonceExtension       []byte
	format               string
	notAfter             time.Time
	intermediateNotAfter time.Time
	intermediateNotCA    bool
	intermediateSigner   *ecdsa.PrivateKey
	leafSigner           *ecdsa.PrivateKey
	x5c                  func(x5c [][]byte) [][]byte
}

type testAppleEnv struct {
	now     time.Time
	clock   *clock.Fixed
	rootKey *ecdsa.PrivateKey
	root    *x509.Certificate
	key     *ecdsa.PrivateKey
	config  schema.AppAuthenticatorAttestation
}

func newTestAppleEnv(t *testing.T) *testAppleEnv {
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)

	rootKey := testGenerateKey(t)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test App Attestation Root CA"},
		NotBefore:             now.Add(-24 * time.Hour),
		NotAfter:              now.Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &rootKey.PublicKey, rootKey)
	require.NoError(t, err)

	root, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return &testAppleEnv{
		now:     now,
		clock:   clock.NewFixed(now),
		rootKey: rootKey,
		root:    root,
		key:     testGenerateKey(t),
		config: schema.AppAuthenticatorAttestation{
			Mode: schema.AppAuthenticatorAttestationModeRequired,
			Apple: []schema.AppAuthenticatorAppleIdentity{
				{TeamID: testAppleTeamID, BundleID: testAppleBundleID, Environment: "production"},
			},
		},
	}
}

func (e *testAppleEnv) verifier() *Verifier {
	return NewVerifier(e.config, WithClock(e.clock), WithAppleRoot(e.root))
}

func (e *testAppleEnv) keyID(t *testing.T) []byte {
	point, err := e.key.PublicKey.ECDH()
	require.NoError(t, err)

	sum := sha256.Sum256(point.Bytes())

	return sum[:]
}

func (e *testAppleEnv) defaultOpts() *testAppleOpts {
	return &testAppleOpts{
		registrationString:   testAppleRegistry,
		rpID:                 testAppleTeamID + "." + testAppleBundleID,
		flags:                0x40,
		aaguid:               []byte("appattest\x00\x00\x00\x00\x00\x00\x00"),
		format:               AttestationFormatAppleAppAttest,
		notAfter:             e.now.Add(72 * time.Hour),
		intermediateNotAfter: e.now.Add(5 * 365 * 24 * time.Hour),
	}
}

func (e *testAppleEnv) input(t *testing.T, opts func(o *testAppleOpts)) AttestationInput {
	o := e.defaultOpts()

	if opts != nil {
		opts(o)
	}

	keyID := e.keyID(t)

	return AttestationInput{
		Format:             AttestationFormatAppleAppAttest,
		KeyID:              base64.RawURLEncoding.EncodeToString(keyID),
		AttestationObject:  base64.RawURLEncoding.EncodeToString(e.object(t, o, keyID)),
		RegistrationString: testAppleRegistry,
	}
}

func (e *testAppleEnv) authData(t *testing.T, o *testAppleOpts, keyID []byte) []byte {
	rpIDHash := sha256.Sum256([]byte(o.rpID))

	credentialID := o.credentialID
	if credentialID == nil {
		credentialID = keyID
	}

	point, err := e.key.PublicKey.ECDH()
	require.NoError(t, err)

	credentialKey, err := cbor.Marshal(map[int]any{1: 2, 3: -7, -1: 1, -2: point.Bytes()[1:33], -3: point.Bytes()[33:]})
	require.NoError(t, err)

	data := append([]byte{}, rpIDHash[:]...)
	data = append(data, o.flags)
	data = binary.BigEndian.AppendUint32(data, o.signCount)
	data = append(data, o.aaguid...)
	data = append(data, new(big.Int).SetInt64(int64(len(credentialID))).FillBytes(make([]byte, 2))...)
	data = append(data, credentialID...)
	data = append(data, credentialKey...)
	data = append(data, o.extensions...)

	return append(data, o.trailing...)
}

func (e *testAppleEnv) chain(t *testing.T, o *testAppleOpts, nonce []byte) [][]byte {
	intKey := testGenerateKey(t)

	intTmpl := &x509.Certificate{
		SerialNumber:          testSerial(t),
		Subject:               pkix.Name{CommonName: "Test App Attestation CA 1"},
		NotBefore:             e.now.Add(-48 * time.Hour),
		NotAfter:              o.intermediateNotAfter,
		IsCA:                  !o.intermediateNotCA,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}

	intParent, intSigner := e.root, e.rootKey

	if o.intermediateSigner != nil {
		intParent, intSigner = &x509.Certificate{Subject: e.root.Subject, PublicKey: &o.intermediateSigner.PublicKey}, o.intermediateSigner
	}

	intDER, err := x509.CreateCertificate(rand.Reader, intTmpl, intParent, &intKey.PublicKey, intSigner)
	require.NoError(t, err)

	intermediate, err := x509.ParseCertificate(intDER)
	require.NoError(t, err)

	ext := o.nonceExtension
	if ext == nil {
		ext = testMarshal(t, struct {
			Nonce []byte `asn1:"explicit,tag:1"`
		}{nonce}, "")
	}

	leafTmpl := &x509.Certificate{
		SerialNumber: testSerial(t),
		Subject:      pkix.Name{CommonName: "App Attest Key"},
		NotBefore:    e.now.Add(-time.Hour),
		NotAfter:     o.notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	if !o.omitNonce {
		leafTmpl.ExtraExtensions = []pkix.Extension{{Id: asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 8, 2}, Value: ext}}
	}

	key := e.key
	if o.certificateKey != nil {
		key = o.certificateKey
	}

	leafParent, leafSigner := intermediate, intKey

	if o.leafSigner != nil {
		leafParent, leafSigner = &x509.Certificate{Subject: intermediate.Subject, PublicKey: &o.leafSigner.PublicKey}, o.leafSigner
	}

	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, leafParent, &key.PublicKey, leafSigner)
	require.NoError(t, err)

	x5c := [][]byte{leafDER, intDER}

	if o.x5c != nil {
		x5c = o.x5c(x5c)
	}

	return x5c
}

func (e *testAppleEnv) object(t *testing.T, o *testAppleOpts, keyID []byte) []byte {
	authData := e.authData(t, o, keyID)

	clientDataHash := sha256.Sum256([]byte(o.registrationString))
	nonce := sha256.Sum256(append(append([]byte{}, authData...), clientDataHash[:]...))

	data, err := cbor.Marshal(map[string]any{
		"fmt":      o.format,
		"attStmt":  map[string]any{"x5c": e.chain(t, o, nonce[:]), "receipt": []byte("receipt")},
		"authData": authData,
	})
	require.NoError(t, err)

	return data
}

func testCBOR(t *testing.T, v any) []byte {
	data, err := cbor.Marshal(v)
	require.NoError(t, err)

	return data
}

func TestShouldVerifyAppleAppAttestAttestation(t *testing.T) {
	env := newTestAppleEnv(t)

	result, err := env.verifier().Verify(context.Background(), env.input(t, nil))

	require.NoError(t, err)
	assert.Equal(t, AttestationResult{
		Format:   AttestationFormatAppleAppAttest,
		Identity: testAppleTeamID + "." + testAppleBundleID,
		Detail:   "production",
		KeyID:    env.keyID(t),
		Attested: true,
	}, result)
}

func TestShouldVerifyAppleAppAttestDevelopmentEnvironment(t *testing.T) {
	env := newTestAppleEnv(t)

	env.config.Apple = []schema.AppAuthenticatorAppleIdentity{
		{TeamID: testAppleTeamID, BundleID: testAppleBundleID, Environment: "production"},
		{TeamID: testAppleTeamID, BundleID: testAppleBundleID, Environment: "development"},
	}

	result, err := env.verifier().Verify(context.Background(), env.input(t, func(o *testAppleOpts) { o.aaguid = []byte("appattestdevelop") }))

	require.NoError(t, err)
	assert.Equal(t, testAppleTeamID+"."+testAppleBundleID, result.Identity)
	assert.Equal(t, "development", result.Detail)
}

func TestShouldVerifyAppleAppAttestWithExtensions(t *testing.T) {
	env := newTestAppleEnv(t)

	extensions := testCBOR(t, map[string]any{"apple_bundle_version_01": "1", "apple_validation_category_01": []byte{3, 0, 0, 0}})

	result, err := env.verifier().Verify(context.Background(), env.input(t, func(o *testAppleOpts) {
		o.flags = 0xc0
		o.extensions = extensions
	}))

	require.NoError(t, err)
	assert.True(t, result.Attested)
}

func TestShouldAcceptAppleChainIncludingRoot(t *testing.T) {
	env := newTestAppleEnv(t)

	result, err := env.verifier().Verify(context.Background(), env.input(t, func(o *testAppleOpts) {
		o.x5c = func(x5c [][]byte) [][]byte { return append(x5c, env.root.Raw) }
	}))

	require.NoError(t, err)
	assert.True(t, result.Attested)
}

func TestShouldRejectInvalidAppleAppAttestAttestation(t *testing.T) {
	otherKey := newTestAppleEnv(t).key
	otherID := sha256.Sum256([]byte("other"))

	testCases := []struct {
		name   string
		opts   func(o *testAppleOpts)
		input  func(env *testAppleEnv, in *AttestationInput)
		config func(c *schema.AppAuthenticatorAttestation)
		err    string
	}{
		{
			name: "ShouldRejectNonceOverOtherRegistrationString",
			opts: func(o *testAppleOpts) { o.registrationString = "other registration string" },
			err:  "nonce does not match",
		},
		{
			name: "ShouldRejectWrongRelyingPartyIDHash",
			opts: func(o *testAppleOpts) { o.rpID = testAppleTeamID + ".com.example.other" },
			err:  "relying party id hash does not match any configured apple identity",
		},
		{
			name: "ShouldRejectNonZeroSignCount",
			opts: func(o *testAppleOpts) { o.signCount = 1 },
			err:  "sign count 1 is not zero",
		},
		{
			name: "ShouldRejectDevelopmentAgainstProduction",
			opts: func(o *testAppleOpts) { o.aaguid = []byte("appattestdevelop") },
			err:  "aaguid does not match the environment of any matching apple identity",
		},
		{
			name: "ShouldRejectProductionAgainstDevelopment",
			config: func(c *schema.AppAuthenticatorAttestation) {
				c.Apple[0].Environment = "development"
			},
			err: "aaguid does not match the environment of any matching apple identity",
		},
		{
			name: "ShouldRejectUnknownAAGUID",
			opts: func(o *testAppleOpts) { o.aaguid = make([]byte, 16) },
			err:  "aaguid does not match the environment of any matching apple identity",
		},
		{
			name: "ShouldRejectCredentialIDNotMatchingKeyID",
			opts: func(o *testAppleOpts) { o.credentialID = otherID[:] },
			err:  "credential id does not match the key id",
		},
		{
			name: "ShouldRejectKeyIDNotMatchingCertificateKey",
			input: func(_ *testAppleEnv, in *AttestationInput) {
				in.KeyID = base64.RawURLEncoding.EncodeToString(otherID[:])
			},
			err: "credential certificate public key does not match the key id",
		},
		{
			name: "ShouldRejectCertificateKeyNotMatchingKeyID",
			opts: func(o *testAppleOpts) { o.certificateKey = otherKey },
			err:  "credential certificate public key does not match the key id",
		},
		{
			name: "ShouldRejectExpiredIntermediate",
			opts: func(o *testAppleOpts) { o.intermediateNotAfter = time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC) },
			err:  "certificate 1 is not valid at the current time",
		},
		{
			name: "ShouldRejectExpiredCredentialCertificate",
			opts: func(o *testAppleOpts) { o.notAfter = time.Date(2026, 11, 12, 9, 30, 0, 0, time.UTC) },
			err:  "certificate 0 is not valid at the current time",
		},
		{
			name: "ShouldRejectIntermediateNotSignedByRoot",
			opts: func(o *testAppleOpts) { o.intermediateSigner = otherKey },
			err:  "certificate 1 is not signed by the trusted root",
		},
		{
			name: "ShouldRejectCredentialCertificateNotSignedByIntermediate",
			opts: func(o *testAppleOpts) { o.leafSigner = otherKey },
			err:  "certificate 0 signature is not valid: x509: ECDSA verification failure",
		},
		{
			name: "ShouldRejectIntermediateNotCA",
			opts: func(o *testAppleOpts) { o.intermediateNotCA = true },
			err:  "certificate 1 is not a certificate authority",
		},
		{
			name: "ShouldRejectEmptyChain",
			opts: func(o *testAppleOpts) { o.x5c = func(_ [][]byte) [][]byte { return nil } },
			err:  "certificate chain is empty",
		},
		{
			name: "ShouldRejectShortChain",
			opts: func(o *testAppleOpts) { o.x5c = func(x5c [][]byte) [][]byte { return x5c[:1] } },
			err:  "certificate chain is too short",
		},
		{
			name: "ShouldRejectLongChain",
			opts: func(o *testAppleOpts) {
				o.x5c = func(x5c [][]byte) [][]byte { return append(x5c, x5c[1], x5c[1], x5c[1], x5c[1]) }
			},
			err: "certificate chain is too long",
		},
		{
			name: "ShouldRejectMalformedCertificate",
			opts: func(o *testAppleOpts) {
				o.x5c = func(x5c [][]byte) [][]byte { return [][]byte{x5c[0], []byte("bad")} }
			},
			err: "error parsing certificate 1: x509: malformed certificate",
		},
		{
			name: "ShouldRejectMissingNonceExtension",
			opts: func(o *testAppleOpts) { o.omitNonce = true },
			err:  "credential certificate does not contain the nonce extension",
		},
		{
			name: "ShouldRejectMalformedNonceExtension",
			opts: func(o *testAppleOpts) { o.nonceExtension = []byte{0x30, 0x00} },
			err:  "error parsing nonce extension: asn1: syntax error: sequence truncated",
		},
		{
			name: "ShouldRejectNonceExtensionWithExtraElements",
			opts: func(o *testAppleOpts) {
				o.nonceExtension = testMarshal(t, struct {
					Nonce []byte `asn1:"explicit,tag:1"`
					Extra []byte `asn1:"explicit,tag:2"`
				}{make([]byte, 32), []byte{1}}, "")
			},
			err: "error parsing nonce extension: extension is not a sequence of exactly one nonce",
		},
		{
			name: "ShouldRejectMissingAttestedCredentialDataFlag",
			opts: func(o *testAppleOpts) { o.flags = 0 },
			err:  "authenticator data does not contain attested credential data",
		},
		{
			name: "ShouldRejectTrailingAuthenticatorData",
			opts: func(o *testAppleOpts) { o.trailing = []byte{0x00} },
			err:  "authenticator data contains trailing data",
		},
		{
			name: "ShouldRejectExtensionsWithoutFlag",
			opts: func(o *testAppleOpts) { o.extensions = testCBOR(t, map[string]any{"a": "b"}) },
			err:  "authenticator data contains trailing data",
		},
		{
			name: "ShouldRejectExtensionsFlagWithoutExtensions",
			opts: func(o *testAppleOpts) { o.flags = 0xc0 },
			err:  "error parsing authenticator data extensions: EOF",
		},
		{
			name: "ShouldRejectShortCredentialID",
			opts: func(o *testAppleOpts) { o.credentialID = []byte{1} },
			err:  "credential id does not match the key id",
		},
		{
			name: "ShouldRejectWrongFormat",
			opts: func(o *testAppleOpts) { o.format = "packed" },
			err:  "attestation object format 'packed' is not 'apple-appattest'",
		},
		{
			name:  "ShouldRejectInvalidKeyIDEncoding",
			input: func(_ *testAppleEnv, in *AttestationInput) { in.KeyID = "not base64url!" },
			err:   "key id is not valid base64url",
		},
		{
			name:  "ShouldRejectShortKeyID",
			input: func(_ *testAppleEnv, in *AttestationInput) { in.KeyID = "AAAA" },
			err:   "key id is not 32 bytes",
		},
		{
			name:  "ShouldRejectInvalidAttestationObjectEncoding",
			input: func(_ *testAppleEnv, in *AttestationInput) { in.AttestationObject = "not base64url!" },
			err:   "attestation object is not valid base64url",
		},
		{
			name: "ShouldRejectDuplicateMapKeys",
			input: func(_ *testAppleEnv, in *AttestationInput) {
				data := slices.Concat(
					[]byte{0xa3},
					testCBOR(t, "fmt"), testCBOR(t, AttestationFormatAppleAppAttest),
					testCBOR(t, "fmt"), testCBOR(t, AttestationFormatAppleAppAttest),
					testCBOR(t, "authData"), testCBOR(t, []byte{}),
				)
				in.AttestationObject = base64.RawURLEncoding.EncodeToString(data)
			},
			err: "error decoding attestation object: cbor: found duplicate map key \"fmt\" at map element index 1",
		},
		{
			name: "ShouldRejectTrailingAttestationObjectData",
			input: func(_ *testAppleEnv, in *AttestationInput) {
				data := append(testCBOR(t, map[string]any{"fmt": AttestationFormatAppleAppAttest}), 0x00)
				in.AttestationObject = base64.RawURLEncoding.EncodeToString(data)
			},
			err: "error decoding attestation object: cbor: 1 bytes of extraneous data starting at index 21",
		},
		{
			name: "ShouldRejectNoAppleIdentities",
			config: func(c *schema.AppAuthenticatorAttestation) {
				c.Apple = nil
			},
			err: "relying party id hash does not match any configured apple identity",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestAppleEnv(t)

			if tc.config != nil {
				tc.config(&env.config)
			}

			in := env.input(t, tc.opts)

			if tc.input != nil {
				tc.input(env, &in)
			}

			_, err := env.verifier().Verify(context.Background(), in)

			assert.EqualError(t, err, "error verifying attestation format 'apple-appattest': "+tc.err)
		})
	}
}

func TestShouldRejectAppleChainWithoutRoot(t *testing.T) {
	env := newTestAppleEnv(t)

	_, err := NewVerifier(env.config, WithClock(env.clock), WithAppleRoot(nil)).Verify(context.Background(), env.input(t, nil))

	assert.EqualError(t, err, "error verifying attestation format 'apple-appattest': no trusted root is configured")
}

func TestShouldVerifyRealAppleProductionAttestation(t *testing.T) {
	data, err := os.ReadFile("testdata/apple/production.txt")
	require.NoError(t, err)

	object, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
	require.NoError(t, err)

	keyID, err := base64.StdEncoding.DecodeString("VnfqjSp0rWyyqNhrfh+9/IhLIvXuYTPAmJEVQwl4dko=")
	require.NoError(t, err)

	config := schema.AppAuthenticatorAttestation{
		Mode:  schema.AppAuthenticatorAttestationModeRequired,
		Apple: []schema.AppAuthenticatorAppleIdentity{{TeamID: "8YE23NZS57", BundleID: "com.kayak.travel", Environment: "production"}},
	}

	in := AttestationInput{
		Format:             AttestationFormatAppleAppAttest,
		KeyID:              base64.RawURLEncoding.EncodeToString(keyID),
		AttestationObject:  base64.RawURLEncoding.EncodeToString(object),
		RegistrationString: "1234567890abcdefgh",
	}

	result, err := NewVerifier(config, WithClock(clock.NewFixed(time.Date(2020, 9, 28, 0, 0, 0, 0, time.UTC)))).Verify(context.Background(), in)

	require.NoError(t, err)
	assert.Equal(t, AttestationResult{
		Format:   AttestationFormatAppleAppAttest,
		Identity: "8YE23NZS57.com.kayak.travel",
		Detail:   "production",
		KeyID:    keyID,
		Attested: true,
	}, result)

	_, err = NewVerifier(config, WithClock(clock.NewFixed(time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)))).Verify(context.Background(), in)
	assert.EqualError(t, err, "error verifying attestation format 'apple-appattest': certificate 0 is not valid at the current time")

	in.RegistrationString = "1234567890abcdefgi"

	_, err = NewVerifier(config, WithClock(clock.NewFixed(time.Date(2020, 9, 28, 0, 0, 0, 0, time.UTC)))).Verify(context.Background(), in)
	assert.EqualError(t, err, "error verifying attestation format 'apple-appattest': nonce does not match")

	_, err = NewVerifier(config, WithClock(clock.NewFixed(time.Date(2020, 9, 28, 0, 0, 0, 0, time.UTC))), WithAppleRoot(newTestAppleEnv(t).root)).Verify(context.Background(), in)
	assert.EqualError(t, err, "error verifying attestation format 'apple-appattest': certificate 1 is not signed by the trusted root")
}

func TestShouldVerifyRealAppleAttestationWithExtensions(t *testing.T) {
	data, err := os.ReadFile("testdata/apple/ios27.txt")
	require.NoError(t, err)

	object, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(string(data)))
	require.NoError(t, err)

	keyID, err := base64.RawURLEncoding.DecodeString("5wvMiiwopebQYSxU76gBmh-9evrXbDwi2hdEIr8FXzM")
	require.NoError(t, err)

	clientDataHash, err := hex.DecodeString("5b1e4a7f3c2d8e91a605b3d472f8196ce347aa582c9d310e8467b2531fc09e77")
	require.NoError(t, err)

	production := schema.AppAuthenticatorAppleIdentity{TeamID: "K38V9TR83U", BundleID: "com.example.webauthn4j.appattest.debug", Environment: "production"}
	development := schema.AppAuthenticatorAppleIdentity{TeamID: "K38V9TR83U", BundleID: "com.example.webauthn4j.appattest.debug", Environment: "development"}

	v := NewVerifier(schema.AppAuthenticatorAttestation{Apple: []schema.AppAuthenticatorAppleIdentity{development, production}})

	identity, err := v.verifyAppleAttestationObject(time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC), keyID, object, clientDataHash)
	require.NoError(t, err)
	assert.Equal(t, production, identity)

	_, err = v.verifyAppleAttestationObject(time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC), keyID, object, clientDataHash)
	assert.EqualError(t, err, "certificate 0 is not valid at the current time")

	v = NewVerifier(schema.AppAuthenticatorAttestation{Apple: []schema.AppAuthenticatorAppleIdentity{development}})

	_, err = v.verifyAppleAttestationObject(time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC), keyID, object, clientDataHash)
	assert.EqualError(t, err, "aaguid does not match the environment of any matching apple identity")
}

func TestShouldLoadAppleAppAttestationRoot(t *testing.T) {
	root := AppleAppAttestationRoot()

	require.NotNil(t, root)

	fingerprint := sha256.Sum256(root.Raw)

	assert.Equal(t, "1cb9823ba28ba6ad2d33a006941de2ae4f513ef1d4e831b9f7e0fa7b6242c932", hex.EncodeToString(fingerprint[:]))
	assert.Equal(t, "Apple App Attestation Root CA", root.Subject.CommonName)
	assert.Len(t, GoogleAttestationRoots(), 5)
}
