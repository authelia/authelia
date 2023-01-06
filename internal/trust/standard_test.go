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
