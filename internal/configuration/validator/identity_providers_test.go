package validator

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldRaiseErrorWhenInvalidOIDCServerConfiguration(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret: "abc",
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option `issuer_private_keys` or 'issuer_private_key' is required")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: option 'clients' must have one or more clients configured")
}

func TestShouldNotRaiseErrorWhenCORSEndpointsValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: keyRSA2048,
			CORS: schema.OpenIDConnectCORS{
				Endpoints: []string{oidc.EndpointAuthorization, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo},
			},
			Clients: []schema.OpenIDConnectClient{
				{
					ID:     "example",
					Secret: tOpenIDConnectPlainTextClientSecret,
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Errors(), 0)
}

func TestShouldRaiseErrorWhenCORSEndpointsNotValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: keyRSA2048,
			CORS: schema.OpenIDConnectCORS{
				Endpoints: []string{oidc.EndpointAuthorization, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo, "invalid_endpoint"},
			},
			Clients: []schema.OpenIDConnectClient{
				{
					ID:     "example",
					Secret: tOpenIDConnectPlainTextClientSecret,
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: cors: option 'endpoints' contains an invalid value 'invalid_endpoint': must be one of 'authorization', 'pushed-authorization-request', 'token', 'introspection', 'revocation', or 'userinfo'")
}

func TestShouldRaiseErrorWhenOIDCPKCEEnforceValueInvalid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: keyRSA2048,
			EnforcePKCE:      testInvalid,
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'enforce_pkce' must be 'never', 'public_clients_only' or 'always', but it's configured as 'invalid'")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: option 'clients' must have one or more clients configured")
}

func TestShouldRaiseErrorWhenOIDCCORSOriginsHasInvalidValues(t *testing.T) {
	validator := schema.NewStructValidator()

	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: keyRSA2048,
			CORS: schema.OpenIDConnectCORS{
				AllowedOrigins:                       utils.URLsFromStringSlice([]string{"https://example.com/", "https://site.example.com/subpath", "https://site.example.com?example=true", "*"}),
				AllowedOriginsFromClientRedirectURIs: true,
			},
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "myclient",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: "two_factor",
					RedirectURIs:        []string{"https://example.com/oauth2_callback", "https://localhost:566/callback", "http://an.example.com/callback", "file://a/file"},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 5)
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: cors: option 'allowed_origins' contains an invalid value 'https://example.com/' as it has a path: origins must only be scheme, hostname, and an optional port")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: cors: option 'allowed_origins' contains an invalid value 'https://site.example.com/subpath' as it has a path: origins must only be scheme, hostname, and an optional port")
	assert.EqualError(t, validator.Errors()[2], "identity_providers: oidc: cors: option 'allowed_origins' contains an invalid value 'https://site.example.com?example=true' as it has a query string: origins must only be scheme, hostname, and an optional port")
	assert.EqualError(t, validator.Errors()[3], "identity_providers: oidc: cors: option 'allowed_origins' contains the wildcard origin '*' with more than one origin but the wildcard origin must be defined by itself")
	assert.EqualError(t, validator.Errors()[4], "identity_providers: oidc: cors: option 'allowed_origins' contains the wildcard origin '*' cannot be specified with option 'allowed_origins_from_client_redirect_uris' enabled")

	require.Len(t, config.OIDC.CORS.AllowedOrigins, 6)
	assert.Equal(t, "*", config.OIDC.CORS.AllowedOrigins[3].String())
	assert.Equal(t, "https://example.com", config.OIDC.CORS.AllowedOrigins[4].String())
}

func TestShouldRaiseErrorWhenOIDCServerNoClients(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: keyRSA2048,
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'clients' must have one or more clients configured")
}

func TestShouldRaiseErrorWhenOIDCServerClientBadValues(t *testing.T) {
	mustParseURL := func(u string) url.URL {
		out, err := url.Parse(u)
		if err != nil {
			panic(err)
		}

		return *out
	}

	testCases := []struct {
		Name    string
		Clients []schema.OpenIDConnectClient
		Errors  []string
	}{
		{
			Name: "EmptyIDAndSecret",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "",
					Secret:              nil,
					AuthorizationPolicy: "",
					RedirectURIs:        []string{},
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: client '': option 'secret' is required",
				"identity_providers: oidc: clients: option 'id' is required but was absent on the clients in positions #1",
			},
		},
		{
			Name: "InvalidPolicy",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-1",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: "a-policy",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: client 'client-1': option 'authorization_policy' must be one of 'one_factor' or 'two_factor' but it's configured as 'a-policy'",
			},
		},
		{
			Name: "ClientIDDuplicated",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-x",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs:        []string{},
				},
				{
					ID:                  "client-x",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs:        []string{},
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: option 'id' must be unique for every client but one or more clients share the following 'id' values 'client-x'",
			},
		},
		{
			Name: "RedirectURIInvalid",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-check-uri-parse",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"http://abc@%two",
					},
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: client 'client-check-uri-parse': option 'redirect_uris' has an invalid value: redirect uri 'http://abc@%two' could not be parsed: parse \"http://abc@%two\": invalid URL escape \"%tw\"",
			},
		},
		{
			Name: "RedirectURINotAbsolute",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-check-uri-abs",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"google.com",
					},
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: client 'client-check-uri-abs': option 'redirect_uris' has an invalid value: redirect uri 'google.com' must have a scheme but it's absent",
			},
		},
		{
			Name: "ValidSectorIdentifier",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-valid-sector",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifier: mustParseURL(exampleDotCom),
				},
			},
		},
		{
			Name: "ValidSectorIdentifierWithPort",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-valid-sector",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifier: mustParseURL("example.com:2000"),
				},
			},
		},
		{
			Name: "InvalidSectorIdentifierInvalidURL",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-invalid-sector",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifier: mustParseURL("https://user:pass@example.com/path?query=abc#fragment"),
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a scheme with the value 'https'",
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a path with the value '/path'",
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a query with the value 'query=abc'",
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a fragment with the value 'fragment'",
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a username with the value 'user'",
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a password",
			},
		},
		{
			Name: "InvalidSectorIdentifierInvalidHost",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-invalid-sector",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifier: mustParseURL("example.com/path?query=abc#fragment"),
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier' with value 'example.com/path?query=abc#fragment': must be a URL with only the host component but appears to be invalid",
			},
		},
		{
			Name: "InvalidConsentMode",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-bad-consent-mode",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					ConsentMode: "cap",
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: client 'client-bad-consent-mode': consent: option 'mode' must be one of 'auto', 'implicit', 'explicit', 'pre-configured', or 'auto' but it's configured as 'cap'",
			},
		},
		{
			Name: "InvalidPKCEChallengeMethod",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-bad-pkce-mode",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					PKCEChallengeMethod: "abc",
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: client 'client-bad-pkce-mode': option 'pkce_challenge_method' must be one of 'plain' or 'S256' but it's configured as 'abc'",
			},
		},
		{
			Name: "InvalidPKCEChallengeMethodLowerCaseS256",
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-bad-pkce-mode-s256",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					PKCEChallengeMethod: "s256",
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: client 'client-bad-pkce-mode-s256': option 'pkce_challenge_method' must be one of 'plain' or 'S256' but it's configured as 's256'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.IdentityProviders{
				OIDC: &schema.OpenIDConnect{
					HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
					IssuerPrivateKey: keyRSA2048,
					Clients:          tc.Clients,
				},
			}

			ValidateIdentityProviders(config, validator)

			errs := validator.Errors()

			require.Len(t, errs, len(tc.Errors))
			for i, errStr := range tc.Errors {
				t.Run(fmt.Sprintf("Error%d", i+1), func(t *testing.T) {
					assert.EqualError(t, errs[i], errStr)
				})
			}
		})
	}
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadScopes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: keyRSA2048,
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "good_id",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: "two_factor",
					Scopes:              []string{"openid", "bad_scope"},
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: clients: client 'good_id': option 'scopes' must only have the values 'openid', 'email', 'profile', 'groups', or 'offline_access' but the values 'bad_scope' are present")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadGrantTypes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: keyRSA2048,
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "good_id",
					Secret:              tOpenIDConnectPBKDF2ClientSecret,
					AuthorizationPolicy: "two_factor",
					GrantTypes:          []string{"bad_grant_type"},
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: clients: client 'good_id': option 'grant_types' must only have the values 'authorization_code', 'implicit', 'client_credentials', or 'refresh_token' but the values 'bad_grant_type' are present")
}

func TestShouldNotErrorOnCertificateValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:             "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerCertificateChain: certRSA2048,
			IssuerPrivateKey:       keyRSA2048,
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "good_id",
					Secret:              tOpenIDConnectPBKDF2ClientSecret,
					AuthorizationPolicy: "two_factor",
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)
}

func TestShouldRaiseErrorOnCertificateNotValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:             "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerCertificateChain: certRSA2048,
			IssuerPrivateKey:       keyRSA4096,
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "good_id",
					Secret:              tOpenIDConnectPBKDF2ClientSecret,
					AuthorizationPolicy: "two_factor",
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: issuer_private_keys: key #1 with key id 'bf1e10': option 'certificate_chain' does not appear to contain the public key for the private key provided by option 'key'")
}

func TestValidateIdentityProvidersOpenIDConnectMinimumParameterEntropy(t *testing.T) {
	testCases := []struct {
		name     string
		have     int
		expected int
		warnings []string
		errors   []string
	}{
		{
			"ShouldNotOverrideCustomValue",
			20,
			20,
			nil,
			nil,
		},
		{
			"ShouldSetDefault",
			0,
			8,
			nil,
			nil,
		},
		{
			"ShouldSetDefaultNegative",
			-2,
			8,
			nil,
			nil,
		},
		{
			"ShouldAllowDisabledAndWarn",
			-1,
			-1,
			[]string{"identity_providers: oidc: option 'minimum_parameter_entropy' is disabled which is considered unsafe and insecure"},
			nil,
		},
		{
			"ShouldWarnOnTooLow",
			2,
			2,
			[]string{"identity_providers: oidc: option 'minimum_parameter_entropy' is configured to an unsafe and insecure value, it should at least be 8 but it's configured to 2"},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.IdentityProviders{
				OIDC: &schema.OpenIDConnect{
					HMACSecret:              "abc",
					IssuerPrivateKey:        keyRSA2048,
					MinimumParameterEntropy: tc.have,
					Clients: []schema.OpenIDConnectClient{
						{
							ID:                  "good_id",
							Secret:              tOpenIDConnectPBKDF2ClientSecret,
							AuthorizationPolicy: "two_factor",
							RedirectURIs: []string{
								"https://google.com/callback",
							},
						},
					},
				},
			}

			ValidateIdentityProviders(config, validator)

			assert.Equal(t, tc.expected, config.OIDC.MinimumParameterEntropy)

			if n := len(tc.warnings); n == 0 {
				assert.Len(t, validator.Warnings(), 0)
			} else {
				require.Len(t, validator.Warnings(), n)

				for i := 0; i < n; i++ {
					assert.EqualError(t, validator.Warnings()[i], tc.warnings[i])
				}
			}

			if n := len(tc.errors); n == 0 {
				assert.Len(t, validator.Errors(), 0)
			} else {
				require.Len(t, validator.Errors(), n)

				for i := 0; i < n; i++ {
					assert.EqualError(t, validator.Errors()[i], tc.errors[i])
				}
			}
		})
	}
}

func TestValidateIdentityProvidersShouldRaiseErrorsOnInvalidClientTypes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:       "hmac1",
			IssuerPrivateKey: keyRSA2048,
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-with-invalid-secret",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					Public:              true,
					AuthorizationPolicy: "two_factor",
					RedirectURIs: []string{
						"https://localhost",
					},
				},
				{
					ID:                  "client-with-bad-redirect-uri",
					Secret:              tOpenIDConnectPBKDF2ClientSecret,
					Public:              false,
					AuthorizationPolicy: "two_factor",
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

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: clients: client 'client-with-invalid-secret': option 'secret' is required to be empty when option 'public' is true")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: clients: client 'client-with-bad-redirect-uri': option 'redirect_uris' has the redirect uri 'urn:ietf:wg:oauth:2.0:oob' when option 'public' is false but this is invalid as this uri is not valid for the openid connect confidential client type")
}

func TestValidateIdentityProvidersShouldNotRaiseErrorsOnValidClientOptions(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:       "hmac1",
			IssuerPrivateKey: keyRSA2048,
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "installed-app-client",
					Public:              true,
					AuthorizationPolicy: "two_factor",
					RedirectURIs: []string{
						oauth2InstalledApp,
					},
				},
				{
					ID:                  "client-with-https-scheme",
					Public:              true,
					AuthorizationPolicy: "two_factor",
					RedirectURIs: []string{
						"https://localhost:9000",
					},
				},
				{
					ID:                  "client-with-loopback",
					Public:              true,
					AuthorizationPolicy: "two_factor",
					RedirectURIs: []string{
						"http://127.0.0.1",
					},
				},
				{
					ID:                  "client-with-pkce-mode-plain",
					Public:              true,
					AuthorizationPolicy: "two_factor",
					RedirectURIs: []string{
						"https://pkce.com",
					},
					PKCEChallengeMethod: "plain",
				},
				{
					ID:                  "client-with-pkce-mode-S256",
					Public:              true,
					AuthorizationPolicy: "two_factor",
					RedirectURIs: []string{
						"https://pkce.com",
					},
					PKCEChallengeMethod: "S256",
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)
}

