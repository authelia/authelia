package oidc

// NewOpenIDConnectWellKnownConfiguration generates a new OpenIDConnectWellKnownConfiguration.
func NewOpenIDConnectWellKnownConfiguration(enablePKCEPlainChallenge, pairwise bool) (config OpenIDConnectWellKnownConfiguration) {
	config = OpenIDConnectWellKnownConfiguration{
		CommonDiscoveryOptions: CommonDiscoveryOptions{
			SubjectTypesSupported: []string{
				"public",
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
				"form_post",
				"query",
				"fragment",
			},
			ScopesSupported: []string{
				ScopeOfflineAccess,
				ScopeOpenID,
				ScopeProfile,
				ScopeGroups,
				ScopeEmail,
			},
			ClaimsSupported: []string{
				"amr",
				"aud",
				"azp",
				"client_id",
				"exp",
				"iat",
				"iss",
				"jti",
				"rat",
				"sub",
				"auth_time",
				"nonce",
				ClaimEmail,
				ClaimEmailVerified,
				ClaimEmailAlts,
				ClaimGroups,
				ClaimPreferredUsername,
				ClaimDisplayName,
			},
		},
		OAuth2DiscoveryOptions: OAuth2DiscoveryOptions{
			CodeChallengeMethodsSupported: []string{
				"S256",
			},
		},
		OpenIDConnectDiscoveryOptions: OpenIDConnectDiscoveryOptions{
			IDTokenSigningAlgValuesSupported: []string{
				"RS256",
			},
			UserinfoSigningAlgValuesSupported: []string{
				"none",
				"RS256",
			},
			RequestObjectSigningAlgValuesSupported: []string{
				"none",
				"RS256",
			},
		},
	}

	if pairwise {
		config.SubjectTypesSupported = append(config.SubjectTypesSupported, "pairwise")
	}

	if enablePKCEPlainChallenge {
		config.CodeChallengeMethodsSupported = append(config.CodeChallengeMethodsSupported, "plain")
	}

	return config
}
