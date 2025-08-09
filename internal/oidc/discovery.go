package oidc

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewOpenIDConnectWellKnownConfiguration generates a new OpenIDConnectWellKnownConfiguration.
func NewOpenIDConnectWellKnownConfiguration(c *schema.IdentityProvidersOpenIDConnect) (config OpenIDConnectWellKnownConfiguration) {
	config = OpenIDConnectWellKnownConfiguration{
		OAuth2WellKnownConfiguration: OAuth2WellKnownConfiguration{
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
					GrantTypeClientCredentials,
					GrantTypeRefreshToken,
					GrantTypeDeviceCode,
				},
				ResponseModesSupported: []string{
					ResponseModeFormPost,
					ResponseModeQuery,
					ResponseModeFragment,
					ResponseModeJWT,
					ResponseModeFormPostJWT,
					ResponseModeQueryJWT,
					ResponseModeFragmentJWT,
				},
				ScopesSupported: []string{
					ScopeOfflineAccess,
					ScopeOpenID,
					ScopeProfile,
					ScopeEmail,
					ScopeAddress,
					ScopePhone,
					ScopeGroups,
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
					ClaimAuthenticationTime,
					ClaimNonce,
					ClaimGroups,
					ClaimSubject,
					ClaimFullName,
					ClaimGivenName,
					ClaimFamilyName,
					ClaimMiddleName,
					ClaimNickname,
					ClaimPreferredUsername,
					ClaimProfile,
					ClaimPicture,
					ClaimWebsite,
					ClaimEmail,
					ClaimEmailVerified,
					ClaimEmailAlts,
					ClaimGender,
					ClaimBirthdate,
					ClaimZoneinfo,
					ClaimLocale,
					ClaimPhoneNumber,
					ClaimPhoneNumberVerified,
					ClaimAddress,
					ClaimUpdatedAt,
				},
				TokenEndpointAuthMethodsSupported: []string{
					ClientAuthMethodClientSecretBasic,
					ClientAuthMethodClientSecretPost,
					ClientAuthMethodClientSecretJWT,
					ClientAuthMethodPrivateKeyJWT,
					ClientAuthMethodNone,
				},
				TokenEndpointAuthSigningAlgValuesSupported: []string{
					SigningAlgHMACUsingSHA256,
					SigningAlgHMACUsingSHA384,
					SigningAlgHMACUsingSHA512,
					SigningAlgRSAUsingSHA256,
					SigningAlgRSAUsingSHA384,
					SigningAlgRSAUsingSHA512,
					SigningAlgECDSAUsingP256AndSHA256,
					SigningAlgECDSAUsingP384AndSHA384,
					SigningAlgECDSAUsingP521AndSHA512,
					SigningAlgRSAPSSUsingSHA256,
					SigningAlgRSAPSSUsingSHA384,
					SigningAlgRSAPSSUsingSHA512,
				},
			},
			OAuth2DiscoveryOptions: OAuth2DiscoveryOptions{
				CodeChallengeMethodsSupported: []string{
					PKCEChallengeMethodSHA256,
				},
				RevocationEndpointAuthMethodsSupported: []string{
					ClientAuthMethodClientSecretBasic,
					ClientAuthMethodClientSecretPost,
					ClientAuthMethodClientSecretJWT,
					ClientAuthMethodPrivateKeyJWT,
					ClientAuthMethodNone,
				},
				RevocationEndpointAuthSigningAlgValuesSupported: []string{
					SigningAlgHMACUsingSHA256,
					SigningAlgHMACUsingSHA384,
					SigningAlgHMACUsingSHA512,
					SigningAlgRSAUsingSHA256,
					SigningAlgRSAUsingSHA384,
					SigningAlgRSAUsingSHA512,
					SigningAlgECDSAUsingP256AndSHA256,
					SigningAlgECDSAUsingP384AndSHA384,
					SigningAlgECDSAUsingP521AndSHA512,
					SigningAlgRSAPSSUsingSHA256,
					SigningAlgRSAPSSUsingSHA384,
					SigningAlgRSAPSSUsingSHA512,
				},
				IntrospectionEndpointAuthMethodsSupported: []string{
					ClientAuthMethodClientSecretBasic,
					ClientAuthMethodClientSecretPost,
					ClientAuthMethodClientSecretJWT,
					ClientAuthMethodPrivateKeyJWT,
				},
			},
			OAuth2DeviceAuthorizationGrantDiscoveryOptions: &OAuth2DeviceAuthorizationGrantDiscoveryOptions{},
			OAuth2JWTIntrospectionResponseDiscoveryOptions: &OAuth2JWTIntrospectionResponseDiscoveryOptions{
				IntrospectionSigningAlgValuesSupported: []string{
					SigningAlgHMACUsingSHA256,
					SigningAlgHMACUsingSHA384,
					SigningAlgHMACUsingSHA512,
					SigningAlgRSAUsingSHA256,
					SigningAlgRSAUsingSHA384,
					SigningAlgRSAUsingSHA512,
					SigningAlgECDSAUsingP256AndSHA256,
					SigningAlgECDSAUsingP384AndSHA384,
					SigningAlgECDSAUsingP521AndSHA512,
					SigningAlgRSAPSSUsingSHA256,
					SigningAlgRSAPSSUsingSHA384,
					SigningAlgRSAPSSUsingSHA512,
					SigningAlgNone,
				},
				IntrospectionEncryptionAlgValuesSupported: []string{
					EncryptionAlgRSA15,
					EncryptionAlgRSAOAEP,
					EncryptionAlgRSAOAEP256,
					EncryptionAlgA128KW,
					EncryptionAlgA192KW,
					EncryptionAlgA256KW,
					EncryptionAlgDirect,
					EncryptionAlgECDHES,
					EncryptionAlgECDHESA128KW,
					EncryptionAlgECDHESA192KW,
					EncryptionAlgECDHESA256KW,
					EncryptionAlgA128GCMKW,
					EncryptionAlgA192GCMKW,
					EncryptionAlgA256GCMKW,
					EncryptionAlgPBES2HS256A128KW,
					EncryptionAlgPBES2HS284A192KW,
					EncryptionAlgPBES2HS512A256KW,
				},
				IntrospectionEncryptionEncValuesSupported: []string{
					EncryptionEncA128CBCHS256,
					EncryptionEncA192CBCHS384,
					EncryptionEncA256CBCHS512,
					EncryptionEncA128GCM,
					EncryptionEncA192GCM,
					EncryptionEncA256GCM,
				},
			},
			OAuth2PushedAuthorizationDiscoveryOptions: &OAuth2PushedAuthorizationDiscoveryOptions{
				RequirePushedAuthorizationRequests: c.RequirePushedAuthorizationRequests,
			},
			OAuth2IssuerIdentificationDiscoveryOptions: &OAuth2IssuerIdentificationDiscoveryOptions{
				AuthorizationResponseIssuerParameterSupported: true,
			},
		},

		OpenIDConnectDiscoveryOptions: OpenIDConnectDiscoveryOptions{
			IDTokenSigningAlgValuesSupported: []string{
				SigningAlgHMACUsingSHA256,
				SigningAlgHMACUsingSHA384,
				SigningAlgHMACUsingSHA512,
				SigningAlgRSAUsingSHA256,
				SigningAlgRSAUsingSHA384,
				SigningAlgRSAUsingSHA512,
				SigningAlgECDSAUsingP256AndSHA256,
				SigningAlgECDSAUsingP384AndSHA384,
				SigningAlgECDSAUsingP521AndSHA512,
				SigningAlgRSAPSSUsingSHA256,
				SigningAlgRSAPSSUsingSHA384,
				SigningAlgRSAPSSUsingSHA512,
			},
			IDTokenEncryptionAlgValuesSupported: []string{
				EncryptionAlgRSA15,
				EncryptionAlgRSAOAEP,
				EncryptionAlgRSAOAEP256,
				EncryptionAlgA128KW,
				EncryptionAlgA192KW,
				EncryptionAlgA256KW,
				EncryptionAlgDirect,
				EncryptionAlgECDHES,
				EncryptionAlgECDHESA128KW,
				EncryptionAlgECDHESA192KW,
				EncryptionAlgECDHESA256KW,
				EncryptionAlgA128GCMKW,
				EncryptionAlgA192GCMKW,
				EncryptionAlgA256GCMKW,
				EncryptionAlgPBES2HS256A128KW,
				EncryptionAlgPBES2HS284A192KW,
				EncryptionAlgPBES2HS512A256KW,
			},
			IDTokenEncryptionEncValuesSupported: []string{
				EncryptionEncA128CBCHS256,
				EncryptionEncA192CBCHS384,
				EncryptionEncA256CBCHS512,
				EncryptionEncA128GCM,
				EncryptionEncA192GCM,
				EncryptionEncA256GCM,
			},
			UserinfoSigningAlgValuesSupported: []string{
				SigningAlgHMACUsingSHA256,
				SigningAlgHMACUsingSHA384,
				SigningAlgHMACUsingSHA512,
				SigningAlgRSAUsingSHA256,
				SigningAlgRSAUsingSHA384,
				SigningAlgRSAUsingSHA512,
				SigningAlgECDSAUsingP256AndSHA256,
				SigningAlgECDSAUsingP384AndSHA384,
				SigningAlgECDSAUsingP521AndSHA512,
				SigningAlgRSAPSSUsingSHA256,
				SigningAlgRSAPSSUsingSHA384,
				SigningAlgRSAPSSUsingSHA512,
				SigningAlgNone,
			},
			UserinfoEncryptionAlgValuesSupported: []string{
				EncryptionAlgRSA15,
				EncryptionAlgRSAOAEP,
				EncryptionAlgRSAOAEP256,
				EncryptionAlgA128KW,
				EncryptionAlgA192KW,
				EncryptionAlgA256KW,
				EncryptionAlgDirect,
				EncryptionAlgECDHES,
				EncryptionAlgECDHESA128KW,
				EncryptionAlgECDHESA192KW,
				EncryptionAlgECDHESA256KW,
				EncryptionAlgA128GCMKW,
				EncryptionAlgA192GCMKW,
				EncryptionAlgA256GCMKW,
				EncryptionAlgPBES2HS256A128KW,
				EncryptionAlgPBES2HS284A192KW,
				EncryptionAlgPBES2HS512A256KW,
			},
			UserinfoEncryptionEncValuesSupported: []string{
				EncryptionEncA128CBCHS256,
				EncryptionEncA192CBCHS384,
				EncryptionEncA256CBCHS512,
				EncryptionEncA128GCM,
				EncryptionEncA192GCM,
				EncryptionEncA256GCM,
			},
			RequestObjectSigningAlgValuesSupported: []string{
				SigningAlgHMACUsingSHA256,
				SigningAlgHMACUsingSHA384,
				SigningAlgHMACUsingSHA512,
				SigningAlgRSAUsingSHA256,
				SigningAlgRSAUsingSHA384,
				SigningAlgRSAUsingSHA512,
				SigningAlgECDSAUsingP256AndSHA256,
				SigningAlgECDSAUsingP384AndSHA384,
				SigningAlgECDSAUsingP521AndSHA512,
				SigningAlgRSAPSSUsingSHA256,
				SigningAlgRSAPSSUsingSHA384,
				SigningAlgRSAPSSUsingSHA512,
				SigningAlgNone,
			},
			RequestObjectEncryptionAlgValuesSupported: []string{
				EncryptionAlgRSA15,
				EncryptionAlgRSAOAEP,
				EncryptionAlgRSAOAEP256,
				EncryptionAlgA128KW,
				EncryptionAlgA192KW,
				EncryptionAlgA256KW,
				EncryptionAlgDirect,
				EncryptionAlgECDHES,
				EncryptionAlgECDHESA128KW,
				EncryptionAlgECDHESA192KW,
				EncryptionAlgECDHESA256KW,
				EncryptionAlgA128GCMKW,
				EncryptionAlgA192GCMKW,
				EncryptionAlgA256GCMKW,
				EncryptionAlgPBES2HS256A128KW,
				EncryptionAlgPBES2HS284A192KW,
				EncryptionAlgPBES2HS512A256KW,
			},
			RequestObjectEncryptionEncValuesSupported: []string{
				EncryptionEncA128CBCHS256,
				EncryptionEncA192CBCHS384,
				EncryptionEncA256CBCHS512,
				EncryptionEncA128GCM,
				EncryptionEncA192GCM,
				EncryptionEncA256GCM,
			},
			ClaimTypesSupported: []string{
				ClaimTypeNormal,
			},
			RequestParameterSupported:     true,
			RequestURIParameterSupported:  true,
			RequireRequestURIRegistration: true,
			ClaimsParameterSupported:      true,
		},
		OpenIDConnectPromptCreateDiscoveryOptions: &OpenIDConnectPromptCreateDiscoveryOptions{
			PromptValuesSupported: []string{
				PromptConsent,
				PromptLogin,
				PromptNone,
				PromptSelectAccount,
			},
		},
		OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions: &OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions{
			AuthorizationSigningAlgValuesSupported: []string{
				SigningAlgHMACUsingSHA256,
				SigningAlgHMACUsingSHA384,
				SigningAlgHMACUsingSHA512,
				SigningAlgRSAUsingSHA256,
				SigningAlgRSAUsingSHA384,
				SigningAlgRSAUsingSHA512,
				SigningAlgECDSAUsingP256AndSHA256,
				SigningAlgECDSAUsingP384AndSHA384,
				SigningAlgECDSAUsingP521AndSHA512,
				SigningAlgRSAPSSUsingSHA256,
				SigningAlgRSAPSSUsingSHA384,
				SigningAlgRSAPSSUsingSHA512,
			},
			AuthorizationEncryptionAlgValuesSupported: []string{
				EncryptionAlgRSA15,
				EncryptionAlgRSAOAEP,
				EncryptionAlgRSAOAEP256,
				EncryptionAlgA128KW,
				EncryptionAlgA192KW,
				EncryptionAlgA256KW,
				EncryptionAlgDirect,
				EncryptionAlgECDHES,
				EncryptionAlgECDHESA128KW,
				EncryptionAlgECDHESA192KW,
				EncryptionAlgECDHESA256KW,
				EncryptionAlgA128GCMKW,
				EncryptionAlgA192GCMKW,
				EncryptionAlgA256GCMKW,
				EncryptionAlgPBES2HS256A128KW,
				EncryptionAlgPBES2HS284A192KW,
				EncryptionAlgPBES2HS512A256KW,
			},
			AuthorizationEncryptionEncValuesSupported: []string{
				EncryptionEncA128CBCHS256,
				EncryptionEncA192CBCHS384,
				EncryptionEncA256CBCHS512,
				EncryptionEncA128GCM,
				EncryptionEncA192GCM,
				EncryptionEncA256GCM,
			},
		},
	}

	if c.EnablePKCEPlainChallenge {
		config.CodeChallengeMethodsSupported = append(config.CodeChallengeMethodsSupported, PKCEChallengeMethodPlain)
	}

	return config
}

