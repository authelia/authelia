package utils

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHostCIDR(t *testing.T) {
	mustParse := func(in string) *net.IPNet {
		_, out, err := net.ParseCIDR(in)
		require.NoError(t, err)

		return out
	}

	testCases := []struct {
		name     string
		have     string
		expected *net.IPNet
		err      string
	}{
		{
			"ShouldParseIPv4",
			"192.168.1.1",
			mustParse("192.168.1.1/32"),
			"",
		},
		{
			"ShouldParseIPv4Zero",
			"0.0.0.0",
			mustParse("0.0.0.0/32"),
			"",
		},
		{
			"ShouldParseIPv4ZeroWithZeroCIDR",
			"0.0.0.0/0",
			mustParse("0.0.0.0/0"),
			"",
		},
		{
			"ShouldParseIPv6",
			"2001:db8:3333:4444:5555:6666:7777:8888",
			mustParse("2001:db8:3333:4444:5555:6666:7777:8888/128"),
			"",
		},
		{
			"ShouldParseIPv4WithCIDR",
			"192.168.1.1/24",
			mustParse("192.168.1.0/24"),
			"",
		},
		{
			"ShouldParseIPv6WithCIDR",
			"2001:db8:3333:4444:5555:6666:7777:8888/56",
			mustParse("2001:db8:3333:4400::/56"),
			"",
		},
		{
			"ShouldParseIPv6Zero",
			"::1",
			mustParse("::1/128"),
			"",
		},
		{
			"ShouldParseIPv6ZeroWithCIDR",
			"::1/0",
			mustParse("::1/0"),
			"",
		},
		{
			"ShouldNotParseMultipleSlashes",
			"192.168.1.1/24/0",
			nil,
			"invalid CIDR address: 192.168.1.1/24/0",
		},
		{
			"ShouldNotParseString",
			"abc",
			nil,
			"invalid CIDR address: abc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ParseHostCIDR(tc.have)
			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
