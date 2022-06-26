package utils

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestShouldSetupDefaultTLSMinVersionOnErr(t *testing.T) {
	schemaTLSConfig := &schema.TLSConfig{
		MinimumVersion: "NotAVersion",
		ServerName:     "golang.org",
		SkipVerify:     true,
	}

	tlsConfig := NewTLSConfig(schemaTLSConfig, tls.VersionTLS12, nil)

	assert.Equal(t, uint16(tls.VersionTLS12), tlsConfig.MinVersion)
	assert.Equal(t, "golang.org", tlsConfig.ServerName)
	assert.True(t, tlsConfig.InsecureSkipVerify)
}

func TestShouldReturnCorrectTLSVersions(t *testing.T) {
	tls13 := uint16(tls.VersionTLS13)
	tls12 := uint16(tls.VersionTLS12)
	tls11 := uint16(tls.VersionTLS11)
	tls10 := uint16(tls.VersionTLS10)

	version, err := TLSStringToTLSConfigVersion(TLS13)
	assert.Equal(t, tls13, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS13)
	assert.Equal(t, tls13, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(TLS12)
	assert.Equal(t, tls12, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS12)
	assert.Equal(t, tls12, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(TLS11)
	assert.Equal(t, tls11, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS11)
	assert.Equal(t, tls11, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(TLS10)
	assert.Equal(t, tls10, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS10)
	assert.Equal(t, tls10, version)
	assert.NoError(t, err)
}

func TestShouldReturnZeroAndErrorOnInvalidTLSVersions(t *testing.T) {
	version, err := TLSStringToTLSConfigVersion("TLS1.4")
	assert.Error(t, err)
	assert.Equal(t, uint16(0), version)
	assert.EqualError(t, err, "supplied tls version isn't supported")

	version, err = TLSStringToTLSConfigVersion("SSL3.0")
	assert.Error(t, err)
	assert.Equal(t, uint16(0), version)
	assert.EqualError(t, err, "supplied tls version isn't supported")
}

func TestShouldReturnErrWhenX509DirectoryNotExist(t *testing.T) {
	pool, warnings, errors := NewX509CertPool("/tmp/asdfzyxabc123/not/a/real/dir")
	assert.NotNil(t, pool)

	if runtime.GOOS == windows {
		require.Len(t, warnings, 1)
		assert.EqualError(t, warnings[0], "could not load system certificate pool which may result in untrusted certificate issues: crypto/x509: system root pool is not available on Windows")
	} else {
		assert.Len(t, warnings, 0)
	}

	require.Len(t, errors, 1)

	if runtime.GOOS == windows {
		assert.EqualError(t, errors[0], "could not read certificates from directory open /tmp/asdfzyxabc123/not/a/real/dir: The system cannot find the path specified.")
	} else {
		assert.EqualError(t, errors[0], "could not read certificates from directory open /tmp/asdfzyxabc123/not/a/real/dir: no such file or directory")
	}
}

func TestShouldNotReturnErrWhenX509DirectoryExist(t *testing.T) {
	pool, warnings, errors := NewX509CertPool("/tmp")
	assert.NotNil(t, pool)

	if runtime.GOOS == windows {
		require.Len(t, warnings, 1)
		assert.EqualError(t, warnings[0], "could not load system certificate pool which may result in untrusted certificate issues: crypto/x509: system root pool is not available on Windows")
	} else {
		assert.Len(t, warnings, 0)
	}

	assert.Len(t, errors, 0)
}

func TestShouldReadCertsFromDirectoryButNotKeys(t *testing.T) {
	pool, warnings, errors := NewX509CertPool("../suites/common/ssl/")
	assert.NotNil(t, pool)
	require.Len(t, errors, 1)

	if runtime.GOOS == "windows" {
		require.Len(t, warnings, 1)
		assert.EqualError(t, warnings[0], "could not load system certificate pool which may result in untrusted certificate issues: crypto/x509: system root pool is not available on Windows")
	} else {
		assert.Len(t, warnings, 0)
	}

	assert.EqualError(t, errors[0], "could not import certificate key.pem")
}

func TestShouldGenerateCertificateAndPersistIt(t *testing.T) {
	testCases := []struct {
		Name              string
		PrivateKeyBuilder PrivateKeyBuilder
	}{
		{
			Name:              "P224",
			PrivateKeyBuilder: ECDSAKeyBuilder{}.WithCurve(elliptic.P224()),
		},
		{
			Name:              "P256",
			PrivateKeyBuilder: ECDSAKeyBuilder{}.WithCurve(elliptic.P256()),
		},
		{
			Name:              "P384",
			PrivateKeyBuilder: ECDSAKeyBuilder{}.WithCurve(elliptic.P384()),
		},
		{
			Name:              "P521",
			PrivateKeyBuilder: ECDSAKeyBuilder{}.WithCurve(elliptic.P521()),
		},
		{
			Name:              "Ed25519",
			PrivateKeyBuilder: Ed25519KeyBuilder{},
		},
		{
			Name:              "RSA",
			PrivateKeyBuilder: RSAKeyBuilder{keySizeInBits: 2048},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			certBytes, keyBytes, err := GenerateCertificate(tc.PrivateKeyBuilder, []string{"authelia.com", "example.org"}, time.Now(), 3*time.Hour, false)
			require.NoError(t, err)
			assert.True(t, len(certBytes) > 0)
			assert.True(t, len(keyBytes) > 0)
		})
	}
}

func TestShouldParseKeySigAlgorithm(t *testing.T) {
	testCases := []struct {
		Name               string
		InputKey, InputSig string
		ExpectedKeyAlg     x509.PublicKeyAlgorithm
		ExpectedSigAlg     x509.SignatureAlgorithm
	}{
		{
			Name:           "ShouldNotParseInvalidKeyAlg",
			InputKey:       "DDD",
			InputSig:       "SHA1",
			ExpectedKeyAlg: x509.UnknownPublicKeyAlgorithm,
			ExpectedSigAlg: x509.UnknownSignatureAlgorithm,
		},
		{
			Name:           "ShouldParseKeyRSASigSHA1",
			InputKey:       "RSA",
			InputSig:       "SHA1",
			ExpectedKeyAlg: x509.RSA,
			ExpectedSigAlg: x509.SHA1WithRSA,
		},
		{
			Name:           "ShouldParseKeyRSASigSHA256",
			InputKey:       "RSA",
			InputSig:       "SHA256",
			ExpectedKeyAlg: x509.RSA,
			ExpectedSigAlg: x509.SHA256WithRSA,
		},
		{
			Name:           "ShouldParseKeyRSASigSHA384",
			InputKey:       "RSA",
			InputSig:       "SHA384",
			ExpectedKeyAlg: x509.RSA,
			ExpectedSigAlg: x509.SHA384WithRSA,
		},
		{
			Name:           "ShouldParseKeyRSASigSHA512",
			InputKey:       "RSA",
			InputSig:       "SHA512",
			ExpectedKeyAlg: x509.RSA,
			ExpectedSigAlg: x509.SHA512WithRSA,
		},
		{
			Name:           "ShouldNotParseKeyRSASigInvalid",
			InputKey:       "RSA",
			InputSig:       "INVALID",
			ExpectedKeyAlg: x509.RSA,
			ExpectedSigAlg: x509.UnknownSignatureAlgorithm,
		},
		{
			Name:           "ShouldParseKeyECDSASigSHA1",
			InputKey:       "ECDSA",
			InputSig:       "SHA1",
			ExpectedKeyAlg: x509.ECDSA,
			ExpectedSigAlg: x509.ECDSAWithSHA1,
		},
		{
			Name:           "ShouldParseKeyECDSASigSHA256",
			InputKey:       "ECDSA",
			InputSig:       "SHA256",
			ExpectedKeyAlg: x509.ECDSA,
			ExpectedSigAlg: x509.ECDSAWithSHA256,
		},
		{
			Name:           "ShouldParseKeyECDSASigSHA384",
			InputKey:       "ECDSA",
			InputSig:       "SHA384",
			ExpectedKeyAlg: x509.ECDSA,
			ExpectedSigAlg: x509.ECDSAWithSHA384,
		},
		{
			Name:           "ShouldParseKeyECDSASigSHA512",
			InputKey:       "ECDSA",
			InputSig:       "SHA512",
			ExpectedKeyAlg: x509.ECDSA,
			ExpectedSigAlg: x509.ECDSAWithSHA512,
		},
		{
			Name:           "ShouldNotParseKeyECDSASigInvalid",
			InputKey:       "ECDSA",
			InputSig:       "INVALID",
			ExpectedKeyAlg: x509.ECDSA,
			ExpectedSigAlg: x509.UnknownSignatureAlgorithm,
		},
		{
			Name:           "ShouldParseKeyEd25519SigSHA1",
			InputKey:       "ED25519",
			InputSig:       "SHA1",
			ExpectedKeyAlg: x509.Ed25519,
			ExpectedSigAlg: x509.PureEd25519,
		},
		{
			Name:           "ShouldParseKeyEd25519SigSHA256",
			InputKey:       "ED25519",
			InputSig:       "SHA256",
			ExpectedKeyAlg: x509.Ed25519,
			ExpectedSigAlg: x509.PureEd25519,
		},
		{
			Name:           "ShouldParseKeyEd25519SigSHA384",
			InputKey:       "ED25519",
			InputSig:       "SHA384",
			ExpectedKeyAlg: x509.Ed25519,
			ExpectedSigAlg: x509.PureEd25519,
		},
		{
			Name:           "ShouldParseKeyEd25519SigSHA512",
			InputKey:       "ED25519",
			InputSig:       "SHA512",
			ExpectedKeyAlg: x509.Ed25519,
			ExpectedSigAlg: x509.PureEd25519,
		},
		{
			Name:           "ShouldParseKeyEd25519SigInvalid",
			InputKey:       "ED25519",
			InputSig:       "INVALID",
			ExpectedKeyAlg: x509.Ed25519,
			ExpectedSigAlg: x509.PureEd25519,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actualKey, actualSig := KeySigAlgorithmFromString(tc.InputKey, tc.InputSig)
			actualKeyLower, actualSigLower := KeySigAlgorithmFromString(strings.ToLower(tc.InputKey), strings.ToLower(tc.InputSig))

			assert.Equal(t, tc.ExpectedKeyAlg, actualKey)
			assert.Equal(t, tc.ExpectedSigAlg, actualSig)
			assert.Equal(t, tc.ExpectedKeyAlg, actualKeyLower)
			assert.Equal(t, tc.ExpectedSigAlg, actualSigLower)
		})
	}
}

func TestShouldParseCurves(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    string
		Expected elliptic.Curve
	}{
		{
			Name:     "P224-Standard",
			Input:    "P224",
			Expected: elliptic.P224(),
		},
		{
			Name:     "P224-Lowercase",
			Input:    "p224",
			Expected: elliptic.P224(),
		},
		{
			Name:     "P224-Hyphenated",
			Input:    "P-224",
			Expected: elliptic.P224(),
		},
		{
			Name:     "P256-Standard",
			Input:    "P256",
			Expected: elliptic.P256(),
		},
		{
			Name:     "P256-Lowercase",
			Input:    "p256",
			Expected: elliptic.P256(),
		},
		{
			Name:     "P256-Hyphenated",
			Input:    "P-256",
			Expected: elliptic.P256(),
		},
		{
			Name:     "P384-Standard",
			Input:    "P384",
			Expected: elliptic.P384(),
		},
		{
			Name:     "P384-Lowercase",
			Input:    "p384",
			Expected: elliptic.P384(),
		},
		{
			Name:     "P384-Hyphenated",
			Input:    "P-384",
			Expected: elliptic.P384(),
		},
		{
			Name:     "P521-Standard",
			Input:    "P521",
			Expected: elliptic.P521(),
		},
		{
			Name:     "P521-Lowercase",
			Input:    "p521",
			Expected: elliptic.P521(),
		},
		{
			Name:     "P521-Hyphenated",
			Input:    "P-521",
			Expected: elliptic.P521(),
		},
		{
			Name:     "Invalid",
			Input:    "521",
			Expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := EllipticCurveFromString(tc.Input)

			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func testMustBuildPrivateKey(b PrivateKeyBuilder) interface{} {
	k, err := b.Build()
	if err != nil {
		panic(err)
	}

	return k
}

func TestPublicKeyFromPrivateKey(t *testing.T) {
	testCases := []struct {
		Name       string
		PrivateKey interface{}
		Expected   interface{}
	}{
		{
			Name:       "RSA2048",
			PrivateKey: testMustBuildPrivateKey(RSAKeyBuilder{}.WithKeySize(512)),
			Expected:   &rsa.PublicKey{},
		},
		{
			Name:       "ECDSA-P256",
			PrivateKey: testMustBuildPrivateKey(ECDSAKeyBuilder{}.WithCurve(elliptic.P256())),
			Expected:   &ecdsa.PublicKey{},
		},
		{
			Name:       "Ed25519",
			PrivateKey: testMustBuildPrivateKey(Ed25519KeyBuilder{}),
			Expected:   ed25519.PublicKey{},
		},
		{
			Name:       "Invalid",
			PrivateKey: 8,
			Expected:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := PublicKeyFromPrivateKey(tc.PrivateKey)

			if tc.Expected == nil {
				assert.Nil(t, actual)
			} else {
				assert.IsType(t, tc.Expected, actual)
			}
		})
	}
}
