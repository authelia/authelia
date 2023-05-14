package validator

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
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

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'issuer_private_key' is required")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: option 'clients' must have one or more clients configured")
}

func TestShouldNotRaiseErrorWhenCORSEndpointsValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
					IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
			IssuerCertificateChain: MustParseX509CertificateChain(testCert1),
			IssuerPrivateKey:       MustParseRSAPrivateKey(testKey1),
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
			IssuerCertificateChain: MustParseX509CertificateChain(testCert1),
			IssuerPrivateKey:       MustParseRSAPrivateKey(testKey2),
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

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'issuer_private_key' does not appear to be the private key the certificate provided by option 'issuer_certificate_chain'")
}

func TestShouldRaiseErrorOnKeySizeTooSmall(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey3),
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

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'issuer_private_key' must be an RSA private key with 2048 bits or more but it only has 1024 bits")
}

func TestShouldRaiseErrorOnKeyInvalidPublicKey(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey3),
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

	config.OIDC.IssuerPrivateKey.PublicKey.N = nil

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'issuer_private_key' must be a valid RSA private key but the provided data is missing the public key bits")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadResponseModes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadUserinfoAlg(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:                       "good_id",
					Secret:                   MustDecodeSecret("$plaintext$good_secret"),
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'userinfo_signing_algorithm' must be one of 'none' or 'RS256' but it's configured as 'rs256'")
}

func TestValidateIdentityProvidersShouldRaiseWarningOnSecurityIssue(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:              "abc",
			IssuerPrivateKey:        MustParseRSAPrivateKey(testKey1),
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
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
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
				assert.Equal(t, oidc.SigningAlgNone, have.Clients[0].UserinfoSigningAlgorithm)
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
				have.Clients[0].UserinfoSigningAlgorithm = oidc.SigningAlgRSAUsingSHA256
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].UserinfoSigningAlgorithm)
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
				have.Clients[0].UserinfoSigningAlgorithm = "rs256"
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
				assert.Equal(t, "rs256", have.Clients[0].UserinfoSigningAlgorithm)
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
				"identity_providers: oidc: client 'test': option 'userinfo_signing_algorithm' must be one of 'none' or 'RS256' but it's configured as 'rs256'",
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
				have.Clients[0].TokenEndpointAuthSigningAlg = "abcinvalid"
				have.Clients[0].Secret = MustDecodeSecret("$plaintext$abc123")
			},
			func(t *testing.T, have *schema.OpenIDConnectConfiguration) {
				assert.Equal(t, "abcinvalid", have.Clients[0].TokenEndpointAuthSigningAlg)
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

func MustDecodeSecret(value string) *schema.PasswordDigest {
	if secret, err := schema.DecodePasswordDigest(value); err != nil {
		panic(err)
	} else {
		return secret
	}
}

func MustParseRSAPrivateKey(data string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(data))
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		panic("not pem encoded")
	}

	if block.Type != "RSA PRIVATE KEY" {
		panic("not private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
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
	testCert1 = `
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

	testKey1 = `
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

	testKey2 = `
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

	testKey3 = `-----BEGIN RSA PRIVATE KEY-----
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

	goodOpenIDConnectClientSecret = "$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng" //nolint:gosec
)
