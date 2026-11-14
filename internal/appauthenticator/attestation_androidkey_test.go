// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package appauthenticator

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

const (
	testAndroidPackage      = "com.example.app"
	testAndroidOtherPackage = "com.example.other"
)

var (
	testAndroidDigest      = sha256.Sum256([]byte("signing certificate"))
	testAndroidOtherDigest = sha256.Sum256([]byte("other signing certificate"))
)

type testStatusServer struct {
	mu       sync.Mutex
	body     string
	fail     bool
	requests int
	server   *httptest.Server
}

func newTestStatusServer(t *testing.T) *testStatusServer {
	s := &testStatusServer{body: `{"entries":{}}`}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.requests++

		if s.fail {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(s.body))
	}))

	t.Cleanup(s.server.Close)

	return s
}

func (s *testStatusServer) set(body string, fail bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.body, s.fail = body, fail
}

func (s *testStatusServer) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.requests
}

type testAndroidChainOpts struct {
	challenge         string
	attestationLevel  asn1.Enumerated
	keyMintLevel      asn1.Enumerated
	packageName       string
	digests           [][]byte
	hardware          func([]asn1.RawValue) []asn1.RawValue
	software          func([]asn1.RawValue) []asn1.RawValue
	certificateKey    *ecdsa.PublicKey
	kdOnIntermediate  bool
	intermediateNotCA bool
	leafSerial        *big.Int
	notAfter          time.Time

	intermediateKeyUsage x509.KeyUsage
	intermediateNotAfter time.Time
	intermediateSerial   *big.Int
	intermediateSigner   *ecdsa.PrivateKey
	leafSigner           *ecdsa.PrivateKey
	appIDInHardware      bool
}

type testAndroidEnv struct {
	now          time.Time
	clock        *clock.Fixed
	rootKey      *ecdsa.PrivateKey
	root         *x509.Certificate
	status       *testStatusServer
	config       schema.AppAuthenticatorAttestation
	approvalKey  *ecdsa.PrivateKey
	transportKey *ecdsa.PrivateKey
}

func newTestAndroidEnv(t *testing.T) *testAndroidEnv {
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)

	rootKey := testGenerateKey(t)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Attestation Root"},
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

	return &testAndroidEnv{
		now:     now,
		clock:   clock.NewFixed(now),
		rootKey: rootKey,
		root:    root,
		status:  newTestStatusServer(t),
		config: schema.AppAuthenticatorAttestation{
			Mode: schema.AppAuthenticatorAttestationModeRequired,
			Android: []schema.AppAuthenticatorAndroidIdentity{
				{PackageName: testAndroidPackage, SigningCertificateDigests: []string{hex.EncodeToString(testAndroidDigest[:])}},
			},
		},
		approvalKey:  testGenerateKey(t),
		transportKey: testGenerateKey(t),
	}
}

func (e *testAndroidEnv) verifier() *Verifier {
	return NewVerifier(e.config, WithClock(e.clock), WithAndroidRoots([]*x509.Certificate{e.root}), WithRevocationURL(e.status.server.URL), WithHTTPClient(e.status.server.Client()))
}

func (e *testAndroidEnv) defaultOpts() *testAndroidChainOpts {
	return &testAndroidChainOpts{
		challenge:            exChallenge,
		attestationLevel:     1,
		keyMintLevel:         1,
		packageName:          testAndroidPackage,
		digests:              [][]byte{testAndroidDigest[:]},
		notAfter:             e.now.Add(365 * 24 * time.Hour),
		intermediateKeyUsage: x509.KeyUsageCertSign,
		intermediateNotAfter: e.now.Add(5 * 365 * 24 * time.Hour),
	}
}

func (e *testAndroidEnv) input(t *testing.T, approval, transport func(o *testAndroidChainOpts)) AttestationInput {
	ao, to := e.defaultOpts(), e.defaultOpts()

	if approval != nil {
		approval(ao)
	}

	if transport != nil {
		transport(to)
	}

	return AttestationInput{
		Format:         AttestationFormatAndroidKey,
		ApprovalChain:  e.chain(t, ao, &e.approvalKey.PublicKey, true),
		TransportChain: e.chain(t, to, &e.transportKey.PublicKey, false),
		ApprovalSPKI:   testSPKI(t, &e.approvalKey.PublicKey),
		TransportSPKI:  testSPKI(t, &e.transportKey.PublicKey),
		Challenge:      exChallenge,
	}
}

