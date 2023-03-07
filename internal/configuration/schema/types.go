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
	"syscall"
	"time"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-crypt/crypt/algorithm/plaintext"
)

// NewAddress returns an *Address and error depending on the ability to parse the string as an Address.
// It also assumes any value without a scheme which looks like a path is the 'unix' scheme, and everything else without
// a scheme is the 'tcp' scheme.
func NewAddress(value string) (address *Address, err error) {
	return NewAddressDefault(value, AddressSchemeTCP, AddressSchemeUnix)
}

// NewAddressDefault returns an *Address and error depending on the ability to parse the string as an Address.
// It also assumes any value without a scheme which looks like a path is the schemeDefaultPath scheme, and everything
// else without a scheme is the schemeDefault scheme.
func NewAddressDefault(value, schemeDefault, schemeDefaultPath string) (address *Address, err error) {
	if len(value) == 0 {
		return &Address{true, false, 0, &url.URL{Scheme: AddressSchemeTCP, Host: ":0"}}, nil
	}

	var u *url.URL

	if regexpHasScheme.MatchString(value) {
		u, err = url.Parse(value)
	} else {
		if strings.HasPrefix(value, "/") {
			u, err = url.Parse(fmt.Sprintf("%s://%s", schemeDefaultPath, value))
		} else {
			u, err = url.Parse(fmt.Sprintf("%s://%s", schemeDefault, value))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not parse string '%s' as address: expected format is [<scheme>://]<ip>[:<port>]: %w", value, err)
	}

	return NewAddressFromURL(u)
}

// NewAddressUnix returns an *Address from a path value.
func NewAddressUnix(path string) Address {
	return Address{true, true, 0, &url.URL{Scheme: AddressSchemeUnix, Path: path}}
}

// NewAddressFromNetworkValues returns an *Address from network values.
func NewAddressFromNetworkValues(network, host string, port int) Address {
	return Address{true, false, port, &url.URL{Scheme: network, Host: fmt.Sprintf("%s:%d", host, port)}}
}

// NewAddressFromURL returns an *Address and error depending on the ability to parse the *url.URL as an Address.
func NewAddressFromURL(u *url.URL) (addr *Address, err error) {
	addr = &Address{
		url: u,
	}

	if err = addr.validate(); err != nil {
		return nil, err
	}

	return addr, nil
}

// AddressTCP is just a type with an underlying type of Address.
type AddressTCP struct {
	Address
}

// AddressUDP is just a type with an underlying type of Address.
type AddressUDP struct {
	Address
}

// AddressLDAP is just a type with an underlying type of Address.
type AddressLDAP struct {
	Address
}

// AddressSMTP is just a type with an underlying type of Address.
type AddressSMTP struct {
	Address
}

// Address represents an address.
type Address struct {
	valid  bool
	socket bool
	port   int

	url *url.URL
}

// Valid returns true if the Address is valid.
func (a *Address) Valid() bool {
	return a.valid
}

// ValidateHTTP returns true if the Address is valid for a HTTP connection listener.
func (a *Address) ValidateHTTP() error {
	switch a.Scheme() {
	case AddressSchemeTCP, AddressSchemeTCP4, AddressSchemeTCP6, AddressSchemeUnix:
		break
	default:
		return fmt.Errorf("scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as '%s'", a.Scheme())
	}

	return nil
}

// ValidateListener returns true if the Address is valid for a connection listener.
func (a *Address) ValidateListener() error {
	switch a.Scheme() {
	case AddressSchemeTCP, AddressSchemeTCP4, AddressSchemeTCP6, AddressSchemeUDP, AddressSchemeUDP4, AddressSchemeUDP6, AddressSchemeUnix:
		break
	default:
		return fmt.Errorf("scheme must be one of 'tcp', 'tcp4', 'tcp6', 'udp', 'udp4', 'udp6', or 'unix' but is configured as '%s'", a.Scheme())
	}

	return nil
}

// ValidateSMTP returns true if the Address is valid for a remote SMTP connection opener.
func (a *Address) ValidateSMTP() error {
	switch a.Scheme() {
	case AddressSchemeSMTP, AddressSchemeSUBMISSION, AddressSchemeSUBMISSIONS:
		break
	default:
		return fmt.Errorf("scheme must be one of 'smtp', 'submission', or 'submissions' but is configured as '%s'", a.Scheme())
	}

	return nil
}

// ValidateLDAP returns true if the Address has a value Scheme for an LDAP connection opener.
func (a *Address) ValidateLDAP() error {
	switch a.Scheme() {
	case AddressSchemeLDAP, AddressSchemeLDAPS, AddressSchemeLDAPI:
		break
	default:
		return fmt.Errorf("scheme must be one of 'ldap', 'ldaps', or 'ldapi' but is configured as '%s'", a.Scheme())
	}

	return nil
}

// IsExplicitlySecure returns true if the address is an explicitly secure.
func (a *Address) IsExplicitlySecure() bool {
	switch a.Scheme() {
	case AddressSchemeSUBMISSIONS, AddressSchemeLDAPS:
		return true
	default:
		return false
	}
}

// String returns a string representation of the Address.
func (a *Address) String() string {
	if !a.valid || a.url == nil {
		return ""
	}

	return a.url.String()
}

// Network returns the Scheme() if it's appropriate for the net packages network arguments otherwise it returns tcp.
func (a *Address) Network() string {
	switch scheme := a.Scheme(); scheme {
	case AddressSchemeTCP, AddressSchemeTCP4, AddressSchemeTCP6, AddressSchemeUDP, AddressSchemeUDP4, AddressSchemeUDP6, AddressSchemeUnix:
		return scheme
	default:
		return AddressSchemeTCP
	}
}

// Scheme returns the *url.URL Scheme field.
func (a *Address) Scheme() string {
	if !a.valid || a.url == nil {
		return ""
	}

	return a.url.Scheme
}

// Hostname returns the output of the *url.URL Hostname func.
func (a *Address) Hostname() string {
	if !a.valid || a.url == nil {
		return ""
	}

	return a.url.Hostname()
}

// SocketHostname returns the correct hostname for a socket connection.
func (a *Address) SocketHostname() string {
	if !a.valid || a.url == nil {
		return ""
	}

	if a.socket {
		return a.url.Path
	}

	return a.url.Hostname()
}

// SetHostname sets the hostname preserving the port.
func (a *Address) SetHostname(hostname string) {
	if !a.valid || a.url == nil {
		return
	}

	if port := a.url.Port(); port == "" {
		a.url.Host = hostname
	} else {
		a.url.Host = fmt.Sprintf("%s:%s", hostname, port)
	}
}

// Port returns the port.
func (a *Address) Port() int {
	return a.port
}

// SetPort sets the port preserving the hostname.
func (a *Address) SetPort(port int) {
	if !a.valid || a.url == nil {
		return
	}

	a.setport(port)
}

// Host returns the *url.URL Host field.
func (a *Address) Host() string {
	if !a.valid || a.url == nil {
		return ""
	}

	return a.url.Host
}

// NetworkAddress returns a string representation of the Address with just the host and port.
func (a *Address) NetworkAddress() string {
	if !a.valid || a.url == nil {
		return ""
	}

	if a.socket {
		return a.url.Path
	}

	return a.url.Host
}

// Listener creates and returns a net.Listener.
func (a *Address) Listener() (net.Listener, error) {
	return a.listener()
}

// Dial creates and returns a dialed net.Conn.
func (a *Address) Dial() (net.Conn, error) {
	if a.url == nil {
		return nil, fmt.Errorf("address url is nil")
	}

	return net.Dial(a.Network(), a.NetworkAddress())
}

// ListenerWithUMask creates and returns a net.Listener with a temporary UMask if the scheme is `unix`.
func (a *Address) ListenerWithUMask(umask int) (ln net.Listener, err error) {
	if !a.socket {
		return a.listener()
	}

	if a.url == nil {
		return nil, fmt.Errorf("address url is nil")
	}

	umask = syscall.Umask(umask)

	ln, err = net.Listen(a.Network(), a.NetworkAddress())

	_ = syscall.Umask(umask)

	return ln, err
}

func (a *Address) listener() (net.Listener, error) {
	if a.url == nil {
		return nil, fmt.Errorf("address url is nil")
	}

	return net.Listen(a.Network(), a.NetworkAddress())
}

func (a *Address) setport(port int) {
	a.port = port
	a.url.Host = net.JoinHostPort(a.url.Hostname(), strconv.Itoa(port))
}

func (a *Address) validate() (err error) {
	if a.url == nil {
		return fmt.Errorf("error validating the address: address url was nil")
	}

	switch {
	case a.url.RawQuery != "":
		return fmt.Errorf("error validating the address: the url '%s' appears to have a query but this is not valid for addresses", a.url.String())
	case a.url.RawFragment != "", a.url.Fragment != "":
		return fmt.Errorf("error validating the address: the url '%s' appears to have a fragment but this is not valid for addresses", a.url.String())
	case a.url.User != nil:
		return fmt.Errorf("error validating the address: the url '%s' appears to have user info but this is not valid for addresses", a.url.String())
	}

	switch a.url.Scheme {
	case AddressSchemeUnix, AddressSchemeLDAPI:
		if err = a.validateUnixSocket(); err != nil {
			return err
		}
	case AddressSchemeTCP, AddressSchemeTCP4, AddressSchemeTCP6, AddressSchemeUDP, AddressSchemeUDP4, AddressSchemeUDP6:
		if err = a.validateTCPUDP(); err != nil {
			return err
		}
	case AddressSchemeLDAP, AddressSchemeLDAPS, AddressSchemeSMTP, AddressSchemeSUBMISSION, AddressSchemeSUBMISSIONS:
		if err = a.validateProtocol(); err != nil {
			return err
		}
	}

	a.valid = true

	return nil
}

func (a *Address) validateProtocol() (err error) {
	port := a.url.Port()

	switch port {
	case "":
		switch a.url.Scheme {
		case AddressSchemeLDAP:
			a.setport(389)
		case AddressSchemeLDAPS:
			a.setport(636)
		case AddressSchemeSMTP:
			a.setport(25)
		case AddressSchemeSUBMISSION:
			a.setport(587)
		case AddressSchemeSUBMISSIONS:
			a.setport(465)
		}
	default:
		actualPort, _ := strconv.Atoi(port)

		a.setport(actualPort)
	}

	return nil
}

func (a *Address) validateTCPUDP() (err error) {
	port := a.url.Port()

	switch port {
	case "":
		a.setport(0)
	default:
		actualPort, _ := strconv.Atoi(port)

		a.setport(actualPort)
	}

	return nil
}

func (a *Address) validateUnixSocket() (err error) {
	switch {
	case a.url.Path == "" && a.url.Scheme != AddressSchemeLDAPI:
		return fmt.Errorf("error validating the unix socket address: could not determine path from '%s'", a.url.String())
	case a.url.Host != "":
		return fmt.Errorf("error validating the unix socket address: the url '%s' appears to have a host but this is not valid for unix sockets: this may occur if you omit the leading forward slash from the socket path", a.url.String())
	}

	a.socket = true

	return nil
}

var cdecoder algorithm.DecoderRegister

// DecodePasswordDigest returns a new PasswordDigest if it can be decoded.
func DecodePasswordDigest(encodedDigest string) (digest *PasswordDigest, err error) {
	if cdecoder == nil {
		if cdecoder, err = crypt.NewDefaultDecoder(); err != nil {
			return nil, fmt.Errorf("failed to initialize decoder: %w", err)
		}

		if err = plaintext.RegisterDecoderPlainText(cdecoder); err != nil {
			return nil, fmt.Errorf("failed to initialize decoder: could not register the plaintext decoder: %w", err)
		}
	}

	var d algorithm.Digest

	if d, err = cdecoder.Decode(encodedDigest); err != nil {
		return nil, err
	}

	return &PasswordDigest{Digest: d}, nil
}

// PasswordDigest is a configuration type for the crypt.Digest.
type PasswordDigest struct {
	algorithm.Digest
}

// IsPlainText returns true if the underlying algorithm.Digest is a *plaintext.Digest.
func (d *PasswordDigest) IsPlainText() bool {
	if d == nil || d.Digest == nil {
		return false
	}

	switch d.Digest.(type) {
	case *plaintext.Digest:
		return true
	default:
		return false
	}
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
