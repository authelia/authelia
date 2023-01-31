package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCookieDomainValid(t *testing.T) {
	testCases := []struct {
		domain   string
		expected bool
	}{
		{"example.com", false},
		{".example.com", false},
		{"*.example.com", false},
		{"authelia.com", false},
		{"duckdns.org", true},
		{"example.duckdns.org", false},
		{"192.168.2.1", false},
		{"localhost", true},
		{"com", true},
		{"randomnada", true},
	}

	for _, tc := range testCases {
		name := "ShouldFail"

		if tc.expected {
			name = "ShouldPass"
		}

		t.Run(tc.domain, func(t *testing.T) {
			t.Run(name, func(t *testing.T) {
				assert.Equal(t, tc.expected, isCookieDomainAPublicSuffix(tc.domain))
			})
		})
	}
}