func TestValidateIdentityProvidersShouldRaiseWarningOnPlainTextClients(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProviders{
		OIDC: &schema.OpenIDConnect{
			HMACSecret:       "hmac1",
			IssuerPrivateKey: keyRSA2048,
			Clients: []schema.OpenIDConnectClient{
				{
					ID:                  "client-with-invalid-secret_standard",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: "two_factor",
					RedirectURIs: []string{
						"https://localhost",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 1)

	assert.EqualError(t, validator.Warnings()[0], "identity_providers: oidc: clients: client 'client-with-invalid-secret_standard': option 'secret' is plaintext but for clients not using the 'token_endpoint_auth_method' of 'client_secret_jwt' it should be a hashed value as plaintext values are deprecated with the exception of 'client_secret_jwt' and will be removed when oidc becomes stable")
}

// All valid schemes are supported as defined in https://datatracker.ietf.org/doc/html/rfc8252#section-7.1
func TestValidateOIDCClientRedirectURIsSupportingPrivateUseURISchemes(t *testing.T) {
	have := &schema.OpenIDConnect{
		Clients: []schema.OpenIDConnectClient{
			{
				ID: "owncloud",
				RedirectURIs: []string{
					"https://www.mywebsite.com",
					"http://www.mywebsite.com",
					"oc://ios.owncloud.com",
					// example given in the RFC https://datatracker.ietf.org/doc/html/rfc8252#section-7.1
					"com.example.app:/oauth2redirect/example-provider",
					oauth2InstalledApp,
				},
			},
		},
	}

	t.Run("public", func(t *testing.T) {
		validator := schema.NewStructValidator()
		have.Clients[0].Public = true
		validateOIDCClientRedirectURIs(0, have, validator, nil)

		assert.Len(t, validator.Warnings(), 0)
		assert.Len(t, validator.Errors(), 0)
	})

	t.Run("not public", func(t *testing.T) {
		validator := schema.NewStructValidator()
		have.Clients[0].Public = false
		validateOIDCClientRedirectURIs(0, have, validator, nil)

		assert.Len(t, validator.Warnings(), 0)
		assert.Len(t, validator.Errors(), 1)
		assert.ElementsMatch(t, validator.Errors(), []error{
			errors.New("identity_providers: oidc: clients: client 'owncloud': option 'redirect_uris' has the redirect uri 'urn:ietf:wg:oauth:2.0:oob' when option 'public' is false but this is invalid as this uri is not valid for the openid connect confidential client type"),
		})
	})
}

func TestValidateOIDCClients(t *testing.T) {
	type tcv struct {
		Scopes        []string
		ResponseTypes []string
		ResponseModes []string
		GrantTypes    []string
	}

	testCasses := []struct {
		name     string
		setup    func(have *schema.OpenIDConnect)
		validate func(t *testing.T, have *schema.OpenIDConnect)
		have     tcv
		expected tcv
		serrs    []string // Soft errors which will be warnings before GA.
		errs     []string
	}{
		{
			"ShouldSetDefaultResponseTypeAndResponseModes",
			nil,
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldIncludeMinimalScope",
			nil,
			nil,
			tcv{
				[]string{oidc.ScopeEmail},
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultResponseModesFlowAuthorizeCode",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultResponseModesFlowImplicit",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeImplicitFlowBoth},
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeImplicitFlowBoth},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeImplicit},
			},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultResponseModesFlowHybrid",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeHybridFlowBoth},
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeHybridFlowBoth},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit},
			},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultResponseModesFlowMixedAuthorizeCodeHybrid",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeHybridFlowBoth},
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeHybridFlowBoth},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit},
			},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultResponseModesFlowMixedAuthorizeCodeImplicit",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowBoth},
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowBoth},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit},
			},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultResponseModesFlowMixedAll",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowBoth, oidc.ResponseTypeHybridFlowBoth},
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowBoth, oidc.ResponseTypeHybridFlowBoth},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit},
			},
			nil,
			nil,
		},
		{
			"ShouldNotOverrideValues",
			nil,
			nil,
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowBoth, oidc.ResponseTypeHybridFlowBoth},
				[]string{oidc.ResponseModeFormPost},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowBoth, oidc.ResponseTypeHybridFlowBoth},
				[]string{oidc.ResponseModeFormPost},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldRaiseErrorOnDuplicateScopes",
			nil,
			nil,
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeOpenID},
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeOpenID},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'scopes' must have unique values but the values 'openid' are duplicated",
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidScopes",
			nil,
			nil,
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeProfile, "group"},
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeProfile, "group"},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'scopes' must only have the values 'openid', 'email', 'profile', 'groups', or 'offline_access' but the values 'group' are present",
			},
		},
		{
			"ShouldRaiseErrorOnMissingAuthorizationCodeFlowResponseTypeWithRefreshTokenValues",
			nil,
			nil,
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeOfflineAccess},
				[]string{oidc.ResponseTypeImplicitFlowBoth},
				nil,
				[]string{oidc.GrantTypeImplicit, oidc.GrantTypeRefreshToken},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeOfflineAccess},
				[]string{oidc.ResponseTypeImplicitFlowBoth},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeImplicit, oidc.GrantTypeRefreshToken},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'scopes' should only have the values 'offline_access' or 'offline' if the client is also configured with a 'response_type' such as 'code', 'code id_token', 'code token', or 'code id_token token' which respond with authorization codes",
				"identity_providers: oidc: clients: client 'test': option 'grant_types' should only have the values 'refresh_token' if the client is also configured with a 'response_type' such as 'code', 'code id_token', 'code token', or 'code id_token token' which respond with authorization codes",
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnDuplicateResponseTypes",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowBoth, oidc.ResponseTypeAuthorizationCodeFlow},
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowBoth, oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'response_types' must have unique values but the values 'code' are duplicated",
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidResponseTypesOrder",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeImplicitFlowBoth, "token id_token"},
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeImplicitFlowBoth, "token id_token"},
				[]string{"form_post", "fragment"},
				[]string{"implicit"},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'response_types' must only have the values 'code', 'id_token', 'token', 'id_token token', 'code id_token', 'code token', or 'code id_token token' but the values 'token id_token' are present",
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidResponseTypes",
			nil,
			nil,
			tcv{
				nil,
				[]string{"not_valid"},
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{"not_valid"},
				[]string{oidc.ResponseModeFormPost},
				nil,
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'response_types' must only have the values 'code', 'id_token', 'token', 'id_token token', 'code id_token', 'code token', or 'code id_token token' but the values 'not_valid' are present",
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidResponseModes",
			nil,
			nil,
			tcv{
				nil,
				nil,
				[]string{"not_valid"},
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{"not_valid"},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'response_modes' must only have the values 'form_post', 'query', or 'fragment' but the values 'not_valid' are present",
			},
		},
		{
			"ShouldRaiseErrorOnDuplicateResponseModes",
			nil,
			nil,
			tcv{
				nil,
				nil,
				[]string{oidc.ResponseModeQuery, oidc.ResponseModeQuery},
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeQuery, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'response_modes' must have unique values but the values 'query' are duplicated",
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidGrantTypes",
			nil,
			nil,
			tcv{
				nil,
				nil,
				nil,
				[]string{"invalid"},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{"invalid"},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' must only have the values 'authorization_code', 'implicit', 'client_credentials', or 'refresh_token' but the values 'invalid' are present",
			},
		},
		{
			"ShouldRaiseErrorOnDuplicateGrantTypes",
			nil,
			nil,
			tcv{
				nil,
				nil,
				nil,
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeAuthorizationCode},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeAuthorizationCode},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' must have unique values but the values 'authorization_code' are duplicated",
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidGrantTypesForPublicClient",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].Public = true
				have.Clients[0].Secret = nil
				have.Clients[0].Scopes = []string{"abc", "123"}
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				[]string{oidc.GrantTypeClientCredentials},
			},
			tcv{
				[]string{"abc", "123"},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeClientCredentials},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' should only have the 'client_credentials' value if it is of the confidential client type but it's of the public client type",
			},
		},
		{
			"ShouldNotRaiseErrorOnValidGrantTypesForConfidentialClient",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].Public = false
				have.Clients[0].Scopes = []string{"scope1"}
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				[]string{oidc.GrantTypeClientCredentials},
			},
			tcv{
				[]string{"scope1"},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeClientCredentials},
			},
			nil,
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidScopeGrantTypesForConfidentialClient",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].Public = false
				have.Clients[0].Scopes = []string{oidc.ScopeOpenID, oidc.ScopeOffline, oidc.ScopeOfflineAccess}
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				[]string{oidc.GrantTypeClientCredentials},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeOffline, oidc.ScopeOfflineAccess},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeClientCredentials},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'scopes' should only have the values 'offline_access' or 'offline' if the client is also configured with a 'response_type' such as 'code', 'code id_token', 'code token', or 'code id_token token' which respond with authorization codes",
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'scopes' has the values 'openid', 'offline', and 'offline_access' however when exclusively utilizing the 'client_credentials' value for the 'grant_types' the values 'openid', 'offline', or 'offline_access' are not allowed",
			},
		},
		{
			"ShouldNotRestrictRefreshOpenIDScopesWithMultipleGrantTypesAndAllowCustomClientCredentials",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].Public = false
				have.Clients[0].Scopes = []string{oidc.ScopeOpenID, oidc.ScopeOffline, oidc.ScopeOfflineAccess, "custom"}
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				[]string{oidc.GrantTypeClientCredentials, oidc.GrantTypeImplicit},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeOffline, oidc.ScopeOfflineAccess, "custom"},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeClientCredentials, oidc.GrantTypeImplicit},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'scopes' should only have the values 'offline_access' or 'offline' if the client is also configured with a 'response_type' such as 'code', 'code id_token', 'code token', or 'code id_token token' which respond with authorization codes",
				"identity_providers: oidc: clients: client 'test': option 'grant_types' should only have grant type values which are valid with the configured 'response_types' for the client but 'implicit' expects a response type for either the implicit or hybrid flow such as 'id_token', 'token', 'id_token token', 'code id_token', 'code token', or 'code id_token token' but the response types are 'code'",
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnGrantTypeRefreshTokenWithoutScopeOfflineAccess",
			nil,
			nil,
			tcv{
				nil,
				nil,
				nil,
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' should only have the 'refresh_token' value if the client is also configured with the 'offline_access' scope",
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnGrantTypeAuthorizationCodeWithoutAuthorizationCodeOrHybridFlow",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeImplicitFlowBoth},
				nil,
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeImplicitFlowBoth},
				[]string{"form_post", "fragment"},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' should only have grant type values which are valid with the configured 'response_types' for the client but 'authorization_code' expects a response type for either the authorization code or hybrid flow such as 'code', 'code id_token', 'code token', or 'code id_token token' but the response types are 'id_token token'",
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnGrantTypeImplicitWithoutImplicitOrHybridFlow",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				nil,
				[]string{oidc.GrantTypeImplicit},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeImplicit},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' should only have grant type values which are valid with the configured 'response_types' for the client but 'implicit' expects a response type for either the implicit or hybrid flow such as 'id_token', 'token', 'id_token token', 'code id_token', 'code token', or 'code id_token token' but the response types are 'code'",
			},
			nil,
		},
		{
			"ShouldValidateCorrectRedirectURIsConfidentialClientType",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].RedirectURIs = []string{
					"https://google.com",
				}
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, []string{"https://google.com"}, have.Clients[0].RedirectURIs)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldValidateCorrectRedirectURIsPublicClientType",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].Public = true
				have.Clients[0].Secret = nil
				have.Clients[0].RedirectURIs = []string{
					oauth2InstalledApp,
				}
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, []string{oauth2InstalledApp}, have.Clients[0].RedirectURIs)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidRedirectURIsPublicOnly",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].RedirectURIs = []string{
					"urn:ietf:wg:oauth:2.0:oob",
				}
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, []string{oauth2InstalledApp}, have.Clients[0].RedirectURIs)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has the redirect uri 'urn:ietf:wg:oauth:2.0:oob' when option 'public' is false but this is invalid as this uri is not valid for the openid connect confidential client type",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidRedirectURIsMalformedURI",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].RedirectURIs = []string{
					"http://abc@%two",
				}
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, []string{"http://abc@%two"}, have.Clients[0].RedirectURIs)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has an invalid value: redirect uri 'http://abc@%two' could not be parsed: parse \"http://abc@%two\": invalid URL escape \"%tw\"",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidRedirectURIsNotAbsolute",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].RedirectURIs = []string{
					"google.com",
				}
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, []string{"google.com"}, have.Clients[0].RedirectURIs)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has an invalid value: redirect uri 'google.com' must have a scheme but it's absent",
			},
		},
		{
			"ShouldRaiseErrorOnDuplicateRedirectURI",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].RedirectURIs = []string{
					"https://google.com",
					"https://google.com",
				}
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, []string{"https://google.com", "https://google.com"}, have.Clients[0].RedirectURIs)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' must have unique values but the values 'https://google.com' are duplicated",
			},
			nil,
		},
		{
			"ShouldNotSetDefaultTokenEndpointClientAuthMethodConfidentialClientType",
			nil,
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, "", have.Clients[0].TokenEndpointAuthMethod)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldNotOverrideValidClientAuthMethod",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretPost
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, oidc.ClientAuthMethodClientSecretPost, have.Clients[0].TokenEndpointAuthMethod)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidClientAuthMethod",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.GrantTypeClientCredentials
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, oidc.GrantTypeClientCredentials, have.Clients[0].TokenEndpointAuthMethod)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_method' must be one of 'none', 'client_secret_post', 'client_secret_basic', 'private_key_jwt', or 'client_secret_jwt' but it's configured as 'client_credentials'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidClientAuthMethodForPublicClientType",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretBasic
				have.Clients[0].Public = true
				have.Clients[0].Secret = nil
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, oidc.ClientAuthMethodClientSecretBasic, have.Clients[0].TokenEndpointAuthMethod)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_method' must be 'none' when configured as the public client type but it's configured as 'client_secret_basic'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidClientAuthMethodForConfidentialClientTypeAuthorizationCodeFlow",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodNone
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, oidc.ClientAuthMethodNone, have.Clients[0].TokenEndpointAuthMethod)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_method' must be one of 'client_secret_post', 'client_secret_basic', or 'private_key_jwt' when configured as the confidential client type unless it only includes implicit flow response types such as 'id_token', 'token', and 'id_token token' but it's configured as 'none'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidClientAuthMethodForConfidentialClientTypeHybridFlow",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodNone
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, oidc.ClientAuthMethodNone, have.Clients[0].TokenEndpointAuthMethod)
			},
			tcv{
				nil,
				[]string{oidc.ResponseTypeHybridFlowToken},
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeHybridFlowToken},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_method' must be one of 'client_secret_post', 'client_secret_basic', or 'private_key_jwt' when configured as the confidential client type unless it only includes implicit flow response types such as 'id_token', 'token', and 'id_token token' but it's configured as 'none'",
			},
		},
		{
			"ShouldSetDefaultUserInfoAlg",
			nil,
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgNone, have.Clients[0].UserinfoSigningAlg)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldNotOverrideUserInfoAlg",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].UserinfoSigningAlg = oidc.SigningAlgRSAUsingSHA256
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].UserinfoSigningAlg)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidUserInfoSigningAlg",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].UserinfoSigningAlg = rs256
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, rs256, have.Clients[0].UserinfoSigningAlg)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'userinfo_signing_alg' must be one of 'RS256' or 'none' but it's configured as 'rs256'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidIDTokenSigningAlg",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].IDTokenSigningAlg = rs256
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, rs256, have.Clients[0].IDTokenSigningAlg)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'id_token_signing_alg' must be one of 'RS256' but it's configured as 'rs256'",
			},
		},
		{
			"ShouldSetDefaultConsentMode",
			nil,
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, "explicit", have.Clients[0].ConsentMode)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultConsentModeAuto",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].ConsentMode = auto
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, "explicit", have.Clients[0].ConsentMode)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultConsentModePreConfigured",
			func(have *schema.OpenIDConnect) {
				d := time.Minute

				have.Clients[0].ConsentMode = ""
				have.Clients[0].ConsentPreConfiguredDuration = &d
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, "pre-configured", have.Clients[0].ConsentMode)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultConsentModeAutoPreConfigured",
			func(have *schema.OpenIDConnect) {
				d := time.Minute

				have.Clients[0].ConsentMode = auto
				have.Clients[0].ConsentPreConfiguredDuration = &d
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, "pre-configured", have.Clients[0].ConsentMode)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldNotOverrideConsentMode",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].ConsentMode = "implicit"
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, "implicit", have.Clients[0].ConsentMode)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldSentConsentPreConfiguredDefaultDuration",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].ConsentMode = "pre-configured"
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, "pre-configured", have.Clients[0].ConsentMode)
				assert.Equal(t, schema.DefaultOpenIDConnectClientConfiguration.ConsentPreConfiguredDuration, have.Clients[0].ConsentPreConfiguredDuration)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldRaiseErrorOnTokenEndpointClientAuthMethodPrivateKeyJWTMustSetAlg",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodPrivateKeyJWT
				have.Clients[0].Secret = tOpenIDConnectPBKDF2ClientSecret
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_signing_alg' is required when option 'token_endpoint_auth_method' is configured to 'private_key_jwt'",
				"identity_providers: oidc: clients: client 'test': option 'public_keys' is required with 'token_endpoint_auth_method' set to 'private_key_jwt'",
			},
		},
		{
			"ShouldRaiseErrorOnTokenEndpointClientAuthMethodPrivateKeyJWTMustSetAlg",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodPrivateKeyJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = "nope"
				have.Clients[0].Secret = tOpenIDConnectPBKDF2ClientSecret
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_signing_alg' must be one of 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', or 'ES512' when option 'token_endpoint_auth_method' is configured to 'private_key_jwt'",
				"identity_providers: oidc: clients: client 'test': option 'public_keys' is required with 'token_endpoint_auth_method' set to 'private_key_jwt'",
			},
		},
		{
			"ShouldRaiseErrorOnTokenEndpointClientAuthMethodPrivateKeyJWTMustSetKnownAlg",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodPrivateKeyJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = oidc.SigningAlgECDSAUsingP384AndSHA384
				have.Clients[0].Secret = tOpenIDConnectPBKDF2ClientSecret
				have.Clients[0].PublicKeys.Values = []schema.JWK{
					{
						KeyID:     "test",
						Key:       keyRSA2048.Public(),
						Algorithm: oidc.SigningAlgRSAUsingSHA256,
					},
				}
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_signing_alg' must be one of registered public key algorithm values 'RS256' when option 'token_endpoint_auth_method' is configured to 'private_key_jwt'",
			},
		},
		{
			"ShouldRaiseErrorOnTokenEndpointClientAuthMethodPrivateKeyJWTMustSetKnownAlg",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodPrivateKeyJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = oidc.SigningAlgECDSAUsingP384AndSHA384
				have.Clients[0].Secret = tOpenIDConnectPBKDF2ClientSecret
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'public_keys' is required with 'token_endpoint_auth_method' set to 'private_key_jwt'",
			},
		},
		{
			"ShouldRaiseErrorOnIncorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].Secret = tOpenIDConnectPBKDF2ClientSecret
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'secret' must be plaintext with option 'token_endpoint_auth_method' with a value of 'client_secret_jwt'",
			},
		},
		{
			"ShouldNotRaiseWarningOrErrorOnCorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].Secret = tOpenIDConnectPlainTextClientSecret
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldRaiseErrorOnIncorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].Secret = MustDecodeSecret("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng")
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'secret' must be plaintext with option 'token_endpoint_auth_method' with a value of 'client_secret_jwt'",
			},
		},
		{
			"ShouldNotRaiseWarningOrErrorOnCorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].Secret = MustDecodeSecret("$plaintext$abc123")
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldSetValidDefaultKeyID",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].IDTokenSigningKeyID = "abcabc123"
				have.Clients[0].UserinfoSigningKeyID = "abc123abc"
				have.Discovery.ResponseObjectSigningKeyIDs = []string{"abcabc123", "abc123abc"}
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidKeyID",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].IDTokenSigningKeyID = "ab"
				have.Clients[0].UserinfoSigningKeyID = "cd"
				have.Discovery.ResponseObjectSigningKeyIDs = []string{"abc123xyz"}
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'userinfo_signing_key_id' must be one of 'abc123xyz' but it's configured as 'cd'",
				"identity_providers: oidc: clients: client 'test': option 'id_token_signing_key_id' must be one of 'abc123xyz' but it's configured as 'ab'",
			},
		},
		{
			"ShouldSetDefaultTokenEndpointAuthSigAlg",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].Secret = tOpenIDConnectPlainTextClientSecret
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgHMACUsingSHA256, have.Clients[0].TokenEndpointAuthSigningAlg)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			nil,
		},
		{
			"ShouldRaiseErrorOnInvalidPublicTokenAuthAlg",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = oidc.SigningAlgHMACUsingSHA256
				have.Clients[0].Secret = nil
				have.Clients[0].Public = true
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgHMACUsingSHA256, have.Clients[0].TokenEndpointAuthSigningAlg)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_method' must be 'none' when configured as the public client type but it's configured as 'client_secret_jwt'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidTokenAuthAlgClientTypeConfidential",
			func(have *schema.OpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = oidc.EndpointToken
				have.Clients[0].Secret = tOpenIDConnectPlainTextClientSecret
			},
			func(t *testing.T, have *schema.OpenIDConnect) {
				assert.Equal(t, oidc.EndpointToken, have.Clients[0].TokenEndpointAuthSigningAlg)
			},
			tcv{
				nil,
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_signing_alg' must be one of 'HS256', 'HS384', or 'HS512' when option 'token_endpoint_auth_method' is configured to 'client_secret_jwt'",
			},
		},
	}

	errDeprecatedFunc := func() {}

	for _, tc := range testCasses {
		t.Run(tc.name, func(t *testing.T) {
			have := &schema.OpenIDConnect{
				Discovery: schema.OpenIDConnectDiscovery{
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
				Clients: []schema.OpenIDConnectClient{
					{
						ID:            "test",
						Secret:        tOpenIDConnectPBKDF2ClientSecret,
						Scopes:        tc.have.Scopes,
						ResponseModes: tc.have.ResponseModes,
						ResponseTypes: tc.have.ResponseTypes,
						GrantTypes:    tc.have.GrantTypes,
					},
				},
			}

			if tc.setup != nil {
				tc.setup(have)
			}

			val := schema.NewStructValidator()

			validateOIDCClient(0, have, val, errDeprecatedFunc)

			t.Run("General", func(t *testing.T) {
				assert.Equal(t, tc.expected.Scopes, have.Clients[0].Scopes)
				assert.Equal(t, tc.expected.ResponseTypes, have.Clients[0].ResponseTypes)
				assert.Equal(t, tc.expected.ResponseModes, have.Clients[0].ResponseModes)
				assert.Equal(t, tc.expected.GrantTypes, have.Clients[0].GrantTypes)

				if tc.validate != nil {
					tc.validate(t, have)
				}
			})

			t.Run("Warnings", func(t *testing.T) {
				require.Len(t, val.Warnings(), len(tc.serrs))
				for i, err := range tc.serrs {
					assert.EqualError(t, val.Warnings()[i], err)
				}
			})

			t.Run("Errors", func(t *testing.T) {
				require.Len(t, val.Errors(), len(tc.errs))
				for i, err := range tc.errs {
					assert.EqualError(t, val.Errors()[i], err)
				}
			})
		})
	}
}

