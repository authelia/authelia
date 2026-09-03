// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

const (
	externalIdentityTypeOpenIDConnect = "openid_connect"
	externalIdentityTypeDiscord       = "discord"
	externalIdentityTypePlex          = "plex"
	externalIdentityTypeGitHub        = "github"
)

var reExternalIdentityProviderID = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,31}$`)

var (
	validExternalIdentityTypes                 = []string{externalIdentityTypeOpenIDConnect, externalIdentityTypeDiscord, externalIdentityTypePlex, externalIdentityTypeGitHub}
	validExternalIdentityAlgs                  = []string{"ES256", "ES384", "ES512", "PS256", "PS384", "PS512", "RS256", "RS384", "RS512"}
	validExternalIdentityAuthMethods           = []string{"client_secret_basic", "client_secret_post", "none"}
	validExternalIdentityConfidentialMethods   = []string{"client_secret_basic", "client_secret_post"}
	validExternalIdentityResponseModes         = []string{"query", "form_post"}
	validExternalIdentityQueryResponseModes    = []string{"query"}
	defaultExternalIdentityOpenIDConnectScopes = []string{"openid", "profile", "email"}
	defaultExternalIdentityDiscordScopes       = []string{"identify", "email"}
	defaultExternalIdentityGitHubScopes        = []string{"read:user", "user:email"}
)

// ValidateAuthenticationBackendExternalIdentity validates and updates the external identity configuration.
func ValidateAuthenticationBackendExternalIdentity(config *schema.AuthenticationBackendExternalIdentity, validator *schema.StructValidator) {
	if len(config.Providers) == 0 {
		validator.Push(errors.New(errFmtExternalIdentityProvidersRequired))

		return
	}

	ids := map[string]int{}

	for i := range config.Providers {
		validateAuthenticationBackendExternalIdentityProvider(i, &config.Providers[i], ids, validator)
	}

	validateAuthenticationBackendExternalIdentityProvidersUnique(config, validator)
}

func validateAuthenticationBackendExternalIdentityProvidersUnique(config *schema.AuthenticationBackendExternalIdentity, validator *schema.StructValidator) {
	var (
		types   = map[string][]string{}
		issuers = map[string][]string{}
		order   []string
		seen    = map[string]bool{}
	)

	for _, provider := range config.Providers {
		if !reExternalIdentityProviderID.MatchString(provider.ID) || seen[provider.ID] {
			continue
		}

		seen[provider.ID] = true

		switch provider.Type {
		case externalIdentityTypeDiscord, externalIdentityTypePlex, externalIdentityTypeGitHub:
			types[provider.Type] = append(types[provider.Type], provider.ID)
		case externalIdentityTypeOpenIDConnect:
			if provider.Issuer == "" {
				continue
			}

			if _, ok := issuers[provider.Issuer]; !ok {
				order = append(order, provider.Issuer)
			}

			issuers[provider.Issuer] = append(issuers[provider.Issuer], provider.ID)
		}
	}

	for _, providerType := range []string{externalIdentityTypeDiscord, externalIdentityTypePlex, externalIdentityTypeGitHub} {
		if len(types[providerType]) > 1 {
			validator.Push(fmt.Errorf(errFmtExternalIdentityProviderTypeDuplicate, providerType, utils.StringJoinAnd(types[providerType])))
		}
	}

	if utils.Dev && os.Getenv(envExternalIdentityOpenIDConnectAllowDuplicateIssuer) == "true" {
		return
	}

	for _, issuer := range order {
		if len(issuers[issuer]) > 1 {
			validator.Push(fmt.Errorf(errFmtExternalIdentityProviderIssuerDuplicate, issuer, utils.StringJoinAnd(issuers[issuer])))
		}
	}
}

func validateAuthenticationBackendExternalIdentityProvider(i int, config *schema.AuthenticationBackendExternalIdentityProvider, ids map[string]int, validator *schema.StructValidator) {
	if !reExternalIdentityProviderID.MatchString(config.ID) {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderID, i+1, reExternalIdentityProviderID.String(), config.ID))

		return
	}

	if _, ok := ids[config.ID]; ok {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderIDDuplicate, config.ID))

		return
	}

	ids[config.ID] = i

	if config.Name == "" {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderOptionRequired, config.ID, "name"))
	}

	if config.ClientID == "" {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderOptionRequired, config.ID, "client_id"))
	}

	switch config.Type {
	case "", externalIdentityTypeOpenIDConnect:
		config.Type = externalIdentityTypeOpenIDConnect

		validateAuthenticationBackendExternalIdentityProviderOpenIDConnect(config, validator)
	case externalIdentityTypeDiscord:
		validateAuthenticationBackendExternalIdentityProviderDiscord(config, validator)
	case externalIdentityTypePlex:
		validateAuthenticationBackendExternalIdentityProviderPlex(config, validator)
	case externalIdentityTypeGitHub:
		validateAuthenticationBackendExternalIdentityProviderGitHub(config, validator)
	default:
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderType, config.ID, utils.StringJoinOr(validExternalIdentityTypes), config.Type))
	}
}

func validateAuthenticationBackendExternalIdentityProviderOpenIDConnect(config *schema.AuthenticationBackendExternalIdentityProvider, validator *schema.StructValidator) {
	validateAuthenticationBackendExternalIdentityProviderIssuer(config, validator)
	validateAuthenticationBackendExternalIdentityProviderScopes(config, "openid", defaultExternalIdentityOpenIDConnectScopes)
	validateAuthenticationBackendExternalIdentityProviderResponseMode(config, validExternalIdentityResponseModes, validator)
	validateAuthenticationBackendExternalIdentityProviderAuthMethod(config, validExternalIdentityAuthMethods, validator)

	switch config.IDTokenSignedResponseAlg {
	case "":
		// With discovery enabled the algorithm is left unset so it is selected from the algorithms the provider
		// advertises. Without discovery there is nothing to select from, so the OpenID Connect default applies.
		if config.Discovery.Disable {
			config.IDTokenSignedResponseAlg = "RS256"
		}
	default:
		if !utils.IsStringInSlice(config.IDTokenSignedResponseAlg, validExternalIdentityAlgs) {
			validator.Push(fmt.Errorf(errFmtExternalIdentityProviderAlg, config.ID, utils.StringJoinOr(validExternalIdentityAlgs), config.IDTokenSignedResponseAlg))
		}
	}

	if config.UserInfoSignedResponseAlg != "" && !utils.IsStringInSlice(config.UserInfoSignedResponseAlg, validExternalIdentityAlgs) {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderUserInfoAlg, config.ID, utils.StringJoinOr(validExternalIdentityAlgs), config.UserInfoSignedResponseAlg))
	}

	validateAuthenticationBackendExternalIdentityProviderPKCE(config, validator)
	validateAuthenticationBackendExternalIdentityProviderAMR(config, validator)
	validateAuthenticationBackendExternalIdentityProviderEndpoints(config, validator)

	if config.Discovery.Disable {
		if config.Endpoints.Authorization == "" {
			validator.Push(fmt.Errorf(errFmtExternalIdentityProviderEndpoint, config.ID, "authorization"))
		}

		if config.Endpoints.Token == "" {
			validator.Push(fmt.Errorf(errFmtExternalIdentityProviderEndpoint, config.ID, "token"))
		}

		if config.Endpoints.JSONWebKeys == "" && len(config.JSONWebKeys) == 0 {
			validator.Push(fmt.Errorf(errFmtExternalIdentityProviderEndpointJWKS, config.ID))
		}

		if config.RequirePushedAuthorizationRequests && config.Endpoints.PushedAuthorizationRequest == "" {
			validator.Push(fmt.Errorf(errFmtExternalIdentityProviderEndpointPAR, config.ID))
		}
	}
}

// validateAuthenticationBackendExternalIdentityProviderDiscord validates a Discord provider. Discord's endpoints are
// fixed and it is not an OpenID Connect 1.0 Provider, so none of the options which describe one apply to it, and
// configuring any of them is rejected rather than silently ignored.
func validateAuthenticationBackendExternalIdentityProviderDiscord(config *schema.AuthenticationBackendExternalIdentityProvider, validator *schema.StructValidator) {
	validateAuthenticationBackendExternalIdentityProviderUnsupported(config, externalIdentityOpenIDConnectOnlyOptions(config), validator)

	validateAuthenticationBackendExternalIdentityProviderScopes(config, "identify", defaultExternalIdentityDiscordScopes)
	validateAuthenticationBackendExternalIdentityProviderResponseMode(config, validExternalIdentityQueryResponseModes, validator)
	validateAuthenticationBackendExternalIdentityProviderAuthMethod(config, validExternalIdentityConfidentialMethods, validator)
	validateAuthenticationBackendExternalIdentityProviderPKCE(config, validator)
	validateAuthenticationBackendExternalIdentityProviderAMRDefault(config, validator)
}

// validateAuthenticationBackendExternalIdentityProviderGitHub validates a GitHub provider. GitHub's endpoints are fixed
// and it is not an OpenID Connect 1.0 Provider, so none of the options which describe one apply to it. GitHub grants
// access to the public profile without any scope, so no scope is mandatory.
func validateAuthenticationBackendExternalIdentityProviderGitHub(config *schema.AuthenticationBackendExternalIdentityProvider, validator *schema.StructValidator) {
	validateAuthenticationBackendExternalIdentityProviderUnsupported(config, externalIdentityOpenIDConnectOnlyOptions(config), validator)

	if len(config.Scopes) == 0 {
		config.Scopes = append([]string(nil), defaultExternalIdentityGitHubScopes...)
	}

	validateAuthenticationBackendExternalIdentityProviderResponseMode(config, validExternalIdentityQueryResponseModes, validator)
	validateAuthenticationBackendExternalIdentityProviderAuthMethod(config, validExternalIdentityConfidentialMethods, validator)
	validateAuthenticationBackendExternalIdentityProviderPKCE(config, validator)
	validateAuthenticationBackendExternalIdentityProviderAMRDefault(config, validator)
}

// validateAuthenticationBackendExternalIdentityProviderPlex validates a Plex provider. Plex is signed in with a PIN
// rather than an OAuth 2.0 authorization code, so there is no client secret, scope, token endpoint, or PKCE either, and
// the client_id is only the client identifier Authelia presents to Plex.
func validateAuthenticationBackendExternalIdentityProviderPlex(config *schema.AuthenticationBackendExternalIdentityProvider, validator *schema.StructValidator) {
	unsupported := append([]externalIdentityOption{
		{"client_secret", config.ClientSecret != ""},
		{"scopes", len(config.Scopes) != 0},
		{"token_endpoint_auth_method", config.TokenEndpointAuthMethod != ""},
		{"pkce.challenge_method", config.PKCE.ChallengeMethod != ""},
	}, externalIdentityOpenIDConnectOnlyOptions(config)...)

	validateAuthenticationBackendExternalIdentityProviderUnsupported(config, unsupported, validator)

	validateAuthenticationBackendExternalIdentityProviderResponseMode(config, validExternalIdentityQueryResponseModes, validator)
	validateAuthenticationBackendExternalIdentityProviderAMRDefault(config, validator)
}

type externalIdentityOption struct {
	name       string
	configured bool
}

func externalIdentityOpenIDConnectOnlyOptions(config *schema.AuthenticationBackendExternalIdentityProvider) []externalIdentityOption {
	return []externalIdentityOption{
		{"issuer", config.Issuer != ""},
		{"id_token_signed_response_alg", config.IDTokenSignedResponseAlg != ""},
		{"authentication_methods_reference.trust", config.AuthenticationMethodsReference.Trust},
		{"authentication_methods_reference.override", config.AuthenticationMethodsReference.Override},
		{"discovery.disable", config.Discovery.Disable},
		{"authorization_response_iss_parameter_supported", config.AuthorizationResponseIssParameterSupported},
		{"require_pushed_authorization_requests", config.RequirePushedAuthorizationRequests},
		{"userinfo_signed_response_alg", config.UserInfoSignedResponseAlg != ""},
		{"endpoints", config.Endpoints != schema.AuthenticationBackendExternalIdentityProviderEndpoints{}},
		{"jwks", len(config.JSONWebKeys) != 0},
	}
}

// validateAuthenticationBackendExternalIdentityProviderUnsupported rejects the options the provider type does not
// support rather than silently ignoring them.
func validateAuthenticationBackendExternalIdentityProviderUnsupported(config *schema.AuthenticationBackendExternalIdentityProvider, options []externalIdentityOption, validator *schema.StructValidator) {
	for _, option := range options {
		if option.configured {
			validator.Push(fmt.Errorf(errFmtExternalIdentityProviderOptionUnsupported, config.ID, option.name, config.Type))
		}
	}
}

func validateAuthenticationBackendExternalIdentityProviderScopes(config *schema.AuthenticationBackendExternalIdentityProvider, required string, defaults []string) {
	if len(config.Scopes) == 0 {
		config.Scopes = append([]string(nil), defaults...)
	} else if !utils.IsStringInSlice(required, config.Scopes) {
		config.Scopes = append([]string{required}, config.Scopes...)
	}
}

func validateAuthenticationBackendExternalIdentityProviderResponseMode(config *schema.AuthenticationBackendExternalIdentityProvider, valid []string, validator *schema.StructValidator) {
	switch {
	case config.ResponseMode == "":
		config.ResponseMode = "query"
	case !utils.IsStringInSlice(config.ResponseMode, valid):
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderResponseMode, config.ID, utils.StringJoinOr(valid), config.ResponseMode))
	}
}

func validateAuthenticationBackendExternalIdentityProviderAuthMethod(config *schema.AuthenticationBackendExternalIdentityProvider, valid []string, validator *schema.StructValidator) {
	switch {
	case config.TokenEndpointAuthMethod == "":
		config.TokenEndpointAuthMethod = "client_secret_basic"
	case !utils.IsStringInSlice(config.TokenEndpointAuthMethod, valid):
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderAuthMethod, config.ID, utils.StringJoinOr(valid), config.TokenEndpointAuthMethod))

		return
	}

	if config.TokenEndpointAuthMethod != "none" && config.ClientSecret == "" {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderSecretRequired, config.ID, config.TokenEndpointAuthMethod))
	}
}

func validateAuthenticationBackendExternalIdentityProviderPKCE(config *schema.AuthenticationBackendExternalIdentityProvider, validator *schema.StructValidator) {
	switch config.PKCE.ChallengeMethod {
	case "":
		config.PKCE.ChallengeMethod = "S256"
	case "S256":
		break
	default:
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderPKCE, config.ID, config.PKCE.ChallengeMethod))
	}
}

// validateAuthenticationBackendExternalIdentityProviderEndpoints validates the explicitly configured endpoints in the
// same manner as the issuer. These values are used verbatim to make requests which carry the client secret, the
// authorization code, and the access token so a plaintext scheme must never be accepted.
func validateAuthenticationBackendExternalIdentityProviderEndpoints(config *schema.AuthenticationBackendExternalIdentityProvider, validator *schema.StructValidator) {
	endpoints := []struct {
		name  string
		value string
	}{
		{"authorization", config.Endpoints.Authorization},
		{"token", config.Endpoints.Token},
		{"userinfo", config.Endpoints.UserInfo},
		{"jwks", config.Endpoints.JSONWebKeys},
		{"pushed_authorization_request", config.Endpoints.PushedAuthorizationRequest},
	}

	for _, endpoint := range endpoints {
		if endpoint.value == "" {
			continue
		}

		parsed, err := url.Parse(endpoint.value)
		if err != nil {
			validator.Push(fmt.Errorf(errFmtExternalIdentityProviderEndpointParse, config.ID, endpoint.name, err))

			continue
		}

		if parsed.Scheme != schemeHTTPS {
			validator.Push(fmt.Errorf(errFmtExternalIdentityProviderEndpointScheme, config.ID, endpoint.name, parsed.Scheme))
		}
	}
}

func validateAuthenticationBackendExternalIdentityProviderIssuer(config *schema.AuthenticationBackendExternalIdentityProvider, validator *schema.StructValidator) {
	if config.Issuer == "" {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderOptionRequired, config.ID, "issuer"))

		return
	}

	issuer, err := url.Parse(config.Issuer)
	if err != nil {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderIssuerParse, config.ID, err))

		return
	}

	if issuer.Scheme != "https" {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderIssuerScheme, config.ID, issuer.Scheme))
	}
}

func validateAuthenticationBackendExternalIdentityProviderAMR(config *schema.AuthenticationBackendExternalIdentityProvider, validator *schema.StructValidator) {
	if config.AuthenticationMethodsReference.Override && len(config.AuthenticationMethodsReference.Default) == 0 {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderAMROverride, config.ID))
	}

	validateAuthenticationBackendExternalIdentityProviderAMRDefault(config, validator)
}

func validateAuthenticationBackendExternalIdentityProviderAMRDefault(config *schema.AuthenticationBackendExternalIdentityProvider, validator *schema.StructValidator) {
	if utils.IsStringInSlice("", config.AuthenticationMethodsReference.Default) {
		validator.Push(fmt.Errorf(errFmtExternalIdentityProviderAMRDefaultEmpty, config.ID))
	}
}
