package validator

import (
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

	assert.EqualError(t, validator.Errors()[0], errFmtOIDCNoPrivateKey)
	assert.EqualError(t, validator.Errors()[1], errFmtOIDCNoClientsConfigured)
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

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: cors: option 'endpoints' contains an invalid value 'invalid_endpoint': must be one of 'authorization', 'pushed-authorization-request', 'token', 'introspection', 'revocation', 'userinfo'")
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

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'enforce_pkce' must be 'never', 'public_clients_only' or 'always', but it is configured as 'invalid'")
	assert.EqualError(t, validator.Errors()[1], errFmtOIDCNoClientsConfigured)
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

	assert.EqualError(t, validator.Errors()[0], errFmtOIDCNoClientsConfigured)
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
				"identity_providers: oidc: one or more clients have been configured with an empty id",
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
			Errors: []string{"identity_providers: oidc: client 'client-1': option 'policy' must be 'one_factor' or 'two_factor' but it is configured as 'a-policy'"},
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
			Errors: []string{errFmtOIDCClientsDuplicateID},
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
				fmt.Sprintf(errFmtOIDCClientRedirectURICantBeParsed, "client-check-uri-parse", "http://abc@%two", errors.New("parse \"http://abc@%two\": invalid URL escape \"%tw\"")),
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
				fmt.Sprintf(errFmtOIDCClientRedirectURIAbsolute, "client-check-uri-abs", "google.com"),
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
				fmt.Sprintf(errFmtOIDCClientInvalidSectorIdentifier, "client-invalid-sector", "https://user:pass@example.com/path?query=abc#fragment", exampleDotCom, "scheme", "https"),
				fmt.Sprintf(errFmtOIDCClientInvalidSectorIdentifier, "client-invalid-sector", "https://user:pass@example.com/path?query=abc#fragment", exampleDotCom, "path", "/path"),
				fmt.Sprintf(errFmtOIDCClientInvalidSectorIdentifier, "client-invalid-sector", "https://user:pass@example.com/path?query=abc#fragment", exampleDotCom, "query", "query=abc"),
				fmt.Sprintf(errFmtOIDCClientInvalidSectorIdentifier, "client-invalid-sector", "https://user:pass@example.com/path?query=abc#fragment", exampleDotCom, "fragment", "fragment"),
				fmt.Sprintf(errFmtOIDCClientInvalidSectorIdentifier, "client-invalid-sector", "https://user:pass@example.com/path?query=abc#fragment", exampleDotCom, "username", "user"),
				fmt.Sprintf(errFmtOIDCClientInvalidSectorIdentifierWithoutValue, "client-invalid-sector", "https://user:pass@example.com/path?query=abc#fragment", exampleDotCom, "password"),
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
				fmt.Sprintf(errFmtOIDCClientInvalidSectorIdentifierHost, "client-invalid-sector", "example.com/path?query=abc#fragment"),
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
				fmt.Sprintf(errFmtOIDCClientInvalidConsentMode, "client-bad-consent-mode", strings.Join(append(validOIDCClientConsentModes, "auto"), "', '"), "cap"),
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
				fmt.Sprintf(errFmtOIDCClientInvalidPKCEChallengeMethod, "client-bad-pkce-mode", "abc"),
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
				fmt.Sprintf(errFmtOIDCClientInvalidPKCEChallengeMethod, "client-bad-pkce-mode-s256", "s256"),
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'scopes' must only have the values 'openid', 'email', 'profile', 'groups', 'offline_access' but one option is configured as 'bad_scope'")
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'grant_types' must only have the values 'implicit', 'refresh_token', 'authorization_code', 'password', 'client_credentials' but one option is configured as 'bad_grant_type'")
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'response_modes' must only have the values 'form_post', 'query', 'fragment' but one option is configured as 'bad_responsemode'")
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: client 'good_id': option 'userinfo_signing_algorithm' must be one of 'none, RS256' but it is configured as 'rs256'")
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

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtOIDCClientPublicInvalidSecret, "client-with-invalid-secret"))
	assert.EqualError(t, validator.Errors()[1], fmt.Sprintf(errFmtOIDCClientRedirectURIPublic, "client-with-bad-redirect-uri", oauth2InstalledApp))
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

	assert.EqualError(t, validator.Warnings()[0], "identity_providers: oidc: client 'client-with-invalid-secret_standard': option 'secret' is plaintext but it should be a hashed value as plaintext values are deprecated and will be removed when oidc becomes stable")
}

func TestValidateIdentityProvidersShouldSetDefaultValues(t *testing.T) {
	timeDay := time.Hour * 24

	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: MustParseRSAPrivateKey(testKey1),
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "a-client",
					Secret: MustDecodeSecret(goodOpenIDConnectClientSecret),
					RedirectURIs: []string{
						"https://google.com",
					},
					ConsentPreConfiguredDuration: &timeDay,
				},
				{
					ID:                       "b-client",
					Description:              "Normal Description",
					Secret:                   MustDecodeSecret(goodOpenIDConnectClientSecret),
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
				{
					ID:     "c-client",
					Secret: MustDecodeSecret(goodOpenIDConnectClientSecret),
					RedirectURIs: []string{
						"https://google.com",
					},
					ConsentMode: "implicit",
				},
				{
					ID:     "d-client",
					Secret: MustDecodeSecret(goodOpenIDConnectClientSecret),
					RedirectURIs: []string{
						"https://google.com",
					},
					ConsentMode: "explicit",
				},
				{
					ID:     "e-client",
					Secret: MustDecodeSecret(goodOpenIDConnectClientSecret),
					RedirectURIs: []string{
						"https://google.com",
					},
					ConsentMode: "pre-configured",
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

	// Assert Clients[0] ends up configured with the correct consent mode.
	require.NotNil(t, config.OIDC.Clients[0].ConsentPreConfiguredDuration)
	assert.Equal(t, time.Hour*24, *config.OIDC.Clients[0].ConsentPreConfiguredDuration)
	assert.Equal(t, "pre-configured", config.OIDC.Clients[0].ConsentMode)

	// Assert Clients[1] ends up configured with the correct consent mode.
	assert.Nil(t, config.OIDC.Clients[1].ConsentPreConfiguredDuration)
	assert.Equal(t, "explicit", config.OIDC.Clients[1].ConsentMode)

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

	assert.Equal(t, "implicit", config.OIDC.Clients[2].ConsentMode)
	assert.Nil(t, config.OIDC.Clients[2].ConsentPreConfiguredDuration)

	assert.Equal(t, "explicit", config.OIDC.Clients[3].ConsentMode)
	assert.Nil(t, config.OIDC.Clients[3].ConsentPreConfiguredDuration)

	assert.Equal(t, "pre-configured", config.OIDC.Clients[4].ConsentMode)
	assert.Equal(t, schema.DefaultOpenIDConnectClientConfiguration.ConsentPreConfiguredDuration, config.OIDC.Clients[4].ConsentPreConfiguredDuration)
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
			oauth2InstalledApp,
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
		assert.Len(t, validator.Errors(), 1)
		assert.ElementsMatch(t, validator.Errors(), []error{
			errors.New("identity_providers: oidc: client 'owncloud': option 'redirect_uris' has the redirect uri 'urn:ietf:wg:oauth:2.0:oob' when option 'public' is false but this is invalid as this uri is not valid for the openid connect confidential client type"),
		})
	})
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
