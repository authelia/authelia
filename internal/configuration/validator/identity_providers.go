package validator

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateIdentityProviders validates and update IdentityProviders configuration.
func ValidateIdentityProviders(config *schema.IdentityProvidersConfiguration, validator *schema.StructValidator) {
	validateOIDC(config.OIDC, validator)
}

func validateOIDC(config *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if config != nil {
		if config.IssuerPrivateKey == "" {
			validator.Push(fmt.Errorf(errFmtOIDCNoPrivateKey))
		}

		if config.AccessTokenLifespan == time.Duration(0) {
			config.AccessTokenLifespan = schema.DefaultOpenIDConnectConfiguration.AccessTokenLifespan
		}

		if config.AuthorizeCodeLifespan == time.Duration(0) {
			config.AuthorizeCodeLifespan = schema.DefaultOpenIDConnectConfiguration.AuthorizeCodeLifespan
		}

		if config.IDTokenLifespan == time.Duration(0) {
			config.IDTokenLifespan = schema.DefaultOpenIDConnectConfiguration.IDTokenLifespan
		}

		if config.RefreshTokenLifespan == time.Duration(0) {
			config.RefreshTokenLifespan = schema.DefaultOpenIDConnectConfiguration.RefreshTokenLifespan
		}

		if config.MinimumParameterEntropy != 0 && config.MinimumParameterEntropy < 8 {
			validator.PushWarning(fmt.Errorf(errFmtOIDCServerInsecureParameterEntropy, config.MinimumParameterEntropy))
		}

		if config.EnforcePKCE == "" {
			config.EnforcePKCE = schema.DefaultOpenIDConnectConfiguration.EnforcePKCE
		}

		if config.EnforcePKCE != "never" && config.EnforcePKCE != "public_clients_only" && config.EnforcePKCE != "always" {
			validator.Push(fmt.Errorf(errFmtOIDCEnforcePKCEInvalidValue, config.EnforcePKCE))
		}

		validateOIDCClients(config, validator)

		if len(config.Clients) == 0 {
			validator.Push(fmt.Errorf(errFmtOIDCNoClientsConfigured))
		}
	}
}

func validateOIDCClients(config *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	invalidID, duplicateIDs := false, false

	var ids []string

	for c, client := range config.Clients {
		if client.ID == "" {
			invalidID = true
		} else {
			if client.Description == "" {
				config.Clients[c].Description = client.ID
			}

			if utils.IsStringInSliceFold(client.ID, ids) {
				duplicateIDs = true
			}
			ids = append(ids, client.ID)
		}

		if client.Public {
			if client.Secret != "" {
				validator.Push(fmt.Errorf(errFmtOIDCClientPublicInvalidSecret, client.ID))
			}
		} else {
			if client.Secret == "" {
				validator.Push(fmt.Errorf(errFmtOIDCClientInvalidSecret, client.ID))
			}
		}

		if client.Policy == "" {
			config.Clients[c].Policy = schema.DefaultOpenIDConnectClientConfiguration.Policy
		} else if client.Policy != policyOneFactor && client.Policy != policyTwoFactor {
			validator.Push(fmt.Errorf(errFmtOIDCClientInvalidPolicy, client.ID, client.Policy))
		}

		validateOIDCClientScopes(c, config, validator)
		validateOIDCClientGrantTypes(c, config, validator)
		validateOIDCClientResponseTypes(c, config, validator)
		validateOIDCClientResponseModes(c, config, validator)
		validateOIDDClientUserinfoAlgorithm(c, config, validator)

		validateOIDCClientRedirectURIs(client, validator)
	}

	if invalidID {
		validator.Push(fmt.Errorf(errFmtOIDCClientsWithEmptyID))
	}

	if duplicateIDs {
		validator.Push(fmt.Errorf(errFmtOIDCClientsDuplicateID))
	}
}

