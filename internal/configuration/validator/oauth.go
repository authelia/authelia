package validator

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateOAuth validates and update OAuth configuration.
func ValidateOAuth(configuration *schema.OAuthConfiguration, validator *schema.StructValidator) {
	validateOIDCServer(configuration.OIDCServer, validator)
}

func validateOIDCServer(configuration *schema.OpenIDConnectServerConfiguration, validator *schema.StructValidator) {
	if configuration != nil {
		if configuration.IssuerPrivateKey == "" {
			validator.Push(fmt.Errorf("OIDC Server issuer private key must be provided"))
		}

		if len(configuration.HMACSecret) != 32 {
			validator.Push(fmt.Errorf(errOAuthOIDCServerHMACLengthMustBe32Fmt, len(configuration.HMACSecret)))
		}

		validateOIDCServerClients(configuration, validator)

		if len(configuration.Clients) == 0 {
			validator.Push(fmt.Errorf("OIDC Server has no clients defined"))
		}
	}
}

func validateOIDCServerClients(configuration *schema.OpenIDConnectServerConfiguration, validator *schema.StructValidator) {
	invalidID, invalidSecret, invalidPolicy, duplicateIDs := false, false, false, false

	var ids []string

	for c, client := range configuration.Clients {
		if client.ID == "" {
			invalidID = true
		} else {
			if utils.IsStringInSliceFold(client.ID, ids) {
				duplicateIDs = true
			}
			ids = append(ids, client.ID)
		}

		if client.Secret == "" {
			invalidSecret = true
		}

		if client.Policy == "" {
			invalidPolicy = true
		}

		if len(client.Scopes) == 0 {
			configuration.Clients[c].Scopes = schema.DefaultOpenIDConnectClientConfiguration.Scopes
		}

		if len(client.GrantTypes) == 0 {
			configuration.Clients[c].GrantTypes = schema.DefaultOpenIDConnectClientConfiguration.GrantTypes
		}

		if len(client.ResponseTypes) == 0 {
			configuration.Clients[c].ResponseTypes = schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes
		}

		validateOIDCServerClientRedirectURIs(client, validator)
	}

	if invalidID {
		validator.Push(fmt.Errorf("OIDC Server has one or more clients with an empty ID"))
	}

	if invalidSecret {
		validator.Push(fmt.Errorf("OIDC Server has one or more clients with an empty secret"))
	}

	if invalidPolicy {
		validator.Push(fmt.Errorf("OIDC Server has one or more clients with an empty policy"))
	}

	if duplicateIDs {
		validator.Push(fmt.Errorf("OIDC Server has clients with duplicate ID's"))
	}
}

func validateOIDCServerClientRedirectURIs(client schema.OpenIDConnectClientConfiguration, validator *schema.StructValidator) {
	for _, redirectURI := range client.RedirectURIs {
		parsedURI, err := url.Parse(redirectURI)

		if err != nil {
			validator.Push(fmt.Errorf(errOAuthOIDCServerClientRedirectURICantBeParsedFmt, redirectURI, err))
			break
		}

		if parsedURI.Scheme != "https" {
			validator.Push(fmt.Errorf(errOAuthOIDCServerClientRedirectURIFmt, redirectURI, parsedURI.Scheme))
		}
	}
}
