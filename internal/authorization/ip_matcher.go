package authorization

import (
	"net"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func selectMatchingNetworkGroups(networks []string, aclNetworks []schema.ACLNetwork) []schema.ACLNetwork {
	var selectedNetworkGroups []schema.ACLNetwork

	for _, network := range networks {
		for _, n := range aclNetworks {
			if n.Name == network {
				selectedNetworkGroups = append(selectedNetworkGroups, n)
			}
		}
	}

	return selectedNetworkGroups
}

func isIPAddressOrCIDR(ip net.IP, network string) bool {
	switch {
	case ip.String() == network:
		return true
	case strings.Contains(network, "/"):
		return parseCIDR(ip, network)
	}

	return false
}

func parseCIDR(ip net.IP, network string) bool {
	_, ipNet, _ := net.ParseCIDR(network)
	return ipNet.Contains(ip)
}

// isIPMatching checks whether user's IP is in one of the network ranges.
func isIPMatching(ip net.IP, networks []string, aclNetworks []schema.ACLNetwork) bool {
	// If no network is provided in the rule, we match any network
	if len(networks) == 0 {
		return true
	}

	matchingNetworkGroups := selectMatchingNetworkGroups(networks, aclNetworks)

	for _, network := range networks {
		if net.ParseIP(network) == nil && !strings.Contains(network, "/") {
			for _, n := range matchingNetworkGroups {
				for _, network := range n.Networks {
					if isIPAddressOrCIDR(ip, network) {
						return true
					}
				}
			}
		} else if isIPAddressOrCIDR(ip, network) {
			return true
		}
	}

	return false
}
