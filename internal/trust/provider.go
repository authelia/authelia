package trust

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// Provider is the trust provider implementation signature.
type Provider interface {
	model.StartupCheck

	// AddTrustedCertificate adds a trusted certificate to the provider.
	AddTrustedCertificate(path string) (err error)

	// GetTrustedCertificates returns the trusted certificates for the provider.
	GetTrustedCertificates() (pool *x509.CertPool)

	// GetTLSConfiguration returns a *tls.Config when provided with a *schema.TLSConfig with the providers trusted certificates.
	GetTLSConfiguration(sconfig *schema.TLSConfig) (config *tls.Config)
}
