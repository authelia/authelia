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
)

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
	dir := t.TempDir()

	pool, warnings, errors := NewX509CertPool(dir)
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
	pool, warnings, errors := NewX509CertPool("../suites/common/pki/")
	assert.NotNil(t, pool)
	require.Len(t, errors, 3)

	if runtime.GOOS == "windows" {
		require.Len(t, warnings, 1)
		assert.EqualError(t, warnings[0], "could not load system certificate pool which may result in untrusted certificate issues: crypto/x509: system root pool is not available on Windows")
	} else {
		assert.Len(t, warnings, 0)
	}

	assert.EqualError(t, errors[0], "could not import certificate private.backend.pem")
	assert.EqualError(t, errors[1], "could not import certificate private.oidc.pem")
	assert.EqualError(t, errors[2], "could not import certificate private.pem")
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

func testMustBuildPrivateKey(b PrivateKeyBuilder) any {
	k, err := b.Build()
	if err != nil {
		panic(err)
	}

	return k
}

func TestPublicKeyFromPrivateKey(t *testing.T) {
	testCases := []struct {
		Name       string
		PrivateKey any
		Expected   any
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

func TestX509ParseKeyUsage(t *testing.T) {
	testCases := []struct {
		name     string
		have     [][]string
		ca       bool
		expected x509.KeyUsage
	}{
		{
			"ShouldParseDefault", nil, false, x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		},
		{
			"ShouldParseDefaultCA", nil, true, x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		},
		{
			"ShouldParseDigitalSignature", [][]string{{"digital_signature"}, {"Digital_Signature"}, {"digitalsignature"}, {"digitalSignature"}}, false, x509.KeyUsageDigitalSignature,
		},
		{
			"ShouldParseKeyEncipherment", [][]string{{"key_encipherment"}, {"Key_Encipherment"}, {"keyencipherment"}, {"keyEncipherment"}}, false, x509.KeyUsageKeyEncipherment,
		},
		{
			"ShouldParseDataEncipherment", [][]string{{"data_encipherment"}, {"Data_Encipherment"}, {"dataencipherment"}, {"dataEncipherment"}}, false, x509.KeyUsageDataEncipherment,
		},
		{
			"ShouldParseKeyAgreement", [][]string{{"key_agreement"}, {"Key_Agreement"}, {"keyagreement"}, {"keyAgreement"}}, false, x509.KeyUsageKeyAgreement,
		},
		{
			"ShouldParseCertSign", [][]string{{"cert_sign"}, {"Cert_Sign"}, {"certsign"}, {"certSign"}, {"certificate_sign"}, {"Certificate_Sign"}, {"certificatesign"}, {"certificateSign"}}, false, x509.KeyUsageCertSign,
		},
		{
			"ShouldParseCRLSign", [][]string{{"crl_sign"}, {"CRL_Sign"}, {"crlsign"}, {"CRLSign"}}, false, x509.KeyUsageCRLSign,
		},
		{
			"ShouldParseEncipherOnly", [][]string{{"encipher_only"}, {"Encipher_Only"}, {"encipheronly"}, {"encipherOnly"}}, false, x509.KeyUsageEncipherOnly,
		},
		{
			"ShouldParseDecipherOnly", [][]string{{"decipher_only"}, {"Decipher_Only"}, {"decipheronly"}, {"decipherOnly"}}, false, x509.KeyUsageDecipherOnly,
		},
		{
			"ShouldParseMulti", [][]string{{"digitalSignature", "keyEncipherment", "dataEncipherment", "certSign", "crlSign"}}, false, x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.have) == 0 {
				actual := X509ParseKeyUsage(nil, tc.ca)

				assert.Equal(t, tc.expected, actual)
			}

			for _, have := range tc.have {
				t.Run(strings.Join(have, ","), func(t *testing.T) {
					actual := X509ParseKeyUsage(have, tc.ca)

					assert.Equal(t, tc.expected, actual)
				})
			}
		})
	}
}

