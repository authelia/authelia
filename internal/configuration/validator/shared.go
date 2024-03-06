package validator

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateTLSConfig sets the default values and validates a schema.TLS.
func ValidateTLSConfig(config *schema.TLS, configDefault *schema.TLS) (err error) {
	if configDefault == nil {
		return errors.New("must provide configDefault")
	}

	if config == nil {
		return
	}

	if config.ServerName == "" {
		config.ServerName = configDefault.ServerName
	}

	if config.MinimumVersion.Value == 0 {
		config.MinimumVersion.Value = configDefault.MinimumVersion.Value
	}

	if config.MaximumVersion.Value == 0 {
		config.MaximumVersion.Value = configDefault.MaximumVersion.Value
	}

	if config.MinimumVersion.MinVersion() < tls.VersionTLS10 {
		return errors.New("option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
	}

	if config.MinimumVersion.MinVersion() > config.MaximumVersion.MaxVersion() {
		return fmt.Errorf("option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version %s is greater than the maximum version %s", config.MinimumVersion.String(), config.MaximumVersion.String())
	}

	if config.CertificateChain.HasCertificates() && config.PrivateKey != nil && !config.CertificateChain.EqualKey(config.PrivateKey) {
		return errors.New("option 'certificate_chain' is invalid: provided certificate chain does not contain the public key for the private key provided")
	}

	return nil
}
