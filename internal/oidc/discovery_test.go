package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOpenIDConnectWellKnownConfiguration(t *testing.T) {
	testCases := []struct {
		desc               string
		pkcePlainChallenge bool
		clients            map[string]*Client

		expectCodeChallengeMethodsSupported, expectSubjectTypesSupported []string
		expectedTokenAuthMethods                                         []string
	}{
		{
			desc:                                "ShouldHaveChallengeMethodsS256ANDSubjectTypesSupportedPublic",
			pkcePlainChallenge:                  false,
			clients:                             map[string]*Client{"a": {}},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256},
			expectSubjectTypesSupported:         []string{SubjectTypePublic},
			expectedTokenAuthMethods:            []string{TokenEndpointAuthMethodClientSecretBasic},
		},
		{
			desc:                                "ShouldHaveChallengeMethodsS256PlainANDSubjectTypesSupportedPublic",
			pkcePlainChallenge:                  true,
			clients:                             map[string]*Client{"a": {}},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256, PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:         []string{SubjectTypePublic},
			expectedTokenAuthMethods:            []string{TokenEndpointAuthMethodClientSecretBasic},
		},
		{
			desc:                                "ShouldHaveChallengeMethodsS256ANDSubjectTypesSupportedPublicPairwise",
			pkcePlainChallenge:                  false,
			clients:                             map[string]*Client{"a": {SectorIdentifier: "yes"}},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256},
			expectSubjectTypesSupported:         []string{SubjectTypePublic, SubjectTypePairwise},
			expectedTokenAuthMethods:            []string{TokenEndpointAuthMethodClientSecretBasic},
		},
		{
			desc:                                "ShouldHaveChallengeMethodsS256PlainANDSubjectTypesSupportedPublicPairwise",
			pkcePlainChallenge:                  true,
			clients:                             map[string]*Client{"a": {SectorIdentifier: "yes"}},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256, PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:         []string{SubjectTypePublic, SubjectTypePairwise},
			expectedTokenAuthMethods:            []string{TokenEndpointAuthMethodClientSecretBasic},
		},
		{
			desc:                                "ShouldHaveTokenAuthMethodsNone",
			pkcePlainChallenge:                  true,
			clients:                             map[string]*Client{"a": {SectorIdentifier: "yes", Public: true}},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256, PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:         []string{SubjectTypePublic, SubjectTypePairwise},
			expectedTokenAuthMethods:            []string{TokenEndpointAuthMethodClientSecretBasic, TokenEndpointAuthMethodNone},
		},
		{
			desc:               "ShouldHaveTokenAuthMethodsNone",
			pkcePlainChallenge: true,
			clients: map[string]*Client{
				"a": {SectorIdentifier: "yes", Public: true},
				"b": {SectorIdentifier: "yes", Public: true},
			},
			expectCodeChallengeMethodsSupported: []string{PKCEChallengeMethodSHA256, PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:         []string{SubjectTypePublic, SubjectTypePairwise},
			expectedTokenAuthMethods:            []string{TokenEndpointAuthMethodClientSecretBasic, TokenEndpointAuthMethodNone},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := NewOpenIDConnectWellKnownConfiguration(tc.pkcePlainChallenge, tc.clients)
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
