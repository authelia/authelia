package schema

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-crypt/crypt"
)

// NewAddressFromString returns an *Address and error depending on the ability to parse the string as an Address.
func NewAddressFromString(a string) (addr *Address, err error) {
	if len(a) == 0 {
		return &Address{true, "tcp", net.ParseIP("0.0.0.0"), 0}, nil
	}

	var u *url.URL

	if regexpHasScheme.MatchString(a) {
		u, err = url.Parse(a)
	} else {
		u, err = url.Parse("tcp://" + a)
	}

	if err != nil {
		return nil, fmt.Errorf("could not parse string '%s' as address: expected format is [<scheme>://]<ip>[:<port>]: %w", a, err)
	}

	return NewAddressFromURL(u)
}

// NewAddressFromURL returns an *Address and error depending on the ability to parse the *url.URL as an Address.
func NewAddressFromURL(u *url.URL) (addr *Address, err error) {
	addr = &Address{
		Scheme: strings.ToLower(u.Scheme),
		IP:     net.ParseIP(u.Hostname()),
	}

	if addr.IP == nil {
		return nil, fmt.Errorf("could not parse ip for address '%s': %s does not appear to be an IP", u.String(), u.Hostname())
	}

	port := u.Port()
	switch port {
	case "":
		break
	default:
		addr.Port, err = strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("could not parse port for address '%s': %w", u.String(), err)
		}
	}

	switch addr.Scheme {
	case "tcp", "udp":
		break
	default:
		return nil, fmt.Errorf("could not parse scheme for address '%s': scheme '%s' is not valid, expected to be one of 'tcp://', 'udp://'", u.String(), addr.Scheme)
	}

	addr.valid = true

	return addr, nil
}

// Address represents an address.
type Address struct {
	valid bool

	Scheme string
	IP     net.IP
	Port   int
}

// Valid returns true if the Address is valid.
func (a Address) Valid() bool {
	return a.valid
}

// String returns a string representation of the Address.
func (a Address) String() string {
	if !a.valid {
		return ""
	}

	return fmt.Sprintf("%s://%s:%d", a.Scheme, a.IP.String(), a.Port)
}

// HostPort returns a string representation of the Address with just the host and port.
func (a Address) HostPort() string {
	if !a.valid {
		return ""
	}

	return fmt.Sprintf("%s:%d", a.IP.String(), a.Port)
}

// Listener creates and returns a net.Listener.
func (a Address) Listener() (net.Listener, error) {
	return net.Listen(a.Scheme, a.HostPort())
}

// NewPasswordDigest returns a new PasswordDigest.
func NewPasswordDigest(value string, plaintext bool) (digest *PasswordDigest, err error) {
	var d crypt.Digest

	switch {
	case plaintext:
		d, err = crypt.DecodeWithPlainText(value)
	default:
		d, err = crypt.Decode(value)
	}

	if err != nil {
		return nil, err
	}

	return &PasswordDigest{d}, err
}

// PasswordDigest is a configuration type for the crypt.Digest.
type PasswordDigest struct {
	crypt.Digest
}

// NewX509CertificateChain creates a new *X509CertificateChain from a given string, parsing each PEM block one by one.
func NewX509CertificateChain(in string) (chain *X509CertificateChain, err error) {
	if in == "" {
		return nil, nil
	}

	chain = &X509CertificateChain{
		certs: []*x509.Certificate{},
	}

	data := []byte(in)

	var (
		block *pem.Block
		cert  *x509.Certificate
	)

	for {
		block, data = pem.Decode(data)

		if block == nil || len(block.Bytes) == 0 {
			return nil, fmt.Errorf("invalid PEM block")
		}

		if block.Type != blockCERTIFICATE {
			return nil, fmt.Errorf("the PEM data chain contains a %s but only certificates are expected", block.Type)
		}

		if cert, err = x509.ParseCertificate(block.Bytes); err != nil {
			return nil, fmt.Errorf("the PEM data chain contains an invalid certificate: %w", err)
		}

		chain.certs = append(chain.certs, cert)

		if len(data) == 0 {
			break
		}
	}

	return chain, nil
}

// NewTLSVersion returns a new TLSVersion given a string.
func NewTLSVersion(input string) (version *TLSVersion, err error) {
	switch strings.ReplaceAll(strings.ToUpper(input), " ", "") {
	case TLSVersion13, Version13:
		return &TLSVersion{tls.VersionTLS13}, nil
	case TLSVersion12, Version12:
		return &TLSVersion{tls.VersionTLS12}, nil
	case TLSVersion11, Version11:
		return &TLSVersion{tls.VersionTLS11}, nil
	case TLSVersion10, Version10:
		return &TLSVersion{tls.VersionTLS10}, nil
	case SSLVersion30:
		return &TLSVersion{tls.VersionSSL30}, nil //nolint:staticcheck
	}

	return nil, ErrTLSVersionNotSupported
}

// TLSVersion is a struct which handles tls.Config versions.
type TLSVersion struct {
	Value uint16
}

// MaxVersion returns the value of this as a MaxVersion value.
func (v *TLSVersion) MaxVersion() uint16 {
	if v.Value == 0 {
		return tls.VersionTLS13
	}

	return v.Value
}