func TestValidateOIDCClientTokenEndpointAuthMethod(t *testing.T) {
	testCasses := []struct {
		name     string
		have     string
		public   bool
		expected string
		errs     []string
	}{
		{"ShouldSetDefaultValueConfidential", "", false, "", nil},
		{"ShouldErrorOnInvalidValue", "abc", false, "abc",
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_method' must be one of 'none', 'client_secret_post', 'client_secret_basic', 'private_key_jwt', or 'client_secret_jwt' but it's configured as 'abc'",
			},
		},
		{"ShouldErrorOnInvalidValueForPublicClient", "client_secret_post", true, "client_secret_post",
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_method' must be 'none' when configured as the public client type but it's configured as 'client_secret_post'",
			},
		},
		{"ShouldErrorOnInvalidValueForConfidentialClient", "none", false, "none",
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_method' must be one of 'client_secret_post', 'client_secret_basic', or 'private_key_jwt' when configured as the confidential client type unless it only includes implicit flow response types such as 'id_token', 'token', and 'id_token token' but it's configured as 'none'",
			},
		},
	}

	for _, tc := range testCasses {
		t.Run(tc.name, func(t *testing.T) {
			have := &schema.OpenIDConnect{
				Clients: []schema.OpenIDConnectClient{
					{
						ID:                      "test",
						Public:                  tc.public,
						TokenEndpointAuthMethod: tc.have,
					},
				},
			}

			val := schema.NewStructValidator()

			validateOIDCClientTokenEndpointAuth(0, have, val)

			assert.Equal(t, tc.expected, have.Clients[0].TokenEndpointAuthMethod)
			assert.Len(t, val.Warnings(), 0)
			require.Len(t, val.Errors(), len(tc.errs))

			if tc.errs != nil {
				for i, err := range tc.errs {
					assert.EqualError(t, val.Errors()[i], err)
				}
			}
		})
	}
}

