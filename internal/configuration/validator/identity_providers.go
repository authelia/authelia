package validator

import (
	"fmt"
	"net/url"
	"time"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateIdentityProviders validates and update IdentityProviders configuration.
func ValidateIdentityProviders(externalURL string, configuration *schema.IdentityProvidersConfiguration, validator *schema.StructValidator) {
	validateOIDC(externalURL, configuration.OIDC, validator)
}

func validateOIDC(externalURL string, configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if configuration != nil {
		if externalURL == "" {
			validator.Push(fmt.Errorf("OIDC Provider cannot be configured without an external_url"))
		}

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
			validator.Push(fmt.Errorf(errIdentityProvidersOIDCServerClientInvalidSecFmt, client.ID))
		}

		if client.Policy == "" {
			configuration.Clients[c].Policy = schema.DefaultOpenIDConnectClientConfiguration.Policy
		} else if client.Policy != oneFactorPolicy && client.Policy != twoFactorPolicy {
			validator.Push(fmt.Errorf(errIdentityProvidersOIDCServerClientInvalidPolicyFmt, client.ID, client.Policy))
		}

		if len(client.Scopes) == 0 {
			configuration.Clients[c].Scopes = schema.DefaultOpenIDConnectClientConfiguration.Scopes
		} else if !utils.IsStringInSlice("openid", client.Scopes) {
			configuration.Clients[c].Scopes = append(configuration.Clients[c].Scopes, "openid")
		}

		if len(client.GrantTypes) == 0 {
			configuration.Clients[c].GrantTypes = schema.DefaultOpenIDConnectClientConfiguration.GrantTypes
		}

		if len(client.ResponseTypes) == 0 {
			configuration.Clients[c].ResponseTypes = schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes
		}

		validateOIDCClientRedirectURIs(client, validator)
	}

	if invalidID {
		validator.Push(fmt.Errorf("OIDC Server has one or more clients with an empty ID"))
	}

	if duplicateIDs {
		validator.Push(fmt.Errorf("OIDC Server has clients with duplicate ID's"))
	}
}

func validateOIDCClientRedirectURIs(client schema.OpenIDConnectClientConfiguration, validator *schema.StructValidator) {
	for _, redirectURI := range client.RedirectURIs {
		parsedURI, err := url.Parse(redirectURI)

		if err != nil {
			validator.Push(fmt.Errorf(errOAuthOIDCServerClientRedirectURICantBeParsedFmt, client.ID, redirectURI, err))
			break
		}

		if parsedURI.Scheme != "https" && parsedURI.Scheme != "http" {
			validator.Push(fmt.Errorf(errOAuthOIDCServerClientRedirectURIFmt, redirectURI, parsedURI.Scheme))
		}
	}
}
