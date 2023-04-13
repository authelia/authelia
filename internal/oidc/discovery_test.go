package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewOpenIDConnectWellKnownConfiguration(t *testing.T) {
	testCases := []struct {
		desc               string
		pkcePlainChallenge bool
		enforcePAR         bool
		clients            map[string]Client

		expectCodeChallengeMethodsSupported, expectSubjectTypesSupported []string
	}{
		{
			desc:                                "ShouldHaveChallengeMethodsS256ANDSubjectTypesSupportedPublic",
			pkcePlainChallenge:                  false,
			clients:                             map[string]Client{"a": &BaseClient{}},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256},
			expectSubjectTypesSupported:         []string{SubjectTypePublic, SubjectTypePairwise},
		},
		{
			desc:                                "ShouldHaveChallengeMethodsS256PlainANDSubjectTypesSupportedPublic",
			pkcePlainChallenge:                  true,
			clients:                             map[string]Client{"a": &BaseClient{}},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256, PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:         []string{SubjectTypePublic, SubjectTypePairwise},
		},
		{
			desc:                                "ShouldHaveChallengeMethodsS256ANDSubjectTypesSupportedPublicPairwise",
			pkcePlainChallenge:                  false,
			clients:                             map[string]Client{"a": &BaseClient{SectorIdentifier: "yes"}},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256},
			expectSubjectTypesSupported:         []string{SubjectTypePublic, SubjectTypePairwise},
		},
		{
			desc:                                "ShouldHaveChallengeMethodsS256PlainANDSubjectTypesSupportedPublicPairwise",
			pkcePlainChallenge:                  true,
			clients:                             map[string]Client{"a": &BaseClient{SectorIdentifier: "yes"}},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256, PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:         []string{SubjectTypePublic, SubjectTypePairwise},
		},
		{
			desc:                                "ShouldHaveTokenAuthMethodsNone",
			pkcePlainChallenge:                  true,
			clients:                             map[string]Client{"a": &BaseClient{SectorIdentifier: "yes"}},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256, PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:         []string{SubjectTypePublic, SubjectTypePairwise},
		},
		{
			desc:               "ShouldHaveTokenAuthMethodsNone",
			pkcePlainChallenge: true,
			clients: map[string]Client{
				"a": &BaseClient{SectorIdentifier: "yes"},
				"b": &BaseClient{SectorIdentifier: "yes"},
			},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256, PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:         []string{SubjectTypePublic, SubjectTypePairwise},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c := schema.OpenIDConnectConfiguration{
				EnablePKCEPlainChallenge: tc.pkcePlainChallenge,
				PAR: schema.OpenIDConnectPARConfiguration{
					Enforce: tc.enforcePAR,
				},
			}

			actual := NewOpenIDConnectWellKnownConfiguration(&c)
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
