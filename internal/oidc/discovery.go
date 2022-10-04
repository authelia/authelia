package oidc

import (
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewOpenIDConnectWellKnownConfiguration generates a new OpenIDConnectWellKnownConfiguration.
func NewOpenIDConnectWellKnownConfiguration(enablePKCEPlainChallenge bool, clients map[string]*Client, algs []string) (config OpenIDConnectWellKnownConfiguration) {
	config = OpenIDConnectWellKnownConfiguration{
		CommonDiscoveryOptions: CommonDiscoveryOptions{
			SubjectTypesSupported: []string{
				SubjectTypePublic,
			},
			ResponseTypesSupported: []string{
				"code",
				"token",
				"id_token",
				"code token",
				"code id_token",
				"token id_token",
				"code token id_token",
				"none",
			},
			ResponseModesSupported: []string{
				ResponseModeFormPost,
				ResponseModeQuery,
				ResponseModeFragment,
			},
			ScopesSupported: []string{
				ScopeOfflineAccess,
				ScopeOpenID,
				ScopeProfile,
				ScopeGroups,
				ScopeEmail,
			},
			ClaimsSupported: []string{
				ClaimAuthenticationMethodsReference,
				ClaimAudience,
				ClaimAuthorizedParty,
				ClaimClientIdentifier,
				ClaimExpirationTime,
				ClaimIssuedAt,
				ClaimIssuer,
				ClaimJWTID,
				ClaimRequestedAt,
				ClaimSubject,
				ClaimAuthenticationTime,
				ClaimNonce,
				ClaimPreferredEmail,
				ClaimEmailVerified,
				ClaimEmailAlts,
				ClaimGroups,
				ClaimPreferredUsername,
				ClaimFullName,
			},
		},
		OAuth2DiscoveryOptions: OAuth2DiscoveryOptions{
			CodeChallengeMethodsSupported: []string{
				PKCEChallengeMethodSHA256,
			},
		},
		OpenIDConnectDiscoveryOptions: OpenIDConnectDiscoveryOptions{
			IDTokenSigningAlgValuesSupported:  algs,
			UserinfoSigningAlgValuesSupported: append(algs, SigningAlgorithmNone),
			RequestObjectSigningAlgValuesSupported: []string{
				SigningAlgorithmNone,
				SigningAlgorithmRSAWithSHA256,
			},
		},
	}

	var pairwise, public bool

	for _, client := range clients {
		if pairwise && public {
			break
		}

		if client.SectorIdentifier != "" {
			pairwise = true
		}

		if !utils.IsStringInSlice(client.TokenEndpointAuthMethod, config.TokenEndpointAuthMethodsSupported) {
			config.TokenEndpointAuthMethodsSupported = append(config.TokenEndpointAuthMethodsSupported, client.TokenEndpointAuthMethod)
		}
	}

	if public {
		config.TokenEndpointAuthMethodsSupported = append(config.TokenEndpointAuthMethodsSupported, TokenEndpointAuthMethodNone)
	}

	if pairwise {
		config.SubjectTypesSupported = append(config.SubjectTypesSupported, SubjectTypePairwise)
	}

	if enablePKCEPlainChallenge {
		config.CodeChallengeMethodsSupported = append(config.CodeChallengeMethodsSupported, PKCEChallengeMethodPlain)
	}

	return config
}
