package utils

import (
	"crypto/elliptic"
	"crypto/tls"
	"runtime"
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
