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
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
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
		return nil, nil, fmt.Errorf("failed to convert certificate in DER format into PEM: %v", err)
	}

	keyDERBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %v", err)
	}

	keyPEMBytes, err := ConvertDERToPEM(keyDERBytes, PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert certificate in DER format into PEM: %v", err)
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

func publicKey(privateKey any) any {
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
	Build() (any, error)
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
func (rkb RSAKeyBuilder) Build() (any, error) {
	return rsa.GenerateKey(rand.Reader, rkb.keySizeInBits)
}

// Ed25519KeyBuilder builder of Ed25519 private key.
type Ed25519KeyBuilder struct{}

// Build an Ed25519 private key.
func (ekb Ed25519KeyBuilder) Build() (any, error) {
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
func (ekb ECDSAKeyBuilder) Build() (any, error) {
	return ecdsa.GenerateKey(ekb.curve, rand.Reader)
}

// ParseX509FromPEM parses PEM bytes and returns a PKCS key.
func ParseX509FromPEM(data []byte) (key any, err error) {
	var (
		block *pem.Block
		rest  []byte
	)

	if block, rest = pem.Decode(data); block == nil {
		return nil, errors.New("error occurred attempting to parse PEM block: either no PEM block was supplied or it was malformed")
	}

	if len(rest) != 0 {
		return nil, errors.New("error occurred attempting to parse PEM block: the block either had trailing data or was otherwise malformed")
	}

	return ParsePEMBlock(block)
}

// ParseX509FromPEMRecursive allows returning the appropriate key type given some PEM encoded input.
// For Keys this is a single value of one of *rsa.PrivateKey, *rsa.PublicKey, *ecdsa.PrivateKey, *ecdsa.PublicKey,
// ed25519.PrivateKey, or ed25519.PublicKey. For certificates this is
// either a *X509.Certificate, or a []*X509.Certificate.
func ParseX509FromPEMRecursive(data []byte) (decoded any, err error) {
	var (
		block        *pem.Block
		multi        bool
		certificates []*x509.Certificate
	)

	for i := 0; true; i++ {
		block, data = pem.Decode(data)

		n := len(data)

		switch {
		case block == nil:
			return nil, fmt.Errorf("error occurred attempting to parse PEM block: either no PEM block was supplied or it was malformed")
		case multi || n != 0:
			switch block.Type {
			case BlockTypeCertificate:
				var certificate *x509.Certificate

				if certificate, err = x509.ParseCertificate(block.Bytes); err != nil {
					return nil, fmt.Errorf("error occurred attempting to parse PEM block: data contains multiple blocks but #%d had an error during parsing: %w", i, err)
				}

				certificates = append(certificates, certificate)
			default:
				return nil, fmt.Errorf("error occurred attempting to parse PEM block: data contains multiple blocks but #%d has a '%s' block type and should have a '%s' block type", i, block.Type, BlockTypeCertificate)
			}

			multi = true
		default:
			if decoded, err = ParsePEMBlock(block); err != nil {
				return nil, err
			}
		}

		if n == 0 {
			break
		}
	}

	switch {
	case multi:
		return certificates, nil
	default:
		return decoded, nil
	}
}

// ParsePEMBlock parses a single PEM block into the relevant X509 data struct.
func ParsePEMBlock(block *pem.Block) (key any, err error) {
	if block == nil {
		return nil, errors.New("failed to parse PEM block as it was empty")
	}

	switch block.Type {
	case BlockTypeRSAPrivateKey:
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case BlockTypeECDSAPrivateKey:
		return x509.ParseECPrivateKey(block.Bytes)
	case BlockTypePKCS8PrivateKey:
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	case BlockTypeRSAPublicKey:
		return x509.ParsePKCS1PublicKey(block.Bytes)
	case BlockTypePKIXPublicKey:
		return x509.ParsePKIXPublicKey(block.Bytes)
	case BlockTypeCertificate:
		return x509.ParseCertificate(block.Bytes)
	case BlockTypeCertificateRequest:
		return x509.ParseCertificateRequest(block.Bytes)
	case BlockTypeX509CRL:
		return x509.ParseRevocationList(block.Bytes)
	default:
		switch {
		case strings.Contains(block.Type, "PRIVATE KEY"):
			return x509.ParsePKCS8PrivateKey(block.Bytes)
		case strings.Contains(block.Type, "PUBLIC KEY"):
			return x509.ParsePKIXPublicKey(block.Bytes)
		default:
			return nil, fmt.Errorf("unknown block type: %s", block.Type)
		}
	}
}

// AssertToX509Certificate converts an interface to an *x509.Certificate.
func AssertToX509Certificate(c any) (certificate *x509.Certificate, ok bool) {
	switch t := c.(type) {
	case x509.Certificate:
		return &t, true
	case *x509.Certificate:
		return t, true
	default:
		return nil, false
	}
}

// IsX509PrivateKey returns true if the provided interface is an rsa.PrivateKey, ecdsa.PrivateKey, or ed25519.PrivateKey.
func IsX509PrivateKey(i any) bool {
	switch i.(type) {
	case rsa.PrivateKey, *rsa.PrivateKey, ecdsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey, *ed25519.PrivateKey:
		return true
	default:
		return false
	}
}

// NewTLSConfig generates a tls.Config from a schema.TLS and a x509.CertPool.
func NewTLSConfig(config *schema.TLS, rootCAs *x509.CertPool) (tlsConfig *tls.Config) {
	if config == nil {
		return nil
	}

	var certificates []tls.Certificate

	if config.PrivateKey != nil && config.CertificateChain.HasCertificates() {
		certificates = []tls.Certificate{
			{
				Certificate: config.CertificateChain.CertificatesRaw(),
				Leaf:        config.CertificateChain.Leaf(),
				PrivateKey:  config.PrivateKey,
			},
		}
	}

	return &tls.Config{
		ServerName:         config.ServerName,
		InsecureSkipVerify: config.SkipVerify, //nolint:gosec // Informed choice by user. Off by default.
		MinVersion:         config.MinimumVersion.MinVersion(),
		MaxVersion:         config.MaximumVersion.MaxVersion(),
		RootCAs:            rootCAs,
		Certificates:       certificates,
	}
}

// NewX509CertPool generates a x509.CertPool from the system PKI and the directory specified using the standard factory.
func NewX509CertPool(directory string) (certPool *x509.CertPool, warnings []error, errors []error) {
	return NewX509CertPoolWithFactory(directory, &StandardX509SystemCertPoolFactory{})
}

// NewX509CertPoolWithFactory generates a x509.CertPool from the system PKI and the directory specified using a specific
// factory.
func NewX509CertPoolWithFactory(directory string, factory X509SystemCertPoolFactory) (certPool *x509.CertPool, warnings []error, errors []error) {
	if factory == nil {
		return nil, nil, []error{fmt.Errorf("failed to create x509 cert pool as no factory was provided")}
	}

	var err error
	if certPool, err = factory.SystemCertPool(); err != nil {
		warnings = append(warnings, fmt.Errorf("could not load system certificate pool which may result in untrusted certificate issues: %v", err))
		certPool = x509.NewCertPool()
	}

	log := logging.Logger()

	log.Tracef("Starting scan of directory %s for certificates", directory)

	if directory == "" {
		return certPool, warnings, errors
	}

	var entries []os.DirEntry

	if entries, err = os.ReadDir(directory); err != nil {
		errors = append(errors, fmt.Errorf("could not read certificates from directory %v", err))

		return certPool, warnings, errors
	}

	for _, entry := range entries {
		nameLower := strings.ToLower(entry.Name())

		if !entry.IsDir() && (strings.HasSuffix(nameLower, ".cer") || strings.HasSuffix(nameLower, ".crt") || strings.HasSuffix(nameLower, ".pem")) {
			certPath := filepath.Join(directory, entry.Name())

			log.Tracef("Found possible cert %s, attempting to add it to the pool", certPath)

			var data []byte

			if data, err = os.ReadFile(certPath); err != nil {
				errors = append(errors, fmt.Errorf("error occurred trying to read certificate: %w", err))
			} else if ok := certPool.AppendCertsFromPEM(data); !ok {
				errors = append(errors, fmt.Errorf("could not import certificate %s", entry.Name()))
			}
		}
	}

	log.Tracef("Finished scan of directory %s for certificates", directory)

	return certPool, warnings, errors
}

// WriteCertificateBytesAsPEMToPath writes a certificate/csr to a file in the PEM format.
func WriteCertificateBytesAsPEMToPath(path string, csr bool, certs ...[]byte) (err error) {
	var out *os.File

	if out, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return err
	}

	if err = WriteCertificateBytesAsPEMToWriter(out, csr, certs...); err != nil {
		_ = out.Close()

		return err
	}

	return out.Close()
}

// WriteCertificateBytesAsPEMToWriter writes a certificate/csr to a io.Writer in the PEM format.
func WriteCertificateBytesAsPEMToWriter(wr io.Writer, csr bool, certs ...[]byte) (err error) {
	blockType := BlockTypeCertificate
	if csr {
		blockType = BlockTypeCertificateRequest
	}

	blocks := make([]*pem.Block, len(certs))

	for i, cert := range certs {
		blocks[i] = &pem.Block{Type: blockType, Bytes: cert}
	}

	return WritePEMBlocksToWriter(wr, blocks...)
}

// WriteKeyToPEM writes a key that can be encoded as a PEM to a file in the PEM format.
func WriteKeyToPEM(key any, path string, legacy bool) (err error) {
	block, err := PEMBlockFromX509Key(key, legacy)
	if err != nil {
		return err
	}

	return WritePEMBlocksToPath(path, block)
}

// WritePEMBlocksToPath writes a set of *pem.Blocks to a file.
func WritePEMBlocksToPath(path string, blocks ...*pem.Block) (err error) {
	var out *os.File

	if out, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return err
	}

	if err = WritePEMBlocksToWriter(out, blocks...); err != nil {
		_ = out.Close()

		return err
	}

	return out.Close()
}

func WritePEMBlocksToWriter(w io.Writer, blocks ...*pem.Block) (err error) {
	for _, block := range blocks {
		if err = pem.Encode(w, block); err != nil {
			return err
		}
	}

	return nil
}

// PEMBlockFromX509Key turns a PublicKey or PrivateKey into a pem.Block.
func PEMBlockFromX509Key(key any, legacy bool) (block *pem.Block, err error) {
	var (
		data      []byte
		blockType string
	)

	switch k := key.(type) {
	case *rsa.PrivateKey:
		if legacy {
			blockType = BlockTypeRSAPrivateKey
			data = x509.MarshalPKCS1PrivateKey(k)

			break
		}

		blockType = BlockTypePKCS8PrivateKey
		data, err = x509.MarshalPKCS8PrivateKey(key)
	case *ecdsa.PrivateKey:
		if legacy {
			blockType = BlockTypeECDSAPrivateKey
			data, err = x509.MarshalECPrivateKey(k)

			break
		}

		blockType = BlockTypePKCS8PrivateKey
		data, err = x509.MarshalPKCS8PrivateKey(key)
	case ed25519.PrivateKey:
		blockType = BlockTypePKCS8PrivateKey
		data, err = x509.MarshalPKCS8PrivateKey(k)
	case *rsa.PublicKey:
		if legacy {
			blockType = BlockTypeRSAPublicKey
			data = x509.MarshalPKCS1PublicKey(k)

			break
		}

		blockType = BlockTypePKIXPublicKey
		data, err = x509.MarshalPKIXPublicKey(key)
	case *ecdsa.PublicKey, ed25519.PublicKey:
		blockType = BlockTypePKIXPublicKey
		data, err = x509.MarshalPKIXPublicKey(k)
	case *x509.Certificate:
		blockType = BlockTypeCertificate
		data = k.Raw
	case *x509.CertificateRequest:
		blockType = BlockTypeCertificateRequest
		data = k.Raw
	case *x509.RevocationList:
		blockType = BlockTypeX509CRL
		data = k.Raw
	default:
		err = fmt.Errorf("failed to match key type: %T", k)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	return &pem.Block{
		Type:    blockType,
		Headers: make(map[string]string),
		Bytes:   data,
	}, nil
}

// KeySigAlgorithmFromString returns a x509.PublicKeyAlgorithm and x509.SignatureAlgorithm given a keyAlgorithm and signatureAlgorithm string.
func KeySigAlgorithmFromString(keyAlgorithm, signatureAlgorithm string) (keyAlg x509.PublicKeyAlgorithm, sigAlg x509.SignatureAlgorithm) {
	keyAlg = PublicKeyAlgorithmFromString(keyAlgorithm)

	if keyAlg == x509.UnknownPublicKeyAlgorithm {
		return x509.UnknownPublicKeyAlgorithm, x509.UnknownSignatureAlgorithm
	}

	switch keyAlg {
	case x509.RSA:
		return keyAlg, RSASignatureAlgorithmFromString(signatureAlgorithm)
	case x509.ECDSA:
		return keyAlg, ECDSASignatureAlgorithmFromString(signatureAlgorithm)
	case x509.Ed25519:
		return keyAlg, x509.PureEd25519
	default:
		return keyAlg, x509.UnknownSignatureAlgorithm
	}
}

// PublicKeyAlgorithmFromString returns a x509.PublicKeyAlgorithm given an appropriate string.
func PublicKeyAlgorithmFromString(algorithm string) (alg x509.PublicKeyAlgorithm) {
	switch strings.ToUpper(algorithm) {
	case KeyAlgorithmRSA:
		return x509.RSA
	case KeyAlgorithmECDSA:
		return x509.ECDSA
	case KeyAlgorithmEd25519:
		return x509.Ed25519
	default:
		return x509.UnknownPublicKeyAlgorithm
	}
}

// RSASignatureAlgorithmFromString returns a x509.SignatureAlgorithm for the RSA x509.PublicKeyAlgorithm given an
// algorithm string.
func RSASignatureAlgorithmFromString(algorithm string) (alg x509.SignatureAlgorithm) {
	switch strings.ToUpper(algorithm) {
	case HashAlgorithmSHA1:
		return x509.SHA1WithRSA
	case HashAlgorithmSHA256:
		return x509.SHA256WithRSA
	case HashAlgorithmSHA384:
		return x509.SHA384WithRSA
	case HashAlgorithmSHA512:
		return x509.SHA512WithRSA
	default:
		return x509.UnknownSignatureAlgorithm
	}
}

// ECDSASignatureAlgorithmFromString returns a x509.SignatureAlgorithm for the ECDSA x509.PublicKeyAlgorithm given an
// algorithm string.
func ECDSASignatureAlgorithmFromString(algorithm string) (alg x509.SignatureAlgorithm) {
	switch strings.ToUpper(algorithm) {
	case HashAlgorithmSHA1:
		return x509.ECDSAWithSHA1
	case HashAlgorithmSHA256:
		return x509.ECDSAWithSHA256
	case HashAlgorithmSHA384:
		return x509.ECDSAWithSHA384
	case HashAlgorithmSHA512:
		return x509.ECDSAWithSHA512
	default:
		return x509.UnknownSignatureAlgorithm
	}
}

// EllipticCurveFromString turns a string into an elliptic.Curve.
func EllipticCurveFromString(curveString string) (curve elliptic.Curve) {
	switch strings.ToUpper(curveString) {
	case EllipticCurveAltP224, EllipticCurveP224:
		return elliptic.P224()
	case EllipticCurveAltP256, EllipticCurveP256:
		return elliptic.P256()
	case EllipticCurveAltP384, EllipticCurveP384:
		return elliptic.P384()
	case EllipticCurveAltP521, EllipticCurveP521:
		return elliptic.P521()
	default:
		return nil
	}
}

// PublicKeyFromPrivateKey returns a PublicKey when provided with a PrivateKey.
func PublicKeyFromPrivateKey(privateKey any) (publicKey any) {
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

// X509ParseKeyUsage parses a list of key usages. If provided with an empty list returns a default of Key Encipherment
// and Digital Signature unless ca is true in which case it returns Cert Sign.
func X509ParseKeyUsage(keyUsages []string, ca bool) (keyUsage x509.KeyUsage) {
	if len(keyUsages) == 0 {
		keyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		if ca {
			keyUsage |= x509.KeyUsageCertSign
		}

		return keyUsage
	}

	for _, keyUsageString := range keyUsages {
		switch strings.ToLower(keyUsageString) {
		case "digitalsignature", "digital_signature":
			keyUsage |= x509.KeyUsageDigitalSignature
		case "keyencipherment", "key_encipherment":
			keyUsage |= x509.KeyUsageKeyEncipherment
		case "dataencipherment", "data_encipherment":
			keyUsage |= x509.KeyUsageDataEncipherment
		case "keyagreement", "key_agreement":
			keyUsage |= x509.KeyUsageKeyAgreement
		case "certsign", "cert_sign", "certificatesign", "certificate_sign":
			keyUsage |= x509.KeyUsageCertSign
		case "crlsign", "crl_sign":
			keyUsage |= x509.KeyUsageCRLSign
		case "encipheronly", "encipher_only":
			keyUsage |= x509.KeyUsageEncipherOnly
		case "decipheronly", "decipher_only":
			keyUsage |= x509.KeyUsageDecipherOnly
		}
	}

	return keyUsage
}

// X509ParseExtendedKeyUsage parses a list of extended key usages. If provided with an empty list returns a default of
// Server Auth unless ca is true in which case it returns a default of Any.
func X509ParseExtendedKeyUsage(extKeyUsages []string, ca bool) (extKeyUsage []x509.ExtKeyUsage) {
	if len(extKeyUsages) == 0 {
		if ca {
			extKeyUsage = []x509.ExtKeyUsage{}
		} else {
			extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		}

		return extKeyUsage
	}

loop:
	for _, extKeyUsageString := range extKeyUsages {
		switch strings.ToLower(extKeyUsageString) {
		case "any":
			extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny}
			break loop
		case "serverauth", "server_auth":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageServerAuth)
		case "clientauth", "client_auth":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageClientAuth)
		case "codesigning", "code_signing":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageCodeSigning)
		case "emailprotection", "email_protection":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageEmailProtection)
		case "ipsecendsystem", "ipsec_endsystem", "ipsec_end_system":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageIPSECEndSystem)
		case "ipsectunnel", "ipsec_tunnel":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageIPSECTunnel)
		case "ipsecuser", "ipsec_user":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageIPSECUser)
		case "ocspsigning", "ocsp_signing":
			extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageOCSPSigning)
		}
	}

	return extKeyUsage
}

