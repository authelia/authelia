package utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

// PEMBlockType represent an enum of the existing PEM block types.
type PEMBlockType int

const (
	// Certificate block type.
	Certificate PEMBlockType = iota
	// PrivateKey block type.
	PrivateKey
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

// GenerateCertificate generate a certificate given a private key. RSA, Ed25519 and ECDSA are officially supported.
func GenerateCertificate(privateKeyBuilder PrivateKeyBuilder, hosts []string, validFrom time.Time, validFor time.Duration, isCA bool) ([]byte, []byte, error) {
	privateKey, err := privateKeyBuilder.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to build private key: %w", err)
	}

	notBefore := validFrom
	notAfter := validFrom.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(privateKey), privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	certPEMBytes, err := ConvertDERToPEM(certDERBytes, Certificate)
	if err != nil {
		return nil, nil, fmt.Errorf("faile to convert certificate in DER format into PEM: %v", err)
	}

	keyDERBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %v", err)
	}

	keyPEMBytes, err := ConvertDERToPEM(keyDERBytes, PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("faile to convert certificate in DER format into PEM: %v", err)
	}

	return certPEMBytes, keyPEMBytes, nil
}

// ConvertDERToPEM convert certificate in DER format into PEM format.
func ConvertDERToPEM(der []byte, blockType PEMBlockType) ([]byte, error) {
	var buf bytes.Buffer

	var blockTypeStr string

	switch blockType {
	case Certificate:
		blockTypeStr = "CERTIFICATE"
	case PrivateKey:
		blockTypeStr = "PRIVATE KEY"
	default:
		return nil, fmt.Errorf("unknown PEM block type %d", blockType)
	}

	if err := pem.Encode(&buf, &pem.Block{Type: blockTypeStr, Bytes: der}); err != nil {
		return nil, fmt.Errorf("failed to encode DER data into PEM: %v", err)
	}

	return buf.Bytes(), nil
}

func publicKey(privateKey interface{}) interface{} {
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

// PrivateKeyBuilder interface for a private key builder.
type PrivateKeyBuilder interface {
	Build() (interface{}, error)
}

// RSAKeyBuilder builder of RSA private key.
type RSAKeyBuilder struct {
	keySizeInBits int
}

// WithKeySize configure the key size to use with RSA.
func (rkb RSAKeyBuilder) WithKeySize(bits int) RSAKeyBuilder {
	rkb.keySizeInBits = bits
	return rkb
}

// Build a RSA private key.
func (rkb RSAKeyBuilder) Build() (interface{}, error) {
	return rsa.GenerateKey(rand.Reader, rkb.keySizeInBits)
}

// Ed25519KeyBuilder builder of Ed25519 private key.
type Ed25519KeyBuilder struct{}

// Build an Ed25519 private key.
func (ekb Ed25519KeyBuilder) Build() (interface{}, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	return priv, err
}

// ECDSAKeyBuilder builder of ECDSA private key.
type ECDSAKeyBuilder struct {
	curve elliptic.Curve
}

// WithCurve configure the curve to use for the ECDSA private key.
func (ekb ECDSAKeyBuilder) WithCurve(curve elliptic.Curve) ECDSAKeyBuilder {
	ekb.curve = curve
	return ekb
}

// Build an ECDSA private key.
func (ekb ECDSAKeyBuilder) Build() (interface{}, error) {
	return ecdsa.GenerateKey(ekb.curve, rand.Reader)
}
