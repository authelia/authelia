package validator

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldRaiseErrorWhenInvalidOIDCServerConfiguration(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "abc",
			IssuerPrivateKey: "",
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "OIDC Server issuer private key must be provided")
	assert.EqualError(t, validator.Errors()[1], "OIDC Server has no clients defined")
}

func TestShouldRaiseErrorWhenOIDCServerIssuerPrivateKeyPathInvalid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "OIDC Server has no clients defined")
}

func TestShouldRaiseErrorWhenOIDCServerClientBadValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "",
					Secret: "",
					Policy: "",
					RedirectURIs: []string{
						"tcp://google.com",
					},
				},
				{
					ID:     "a-client",
					Secret: "a-secret",
					Policy: "a-policy",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
				{
					ID:     "a-client",
					Secret: "a-secret",
					Policy: "a-policy",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
				{
					ID:     "client-check-uri-parse",
					Secret: "a-secret",
					Policy: twoFactorPolicy,
					RedirectURIs: []string{
						"http://abc@%two",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 7)

	assert.Equal(t, schema.DefaultOpenIDConnectClientConfiguration.Policy, config.OIDC.Clients[0].Policy)
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtOIDCServerClientInvalidSec, ""))
	assert.EqualError(t, validator.Errors()[1], fmt.Sprintf(errFmtOIDCServerClientRedirectURI, "", "tcp://google.com", "tcp"))
	assert.EqualError(t, validator.Errors()[2], fmt.Sprintf(errFmtOIDCServerClientInvalidPolicy, "a-client", "a-policy"))
	assert.EqualError(t, validator.Errors()[3], fmt.Sprintf(errFmtOIDCServerClientInvalidPolicy, "a-client", "a-policy"))
	assert.EqualError(t, validator.Errors()[4], fmt.Sprintf(errFmtOIDCServerClientRedirectURICantBeParsed, "client-check-uri-parse", "http://abc@%two", errors.New("parse \"http://abc@%two\": invalid URL escape \"%tw\"")))
	assert.EqualError(t, validator.Errors()[5], "OIDC Server has one or more clients with an empty ID")
	assert.EqualError(t, validator.Errors()[6], "OIDC Server has clients with duplicate ID's")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadScopes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "good_id",
					Secret: "good_secret",
					Policy: "two_factor",
					Scopes: []string{"openid", "bad_scope"},
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "OIDC Client with ID 'good_id' has an invalid scope "+
		"'bad_scope', must be one of: 'openid', 'email', 'profile', 'groups', 'offline_access'")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadGrantTypes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:         "good_id",
					Secret:     "good_secret",
					Policy:     "two_factor",
					GrantTypes: []string{"bad_grant_type"},
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "OIDC Client with ID 'good_id' has an invalid grant type "+
		"'bad_grant_type', must be one of: 'implicit', 'refresh_token', 'authorization_code', "+
		"'password', 'client_credentials'")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadResponseTypes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:            "good_id",
					Secret:        "good_secret",
					Policy:        "two_factor",
					ResponseTypes: []string{"bad_response_type"},
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "OIDC Client with ID 'good_id' has an invalid response type "+
		"'bad_response_type', must be one of: 'code', 'code id_token', 'id_token', 'token id_token', 'token', "+
		"'token id_token code'")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadResponseModes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:            "good_id",
					Secret:        "good_secret",
					Policy:        "two_factor",
					ResponseModes: []string{"bad_responsemode"},
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "OIDC Client with ID 'good_id' has an invalid response mode "+
		"'bad_responsemode', must be one of: 'form_post', 'query', 'fragment'")
}