func (e *testAndroidEnv) chain(t *testing.T, o *testAndroidChainOpts, key *ecdsa.PublicKey, approval bool) []string {
	kd := testKeyDescription(t, o, approval)

	ext := pkix.Extension{Id: oidAndroidKeyDescription, Value: kd}

	intKey := testGenerateKey(t)

	intSerial := o.intermediateSerial
	if intSerial == nil {
		intSerial = testSerial(t)
	}

	intTmpl := &x509.Certificate{
		SerialNumber:          intSerial,
		Subject:               pkix.Name{CommonName: "Test Attestation Intermediate"},
		NotBefore:             e.now.Add(-48 * time.Hour),
		NotAfter:              o.intermediateNotAfter,
		IsCA:                  !o.intermediateNotCA,
		BasicConstraintsValid: true,
		KeyUsage:              o.intermediateKeyUsage,
	}

	if o.kdOnIntermediate {
		intTmpl.ExtraExtensions = []pkix.Extension{ext}
	}

	intParent, intSigner := e.root, e.rootKey

	if o.intermediateSigner != nil {
		intParent, intSigner = &x509.Certificate{Subject: e.root.Subject, PublicKey: &o.intermediateSigner.PublicKey}, o.intermediateSigner
	}

	intDER, err := x509.CreateCertificate(rand.Reader, intTmpl, intParent, &intKey.PublicKey, intSigner)
	require.NoError(t, err)

	intermediate, err := x509.ParseCertificate(intDER)
	require.NoError(t, err)

	serial := o.leafSerial
	if serial == nil {
		serial = testSerial(t)
	}

	if o.certificateKey != nil {
		key = o.certificateKey
	}

	leafTmpl := &x509.Certificate{
		SerialNumber:    serial,
		Subject:         pkix.Name{CommonName: "Android Keystore Key"},
		NotBefore:       e.now.Add(-time.Hour),
		NotAfter:        o.notAfter,
		KeyUsage:        x509.KeyUsageDigitalSignature,
		ExtraExtensions: []pkix.Extension{ext},
	}

	leafParent, leafSigner := intermediate, intKey

	if o.leafSigner != nil {
		leafParent, leafSigner = &x509.Certificate{Subject: intermediate.Subject, PublicKey: &o.leafSigner.PublicKey}, o.leafSigner
	}

	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, leafParent, key, leafSigner)
	require.NoError(t, err)

	return []string{
		base64.RawURLEncoding.EncodeToString(leafDER),
		base64.RawURLEncoding.EncodeToString(intDER),
		base64.RawURLEncoding.EncodeToString(e.root.Raw),
	}
}

func testKeyDescription(t *testing.T, o *testAndroidChainOpts, approval bool) []byte {
	appID, err := asn1.Marshal(attestationApplicationID{
		PackageInfos:     []attestationPackageInfo{{PackageName: []byte(o.packageName), Version: 42}},
		SignatureDigests: o.digests,
	})
	require.NoError(t, err)

	rootOfTrust := testMarshal(t, struct {
		VerifiedBootKey   []byte
		DeviceLocked      bool
		VerifiedBootState asn1.Enumerated
		VerifiedBootHash  []byte
	}{make([]byte, 32), true, 0, make([]byte, 32)}, "")

	hardware := []asn1.RawValue{
		testTagged(1, testMarshal(t, []int{2}, "set")),
		testTagged(2, testMarshal(t, 3, "")),
		testTagged(3, testMarshal(t, 256, "")),
		testTagged(5, testMarshal(t, []int{4}, "set")),
		testTagged(10, testMarshal(t, 1, "")),
	}

	if approval {
		hardware = append(hardware, testTagged(504, testMarshal(t, 2, "")))
	} else {
		hardware = append(hardware, testTagged(503, asn1.NullBytes))
	}

	hardware = append(hardware,
		testTagged(702, testMarshal(t, 0, "")),
		testTagged(704, rootOfTrust),
		testTagged(705, testMarshal(t, 150000, "")),
	)

	software := []asn1.RawValue{
		testTagged(701, testMarshal(t, int64(1794125087000), "")),
	}

	if o.appIDInHardware {
		hardware = append(hardware, testTagged(709, testMarshal(t, appID, "")))
	} else {
		software = append(software, testTagged(709, testMarshal(t, appID, "")))
	}

	if o.hardware != nil {
		hardware = o.hardware(hardware)
	}

	if o.software != nil {
		software = o.software(software)
	}

	return testMarshal(t, keyDescription{
		AttestationVersion:       300,
		AttestationSecurityLevel: o.attestationLevel,
		KeyMintVersion:           300,
		KeyMintSecurityLevel:     o.keyMintLevel,
		AttestationChallenge:     []byte(o.challenge),
		UniqueID:                 []byte{},
		SoftwareEnforced:         software,
		HardwareEnforced:         hardware,
	}, "")
}

