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

// AddTrustedCertificate adds a trusted certificate to this provider.
func (t *StandardTrustProvider) AddTrustedCertificate(path string) (err error) {
	if t.pool == nil {
		t.pool = t.get()
	}

	var data []byte

	if data, err = os.ReadFile(path); err != nil {
		return fmt.Errorf("failed to read certificate: %w", err)
	}

	certs := loadPEMCertificates(data)

	if len(certs) == 0 {
		return fmt.Errorf("failed to read certificate: certificate at path '%s' does not contain PEM encoded certificate blocks", path)
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
	if t.pool == nil {
		t.mu.Lock()

		t.pool = t.get()

		t.mu.Unlock()
	}

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

	if t.pool == nil {
		t.mu.Lock()

		t.pool = t.get()

		t.mu.Unlock()
	}

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
		entries    []os.DirEntry
		data       []byte
		totalFound int
	)

	for _, dir := range t.dirs {
		if entries, err = os.ReadDir(dir); err != nil {
			return err
		}

		if len(entries) == 0 {
			t.log.Tracef("No files found in scan of directory '%s' for potential additional trusted certificates", dir)

			continue
		}

		var found int

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(entry.Name()))

			switch ext {
			case extCER, extCRT, extPEM:
				name := filepath.Join(dir, entry.Name())

				t.log.WithField("path", name).Tracef("Found possible certertificate, attempting to add it to the pool")

				if data, err = os.ReadFile(name); err != nil {
					return fmt.Errorf("failed to read certificate: %w", err)
				}

				certs := loadPEMCertificates(data)

				c := len(certs)

				if c == 0 {
					return fmt.Errorf("failed to read certificate: certificate at path '%s' does not contain PEM encoded certificate blocks", name)
				}

				found += c

				for i := 0; i < c; i++ {
					pool.AddCert(certs[i])
				}
			default:
				continue
			}
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