func TestX509ParseExtendedKeyUsage(t *testing.T) {
	testCases := []struct {
		name     string
		have     [][]string
		ca       bool
		expected []x509.ExtKeyUsage
	}{
		{"ShouldParseDefault", nil, false, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}},
		{"ShouldParseDefaultCA", nil, true, []x509.ExtKeyUsage{}},
		{"ShouldParseAny", [][]string{{"any"}, {"Any"}, {"any", "server_auth"}}, false, []x509.ExtKeyUsage{x509.ExtKeyUsageAny}},
		{"ShouldParseServerAuth", [][]string{{"server_auth"}, {"Server_Auth"}, {"serverauth"}, {"serverAuth"}}, false, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}},
		{"ShouldParseClientAuth", [][]string{{"client_auth"}, {"Client_Auth"}, {"clientauth"}, {"clientAuth"}}, false, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}},
		{"ShouldParseCodeSigning", [][]string{{"code_signing"}, {"Code_Signing"}, {"codesigning"}, {"codeSigning"}}, false, []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning}},
		{"ShouldParseEmailProtection", [][]string{{"email_protection"}, {"Email_Protection"}, {"emailprotection"}, {"emailProtection"}}, false, []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection}},
		{"ShouldParseIPSECEndSystem", [][]string{{"ipsec_endsystem"}, {"IPSEC_Endsystem"}, {"ipsec_end_system"}, {"IPSEC_End_System"}, {"ipsecendsystem"}, {"ipsecEndSystem"}}, false, []x509.ExtKeyUsage{x509.ExtKeyUsageIPSECEndSystem}},
		{"ShouldParseIPSECTunnel", [][]string{{"ipsec_tunnel"}, {"IPSEC_Tunnel"}, {"ipsectunnel"}, {"ipsecTunnel"}}, false, []x509.ExtKeyUsage{x509.ExtKeyUsageIPSECTunnel}},
		{"ShouldParseIPSECUser", [][]string{{"ipsec_user"}, {"IPSEC_User"}, {"ipsecuser"}, {"ipsecUser"}}, false, []x509.ExtKeyUsage{x509.ExtKeyUsageIPSECUser}},
		{"ShouldParseOCSPSigning", [][]string{{"ocsp_signing"}, {"OCSP_Signing"}, {"ocspsigning"}, {"ocspSigning"}}, false, []x509.ExtKeyUsage{x509.ExtKeyUsageOCSPSigning}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.have) == 0 {
				actual := X509ParseExtendedKeyUsage(nil, tc.ca)

				assert.Equal(t, tc.expected, actual)
			}

			for _, have := range tc.have {
				t.Run(strings.Join(have, ","), func(t *testing.T) {
					actual := X509ParseExtendedKeyUsage(have, tc.ca)

					assert.Equal(t, tc.expected, actual)
				})
			}
		})
	}
}

func TestTLSVersionFromBytesString(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected int
		err      string
	}{
		{
			"ShouldDecodeSSL3.0",
			"0300",
			tls.VersionSSL30, //nolint:staticcheck
			"",
		},
		{
			"ShouldDecodeTLS1.0",
			"0301",
			tls.VersionTLS10,
			"",
		},
		{
			"ShouldDecodeTLS1.1",
			"0302",
			tls.VersionTLS11,
			"",
		},
		{
			"ShouldDecodeTLS1.2",
			"0303",
			tls.VersionTLS12,
			"",
		},
		{
			"ShouldDecodeTLS1.3",
			"0304",
			tls.VersionTLS13,
			"",
		},
		{
			"ShouldNotDecodeUnknownVersion",
			"ffff",
			-1,
			"tls version 0xffff is unknown",
		},
		{
			"ShouldNotDecodeUnknownLength",
			"ff",
			-1,
			"the input size was incorrect: should be 4 but was 2",
		},
		{
			"ShouldNotDecodeUnknownCharacters",
			"zzzz",
			-1,
			"failed to decode hex: encoding/hex: invalid byte: U+007A 'z'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := TLSVersionFromBytesString(tc.have)

			assert.Equal(t, tc.expected, actual)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}