func validateOIDCClientScopes(c int, configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if len(configuration.Clients[c].Scopes) == 0 {
		configuration.Clients[c].Scopes = schema.DefaultOpenIDConnectClientConfiguration.Scopes
		return
	}

	if !utils.IsStringInSlice("openid", configuration.Clients[c].Scopes) {
		configuration.Clients[c].Scopes = append(configuration.Clients[c].Scopes, "openid")
	}

	for _, scope := range configuration.Clients[c].Scopes {
		if !utils.IsStringInSlice(scope, validOIDCScopes) {
			validator.Push(fmt.Errorf(
				errFmtOIDCClientInvalidEntry,
				configuration.Clients[c].ID, "scopes", strings.Join(validOIDCScopes, "', '"), scope))
		}
	}
}

func validateOIDCClientGrantTypes(c int, configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if len(configuration.Clients[c].GrantTypes) == 0 {
		configuration.Clients[c].GrantTypes = schema.DefaultOpenIDConnectClientConfiguration.GrantTypes
		return
	}

	for _, grantType := range configuration.Clients[c].GrantTypes {
		if !utils.IsStringInSlice(grantType, validOIDCGrantTypes) {
			validator.Push(fmt.Errorf(
				errFmtOIDCClientInvalidEntry,
				configuration.Clients[c].ID, "grant_types", strings.Join(validOIDCGrantTypes, "', '"), grantType))
		}
	}
}

func validateOIDCClientResponseTypes(c int, configuration *schema.OpenIDConnectConfiguration, _ *schema.StructValidator) {
	if len(configuration.Clients[c].ResponseTypes) == 0 {
		configuration.Clients[c].ResponseTypes = schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes
		return
	}
}

func validateOIDCClientResponseModes(c int, configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if len(configuration.Clients[c].ResponseModes) == 0 {
		configuration.Clients[c].ResponseModes = schema.DefaultOpenIDConnectClientConfiguration.ResponseModes
		return
	}

	for _, responseMode := range configuration.Clients[c].ResponseModes {
		if !utils.IsStringInSlice(responseMode, validOIDCResponseModes) {
			validator.Push(fmt.Errorf(
				errFmtOIDCClientInvalidEntry,
				configuration.Clients[c].ID, "response_modes", strings.Join(validOIDCResponseModes, "', '"), responseMode))
		}
	}
}

func validateOIDDClientUserinfoAlgorithm(c int, configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if configuration.Clients[c].UserinfoSigningAlgorithm == "" {
		configuration.Clients[c].UserinfoSigningAlgorithm = schema.DefaultOpenIDConnectClientConfiguration.UserinfoSigningAlgorithm
	} else if !utils.IsStringInSlice(configuration.Clients[c].UserinfoSigningAlgorithm, validOIDCUserinfoAlgorithms) {
		validator.Push(fmt.Errorf(errFmtOIDCClientInvalidUserinfoAlgorithm,
			configuration.Clients[c].ID, strings.Join(validOIDCUserinfoAlgorithms, ", "), configuration.Clients[c].UserinfoSigningAlgorithm))
	}
}

func validateOIDCClientRedirectURIs(client schema.OpenIDConnectClientConfiguration, validator *schema.StructValidator) {
	for _, redirectURI := range client.RedirectURIs {
		if redirectURI == oauth2InstalledApp {
			if client.Public {
				continue
			}

			validator.Push(fmt.Errorf(errFmtOIDCClientRedirectURIPublic, client.ID, oauth2InstalledApp))

			continue
		}

		parsedURL, err := url.Parse(redirectURI)
		if err != nil {
			validator.Push(fmt.Errorf(errFmtOIDCClientRedirectURICantBeParsed, client.ID, redirectURI, err))
			continue
		}

		if !parsedURL.IsAbs() {
			validator.Push(fmt.Errorf(errFmtOIDCClientRedirectURIAbsolute, client.ID, redirectURI))
			return
		}

		if !client.Public && parsedURL.Scheme != schemeHTTPS && parsedURL.Scheme != schemeHTTP {
			validator.Push(fmt.Errorf(errFmtOIDCClientRedirectURI, client.ID, redirectURI, parsedURL.Scheme))
		}
	}
}