func testTagged(tag int, inner []byte) asn1.RawValue {
	return asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: tag, IsCompound: true, Bytes: inner}
}

func testMarshal(t *testing.T, v any, params string) []byte {
	b, err := asn1.MarshalWithParams(v, params)
	require.NoError(t, err)

	return b
}

func testWithout(tag int) func([]asn1.RawValue) []asn1.RawValue {
	return func(list []asn1.RawValue) []asn1.RawValue {
		out := make([]asn1.RawValue, 0, len(list))

		for _, v := range list {
			if v.Tag != tag {
				out = append(out, v)
			}
		}

		return out
	}
}

func testWith(v asn1.RawValue) func([]asn1.RawValue) []asn1.RawValue {
	return func(list []asn1.RawValue) []asn1.RawValue {
		return append(testWithout(v.Tag)(list), v)
	}
}

func testGenerateKey(t *testing.T) *ecdsa.PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	return key
}

func testSPKI(t *testing.T, key *ecdsa.PublicKey) []byte {
	spki, err := x509.MarshalPKIXPublicKey(key)
	require.NoError(t, err)

	return spki
}

func testSerial(t *testing.T) *big.Int {
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	require.NoError(t, err)

	return serial.Add(serial, big.NewInt(1))
}

func TestShouldVerifyAndroidKeyAttestation(t *testing.T) {
	env := newTestAndroidEnv(t)

	result, err := env.verifier().Verify(context.Background(), env.input(t, nil, nil))

	require.NoError(t, err)
	assert.Equal(t, AttestationResult{Format: AttestationFormatAndroidKey, Identity: testAndroidPackage, Detail: "TrustedEnvironment", Attested: true}, result)
}

func TestShouldReportLowerAndroidSecurityLevel(t *testing.T) {
	env := newTestAndroidEnv(t)

	strongbox := func(o *testAndroidChainOpts) {
		o.attestationLevel, o.keyMintLevel = 2, 2
	}

	result, err := env.verifier().Verify(context.Background(), env.input(t, strongbox, strongbox))
	require.NoError(t, err)
	assert.Equal(t, "StrongBox", result.Detail)

	result, err = env.verifier().Verify(context.Background(), env.input(t, strongbox, nil))
	require.NoError(t, err)
	assert.Equal(t, "TrustedEnvironment", result.Detail)
}

func TestShouldAcceptZeroAuthTimeout(t *testing.T) {
	env := newTestAndroidEnv(t)

	_, err := env.verifier().Verify(context.Background(), env.input(t, func(o *testAndroidChainOpts) {
		o.hardware = testWith(testTagged(505, []byte{0x02, 0x01, 0x00}))
	}, nil))

	assert.NoError(t, err)
}

