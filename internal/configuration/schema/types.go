package schema

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
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
	case "tcp", "udp", "http", "https":
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
