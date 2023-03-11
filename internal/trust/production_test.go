package trust

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProvider(t *testing.T) {
	provider := NewProduction()

	assert.NotNil(t, provider.GetCertPool())

	provider = NewProduction()

	assert.NoError(t, provider.StartupCheck())
	assert.NotNil(t, provider.GetCertPool())
}

func TestNewProvider_WithDirectories(t *testing.T) {
	provider := NewProduction(WithCertificatePaths("../suites/common/pki/"))

	assert.NoError(t, provider.StartupCheck())
	assert.NotNil(t, provider.GetCertPool())
}
