package validator

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
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
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret: "abc",
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'issuer_private_key' or `issuer_jwks` is required")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: option 'clients' must have one or more clients configured")
}

func TestShouldNotRaiseErrorWhenCORSEndpointsValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
			CORS: schema.OpenIDConnectCORSConfiguration{
				Endpoints: []string{oidc.EndpointAuthorization, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo},
			},
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "example",
					Secret: MustDecodeSecret("$plaintext$example"),
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Errors(), 0)
}

func TestShouldRaiseErrorWhenCORSEndpointsNotValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
			CORS: schema.OpenIDConnectCORSConfiguration{
				Endpoints: []string{oidc.EndpointAuthorization, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo, "invalid_endpoint"},
			},
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "example",
					Secret: MustDecodeSecret("$plaintext$example"),
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
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
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

	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
			CORS: schema.OpenIDConnectCORSConfiguration{
				AllowedOrigins:                       utils.URLsFromStringSlice([]string{"https://example.com/", "https://site.example.com/subpath", "https://site.example.com?example=true", "*"}),
				AllowedOriginsFromClientRedirectURIs: true,
			},
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:           "myclient",
					Secret:       MustDecodeSecret("$plaintext$jk12nb3klqwmnelqkwenm"),
					Policy:       "two_factor",
					RedirectURIs: []string{"https://example.com/oauth2_callback", "https://localhost:566/callback", "http://an.example.com/callback", "file://a/file"},
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
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
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
		Clients []schema.OpenIDConnectClientConfiguration
		Errors  []string
	}{
		{
			Name: "EmptyIDAndSecret",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:           "",
					Secret:       nil,
					Policy:       "",
					RedirectURIs: []string{},
				},
			},
			Errors: []string{
				"identity_providers: oidc: client '': option 'secret' is required",
				"identity_providers: oidc: clients: option 'id' is required but was absent on the clients in positions #1",
			},
		},
		{
			Name: "InvalidPolicy",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-1",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: "a-policy",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
			},
			Errors: []string{
				"identity_providers: oidc: client 'client-1': option 'policy' must be one of 'one_factor' or 'two_factor' but it's configured as 'a-policy'",
			},
		},
		{
			Name: "ClientIDDuplicated",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:           "client-x",
					Secret:       MustDecodeSecret("$plaintext$a-secret"),
					Policy:       policyTwoFactor,
					RedirectURIs: []string{},
				},
				{
					ID:           "client-x",
					Secret:       MustDecodeSecret("$plaintext$a-secret"),
					Policy:       policyTwoFactor,
					RedirectURIs: []string{},
				},
			},
			Errors: []string{
				"identity_providers: oidc: clients: option 'id' must be unique for every client but one or more clients share the following 'id' values 'client-x'",
			},
		},
		{
			Name: "RedirectURIInvalid",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-check-uri-parse",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"http://abc@%two",
					},
				},
			},
			Errors: []string{
				"identity_providers: oidc: client 'client-check-uri-parse': option 'redirect_uris' has an invalid value: redirect uri 'http://abc@%two' could not be parsed: parse \"http://abc@%two\": invalid URL escape \"%tw\"",
			},
		},
		{
			Name: "RedirectURINotAbsolute",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-check-uri-abs",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"google.com",
					},
				},
			},
			Errors: []string{
				"identity_providers: oidc: client 'client-check-uri-abs': option 'redirect_uris' has an invalid value: redirect uri 'google.com' must have a scheme but it's absent",
			},
		},
		{
			Name: "ValidSectorIdentifier",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-valid-sector",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifier: mustParseURL(exampleDotCom),
				},
			},
		},
		{
			Name: "ValidSectorIdentifierWithPort",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-valid-sector",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifier: mustParseURL("example.com:2000"),
				},
			},
		},
		{
			Name: "InvalidSectorIdentifierInvalidURL",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-invalid-sector",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifier: mustParseURL("https://user:pass@example.com/path?query=abc#fragment"),
				},
			},
			Errors: []string{
				"identity_providers: oidc: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a scheme with the value 'https'",
				"identity_providers: oidc: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a path with the value '/path'",
				"identity_providers: oidc: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a query with the value 'query=abc'",
				"identity_providers: oidc: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a fragment with the value 'fragment'",
				"identity_providers: oidc: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a username with the value 'user'",
				"identity_providers: oidc: client 'client-invalid-sector': option 'sector_identifier' with value 'https://user:pass@example.com/path?query=abc#fragment': must be a URL with only the host component for example 'example.com' but it has a password",
			},
		},
		{
			Name: "InvalidSectorIdentifierInvalidHost",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-invalid-sector",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifier: mustParseURL("example.com/path?query=abc#fragment"),
				},
			},
			Errors: []string{
				"identity_providers: oidc: client 'client-invalid-sector': option 'sector_identifier' with value 'example.com/path?query=abc#fragment': must be a URL with only the host component but appears to be invalid",
			},
		},
		{
			Name: "InvalidConsentMode",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-bad-consent-mode",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					ConsentMode: "cap",
				},
			},
			Errors: []string{
				"identity_providers: oidc: client 'client-bad-consent-mode': consent: option 'mode' must be one of 'auto', 'implicit', 'explicit', 'pre-configured', or 'auto' but it's configured as 'cap'",
			},
		},
		{
			Name: "InvalidPKCEChallengeMethod",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-bad-pkce-mode",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					PKCEChallengeMethod: "abc",
				},
			},
			Errors: []string{
				"identity_providers: oidc: client 'client-bad-pkce-mode': option 'pkce_challenge_method' must be one of 'plain' or 'S256' but it's configured as 'abc'",
			},
		},
		{
			Name: "InvalidPKCEChallengeMethodLowerCaseS256",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-bad-pkce-mode-s256",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					PKCEChallengeMethod: "s256",
				},
			},
			Errors: []string{
				"identity_providers: oidc: client 'client-bad-pkce-mode-s256': option 'pkce_challenge_method' must be one of 'plain' or 'S256' but it's configured as 's256'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.IdentityProvidersConfiguration{
				OIDC: &schema.OpenIDConnectConfiguration{
					HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
					IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
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
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "good_id",
					Secret: MustDecodeSecret("$plaintext$good_secret"),
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'scopes' must only have the values 'openid', 'email', 'profile', 'groups', or 'offline_access' but the values 'bad_scope' are present")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadGrantTypes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:         "good_id",
					Secret:     MustDecodeSecret(goodOpenIDConnectClientSecret),
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'grant_types' must only have the values 'implicit', 'refresh_token', or 'authorization_code' but the values 'bad_grant_type' are present")
}

func TestShouldNotErrorOnCertificateValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:             "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerCertificateChain: MustParseX509CertificateChain(testCertRSA1),
			IssuerPrivateKey:       MustParseRSAPrivateKey(testKeyRSA1),
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "good_id",
					Secret: MustDecodeSecret(goodOpenIDConnectClientSecret),
					Policy: "two_factor",
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
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:             "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerCertificateChain: MustParseX509CertificateChain(testCertRSA1),
			IssuerPrivateKey:       MustParseRSAPrivateKey(testKeyRSA2),
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "good_id",
					Secret: MustDecodeSecret(goodOpenIDConnectClientSecret),
					Policy: "two_factor",
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

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: issuer_jwks: key #0 with key id '5ee437': option 'key' does not appear to be the private key the certificate provided by option 'certificate_chain'")
}

func TestShouldRaiseErrorOnKeySizeTooSmall(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA3),
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "good_id",
					Secret: MustDecodeSecret(goodOpenIDConnectClientSecret),
					Policy: "two_factor",
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

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: issuer_jwks: key #0 with key id '739c2e': option 'key' is an RSA 1024 bit private key but it must be a RSA 2048 bit private key")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadResponseModes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:            "good_id",
					Secret:        MustDecodeSecret("$plaintext$good_secret"),
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'response_modes' must only have the values 'form_post', 'query', or 'fragment' but the values 'bad_responsemode' are present")
}

func TestValidateIdentityProvidersShouldRaiseWarningOnSecurityIssue(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:              "abc",
			IssuerPrivateKey:        MustParseRSAPrivateKey(testKeyRSA1),
			MinimumParameterEntropy: 1,
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "good_id",
					Secret: MustDecodeSecret(goodOpenIDConnectClientSecret),
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
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-with-invalid-secret",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://localhost",
					},
				},
				{
					ID:     "client-with-bad-redirect-uri",
					Secret: MustDecodeSecret(goodOpenIDConnectClientSecret),
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

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'client-with-invalid-secret': option 'secret' is required to be empty when option 'public' is true")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: client 'client-with-bad-redirect-uri': option 'redirect_uris' has the redirect uri 'urn:ietf:wg:oauth:2.0:oob' when option 'public' is false but this is invalid as this uri is not valid for the openid connect confidential client type")
}