func TestShouldRejectInvalidAndroidKeyAttestation(t *testing.T) {
	otherSigner := testGenerateKey(t)
	otherKey := &otherSigner.PublicKey

	decodedChallenge, err := base64.RawURLEncoding.DecodeString(exChallenge)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		config    func(c *schema.AppAuthenticatorAttestation)
		approval  func(o *testAndroidChainOpts)
		transport func(o *testAndroidChainOpts)
		input     func(in *AttestationInput)
		err       string
	}{
		{
			name:     "ShouldRejectWrongChallenge",
			approval: func(o *testAndroidChainOpts) { o.challenge = exVerifier },
			err:      "approval certificate chain: attestation challenge does not match",
		},
		{
			name:     "ShouldRejectDecodedChallenge",
			approval: func(o *testAndroidChainOpts) { o.challenge = string(decodedChallenge) },
			err:      "approval certificate chain: attestation challenge does not match",
		},
		{
			name:     "ShouldRejectSoftwareAttestationSecurityLevel",
			approval: func(o *testAndroidChainOpts) { o.attestationLevel = 0 },
			err:      "approval certificate chain: attestation security level 'Software' is not hardware backed",
		},
		{
			name:      "ShouldRejectSoftwareKeyMintSecurityLevel",
			transport: func(o *testAndroidChainOpts) { o.keyMintLevel = 0 },
			err:       "transport certificate chain: keymint security level 'Software' is not hardware backed",
		},
		{
			name:     "ShouldRejectLeafKeyMismatch",
			approval: func(o *testAndroidChainOpts) { o.certificateKey = otherKey },
			err:      "approval certificate chain: leaf public key does not match the submitted public key",
		},
		{
			name:     "ShouldRejectKeyDescriptionOnIntermediate",
			approval: func(o *testAndroidChainOpts) { o.kdOnIntermediate = true },
			err:      "approval certificate chain: certificate 1 contains the key description extension",
		},
		{
			name:     "ShouldRejectIntermediateNotCA",
			approval: func(o *testAndroidChainOpts) { o.intermediateNotCA = true },
			err:      "approval certificate chain: certificate 1 is not a certificate authority",
		},
		{
			name:     "ShouldRejectPackageNameMismatch",
			approval: func(o *testAndroidChainOpts) { o.packageName = testAndroidOtherPackage },
			err:      "approval certificate chain: attestation application id does not match any configured android identity",
		},
		{
			name:     "ShouldRejectDigestNotConfigured",
			approval: func(o *testAndroidChainOpts) { o.digests = [][]byte{testAndroidDigest[:], testAndroidOtherDigest[:]} },
			err:      "approval certificate chain: attestation application id does not match any configured android identity",
		},
		{
			name:     "ShouldRejectEmptyDigests",
			approval: func(o *testAndroidChainOpts) { o.digests = [][]byte{} },
			err:      "approval certificate chain: attestation application id does not match any configured android identity",
		},
		{
			name: "ShouldRejectChainsMatchingDifferentIdentities",
			config: func(c *schema.AppAuthenticatorAttestation) {
				c.Android = append(c.Android, schema.AppAuthenticatorAndroidIdentity{PackageName: testAndroidOtherPackage, SigningCertificateDigests: []string{hex.EncodeToString(testAndroidOtherDigest[:])}})
			},
			transport: func(o *testAndroidChainOpts) {
				o.packageName, o.digests = testAndroidOtherPackage, [][]byte{testAndroidOtherDigest[:]}
			},
			err: "certificate chains match different android identities",
		},
		{
			name:     "ShouldRejectApprovalNoAuthRequired",
			approval: func(o *testAndroidChainOpts) { o.hardware = testWith(testTagged(503, asn1.NullBytes)) },
			err:      "approval certificate chain: no auth required is present",
		},
		{
			name:     "ShouldRejectApprovalSoftwareNoAuthRequired",
			approval: func(o *testAndroidChainOpts) { o.software = testWith(testTagged(503, asn1.NullBytes)) },
			err:      "approval certificate chain: no auth required is present",
		},
		{
			name:     "ShouldRejectApprovalUserAuthTypeWithPassword",
			approval: func(o *testAndroidChainOpts) { o.hardware = testWith(testTagged(504, []byte{0x02, 0x01, 0x03})) },
			err:      "approval certificate chain: user auth type 3 is not biometric only",
		},
		{
			name:     "ShouldRejectApprovalUserAuthTypeAbsent",
			approval: func(o *testAndroidChainOpts) { o.hardware = testWithout(504) },
			err:      "approval certificate chain: user auth type is absent",
		},
		{
			name:     "ShouldRejectApprovalAuthTimeout",
			approval: func(o *testAndroidChainOpts) { o.hardware = testWith(testTagged(505, []byte{0x02, 0x01, 0x1e})) },
			err:      "approval certificate chain: auth timeout 30 is not zero",
		},
		{
			name:      "ShouldRejectTransportWithoutNoAuthRequired",
			transport: func(o *testAndroidChainOpts) { o.hardware = testWithout(503) },
			err:       "transport certificate chain: no auth required is absent",
		},
		{
			name:     "ShouldRejectWrongAlgorithm",
			approval: func(o *testAndroidChainOpts) { o.hardware = testWith(testTagged(2, []byte{0x02, 0x01, 0x01})) },
			err:      "approval certificate chain: algorithm 1 is not EC",
		},
		{
			name:     "ShouldRejectImportedKey",
			approval: func(o *testAndroidChainOpts) { o.hardware = testWith(testTagged(702, []byte{0x02, 0x01, 0x02})) },
			err:      "approval certificate chain: origin 2 is not generated",
		},
		{
			name:     "ShouldRejectMissingPurpose",
			approval: func(o *testAndroidChainOpts) { o.hardware = testWithout(1) },
			err:      "approval certificate chain: purpose does not contain sign",
		},
		{
			name:     "ShouldRejectExpiredLeaf",
			approval: func(o *testAndroidChainOpts) { o.notAfter = time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC) },
			err:      "approval certificate chain: certificate 0 is not valid at the current time",
		},
		{
			name:  "ShouldRejectEmptyChain",
			input: func(in *AttestationInput) { in.TransportChain = nil },
			err:   "transport certificate chain: chain is empty",
		},
		{
			name:  "ShouldRejectBadEncoding",
			input: func(in *AttestationInput) { in.ApprovalChain[1] = "!!" },
			err:   "approval certificate chain: certificate 1 is not valid base64url",
		},
		{
			name:  "ShouldRejectUntrustedRoot",
			input: func(in *AttestationInput) { in.ApprovalChain = in.ApprovalChain[:2] },
			err:   "approval certificate chain: chain does not end in a trusted root",
		},
		{
			name:     "ShouldRejectExpiredIntermediate",
			approval: func(o *testAndroidChainOpts) { o.intermediateNotAfter = time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC) },
			err:      "approval certificate chain: certificate 1 is not valid at the current time",
		},
		{
			name:     "ShouldRejectIntermediateNotSignedByRoot",
			approval: func(o *testAndroidChainOpts) { o.intermediateSigner = otherSigner },
			err:      "approval certificate chain: certificate 1 signature is not valid: x509: ECDSA verification failure",
		},
		{
			name:      "ShouldRejectLeafNotSignedByIntermediate",
			transport: func(o *testAndroidChainOpts) { o.leafSigner = otherSigner },
			err:       "transport certificate chain: certificate 0 signature is not valid: x509: ECDSA verification failure",
		},
		{
			name: "ShouldRejectSwappedLeafAndIntermediate",
			input: func(in *AttestationInput) {
				in.ApprovalChain[0], in.ApprovalChain[1] = in.ApprovalChain[1], in.ApprovalChain[0]
			},
			err: "approval certificate chain: certificate 0 signature is not valid: x509: ECDSA verification failure",
		},
		{
			name:     "ShouldRejectIntermediateWithoutKeyCertSign",
			approval: func(o *testAndroidChainOpts) { o.intermediateKeyUsage = x509.KeyUsageDigitalSignature },
			err:      "approval certificate chain: certificate 1 key usage does not include certificate signing",
		},
		{
			name: "ShouldRejectApprovalUserAuthTypeOnlyInSoftware",
			approval: func(o *testAndroidChainOpts) {
				o.hardware = testWithout(504)
				o.software = testWith(testTagged(504, []byte{0x02, 0x01, 0x02}))
			},
			err: "approval certificate chain: user auth type is absent",
		},
		{
			name:     "ShouldRejectApprovalAuthTimeoutOnlyInSoftware",
			approval: func(o *testAndroidChainOpts) { o.software = testWith(testTagged(505, []byte{0x02, 0x01, 0x1e})) },
			err:      "approval certificate chain: auth timeout 30 is not zero",
		},
		{
			name: "ShouldRejectTransportNoAuthRequiredOnlyInSoftware",
			transport: func(o *testAndroidChainOpts) {
				o.hardware = testWithout(503)
				o.software = testWith(testTagged(503, asn1.NullBytes))
			},
			err: "transport certificate chain: no auth required is absent",
		},
		{
			name:     "ShouldRejectApplicationIDOnlyInHardware",
			approval: func(o *testAndroidChainOpts) { o.appIDInHardware = true },
			err:      "approval certificate chain: attestation application id is absent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestAndroidEnv(t)

			if tc.config != nil {
				tc.config(&env.config)
			}

			in := env.input(t, tc.approval, tc.transport)

			if tc.input != nil {
				tc.input(&in)
			}

			_, err := env.verifier().Verify(context.Background(), in)

			assert.EqualError(t, err, "error verifying attestation format 'android-key': "+tc.err)
		})
	}
}

