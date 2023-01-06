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

	// AddTrustedCertificate adds a trusted *x509.Certificate to this provider.
	AddTrustedCertificate(cert *x509.Certificate) (err error)

	// AddTrustedCertificateFromPath adds a trusted certificates from a path to the provider. If the path is a directory
	// the directory is scanned for .crt, .cer, and .pem files.
	AddTrustedCertificateFromPath(path string) (err error)

	// GetTrustedCertificates returns the trusted certificates for the provider.
	GetTrustedCertificates() (pool *x509.CertPool)

	// GetTLSConfiguration returns a *tls.Config when provided with a *schema.TLSConfig with the providers trusted certificates.
	GetTLSConfiguration(sconfig *schema.TLSConfig) (config *tls.Config)
}
