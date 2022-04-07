package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOpenIDConnectWellKnownConfiguration(t *testing.T) {
	testCases := []struct {
		desc                                                             string
		pkcePlainChallenge, pairwise                                     bool
		expectCodeChallengeMethodsSupported, expectSubjectTypesSupported []string
	}{
		{
			desc:                                "ShouldHaveChallengeMethodsS256ANDSubjectTypesSupportedPublic",
			pkcePlainChallenge:                  false,
			pairwise:                            false,
			expectCodeChallengeMethodsSupported: []string{"S256"},
			expectSubjectTypesSupported:         []string{"public"},
		},
		{
			desc:                                "ShouldHaveChallengeMethodsS256PlainANDSubjectTypesSupportedPublic",
			pkcePlainChallenge:                  true,
			pairwise:                            false,
			expectCodeChallengeMethodsSupported: []string{"S256", "plain"},
			expectSubjectTypesSupported:         []string{"public"},
		},
		{
			desc:                                "ShouldHaveChallengeMethodsS256ANDSubjectTypesSupportedPublicPairwise",
			pkcePlainChallenge:                  false,
			pairwise:                            true,
			expectCodeChallengeMethodsSupported: []string{"S256"},
			expectSubjectTypesSupported:         []string{"public", "pairwise"},
		},
		{
			desc:                                "ShouldHaveChallengeMethodsS256PlainANDSubjectTypesSupportedPublicPairwise",
			pkcePlainChallenge:                  true,
			pairwise:                            true,
			expectCodeChallengeMethodsSupported: []string{"S256", "plain"},
			expectSubjectTypesSupported:         []string{"public", "pairwise"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := NewOpenIDConnectWellKnownConfiguration(tc.pkcePlainChallenge, tc.pairwise)
			for _, codeChallengeMethod := range tc.expectCodeChallengeMethodsSupported {
				assert.Contains(t, actual.CodeChallengeMethodsSupported, codeChallengeMethod)
			}

			for _, subjectType := range tc.expectSubjectTypesSupported {
				assert.Contains(t, actual.SubjectTypesSupported, subjectType)
			}

			for _, codeChallengeMethod := range actual.CodeChallengeMethodsSupported {
				assert.Contains(t, tc.expectCodeChallengeMethodsSupported, codeChallengeMethod)
			}

			for _, subjectType := range actual.SubjectTypesSupported {
				assert.Contains(t, tc.expectSubjectTypesSupported, subjectType)
			}
		})
	}
}
