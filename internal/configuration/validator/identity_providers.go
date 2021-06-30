package validator

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateIdentityProviders validates and update IdentityProviders configuration.
func ValidateIdentityProviders(configuration *schema.IdentityProvidersConfiguration, validator *schema.StructValidator) {
	validateOIDC(configuration.OIDC, validator)
}

func validateOIDC(configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if configuration != nil {
		if configuration.IssuerPrivateKey == "" {
			validator.Push(fmt.Errorf("OIDC Server issuer private key must be provided"))
		}

		if configuration.AccessTokenLifespan == time.Duration(0) {
			configuration.AccessTokenLifespan = schema.DefaultOpenIDConnectConfiguration.AccessTokenLifespan
		}

		if configuration.AuthorizeCodeLifespan == time.Duration(0) {
			configuration.AuthorizeCodeLifespan = schema.DefaultOpenIDConnectConfiguration.AuthorizeCodeLifespan
		}

		if configuration.IDTokenLifespan == time.Duration(0) {
			configuration.IDTokenLifespan = schema.DefaultOpenIDConnectConfiguration.IDTokenLifespan
		}

		if configuration.RefreshTokenLifespan == time.Duration(0) {
			configuration.RefreshTokenLifespan = schema.DefaultOpenIDConnectConfiguration.RefreshTokenLifespan
		}

		validateOIDCClients(configuration, validator)

		if len(configuration.Clients) == 0 {
			validator.Push(fmt.Errorf("OIDC Server has no clients defined"))
		}
	}
}

func validateOIDCClients(configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	invalidID, duplicateIDs := false, false

	var ids []string

	for c, client := range configuration.Clients {
		if client.ID == "" {
			invalidID = true
		} else {
			if client.Description == "" {
				configuration.Clients[c].Description = client.ID
			}

			if utils.IsStringInSliceFold(client.ID, ids) {
				duplicateIDs = true
			}
			ids = append(ids, client.ID)
		}

		if client.Secret == "" {
			validator.Push(fmt.Errorf(errFmtOIDCServerClientInvalidSecret, client.ID))
		}

		if client.Policy == "" {
			configuration.Clients[c].Policy = schema.DefaultOpenIDConnectClientConfiguration.Policy
		} else if client.Policy != oneFactorPolicy && client.Policy != twoFactorPolicy {
			validator.Push(fmt.Errorf(errFmtOIDCServerClientInvalidPolicy, client.ID, client.Policy))
		}

		validateOIDCClientScopes(c, configuration, validator)
		validateOIDCClientGrantTypes(c, configuration, validator)
		validateOIDCClientResponseTypes(c, configuration, validator)
		validateOIDCClientResponseModes(c, configuration, validator)

		validateOIDCClientRedirectURIs(client, validator)
	}

	if invalidID {
		validator.Push(fmt.Errorf("OIDC Server has one or more clients with an empty ID"))
	}

	if duplicateIDs {
		validator.Push(fmt.Errorf("OIDC Server has clients with duplicate ID's"))
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
		if !utils.IsStringInSlice(scope, validScopes) {
			validator.Push(fmt.Errorf(
				errFmtOIDCServerClientInvalidScope,
				configuration.Clients[c].ID, scope, strings.Join(validScopes, "', '")))
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
				errFmtOIDCServerClientInvalidGrantType,
				configuration.Clients[c].ID, grantType, strings.Join(validOIDCGrantTypes, "', '")))
		}
	}
}

func validateOIDCClientResponseTypes(c int, configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if len(configuration.Clients[c].ResponseTypes) == 0 {
		configuration.Clients[c].ResponseTypes = schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes
		return
	}

	for _, responseType := range configuration.Clients[c].ResponseTypes {
		if !utils.IsStringInSlice(responseType, validOIDCResponseTypes) {
			validator.Push(fmt.Errorf(
				errFmtOIDCServerClientInvalidResponseType,
				configuration.Clients[c].ID, responseType, strings.Join(validOIDCResponseTypes, "', '")))
		}
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
				errFmtOIDCServerClientInvalidResponseMode,
				configuration.Clients[c].ID, responseMode, strings.Join(validOIDCResponseModes, "', '")))
		}
	}
}

func validateOIDCClientRedirectURIs(client schema.OpenIDConnectClientConfiguration, validator *schema.StructValidator) {
	for _, redirectURI := range client.RedirectURIs {
		parsedURI, err := url.Parse(redirectURI)

		if err != nil {
			validator.Push(fmt.Errorf(errFmtOIDCServerClientRedirectURICantBeParsed, client.ID, redirectURI, err))
			break
		}

		if parsedURI.Scheme != schemeHTTPS && parsedURI.Scheme != schemeHTTP {
			validator.Push(fmt.Errorf(errFmtOIDCServerClientRedirectURI, client.ID, redirectURI, parsedURI.Scheme))
		}
	}
}