func TestValidateOIDCClientJWKS(t *testing.T) {
	frankenchain := schema.NewX509CertificateChainFromCerts([]*x509.Certificate{certRSA2048.Leaf(), certRSA1024.Leaf()})
	frankenkey := &rsa.PrivateKey{}

	*frankenkey = *keyRSA2048

	frankenkey.PublicKey.N = nil

	testCases := []struct {
		name     string
		haveURI  *url.URL
		haveJWKS []schema.JWK
		setup    func(config *schema.OpenIDConnect)
		expected func(t *testing.T, config *schema.OpenIDConnect)
		errs     []string
	}{
		{
			"ShouldValidateURL",
			MustParseURL("https://example.com"),
			nil,
			nil,
			nil,
			nil,
		},
		{
			"ShouldErrorOnHTTPURL",
			MustParseURL("http://example.com"),
			nil,
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: option 'uri' must have the 'https' scheme but the scheme is 'http'",
			},
		},
		{
			"ShouldErrorOnBothDefined",
			MustParseURL("http://example.com"),
			[]schema.JWK{
				{KeyID: "test"},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: option 'uri' must not be defined at the same time as option 'values'",
			},
		},
		{
			"ShouldAllowGoodKey",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048PKCS8.Public()},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldAllowGoodKeyWithCertificate",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048.Public(), CertificateChain: certRSA2048},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldErrorOnPrivateKey",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048PKCS8},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: values: key #1 with key id 'test': option 'key' must be a RSA public key or ECDSA public key but it's type is *rsa.PrivateKey",
			},
		},
		{
			"ShouldErrorOnMissingKID",
			nil,
			[]schema.JWK{
				{Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048PKCS8.Public()},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: values: key #1: option 'key_id' must be provided",
			},
		},
		{
			"ShouldFailOnNonKey",
			nil,
			[]schema.JWK{
				{Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: "nokey", KeyID: "KeyID"},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: values: key #1 with key id 'KeyID': option 'key' failed to get key properties: the key type 'string' is unknown or not valid for the configuration",
			},
		},
		{
			"ShouldFailOnBadUseAlg",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: "enc", Algorithm: "bad", Key: keyRSA2048PKCS8.Public()},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: values: key #1 with key id 'test': option 'use' must be one of 'sig' but it's configured as 'enc'",
				"identity_providers: oidc: clients: client 'test': public_keys: values: key #1 with key id 'test': option 'algorithm' must be one of 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', or 'ES512' but it's configured as 'bad'",
			},
		},
		{
			"ShouldFailOnEmptyKey",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: nil},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: values: key #1: option 'key' must be provided",
			},
		},
		{
			"ShouldFailOnMalformedKey",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: frankenkey.Public()},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: values: key #1: option 'key' option 'key' must be a valid private key but the provided data is malformed as it's missing the public key bits",
			},
		},
		{
			"ShouldFailOnBadKeySize",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA1024.Public()},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: values: key #1 with key id 'test': option 'key' is an RSA 1024 bit private key but it must at minimum be a RSA 2048 bit private key",
			},
		},
		{
			"ShouldFailOnMismatchedKeyCert",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048PKCS8.Public(), CertificateChain: certRSA2048},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: values: key #1 with key id 'test': option 'certificate_chain' does not appear to contain the public key for the public key provided by option 'key'",
			},
		},
		{
			"ShouldFailOnMismatchedCertChain",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048.Public(), CertificateChain: frankenchain},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': public_keys: values: key #1 with key id 'test': option 'certificate_chain' produced an error during validation of the chain: certificate #1 in chain is not signed properly by certificate #2 in chain: x509: invalid signature: parent certificate cannot sign this kind of certificate",
			},
		},
		{
			"ShouldSetDefaultUseAlgRSA",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: "", Algorithm: "", Key: keyRSA2048PKCS8.Public()},
			},
			nil,
			func(t *testing.T, config *schema.OpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].PublicKeys.Values[0].Use)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, config.Clients[0].PublicKeys.Values[0].Algorithm)
			},
			nil,
		},
		{
			"ShouldSetDefaultUseAlgECDSA256",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: "", Algorithm: "", Key: keyECDSAP256.Public()},
			},
			nil,
			func(t *testing.T, config *schema.OpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].PublicKeys.Values[0].Use)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP256AndSHA256, config.Clients[0].PublicKeys.Values[0].Algorithm)
			},
			nil,
		},
		{
			"ShouldSetDefaultUseAlgECDSA384",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: "", Algorithm: "", Key: keyECDSAP384.Public()},
			},
			nil,
			func(t *testing.T, config *schema.OpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].PublicKeys.Values[0].Use)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP384AndSHA384, config.Clients[0].PublicKeys.Values[0].Algorithm)
			},
			nil,
		},
		{
			"ShouldSetDefaultUseAlgECDSA521",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: "", Algorithm: "", Key: keyECDSAP521.Public()},
			},
			nil,
			func(t *testing.T, config *schema.OpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].PublicKeys.Values[0].Use)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, config.Clients[0].PublicKeys.Values[0].Algorithm)
			},
			nil,
		},
		{
			"ShouldConfigureRegisteredRequestObjectAlgs",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: "", Algorithm: "", Key: keyECDSAP521.Public()},
			},
			nil,
			func(t *testing.T, config *schema.OpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].PublicKeys.Values[0].Use)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, config.Clients[0].PublicKeys.Values[0].Algorithm)

				assert.Equal(t, []string{oidc.SigningAlgECDSAUsingP521AndSHA512}, config.Discovery.RequestObjectSigningAlgs)
			},
			nil,
		},
		{
			"ShouldOnlyAllowRequetsObjectSigningAlgsThatTheClientHasKeysFor",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: "", Algorithm: "", Key: keyECDSAP521.Public()},
			},
			func(config *schema.OpenIDConnect) {
				config.Clients[0].RequestObjectSigningAlg = oidc.SigningAlgRSAUsingSHA512
			},
			func(t *testing.T, config *schema.OpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].PublicKeys.Values[0].Use)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, config.Clients[0].PublicKeys.Values[0].Algorithm)

				assert.Equal(t, []string{oidc.SigningAlgECDSAUsingP521AndSHA512}, config.Discovery.RequestObjectSigningAlgs)
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'request_object_signing_alg' must be one of 'ES512' configured in the client option 'public_keys'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.OpenIDConnect{
				Clients: []schema.OpenIDConnectClient{
					{
						ID: "test",
						PublicKeys: schema.OpenIDConnectClientPublicKeys{
							URI:    tc.haveURI,
							Values: tc.haveJWKS,
						},
					},
				},
			}

			if tc.setup != nil {
				tc.setup(config)
			}

			val := schema.NewStructValidator()

			validateOIDCClientPublicKeys(0, config, val)

			if tc.expected != nil {
				tc.expected(t, config)
			}

			n := len(tc.errs)

			assert.Len(t, val.Warnings(), 0)

			theErrors := val.Errors()
			require.Len(t, theErrors, n)

			for i := 0; i < n; i++ {
				assert.EqualError(t, theErrors[i], tc.errs[i])
			}
		})
	}
}

