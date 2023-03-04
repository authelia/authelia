package trust

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProvider(t *testing.T) {
	provider := NewProvider()

	assert.NotNil(t, provider.GetTrustedCertificates())

	provider = NewProvider()

	assert.NoError(t, provider.StartupCheck())
	assert.NotNil(t, provider.GetTrustedCertificates())
}

func TestNewProvider_WithDirectories(t *testing.T) {
	provider := NewProvider(WithPaths("../suites/common/pki/"), WithSystem(true))

	pool := provider.GetTrustedCertificates()

	assert.NotNil(t, pool)

	assert.Equal(t, pool, provider.GetTrustedCertificates())

	assert.NoError(t, provider.StartupCheck())

	poolx := provider.GetTrustedCertificates()

	assert.NotEqual(t, pool, poolx)
}
