package utils

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
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
	assert.EqualError(t, err, "supplied TLS version isn't supported")

	version, err = TLSStringToTLSConfigVersion("SSL3.0")
	assert.Error(t, err)
	assert.Equal(t, uint16(0), version)
	assert.EqualError(t, err, "supplied TLS version isn't supported")
}

func TestShouldReturnErrWhenX509DirectoryNotExist(t *testing.T) {
	pool, errs, nonFatalErrs := NewX509CertPool("/tmp/asdfzyxabc123/not/a/real/dir")
	assert.NotNil(t, pool)
	assert.Len(t, nonFatalErrs, 0)
	require.Len(t, errs, 1)
	assert.EqualError(t, errs[0], "could not read certificates from directory open /tmp/asdfzyxabc123/not/a/real/dir: no such file or directory")
}

func TestShouldNotReturnErrWhenX509DirectoryExist(t *testing.T) {
	pool, errs, nonFatalErrs := NewX509CertPool("/tmp")
	assert.NotNil(t, pool)
	assert.Len(t, nonFatalErrs, 0)
	assert.Len(t, errs, 0)
}

func TestShouldRaiseNonFatalErrWhenNotifierTrustedCertConfigured(t *testing.T) {
	pool, errs, nonFatalErrs := NewX509CertPool("/tmp")
	assert.NotNil(t, pool)
	assert.Len(t, nonFatalErrs, 0)
	assert.Len(t, errs, 0)
}

func TestShouldReadCertsFromDirectoryButNotKeys(t *testing.T) {
	pool, errs, nonFatalErrs := NewX509CertPool("../suites/common/ssl/")
	assert.NotNil(t, pool)
	require.Len(t, errs, 1)
	assert.Len(t, nonFatalErrs, 0)
	assert.EqualError(t, errs[0], "could not import certificate key.pem")
}
