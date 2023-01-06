package trust

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// NewProvider returns a new provider.
func NewProvider(dirs ...string) *StandardTrustProvider {
	return &StandardTrustProvider{
		dirs: dirs,
		mu:   &sync.Mutex{},
		log:  logging.Logger(),
	}
}

// StandardTrustProvider is a trust.Provider used for production operations.
// Should only be initialized via trust.NewProvider.
type StandardTrustProvider struct {
	dirs []string
	mu   *sync.Mutex
	log  *logrus.Logger
	pool *x509.CertPool
}

// StartupCheck implements the startup check provider interface.
func (t *StandardTrustProvider) StartupCheck() (err error) {
	return t.reload()
}

// AddTrustedCertificate adds a trusted *x509.Certificate to this provider.
func (t *StandardTrustProvider) AddTrustedCertificate(cert *x509.Certificate) (err error) {
	t.init()

	if cert == nil {
		return fmt.Errorf("certificate was not provided")
	}

	t.mu.Lock()

	t.pool.AddCert(cert)

	t.mu.Unlock()

	return nil
}

// AddTrustedCertificatesFromPEM adds the *x509.Certificate content of a PEM block to this provider.
func (t *StandardTrustProvider) AddTrustedCertificatesFromPEM(blocks []byte) (err error) {
	return nil
}

// AddTrustedCertificateFromPath adds a trusted certificates from a path to this provider. If the path is a directory
// the directory is scanned for .crt, .cer, and .pem files.
func (t *StandardTrustProvider) AddTrustedCertificateFromPath(path string) (err error) {
	t.init()

	var (
		info os.FileInfo
		data []byte
	)

	if info, err = os.Stat(path); err != nil {
		return fmt.Errorf("failed to stat certificate: %w", err)
	}

	var certs []*x509.Certificate

	if info.IsDir() {
		var found int

		if found, certs, err = t.load(path); err != nil {
			return fmt.Errorf("failed to read certificate directory '%s': %w", path, err)
		}

		if found == 0 {
			return fmt.Errorf("failed to read certificate directory: directory at path '%s' does not contain any certificates", path)
		}
	} else {
		if data, err = os.ReadFile(path); err != nil {
			return fmt.Errorf("failed to read certificate: %w", err)
		}

		if certs, err = loadPEMCertificates(data); err != nil {
			return fmt.Errorf("failed to read certificate: certificate at path '%s': %w", path, err)
		}

		if len(certs) == 0 {
			return fmt.Errorf("failed to read certificate: certificate at path '%s' does not contain PEM encoded certificate blocks", path)
		}
	}

	t.mu.Lock()

	for i := 0; i < len(certs); i++ {
		t.pool.AddCert(certs[i])
	}

	t.mu.Unlock()

	return nil
}

// GetTrustedCertificates returns the trusted certificates for this provider.
func (t *StandardTrustProvider) GetTrustedCertificates() (pool *x509.CertPool) {
	t.init()

	pool = &x509.CertPool{}

	t.mu.Lock()

	*pool = *t.pool

	t.mu.Unlock()

	return pool
}

// GetTLSConfiguration returns a *tls.Config when provided with a *schema.TLSConfig with this providers trusted certificates.
func (t *StandardTrustProvider) GetTLSConfiguration(sconfig *schema.TLSConfig) (config *tls.Config) {
	if sconfig == nil {
		return nil
	}

	t.init()

	var certificates []tls.Certificate

	if sconfig.PrivateKey != nil && sconfig.CertificateChain.HasCertificates() {
		certificates = []tls.Certificate{
			{
				Certificate: sconfig.CertificateChain.CertificatesRaw(),
				Leaf:        sconfig.CertificateChain.Leaf(),
				PrivateKey:  sconfig.PrivateKey,
			},
		}
	}

	rootCAs := &x509.CertPool{}

	t.mu.Lock()

	*rootCAs = *t.pool

	t.mu.Unlock()

	return &tls.Config{
		ServerName:         sconfig.ServerName,
		InsecureSkipVerify: sconfig.SkipVerify, //nolint:gosec // Informed choice by user. Off by default.
		MinVersion:         sconfig.MinimumVersion.MinVersion(),
		MaxVersion:         sconfig.MaximumVersion.MaxVersion(),
		RootCAs:            rootCAs,
		Certificates:       certificates,
	}
}

func (t *StandardTrustProvider) init() {
	if t.pool == nil {
		t.mu.Lock()

		t.pool = t.get()

		t.mu.Unlock()
	}
}

func (t *StandardTrustProvider) get() (pool *x509.CertPool) {
	var err error

	if pool, err = x509.SystemCertPool(); err != nil {
		pool = x509.NewCertPool()

		t.log.WithError(err).Warnf("Error occurred loading the system certificate pool")
	}

	return pool
}

func (t *StandardTrustProvider) reload() (err error) {
	pool := t.get()

	if len(t.dirs) == 0 {
		t.mu.Lock()

		t.pool = pool

		t.mu.Unlock()

		t.log.Tracef("Skipping scan of certificate directories are not defined")

		return nil
	}

	t.log.Debugf("Starting scan of directories '%s' for potential additional trusted certificates", strings.Join(t.dirs, "', '"))

	var (
		totalFound int
	)

	for _, dir := range t.dirs {
		var (
			found int
			certs []*x509.Certificate
		)

		if found, certs, err = t.load(dir); err != nil {
			return err
		}

		if found == 0 {
			t.log.Tracef("No files found in scan of directory '%s' for potential additional trusted certificates", dir)

			continue
		}

		for i := 0; i < found; i++ {
			pool.AddCert(certs[i])
		}

		t.log.WithField("found", found).Tracef("Finished scan of directory '%s' for potential additional trusted certificates", dir)

		totalFound += found
	}

	t.log.WithField("found", totalFound).Debugf("Finished scan of directories '%s' for potential additional trusted certificates", strings.Join(t.dirs, "', '"))

	t.mu.Lock()

	t.pool = pool

	t.mu.Unlock()

	return nil
}

func (t *StandardTrustProvider) load(dir string) (found int, certs []*x509.Certificate, err error) {
	var (
		entries []os.DirEntry
		data    []byte
	)

	if entries, err = os.ReadDir(dir); err != nil {
		return found, nil, err
	}

	if len(entries) == 0 {
		t.log.Tracef("No files found in scan of directory '%s' for potential additional trusted certificates", dir)

		return 0, nil, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))

		switch ext {
		case extCER, extCRT, extPEM:
			path := filepath.Join(dir, entry.Name())

			t.log.WithField("path", path).Tracef("Found possible certertificate, attempting to add it to the pool")

			if data, err = os.ReadFile(path); err != nil {
				return 0, nil, fmt.Errorf("failed to read certificate: %w", err)
			}

			var loaded []*x509.Certificate

			if loaded, err = loadPEMCertificates(data); err != nil {
				return 0, nil, fmt.Errorf("failed to read certificate: certificate at path '%s': %w", path, err)
			}

			c := len(loaded)

			if c == 0 {
				return 0, nil, fmt.Errorf("failed to read certificate: certificate at path '%s' does not contain PEM encoded certificate blocks", path)
			}

			certs = append(certs, loaded...)

			found += c
		default:
			continue
		}
	}

	t.log.WithField("found", found).Tracef("Finished scan of directory '%s' for potential additional trusted certificates", dir)

	return found, certs, nil
}