func TestValidateOIDCIssuer(t *testing.T) {
	frankenchain := schema.NewX509CertificateChainFromCerts([]*x509.Certificate{certRSA2048.Leaf(), certRSA1024.Leaf()})
	frankenkey := &rsa.PrivateKey{}

	*frankenkey = *keyRSA2048

	frankenkey.PublicKey.N = nil

	testCases := []struct {
		name     string
		have     schema.OpenIDConnect
		expected schema.OpenIDConnect
		errs     []string
	}{
		{
			"ShouldMapLegacyConfiguration",
			schema.OpenIDConnect{
				IssuerPrivateKey: keyRSA2048,
			},
			schema.OpenIDConnect{
				IssuerPrivateKey: keyRSA2048,
				IssuerPrivateKeys: []schema.JWK{
					{KeyID: "1f8bfc", Key: keyRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "1f8bfc"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			nil,
		},
		{
			"ShouldSetDefaultKeyValues",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048},
					{Key: keyECDSAP256, CertificateChain: certECDSAP256},
					{Key: keyECDSAP384, CertificateChain: certECDSAP384},
					{Key: keyECDSAP521, CertificateChain: certECDSAP521},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "1f8bfc"},
					{Key: keyECDSAP256, CertificateChain: certECDSAP256, Algorithm: oidc.SigningAlgECDSAUsingP256AndSHA256, Use: oidc.KeyUseSignature, KeyID: "1e7788"},
					{Key: keyECDSAP384, CertificateChain: certECDSAP384, Algorithm: oidc.SigningAlgECDSAUsingP384AndSHA384, Use: oidc.KeyUseSignature, KeyID: "ba8508"},
					{Key: keyECDSAP521, CertificateChain: certECDSAP521, Algorithm: oidc.SigningAlgECDSAUsingP521AndSHA512, Use: oidc.KeyUseSignature, KeyID: "7ecbac"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs: map[string]string{
						oidc.SigningAlgRSAUsingSHA256:          "1f8bfc",
						oidc.SigningAlgECDSAUsingP256AndSHA256: "1e7788",
						oidc.SigningAlgECDSAUsingP384AndSHA384: "ba8508",
						oidc.SigningAlgECDSAUsingP521AndSHA512: "7ecbac",
					},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgECDSAUsingP384AndSHA384, oidc.SigningAlgECDSAUsingP521AndSHA512},
				},
			},
			nil,
		},
		{
			"ShouldNotRaiseErrorsMultipleRSA256Keys",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048},
					{Key: keyRSA4096, CertificateChain: certRSA4096},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "1f8bfc"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "bf1e10"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "1f8bfc"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			nil,
		},
		{
			"ShouldRaiseErrorsDuplicateRSA256Keys",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA512},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA512, Use: oidc.KeyUseSignature, KeyID: "bf1e10"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA512: "bf1e10"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA512},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: keys: must at least have one key supporting the 'RS256' algorithm but only has 'RS512'",
			},
		},
		{
			"ShouldRaiseErrorOnBadCurve",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096},
					{Key: keyECDSAP224, CertificateChain: certECDSAP224},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "bf1e10"},
					{Key: keyECDSAP224, CertificateChain: certECDSAP224},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "bf1e10"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #2: option 'key' failed to calculate thumbprint to configure key id value: square/go-jose: unsupported/unknown elliptic curve",
			},
		},
		{
			"ShouldRaiseErrorOnBadRSAKey",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA1024, CertificateChain: certRSA1024},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA1024, CertificateChain: certRSA1024, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "cf375e"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "cf375e"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #1 with key id 'cf375e': option 'key' is an RSA 1024 bit private key but it must at minimum be a RSA 2048 bit private key",
			},
		},
		{
			"ShouldRaiseErrorOnBadAlg",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: "invalid"},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: "invalid", Use: oidc.KeyUseSignature, KeyID: "bf1e10"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{},
					ResponseObjectSigningAlgs: []string{"invalid"},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #1 with key id 'bf1e10': option 'algorithm' must be one of 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', or 'ES512' but it's configured as 'invalid'",
				"identity_providers: oidc: issuer_private_keys: keys: must at least have one key supporting the 'RS256' algorithm but only has 'invalid'",
			},
		},
		{
			"ShouldRaiseErrorOnBadUse",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Use: "invalid"},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: "invalid", KeyID: "bf1e10"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "bf1e10"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #1 with key id 'bf1e10': option 'use' must be one of 'sig' but it's configured as 'invalid'",
			},
		},
		{
			"ShouldRaiseErrorOnBadKeyIDLength",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "thisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolong"},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "thisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolong"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "thisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolong"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #1 with key id 'thisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolong': option `key_id` must be 100 characters or less",
			},
		},
		{
			"ShouldRaiseErrorOnBadKeyIDCharacters",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x@x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "-xx"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "xx."},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x@x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "-xx"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "xx."},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "x@x"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #1 with key id 'x@x': option 'key_id' must only contain RFC3986 unreserved characters and must only start and end with alphanumeric characters",
				"identity_providers: oidc: issuer_private_keys: key #2 with key id '-xx': option 'key_id' must only contain RFC3986 unreserved characters and must only start and end with alphanumeric characters",
				"identity_providers: oidc: issuer_private_keys: key #3 with key id 'xx.': option 'key_id' must only contain RFC3986 unreserved characters and must only start and end with alphanumeric characters",
			},
		},
		{
			"ShouldNotRaiseErrorOnGoodKeyIDCharacters",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x-x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "xx"},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x-x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "xx"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "x-x"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnBadKeyIDDuplicates",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x"},
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAPSSUsingSHA256, KeyID: "x"},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x"},
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAPSSUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "x", oidc.SigningAlgRSAPSSUsingSHA256: "x"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgRSAPSSUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #2 with key id 'x': option 'key_id' must be unique",
			},
		},
		{
			"ShouldRaiseErrorOnEd25519Keys",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyEd2519, CertificateChain: certEd15519},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyEd2519, CertificateChain: certEd15519, KeyID: "14dfd3"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{},
					ResponseObjectSigningAlgs: []string(nil),
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #1 with key id '14dfd3': option 'key' must be a RSA private key or ECDSA private key but it's type is ed25519.PrivateKey",
			},
		},
		{
			"ShouldRaiseErrorOnCertificateAsKey",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: publicRSA2048Pair},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: publicRSA2048Pair, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "9a0e71"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "9a0e71"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #1 with key id '9a0e71': option 'key' must be a RSA private key or ECDSA private key but it's type is *rsa.PublicKey",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidChain",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: frankenchain},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: frankenchain, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "1f8bfc"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{oidc.SigningAlgRSAUsingSHA256: "1f8bfc"},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #1 with key id '1f8bfc': option 'certificate_chain' produced an error during validation of the chain: certificate #1 in chain is not signed properly by certificate #2 in chain: x509: invalid signature: parent certificate cannot sign this kind of certificate",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidPrivateKeyN",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: frankenkey},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: frankenkey},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{},
					ResponseObjectSigningAlgs: []string(nil),
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #1: option 'key' must be a valid private key but the provided data is malformed as it's missing the public key bits",
			},
		},
		{
			"ShouldRaiseErrorOnCertForKey",
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: certRSA2048},
				},
			},
			schema.OpenIDConnect{
				IssuerPrivateKeys: []schema.JWK{
					{Key: certRSA2048},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyIDs:             map[string]string{},
					ResponseObjectSigningAlgs: []string(nil),
				},
			},
			[]string{
				"identity_providers: oidc: issuer_private_keys: key #1 with key id '': option 'key' failed to get key properties: the key type 'schema.X509CertificateChain' is unknown or not valid for the configuration",
			},
		},
	}

	var n int

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val := schema.NewStructValidator()

			validateOIDCIssuer(&tc.have, val)

			assert.Equal(t, tc.expected.Discovery.DefaultKeyIDs, tc.have.Discovery.DefaultKeyIDs)
			assert.Equal(t, tc.expected.Discovery.ResponseObjectSigningAlgs, tc.have.Discovery.ResponseObjectSigningAlgs)
			assert.Equal(t, tc.expected.IssuerPrivateKey, tc.have.IssuerPrivateKey)
			assert.Equal(t, tc.expected.IssuerCertificateChain, tc.have.IssuerCertificateChain)

			n = len(tc.expected.IssuerPrivateKeys)

			require.Len(t, tc.have.IssuerPrivateKeys, n)

			for i := 0; i < n; i++ {
				t.Run(fmt.Sprintf("Key%d", i), func(t *testing.T) {
					assert.Equal(t, tc.expected.IssuerPrivateKeys[i].Algorithm, tc.have.IssuerPrivateKeys[i].Algorithm)
					assert.Equal(t, tc.expected.IssuerPrivateKeys[i].Use, tc.have.IssuerPrivateKeys[i].Use)
					assert.Equal(t, tc.expected.IssuerPrivateKeys[i].KeyID, tc.have.IssuerPrivateKeys[i].KeyID)
					assert.Equal(t, tc.expected.IssuerPrivateKeys[i].Key, tc.have.IssuerPrivateKeys[i].Key)
					assert.Equal(t, tc.expected.IssuerPrivateKeys[i].CertificateChain, tc.have.IssuerPrivateKeys[i].CertificateChain)
				})
			}

			n = len(tc.errs)

			require.Len(t, val.Errors(), n)

			for i := 0; i < n; i++ {
				assert.EqualError(t, val.Errors()[i], tc.errs[i])
			}
		})
	}
}

