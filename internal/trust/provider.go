package trust

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// CertificateProvider is the certificate trust provider implementation signature.
type CertificateProvider interface {
	model.StartupCheck

	// AddTrustedCertificate adds a trusted *x509.Certificate to this provider.
	AddTrustedCertificate(cert *x509.Certificate) (err error)

	// AddTrustedCertificatesFromBytes adds the *x509.Certificate content of a DER binary encoded block or PEM encoded
	// blocks to this provider.
	AddTrustedCertificatesFromBytes(data []byte) (err error)

	// AddTrustedCertificateFromPath adds a trusted certificates from a path to the provider. If the path is a directory
	// the directory is scanned for .crt, .cer, and .pem files.
	AddTrustedCertificateFromPath(path string) (err error)

	// GetCertPool returns the trusted certificates for the provider.
	GetCertPool() (pool *x509.CertPool)

	// NewTLSConfig returns a *tls.Config when provided with a *schema.TLSConfig and a *x509.CertPool.
	NewTLSConfig(c *schema.TLSConfig, rootCAs *x509.CertPool) (config *tls.Config)

	// GetTLSConfig returns a *tls.Config when provided with a *schema.TLSConfig with the providers trusted certificates.
	GetTLSConfig(sconfig *schema.TLSConfig) (config *tls.Config)
}
