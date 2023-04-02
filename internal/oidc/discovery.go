package oidc

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewOpenIDConnectWellKnownConfiguration generates a new OpenIDConnectWellKnownConfiguration.
func NewOpenIDConnectWellKnownConfiguration(c *schema.OpenIDConnectConfiguration) (config OpenIDConnectWellKnownConfiguration) {
	config = OpenIDConnectWellKnownConfiguration{
		CommonDiscoveryOptions: CommonDiscoveryOptions{
			SubjectTypesSupported: []string{
				SubjectTypePublic,
				SubjectTypePairwise,
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
		},
		OpenIDConnectPromptCreateDiscoveryOptions: OpenIDConnectPromptCreateDiscoveryOptions{
			PromptValuesSupported: []string{
				PromptNone,
				PromptConsent,
			},
		},
		PushedAuthorizationDiscoveryOptions: PushedAuthorizationDiscoveryOptions{
			RequirePushedAuthorizationRequests: c.PAR.Enforce,
		},
	}

	if c.EnablePKCEPlainChallenge {
		config.CodeChallengeMethodsSupported = append(config.CodeChallengeMethodsSupported, PKCEChallengeMethodPlain)
	}

	return config
}