func MustDecodeSecret(value string) *schema.PasswordDigest {
	if secret, err := schema.DecodePasswordDigest(value); err != nil {
		panic(err)
	} else {
		return secret
	}
}

func MustLoadCrypto(alg, mod, ext string, extra ...string) any {
	fparts := []string{alg, mod}
	if len(extra) != 0 {
		fparts = append(fparts, extra...)
	}

	var (
		data    []byte
		decoded any
		err     error
	)

	if data, err = os.ReadFile(fmt.Sprintf(pathCrypto, strings.Join(fparts, "_"), ext)); err != nil {
		panic(err)
	}

	if decoded, err = utils.ParseX509FromPEMRecursive(data); err != nil {
		panic(err)
	}

	return decoded
}

func MustLoadCertificateChain(alg, op string) schema.X509CertificateChain {
	decoded := MustLoadCrypto(alg, op, "crt")

	switch cert := decoded.(type) {
	case *x509.Certificate:
		return schema.NewX509CertificateChainFromCerts([]*x509.Certificate{cert})
	case []*x509.Certificate:
		return schema.NewX509CertificateChainFromCerts(cert)
	default:
		panic(fmt.Errorf("the key was not a *x509.Certificate or []*x509.Certificate, it's a %T", cert))
	}
}

func MustLoadEd15519PrivateKey(curve string, extra ...string) ed25519.PrivateKey {
	decoded := MustLoadCrypto("ED25519", curve, "pem", extra...)

	key, ok := decoded.(ed25519.PrivateKey)
	if !ok {
		panic(fmt.Errorf("the key was not a ed25519.PrivateKey, it's a %T", key))
	}

	return key
}

