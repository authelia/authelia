package authorization

import (
	"net"
	"strings"
)

// isIPMatching check whether user's IP is in one of the network ranges.
func isIPMatching(ip net.IP, networks []string) bool {
	// If no network is provided in the rule, we match any network
	if len(networks) == 0 {
		return true
	}

	for _, network := range networks {
		if !strings.Contains(network, "/") {
			if ip.String() == network {
				return true
			}
			continue
		}
		_, ipNet, err := net.ParseCIDR(network)
		if err != nil {
			// TODO(c.michaud): make sure the rule is valid at startup to
			// to such a case here.
			continue
		}

		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}