// Copy the values of the OAuth2WellKnownConfiguration and return it as a new struct.
func (opts OAuth2WellKnownConfiguration) Copy() (optsCopy OAuth2WellKnownConfiguration) {
	optsCopy = OAuth2WellKnownConfiguration{
		CommonDiscoveryOptions: opts.CommonDiscoveryOptions,
		OAuth2DiscoveryOptions: opts.OAuth2DiscoveryOptions,
	}

	if opts.OAuth2DeviceAuthorizationGrantDiscoveryOptions != nil {
		optsCopy.OAuth2DeviceAuthorizationGrantDiscoveryOptions = &OAuth2DeviceAuthorizationGrantDiscoveryOptions{}
		*optsCopy.OAuth2DeviceAuthorizationGrantDiscoveryOptions = *opts.OAuth2DeviceAuthorizationGrantDiscoveryOptions
	}

	if opts.OAuth2MutualTLSClientAuthenticationDiscoveryOptions != nil {
		optsCopy.OAuth2MutualTLSClientAuthenticationDiscoveryOptions = &OAuth2MutualTLSClientAuthenticationDiscoveryOptions{}
		*optsCopy.OAuth2MutualTLSClientAuthenticationDiscoveryOptions = *opts.OAuth2MutualTLSClientAuthenticationDiscoveryOptions
	}

	if opts.OAuth2IssuerIdentificationDiscoveryOptions != nil {
		optsCopy.OAuth2IssuerIdentificationDiscoveryOptions = &OAuth2IssuerIdentificationDiscoveryOptions{}
		*optsCopy.OAuth2IssuerIdentificationDiscoveryOptions = *opts.OAuth2IssuerIdentificationDiscoveryOptions
	}

	if opts.OAuth2JWTIntrospectionResponseDiscoveryOptions != nil {
		optsCopy.OAuth2JWTIntrospectionResponseDiscoveryOptions = &OAuth2JWTIntrospectionResponseDiscoveryOptions{}
		*optsCopy.OAuth2JWTIntrospectionResponseDiscoveryOptions = *opts.OAuth2JWTIntrospectionResponseDiscoveryOptions
	}

	if opts.OAuth2JWTSecuredAuthorizationRequestDiscoveryOptions != nil {
		optsCopy.OAuth2JWTSecuredAuthorizationRequestDiscoveryOptions = &OAuth2JWTSecuredAuthorizationRequestDiscoveryOptions{}
		*optsCopy.OAuth2JWTSecuredAuthorizationRequestDiscoveryOptions = *opts.OAuth2JWTSecuredAuthorizationRequestDiscoveryOptions
	}

	if opts.OAuth2PushedAuthorizationDiscoveryOptions != nil {
		optsCopy.OAuth2PushedAuthorizationDiscoveryOptions = &OAuth2PushedAuthorizationDiscoveryOptions{}
		*optsCopy.OAuth2PushedAuthorizationDiscoveryOptions = *opts.OAuth2PushedAuthorizationDiscoveryOptions
	}

	return optsCopy
}

