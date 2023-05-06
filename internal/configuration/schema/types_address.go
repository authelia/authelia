package schema

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"syscall"
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
		return nil, fmt.Errorf("could not parse string '%s' as address: expected format is [<scheme>://]<hostname>[:<port>]: %w", value, err)
	}

	return NewAddressFromURL(u)
}

// NewAddressFromNetworkValuesDefault returns an *Address and error depending on the ability to parse the string as an Address.
// It also assumes any value without a scheme which looks like a path is the schemeDefaultPath scheme, and everything
// else without a scheme is the schemeDefault scheme.
func NewAddressFromNetworkValuesDefault(value string, port int, schemeDefault, schemeDefaultPath string) (address *Address, err error) {
	var u *url.URL

	if regexpHasScheme.MatchString(value) {
		u, err = url.Parse(value)
	} else {
		switch {
		case strings.HasPrefix(value, "/"):
			u, err = url.Parse(fmt.Sprintf("%s://%s", schemeDefaultPath, value))
		case port > 0:
			u, err = url.Parse(fmt.Sprintf("%s://%s:%d", schemeDefault, value, port))
		default:
			u, err = url.Parse(fmt.Sprintf("%s://%s", schemeDefault, value))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not parse string '%s' as address: expected format is [<scheme>://]<hostname>[:<port>]: %w", value, err)
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

// NewSMTPAddress returns an *AddressSMTP from SMTP values.
func NewSMTPAddress(scheme, host string, port int) *AddressSMTP {
	if port == 0 {
		switch scheme {
		case AddressSchemeSUBMISSIONS:
			port = 465
		case AddressSchemeSUBMISSION:
			port = 587
		default:
			port = 25
		}
	}

	if scheme == "" {
		switch port {
		case 465:
			scheme = AddressSchemeSUBMISSIONS
		case 587:
			scheme = AddressSchemeSUBMISSION
		default:
			scheme = AddressSchemeSMTP
		}
	}

	return &AddressSMTP{Address: Address{true, false, port, &url.URL{Scheme: scheme, Host: fmt.Sprintf("%s:%d", host, port)}}}
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

// IsUnixDomainSocket returns true if the address has been determined to be a Unix Domain Socket.
func (a *Address) IsUnixDomainSocket() bool {
	return a.socket
}

// IsTCP returns true if the address is one of the TCP schemes (not including application schemes that use TCP).
func (a *Address) IsTCP() bool {
	switch a.Scheme() {
	case AddressSchemeTCP, AddressSchemeTCP4, AddressSchemeTCP6:
		return true
	default:
		return false
	}
}

// IsUDP returns true if the address is one of the UDP schemes (not including application schemes that use UDP).
func (a *Address) IsUDP() bool {
	switch a.Scheme() {
	case AddressSchemeUDP, AddressSchemeUDP4, AddressSchemeUDP6:
		return true
	default:
		return false
	}
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

// ValidateHTTP returns true if the Address is valid for a HTTP connection listener.
func (a *Address) ValidateHTTP() error {
	if a.IsTCP() {
		return nil
	}

	switch a.Scheme() {
	case AddressSchemeUnix:
		return nil
	default:
		return fmt.Errorf("scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as '%s'", a.Scheme())
	}
}

// ValidateSMTP returns true if the Address is valid for a remote SMTP connection opener.
func (a *Address) ValidateSMTP() error {
	switch a.Scheme() {
	case AddressSchemeSMTP, AddressSchemeSUBMISSION, AddressSchemeSUBMISSIONS:
		return nil
	default:
		return fmt.Errorf("scheme must be one of 'smtp', 'submission', or 'submissions' but is configured as '%s'", a.Scheme())
	}
}

// ValidateSQL returns true if the Address is valid for a remote SQL connection opener.
func (a *Address) ValidateSQL() error {
	if a.IsTCP() {
		return nil
	}

	switch a.Scheme() {
	case AddressSchemeUnix:
		return nil
	default:
		return fmt.Errorf("scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as '%s'", a.Scheme())
	}
}

// ValidateLDAP returns true if the Address has a value Scheme for an LDAP connection opener.
func (a *Address) ValidateLDAP() error {
	switch a.Scheme() {
	case AddressSchemeLDAP, AddressSchemeLDAPS, AddressSchemeLDAPI:
		return nil
	default:
		return fmt.Errorf("scheme must be one of 'ldap', 'ldaps', or 'ldapi' but is configured as '%s'", a.Scheme())
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
