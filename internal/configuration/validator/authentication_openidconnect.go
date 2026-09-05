package validator

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

var reOIDCRPProviderID = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,31}$`)

var (
	validOIDCRPAlgs        = []string{"ES256", "ES384", "ES512", "PS256", "PS384", "PS512", "RS256", "RS384", "RS512"}
	validOIDCRPAuthMethods = []string{"client_secret_basic", "client_secret_post", "none"}
)

// ValidateAuthenticationBackendOpenIDConnect validates and updates the OpenID Connect 1.0 Relying Party configuration.
func ValidateAuthenticationBackendOpenIDConnect(config *schema.AuthenticationBackendOpenIDConnect, validator *schema.StructValidator) {
	if len(config.Providers) == 0 {
		validator.Push(errors.New(errFmtOIDCRPProvidersRequired))

		return
	}

	ids := map[string]int{}

	for i := range config.Providers {
		validateAuthenticationBackendOpenIDConnectProvider(i, &config.Providers[i], ids, validator)
	}
}

//nolint:gocyclo
func validateAuthenticationBackendOpenIDConnectProvider(i int, config *schema.AuthenticationBackendOpenIDConnectProvider, ids map[string]int, validator *schema.StructValidator) {
	if !reOIDCRPProviderID.MatchString(config.ID) {
		validator.Push(fmt.Errorf(errFmtOIDCRPProviderID, i+1, reOIDCRPProviderID.String(), config.ID))

		return
	}

	if _, ok := ids[config.ID]; ok {
		validator.Push(fmt.Errorf(errFmtOIDCRPProviderIDDuplicate, config.ID))

		return
	}

	ids[config.ID] = i

	if config.Name == "" {
		validator.Push(fmt.Errorf(errFmtOIDCRPProviderOptionRequired, config.ID, "name"))
	}

	if config.ClientID == "" {
		validator.Push(fmt.Errorf(errFmtOIDCRPProviderOptionRequired, config.ID, "client_id"))
	}

	validateAuthenticationBackendOpenIDConnectProviderIssuer(config, validator)

	if len(config.Scopes) == 0 {
		config.Scopes = []string{"openid", "profile", "email"}
	} else if !utils.IsStringInSlice("openid", config.Scopes) {
		config.Scopes = append([]string{"openid"}, config.Scopes...)
	}

	switch config.TokenEndpointAuthMethod {
	case "":
		config.TokenEndpointAuthMethod = "client_secret_basic"
	case "client_secret_basic", "client_secret_post", "none":
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCRPProviderAuthMethod, config.ID, utils.StringJoinOr(validOIDCRPAuthMethods), config.TokenEndpointAuthMethod))
	}

	if config.TokenEndpointAuthMethod != "none" && config.ClientSecret == "" {
		validator.Push(fmt.Errorf(errFmtOIDCRPProviderSecretRequired, config.ID, config.TokenEndpointAuthMethod))
	}

	switch config.IDTokenSignedResponseAlg {
	case "":
		config.IDTokenSignedResponseAlg = "RS256"
	default:
		if !utils.IsStringInSlice(config.IDTokenSignedResponseAlg, validOIDCRPAlgs) {
			validator.Push(fmt.Errorf(errFmtOIDCRPProviderAlg, config.ID, utils.StringJoinOr(validOIDCRPAlgs), config.IDTokenSignedResponseAlg))
		}
	}

	switch config.PKCE.ChallengeMethod {
	case "":
		config.PKCE.ChallengeMethod = "S256"
	case "S256":
		break
	default:
		validator.Push(fmt.Errorf(errFmtOIDCRPProviderPKCE, config.ID, config.PKCE.ChallengeMethod))
	}

	validateAuthenticationBackendOpenIDConnectProviderEndpoints(config, validator)

	if config.Discovery.Disable {
		if config.Endpoints.Authorization == "" {
			validator.Push(fmt.Errorf(errFmtOIDCRPProviderEndpoint, config.ID, "authorization"))
		}

		if config.Endpoints.Token == "" {
			validator.Push(fmt.Errorf(errFmtOIDCRPProviderEndpoint, config.ID, "token"))
		}

		if config.Endpoints.JSONWebKeys == "" && len(config.JSONWebKeys) == 0 {
			validator.Push(fmt.Errorf(errFmtOIDCRPProviderEndpointJWKS, config.ID))
		}
	}
}

// validateAuthenticationBackendOpenIDConnectProviderEndpoints validates the explicitly configured endpoints in the
// same manner as the issuer. These values are used verbatim to make requests which carry the client secret and the
// authorization code so a plaintext scheme must never be accepted.
func validateAuthenticationBackendOpenIDConnectProviderEndpoints(config *schema.AuthenticationBackendOpenIDConnectProvider, validator *schema.StructValidator) {
	endpoints := []struct {
		name  string
		value string
	}{
		{"authorization", config.Endpoints.Authorization},
		{"token", config.Endpoints.Token},
		{"jwks", config.Endpoints.JSONWebKeys},
	}

	for _, endpoint := range endpoints {
		if endpoint.value == "" {
			continue
		}

		parsed, err := url.Parse(endpoint.value)
		if err != nil {
			validator.Push(fmt.Errorf(errFmtOIDCRPProviderEndpointParse, config.ID, endpoint.name, err))

			continue
		}

		if parsed.Scheme != schemeHTTPS {
			validator.Push(fmt.Errorf(errFmtOIDCRPProviderEndpointScheme, config.ID, endpoint.name, parsed.Scheme))
		}
	}
}

func validateAuthenticationBackendOpenIDConnectProviderIssuer(config *schema.AuthenticationBackendOpenIDConnectProvider, validator *schema.StructValidator) {
	if config.Issuer == "" {
		validator.Push(fmt.Errorf(errFmtOIDCRPProviderOptionRequired, config.ID, "issuer"))

		return
	}

	issuer, err := url.Parse(config.Issuer)
	if err != nil {
		validator.Push(fmt.Errorf(errFmtOIDCRPProviderIssuerParse, config.ID, err))

		return
	}

	if issuer.Scheme != "https" {
		validator.Push(fmt.Errorf(errFmtOIDCRPProviderIssuerScheme, config.ID, issuer.Scheme))
	}
}
