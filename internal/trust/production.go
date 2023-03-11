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

// NewProduction returns a new provider.
func NewProduction(opts ...ProductionOpt) (provider *Production) {
	provider = &Production{
		mu: &sync.Mutex{},
		config: Config{
			System:                 true,
			ValidationReturnErrors: true,
			ValidateNotAfter:       true,
			ValidateNotBefore:      true,
		},
		log: logging.Logger().WithFields(map[string]any{"service": "trust"}),
	}

	for _, opt := range opts {
		opt(provider)
	}

	return provider
}

// Production is a trust.CertificateProvider used for production operations. Should only be initialized via trust.NewProduction.
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

	// ValidationReturnErrors ensures that errors during validation are returned during validation. If disabled the
	// errors will instead be logged as warnings.
	ValidationReturnErrors bool

	// ValidateNotAfter enforces checks on the not after value of certificates.
	ValidateNotAfter bool

	// ValidateNotBefore enforces checks on the not before value of certificates.
	ValidateNotBefore bool
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

// AddTrustedCertificatesFromBytes adds the *x509.Certificate content of a DER binary encoded block or PEM encoded
// blocks to this provider.
func (t *Production) AddTrustedCertificatesFromBytes(data []byte) (err error) {
	t.init()

	found, certs, err := t.readFromBytes("", data)
	if err != nil {
		return err
	}

	if found == 0 {
		return fmt.Errorf("no certificates found in data")
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

// GetCertPool returns the trusted certificates for this provider.
func (t *Production) GetCertPool() (pool *x509.CertPool) {
	t.init()

	t.mu.Lock()

	defer t.mu.Unlock()

	return t.pool.Clone()
}

// GetTLSConfig returns a *tls.Config when provided with a *schema.TLSConfig with this providers trusted certificates.
func (t *Production) GetTLSConfig(c *schema.TLSConfig) (config *tls.Config) {
	if c == nil {
		return nil
	}

	t.init()

	t.mu.Lock()

	rootCAs := t.pool.Clone()

	t.mu.Unlock()

	return t.NewTLSConfig(c, rootCAs)
}

func (t *Production) NewTLSConfig(c *schema.TLSConfig, rootCAs *x509.CertPool) (config *tls.Config) {
	if c == nil {
		return nil
	}

	var certificates []tls.Certificate

	if c.PrivateKey != nil && c.CertificateChain.HasCertificates() {
		certificates = []tls.Certificate{
			{
				Certificate: c.CertificateChain.CertificatesRaw(),
				Leaf:        c.CertificateChain.Leaf(),
				PrivateKey:  c.PrivateKey,
			},
		}
	}

	return &tls.Config{
		ServerName:         c.ServerName,
		InsecureSkipVerify: c.SkipVerify, //nolint:gosec // Informed choice by user. Off by default.
		MinVersion:         c.MinimumVersion.MinVersion(),
		MaxVersion:         c.MaximumVersion.MaxVersion(),
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

func (t *Production) readFromBytes(ext string, data []byte) (found int, certs []*x509.Certificate, err error) {
	isPEM := ext == extPEM

	if !isPEM {
		if certs, err = x509.ParseCertificates(data); err == nil {
			return len(certs), certs, nil
		}
	}

	var (
		cert  *x509.Certificate
		block *pem.Block
	)

	for len(data) > 0 {
		if block, data = pem.Decode(data); block == nil {
			if len(certs) != 0 {
				break
			}

			return 0, nil, fmt.Errorf("failed to parse certificate: the file contained no PEM blocks and was not DER binary encoded")
		}

		if block.Type != "CERTIFICATE" {
			if isPEM {
				continue
			}

			return 0, nil, fmt.Errorf("failed to parse certificate PEM block: the PEM block is not a certificate, it's a '%s'", block.Type)
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

func (t *Production) readFromFile(name string) (found int, certs []*x509.Certificate, err error) {
	var (
		data []byte
	)

	if data, err = os.ReadFile(name); err != nil {
		return 0, nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	return t.readFromBytes(strings.ToLower(filepath.Ext(name)), data)
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

func (t *Production) validate(cert *x509.Certificate) (err error) {
	now := time.Now()

	if t.config.ValidateNotAfter && cert.NotAfter.Before(now) {
		switch {
		case t.config.ValidationReturnErrors:
			return fmt.Errorf("failed to load certificate which is expired with signature %s: not after %d (now is %d)", cert.Signature, cert.NotAfter.Unix(), now.Unix())
		default:
			t.log.WithFields(map[string]any{"signature": string(cert.Signature), "common name": cert.Subject.CommonName, "expires": cert.NotAfter.Unix()}).Warn("Certificate which has expired was loaded")
		}
	}

	if t.config.ValidateNotBefore && !cert.NotBefore.IsZero() && cert.NotBefore.After(now) {
		switch {
		case t.config.ValidationReturnErrors:
			return fmt.Errorf("failed to load certificate which is not yet valid with signature %s: not before %d (now is %d)", cert.Signature, cert.NotBefore.Unix(), now.Unix())
		default:
			t.log.WithFields(map[string]any{"signature": string(cert.Signature), "common name": cert.Subject.CommonName, "not before": cert.NotBefore.Unix()}).Warn("Certificate which is only valid in the future was loaded")
		}
	}

	return nil
}