func TestShouldRejectAndroidRevokedSerial(t *testing.T) {
	testCases := []struct {
		name string
		key  func(serial *big.Int) string
	}{
		{"ShouldMatchHex", func(serial *big.Int) string { return serial.Text(16) }},
		{"ShouldMatchDecimal", func(serial *big.Int) string { return serial.Text(10) }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestAndroidEnv(t)

			serial := testSerial(t)

			env.status.set(`{"entries":{"`+tc.key(serial)+`":{"status":"REVOKED","reason":"KEY_COMPROMISE"}}}`, false)

			_, err := env.verifier().Verify(context.Background(), env.input(t, nil, func(o *testAndroidChainOpts) { o.leafSerial = serial }))

			assert.EqualError(t, err, "error verifying attestation format 'android-key': transport certificate chain: certificate 0 has revocation status 'REVOKED'")
		})
	}
}

func TestShouldRejectAndroidRevokedIntermediate(t *testing.T) {
	env := newTestAndroidEnv(t)

	serial := testSerial(t)

	env.status.set(`{"entries":{"`+serial.Text(16)+`":{"status":"SUSPENDED"}}}`, false)

	_, err := env.verifier().Verify(context.Background(), env.input(t, func(o *testAndroidChainOpts) { o.intermediateSerial = serial }, nil))

	assert.EqualError(t, err, "error verifying attestation format 'android-key': approval certificate chain: certificate 1 has revocation status 'SUSPENDED'")
}

