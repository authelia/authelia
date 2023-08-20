package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidateTLSConfig(t *testing.T) {
	var (
		config, configDefault *schema.TLSConfig
	)

	assert.EqualError(t, ValidateTLSConfig(config, configDefault), "must provide configDefault")

	configDefault = &schema.TLSConfig{}

	assert.NoError(t, ValidateTLSConfig(config, configDefault))

	config = &schema.TLSConfig{}

	assert.NoError(t, ValidateTLSConfig(config, configDefault))

	config.PrivateKey = keyRSA2048
	config.CertificateChain = certRSA4096

	assert.EqualError(t, ValidateTLSConfig(config, configDefault), "option 'certificates' is invalid: provided certificate does not contain the public key for the private key provided")
}
