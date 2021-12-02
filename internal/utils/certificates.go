package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// NewTLSConfig generates a tls.Config from a schema.TLSConfig and a x509.CertPool.
func NewTLSConfig(config *schema.TLSConfig, defaultMinVersion uint16, certPool *x509.CertPool) (tlsConfig *tls.Config) {
	minVersion, err := TLSStringToTLSConfigVersion(config.MinimumVersion)
	if err != nil {
		minVersion = defaultMinVersion
	}

	return &tls.Config{
		ServerName:         config.ServerName,
		InsecureSkipVerify: config.SkipVerify, //nolint:gosec // Informed choice by user. Off by default.
		MinVersion:         minVersion,
		RootCAs:            certPool,
	}
}

// NewX509CertPool generates a x509.CertPool from the system PKI and the directory specified.
func NewX509CertPool(directory string) (certPool *x509.CertPool, warnings []error, errors []error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		warnings = append(warnings, fmt.Errorf("could not load system certificate pool which may result in untrusted certificate issues: %v", err))
		certPool = x509.NewCertPool()
	}

	logger := logging.Logger()

	logger.Tracef("Starting scan of directory %s for certificates", directory)

	if directory != "" {
		certsFileInfo, err := os.ReadDir(directory)
		if err != nil {
			errors = append(errors, fmt.Errorf("could not read certificates from directory %v", err))
		} else {
			for _, certFileInfo := range certsFileInfo {
				nameLower := strings.ToLower(certFileInfo.Name())

				if !certFileInfo.IsDir() && (strings.HasSuffix(nameLower, ".cer") || strings.HasSuffix(nameLower, ".crt") || strings.HasSuffix(nameLower, ".pem")) {
					certPath := filepath.Join(directory, certFileInfo.Name())

					logger.Tracef("Found possible cert %s, attempting to add it to the pool", certPath)

					certBytes, err := os.ReadFile(certPath)
					if err != nil {
						errors = append(errors, fmt.Errorf("could not read certificate %v", err))
					} else if ok := certPool.AppendCertsFromPEM(certBytes); !ok {
						errors = append(errors, fmt.Errorf("could not import certificate %s", certFileInfo.Name()))
					}
				}
			}
		}
	}

	logger.Tracef("Finished scan of directory %s for certificates", directory)

	return certPool, warnings, errors
}

// TLSStringToTLSConfigVersion returns a go crypto/tls version for a tls.Config based on string input.
func TLSStringToTLSConfigVersion(input string) (version uint16, err error) {
	switch strings.ToUpper(input) {
	case "TLS1.3", TLS13:
		return tls.VersionTLS13, nil
	case "TLS1.2", TLS12:
		return tls.VersionTLS12, nil
	case "TLS1.1", TLS11:
		return tls.VersionTLS11, nil
	case "TLS1.0", TLS10:
		return tls.VersionTLS10, nil
	}

	return 0, ErrTLSVersionNotSupported
}