func TestValidateIdentityProvidersShouldNotRaiseErrorsOnValidClientOptions(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "hmac1",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
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
				{
					ID:     "client-with-pkce-mode-plain",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://pkce.com",
					},
					PKCEChallengeMethod: "plain",
				},
				{
					ID:     "client-with-pkce-mode-S256",
					Public: true,
					Policy: "two_factor",
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
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "hmac1",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKeyRSA1),
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-with-invalid-secret_standard",
					Secret: MustDecodeSecret("$plaintext$a-secret"),
					Policy: "two_factor",
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

	assert.EqualError(t, validator.Warnings()[0], "identity_providers: oidc: client 'client-with-invalid-secret_standard': option 'secret' is plaintext but for clients not using the 'token_endpoint_auth_method' of 'client_secret_jwt' it should be a hashed value as plaintext values are deprecated with the exception of 'client_secret_jwt' and will be removed when oidc becomes stable")
}

// All valid schemes are supported as defined in https://datatracker.ietf.org/doc/html/rfc8252#section-7.1
func TestValidateOIDCClientRedirectURIsSupportingPrivateUseURISchemes(t *testing.T) {
	have := &schema.OpenIDConnectConfiguration{
		Clients: []schema.OpenIDConnectClientConfiguration{
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
			errors.New("identity_providers: oidc: client 'owncloud': option 'redirect_uris' has the redirect uri 'urn:ietf:wg:oauth:2.0:oob' when option 'public' is false but this is invalid as this uri is not valid for the openid connect confidential client type"),
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
		setup    func(have *schema.OpenIDConnectConfiguration)
		validate func(t *testing.T, have *schema.OpenIDConnectConfiguration)
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
				"identity_providers: oidc: client 'test': option 'scopes' must have unique values but the values 'openid' are duplicated",
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
				"identity_providers: oidc: client 'test': option 'scopes' must only have the values 'openid', 'email', 'profile', 'groups', or 'offline_access' but the values 'group' are present",
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
				"identity_providers: oidc: client 'test': option 'scopes' should only have the values 'offline_access' or 'offline' if the client is also configured with a 'response_type' such as 'code', 'code id_token', 'code token', or 'code id_token token' which respond with authorization codes",
				"identity_providers: oidc: client 'test': option 'grant_types' should only have the values 'refresh_token' if the client is also configured with a 'response_type' such as 'code', 'code id_token', 'code token', or 'code id_token token' which respond with authorization codes",
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
				"identity_providers: oidc: client 'test': option 'response_types' must have unique values but the values 'code' are duplicated",
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
				"identity_providers: oidc: client 'test': option 'response_types' must only have the values 'code', 'id_token', 'token', 'id_token token', 'code id_token', 'code token', or 'code id_token token' but the values 'token id_token' are present",
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
				"identity_providers: oidc: client 'test': option 'response_types' must only have the values 'code', 'id_token', 'token', 'id_token token', 'code id_token', 'code token', or 'code id_token token' but the values 'not_valid' are present",
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
				"identity_providers: oidc: client 'test': option 'response_modes' must only have the values 'form_post', 'query', or 'fragment' but the values 'not_valid' are present",
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
				"identity_providers: oidc: client 'test': option 'response_modes' must have unique values but the values 'query' are duplicated",
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
				"identity_providers: oidc: client 'test': option 'grant_types' must only have the values 'implicit', 'refresh_token', or 'authorization_code' but the values 'invalid' are present",
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
				"identity_providers: oidc: client 'test': option 'grant_types' must have unique values but the values 'authorization_code' are duplicated",
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
				"identity_providers: oidc: client 'test': option 'grant_types' should only have the 'refresh_token' value if the client is also configured with the 'offline_access' scope",
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
				"identity_providers: oidc: client 'test': option 'grant_types' should only have grant type values which are valid with the configured 'response_types' for the client but 'authorization_code' expects a response type for either the authorization code or hybrid flow such as 'code', 'code id_token', 'code token', or 'code id_token token' but the response types are 'id_token token'",
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
				"identity_providers: oidc: client 'test': option 'grant_types' should only have grant type values which are valid with the configured 'response_types' for the client but 'implicit' expects a response type for either the implicit or hybrid flow such as 'id_token', 'token', 'id_token token', 'code id_token', 'code token', or 'code id_token token' but the response types are 'code'",
			},
			nil,
		},
		{
			"ShouldValidateCorrectRedirectURIsConfidentialClientType",
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].RedirectURIs = []string{
					"https://google.com",
				}
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].Public = true
				have.Clients[0].Secret = nil
				have.Clients[0].RedirectURIs = []string{
					oauth2InstalledApp,
				}
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].RedirectURIs = []string{
					"urn:ietf:wg:oauth:2.0:oob",
				}
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'redirect_uris' has the redirect uri 'urn:ietf:wg:oauth:2.0:oob' when option 'public' is false but this is invalid as this uri is not valid for the openid connect confidential client type",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidRedirectURIsMalformedURI",
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].RedirectURIs = []string{
					"http://abc@%two",
				}
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'redirect_uris' has an invalid value: redirect uri 'http://abc@%two' could not be parsed: parse \"http://abc@%two\": invalid URL escape \"%tw\"",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidRedirectURIsNotAbsolute",
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].RedirectURIs = []string{
					"google.com",
				}
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'redirect_uris' has an invalid value: redirect uri 'google.com' must have a scheme but it's absent",
			},
		},
		{
			"ShouldRaiseErrorOnDuplicateRedirectURI",
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].RedirectURIs = []string{
					"https://google.com",
					"https://google.com",
				}
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'redirect_uris' must have unique values but the values 'https://google.com' are duplicated",
			},
			nil,
		},
		{
			"ShouldNotSetDefaultTokenEndpointClientAuthMethodConfidentialClientType",
			nil,
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretPost
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].TokenEndpointAuthMethod = "client_credentials"
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
				assert.Equal(t, "client_credentials", have.Clients[0].TokenEndpointAuthMethod)
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
				"identity_providers: oidc: client 'test': option 'token_endpoint_auth_method' must be one of 'none', 'client_secret_post', 'client_secret_basic', or 'client_secret_jwt' but it's configured as 'client_credentials'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidClientAuthMethodForPublicClientType",
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretBasic
				have.Clients[0].Public = true
				have.Clients[0].Secret = nil
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'token_endpoint_auth_method' must be 'none' when configured as the public client type but it's configured as 'client_secret_basic'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidClientAuthMethodForConfidentialClientTypeAuthorizationCodeFlow",
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodNone
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'token_endpoint_auth_method' must be one of 'client_secret_post' or 'client_secret_basic' when configured as the confidential client type unless it only includes implicit flow response types such as 'id_token', 'token', and 'id_token token' but it's configured as 'none'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidClientAuthMethodForConfidentialClientTypeHybridFlow",
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodNone
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'token_endpoint_auth_method' must be one of 'client_secret_post' or 'client_secret_basic' when configured as the confidential client type unless it only includes implicit flow response types such as 'id_token', 'token', and 'id_token token' but it's configured as 'none'",
			},
		},
		{
			"ShouldSetDefaultUserInfoAlg",
			nil,
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].UserinfoSigningAlg = oidc.SigningAlgRSAUsingSHA256
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			"ShouldRaiseErrorOnInvalidUserInfoAlg",
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].UserinfoSigningAlg = "rs256"
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
				assert.Equal(t, "rs256", have.Clients[0].UserinfoSigningAlg)
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
				"identity_providers: oidc: client 'test': option 'userinfo_signing_algorithm' must be one of 'RS256' or 'none' but it's configured as 'rs256'",
			},
		},
		{
			"ShouldSetDefaultConsentMode",
			nil,
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].ConsentMode = auto
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				d := time.Minute

				have.Clients[0].ConsentMode = ""
				have.Clients[0].ConsentPreConfiguredDuration = &d
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				d := time.Minute

				have.Clients[0].ConsentMode = auto
				have.Clients[0].ConsentPreConfiguredDuration = &d
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].ConsentMode = "implicit"
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].ConsentMode = "pre-configured"
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			"ShouldRaiseErrorOnIncorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'secret' must be plaintext with option 'token_endpoint_auth_method' with a value of 'client_secret_jwt'",
			},
		},
		{
			"ShouldNotRaiseWarningOrErrorOnCorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.OpenIDConnectConfiguration) {
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
			"ShouldRaiseErrorOnIncorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'secret' must be plaintext with option 'token_endpoint_auth_method' with a value of 'client_secret_jwt'",
			},
		},
		{
			"ShouldNotRaiseWarningOrErrorOnCorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.OpenIDConnectConfiguration) {
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
			"ShouldSetDefaultTokenEndpointAuthSigAlg",
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].Secret = MustDecodeSecret("$plaintext$abc123")
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = oidc.SigningAlgHMACUsingSHA256
				have.Clients[0].Secret = nil
				have.Clients[0].Public = true
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'token_endpoint_auth_method' must be 'none' when configured as the public client type but it's configured as 'client_secret_jwt'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidTokenAuthAlgClientTypeConfidential",
			func(have *schema.OpenIDConnectConfiguration) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = oidc.EndpointToken
				have.Clients[0].Secret = MustDecodeSecret("$plaintext$abc123")
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
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
				"identity_providers: oidc: client 'test': option 'token_endpoint_auth_signing_alg' must be 'HS256', 'HS384', or 'HS512' when option 'token_endpoint_auth_method' is client_secret_jwt",
			},
		},
	}

	errDeprecatedFunc := func() {}

	for _, tc := range testCasses {
		t.Run(tc.name, func(t *testing.T) {
			have := &schema.OpenIDConnectConfiguration{
				Discovery: schema.OpenIDConnectDiscovery{
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
				Clients: []schema.OpenIDConnectClientConfiguration{
					{
						ID:            "test",
						Secret:        MustDecodeSecret("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng"),
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
				"identity_providers: oidc: client 'test': option 'token_endpoint_auth_method' must be one of 'none', 'client_secret_post', 'client_secret_basic', or 'client_secret_jwt' but it's configured as 'abc'",
			},
		},
		{"ShouldErrorOnInvalidValueForPublicClient", "client_secret_post", true, "client_secret_post",
			[]string{
				"identity_providers: oidc: client 'test': option 'token_endpoint_auth_method' must be 'none' when configured as the public client type but it's configured as 'client_secret_post'",
			},
		},
		{"ShouldErrorOnInvalidValueForConfidentialClient", "none", false, "none",
			[]string{
				"identity_providers: oidc: client 'test': option 'token_endpoint_auth_method' must be one of 'client_secret_post' or 'client_secret_basic' when configured as the confidential client type unless it only includes implicit flow response types such as 'id_token', 'token', and 'id_token token' but it's configured as 'none'",
			},
		},
	}

	for _, tc := range testCasses {
		t.Run(tc.name, func(t *testing.T) {
			have := &schema.OpenIDConnectConfiguration{
				Clients: []schema.OpenIDConnectClientConfiguration{
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

func TestValidateOIDCIssuer(t *testing.T) {
	keyRSA1024 := MustParseRSAPrivateKey(testKeyRSA1024)
	keyRSA2048 := MustParseRSAPrivateKey(testKeyRSA2048)
	keyRSA4096 := MustParseRSAPrivateKey(testKeyRSA4096)
	keyECDSAP224 := MustParseECDSAPrivateKey(testKeyECDSAWithP224)
	keyECDSAP256 := MustParseECDSAPrivateKey(testKeyECDSAWithP256)
	keyECDSAP384 := MustParseECDSAPrivateKey(testKeyECDSAWithP384)
	keyECDSAP521 := MustParseECDSAPrivateKey(testKeyECDSAWithP521)

	assert.NotNil(t, keyECDSAP224)

	certRSA1024 := MustParseX509CertificateChain(testCertRSA1024)
	certRSA2048 := MustParseX509CertificateChain(testCertRSA2048)
	certRSA4096 := MustParseX509CertificateChain(testCertRSA4096)
	certECDSAP224 := MustParseX509CertificateChain(testCertECDSAWithP224)
	certECDSAP256 := MustParseX509CertificateChain(testCertECDSAWithP256)
	certECDSAP384 := MustParseX509CertificateChain(testCertECDSAWithP384)
	certECDSAP521 := MustParseX509CertificateChain(testCertECDSAWithP521)

	assert.NotNil(t, certECDSAP224)

	testCases := []struct {
		name     string
		have     schema.OpenIDConnectConfiguration
		expected schema.OpenIDConnectConfiguration
		errs     []string
	}{
		{
			"ShouldMapLegacyConfiguration",
			schema.OpenIDConnectConfiguration{
				IssuerPrivateKey: keyRSA2048,
			},
			schema.OpenIDConnectConfiguration{
				IssuerPrivateKey: keyRSA2048,
				IssuerJWKS: []schema.JWK{
					{KeyID: "e7dfdc", Key: keyRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "e7dfdc",
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			nil,
		},
		{
			"ShouldSetDefaultKeyValues",
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048},
					{Key: keyECDSAP256, CertificateChain: certECDSAP256},
					{Key: keyECDSAP384, CertificateChain: certECDSAP384},
					{Key: keyECDSAP521, CertificateChain: certECDSAP521},
				},
			},
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "e7dfdc"},
					{Key: keyECDSAP256, CertificateChain: certECDSAP256, Algorithm: oidc.SigningAlgECDSAUsingP256AndSHA256, Use: oidc.KeyUseSignature, KeyID: "29b3f2"},
					{Key: keyECDSAP384, CertificateChain: certECDSAP384, Algorithm: oidc.SigningAlgECDSAUsingP384AndSHA384, Use: oidc.KeyUseSignature, KeyID: "e968b4"},
					{Key: keyECDSAP521, CertificateChain: certECDSAP521, Algorithm: oidc.SigningAlgECDSAUsingP521AndSHA512, Use: oidc.KeyUseSignature, KeyID: "6b20c3"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "e7dfdc",
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgECDSAUsingP384AndSHA384, oidc.SigningAlgECDSAUsingP521AndSHA512},
				},
			},
			nil,
		},
		{
			"ShouldRaiseErrorsDuplicateRSA256Keys",
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048},
					{Key: keyRSA4096, CertificateChain: certRSA4096},
				},
			},
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "e7dfdc"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "9c7423"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "e7dfdc",
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_jwks: key #1 with key id '9c7423': option 'algorithm' must be unique but another key is using it",
			},
		},
		{
			"ShouldRaiseErrorsDuplicateRSA256Keys",
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA512},
				},
			},
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA512, Use: oidc.KeyUseSignature, KeyID: "9c7423"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "",
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA512},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_jwks: keys: must at least have one key supporting the 'RS256' algorithm but only has 'RS512'",
			},
		},
		{
			"ShouldRaiseErrorOnBadCurve",
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096},
					{Key: keyECDSAP224, CertificateChain: certECDSAP224},
				},
			},
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "9c7423"},
					{Key: keyECDSAP224, CertificateChain: certECDSAP224},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "9c7423",
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_jwks: key #1: option 'key' failed to calculate thumbprint to configure key id value: square/go-jose: unsupported/unknown elliptic curve",
			},
		},
		{
			"ShouldRaiseErrorOnBadRSAKey",
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA1024, CertificateChain: certRSA1024},
				},
			},
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA1024, CertificateChain: certRSA1024, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "a9c018"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "a9c018",
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_jwks: key #0 with key id 'a9c018': option 'key' is an RSA 1024 bit private key but it must be a RSA 2048 bit private key",
			},
		},
		{
			"ShouldRaiseErrorOnBadAlg",
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: "invalid"},
				},
			},
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: "invalid", Use: oidc.KeyUseSignature, KeyID: "9c7423"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "",
					RegisteredJWKSigningAlgs: []string{"invalid"},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_jwks: key #0 with key id '9c7423': option 'algorithm' must be one of 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', or 'ES512' but it's configured as 'invalid'",
				"identity_providers: oidc: issuer_jwks: keys: must at least have one key supporting the 'RS256' algorithm but only has 'invalid'",
			},
		},
		{
			"ShouldRaiseErrorOnBadUse",
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Use: "invalid"},
				},
			},
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: "invalid", KeyID: "9c7423"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "9c7423",
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_jwks: key #0 with key id '9c7423': option 'use' must be one of 'sig' but it's configured as 'invalid'",
			},
		},
		{
			"ShouldRaiseErrorOnBadKeyIDLength",
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "thisistoolong"},
				},
			},
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "thisistoolong"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "thisistoolong",
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_jwks: key #0 with key id 'thisistoolong': option `key_id`` must be 7 characters or less",
			},
		},
		{
			"ShouldRaiseErrorOnBadKeyIDCharacters",
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x@x"},
				},
			},
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x@x"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "x@x",
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_jwks: key #0 with key id 'x@x': option 'key_id' must only have alphanumeric characters",
			},
		},
		{
			"ShouldRaiseErrorOnBadKeyIDDuplicates",
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x"},
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAPSSUsingSHA256, KeyID: "x"},
				},
			},
			schema.OpenIDConnectConfiguration{
				IssuerJWKS: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x"},
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAPSSUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x"},
				},
				Discovery: schema.OpenIDConnectDiscovery{
					DefaultKeyID:             "x",
					RegisteredJWKSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgRSAPSSUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: issuer_jwks: key #1 with key id 'x': option 'key_id' must be unique",
			},
		},
	}

	var n int

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val := schema.NewStructValidator()

			validateOIDCIssuer(&tc.have, val)

			assert.Equal(t, tc.expected.Discovery.DefaultKeyID, tc.have.Discovery.DefaultKeyID)
			assert.Equal(t, tc.expected.Discovery.RegisteredJWKSigningAlgs, tc.have.Discovery.RegisteredJWKSigningAlgs)
			assert.Equal(t, tc.expected.IssuerPrivateKey, tc.have.IssuerPrivateKey)
			assert.Equal(t, tc.expected.IssuerCertificateChain, tc.have.IssuerCertificateChain)

			n = len(tc.expected.IssuerJWKS)

			require.Len(t, tc.have.IssuerJWKS, n)

			for i := 0; i < n; i++ {
				t.Run(fmt.Sprintf("Key%d", i), func(t *testing.T) {
					assert.Equal(t, tc.expected.IssuerJWKS[i].Algorithm, tc.have.IssuerJWKS[i].Algorithm)
					assert.Equal(t, tc.expected.IssuerJWKS[i].Use, tc.have.IssuerJWKS[i].Use)
					assert.Equal(t, tc.expected.IssuerJWKS[i].KeyID, tc.have.IssuerJWKS[i].KeyID)
					assert.Equal(t, tc.expected.IssuerJWKS[i].Key, tc.have.IssuerJWKS[i].Key)
					assert.Equal(t, tc.expected.IssuerJWKS[i].CertificateChain, tc.have.IssuerJWKS[i].CertificateChain)
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

func MustParseRSAPrivateKey(data string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(data))
	if block == nil || len(block.Bytes) == 0 {
		panic("not pem encoded")
	}

	if block.Type != "RSA PRIVATE KEY" {
		panic("not rsa private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return key
}

func MustParseECDSAPrivateKey(data string) *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(strings.TrimSpace(data)))
	if block == nil || len(block.Bytes) == 0 {
		panic("not pem encoded")
	}

	if block.Type != "EC PRIVATE KEY" {
		panic("not ecdsa private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return key
}

func MustParseX509CertificateChain(data string) schema.X509CertificateChain {
	chain, err := schema.NewX509CertificateChain(data)

	if err != nil {
		panic(err)
	}

	return *chain
}

var (
	testCertRSA1 = `
-----BEGIN CERTIFICATE-----
MIIC5jCCAc6gAwIBAgIRAJZ+6KrHw95zIDgm2arCTCgwDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwHhcNMjIwOTA4MDIyNDQyWhcNMjMwOTA4MDIy
NDQyWjATMREwDwYDVQQKEwhBdXRoZWxpYTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAMAE7muDAJtLsV3WgOpjrZ1JD1RlhuSOa3V+4zo2NYFQSdZW18SZ
fYYgUwLOleEy3VQ3N9MEFh/rWNHYHdsBjDvz/Q1EzAlXqthGd0Sic/UDYtrahrko
jCSkZCQ5YVO9ivMRth6XdUlu7RHVYY3aSOWPx2wiw9cdN+e4p73W6KwyzT7ezbUD
0Nng0Z7CNQTLHv3LBsLUODc4aVOvp2B4aAaw6cn990buKMvUuo2ge9gh0c5gIOM5
dU7xOGAt7RzwCIHnG4CGAWPFuuS215ZeelgQr/9/fhtzDqSuBZw5f10vXnAyBwei
vN6Kffj2RXB+koFwBguT84A6cfmxWllGNF0CAwEAAaM1MDMwDgYDVR0PAQH/BAQD
AgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwDQYJKoZIhvcN
AQELBQADggEBAFvORjj7RGoIc3q0fv6QjuncZ0Mu1/24O0smCr6tq5d6RQBRpb1M
jEsbTMLZErrHbyw/DWC75eJhW6T+6HiVTo6brBXkmDL+QGkLgRNOkZla6cnmIpmL
bf9iPmmcThscQERgYZzNg19zqK8JAQU/6PgU/N6OXTL/mQQoB972ET9dUl7lGx1Q
2l8XBe8t4QTp4t1xd3c4azxWvFNpzWBjC5eBWiVHLJmFXr4xpcnPFYFETOkvEqwt
pMQ2x895BoLrep6b+g0xeF4pmmIQwA9KrUVr++gpYaRzytaOIYwcIPMzt9iLWKQe
6ZSOrTVi8pPugYXp+LhVk/WI7r8EWtyADu0=
-----END CERTIFICATE-----`

	testKeyRSA1 = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAwATua4MAm0uxXdaA6mOtnUkPVGWG5I5rdX7jOjY1gVBJ1lbX
xJl9hiBTAs6V4TLdVDc30wQWH+tY0dgd2wGMO/P9DUTMCVeq2EZ3RKJz9QNi2tqG
uSiMJKRkJDlhU72K8xG2Hpd1SW7tEdVhjdpI5Y/HbCLD1x0357invdborDLNPt7N
tQPQ2eDRnsI1BMse/csGwtQ4NzhpU6+nYHhoBrDpyf33Ru4oy9S6jaB72CHRzmAg
4zl1TvE4YC3tHPAIgecbgIYBY8W65LbXll56WBCv/39+G3MOpK4FnDl/XS9ecDIH
B6K83op9+PZFcH6SgXAGC5PzgDpx+bFaWUY0XQIDAQABAoIBAQClcdpHcglMxOwe
kRpkWdwWAAQgUJXoSbnW86wu1NRHBfmInyyrrSBVN3aunXbQITZIQIdt3kB94haW
P6KBt5Svd2saSqOOjSWb0SMkVOCaQ/+h19VqpcASNj4+Y94y+8ZD5ofHVfJtghDr
Y7H5OhHDEZ3e0xlwODGaCyUkUY4KBv/oIlILoh4phbDYHkZH8AzDnEiyVE1JAWlN
voAQysgSU7eEnNCi1S07jl5bY+MD3XpJkAfQsJYhqYT/qetStZ12PuXjpbIr3y53
qjCrKeWTyDN+gOznyIGuiR6nvXeQAw/o9hZiah4RuHXTPs/3GAcRXcuMR0pbgJ+B
yfX6eLK1AoGBAPKkJKPYJD2NHukAelNbT2OgeDIgJmfOIvLa73/x2hXvWwa4VwIC
POuKXtT/a02J4pYMGlaKXfHgLmaW2HPObOIjpxqgRIswsiKS1AbaDtkWnhaS1/SJ
oZ7Fk8DdX+1QT4J/fj/2uxRT0GhXdMxDpK7ekpmRE+APPCGhmOMgmWszAoGBAMqX
Ts1RdGWgDxLi15rDqdqRBARJG7Om/xC2voXVmbAb4Q+QoNrNeiIAM2usuhrVuj5V
c16m9fxswRNYqQBYyShDi5wp5a8UjfqDpzJdku2bmrBaL+XVq8PY+oTK6KS3ss8U
CGQ8P6Phz5JGavn/nDMRZ4EwEWqbEMUqJAJlpmIvAoGAQ9Wj8LJyn0qew6FAkaFL
dpzcPZdDZW35009l+a0RvWQnXJ+Yo5UglvEeRgoKY6kS0cQccOlKDl8QWdn+NZIW
WrqA8y6vOwKoKoZGBIxd7k8mb0UqXtFDf/HYtuis8tmrAN7H2vYNo0czUphwrNKU
bdcHwSsQFWns87IL3iO1AIUCgYBzmBX8jOePPN6c9hXzVoVKEshp8ZT+0uBilwLq
tk/07lNiYDGH5woy8E5mt62QtjaIbpVfgoCEwUEBWutDKWXNtYypVDabyWyhbhEu
abn2HX0L9smxqFNTcjCvKF/J7I74HQQUvVPKnIOlgMx1TOXBNcMLMXQekc/lz/+v
5nQjPQKBgQDjdJABeiy9tU0tzLWUVc5QoQKnlfSJoFLis46REb1yHwU9OjTc05Wx
5lAXdTjDmnelDdGWNWHjWOiKSkTxhvQD3jXriI5y8Sdxe3zS3ikYvbMbi97GJz0O
5oyNJo6/froW1dLkJJWR8hg2PQbtoOo6l9HHSd91BnJJ4qFbq9ZrXQ==
-----END RSA PRIVATE KEY-----`

	testKeyRSA2 = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA6z1LOg1ZCqb0lytXWZ+MRBpMHEXOoTOLYgfZXt1IYyE3Z758
cyalk0NYQhY5cZDsXPYWPvAHiPMUxutWkoxFwby56S+AbIMa3/Is+ILrHRJs8Exn
ZkpyrYFxPX12app2kErdmAkHSx0Z5/kuXiz96PHs8S8/ZbyZolLHzdfLtSzjvRm5
Zue5iFzsf19NJz5CIBfv8g5lRwtE8wNJoRSpn1xq7fqfuA0weDNFPzjlNWRLy6aa
rK7qJexRkmkCs4sLgyl+9NODYJpvmN8E1yhyC27E0joI6rBFVW7Ihv+cSPCdDzGp
EWe81x3AeqAa3mjVqkiq4u4Z2i8JDgBaPboqJwIDAQABAoIBAAFdLZ58jVOefDSU
L8F5R1rtvBs93GDa56f926jNJ6pLewLC+/2+757W+SAI+PRLntM7Kg3bXm/Q2QH+
Q1Y+MflZmspbWCdI61L5GIGoYKyeers59i+FpvySj5GHtLQRiTZ0+Kv1AXHSDWBm
9XneUOqU3IbZe0ifu1RRno72/VtjkGXbW8Mkkw+ohyGbIeTx/0/JQ6sSNZTT3Vk7
8i4IXptq3HSF0/vqZuah8rShoeNq72pD1YLM9YPdL5by1QkDLnqATDiCpLBTCaNV
I8sqYEun+HYbQzBj8ZACG2JVZpEEidONWQHw5BPWO95DSZYrVnEkuCqeH+u5vYt7
CHuJ3AECgYEA+W3v5z+j91w1VPHS0VB3SCDMouycAMIUnJPAbt+0LPP0scUFsBGE
hPAKddC54pmMZRQ2KIwBKiyWfCrJ8Xz8Yogn7fJgmwTHidJBr2WQpIEkNGlK3Dzi
jXL2sh0yC7sHvn0DqiQ79l/e7yRbSnv2wrTJEczOOH2haD7/tBRyCYECgYEA8W+q
E9YyGvEltnPFaOxofNZ8LHVcZSsQI5b6fc0iE7fjxFqeXPXEwGSOTwqQLQRiHn9b
CfPmIG4Vhyq0otVmlPvUnfBZ2OK+tl5X2/mQFO3ROMdvpi0KYa994uqfJdSTaqLn
jjoKFB906UFHnDQDLZUNiV1WwnkTglgLc+xrd6cCgYEAqqthyv6NyBTM3Tm2gcio
Ra9Dtntl51LlXZnvwy3IkDXBCd6BHM9vuLKyxZiziGx+Vy90O1xI872cnot8sINQ
Am+dur/tAEVN72zxyv0Y8qb2yfH96iKy9gxi5s75TnOEQgAygLnYWaWR2lorKRUX
bHTdXBOiS58S0UzCFEslGIECgYBqkO4SKWYeTDhoKvuEj2yjRYyzlu28XeCWxOo1
otiauX0YSyNBRt2cSgYiTzhKFng0m+QUJYp63/wymB/5C5Zmxi0XtWIDADpLhqLj
HmmBQ2Mo26alQ5YkffBju0mZyhVzaQop1eZi8WuKFV1FThPlB7hc3E0SM5zv2Grd
tQnOWwKBgQC40yZY0PcjuILhy+sIc0Wvh7LUA7taSdTye149kRvbvsCDN7Jh75lM
USjhLXY0Nld2zBm9r8wMb81mXH29uvD+tDqqsICvyuKlA/tyzXR+QTr7dCVKVwu0
1YjCJ36UpTsLre2f8nOSLtNmRfDPtbOE2mkOoO9dD9UU0XZwnvn9xw==
-----END RSA PRIVATE KEY-----`

	testKeyRSA3 = `
-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDBi7fdmUmlpWklpgAvNUdhDrpsDVqAHuEzVApK6f6ohYAi0/q2
+YmOwyPKDSrOc6Sy1myJtV3FbZGvYaQhnokc4bnkS9DH0lY+6Hk2vKps5PrhRY/q
1EjnfwXvzhAzb25rGFwKcSvfvndMTVvxgqXVob+3pRt9maD6HFHAh2/NCQIDAQAB
AoGACT2bfLgJ3R/FomeHkLlxe//RBMGqdX2D8QhtKWB8qR0engsS6FOHrspAVjBE
v/Cjh2pXake/f2KY1w/JX1WLZEFXja2RFPeeDiiC/4S7pKCySUVeHO9rQ4SY5Frg
/s/QWWtmq7+1iu2DXhdGJA6fIurzSoDgUXo3NGFCYqIFaAECQQDUi9AAgEljmc2q
dAUQD0KNTcJFkpTafhfPiYc2GT1vS/bArtXRmvJmbIiRfVuGbM8z5ES7JGd5FyYL
i2WCCzUBAkEA6R14GVhN8NIPWEUrzjgOvjKlc2ZHskT3dYb3djpm69TK7GjLtHyq
qO7l4VJowsXI+o/6YucagF6+rH0O0VrwCQJBAM8twYDbi63knA8MrGqtFUg7haTf
bu1Tf84y1nOrQrEcMNg9E/sOuD2SicSXlwF/SrHgTgbFQ39LSzBxnm6WkgECQQCh
AQmB98tdGLggbyXiODV2h+Rd37aFGb0QHzerIIsVNtMwlPCcp733D4kWJqTUYWZ+
KBL3XEahgs6Os5EYZ4aBAkEAjKE+2/nBYUdHVusjMXeNsE5rqwJND5zvYzmToG7+
xhv4RUAe4dHL4IDQoQRjhr3Nw+JYvtzBx0Iq/178xMnGKg==
-----END RSA PRIVATE KEY-----`

	testCertRSA1024 = `
-----BEGIN CERTIFICATE-----
MIIB4zCCAUygAwIBAgIRANSysyC3vJlv86Ttmi8M97owDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwIBcNMjMwNDE3MTM0MTM3WhgPMjEwMDAxMDEw
MDAwMDBaMBMxETAPBgNVBAoTCEF1dGhlbGlhMIGfMA0GCSqGSIb3DQEBAQUAA4GN
ADCBiQKBgQC4ntJ/qcs9yBQihZkrF5v2Pdp6Rr8uNc4GDjuOsVGUohpwcjVobAuj
AuvCG646cnekbkJOm1bY+38F+nfWJ7ny9RYMp1ng6xWR6vpzZiPyJI89FQU3gd8f
WDI5Xn2ZvrSqfgEJhXMAWn7EPgUajlbLoPzYFCKSChIpR9umk5DBnQIDAQABozUw
MzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/
BAIwADANBgkqhkiG9w0BAQsFAAOBgQBjpYkj+iE9XoA0q8Iq8+CYwlRwQ76jHKgy
z+0JCJE10ysuDPqRJEGJR3vfOs6VyNTGcvdCemPkTEYYAikaT4ydRNqIwefuHlx0
7Abr/GUkZpRdTNfitAZbN4HpHpxZhx/A4yNutwGLiZSzqsn1r1VxTymSkNLa680X
84rsVRZppA==
-----END CERTIFICATE-----`

	testKeyRSA1024 = `
-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQC4ntJ/qcs9yBQihZkrF5v2Pdp6Rr8uNc4GDjuOsVGUohpwcjVo
bAujAuvCG646cnekbkJOm1bY+38F+nfWJ7ny9RYMp1ng6xWR6vpzZiPyJI89FQU3
gd8fWDI5Xn2ZvrSqfgEJhXMAWn7EPgUajlbLoPzYFCKSChIpR9umk5DBnQIDAQAB
AoGAHNF93jus5An1SqY8EIPw7nEdR3T/psDzVfKmzVFUgLUFF4RcXd5vupRcJMKZ
Ybo4fsxPQWHyHpCzdUVxq1YsKkK5qaAGjUfyHKP9yS/ZTKzA4BQJ+mOxagdZ79PB
dxrxtRWz58x++537TNGAUNziG7zaLOmdwlqul4bYjHt5FCkCQQDhPwFMbrpjT4oM
fDuy1bWS4t3X4VVBZfEQT3WeLu4qHzCnBbEszL/q3bXtKlbiqcWRwhrOvCqkmY8v
XBAb21yTAkEA0dPVHCcgXKWIytz7DOGEcBoO8ANw/918VoA9LW560pL5B1qzYAl+
7Ecl6zoZLJPVY1BQ+HVE5tLih84hmlp3DwJAZnHQdmHaFfcEE3+ha0n1plPWkCwl
KXRi+ocZOJOhsLi02RImrfiFxR2Hc9GQ6NBMUmnU5XgBcRGCZQjbLsBLTwJAEct7
SVXwIqtPPJUdHWyKxM8Q8T35eVmZT+S0S4QRGoaoY/1HNR/ZCcTG7HoS5HrtH+0R
0OBxJXpBB+9tXh/J9QJAJDVE0lcJWKeUl/W3P/pzCXZHdUtqoYUFuuuJyFnAAFFU
CKi6wGKnfsc0v01tVpooyThJ+4Z9eTotNGp6ke1tTA==
-----END RSA PRIVATE KEY-----`

	testCertRSA2048 = `
-----BEGIN CERTIFICATE-----
MIIC6DCCAdCgAwIBAgIRAKtA7n5MZorcd0TquNdXF74wDQYJKoZIhvcNAQELBQAw
EzERMA8GA1UEChMIQXV0aGVsaWEwIBcNMjMwNDE3MTM0NDI4WhgPMjEwMDAxMDEw
MDAwMDBaMBMxETAPBgNVBAoTCEF1dGhlbGlhMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAugGNhIscNEpZtxdrTuRAyeGjSERuue56uF0lMDvh8YDi0alQ
y4q80FmIwfP9lWmji9IkMvf1iD0M2I55WBPrL9ddqtpdkIfLgz6eX791LKtF9n8u
PwuX2jUDtWdJrMlOIJ8wCcTjyCFzjTFujAtXGffjWt4tlKCWZZUkJqmyfBiQIag6
ZFb1S6VXFXFpWTOIc41X2VBmzpSLnfqEDqgp/KMDja1tDYAhFh3IAFSzBpXqUGT4
cH2wJUcrngdaiHR/2ToJRu66jK8akB35gjmKiF/Tp9pUI7/rBgPeFWDCZuuZa3k3
brfjJkynQAoCNajfxy8cglCxAuG+jWubFDPGewIDAQABozUwMzAOBgNVHQ8BAf8E
BAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIwADANBgkqhkiG
9w0BAQsFAAOCAQEAoIXTZcKC13KW1GNhzx9ECFTs/gatfjkYONRz+M2wjpVGHzsn
JUPXhoT1SL9WdWYXVVCUXrQge9n9n6IccDjwddnoWL2JMXnNd2PAJLgwE2Xfd40o
U1CLTwvsVNCXjoQLyjkg+SbWGqApVS3oj+A8RTtSBbztP+CoOqbyD3Roo1sFHeE8
PXYboT5bIIaU7DaxhItGHwVDLLOSD72FP/5i+ZmFse2EzUUdyi6d4FSjk7pZCX1T
/2w/bqk3zRemBqDwTnH+sMhPUPvcOg6AIR5YdjWYSDz45sDdgBpUZYYTPfSz2nUL
PJwsB/gk0asMwSYprat6sJ0X3xrtg1ak3a7LkQ==
-----END CERTIFICATE-----`

	testKeyRSA2048 = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAugGNhIscNEpZtxdrTuRAyeGjSERuue56uF0lMDvh8YDi0alQ
y4q80FmIwfP9lWmji9IkMvf1iD0M2I55WBPrL9ddqtpdkIfLgz6eX791LKtF9n8u
PwuX2jUDtWdJrMlOIJ8wCcTjyCFzjTFujAtXGffjWt4tlKCWZZUkJqmyfBiQIag6
ZFb1S6VXFXFpWTOIc41X2VBmzpSLnfqEDqgp/KMDja1tDYAhFh3IAFSzBpXqUGT4
cH2wJUcrngdaiHR/2ToJRu66jK8akB35gjmKiF/Tp9pUI7/rBgPeFWDCZuuZa3k3
brfjJkynQAoCNajfxy8cglCxAuG+jWubFDPGewIDAQABAoIBAD7IkWUAs4du5TNo
wz7AyqGZ+MxG1P0LYv7h6dCLFeu3blgIh438iVjmL8QPwDNzkdF7H97YVVckDDb4
eDrjlknyrtohlN1ZCLeHJlv5OurV8OqP6SM8nYf4xwSvFW4uEKHwOX3CqIP/zooE
+mRo24CXbHVacxYs0jb9jVNDikxaRh57aL9wqvg8/SUhe2T2VXzTO9z0ZYNuIyF/
EkOLkRM1FY7o3zHuYeRN2Y/QbS7yOT4aIodHBtwvObWs8g6Sq8cSbss72x22fY8J
uPwDASy93EOfn7phx9rGq6ioy5XpZtiGHPs+dThftBoOUCq3x2PxdVjGRwnZ1LRd
kEhWm3kCgYEA8LelRt6go8tqUGj82Y+RkKlbviO7+P8Hol1B1fV+wIlLK6CRm6Mb
3PB7fRdBsxcuBFf8WfVdfIkSHEi2uG+Ehb54AySTdsgRYHSQtSYMvKyKxHcsEJ6x
5uNK425N6wVS86MGoIK5li6jpVwdnA8bE6cvEMBRMXlTTIzvwM1+AV8CgYEAxdCw
qqaS6aYBH5Ol0hUi3wW6HaLUOVGksrn3ZTfAGsgoX/mgF8qinz1NmuSzKoevvVK9
q9TC+GDrVNQcD7s2kQQD3Ni8Jb3ok360HcqRN+n8O5+v/t7OjHSsqSRmnQX/lvWO
+65r442ziHXrhFT2CwmY3zBHXe2+MBrGHmrgRGUCgYEAkRoSbdrrOHD44Am5SSfq
1inQnJgLyjdpEa1nbyLxyfu4rU64FvpGZHMt7SSkvODfI00qV8u5E8XIffYy9pB6
cOh0jWhx36sQFnWNeTS7fsv/RhiUHlya3pPqY5ftLhtiemyuJPlIB8iLarVRP+43
IyynCVD0YH9DACUArNbx+r8CgYBpd9EZy2I9DPNAYLpifj5vZmBK+MvqG6uSVzCe
WNEl9l4AfdlrlfCKsma0FQepv1pluL3D5dZmE1aljcnAYXLAcsGUeEIoZU6hhUaH
M7+lbi27pHJzk1vQ60w7ilrjkZUqaZZofiCr3JtCQIznq1zbmaxWIymJ3P4wK7ZB
9X3JOQKBgQDHPBrRVP5Od7dd/C+yi7p2CujuLbV4vKczpNf7Kj+OwQZtnZxjzi95
ObquHDkrz5+OhkvZwKNKO89UQzvhT+7gpQpZJ/gdVl1oNuZ5gAsJ4aDWX+NZ00Z0
Nkb2HpYR8xlnt6rFV465SYun1CIz5h1sUD1T75LDK340SrMbQq2Ilw==
-----END RSA PRIVATE KEY-----`

	testCertRSA4096 = `
-----BEGIN CERTIFICATE-----
MIIE5zCCAs+gAwIBAgIQFaK8pCAGUonNI4J2+aHgGzANBgkqhkiG9w0BAQsFADAT
MREwDwYDVQQKEwhBdXRoZWxpYTAgFw0yMzA0MTcxMzQ1MTBaGA8yMTAwMDEwMTAw
MDAwMFowEzERMA8GA1UEChMIQXV0aGVsaWEwggIiMA0GCSqGSIb3DQEBAQUAA4IC
DwAwggIKAoICAQCjdNqfv8DzXBfR4XXsskYFcUYSLx17BrQ1hQqyLCTaAluH80OE
7HyZbbklKyvw4ig5Rk8Slq1pv6JBK3dEsOWWth2BNZEQyyGXa/aa77Zj6FxdDRHb
H/retzJzxTaaiu+F50OU0PkE5clj/V5JsgKwm70GwOy6zLkGnlv0k0I8HzTYJd1u
9qfKLZpmkJ0oiR4TogEgZA+9atCxzAFbcrynEnStIrepxvad9oOlzFLxArFe5Ai8
IsT3hgIo1SHFSVMhPdfsXVt1nIo5h2Ol2Ry932sIIDypNc70KsYgzQ4jC/6iRni4
saKoUp9IIDCRl+zjlLM1csiufhzC8U0g+UvWFkzigTW4J+CneQk9nnb5BtfCAiir
6WjOicQJff9EuvQFYASljQypH8hunKcH9YWtT/DGRThpWRgDMMalHnEprC4uSrYy
1QajLCi0ncJIArW3SdyePc7tRebNIxY3/Phj5kMwfV+ypIso5nJyu41AQVaRT7U5
+YHydvg3FGOa7JDUm1a27BQgpocz5yU5aazbffmPz/eRPqMa/YRsmzBLnuhkR2Fo
6aRoU0a3zQe9LcrP8gdxr2ZQqZYzdVJ2feywaeH6RN6jl3S2IlH/j71G0dyi3nSN
nC4pe7CHH/wtE6NYCzoPcpZIcDqWt1aCFKYCK1MacmWSycxzv0dzJo+CnwIDAQAB
ozUwMzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0T
AQH/BAIwADANBgkqhkiG9w0BAQsFAAOCAgEAL65NXfJDjTF/GeoB6s8V3fvuwviu
RVQZsbKO4i7nCgs7lFJbXDkW+ybXhq1fmOeCnD1BaF5wwBNB4rgb70PT2eGpKbC7
7U4sgghU7CYSS5sc3dP2xAS1NaCrTAa3m8tlzPqhkR1VkZ9X16fG7kTp/CPXMKMT
zoBj+v3Iv3Gf+gSJu6a2PmvAPuV4w78PH9Lz2sZ7FkZYOYhzNgUS++vL2k53DIW9
6mIqMhm6y9yzsxJx9FXKBWqNuqpL3Gp5KL0XWFy4JIbnpX1b0J4C8yAvF5QS8viO
3VFFcgGB5VWS6vMDAp/c6O+9Rzg0ZbfnLYAHxeSMGZ/Zkf/TnHNjKASFmEu912iH
c6ulT8hVxwTxi/P8eereFdsMib8Z0z871e/2KGZN9bwycVIsqZIllTM3vqcwc9wi
uu6eoScqx25qut2G7K2aQxtfPHmPfyh7/Ft3jZra3apzEuRE3KRBVWaSbDI2SjoP
LESJpBtFPnKOt2p9p/70iODv2lrfoMpj4eXXztJAJFUi4KkczomrU1WtJDc8J5Pp
9tBiNFR1bBKE4+9kwY+6x8LMJs94XjlbG7stoPki41qGR/8Th+n33GcIF4n9Up9l
2XR5/Iqewj2FJAkiYcalFasScU/hLTyjJzpYMOtAvVbBgvYm8IQ4Q5VBkQPPe6a4
P+3smf8j9ywptqk=
-----END CERTIFICATE-----`

	testKeyRSA4096 = `
-----BEGIN RSA PRIVATE KEY-----
MIIJJwIBAAKCAgEAo3Tan7/A81wX0eF17LJGBXFGEi8dewa0NYUKsiwk2gJbh/ND
hOx8mW25JSsr8OIoOUZPEpatab+iQSt3RLDllrYdgTWREMshl2v2mu+2Y+hcXQ0R
2x/63rcyc8U2morvhedDlND5BOXJY/1eSbICsJu9BsDsusy5Bp5b9JNCPB802CXd
bvanyi2aZpCdKIkeE6IBIGQPvWrQscwBW3K8pxJ0rSK3qcb2nfaDpcxS8QKxXuQI
vCLE94YCKNUhxUlTIT3X7F1bdZyKOYdjpdkcvd9rCCA8qTXO9CrGIM0OIwv+okZ4
uLGiqFKfSCAwkZfs45SzNXLIrn4cwvFNIPlL1hZM4oE1uCfgp3kJPZ52+QbXwgIo
q+lozonECX3/RLr0BWAEpY0MqR/IbpynB/WFrU/wxkU4aVkYAzDGpR5xKawuLkq2
MtUGoywotJ3CSAK1t0ncnj3O7UXmzSMWN/z4Y+ZDMH1fsqSLKOZycruNQEFWkU+1
OfmB8nb4NxRjmuyQ1JtWtuwUIKaHM+clOWms2335j8/3kT6jGv2EbJswS57oZEdh
aOmkaFNGt80HvS3Kz/IHca9mUKmWM3VSdn3ssGnh+kTeo5d0tiJR/4+9RtHcot50
jZwuKXuwhx/8LROjWAs6D3KWSHA6lrdWghSmAitTGnJlksnMc79HcyaPgp8CAwEA
AQKCAgAO31QBEwZwXiHAs/3x0mqylhLlFqpdBkghUoCdo4ya1XoUjZrIHmhb4XLm
Id52pW05gN8y9sjChXAy88x/UIUjSGC43/HaEFF3IJiokkULJBo7UTQdtvQxjYOm
qvwD5b5Tda5dfQIbYvkHAwewNuUtwo3ZbnZbrMLtCj2drERrif9Z52AVd5XevHV+
/Yt/I7K74JKvqssP1gc1FjXNZ0wo+3HoSu9hIDxSNRrXXBbz3OXcl20ACT3Ys7XA
l1viQoCw1pqt4/StZ9ff0iTL80w9LnXjoGNEliPFbZrnYyD1KWM6yqSzUV5WaGYb
vuoMZUFll6MSquX9knX1etUkueofZE631DEtXagjdgoPwShxHzJM6dPqrelm79u0
FWrptMCKActwGnFMOow8jDBnJ/ns7g95fKWthCZgyHhJHhs5EpNFEc1WTRuV+/Yb
564oJlgwTrSsNdBtdtmNLVFX2YqMXjeFkJuiJDFHEFpV8fMmHgVko0ihpFrohSza
ftfwCQKkd/L4huN4hMfZYwVd9rdeUqsJGXASnDu0f4Q1Bq6H9zQ//A+wBkh70Sq+
2cYsW/GryI4h/Q5677xtKUA9IDMIz2opBBewjCJH+BRPUobcuyq/xWod217mo8gn
mHN5/8ysFPWJH42wL3aO20L0XUrps4n9ni5qW7sbiGg/VZCQyQKCAQEA2WpR4Csw
HOkhfnvay7NOEAiud/MLPLNPVzQioR+SdOqWuQ81Dx1iZf6ZkMJzKi7l/I9txG39
QAh3aEUXStpjcmQygn+LS9zmZsaACnii/kBqNEJ7EVDkP/kL9qecgAUwSPB9xD8P
1SpcUVqcXfDpN9jQduLTHHcXgs5vUIUwFcC3OJKTMwWbSxMW2WmBsWrI/QqjprI3
sFKHKOPs+RRALUF44L8jMnODzwbMEK1Bui+eFiNOI5QbU2cd/uVkKe+PZTMexziX
cBUa7gzrE2dDccLkVup8dRsXjZJlu2ET8VZYncLidjBzAiTy7E8hZwuHzzovvppr
c2PHvlX0JtrEvQKCAQEAwHcPEOCQP/cswnYgRSsI3uDCZN3PpiMQd1qKYiIQItKT
hBRPvYqzvDcsprSnqvcWqd6jlbvyhgZyi8xQt+Iyduib3scoOZ2w8P94B5YR3rT7
u+BVd/A4iUxU0f8FfZ6sbwE5Q76XZdWwvFuy+lloM8y4BFe//VicA0JO7Rt4lFN2
31WSGwa7nJuPbbIPCeY9hPQVkdgsVgSksW+TjuUxiRVUvcF60rAuErcDBnO4XcC6
f/qGlLAUAzmQQqI0WpfldAFlYCrZi5UuJJaA7/w7h0a9MQmPx4UCcg4D1IZiyt0U
psueRzFBAndsgVLd+/v5t8Tkrz3M3/Dsw9yDg8RwiwKCAQB3gD7cjiB145YrZXxP
dpCzs3HiME6+4Hf9oIRgN3BSnxaVRUyOsEIDebuCm76dMwXqmhNlYmdOqNipEUDK
PdtnZrd0jxJLcnGZkAWUu9YrFdDKRLhMPkAXAZaXzmzw2Ok/TiBym47iRdRUSw+j
euVVcvCyR95tyO+9UCZTBcH2UuTiTX5nDu/ahfWLLrjAgcdTfmORHmgJnHL6AL2h
8oWL2m7MaYK5GlEam8vSZsi3w7CKzoEGgUO7xfPwxLkXa7tPjpeePPbP/mm86pDT
K3EguFS1iVE7NNbvU8ZjBermPeWbYSEEgYDVbuWvCZd8ghP1zS+s/keNNwz1C12V
da2pAoIBAFsBsSkM1oi4ivykuJucPsSMyL7DN6XaXLXjJR5D9xdQNRq2NAJvLI/q
Ev383G92CMxoDzgFOCdxswYxpVVd6vjZAqMzzux3iSxb0Fjd+DMzpvjumdttxn39
jvoBOYpt1iFjFb3XyGUJx1k5jwbb8e7UdYrwJ0NXe+X6m7F4VOrmEIaIQt7urxXd
ZNO852mJ6jsM44okCsrdxTZ1iPN/oo2sfXaAn2AymIaW7SJG473JHSbYwnxaSgxA
Utt/MXxI6OGSq2nuuRFMiBYa6HsR7OAJbfpbCBaS6VYfFGaQ6PP91/8Ktxv4yUGu
UKtSEM9PFYR04KGQemjF1l7CzZkn8QMCggEAetZUlucrPeejwSaIJvxcyO59Vp73
I5Yp6wnIQ3SZ3SIyx3GRJNU6uB8S6GR9Zu6CzdHSjuU9oBdmd4WdRle9H+8hooq0
xWbtpZE90cXvx36Z1IqILSu1ZTJrdsCxTiU9vQmg23jRl5z8K2YnsN4ury7qiZBt
SPD051WfyTeyLG3A8gx9ugw3vyJwXgvE6d82BJvJoWS9IzuK7xGULD80zO/l18P+
9ixzCh3rTi46FWESKu58HYNtWdTb0NolrqKexAx+IXhPBc8Zj7+Ip5u5s6XDV0Ek
tk96Zvf7GMFfPqRP6gidXeouyTu3pdcR7bCT9bVC0AcT/4n/L2D99ftFYQ==
-----END RSA PRIVATE KEY-----`

	testCertECDSAWithP224 = `
-----BEGIN CERTIFICATE-----
MIIBRzCB9qADAgECAhB51uvUDHkaxlSEs8cgoYBRMAoGCCqGSM49BAMCMBMxETAP
BgNVBAoTCEF1dGhlbGlhMCAXDTIzMDQxNzEzMTIwMloYDzIxMDAwMTAxMDAwMDAw
WjATMREwDwYDVQQKEwhBdXRoZWxpYTBOMBAGByqGSM49AgEGBSuBBAAhAzoABJa4
oEFZqEbmsnKWXEfNWTiqyEq6YiWbVFIH/PGijaRmsYpKC2UBGsscN4DziAUHBlqX
KLA/lsRjozUwMzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEw
DAYDVR0TAQH/BAIwADAKBggqhkjOPQQDAgNAADA9Ah0Aq03epx31NN1fTorB/rrz
Muu9Taw8YxaZxvjLaQIcbNHGY5bFYxi04ahvN1rYi2sJEn66SaWut+lBIw==
-----END CERTIFICATE-----`

	testKeyECDSAWithP224 = `
-----BEGIN EC PRIVATE KEY-----
MGgCAQEEHM0126u3fW5scirH39HU9FgPTZOHPxg2NgbSQQ+gBwYFK4EEACGhPAM6
AASWuKBBWahG5rJyllxHzVk4qshKumIlm1RSB/zxoo2kZrGKSgtlARrLHDeA84gF
BwZalyiwP5bEYw==
-----END EC PRIVATE KEY-----`

	testCertECDSAWithP256 = `
-----BEGIN CERTIFICATE-----
MIIBWzCCAQKgAwIBAgIRANqK3vKflYMr/2HGVd3aOR0wCgYIKoZIzj0EAwIwEzER
MA8GA1UEChMIQXV0aGVsaWEwIBcNMjMwNDE3MTMxNjI5WhgPMjEwMDAxMDEwMDAw
MDBaMBMxETAPBgNVBAoTCEF1dGhlbGlhMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcD
QgAEnnBdDSXbTgHtrc5vcJ2xz6qyGXM8PJgENjgQgn5WFVQCSZnKp08+mzeDiHrM
67KmISfxSAjoeCJV+dP6JfxIVqM1MDMwDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQM
MAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwCgYIKoZIzj0EAwIDRwAwRAIgOo+m
1yQsTmqOaKak9MY2q7CdBI9Di8vPK/sE/x5JIPYCIA/lyI/sG1EEdLT8g3M4Joc3
VK7cBHjmftnZL6kiS+Dn
-----END CERTIFICATE-----`

	testKeyECDSAWithP256 = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHL87FDsqijXFhRJ5VgYiOz2ko6xxP7aP7i4v3Eowf4KoAoGCCqGSM49
AwEHoUQDQgAEnnBdDSXbTgHtrc5vcJ2xz6qyGXM8PJgENjgQgn5WFVQCSZnKp08+
mzeDiHrM67KmISfxSAjoeCJV+dP6JfxIVg==
-----END EC PRIVATE KEY-----`

	testCertECDSAWithP384 = `
-----BEGIN CERTIFICATE-----
MIIBmDCCAR6gAwIBAgIQWFgOoTSBNa4F1A+Uk5fBhTAKBggqhkjOPQQDAjATMREw
DwYDVQQKEwhBdXRoZWxpYTAgFw0yMzA0MTcxMzE1MDBaGA8yMTAwMDEwMTAwMDAw
MFowEzERMA8GA1UEChMIQXV0aGVsaWEwdjAQBgcqhkjOPQIBBgUrgQQAIgNiAARq
Fk2dSauZd2mW0ZXuxZ0k2a5PInZOs3wbzjJr67RPzmPMNGt5dVHtbOTLr9MAcm21
E6/4CLQZ+wMq4Zxuhoa02VN4lQBFOzWFPwVTa0lcOUCkJ7E7JWXiZjX80ROyqDOj
NTAzMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMB
Af8EAjAAMAoGCCqGSM49BAMCA2gAMGUCMQCJHEN22ouKJr0usue9/bUsJltPrgSW
v7NjjQ9hY96JAwBQpTxX6EksQdnl44Q/LLACMHBZn3weWvq8frMOAmAvOomMsnMp
H7tweTJNXh4V8XdtR2GGxAAYbq/ShyxrpQ6LVA==
-----END CERTIFICATE-----`

	testKeyECDSAWithP384 = `
-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDBd2neGG9Ax14sDR0V0TYSXIBxNWZwYr7OAFd57MRUZ/+BkHvQEMOoV
umd/tOgGjEagBwYFK4EEACKhZANiAARqFk2dSauZd2mW0ZXuxZ0k2a5PInZOs3wb
zjJr67RPzmPMNGt5dVHtbOTLr9MAcm21E6/4CLQZ+wMq4Zxuhoa02VN4lQBFOzWF
PwVTa0lcOUCkJ7E7JWXiZjX80ROyqDM=
-----END EC PRIVATE KEY-----`

	testCertECDSAWithP521 = `
-----BEGIN CERTIFICATE-----
MIIB5DCCAUWgAwIBAgIRAIpQUsZLrSAJ7+PY4U0MIaYwCgYIKoZIzj0EAwIwEzER
MA8GA1UEChMIQXV0aGVsaWEwIBcNMjMwNDE3MTMxNTM2WhgPMjEwMDAxMDEwMDAw
MDBaMBMxETAPBgNVBAoTCEF1dGhlbGlhMIGbMBAGByqGSM49AgEGBSuBBAAjA4GG
AAQAO8GuJvWACDYuO1ZhMdbrINK8AM8B2xFn5nSvAHAgYolyXz8yxLjmFT1/ifQZ
QjnocX4j/zOGIt1f1OXQvPSRaiQAzWlFIejCKChBK0hiDqfTyzDgrJGiCobL1bgr
yxO3oDg70YeN3mr0OkMvdrIBjpGpGkt5AX6XyaIau9ogZJz6gyOjNTAzMA4GA1Ud
DwEB/wQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMAoG
CCqGSM49BAMCA4GMADCBiAJCAKaUek+/zGw7Tt3L5rqQhXFzWI4mfci+jD99/JHY
UT/FX2Co/tjEcyty46mMWsw6E7q6XCJ6gx38SCWe7wRdMXQMAkIA6rlkntqo6r+j
PoTtPVmbkFFc5ficw+xlmhuKblKrq+u8sbnm+J62C8pXuzSc8dtKEe0+oORD5HH9
YGuoIKNL2vg=
-----END CERTIFICATE-----`

	testKeyECDSAWithP521 = `
-----BEGIN EC PRIVATE KEY-----
MIHcAgEBBEIAe0mKO82UiFUDM3M3CgyEKkXuXnt0m2DAnW3Yf2nadim00n/XsGw7
+ID6Zz5Xhazpx7WFNNhtrjbNQOKbsQNndPugBwYFK4EEACOhgYkDgYYABAA7wa4m
9YAINi47VmEx1usg0rwAzwHbEWfmdK8AcCBiiXJfPzLEuOYVPX+J9BlCOehxfiP/
M4Yi3V/U5dC89JFqJADNaUUh6MIoKEErSGIOp9PLMOCskaIKhsvVuCvLE7egODvR
h43eavQ6Qy92sgGOkakaS3kBfpfJohq72iBknPqDIw==
-----END EC PRIVATE KEY-----`

	goodOpenIDConnectClientSecret = "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng" //nolint:gosec
)
