package trust

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// NewProvider returns a new provider.
func NewProvider(opts ...Opt) (provider *Production) {
	provider = &Production{
		mu:     &sync.Mutex{},
		config: Config{},
		log:    logging.Logger().WithFields(map[string]any{"service": "trust"}),
	}

	for _, opt := range opts {
		opt(provider)
	}

	return provider
}

// Production is a trust.Provider used for production operations. Should only be initialized via trust.NewProvider.
type Production struct {
	mu     *sync.Mutex
	log    *logrus.Entry
	config Config
	pool   *x509.CertPool
}

type Config struct {
	// List of paths to load certificates from.
	Paths []string

	// List of statically configured certificates to include.
	Static []*x509.Certificate

	// System allows trusting of system certificates.
	System bool

	// Invalid allows importing of expired or future certificates and instead just logs a warning.
	Invalid bool

	// Expired enforces checks on the expired status of certificates.
	Expired bool

	// Future enforces checks on the not yet valid status of certificates.
	Future bool
}

// StartupCheck implements the startup check provider interface.
func (t *Production) StartupCheck() (err error) {
	return t.reload()
}

// AddTrustedCertificate adds a trusted *x509.Certificate to this provider.
func (t *Production) AddTrustedCertificate(cert *x509.Certificate) (err error) {
	t.init()

	if cert == nil {
		return fmt.Errorf("certificate was not provided")
	}

	if err = t.validate(cert); err != nil {
		return err
	}

	t.mu.Lock()

	t.pool.AddCert(cert)

	t.mu.Unlock()

	return nil
}

// AddTrustedCertificatesFromPEM adds the *x509.Certificate content of a PEM block to this provider.
func (t *Production) AddTrustedCertificatesFromPEM(blocks []byte) (err error) {
	return nil
}

// AddTrustedCertificateFromPath adds a trusted certificates from a path to this provider. If the path is a directory
// the directory is scanned for .crt, .cer, and .pem files.
func (t *Production) AddTrustedCertificateFromPath(path string) (err error) {
	t.init()

	found, certs, err := t.read(path)
	if err != nil {
		return err
	}

	if found == 0 {
		return fmt.Errorf("no certificates found in path '%s'", path)
	}

	t.mu.Lock()

	pool := t.pool.Clone()

	t.mu.Unlock()

	for i := 0; i < len(certs); i++ {
		if err = t.validate(certs[i]); err != nil {
			return err
		}

		pool.AddCert(certs[i])
	}

	t.mu.Lock()

	t.pool = pool

	t.mu.Unlock()

	return nil
}

// GetTrustedCertificates returns the trusted certificates for this provider.
func (t *Production) GetTrustedCertificates() (pool *x509.CertPool) {
	t.init()

	t.mu.Lock()

	pool = t.pool.Clone()

	t.mu.Unlock()

	return pool
}