func TestShouldAcceptExpiredPinnedRootCertificate(t *testing.T) {
	env := newTestAndroidEnv(t)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               env.root.Subject,
		NotBefore:             env.now.Add(-10 * 365 * 24 * time.Hour),
		NotAfter:              env.now.Add(-24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &env.rootKey.PublicKey, env.rootKey)
	require.NoError(t, err)

	in := env.input(t, nil, nil)
	in.ApprovalChain[2] = base64.RawURLEncoding.EncodeToString(der)
	in.TransportChain[2] = base64.RawURLEncoding.EncodeToString(der)

	result, err := env.verifier().Verify(context.Background(), in)

	require.NoError(t, err)
	assert.True(t, result.Attested)
}

func TestShouldParseRealAndroidAttestationChain(t *testing.T) {
	data, err := os.ReadFile("testdata/android/tee_chain.txt")
	require.NoError(t, err)

	certs, err := decodeAndroidChain(strings.Fields(string(data)))
	require.NoError(t, err)
	require.Len(t, certs, 5)

	v := NewVerifier(schema.AppAuthenticatorAttestation{})

	assert.NoError(t, v.verifyAndroidChainTrust(time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), certs))
	assert.EqualError(t, v.verifyAndroidChainTrust(time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC), certs), "certificate 1 is not valid at the current time")

	kd, err := parseAndroidKeyDescription(certs)
	require.NoError(t, err)

	assert.Equal(t, 300, kd.AttestationVersion)
	assert.Equal(t, asn1.Enumerated(1), kd.AttestationSecurityLevel)
	assert.Equal(t, 300, kd.KeyMintVersion)
	assert.Equal(t, asn1.Enumerated(1), kd.KeyMintSecurityLevel)
	assert.Equal(t, "5652e2dc45549a96f96afa225502f87fadc08a60bc021392c0be8c5062fd5f5e", hex.EncodeToString(kd.AttestationChallenge))

	level, err := androidSecurityLevel(kd)
	require.NoError(t, err)
	assert.Equal(t, "TrustedEnvironment", androidSecurityLevelName(level))

	hardware, err := parseAuthorizationList(kd.HardwareEnforced)
	require.NoError(t, err)

	software, err := parseAuthorizationList(kd.SoftwareEnforced)
	require.NoError(t, err)

	assert.NoError(t, checkAndroidHardwareEnforced(hardware))
	assert.True(t, hardware.has(androidTagUserAuthType))
	assert.False(t, hardware.has(androidTagNoAuthRequired))
	assert.False(t, software.has(androidTagNoAuthRequired))

	appID, err := parseAttestationApplicationID(software)
	require.NoError(t, err)

	require.Len(t, appID.PackageInfos, 2)
	assert.Equal(t, "com.google.android.gsf", string(appID.PackageInfos[0].PackageName))
	assert.Equal(t, int64(35), appID.PackageInfos[0].Version)
	assert.Equal(t, "com.google.android.gms", string(appID.PackageInfos[1].PackageName))
	assert.Equal(t, int64(250232035), appID.PackageInfos[1].Version)
	require.Len(t, appID.SignatureDigests, 1)
	assert.Equal(t, "f0fd6c5b410f25cb25c3b53346c8972fae30f8ee7411df910480ad6b2d60db83", hex.EncodeToString(appID.SignatureDigests[0]))

	v = NewVerifier(schema.AppAuthenticatorAttestation{Android: []schema.AppAuthenticatorAndroidIdentity{
		{PackageName: "com.google.android.gms", SigningCertificateDigests: []string{"f0fd6c5b410f25cb25c3b53346c8972fae30f8ee7411df910480ad6b2d60db83"}},
	}})

	identities, err := v.matchAndroidIdentities(software)
	require.NoError(t, err)
	assert.Equal(t, []int{0}, identities)
}