// TLSVersionFromBytesString converts a given 4 byte hexadecimal string into the appropriate TLS version.
func TLSVersionFromBytesString(input string) (version int, err error) {
	if n := len(input); n != 4 {
		return -1, fmt.Errorf("the input size was incorrect: should be 4 but was %d", n)
	}

	decoded, err := hex.DecodeString(input)
	if err != nil {
		return -1, fmt.Errorf("failed to decode hex: %w", err)
	}

	value := binary.BigEndian.Uint16(decoded)

	version = int(value)

	switch version {
	case tls.VersionSSL30: //nolint:staticcheck
		return tls.VersionSSL30, nil //nolint:staticcheck
	case tls.VersionTLS10:
		return tls.VersionTLS10, nil
	case tls.VersionTLS11:
		return tls.VersionTLS11, nil
	case tls.VersionTLS12:
		return tls.VersionTLS12, nil
	case tls.VersionTLS13:
		return tls.VersionTLS13, nil
	default:
		return -1, fmt.Errorf("tls version 0x%x is unknown", version)
	}
}

// IsInsecureCipherSuite returns true if a cipher suite is insecure.
func IsInsecureCipherSuite(cipherSuite uint16) bool {
	for _, suite := range tls.InsecureCipherSuites() {
		if suite.ID == cipherSuite {
			return true
		}
	}

	return false
}

// UnsafeGetIntermediatesFromPeerCertificates attempts to find valid intermediates from the provided peer certificates.
//
// CRITICAL: This function should not be used for production code as it may not produce the correct output to properly
// verify the chain. This function is intended to be used for testing purposes only.
func UnsafeGetIntermediatesFromPeerCertificates(certs []*x509.Certificate, roots, ints *x509.CertPool) (intermediates *x509.CertPool) {
	var err error

	n := len(certs) - 1

	opts := x509.VerifyOptions{}

	if roots != nil {
		opts.Roots = roots.Clone()
	}

	if ints != nil {
		opts.Intermediates = ints.Clone()
	}

	for i := n; i >= 0; i-- {
		if _, err = certs[i].Verify(opts); err == nil {
			continue
		}

		if i == n {
			// No certs in the chain are valid.
			break
		}

		// Intentionally only add the certificates within the trust chain.
		if certs[i+1].IsCA {
			if _, err = certs[i+1].Verify(opts); err == nil {
				opts.Intermediates.AddCert(certs[i+1])
			}
		}
	}

	return opts.Intermediates
}
