package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidateTLSConfig(t *testing.T) {
	var (
		config, configDefault *schema.TLS
	)

	assert.EqualError(t, ValidateTLSConfig(config, configDefault), "must provide configDefault")

	configDefault = &schema.TLS{}

	assert.NoError(t, ValidateTLSConfig(config, configDefault))

	config = &schema.TLS{}

	assert.NoError(t, ValidateTLSConfig(config, configDefault))

	config.PrivateKey = keyRSA2048
	config.CertificateChain = certRSA4096

	assert.EqualError(t, ValidateTLSConfig(config, configDefault), "option 'certificate_chain' is invalid: provided certificate chain does not contain the public key for the private key provided")
}
