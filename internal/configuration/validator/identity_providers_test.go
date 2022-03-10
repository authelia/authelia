package validator

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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

	assert.EqualError(t, validator.Errors()[0], errFmtOIDCNoPrivateKey)
	assert.EqualError(t, validator.Errors()[1], errFmtOIDCNoClientsConfigured)
}

func TestShouldRaiseErrorWhenOIDCPKCEEnforceValueInvalid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			EnforcePKCE:      "invalid",
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'enforce_pkce' must be 'never', 'public_clients_only' or 'always', but it is configured as 'invalid'")
	assert.EqualError(t, validator.Errors()[1], errFmtOIDCNoClientsConfigured)
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

	assert.EqualError(t, validator.Errors()[0], errFmtOIDCNoClientsConfigured)
}

func TestShouldRaiseErrorWhenOIDCServerClientBadValues(t *testing.T) {
	testCases := []struct {
		Name    string
		Clients []schema.OpenIDConnectClientConfiguration
		Errors  []error
	}{
		{
			Name: "empty",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:           "",
					Secret:       "",
					Policy:       "",
					RedirectURIs: []string{},
				},
			},
			Errors: []error{
				fmt.Errorf(errFmtOIDCClientInvalidSecret, ""),
				errors.New(errFmtOIDCClientsWithEmptyID),
			},
		},
		{
			Name: "client-1",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-1",
					Secret: "a-secret",
					Policy: "a-policy",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
			},
			Errors: []error{fmt.Errorf(errFmtOIDCClientInvalidPolicy, "client-1", "a-policy")},
		},
		{
			Name: "client-duplicate",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:           "client-x",
					Secret:       "a-secret",
					Policy:       policyTwoFactor,
					RedirectURIs: []string{},
				},
				{
					ID:           "client-x",
					Secret:       "a-secret",
					Policy:       policyTwoFactor,
					RedirectURIs: []string{},
				},
			},
			Errors: []error{errors.New(errFmtOIDCClientsDuplicateID)},
		},
		{
			Name: "client-check-uri-parse",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-check-uri-parse",
					Secret: "a-secret",
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"http://abc@%two",
					},
				},
			},
			Errors: []error{
				fmt.Errorf(errFmtOIDCClientRedirectURICantBeParsed, "client-check-uri-parse", "http://abc@%two", errors.New("parse \"http://abc@%two\": invalid URL escape \"%tw\"")),
			},
		},
		{
			Name: "client-check-uri-abs",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-check-uri-abs",
					Secret: "a-secret",
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"google.com",
					},
				},
			},
			Errors: []error{
				fmt.Errorf(errFmtOIDCClientRedirectURIAbsolute, "client-check-uri-abs", "google.com"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.IdentityProvidersConfiguration{
				OIDC: &schema.OpenIDConnectConfiguration{
					HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
					IssuerPrivateKey: "key-material",
					Clients:          tc.Clients,
				},
			}

			ValidateIdentityProviders(config, validator)

			assert.ElementsMatch(t, validator.Errors(), tc.Errors)
		})
	}
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'scopes' must only have the values 'openid', 'email', 'profile', 'groups', 'offline_access' but one option is configured as 'bad_scope'")
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'grant_types' must only have the values 'implicit', 'refresh_token', 'authorization_code', 'password', 'client_credentials' but one option is configured as 'bad_grant_type'")
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'response_modes' must only have the values 'form_post', 'query', 'fragment' but one option is configured as 'bad_responsemode'")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadUserinfoAlg(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:                       "good_id",
					Secret:                   "good_secret",
					Policy:                   "two_factor",
					UserinfoSigningAlgorithm: "rs256",
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'userinfo_signing_algorithm' must be one of 'none, RS256' but it is configured as 'rs256'")
}