// GetTLSConfiguration returns a *tls.Config when provided with a *schema.TLSConfig with this providers trusted certificates.
func (t *Production) GetTLSConfiguration(sconfig *schema.TLSConfig) (config *tls.Config) {
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

	t.mu.Lock()

	rootCAs := t.pool.Clone()

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

func (t *Production) init() {
	if t.pool == nil {
		pool := t.new()

		t.mu.Lock()

		t.pool = pool

		t.mu.Unlock()
	}
}

func (t *Production) new() (pool *x509.CertPool) {
	var err error

	if !t.config.System {
		return x509.NewCertPool()
	}

	if pool, err = x509.SystemCertPool(); err != nil {
		t.log.WithError(err).Warnf("Error occurred loading the system certificate pool")

		pool = x509.NewCertPool()
	}

	return pool
}

func (t *Production) reload() (err error) {
	pool := t.new()

	t.log.Debug("Started load of trusted certificates pool")

	var (
		totalFound int
		totalCerts []*x509.Certificate
	)

	totalFound += len(t.config.Static)
	totalCerts = append(totalCerts, t.config.Static...)

	if len(t.config.Paths) == 0 {
		t.log.Trace("Skipping certificate scan of directories as none are defined")
	} else {
		for _, path := range t.config.Paths {
			var (
				found int
				certs []*x509.Certificate
			)

			if found, certs, err = t.read(path); err != nil {
				return err
			}

			if found == 0 {
				t.log.Tracef("No files found in scan of directory '%s' for potential additional trusted certificates", path)

				continue
			}

			totalCerts = append(totalCerts, certs...)

			totalFound += found
		}
	}

	t.log.WithField("found", totalFound).Debug("Finished load of trusted certificates pool")

	for i := 0; i < len(totalCerts); i++ {
		if err = t.validate(totalCerts[i]); err != nil {
			return err
		}

		pool.AddCert(totalCerts[i])
	}

	t.mu.Lock()

	t.pool = pool

	t.mu.Unlock()

	return nil
}

func (t *Production) read(name string) (found int, certs []*x509.Certificate, err error) {
	var info os.FileInfo

	if info, err = os.Stat(name); err != nil {
		return 0, nil, err
	}

	switch {
	case info.IsDir():
		return t.readFromDirectory(name)
	default:
		return t.readFromFile(name)
	}
}

func (t *Production) readFromFile(name string) (found int, certs []*x509.Certificate, err error) {
	ext := strings.ToLower(filepath.Ext(name))

	var (
		cert  *x509.Certificate
		block *pem.Block
		data  []byte
	)

	if data, err = os.ReadFile(name); err != nil {
		return 0, nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	for len(data) > 0 {
		if block, data = pem.Decode(data); block == nil {
			if len(certs) != 0 {
				break
			}

			return 0, nil, fmt.Errorf("failed to parse certificate: the file contained no PEM blocks")
		}

		if block.Type != "CERTIFICATE" {
			switch ext {
			case extPEM:
				continue
			default:
				return 0, nil, fmt.Errorf("failed to parse certificate PEM block: the PEM block is not a certificate, it's a '%s'", block.Type)
			}
		}

		if len(block.Headers) != 0 {
			return 0, nil, fmt.Errorf("failed to parse certificate PEM block: the PEM block has additional unexpected headers")
		}

		if cert, err = x509.ParseCertificate(block.Bytes); err != nil {
			return 0, nil, fmt.Errorf("failed to parse certificate PEM block: %w", err)
		}

		certs = append(certs, cert)
	}

	return len(certs), certs, nil
}

func (t *Production) readFromDirectory(name string) (found int, certs []*x509.Certificate, err error) {
	var entries []os.DirEntry

	if name, err = filepath.Abs(name); err != nil {
		return 0, nil, fmt.Errorf("could not determine absolute path: %w", err)
	}

	t.log.WithFields(map[string]any{"directory": name}).Debug("Starting certificate scan on directory")

	if entries, err = os.ReadDir(name); err != nil {
		return 0, nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))

		switch ext {
		case extCER, extCRT, extPEM:
			path := filepath.Join(name, entry.Name())

			t.log.WithFields(map[string]any{"directory": name, "name": entry.Name()}).Trace("Certificate scan on directory discovered a potential certificate")

			var (
				f int
				c []*x509.Certificate
			)

			if f, c, err = t.readFromFile(path); err != nil {
				return 0, nil, err
			}

			found += f

			certs = append(certs, c...)
		default:
			continue
		}
	}

	return found, certs, nil
}

/*
func (t *Production) load(dir string) (found int, certs []*x509.Certificate, err error) {
	var (
		entries []os.DirEntry
		data    []byte
	)

	t.log.WithFields(map[string]any{"directory": dir}).Debug("Starting certificate scan on directory")

	if entries, err = os.ReadDir(dir); err != nil {
		return found, nil, err
	}

	if len(entries) == 0 {
		t.log.WithFields(map[string]any{"directory": dir}).Trace("Finished certificate scan on empty directory")

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

			t.log.WithFields(map[string]any{"directory": dir, "name": entry.Name()}).Trace("Certificate scan on directory discovered a potential certificate")

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

	t.log.WithFields(map[string]any{"directory": dir, "found": found}).Debug("Finished certificate scan on directory")

	return found, certs, nil
}
*/

func (t *Production) validate(cert *x509.Certificate) (err error) {
	now := time.Now()

	if t.config.Expired && cert.NotAfter.Before(now) {
		switch {
		case t.config.Invalid:
			t.log.WithFields(map[string]any{"signature": string(cert.Signature), "common name": cert.Subject.CommonName, "expires": cert.NotAfter.Unix()}).Warn("Certificate which has expired was loaded")
		default:
			return fmt.Errorf("failed to load certificate which is expired with signature %s: not after %d (now is %d)", cert.Signature, cert.NotAfter.Unix(), now.Unix())
		}
	}

	if t.config.Future && !cert.NotBefore.IsZero() && cert.NotBefore.After(now) {
		switch {
		case t.config.Invalid:
			t.log.WithFields(map[string]any{"signature": string(cert.Signature), "common name": cert.Subject.CommonName, "not before": cert.NotBefore.Unix()}).Warn("Certificate which is only valid in the future was loaded")
		default:
			return fmt.Errorf("failed to load certificate which is not yet valid with signature %s: not before %d (now is %d)", cert.Signature, cert.NotBefore.Unix(), now.Unix())
		}
	}

	return nil
}