func MustLoadECDSAPrivateKey(curve string, extra ...string) *ecdsa.PrivateKey {
	decoded := MustLoadCrypto("ECDSA", curve, "pem", extra...)

	key, ok := decoded.(*ecdsa.PrivateKey)
	if !ok {
		panic(fmt.Errorf("the key was not a *ecdsa.PrivateKey, it's a %T", key))
	}

	return key
}

func MustLoadRSAPublicKey(bits string, extra ...string) *rsa.PublicKey {
	decoded := MustLoadCrypto("RSA", bits, "pem", extra...)

	key, ok := decoded.(*rsa.PublicKey)
	if !ok {
		panic(fmt.Errorf("the key was not a *rsa.PublicKey, it's a %T", key))
	}

	return key
}

func MustLoadRSAPrivateKey(bits string, extra ...string) *rsa.PrivateKey {
	decoded := MustLoadCrypto("RSA", bits, "pem", extra...)

	key, ok := decoded.(*rsa.PrivateKey)
	if !ok {
		panic(fmt.Errorf("the key was not a *rsa.PrivateKey, it's a %T", key))
	}

	return key
}

const (
	pathCrypto = "../test_resources/crypto/%s.%s"
)

var (
	tOpenIDConnectPBKDF2ClientSecret, tOpenIDConnectPlainTextClientSecret *schema.PasswordDigest

	// Standard RSA key pair.
	publicRSA2048Pair  *rsa.PublicKey
	privateRSA2048Pair *rsa.PrivateKey

	// Standard RSA key / certificate pairs.
	keyRSA1024, keyRSA2048, keyRSA2048PKCS8, keyRSA4096 *rsa.PrivateKey
	certRSA1024, certRSA2048, certRSA4096               schema.X509CertificateChain

	// Standard ECDSA key / certificate pairs.
	keyECDSAP224, keyECDSAP256, keyECDSAP384, keyECDSAP521     *ecdsa.PrivateKey
	certECDSAP224, certECDSAP256, certECDSAP384, certECDSAP521 schema.X509CertificateChain

	// Standard ECDSA key / certificate pairs.
	keyECDSAP256PKCS8, keyECDSAP384PKCS8, keyECDSAP521PKCS8                        *ecdsa.PrivateKey
	certECDSAP224PKCS8, certECDSAP256PKCS8, certECDSAP384PKCS8, certECDSAP521PKCS8 schema.X509CertificateChain //nolint:unused

	// Ed15519 key / certificate pair.
	keyEd2519   ed25519.PrivateKey
	certEd15519 schema.X509CertificateChain
)

func init() {
	tOpenIDConnectPBKDF2ClientSecret = MustDecodeSecret("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng")
	tOpenIDConnectPlainTextClientSecret = MustDecodeSecret("$plaintext$example")

	publicRSA2048Pair = MustLoadRSAPublicKey("2048", "PAIR", "PUBLIC")
	privateRSA2048Pair = MustLoadRSAPrivateKey("2048", "PAIR", "PRIVATE")

	keyRSA1024 = MustLoadRSAPrivateKey("1024")
	keyRSA2048 = MustLoadRSAPrivateKey("2048")
	keyRSA4096 = MustLoadRSAPrivateKey("4096")
	keyECDSAP224 = MustLoadECDSAPrivateKey("P224")
	keyECDSAP256 = MustLoadECDSAPrivateKey("P256")
	keyECDSAP384 = MustLoadECDSAPrivateKey("P384")
	keyECDSAP521 = MustLoadECDSAPrivateKey("P521")
	keyEd2519 = MustLoadEd15519PrivateKey("PKCS8")

	keyRSA2048PKCS8 = MustLoadRSAPrivateKey("2048", "PKCS8")
	keyECDSAP256PKCS8 = MustLoadECDSAPrivateKey("P256", "PKCS8")
	keyECDSAP384PKCS8 = MustLoadECDSAPrivateKey("P384", "PKCS8")
	keyECDSAP521PKCS8 = MustLoadECDSAPrivateKey("P521", "PKCS8")

	certRSA1024 = MustLoadCertificateChain("RSA", "1024")
	certRSA2048 = MustLoadCertificateChain("RSA", "2048")
	certRSA4096 = MustLoadCertificateChain("RSA", "4096")
	certECDSAP224 = MustLoadCertificateChain("ECDSA", "P224")
	certECDSAP256 = MustLoadCertificateChain("ECDSA", "P256")
	certECDSAP384 = MustLoadCertificateChain("ECDSA", "P384")
	certECDSAP521 = MustLoadCertificateChain("ECDSA", "P521")
	certEd15519 = MustLoadCertificateChain("ED25519", "PKCS8")
}
