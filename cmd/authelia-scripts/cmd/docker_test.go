// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeTagShouldMakeBranchNamesSafeForDockerTags(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected string
	}{
		{"ShouldFoldTheTypeSeparator", "fix/oidc-claims", "fix-oidc-claims"},
		{"ShouldFoldEveryRunToOneSeparator", "renovate/docker-haproxy/3.x", "renovate-docker-haproxy-3.x"},
		{"ShouldPreserveAlreadySafeNames", "feat-oidc-dpop", "feat-oidc-dpop"},
		{"ShouldPreserveDotsAndUnderscores", "fix/v4.38_backport", "fix-v4.38_backport"},
		{"ShouldFoldWhitespaceAndOtherUnsafeRunes", "fix/oidc claims (#123)", "fix-oidc-claims-123-"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, sanitizeTag(tc.have))
		})
	}
}
