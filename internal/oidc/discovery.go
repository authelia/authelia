package oidc

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewOpenIDConnectWellKnownConfiguration generates a new OpenIDConnectWellKnownConfiguration.
func NewOpenIDConnectWellKnownConfiguration(c *schema.OpenIDConnectConfiguration, clients map[string]*Client) (config OpenIDConnectWellKnownConfiguration) {
	config = OpenIDConnectWellKnownConfiguration{
		CommonDiscoveryOptions: CommonDiscoveryOptions{
			SubjectTypesSupported: []string{
				SubjectTypePublic,
			},
			ResponseTypesSupported: []string{
				ResponseTypeAuthorizationCodeFlow,
				ResponseTypeImplicitFlowIDToken,
				ResponseTypeImplicitFlowToken,
				ResponseTypeImplicitFlowBoth,
				ResponseTypeHybridFlowIDToken,
				ResponseTypeHybridFlowToken,
				ResponseTypeHybridFlowBoth,
			},
			GrantTypesSupported: []string{
				GrantTypeAuthorizationCode,
				GrantTypeImplicit,
				GrantTypeRefreshToken,
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
			TokenEndpointAuthMethodsSupported: []string{
				ClientAuthMethodClientSecretBasic,
				ClientAuthMethodClientSecretPost,
				ClientAuthMethodNone,
			},
		},
		OAuth2DiscoveryOptions: OAuth2DiscoveryOptions{
			CodeChallengeMethodsSupported: []string{
				PKCEChallengeMethodSHA256,
			},
		},
		OpenIDConnectDiscoveryOptions: OpenIDConnectDiscoveryOptions{
			IDTokenSigningAlgValuesSupported: []string{
				SigningAlgorithmRSAWithSHA256,
			},
			UserinfoSigningAlgValuesSupported: []string{
				SigningAlgorithmNone,
				SigningAlgorithmRSAWithSHA256,
			},
			RequestObjectSigningAlgValuesSupported: []string{
				SigningAlgorithmNone,
				SigningAlgorithmRSAWithSHA256,
			},
		},
		PushedAuthorizationDiscoveryOptions: PushedAuthorizationDiscoveryOptions{
			RequirePushedAuthorizationRequests: c.PAR.Enforce,
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
	}

	if pairwise {
		config.SubjectTypesSupported = append(config.SubjectTypesSupported, SubjectTypePairwise)
	}

	if c.EnablePKCEPlainChallenge {
		config.CodeChallengeMethodsSupported = append(config.CodeChallengeMethodsSupported, PKCEChallengeMethodPlain)
	}

	return config
}