func TestShouldCacheAndroidStatusList(t *testing.T) {
	env := newTestAndroidEnv(t)
	v := env.verifier()

	_, err := v.Verify(context.Background(), env.input(t, nil, nil))
	require.NoError(t, err)
	assert.Equal(t, 1, env.status.count())

	env.clock.Set(env.now.Add(23 * time.Hour))

	_, err = v.Verify(context.Background(), env.input(t, nil, nil))
	require.NoError(t, err)
	assert.Equal(t, 1, env.status.count())

	env.status.set(`{"entries":{}}`, true)
	env.clock.Set(env.now.Add(48 * time.Hour))

	_, err = v.Verify(context.Background(), env.input(t, nil, nil))
	require.NoError(t, err)
	assert.Equal(t, 2, env.status.count())

	env.clock.Set(env.now.Add(48*time.Hour + 30*time.Second))

	_, err = v.Verify(context.Background(), env.input(t, nil, nil))
	require.NoError(t, err)
	assert.Equal(t, 2, env.status.count())

	env.clock.Set(env.now.Add(48*time.Hour + 2*time.Minute))

	_, err = v.Verify(context.Background(), env.input(t, nil, nil))
	require.NoError(t, err)
	assert.Equal(t, 3, env.status.count())

	env.status.set(`{"entries":{}}`, false)
	env.clock.Set(env.now.Add(72 * time.Hour))

	_, err = v.Verify(context.Background(), env.input(t, nil, nil))
	require.NoError(t, err)
	assert.Equal(t, 4, env.status.count())

	env.status.set(`{"entries":{}}`, true)

	env.clock.Set(env.now.Add(72*time.Hour + 8*24*time.Hour))

	_, err = v.Verify(context.Background(), env.input(t, nil, nil))
	assert.ErrorContains(t, err, "approval certificate chain: error checking revocation: no usable attestation status list: ")
}

func TestShouldFailClosedWithoutAndroidStatusList(t *testing.T) {
	env := newTestAndroidEnv(t)

	env.status.set("", true)

	_, err := env.verifier().Verify(context.Background(), env.input(t, nil, nil))

	assert.EqualError(t, err, "error verifying attestation format 'android-key': approval certificate chain: error checking revocation: no usable attestation status list: unexpected status code 500")
}

func TestShouldRejectMalformedAndroidStatusList(t *testing.T) {
	env := newTestAndroidEnv(t)

	env.status.set(`{"other":{}}`, false)

	_, err := env.verifier().Verify(context.Background(), env.input(t, nil, nil))

	assert.EqualError(t, err, "error verifying attestation format 'android-key': approval certificate chain: error checking revocation: no usable attestation status list: status list has no entries")
}

func TestShouldLoadGoogleAttestationRoots(t *testing.T) {
	roots := GoogleAttestationRoots()

	require.Len(t, roots, 5)

	var ec int

	for _, root := range roots {
		if _, ok := root.PublicKey.(*ecdsa.PublicKey); ok {
			ec++
		}
	}

	assert.Equal(t, 1, ec)
}