func TestValidateIdentityProvidersShouldRaiseWarningOnSecurityIssue(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:              "abc",
			IssuerPrivateKey:        "abc",
			MinimumParameterEntropy: 1,
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "good_id",
					Secret: "good_secret",
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 1)

	assert.EqualError(t, validator.Warnings()[0], "openid connect provider: SECURITY ISSUE - minimum parameter entropy is configured to an unsafe value, it should be above 8 but it's configured to 1")
}

func TestValidateIdentityProvidersShouldRaiseErrorsOnInvalidClientTypes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "hmac1",
			IssuerPrivateKey: "key2",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-with-invalid-secret",
					Secret: "a-secret",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://localhost",
					},
				},
				{
					ID:     "client-with-bad-redirect-uri",
					Secret: "a-secret",
					Public: false,
					Policy: "two_factor",
					RedirectURIs: []string{
						oauth2InstalledApp,
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 2)
	assert.Len(t, validator.Warnings(), 0)

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtOIDCClientPublicInvalidSecret, "client-with-invalid-secret"))
	assert.EqualError(t, validator.Errors()[1], fmt.Sprintf(errFmtOIDCClientRedirectURIPublic, "client-with-bad-redirect-uri", oauth2InstalledApp))
}

func TestValidateIdentityProvidersShouldNotRaiseErrorsOnValidPublicClients(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "hmac1",
			IssuerPrivateKey: "key2",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "installed-app-client",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						oauth2InstalledApp,
					},
				},
				{
					ID:     "client-with-https-scheme",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://localhost:9000",
					},
				},
				{
					ID:     "client-with-loopback",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"http://127.0.0.1",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)
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
					ID:                       "b-client",
					Description:              "Normal Description",
					Secret:                   "b-client-secret",
					Policy:                   policyOneFactor,
					UserinfoSigningAlgorithm: "RS256",
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
	assert.Equal(t, policyTwoFactor, config.OIDC.Clients[0].Policy)
	assert.Equal(t, policyOneFactor, config.OIDC.Clients[1].Policy)

	assert.Equal(t, "none", config.OIDC.Clients[0].UserinfoSigningAlgorithm)
	assert.Equal(t, "RS256", config.OIDC.Clients[1].UserinfoSigningAlgorithm)

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
	assert.Equal(t, time.Minute*90, config.OIDC.RefreshTokenLifespan)
}

// All valid schemes are supported as defined in https://datatracker.ietf.org/doc/html/rfc8252#section-7.1
func TestValidateOIDCClientRedirectURIsSupportingPrivateUseURISchemes(t *testing.T) {
	conf := schema.OpenIDConnectClientConfiguration{
		ID: "owncloud",
		RedirectURIs: []string{
			"https://www.mywebsite.com",
			"http://www.mywebsite.com",
			"oc://ios.owncloud.com",
			// example given in the RFC https://datatracker.ietf.org/doc/html/rfc8252#section-7.1
			"com.example.app:/oauth2redirect/example-provider",
		},
	}

	t.Run("public", func(t *testing.T) {
		validator := schema.NewStructValidator()
		conf.Public = true
		validateOIDCClientRedirectURIs(conf, validator)

		assert.Len(t, validator.Warnings(), 0)
		assert.Len(t, validator.Errors(), 0)
	})

	t.Run("not public", func(t *testing.T) {
		validator := schema.NewStructValidator()
		conf.Public = false
		validateOIDCClientRedirectURIs(conf, validator)

		assert.Len(t, validator.Warnings(), 0)
		assert.Len(t, validator.Errors(), 2)
		assert.ElementsMatch(t, validator.Errors(), []error{
			errors.New("identity_providers: oidc: client 'owncloud': option 'redirect_uris' has an invalid value: redirect uri 'oc://ios.owncloud.com' must have a scheme of 'http' or 'https' but 'oc' is configured"),
			errors.New("identity_providers: oidc: client 'owncloud': option 'redirect_uris' has an invalid value: redirect uri 'com.example.app:/oauth2redirect/example-provider' must have a scheme of 'http' or 'https' but 'com.example.app' is configured"),
		})
	})
}
