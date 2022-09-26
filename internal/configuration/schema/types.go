package schema

import (
	"bytes"
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

		if block.Type != pemBlockTypeCertificate {
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

// NewX509KeyPair is a wrapper around the tls.X509KeyPair function which splits the certificates from the private key.
func NewX509KeyPair(data string) (pair *X509KeyPair, err error) {
	var (
		certs []*pem.Block
		key   *pem.Block
	)

	bytesPEM := []byte(data)

	var block *pem.Block

	for {
		block, bytesPEM = pem.Decode(bytesPEM)

		switch {
		case block == nil:
			switch len(bytesPEM) {
			case 0:
				break
			default:
				return nil, fmt.Errorf("x509 key pair was provided invalid data: isn't an encoded PEM block")
			}
		case block.Type == pemBlockTypeCertificate:
			certs = append(certs, block)
		case block.Type == "PRIVATE KEY", strings.HasSuffix(block.Type, " PRIVATE KEY"):
			if key != nil {
				return nil, fmt.Errorf("x509 key pair was provided more than one private key")
			}

			key = block
		default:
			return nil, fmt.Errorf("x509 key pair must only contain certificates and private keys")
		}

		if len(bytesPEM) == 0 {
			break
		}
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("x509 key pair failed to decode: no certificate PEM block provided")
	}

	if key == nil {
		return nil, fmt.Errorf("x509 key pair failed to decode: no private key PEM block provided")
	}

	buf := &bytes.Buffer{}

	for _, cert := range certs {
		if err = pem.Encode(buf, cert); err != nil {
			return nil, fmt.Errorf("x509 key pair failed to encode certificate: %w", err)
		}
	}

	bufk := &bytes.Buffer{}

	if err = pem.Encode(bufk, key); err != nil {
		return nil, fmt.Errorf("x509 key pair failed to encode private key: %w", err)
	}

	var tpair tls.Certificate

	if tpair, err = tls.X509KeyPair(buf.Bytes(), bufk.Bytes()); err != nil {
		return nil, fmt.Errorf("x509 key pair failed to derive tls.Certificate: %w", err)
	}

	return &X509KeyPair{pair: tpair}, nil
}

// X509KeyPair is a tls.Certificate that is ensured to have a private key and certificate.
type X509KeyPair struct {
	pair tls.Certificate
}

// Certificate returns the certificate.
func (x *X509KeyPair) Certificate() tls.Certificate {
	return x.pair
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
	if !c.HasCertificates() {
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
func (c *X509CertificateChain) Certificates() []*x509.Certificate {
	return c.certs
}

// Validate the X509CertificateChain ensuring the certificates were provided in the correct order
// (with nth being signed by the nth+1), and that all of the certificates are valid based on the current time.
func (c *X509CertificateChain) Validate() (err error) {
	n := len(c.certs)
	now := time.Now()

	for i, cert := range c.certs {
		if (i == 0 || i == n) && !cert.BasicConstraintsValid {
			return fmt.Errorf("certificate #%d in chain is invalid as the basic constraints for the certificate are not valid", i+1)
		}

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

// NewTLSVersion parses TLS versions to the unit16 representation of the version and stores it as a *schema.TLSVersion.
func NewTLSVersion(input string) (version *TLSVersion, err error) {
	switch strings.ToUpper(input) {
	case "1.3", prefixTLS + vOneThree, prefixTLS + " " + vOneThree:
		return &TLSVersion{tls.VersionTLS13}, nil
	case "1.2", prefixTLS + vOneTwo, prefixTLS + " " + vOneTwo:
		return &TLSVersion{tls.VersionTLS12}, nil
	case "1.1", prefixTLS + vOneOne, prefixTLS + " " + vOneOne:
		return &TLSVersion{tls.VersionTLS11}, nil
	case "1.0", prefixTLS + vOneZero, prefixTLS + " " + vOneZero:
		return &TLSVersion{tls.VersionTLS10}, nil
	default:
		return nil, fmt.Errorf("tls version '%s' is unknown or unsupported", input)
	}
}

// TLSVersion represents a TLS version like tls.VersionTLS13.
type TLSVersion struct {
	value uint16
}

// Version returns the actual version.
func (v *TLSVersion) Version() uint16 {
	return v.value
}
