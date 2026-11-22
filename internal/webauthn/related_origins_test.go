// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelatedOriginLabeler(t *testing.T) {
	testCases := []struct {
		name     string
		origin   string
		expected string
		err      string
	}{
		{"ShouldLabelSimpleDomain", "https://example.com", "example", ""},
		{"ShouldLabelSubdomain", "https://auth.example.com", "example", ""},
		{"ShouldLabelMultiLabelPublicSuffix", "https://example.co.uk", "example", ""},
		{"ShouldLabelMultiLabelPublicSuffixSubdomain", "https://auth.example.com.au", "example", ""},
		{"ShouldLabelUpperCaseTrailingDot", "https://Auth.Example.COM.:8443", "example", ""},
		{"ShouldLabelIPv4", "https://192.168.1.1", "192.168.1.1", ""},
		{"ShouldLabelIPv6", "https://[::1]:9091", "::1", ""},
		{"ShouldLabelSingleLabel", "https://localhost", "localhost", ""},
		{"ShouldFallbackPublicSuffixOnly", "https://co.uk", "co", ""},
		{"ShouldErrorNoHost", "https://", "", "error determining the label of origin 'https://': the origin has no host component"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			label, err := RelatedOriginLabeler(tc.origin)

			if tc.err == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, label)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}