// MinVersion returns the value of this as a MinVersion value.
func (v *TLSVersion) MinVersion() uint16 {
	if v.Value == 0 {
		return tls.VersionTLS12
	}

	return v.Value
}

// String provides the Stringer.
func (v *TLSVersion) String() string {
	switch v.Value {
	case tls.VersionTLS10:
		return TLSVersion10
	case tls.VersionTLS11:
		return TLSVersion11
	case tls.VersionTLS12:
		return TLSVersion12
	case tls.VersionTLS13:
		return TLSVersion13
	case tls.VersionSSL30: //nolint:staticcheck
		return SSLVersion30
	default:
		return ""
	}
}

// CryptographicPrivateKey represents the actual crypto.PrivateKey interface.
type CryptographicPrivateKey interface {
	Public() crypto.PublicKey
	Equal(x crypto.PrivateKey) bool
}

// X509CertificateChain is a helper struct that holds a list of *x509.Certificate's.
type X509CertificateChain struct {
	certs []*x509.Certificate
}

// Thumbprint returns the Thumbprint for the first certificate.
func (c *X509CertificateChain) Thumbprint(hash crypto.Hash) []byte {
	if len(c.certs) == 0 {
		return nil
	}

	h := hash.New()

	h.Write(c.certs[0].Raw)

	return h.Sum(nil)
}

// HasCertificates returns true if the chain has any certificates.
func (c *X509CertificateChain) HasCertificates() (has bool) {
	return len(c.certs) != 0
}

// Equal checks if the provided *x509.Certificate is equal to the first *x509.Certificate in the chain.
func (c *X509CertificateChain) Equal(other *x509.Certificate) (equal bool) {
	if len(c.certs) == 0 {
		return false
	}

	return c.certs[0].Equal(other)
}

// EqualKey checks if the provided key (public or private) has a public key equal to the first public key in this chain.
//
//nolint:gocyclo // This is an adequately clear function even with the complexity.
func (c *X509CertificateChain) EqualKey(other any) (equal bool) {
	if len(c.certs) == 0 || other == nil {
		return false
	}

	switch key := other.(type) {
	case *rsa.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case rsa.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case *rsa.PrivateKey:
		return key.PublicKey.Equal(c.certs[0].PublicKey)
	case rsa.PrivateKey:
		return key.PublicKey.Equal(c.certs[0].PublicKey)
	case *ecdsa.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case ecdsa.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case *ecdsa.PrivateKey:
		return key.PublicKey.Equal(c.certs[0].PublicKey)
	case ecdsa.PrivateKey:
		return key.PublicKey.Equal(c.certs[0].PublicKey)
	case *ed25519.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case ed25519.PublicKey:
		return key.Equal(c.certs[0].PublicKey)
	case *ed25519.PrivateKey:
		switch pub := key.Public().(type) {
		case *ed25519.PublicKey:
			return pub.Equal(c.certs[0].PublicKey)
		case ed25519.PublicKey:
			return pub.Equal(c.certs[0].PublicKey)
		default:
			return false
		}
	case ed25519.PrivateKey:
		switch pub := key.Public().(type) {
		case *ed25519.PublicKey:
			return pub.Equal(c.certs[0].PublicKey)
		case ed25519.PublicKey:
			return pub.Equal(c.certs[0].PublicKey)
		default:
			return false
		}
	default:
		return false
	}
}

// Certificates for this X509CertificateChain.
func (c *X509CertificateChain) Certificates() (certificates []*x509.Certificate) {
	return c.certs
}

// CertificatesRaw for this X509CertificateChain.
func (c *X509CertificateChain) CertificatesRaw() (certificates [][]byte) {
	if !c.HasCertificates() {
		return nil
	}

	for _, cert := range c.certs {
		certificates = append(certificates, cert.Raw)
	}

	return certificates
}

// Leaf returns the first certificate if available for use with tls.Certificate.
func (c *X509CertificateChain) Leaf() (leaf *x509.Certificate) {
	if !c.HasCertificates() {
		return nil
	}

	return c.certs[0]
}

// Validate the X509CertificateChain ensuring the certificates were provided in the correct order
// (with nth being signed by the nth+1), and that all of the certificates are valid based on the current time.
func (c *X509CertificateChain) Validate() (err error) {
	n := len(c.certs)
	now := time.Now()

	for i, cert := range c.certs {
		if !cert.NotBefore.IsZero() && cert.NotBefore.After(now) {
			return fmt.Errorf("certificate #%d in chain is invalid before %d but the time is %d", i+1, cert.NotBefore.Unix(), now.Unix())
		}

		if cert.NotAfter.Before(now) {
			return fmt.Errorf("certificate #%d in chain is invalid after %d but the time is %d", i+1, cert.NotAfter.Unix(), now.Unix())
		}

		if i+1 >= n {
			break
		}

		if err = cert.CheckSignatureFrom(c.certs[i+1]); err != nil {
			return fmt.Errorf("certificate #%d in chain is not signed properly by certificate #%d in chain: %w", i+1, i+2, err)
		}
	}

	return nil
}
