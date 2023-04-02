// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMACCoreStrategy_TrimPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		part     string
		expected string
	}{
		{"ShouldTrimAutheliaPrefix", "authelia_at_example", tokenPrefixPartAccessToken, "example"},
		{"ShouldTrimOryPrefix", "ory_at_example", tokenPrefixPartAccessToken, "example"},
		{"ShouldTrimOnlyAutheliaPrefix", "authelia_at_ory_at_example", tokenPrefixPartAccessToken, "ory_at_example"},
		{"ShouldTrimOnlyOryPrefix", "ory_at_authelia_at_example", tokenPrefixPartAccessToken, "authelia_at_example"},
		{"ShouldNotTrimGitHubPrefix", "gh_at_example", tokenPrefixPartAccessToken, "gh_at_example"},
	}

	strategy := &HMACCoreStrategy{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, strategy.trimPrefix(tc.have, tc.part))
		})
	}
}

func TestHMACCoreStrategy_GetSetPrefix(t *testing.T) {
	testCases := []struct {
		name        string
		have        string
		expectedSet string
		expectedGet string
	}{
		{"ShouldAddPrefix", "example", "authelia_%s_example", "authelia_%s_"},
	}

	strategy := &HMACCoreStrategy{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, part := range []string{tokenPrefixPartAccessToken, tokenPrefixPartAuthorizeCode, tokenPrefixPartRefreshToken} {
				t.Run(strings.ToUpper(part), func(t *testing.T) {
					assert.Equal(t, fmt.Sprintf(tc.expectedSet, part), strategy.setPrefix(tc.have, part))
					assert.Equal(t, fmt.Sprintf(tc.expectedGet, part), strategy.getPrefix(part))
				})
			}
		})
	}
}