// Copy the values of the OpenIDConnectWellKnownConfiguration and return it as a new struct.
func (opts OpenIDConnectWellKnownConfiguration) Copy() (optsCopy OpenIDConnectWellKnownConfiguration) {
	optsCopy = OpenIDConnectWellKnownConfiguration{
		OAuth2WellKnownConfiguration:  opts.OAuth2WellKnownConfiguration.Copy(),
		OpenIDConnectDiscoveryOptions: opts.OpenIDConnectDiscoveryOptions,
	}

	if opts.OpenIDConnectFrontChannelLogoutDiscoveryOptions != nil {
		optsCopy.OpenIDConnectFrontChannelLogoutDiscoveryOptions = &OpenIDConnectFrontChannelLogoutDiscoveryOptions{}
		*optsCopy.OpenIDConnectFrontChannelLogoutDiscoveryOptions = *opts.OpenIDConnectFrontChannelLogoutDiscoveryOptions
	}

	if opts.OpenIDConnectBackChannelLogoutDiscoveryOptions != nil {
		optsCopy.OpenIDConnectBackChannelLogoutDiscoveryOptions = &OpenIDConnectBackChannelLogoutDiscoveryOptions{}
		*optsCopy.OpenIDConnectBackChannelLogoutDiscoveryOptions = *opts.OpenIDConnectBackChannelLogoutDiscoveryOptions
	}

	if opts.OpenIDConnectSessionManagementDiscoveryOptions != nil {
		optsCopy.OpenIDConnectSessionManagementDiscoveryOptions = &OpenIDConnectSessionManagementDiscoveryOptions{}
		*optsCopy.OpenIDConnectSessionManagementDiscoveryOptions = *opts.OpenIDConnectSessionManagementDiscoveryOptions
	}

	if opts.OpenIDConnectRPInitiatedLogoutDiscoveryOptions != nil {
		optsCopy.OpenIDConnectRPInitiatedLogoutDiscoveryOptions = &OpenIDConnectRPInitiatedLogoutDiscoveryOptions{}
		*optsCopy.OpenIDConnectRPInitiatedLogoutDiscoveryOptions = *opts.OpenIDConnectRPInitiatedLogoutDiscoveryOptions
	}

	if opts.OpenIDConnectPromptCreateDiscoveryOptions != nil {
		optsCopy.OpenIDConnectPromptCreateDiscoveryOptions = &OpenIDConnectPromptCreateDiscoveryOptions{}
		*optsCopy.OpenIDConnectPromptCreateDiscoveryOptions = *opts.OpenIDConnectPromptCreateDiscoveryOptions
	}

	if opts.OpenIDConnectClientInitiatedBackChannelAuthFlowDiscoveryOptions != nil {
		optsCopy.OpenIDConnectClientInitiatedBackChannelAuthFlowDiscoveryOptions = &OpenIDConnectClientInitiatedBackChannelAuthFlowDiscoveryOptions{}
		*optsCopy.OpenIDConnectClientInitiatedBackChannelAuthFlowDiscoveryOptions = *opts.OpenIDConnectClientInitiatedBackChannelAuthFlowDiscoveryOptions
	}

	if opts.OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions != nil {
		optsCopy.OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions = &OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions{}
		*optsCopy.OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions = *opts.OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions
	}

	if opts.OpenIDFederationDiscoveryOptions != nil {
		optsCopy.OpenIDFederationDiscoveryOptions = &OpenIDFederationDiscoveryOptions{}
		*optsCopy.OpenIDFederationDiscoveryOptions = *opts.OpenIDFederationDiscoveryOptions
	}

	if opts.OpenIDConnectIdentityAssurance != nil {
		optsCopy.OpenIDConnectIdentityAssurance = &OpenIDConnectIdentityAssurance{}
		*optsCopy.OpenIDConnectIdentityAssurance = *opts.OpenIDConnectIdentityAssurance
	}

	return optsCopy
}
