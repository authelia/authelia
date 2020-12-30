package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
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

//nolint:gocyclo // TODO: Remove in 4.28. Should be able to remove the nolint during the removal of deprecated config.
// NewX509CertPool generates a x509.CertPool from the system PKI and the directory specified.
func NewX509CertPool(directory string, config *schema.Configuration) (certPool *x509.CertPool, errors []error, nonFatalErrors []error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		nonFatalErrors = append(nonFatalErrors, fmt.Errorf("could not load system certificate pool which may result in untruested certificate issues: %v", err))
		certPool = x509.NewCertPool()
	}

	if directory != "" {
		certsFileInfo, err := ioutil.ReadDir(directory)
		if err != nil {
			errors = append(errors, fmt.Errorf("could not read certificates from directory %v", err))
		} else {
			for _, certFileInfo := range certsFileInfo {
				nameLower := strings.ToLower(certFileInfo.Name())

				if !certFileInfo.IsDir() && (strings.HasSuffix(nameLower, ".cer") || strings.HasSuffix(nameLower, ".pem")) {
					certBytes, err := ioutil.ReadFile(certFileInfo.Name())
					if err != nil {
						errors = append(errors, fmt.Errorf("could not read certificate %v", err))
					} else if ok := certPool.AppendCertsFromPEM(certBytes); !ok {
						errors = append(errors, fmt.Errorf("could not import certificate %s", certFileInfo.Name()))
					}
				}
			}
		}
	}

	// Deprecated. Maps deprecated values to the new ones. TODO: Remove in 4.28.
	if config != nil && config.Notifier != nil && config.Notifier.SMTP != nil && config.Notifier.SMTP.TrustedCert != "" {
		nonFatalErrors = append(nonFatalErrors, fmt.Errorf("defining the trusted cert in the SMTP notifier is deprecated and will be removed in 4.28.0, please use the global certificates_directory instead"))

		if exists, _ := FileExists(config.Notifier.SMTP.TrustedCert); exists {
			pem, err := ioutil.ReadFile(config.Notifier.SMTP.TrustedCert)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to read legacy SMTP trusted_cert (see the new certificates_directory option) certificate %s with error: %s", config.Notifier.SMTP.TrustedCert, err))
			} else if ok := certPool.AppendCertsFromPEM(pem); !ok {
				errors = append(errors, fmt.Errorf("could not import legacy SMTP trusted_cert (see the new certificates_directory option) certificate %s", config.Notifier.SMTP.TrustedCert))
			}
		} else {
			errors = append(errors, fmt.Errorf("could not import legacy SMTP trusted_cert (see the new certificates_directory option) certificate %s (file does not exist)", config.Notifier.SMTP.TrustedCert))
		}
	}

	return certPool, errors, nonFatalErrors
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
