package schema

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
)

var regexpAddress = regexp.MustCompile(`^((?P<Scheme>\w+)://)?((?P<IPv4>((((25[0-5]|2[0-4]\d|[01]?\d\d?)(\.)){3})(25[0-5]|2[0-4]\d|[01]?\d\d?)))|(\[(?P<IPv6>([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4})?:)?((25[0-5]|(2[0-4]|1?\d)?\d)\.){3}(25[0-5]|(2[0-4]|1?\d)?\d)|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1?\d)?\d)\.){3}(25[0-5]|(2[0-4]|1?\d)?\d))\]):)?(?P<Port>\d+)$`)

const tcp = "tcp"

// NewAddress produces a valid address from input.
func NewAddress(scheme string, ip net.IP, port int) Address {
	return Address{
		valid:  true,
		Scheme: scheme,
		IP:     ip,
		Port:   port,
	}
}

// NewAddressFromString parses a string and returns an *Address or error.
func NewAddressFromString(addr string) (address *Address, err error) {
	if addr == "" {
		return &Address{}, nil
	}

	if !regexpAddress.MatchString(addr) {
		return nil, fmt.Errorf("the string '%s' does not appear to be a valid address", addr)
	}

	address = &Address{
		valid: true,
	}

	submatches := regexpAddress.FindStringSubmatch(addr)

	var ip, port string

	for i, name := range regexpAddress.SubexpNames() {
		switch name {
		case "Scheme":
			address.Scheme = submatches[i]
		case "IPv4":
			ip = submatches[i]

			if address.Scheme == "" || address.Scheme == tcp {
				address.Scheme = "tcp4"
			}
		case "IPv6":
			ip = submatches[i]

			if address.Scheme == "" || address.Scheme == tcp {
				address.Scheme = "tcp6"
			}
		case "Port":
			port = submatches[i]
		}
	}

	if address.IP = net.ParseIP(ip); address.IP == nil {
		return nil, fmt.Errorf("failed to parse '%s' as an IP address", ip)
	}

	address.Port, _ = strconv.Atoi(port)

	if address.Port <= 0 || address.Port > 65535 {
		return nil, fmt.Errorf("failed to parse address port '%d' is invalid: ports must be between 1 and 65535", address.Port)
	}

	return address, nil
}

// Address represents an address.
type Address struct {
	valid bool

	Scheme string
	net.IP
	Port int
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