func TestValidateIdentityProvidersShouldSetDefaultValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "../../../README.md",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "a-client",
					Secret: "a-client-secret",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
				{
					ID:          "b-client",
					Description: "Normal Description",
					Secret:      "b-client-secret",
					Policy:      oneFactorPolicy,
					RedirectURIs: []string{
						"https://google.com",
					},
					Scopes: []string{
						"groups",
					},
					GrantTypes: []string{
						"refresh_token",
					},
					ResponseTypes: []string{
						"token",
						"code",
					},
					ResponseModes: []string{
						"form_post",
						"fragment",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	// Assert Clients[0] Policy is set to the default, and the default doesn't override Clients[1]'s Policy.
	assert.Equal(t, config.OIDC.Clients[0].Policy, twoFactorPolicy)
	assert.Equal(t, config.OIDC.Clients[1].Policy, oneFactorPolicy)

	// Assert Clients[0] Description is set to the Clients[0] ID, and Clients[1]'s Description is not overridden.
	assert.Equal(t, config.OIDC.Clients[0].ID, config.OIDC.Clients[0].Description)
	assert.Equal(t, "Normal Description", config.OIDC.Clients[1].Description)

	// Assert Clients[0] ends up configured with the default Scopes.
	require.Len(t, config.OIDC.Clients[0].Scopes, 4)
	assert.Equal(t, "openid", config.OIDC.Clients[0].Scopes[0])
	assert.Equal(t, "groups", config.OIDC.Clients[0].Scopes[1])
	assert.Equal(t, "profile", config.OIDC.Clients[0].Scopes[2])
	assert.Equal(t, "email", config.OIDC.Clients[0].Scopes[3])

	// Assert Clients[1] ends up configured with the configured Scopes and the openid Scope.
	require.Len(t, config.OIDC.Clients[1].Scopes, 2)
	assert.Equal(t, "groups", config.OIDC.Clients[1].Scopes[0])
	assert.Equal(t, "openid", config.OIDC.Clients[1].Scopes[1])

	// Assert Clients[0] ends up configured with the default GrantTypes.
	require.Len(t, config.OIDC.Clients[0].GrantTypes, 2)
	assert.Equal(t, "refresh_token", config.OIDC.Clients[0].GrantTypes[0])
	assert.Equal(t, "authorization_code", config.OIDC.Clients[0].GrantTypes[1])

	// Assert Clients[1] ends up configured with only the configured GrantTypes.
	require.Len(t, config.OIDC.Clients[1].GrantTypes, 1)
	assert.Equal(t, "refresh_token", config.OIDC.Clients[1].GrantTypes[0])

	// Assert Clients[0] ends up configured with the default ResponseTypes.
	require.Len(t, config.OIDC.Clients[0].ResponseTypes, 1)
	assert.Equal(t, "code", config.OIDC.Clients[0].ResponseTypes[0])

	// Assert Clients[1] ends up configured only with the configured ResponseTypes.
	require.Len(t, config.OIDC.Clients[1].ResponseTypes, 2)
	assert.Equal(t, "token", config.OIDC.Clients[1].ResponseTypes[0])
	assert.Equal(t, "code", config.OIDC.Clients[1].ResponseTypes[1])

	// Assert Clients[0] ends up configured with the default ResponseModes.
	require.Len(t, config.OIDC.Clients[0].ResponseModes, 3)
	assert.Equal(t, "form_post", config.OIDC.Clients[0].ResponseModes[0])
	assert.Equal(t, "query", config.OIDC.Clients[0].ResponseModes[1])
	assert.Equal(t, "fragment", config.OIDC.Clients[0].ResponseModes[2])

	// Assert Clients[1] ends up configured only with the configured ResponseModes.
	require.Len(t, config.OIDC.Clients[1].ResponseModes, 2)
	assert.Equal(t, "form_post", config.OIDC.Clients[1].ResponseModes[0])
	assert.Equal(t, "fragment", config.OIDC.Clients[1].ResponseModes[1])

	assert.Equal(t, false, config.OIDC.EnableClientDebugMessages)
	assert.Equal(t, time.Hour, config.OIDC.AccessTokenLifespan)
	assert.Equal(t, time.Minute, config.OIDC.AuthorizeCodeLifespan)
	assert.Equal(t, time.Hour, config.OIDC.IDTokenLifespan)
	assert.Equal(t, time.Hour*24*30, config.OIDC.RefreshTokenLifespan)
}
