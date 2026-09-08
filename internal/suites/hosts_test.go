//go:build !externalsuites
// +build !externalsuites

package suites

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostEntriesShouldFollowTheSuiteSubnet(t *testing.T) {
	testCases := []struct {
		name    string
		subnet  string
		portal  string
		backend string
	}{
		{"ShouldUseTheDefaultSubnetWhenUnset", "", "192.168.240.100", "192.168.240.50"},
		{"ShouldUseTheSlottedSubnetWhenSet", "10.240.2", "10.240.2.100", "10.240.2.50"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SUITE_SUBNET", tc.subnet)

			addresses := hostAddresses()

			assert.Equal(t, tc.portal, addresses["login.example.com"])
			assert.Equal(t, tc.portal, addresses["login.example2.com"])
			assert.Equal(t, tc.backend, addresses["authelia.example.com"])
			assert.Equal(t, "127.0.0.1", addresses["local.example.com"])
		})
	}
}

func TestHostResolverRulesShouldMapEveryEntry(t *testing.T) {
	t.Setenv("SUITE_SUBNET", "10.240.2")

	rules := HostResolverRules()

	assert.Contains(t, rules, "MAP login.example.com 10.240.2.100")
	assert.Contains(t, rules, "MAP proxy-client1.example.com 10.240.2.201")
	assert.Contains(t, rules, "MAP ssh.example.com 10.240.2.130")
	assert.Contains(t, rules, "MAP local.example.com 127.0.0.1")

	// One rule per entry, so no ordering rule of Chrome's has to be relied on.
	require.Len(t, strings.Split(rules, ","), len(HostEntries()))
}

func TestResolveAddrShouldSubstituteSuiteDomainsOnly(t *testing.T) {
	t.Setenv("SUITE_SUBNET", "10.240.2")

	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldResolveThePortal", "login.example.com:8080", "10.240.2.100:8080"},
		{"ShouldResolveTheBackend", "authelia.example.com:9091", "10.240.2.50:9091"},
		{"ShouldResolveTheSSHHost", "ssh.example.com:22", "10.240.2.130:22"},
		{"ShouldLeaveUnknownDomainsAlone", "example.org:443", "example.org:443"},
		{"ShouldLeaveAddressesWithoutAPortAlone", "login.example.com", "login.example.com"},
		{"ShouldLeaveLiteralAddressesAlone", "10.240.2.100:8080", "10.240.2.100:8080"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ResolveAddr(tc.have))
		})
	}
}
