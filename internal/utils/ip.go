package utils

import (
	"net"
	"strings"
)

// ParseHostCIDR parses a raw string as a *net.IPNet similar to net.ParseCIDR, in fact it leverages it. The only
// differences between the functions is if the input does not contain a single '/' it first parses it with net.ParseIP
// to determine if it's a IPv4 or IPv6 and then adds the relevant CIDR suffix for a single host, and it only returns the
// *net.IPNet and error, discarding the net.IP.
func ParseHostCIDR(s string) (cidr *net.IPNet, err error) {
	switch strings.Count(s, "/") {
	case 1:
		_, cidr, err = net.ParseCIDR(s)

		return
	case 0:
		switch n := net.ParseIP(s); {
		case n == nil:
			return nil, &net.ParseError{Type: "CIDR address", Text: s}
		case n.To4() == nil:
			_, cidr, err = net.ParseCIDR(s + "/128")
		default:
			_, cidr, err = net.ParseCIDR(s + "/32")
		}

		return
	default:
		return nil, &net.ParseError{Type: "CIDR address", Text: s}
	}
}
