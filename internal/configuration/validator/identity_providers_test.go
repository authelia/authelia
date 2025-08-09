package validator

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
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
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret: "abc",
			},
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option `jwks` is required")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: option 'clients' must have one or more clients configured")
}

func TestShouldRaiseErrorWhenInvalidOIDCServerConfigurationBothKeyTypesSpecified(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "abc",
				IssuerPrivateKey: keyRSA2048,
				JSONWebKeys: []schema.JWK{
					{
						Use:       "sig",
						Algorithm: oidc.SigningAlgRSAUsingSHA256,
						Key:       keyRSA4096,
					},
				},
			},
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option `jwks` must not be configured at the same time as 'issuer_private_key' or 'issuer_certificate_chain'")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: option 'clients' must have one or more clients configured")
}

func TestShouldNotRaiseErrorWhenCORSEndpointsValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
				IssuerPrivateKey: keyRSA2048,
				CORS: schema.IdentityProvidersOpenIDConnectCORS{
					Endpoints: []string{oidc.EndpointAuthorization, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:     "example",
						Secret: tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	assert.Len(t, validator.Errors(), 0)
}

func TestShouldRaiseErrorWhenCORSEndpointsNotValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
				IssuerPrivateKey: keyRSA2048,
				CORS: schema.IdentityProvidersOpenIDConnectCORS{
					Endpoints: []string{oidc.EndpointAuthorization, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo, "invalid_endpoint"},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:     "example",
						Secret: tOpenIDConnectPlainTextClientSecret,
					},
				},
			},
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: cors: option 'endpoints' contains an invalid value 'invalid_endpoint': must be one of 'authorization', 'device-authorization', 'pushed-authorization-request', 'token', 'introspection', 'revocation', or 'userinfo'")
}

func TestShouldRaiseErrorWhenOIDCPKCEEnforceValueInvalid(t *testing.T) {
	validator := schema.NewStructValidator()

	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
				IssuerPrivateKey: keyRSA2048,
				EnforcePKCE:      testInvalid,
			},
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'enforce_pkce' must be 'never', 'public_clients_only' or 'always', but it's configured as 'invalid'")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: option 'clients' must have one or more clients configured")
}

func TestShouldRaiseErrorWhenOIDCCORSOriginsHasInvalidValues(t *testing.T) {
	validator := schema.NewStructValidator()

	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
				IssuerPrivateKey: keyRSA2048,
				CORS: schema.IdentityProvidersOpenIDConnectCORS{
					AllowedOrigins:                       utils.URLsFromStringSlice([]string{"https://example.com/", "https://site.example.com/subpath", "https://site.example.com?example=true", "*"}),
					AllowedOriginsFromClientRedirectURIs: true,
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  "myclient",
						Secret:              tOpenIDConnectPlainTextClientSecret,
						AuthorizationPolicy: "two_factor",
						RedirectURIs:        []string{"https://example.com/oauth2_callback", "https://localhost:566/callback", "http://an.example.com/callback", "file://a/file"},
					},
				},
			},
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	require.Len(t, validator.Errors(), 5)
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: cors: option 'allowed_origins' contains an invalid value 'https://example.com/' as it has a path: origins must only be scheme, hostname, and an optional port")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: cors: option 'allowed_origins' contains an invalid value 'https://site.example.com/subpath' as it has a path: origins must only be scheme, hostname, and an optional port")
	assert.EqualError(t, validator.Errors()[2], "identity_providers: oidc: cors: option 'allowed_origins' contains an invalid value 'https://site.example.com?example=true' as it has a query string: origins must only be scheme, hostname, and an optional port")
	assert.EqualError(t, validator.Errors()[3], "identity_providers: oidc: cors: option 'allowed_origins' contains the wildcard origin '*' with more than one origin but the wildcard origin must be defined by itself")
	assert.EqualError(t, validator.Errors()[4], "identity_providers: oidc: cors: option 'allowed_origins' contains the wildcard origin '*' cannot be specified with option 'allowed_origins_from_client_redirect_uris' enabled")

	require.Len(t, config.IdentityProviders.OIDC.CORS.AllowedOrigins, 6)
	assert.Equal(t, "*", config.IdentityProviders.OIDC.CORS.AllowedOrigins[3].String())
	assert.Equal(t, "https://example.com", config.IdentityProviders.OIDC.CORS.AllowedOrigins[4].String())
}

func TestShouldRaiseErrorWhenOIDCServerNoClients(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
				IssuerPrivateKey: keyRSA2048,
			},
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: option 'clients' must have one or more clients configured")
}

func TestShouldRaiseErrorWhenOIDCServerClientBadValues(t *testing.T) {
	mux := http.NewServeMux()

	handleString := func(in string) http.HandlerFunc {
		var h http.HandlerFunc = func(rw http.ResponseWriter, r *http.Request) {
			_, _ = rw.Write([]byte(in))
		}

		return h
	}

	mux.Handle("/sector-1.json", handleString(`["https://google.com"]`))

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	root, err := url.ParseRequestURI(server.URL)
	require.NoError(t, err)

	mustParseURL := func(u string) *url.URL {
		out, err := url.Parse(u)
		if err != nil {
			panic(err)
		}

		return out
	}

	testCases := []struct {
		name    string
		clients []schema.IdentityProvidersOpenIDConnectClient
		warns   []string
		errors  []string
		test    func(t *testing.T, actual []schema.IdentityProvidersOpenIDConnectClient)
	}{
		{
			name: "EmptyIDAndSecret",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "",
					Secret:              nil,
					AuthorizationPolicy: "",
					RedirectURIs:        []string{},
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client '': option 'client_secret' is required",
				"identity_providers: oidc: clients: option 'id' is required but was absent on the clients in positions #1",
			},
		},
		{
			name: "BadIDTooLong",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:     "abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123",
					Secret: tOpenIDConnectPlainTextClientSecret,
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123': option 'id' must not be more than 100 characters but it has 108 characters",
			},
		},
		{
			name: "BadIDInvalidCharacters",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:     "@!#!@$!@#*()!&@%*(!^@#*()!&@^%!(_@#&",
					Secret: tOpenIDConnectPlainTextClientSecret,
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client '@!#!@$!@#*()!&@%*(!^@#*()!&@^%!(_@#&': option 'id' must only contain RFC3986 unreserved characters",
			},
		},
		{
			name: "InvalidPolicy",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-1",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: "a-policy",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-1': option 'authorization_policy' must be one of 'one_factor' or 'two_factor' but it's configured as 'a-policy'",
			},
		},
		{
			name: "InvalidPolicyCCG",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-1",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: "a-policy",
					RedirectURIs: []string{
						"https://google.com",
					},
					GrantTypes: []string{
						oidc.GrantTypeClientCredentials,
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-1': option 'authorization_policy' must be one of 'one_factor' but it's configured as 'a-policy'",
			},
		},
		{
			name: "ClientIDDuplicated",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
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
			errors: []string{
				"identity_providers: oidc: clients: option 'id' must be unique for every client but one or more clients share the following 'id' values 'client-x'",
			},
		},
		{
			name: "RedirectURIInvalid",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-check-uri-parse",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"http://abc@%two",
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-check-uri-parse': option 'redirect_uris' has an invalid value: redirect uri 'http://abc@%two' could not be parsed: parse \"http://abc@%two\": invalid URL escape \"%tw\"",
			},
		},
		{
			name: "RedirectURINotAbsolute",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-check-uri-abs",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"google.com",
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-check-uri-abs': option 'redirect_uris' has an invalid value: redirect uri 'google.com' must have a scheme but it's absent",
			},
		},
		{
			name: "RequestURINotValidURI",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-check-uri-parse",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RequestURIs: []string{
						"http://abc@%two",
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-check-uri-parse': option 'request_uris' has an invalid value: request uri 'http://abc@%two' could not be parsed: parse \"http://abc@%two\": invalid URL escape \"%tw\"",
			},
		},
		{
			name: "RequestURINotAbsolute",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-check-uri-parse",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RequestURIs: []string{
						exampleDotCom,
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-check-uri-parse': option 'request_uris' has an invalid value: request uri 'example.com' must have a scheme but it's absent",
			},
		},
		{
			name: "RequestURINotSecure",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-check-uri-parse",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RequestURIs: []string{
						"http://" + exampleDotCom,
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-check-uri-parse': option 'request_uris' has an invalid scheme: scheme must be 'https' but request uri 'http://example.com' has a 'http' scheme",
			},
		},
		{
			name: "ValidSectorIdentifier",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-valid-sector",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifierURI: root.JoinPath("sector-1.json"),
				},
			},
		},
		{
			name: "InvalidSectorIdentifierInvalidURL",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-invalid-sector",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifierURI: mustParseURL("https://user:pass@example.com/path?query=abc#fragment"),
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier_uri' with value 'https://user:pass@example.com/path?query=abc#fragment': must not have a fragment but it has a fragment with the value 'fragment'",
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier_uri' with value 'https://user:pass@example.com/path?query=abc#fragment': must not have a username but it has a username with the value 'user'",
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier_uri' with value 'https://user:pass@example.com/path?query=abc#fragment': must not have a password but it has a password with the value 'pass'",
			},
		},
		{
			name: "InvalidSectorIdentifierInvalidHost",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-invalid-sector",
					Secret:              tOpenIDConnectPBKDF2ClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifierURI: mustParseURL("example.com/path?query=abc#fragment"),
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier_uri' with value 'example.com/path?query=abc#fragment': must not have a fragment but it has a fragment with the value 'fragment'",
			},
			warns: []string{
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier_uri' with value 'example.com/path?query=abc#fragment': should be an absolute URI",
				"identity_providers: oidc: clients: warnings for clients above indicate deprecated functionality and it's strongly suggested these issues are checked and fixed if they're legitimate issues or reported if they are not as in a future version these warnings will become errors",
			},
		},
		{
			name: "InvalidSectorIdentifierInvalidScheme",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-invalid-sector",
					Secret:              tOpenIDConnectPBKDF2ClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifierURI: mustParseURL("http://example.com/path?query=abc"),
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-invalid-sector': option 'sector_identifier_uri' with value 'http://example.com/path?query=abc': must have the 'https' scheme but has the 'http' scheme",
			},
		},
		{
			name: "EmptySectorIdentifier",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-invalid-sector",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					SectorIdentifierURI: mustParseURL(""),
				},
			},
			test: func(t *testing.T, actual []schema.IdentityProvidersOpenIDConnectClient) {
				assert.Nil(t, actual[0].SectorIdentifierURI)
				assert.True(t, actual[0].SectorIdentifierURI == nil)
			},
		},
		{
			name: "InvalidConsentMode",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
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
			errors: []string{
				"identity_providers: oidc: clients: client 'client-bad-consent-mode': consent: option 'mode' must be one of 'auto', 'implicit', 'explicit', 'pre-configured', or 'auto' but it's configured as 'cap'",
			},
		},
		{
			name: "InvalidPKCEChallengeMethod",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
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
			errors: []string{
				"identity_providers: oidc: clients: client 'client-bad-pkce-mode': option 'pkce_challenge_method' must be one of 'plain' or 'S256' but it's configured as 'abc'",
			},
		},
		{
			name: "InvalidPKCEChallengeMethodLowerCaseS256",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
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
			errors: []string{
				"identity_providers: oidc: clients: client 'client-bad-pkce-mode-s256': option 'pkce_challenge_method' must be one of 'plain' or 'S256' but it's configured as 's256'",
			},
		},
		{
			name: "ValidRequestedAudienceMode",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-good-ram",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					RequestedAudienceMode: "explicit",
				},
			},
		},
		{
			name: "SetDefaultAudienceMode",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-no-ram",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					RequestedAudienceMode: "",
				},
			},
			test: func(t *testing.T, actual []schema.IdentityProvidersOpenIDConnectClient) {
				assert.Equal(t, "explicit", actual[0].RequestedAudienceMode)
			},
		},
		{
			name: "InvalidRequestedAudienceMode",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-bad-ram",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					RequestedAudienceMode: "magic",
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-bad-ram': option 'requested_audience_mode' must be one of 'explicit' or 'implicit' but it's configured as 'magic'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
						IssuerPrivateKey: keyRSA2048,
						Clients:          tc.clients,
					},
				},
			}

			ctx := NewValidateCtx()
			ctx.tlsconfig = &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
			}

			ValidateIdentityProviders(ctx, config, validator)

			errs := validator.Errors()

			require.Len(t, errs, len(tc.errors))

			for i, errStr := range tc.errors {
				t.Run(fmt.Sprintf("Error%d", i+1), func(t *testing.T) {
					assert.EqualError(t, errs[i], errStr)
				})
			}

			if tc.test != nil {
				tc.test(t, config.IdentityProviders.OIDC.Clients)
			}

			warns := validator.Warnings()

			if len(tc.warns) != 0 {
				require.Len(t, warns, len(tc.warns))

				for i, errStr := range tc.warns {
					t.Run(fmt.Sprintf("Warning%d", i+1), func(t *testing.T) {
						assert.EqualError(t, warns[i], errStr)
					})
				}
			}
		})
	}
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadGrantTypes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
				IssuerPrivateKey: keyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
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
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: clients: client 'good_id': option 'grant_types' must only have the values 'authorization_code', 'implicit', 'client_credentials', 'refresh_token', or 'urn:ietf:params:oauth:grant-type:device_code' but the values 'bad_grant_type' are present")
}

func TestShouldNotErrorOnCertificateValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:             "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
				IssuerCertificateChain: certRSA2048,
				IssuerPrivateKey:       keyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
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
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)
}

func TestShouldRaiseErrorOnCertificateNotValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:             "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
				IssuerCertificateChain: certRSA2048,
				IssuerPrivateKey:       keyRSA4096,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
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
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: jwks: key #1 with key id 'c4c7ca-rs256': option 'certificate_chain' does not appear to contain the public key for the private key provided by option 'key'")
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
			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret:              "abc",
						IssuerPrivateKey:        keyRSA2048,
						MinimumParameterEntropy: tc.have,
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
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
				},
			}

			ValidateIdentityProviders(NewValidateCtx(), config, validator)

			assert.Equal(t, tc.expected, config.IdentityProviders.OIDC.MinimumParameterEntropy)

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
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "hmac1",
				IssuerPrivateKey: keyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
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
							oidc.RedirectURISpecialOAuth2InstalledApp,
						},
					},
				},
			},
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	require.Len(t, validator.Errors(), 2)
	assert.Len(t, validator.Warnings(), 1)

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: clients: client 'client-with-invalid-secret': option 'client_secret' is required to be empty when option 'public' is true")
	assert.EqualError(t, validator.Errors()[1], "identity_providers: oidc: clients: client 'client-with-bad-redirect-uri': option 'redirect_uris' has the redirect uri 'urn:ietf:wg:oauth:2.0:oob' when option 'public' is false but this is invalid as this uri is not valid for the openid connect confidential client type")
	assert.EqualError(t, validator.Warnings()[0], "identity_providers: oidc: clients: client 'client-with-invalid-secret': option 'client_secret' is plaintext but for clients not using any endpoint authentication method 'client_secret_jwt' it should be a hashed value as plaintext values are deprecated with the exception of 'client_secret_jwt' and will be removed in the near future")
}

func TestValidateIdentityProvidersShouldNotRaiseErrorsOnValidClientOptions(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "hmac1",
				IssuerPrivateKey: keyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  "installed-app-client",
						Public:              true,
						AuthorizationPolicy: "two_factor",
						RedirectURIs: []string{
							oidc.RedirectURISpecialOAuth2InstalledApp,
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
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)
}

func TestValidateIdentityProvidersShouldRaiseWarningOnPlainTextClients(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "hmac1",
				IssuerPrivateKey: keyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
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
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	assert.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 1)

	assert.EqualError(t, validator.Warnings()[0], "identity_providers: oidc: clients: client 'client-with-invalid-secret_standard': option 'client_secret' is plaintext but for clients not using any endpoint authentication method 'client_secret_jwt' it should be a hashed value as plaintext values are deprecated with the exception of 'client_secret_jwt' and will be removed in the near future")
}

// All valid schemes are supported as defined in https://datatracker.ietf.org/doc/html/rfc8252#section-7.1
func TestValidateOIDCClientRedirectURIsSupportingPrivateUseURISchemes(t *testing.T) {
	have := &schema.IdentityProvidersOpenIDConnect{
		Clients: []schema.IdentityProvidersOpenIDConnectClient{
			{
				ID: "owncloud",
				RedirectURIs: []string{
					"https://www.mywebsite.com",
					"http://www.mywebsite.com",
					"oc://ios.owncloud.com",
					// example given in the RFC https://datatracker.ietf.org/doc/html/rfc8252#section-7.1
					"com.example.app:/oauth2redirect/example-provider",
					oidc.RedirectURISpecialOAuth2InstalledApp,
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

	const (
		abcabc123 = "abcabc123"
		abc123abc = "abc123abc"
	)

	testCasses := []struct {
		name     string
		setup    func(have *schema.IdentityProvidersOpenIDConnect)
		validate func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect)
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
			"ShouldNotIncludeOldMinimalScope",
			nil,
			nil,
			tcv{
				[]string{oidc.ScopeEmail},
				nil,
				nil,
				nil,
			},
			tcv{
				[]string{oidc.ScopeEmail},
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
			"ShouldRaiseWarningOnInvalidScopes",
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
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'scopes' only expects the values 'openid', 'email', 'profile', 'address', 'phone', 'groups', 'offline_access', 'offline', or 'authelia.bearer.authz' but the unknown values 'group' are present and should generally only be used if a particular client requires a scope outside of our standard scopes",
			},
			nil,
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
				"identity_providers: oidc: clients: client 'test': option 'response_modes' must only have the values 'form_post', 'query', 'fragment', 'jwt', 'form_post.jwt', 'query.jwt', or 'fragment.jwt' but the values 'not_valid' are present",
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
				"identity_providers: oidc: clients: client 'test': option 'grant_types' must only have the values 'authorization_code', 'implicit', 'client_credentials', 'refresh_token', or 'urn:ietf:params:oauth:grant-type:device_code' but the values 'invalid' are present",
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
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
				"identity_providers: oidc: clients: client 'test': option 'scopes' has the values 'openid', 'offline', and 'offline_access' however when utilizing the 'client_credentials' value for the 'grant_types' the values 'openid', 'offline', or 'offline_access' are not allowed",
			},
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].RedirectURIs = []string{
					"https://google.com",
				}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, schema.IdentityProvidersOpenIDConnectClientURIs([]string{"https://google.com"}), have.Clients[0].RedirectURIs)
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].Public = true
				have.Clients[0].Secret = nil
				have.Clients[0].RedirectURIs = []string{
					oidc.RedirectURISpecialOAuth2InstalledApp,
				}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, schema.IdentityProvidersOpenIDConnectClientURIs([]string{oidc.RedirectURISpecialOAuth2InstalledApp}), have.Clients[0].RedirectURIs)
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].RedirectURIs = []string{
					"urn:ietf:wg:oauth:2.0:oob",
				}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, schema.IdentityProvidersOpenIDConnectClientURIs([]string{oidc.RedirectURISpecialOAuth2InstalledApp}), have.Clients[0].RedirectURIs)
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].RedirectURIs = []string{
					"http://abc@%two",
				}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, schema.IdentityProvidersOpenIDConnectClientURIs([]string{"http://abc@%two"}), have.Clients[0].RedirectURIs)
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].RedirectURIs = []string{
					"google.com",
				}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, schema.IdentityProvidersOpenIDConnectClientURIs([]string{"google.com"}), have.Clients[0].RedirectURIs)
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].RedirectURIs = []string{
					"https://google.com",
					"https://google.com",
				}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, schema.IdentityProvidersOpenIDConnectClientURIs([]string{"https://google.com", "https://google.com"}), have.Clients[0].RedirectURIs)
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
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			nil,
		},
		{
			"ShouldNotOverrideValidClientAuthMethod",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretPost
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.GrantTypeClientCredentials
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretBasic
				have.Clients[0].Public = true
				have.Clients[0].Secret = nil
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodNone
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
				"identity_providers: oidc: clients: client 'test': option 'client_secret' is required to be empty when option 'token_endpoint_auth_method' is configured as 'none'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidClientAuthMethodForConfidentialClientTypeHybridFlow",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodNone
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
				"identity_providers: oidc: clients: client 'test': option 'client_secret' is required to be empty when option 'token_endpoint_auth_method' is configured as 'none'",
			},
		},
		{
			"ShouldSetDefaultResponseSigningAlg",
			nil,
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgNone, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgNone, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].IDTokenSignedResponseAlg)
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
			"ShouldConfigureKID",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.JSONWebKeys = []schema.JWK{
					{
						KeyID:     "idES512",
						Algorithm: oidc.SigningAlgECDSAUsingP521AndSHA512,
					},
					{
						KeyID:     "idRS384",
						Algorithm: oidc.SigningAlgRSAUsingSHA384,
					},
					{
						KeyID:     "idRS512",
						Algorithm: oidc.SigningAlgRSAUsingSHA512,
					},
				}

				have.DiscoverySignedResponseAlg = oidc.SigningAlgRSAUsingSHA384

				have.Clients[0].IntrospectionSignedResponseAlg = oidc.SigningAlgRSAUsingSHA384
				have.Clients[0].UserinfoSignedResponseAlg = oidc.SigningAlgRSAUsingSHA512
				have.Clients[0].IDTokenSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512
				have.Clients[0].AccessTokenSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512
				have.Clients[0].AuthorizationSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512

				have.Discovery.ResponseObjectSigningAlgs = []string{oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAUsingSHA512, oidc.SigningAlgECDSAUsingP521AndSHA512}
				have.Discovery.ResponseObjectSigningKeyIDs = []string{id + oidc.SigningAlgRSAUsingSHA384, id + oidc.SigningAlgRSAUsingSHA512, id + oidc.SigningAlgECDSAUsingP521AndSHA512}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA384, have.DiscoverySignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA384, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA512, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AccessTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AuthorizationSignedResponseAlg)

				assert.Equal(t, id+oidc.SigningAlgRSAUsingSHA384, have.DiscoverySignedResponseKeyID)
				assert.Equal(t, id+oidc.SigningAlgRSAUsingSHA384, have.Clients[0].IntrospectionSignedResponseKeyID)
				assert.Equal(t, id+oidc.SigningAlgRSAUsingSHA512, have.Clients[0].UserinfoSignedResponseKeyID)
				assert.Equal(t, id+oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].IDTokenSignedResponseKeyID)
				assert.Equal(t, id+oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AccessTokenSignedResponseKeyID)
				assert.Equal(t, id+oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AuthorizationSignedResponseKeyID)
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
			"ShouldConfigureAlg",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.JSONWebKeys = []schema.JWK{
					{
						KeyID:     "idES512",
						Algorithm: oidc.SigningAlgECDSAUsingP521AndSHA512,
					},
					{
						KeyID:     "idRS384",
						Algorithm: oidc.SigningAlgRSAUsingSHA384,
					},
					{
						KeyID:     "idRS512",
						Algorithm: oidc.SigningAlgRSAUsingSHA512,
					},
				}

				have.Clients[0].IntrospectionSignedResponseKeyID = id + oidc.SigningAlgRSAUsingSHA384
				have.Clients[0].UserinfoSignedResponseKeyID = id + oidc.SigningAlgRSAUsingSHA512
				have.Clients[0].IDTokenSignedResponseKeyID = id + oidc.SigningAlgECDSAUsingP521AndSHA512
				have.Clients[0].AccessTokenSignedResponseKeyID = id + oidc.SigningAlgECDSAUsingP521AndSHA512
				have.Clients[0].AuthorizationSignedResponseKeyID = id + oidc.SigningAlgECDSAUsingP521AndSHA512

				have.Discovery.ResponseObjectSigningAlgs = []string{oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAUsingSHA512, oidc.SigningAlgECDSAUsingP521AndSHA512}
				have.Discovery.ResponseObjectSigningKeyIDs = []string{id + oidc.SigningAlgRSAUsingSHA384, id + oidc.SigningAlgRSAUsingSHA512, id + oidc.SigningAlgECDSAUsingP521AndSHA512}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA384, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA512, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AccessTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AuthorizationSignedResponseAlg)

				assert.Equal(t, id+oidc.SigningAlgRSAUsingSHA384, have.Clients[0].IntrospectionSignedResponseKeyID)
				assert.Equal(t, id+oidc.SigningAlgRSAUsingSHA512, have.Clients[0].UserinfoSignedResponseKeyID)
				assert.Equal(t, id+oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].IDTokenSignedResponseKeyID)
				assert.Equal(t, id+oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AccessTokenSignedResponseKeyID)
				assert.Equal(t, id+oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AuthorizationSignedResponseKeyID)
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
			"ShouldSetDefaultAlgKID",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.JSONWebKeys = []schema.JWK{
					{
						KeyID:     "idRS256",
						Algorithm: oidc.SigningAlgRSAUsingSHA256,
					},
					{
						KeyID:     "idES512",
						Algorithm: oidc.SigningAlgECDSAUsingP521AndSHA512,
					},
					{
						KeyID:     "idRS384",
						Algorithm: oidc.SigningAlgRSAUsingSHA384,
					},
					{
						KeyID:     "idRS512",
						Algorithm: oidc.SigningAlgRSAUsingSHA512,
					},
				}

				have.Discovery.ResponseObjectSigningAlgs = []string{oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAUsingSHA512, oidc.SigningAlgECDSAUsingP521AndSHA512}
				have.Discovery.ResponseObjectSigningKeyIDs = []string{id + oidc.SigningAlgRSAUsingSHA256, id + oidc.SigningAlgRSAUsingSHA384, id + oidc.SigningAlgRSAUsingSHA512, id + oidc.SigningAlgECDSAUsingP521AndSHA512}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgNone, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgNone, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgNone, have.Clients[0].AccessTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].AuthorizationSignedResponseAlg)

				assert.Equal(t, "", have.Clients[0].IntrospectionSignedResponseKeyID)
				assert.Equal(t, "", have.Clients[0].UserinfoSignedResponseKeyID)
				assert.Equal(t, "idRS256", have.Clients[0].IDTokenSignedResponseKeyID)
				assert.Equal(t, "", have.Clients[0].AccessTokenSignedResponseKeyID)
				assert.Equal(t, "idRS256", have.Clients[0].AuthorizationSignedResponseKeyID)
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
			"ShouldNotOverrideResponseSigningAlg",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].IntrospectionSignedResponseAlg = oidc.SigningAlgRSAUsingSHA384
				have.Clients[0].UserinfoSignedResponseAlg = oidc.SigningAlgRSAUsingSHA512
				have.Clients[0].IDTokenSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512
				have.Clients[0].AccessTokenSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512
				have.Clients[0].AuthorizationSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512

				have.Discovery.ResponseObjectSigningAlgs = []string{oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAUsingSHA512, oidc.SigningAlgECDSAUsingP521AndSHA512}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA384, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA512, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AccessTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AuthorizationSignedResponseAlg)
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
			"ShouldConfigureKeyIDFromAlg",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].IntrospectionSignedResponseAlg = oidc.SigningAlgRSAUsingSHA384
				have.Clients[0].UserinfoSignedResponseAlg = oidc.SigningAlgRSAUsingSHA512
				have.Clients[0].IDTokenSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512
				have.Clients[0].AccessTokenSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512
				have.Clients[0].AuthorizationSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512

				have.Discovery.ResponseObjectSigningAlgs = []string{oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAUsingSHA512, oidc.SigningAlgECDSAUsingP521AndSHA512}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA384, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA512, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AccessTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AuthorizationSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AuthorizationSignedResponseAlg)
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
			"ShouldConfigureKeyIDFromAlg",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].IntrospectionSignedResponseAlg = oidc.SigningAlgRSAUsingSHA384
				have.Clients[0].UserinfoSignedResponseAlg = oidc.SigningAlgRSAUsingSHA512
				have.Clients[0].IDTokenSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512
				have.Clients[0].AuthorizationSignedResponseAlg = oidc.SigningAlgECDSAUsingP521AndSHA512

				have.Discovery.ResponseObjectSigningAlgs = []string{oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAUsingSHA512, oidc.SigningAlgECDSAUsingP521AndSHA512}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA384, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA512, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, have.Clients[0].AuthorizationSignedResponseAlg)
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
			"ShouldRaiseErrorOnInvalidSigningAlgForIDTokensOrAuthSignedAlg",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].IntrospectionSignedResponseAlg = oidc.SigningAlgRSAUsingSHA384
				have.Clients[0].UserinfoSignedResponseAlg = oidc.SigningAlgRSAUsingSHA512
				have.Clients[0].IDTokenSignedResponseAlg = oidc.SigningAlgNone
				have.Clients[0].AuthorizationSignedResponseAlg = oidc.SigningAlgNone

				have.Discovery.ResponseObjectSigningAlgs = []string{oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAUsingSHA512, oidc.SigningAlgECDSAUsingP521AndSHA512}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA384, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA512, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgNone, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, oidc.SigningAlgNone, have.Clients[0].AuthorizationSignedResponseAlg)
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
				"identity_providers: oidc: clients: client 'test': option 'authorization_signed_response_alg' must be one of 'RS384', 'RS512', or 'ES512' but it's configured as 'none'",
				"identity_providers: oidc: clients: client 'test': option 'id_token_signed_response_alg' must be one of 'RS384', 'RS512', or 'ES512' but it's configured as 'none'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidLifespanNone",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].Lifespan = rs256
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, rs256, have.Clients[0].Lifespan)
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
				"identity_providers: oidc: clients: client 'test': option 'lifespan' must not be configured when no custom lifespans are configured but it's configured as 'rs256'",
			},
		},
		{
			name: "ShouldRaiseErrorOnInvalidLifespan",
			setup: func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].Lifespan = rs256
				have.Discovery.Lifespans = []string{"example"}
			},
			validate: func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, rs256, have.Clients[0].Lifespan)
			},
			have: tcv{
				nil,
				nil,
				nil,
				nil,
			},
			expected: tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			errs: []string{
				"identity_providers: oidc: clients: client 'test': option 'lifespan' must be one of 'example' but it's configured as 'rs256'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidResponseSigningAlg",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.DiscoverySignedResponseAlg = rs256
				have.Clients[0].AuthorizationSignedResponseAlg = rs256
				have.Clients[0].IntrospectionSignedResponseAlg = rs256
				have.Clients[0].IDTokenSignedResponseAlg = rs256
				have.Clients[0].UserinfoSignedResponseAlg = rs256
				have.Clients[0].AccessTokenSignedResponseAlg = rs256
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, rs256, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, rs256, have.Clients[0].AuthorizationSignedResponseAlg)
				assert.Equal(t, rs256, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, rs256, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, rs256, have.Clients[0].AccessTokenSignedResponseAlg)
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
				"identity_providers: oidc: option 'discovery_signed_response_alg' must be one of 'RS256' or 'none' but it's configured as 'rs256'",
				"identity_providers: oidc: clients: client 'test': option 'authorization_signed_response_alg' must be one of 'RS256' or 'none' but it's configured as 'rs256'",
				"identity_providers: oidc: clients: client 'test': option 'id_token_signed_response_alg' must be one of 'RS256' but it's configured as 'rs256'",
				"identity_providers: oidc: clients: client 'test': option 'access_token_signed_response_alg' must be one of 'RS256' or 'none' but it's configured as 'rs256'",
				"identity_providers: oidc: clients: client 'test': option 'userinfo_signed_response_alg' must be one of 'RS256' or 'none' but it's configured as 'rs256'",
				"identity_providers: oidc: clients: client 'test': option 'introspection_signed_response_alg' must be one of 'RS256' or 'none' but it's configured as 'rs256'",
			},
		},
		{
			"ShouldHandleBearerErrorsMisconfiguredPublicClientType",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0] = schema.IdentityProvidersOpenIDConnectClient{
					ID:                                 "abc",
					Secret:                             nil,
					Public:                             true,
					RedirectURIs:                       []string{"http://localhost"},
					Audience:                           nil,
					Scopes:                             []string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOpenID},
					GrantTypes:                         []string{oidc.GrantTypeImplicit},
					ResponseTypes:                      []string{oidc.ResponseTypeImplicitFlowBoth},
					ResponseModes:                      []string{oidc.ResponseModeQuery},
					AuthorizationPolicy:                "",
					RequestedAudienceMode:              "",
					ConsentMode:                        oidc.ClientConsentModeImplicit.String(),
					RequirePushedAuthorizationRequests: false,
					RequirePKCE:                        false,
					PKCEChallengeMethod:                "",
					TokenEndpointAuthMethod:            "",
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
				[]string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOpenID},
				[]string{oidc.ResponseTypeImplicitFlowBoth},
				[]string{oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeImplicit},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'abc': option 'scopes' must only have the values 'offline_access', 'offline', and 'authelia.bearer.authz' when configured with scope 'authelia.bearer.authz' but the values 'authelia.bearer.authz' and 'openid' are present",
				"identity_providers: oidc: clients: client 'abc': option 'grant_types' must only have the values 'authorization_code', 'refresh_token', and 'client_credentials' when configured with scope 'authelia.bearer.authz' but the values 'implicit' are present",
				"identity_providers: oidc: clients: client 'abc': option 'audience' must be configured when configured with scope 'authelia.bearer.authz' but it's absent",
				"identity_providers: oidc: clients: client 'abc': option 'require_pushed_authorization_requests' must be configured as 'true' when configured with scope 'authelia.bearer.authz' but it's configured as 'false'",
				"identity_providers: oidc: clients: client 'abc': option 'require_pkce' must be configured as 'true' when configured with scope 'authelia.bearer.authz' but it's configured as 'false'",
				"identity_providers: oidc: clients: client 'abc': option 'consent_mode' must be configured as 'explicit' when configured with scope 'authelia.bearer.authz' but it's configured as 'implicit'",
				"identity_providers: oidc: clients: client 'abc': option 'response_types' must only have the values 'code' when configured with scope 'authelia.bearer.authz' but the values 'id_token token' are present",
				"identity_providers: oidc: clients: client 'abc': option 'response_modes' must only have the values 'form_post' and 'form_post.jwt' when configured with scope 'authelia.bearer.authz' but the values 'query' are present",
				"identity_providers: oidc: clients: client 'abc': option 'token_endpoint_auth_method' must be configured as 'none' when configured with scope 'authelia.bearer.authz' and the 'public' client type but it's configured as ''",
			},
		},
		{
			"ShouldHandleBearerErrorsMisconfiguredConfidentialClientType",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0] = schema.IdentityProvidersOpenIDConnectClient{
					ID:                                 "abc",
					Secret:                             tOpenIDConnectPBKDF2ClientSecret,
					Public:                             false,
					RedirectURIs:                       []string{"http://localhost"},
					Audience:                           nil,
					Scopes:                             []string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOpenID},
					GrantTypes:                         []string{oidc.GrantTypeImplicit},
					ResponseTypes:                      []string{oidc.ResponseTypeImplicitFlowBoth},
					ResponseModes:                      []string{oidc.ResponseModeQuery},
					AuthorizationPolicy:                "",
					RequestedAudienceMode:              "",
					ConsentMode:                        oidc.ClientConsentModeImplicit.String(),
					RequirePushedAuthorizationRequests: false,
					RequirePKCE:                        true,
					PKCEChallengeMethod:                "",
					TokenEndpointAuthMethod:            "",
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
				[]string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOpenID},
				[]string{oidc.ResponseTypeImplicitFlowBoth},
				[]string{oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeImplicit},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'abc': option 'scopes' must only have the values 'offline_access', 'offline', and 'authelia.bearer.authz' when configured with scope 'authelia.bearer.authz' but the values 'authelia.bearer.authz' and 'openid' are present",
				"identity_providers: oidc: clients: client 'abc': option 'grant_types' must only have the values 'authorization_code', 'refresh_token', and 'client_credentials' when configured with scope 'authelia.bearer.authz' but the values 'implicit' are present",
				"identity_providers: oidc: clients: client 'abc': option 'audience' must be configured when configured with scope 'authelia.bearer.authz' but it's absent",
				"identity_providers: oidc: clients: client 'abc': option 'require_pushed_authorization_requests' must be configured as 'true' when configured with scope 'authelia.bearer.authz' but it's configured as 'false'",
				"identity_providers: oidc: clients: client 'abc': option 'pkce_challenge_method' must be configured as 'S256' when configured with scope 'authelia.bearer.authz' but it's configured as ''",
				"identity_providers: oidc: clients: client 'abc': option 'consent_mode' must be configured as 'explicit' when configured with scope 'authelia.bearer.authz' but it's configured as 'implicit'",
				"identity_providers: oidc: clients: client 'abc': option 'response_types' must only have the values 'code' when configured with scope 'authelia.bearer.authz' but the values 'id_token token' are present",
				"identity_providers: oidc: clients: client 'abc': option 'response_modes' must only have the values 'form_post' and 'form_post.jwt' when configured with scope 'authelia.bearer.authz' but the values 'query' are present",
				"identity_providers: oidc: clients: client 'abc': option 'token_endpoint_auth_method' must be configured as 'client_secret_basic', 'client_secret_jwt', or 'private_key_jwt' when configured with scope 'authelia.bearer.authz' and the 'confidential' client type but it's configured as ''",
			},
		},
		{
			"ShouldHandleBearerErrorsMisconfiguredConfidentialClientTypeClientCredentials",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0] = schema.IdentityProvidersOpenIDConnectClient{
					ID:                                 "abc",
					Secret:                             tOpenIDConnectPBKDF2ClientSecret,
					Public:                             false,
					RedirectURIs:                       []string{"http://localhost"},
					Audience:                           nil,
					Scopes:                             []string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOpenID},
					GrantTypes:                         []string{oidc.GrantTypeClientCredentials},
					ResponseTypes:                      nil,
					ResponseModes:                      nil,
					AuthorizationPolicy:                "",
					RequestedAudienceMode:              "",
					ConsentMode:                        oidc.ClientConsentModeImplicit.String(),
					RequirePushedAuthorizationRequests: false,
					RequirePKCE:                        true,
					PKCEChallengeMethod:                "",
					TokenEndpointAuthMethod:            "",
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
				[]string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOpenID},
				[]string(nil),
				[]string(nil),
				[]string{oidc.GrantTypeClientCredentials},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'abc': option 'scopes' must only have the values 'offline_access', 'offline', and 'authelia.bearer.authz' when configured with scope 'authelia.bearer.authz' but the values 'authelia.bearer.authz' and 'openid' are present",
				"identity_providers: oidc: clients: client 'abc': option 'audience' must be configured when configured with scope 'authelia.bearer.authz' but it's absent",
				"identity_providers: oidc: clients: client 'abc': option 'token_endpoint_auth_method' must be configured as 'client_secret_basic', 'client_secret_jwt', or 'private_key_jwt' when configured with scope 'authelia.bearer.authz' and the 'confidential' client type but it's configured as ''",
				"identity_providers: oidc: clients: client 'abc': option 'scopes' has the values 'authelia.bearer.authz' and 'openid' however when utilizing the 'client_credentials' value for the 'grant_types' the values 'openid' are not allowed",
			},
		},
		{
			"ShouldHandleBearerErrorsNotExplicit",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0] = schema.IdentityProvidersOpenIDConnectClient{
					ID:                                 "abc",
					Secret:                             nil,
					Public:                             true,
					RedirectURIs:                       []string{"http://localhost"},
					Audience:                           nil,
					Scopes:                             []string{oidc.ScopeAutheliaBearerAuthz},
					RequirePushedAuthorizationRequests: false,
					RequirePKCE:                        false,
					PKCEChallengeMethod:                "",
					TokenEndpointAuthMethod:            "",
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
				[]string{oidc.ScopeAutheliaBearerAuthz},
				nil,
				nil,
				nil,
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'abc': option 'grant_types' must only have the values 'authorization_code', 'refresh_token', and 'client_credentials' when configured with scope 'authelia.bearer.authz' but it's not configured",
				"identity_providers: oidc: clients: client 'abc': option 'audience' must be configured when configured with scope 'authelia.bearer.authz' but it's absent",
				"identity_providers: oidc: clients: client 'abc': option 'require_pushed_authorization_requests' must be configured as 'true' when configured with scope 'authelia.bearer.authz' but it's configured as 'false'",
				"identity_providers: oidc: clients: client 'abc': option 'require_pkce' must be configured as 'true' when configured with scope 'authelia.bearer.authz' but it's configured as 'false'",
				"identity_providers: oidc: clients: client 'abc': option 'consent_mode' must be configured as 'explicit' when configured with scope 'authelia.bearer.authz' but it's configured as ''",
				"identity_providers: oidc: clients: client 'abc': option 'response_types' must only have the values 'code' when configured with scope 'authelia.bearer.authz' but it's not configured",
				"identity_providers: oidc: clients: client 'abc': option 'response_modes' must only have the values 'form_post' and 'form_post.jwt' when configured with scope 'authelia.bearer.authz' but it's not configured",
				"identity_providers: oidc: clients: client 'abc': option 'token_endpoint_auth_method' must be configured as 'none' when configured with scope 'authelia.bearer.authz' and the 'public' client type but it's configured as ''",
			},
		},
		{
			"ShouldHandleBearerValidConfidentialClientType",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0] = schema.IdentityProvidersOpenIDConnectClient{
					ID:                                 "abc",
					Secret:                             tOpenIDConnectPBKDF2ClientSecret,
					Public:                             false,
					RedirectURIs:                       []string{"http://localhost"},
					Audience:                           []string{"https://app.example.com"},
					Scopes:                             []string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess},
					GrantTypes:                         []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken},
					ResponseTypes:                      []string{oidc.ResponseTypeAuthorizationCodeFlow},
					ResponseModes:                      []string{oidc.ResponseModeFormPost, oidc.ResponseModeFormPostJWT},
					AuthorizationPolicy:                "",
					RequestedAudienceMode:              "",
					ConsentMode:                        oidc.ClientConsentModeExplicit.String(),
					RequirePushedAuthorizationRequests: true,
					RequirePKCE:                        true,
					PKCEChallengeMethod:                oidc.PKCEChallengeMethodSHA256,
					TokenEndpointAuthMethod:            oidc.ClientAuthMethodClientSecretBasic,
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
				[]string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeFormPostJWT},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken},
			},
			nil,
			nil,
		},
		{
			"ShouldHandleBearerValidPublicClientType",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0] = schema.IdentityProvidersOpenIDConnectClient{
					ID:                                 "abc",
					Secret:                             nil,
					Public:                             true,
					RedirectURIs:                       []string{"http://localhost"},
					Audience:                           []string{"https://app.example.com"},
					Scopes:                             []string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess},
					GrantTypes:                         []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken},
					ResponseTypes:                      []string{oidc.ResponseTypeAuthorizationCodeFlow},
					ResponseModes:                      []string{oidc.ResponseModeFormPost},
					AuthorizationPolicy:                "",
					RequestedAudienceMode:              "",
					ConsentMode:                        oidc.ClientConsentModeExplicit.String(),
					RequirePushedAuthorizationRequests: true,
					RequirePKCE:                        true,
					PKCEChallengeMethod:                oidc.PKCEChallengeMethodSHA256,
					TokenEndpointAuthMethod:            oidc.ClientAuthMethodNone,
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
				[]string{oidc.ScopeAutheliaBearerAuthz, oidc.ScopeOfflineAccess},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken},
			},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultConsentMode",
			nil,
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].ConsentMode = auto
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				d := time.Minute

				have.Clients[0].ConsentMode = ""
				have.Clients[0].ConsentPreConfiguredDuration = &d
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				d := time.Minute

				have.Clients[0].ConsentMode = auto
				have.Clients[0].ConsentPreConfiguredDuration = &d
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].ConsentMode = "implicit"
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].ConsentMode = "pre-configured"
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
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
				"identity_providers: oidc: clients: client 'test': option 'jwks_uri' or 'jwks' is required with 'token_endpoint_auth_method' set to 'private_key_jwt'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' is required to be empty when option 'token_endpoint_auth_method' is configured as 'private_key_jwt'",
			},
		},
		{
			"ShouldRaiseErrorOnTokenEndpointClientAuthMethodPrivateKeyJWTMustSetKnownAlg",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodPrivateKeyJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = "nope"
				have.Clients[0].Secret = nil
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
				"identity_providers: oidc: clients: client 'test': option 'jwks_uri' or 'jwks' is required with 'token_endpoint_auth_method' set to 'private_key_jwt'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' is required",
			},
		},
		{
			"ShouldRaiseErrorOnTokenEndpointClientAuthMethodPrivateKeyJWTMustSetRegisteredAlg",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodPrivateKeyJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = oidc.SigningAlgECDSAUsingP384AndSHA384
				have.Clients[0].Secret = nil
				have.Clients[0].JSONWebKeys = []schema.JWK{
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
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_signing_alg' must be one of the registered public key algorithm values 'RS256' when option 'token_endpoint_auth_method' is configured to 'private_key_jwt'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' is required",
			},
		},
		{
			"ShouldRaiseErrorOnTokenEndpointClientAuthMethodPrivateKeyJWTMustHavePublicKeys",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodPrivateKeyJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = oidc.SigningAlgECDSAUsingP384AndSHA384
				have.Clients[0].Secret = nil
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
				"identity_providers: oidc: clients: client 'test': option 'jwks_uri' or 'jwks' is required with 'token_endpoint_auth_method' set to 'private_key_jwt'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' is required",
			},
		},
		{
			"ShouldRaiseErrorOnIncorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.IdentityProvidersOpenIDConnect) {
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
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'token_endpoint_auth_method' with a value of 'client_secret_jwt'",
			},
		},
		{
			"ShouldNotRaiseWarningOrErrorOnCorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].RevocationEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].IntrospectionEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].PushedAuthorizationRequestEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
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
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'token_endpoint_auth_method' with a value of 'client_secret_jwt'",
			},
		},
		{
			"ShouldNotRaiseWarningOrErrorOnCorrectlyConfiguredTokenEndpointClientAuthMethodClientSecretJWT",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].RevocationEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].IntrospectionEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].PushedAuthorizationRequestEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].AuthorizationSignedResponseKeyID = abcabc123
				have.Clients[0].IDTokenSignedResponseKeyID = abcabc123
				have.Clients[0].UserinfoSignedResponseKeyID = abc123abc
				have.Clients[0].AccessTokenSignedResponseKeyID = abc123abc
				have.Clients[0].IntrospectionSignedResponseKeyID = abc123abc
				have.Discovery.ResponseObjectSigningKeyIDs = []string{abcabc123, abc123abc}
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.DiscoverySignedResponseKeyID = "ij"
				have.Clients[0].AuthorizationSignedResponseKeyID = "01"
				have.Clients[0].IDTokenSignedResponseKeyID = "ab"
				have.Clients[0].UserinfoSignedResponseKeyID = "cd"
				have.Clients[0].IntrospectionSignedResponseKeyID = "ef"
				have.Clients[0].AccessTokenSignedResponseKeyID = "gh"
				have.Discovery.ResponseObjectSigningKeyIDs = []string{"abc123xyz"}
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, "ij", have.DiscoverySignedResponseKeyID)
				assert.Equal(t, "ef", have.Clients[0].IntrospectionSignedResponseKeyID)
				assert.Equal(t, "01", have.Clients[0].AuthorizationSignedResponseKeyID)
				assert.Equal(t, "ab", have.Clients[0].IDTokenSignedResponseKeyID)
				assert.Equal(t, "cd", have.Clients[0].UserinfoSignedResponseKeyID)
				assert.Equal(t, "gh", have.Clients[0].AccessTokenSignedResponseKeyID)
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
				"identity_providers: oidc: option 'discovery_signed_response_key_id' must be one of 'abc123xyz' but it's configured as 'ij'",
				"identity_providers: oidc: clients: client 'test': option 'authorization_signed_response_key_id' must be one of 'abc123xyz' but it's configured as '01'",
				"identity_providers: oidc: clients: client 'test': option 'id_token_signed_response_key_id' must be one of 'abc123xyz' but it's configured as 'ab'",
				"identity_providers: oidc: clients: client 'test': option 'access_token_signed_response_key_id' must be one of 'abc123xyz' but it's configured as 'gh'",
				"identity_providers: oidc: clients: client 'test': option 'userinfo_signed_response_key_id' must be one of 'abc123xyz' but it's configured as 'cd'",
				"identity_providers: oidc: clients: client 'test': option 'introspection_signed_response_key_id' must be one of 'abc123xyz' but it's configured as 'ef'",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidEncryptionAlgs",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].AuthorizationEncryptedResponseAlg = "01alg"
				have.Clients[0].IDTokenEncryptedResponseAlg = "abalg"
				have.Clients[0].UserinfoEncryptedResponseAlg = "cdalg"
				have.Clients[0].IntrospectionEncryptedResponseAlg = "efalg"
				have.Clients[0].AccessTokenEncryptedResponseAlg = "ghalg"

				have.Clients[0].AuthorizationEncryptedResponseEnc = "01enc"
				have.Clients[0].IDTokenEncryptedResponseEnc = "abenc"
				have.Clients[0].UserinfoEncryptedResponseEnc = "cdenc"
				have.Clients[0].IntrospectionEncryptedResponseEnc = "efenc"
				have.Clients[0].AccessTokenEncryptedResponseEnc = "ghenc"
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, "01alg", have.Clients[0].AuthorizationEncryptedResponseAlg)
				assert.Equal(t, "abalg", have.Clients[0].IDTokenEncryptedResponseAlg)
				assert.Equal(t, "cdalg", have.Clients[0].UserinfoEncryptedResponseAlg)
				assert.Equal(t, "efalg", have.Clients[0].IntrospectionEncryptedResponseAlg)
				assert.Equal(t, "ghalg", have.Clients[0].AccessTokenEncryptedResponseAlg)

				assert.Equal(t, "01enc", have.Clients[0].AuthorizationEncryptedResponseEnc)
				assert.Equal(t, "abenc", have.Clients[0].IDTokenEncryptedResponseEnc)
				assert.Equal(t, "cdenc", have.Clients[0].UserinfoEncryptedResponseEnc)
				assert.Equal(t, "efenc", have.Clients[0].IntrospectionEncryptedResponseEnc)
				assert.Equal(t, "ghenc", have.Clients[0].AccessTokenEncryptedResponseEnc)
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
				"identity_providers: oidc: clients: client 'test': option 'authorization_encrypted_response_alg' must be one of 'none', 'RSA1_5', 'RSA-OAEP', 'RSA-OAEP-256', 'ECDH-ES', 'ECDH-ES+A128KW', 'ECDH-ES+A192KW', 'ECDH-ES+A256KW', 'A128KW', 'A192KW', 'A256KW', 'A128GCMKW', 'A192GCMKW', 'A256GCMKW', 'PBES2-HS256+A128KW', 'PBES2-HS384+A192KW', or 'PBES2-HS512+A256KW' but it's configured as '01alg'",
				"identity_providers: oidc: clients: client 'test': option 'authorization_encrypted_response_enc' must be one of 'A128GCM', 'A192GCM', 'A256GCM', 'A128CBC-HS256', 'A192CBC-HS384', or 'A256CBC-HS512' but it's configured as '01enc'",
				"identity_providers: oidc: clients: client 'test': option 'id_token_encrypted_response_alg' must be one of 'none', 'RSA1_5', 'RSA-OAEP', 'RSA-OAEP-256', 'ECDH-ES', 'ECDH-ES+A128KW', 'ECDH-ES+A192KW', 'ECDH-ES+A256KW', 'A128KW', 'A192KW', 'A256KW', 'A128GCMKW', 'A192GCMKW', 'A256GCMKW', 'PBES2-HS256+A128KW', 'PBES2-HS384+A192KW', or 'PBES2-HS512+A256KW' but it's configured as 'abalg'",
				"identity_providers: oidc: clients: client 'test': option 'id_token_encrypted_response_enc' must be one of 'A128GCM', 'A192GCM', 'A256GCM', 'A128CBC-HS256', 'A192CBC-HS384', or 'A256CBC-HS512' but it's configured as 'abenc'",
				"identity_providers: oidc: clients: client 'test': option 'access_token_encrypted_response_alg' must be one of 'none', 'RSA1_5', 'RSA-OAEP', 'RSA-OAEP-256', 'ECDH-ES', 'ECDH-ES+A128KW', 'ECDH-ES+A192KW', 'ECDH-ES+A256KW', 'A128KW', 'A192KW', 'A256KW', 'A128GCMKW', 'A192GCMKW', 'A256GCMKW', 'PBES2-HS256+A128KW', 'PBES2-HS384+A192KW', or 'PBES2-HS512+A256KW' but it's configured as 'ghalg'",
				"identity_providers: oidc: clients: client 'test': option 'access_token_encrypted_response_enc' must be one of 'A128GCM', 'A192GCM', 'A256GCM', 'A128CBC-HS256', 'A192CBC-HS384', or 'A256CBC-HS512' but it's configured as 'ghenc'",
				"identity_providers: oidc: clients: client 'test': option 'userinfo_encrypted_response_alg' must be one of 'none', 'RSA1_5', 'RSA-OAEP', 'RSA-OAEP-256', 'ECDH-ES', 'ECDH-ES+A128KW', 'ECDH-ES+A192KW', 'ECDH-ES+A256KW', 'A128KW', 'A192KW', 'A256KW', 'A128GCMKW', 'A192GCMKW', 'A256GCMKW', 'PBES2-HS256+A128KW', 'PBES2-HS384+A192KW', or 'PBES2-HS512+A256KW' but it's configured as 'cdalg'",
				"identity_providers: oidc: clients: client 'test': option 'userinfo_encrypted_response_enc' must be one of 'A128GCM', 'A192GCM', 'A256GCM', 'A128CBC-HS256', 'A192CBC-HS384', or 'A256CBC-HS512' but it's configured as 'cdenc'",
				"identity_providers: oidc: clients: client 'test': option 'introspection_encrypted_response_alg' must be one of 'none', 'RSA1_5', 'RSA-OAEP', 'RSA-OAEP-256', 'ECDH-ES', 'ECDH-ES+A128KW', 'ECDH-ES+A192KW', 'ECDH-ES+A256KW', 'A128KW', 'A192KW', 'A256KW', 'A128GCMKW', 'A192GCMKW', 'A256GCMKW', 'PBES2-HS256+A128KW', 'PBES2-HS384+A192KW', or 'PBES2-HS512+A256KW' but it's configured as 'efalg'",
				"identity_providers: oidc: clients: client 'test': option 'introspection_encrypted_response_enc' must be one of 'A128GCM', 'A192GCM', 'A256GCM', 'A128CBC-HS256', 'A192CBC-HS384', or 'A256CBC-HS512' but it's configured as 'efenc'",
			},
		},
		{
			"ShouldRaiseErrorsOnValidEncryptionAlgsSetWithoutSigningAlgs",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].AuthorizationEncryptedResponseAlg = oidc.EncryptionAlgECDHES
				have.Clients[0].IDTokenEncryptedResponseAlg = oidc.EncryptionAlgECDHES
				have.Clients[0].UserinfoEncryptedResponseAlg = oidc.EncryptionAlgECDHES
				have.Clients[0].IntrospectionEncryptedResponseAlg = oidc.EncryptionAlgECDHES
				have.Clients[0].AccessTokenEncryptedResponseAlg = oidc.EncryptionAlgECDHES

				have.Clients[0].AuthorizationEncryptedResponseEnc = ""
				have.Clients[0].IDTokenEncryptedResponseEnc = ""
				have.Clients[0].UserinfoEncryptedResponseEnc = ""
				have.Clients[0].IntrospectionEncryptedResponseEnc = ""
				have.Clients[0].AccessTokenEncryptedResponseEnc = ""
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.EncryptionAlgECDHES, have.Clients[0].AuthorizationEncryptedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgECDHES, have.Clients[0].IDTokenEncryptedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgECDHES, have.Clients[0].UserinfoEncryptedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgECDHES, have.Clients[0].IntrospectionEncryptedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgECDHES, have.Clients[0].AccessTokenEncryptedResponseAlg)

				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].AuthorizationEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].IDTokenEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].UserinfoEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].IntrospectionEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].AccessTokenEncryptedResponseEnc)
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
				"identity_providers: oidc: clients: client 'test': option 'jwks_uri' or 'jwks' must be configured when 'authorization_encrypted_response_alg' is set to 'ECDH-ES'",
				"identity_providers: oidc: clients: client 'test': option 'jwks_uri' or 'jwks' must be configured when 'id_token_encrypted_response_alg' is set to 'ECDH-ES'",
				"identity_providers: oidc: clients: client 'test': option 'jwks_uri' or 'jwks' must be configured when 'access_token_encrypted_response_alg' is set to 'ECDH-ES'",
				"identity_providers: oidc: clients: client 'test': option 'access_token_encrypted_response_alg' must either not be configured or set to 'none' if 'access_token_signed_response_alg' is set to 'none'",
				"identity_providers: oidc: clients: client 'test': option 'jwks_uri' or 'jwks' must be configured when 'userinfo_encrypted_response_alg' is set to 'ECDH-ES'",
				"identity_providers: oidc: clients: client 'test': option 'userinfo_encrypted_response_alg' must either not be configured or set to 'none' if 'userinfo_signed_response_alg' is set to 'none'",
				"identity_providers: oidc: clients: client 'test': option 'jwks_uri' or 'jwks' must be configured when 'introspection_encrypted_response_alg' is set to 'ECDH-ES'",
				"identity_providers: oidc: clients: client 'test': option 'introspection_encrypted_response_alg' must either not be configured or set to 'none' if 'introspection_signed_response_alg' is set to 'none'",
			},
		},
		{
			"ShouldNotRaiseErrorsOnValidEncryptionAlgsSetWithSigningAlgs",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].JSONWebKeysURI = &url.URL{Scheme: "https", Host: "example.com", Path: "/jwks.json"}

				have.Clients[0].AuthorizationEncryptedResponseAlg = oidc.EncryptionAlgECDHES
				have.Clients[0].AuthorizationSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
				have.Clients[0].IDTokenEncryptedResponseAlg = oidc.EncryptionAlgECDHES
				have.Clients[0].UserinfoEncryptedResponseAlg = oidc.EncryptionAlgECDHES
				have.Clients[0].UserinfoSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
				have.Clients[0].IntrospectionEncryptedResponseAlg = oidc.EncryptionAlgECDHES
				have.Clients[0].IntrospectionSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
				have.Clients[0].AccessTokenEncryptedResponseAlg = oidc.EncryptionAlgECDHES
				have.Clients[0].AccessTokenSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256

				have.Clients[0].AuthorizationEncryptedResponseEnc = ""
				have.Clients[0].IDTokenEncryptedResponseEnc = ""
				have.Clients[0].UserinfoEncryptedResponseEnc = ""
				have.Clients[0].IntrospectionEncryptedResponseEnc = ""
				have.Clients[0].AccessTokenEncryptedResponseEnc = ""
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.EncryptionAlgECDHES, have.Clients[0].AuthorizationEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].AuthorizationSignedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgECDHES, have.Clients[0].IDTokenEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgECDHES, have.Clients[0].UserinfoEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgECDHES, have.Clients[0].IntrospectionEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgECDHES, have.Clients[0].AccessTokenEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].AccessTokenSignedResponseAlg)

				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].AuthorizationEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].IDTokenEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].UserinfoEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].IntrospectionEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].AccessTokenEncryptedResponseEnc)

				assert.True(t, have.Discovery.JWTResponseAccessTokens)
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
			"ShouldRaiseErrorsClientSecretNotPlainText",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].AuthorizationEncryptedResponseAlg = oidc.EncryptionAlgA128KW
				have.Clients[0].AuthorizationSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
				have.Clients[0].IDTokenEncryptedResponseAlg = oidc.EncryptionAlgA128KW
				have.Clients[0].UserinfoEncryptedResponseAlg = oidc.EncryptionAlgA128KW
				have.Clients[0].UserinfoSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
				have.Clients[0].IntrospectionEncryptedResponseAlg = oidc.EncryptionAlgA128KW
				have.Clients[0].IntrospectionSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
				have.Clients[0].AccessTokenEncryptedResponseAlg = oidc.EncryptionAlgA128KW
				have.Clients[0].AccessTokenSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256

				have.Clients[0].AuthorizationEncryptedResponseEnc = ""
				have.Clients[0].IDTokenEncryptedResponseEnc = ""
				have.Clients[0].UserinfoEncryptedResponseEnc = ""
				have.Clients[0].IntrospectionEncryptedResponseEnc = ""
				have.Clients[0].AccessTokenEncryptedResponseEnc = ""
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.EncryptionAlgA128KW, have.Clients[0].AuthorizationEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].AuthorizationSignedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgA128KW, have.Clients[0].IDTokenEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgA128KW, have.Clients[0].UserinfoEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgA128KW, have.Clients[0].IntrospectionEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, oidc.EncryptionAlgA128KW, have.Clients[0].AccessTokenEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, have.Clients[0].AccessTokenSignedResponseAlg)

				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].AuthorizationEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].IDTokenEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].UserinfoEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].IntrospectionEncryptedResponseEnc)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].AccessTokenEncryptedResponseEnc)
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
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'authorization_encrypted_response_alg' with a value of 'A128KW'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'id_token_encrypted_response_alg' with a value of 'A128KW'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'access_token_encrypted_response_alg' with a value of 'A128KW'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'userinfo_encrypted_response_alg' with a value of 'A128KW'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'introspection_encrypted_response_alg' with a value of 'A128KW'",
			},
		},
		{
			"ShouldRaiseErrorsClientSecretNotPlainTextSig",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].AuthorizationSignedResponseAlg = oidc.SigningAlgHMACUsingSHA256
				have.Clients[0].IDTokenSignedResponseAlg = oidc.SigningAlgHMACUsingSHA256
				have.Clients[0].UserinfoSignedResponseAlg = oidc.SigningAlgHMACUsingSHA256
				have.Clients[0].IntrospectionSignedResponseAlg = oidc.SigningAlgHMACUsingSHA256
				have.Clients[0].AccessTokenSignedResponseAlg = oidc.SigningAlgHMACUsingSHA256
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, "", have.Clients[0].AuthorizationEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgHMACUsingSHA256, have.Clients[0].AuthorizationSignedResponseAlg)
				assert.Equal(t, "", have.Clients[0].IDTokenEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgHMACUsingSHA256, have.Clients[0].IDTokenSignedResponseAlg)
				assert.Equal(t, "", have.Clients[0].UserinfoEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgHMACUsingSHA256, have.Clients[0].UserinfoSignedResponseAlg)
				assert.Equal(t, "", have.Clients[0].IntrospectionEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgHMACUsingSHA256, have.Clients[0].IntrospectionSignedResponseAlg)
				assert.Equal(t, "", have.Clients[0].AccessTokenEncryptedResponseAlg)
				assert.Equal(t, oidc.SigningAlgHMACUsingSHA256, have.Clients[0].AccessTokenSignedResponseAlg)

				assert.Equal(t, "", have.Clients[0].AuthorizationEncryptedResponseEnc)
				assert.Equal(t, "", have.Clients[0].IDTokenEncryptedResponseEnc)
				assert.Equal(t, "", have.Clients[0].UserinfoEncryptedResponseEnc)
				assert.Equal(t, "", have.Clients[0].IntrospectionEncryptedResponseEnc)
				assert.Equal(t, "", have.Clients[0].AccessTokenEncryptedResponseEnc)
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
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'authorization_signed_response_alg' with a value of 'HS256'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'id_token_signed_response_alg' with a value of 'HS256'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'access_token_signed_response_alg' with a value of 'HS256'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'userinfo_signed_response_alg' with a value of 'HS256'",
				"identity_providers: oidc: clients: client 'test': option 'client_secret' must be plaintext with option 'introspection_signed_response_alg' with a value of 'HS256'",
			},
		},
		{
			"ShouldSetDefaultTokenEndpointAuthSigAlg",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].RevocationEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].IntrospectionEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].PushedAuthorizationRequestEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].Secret = tOpenIDConnectPlainTextClientSecret
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = oidc.SigningAlgHMACUsingSHA256
				have.Clients[0].Secret = nil
				have.Clients[0].Public = true
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].RevocationEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].IntrospectionEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].PushedAuthorizationRequestEndpointAuthMethod = oidc.ClientAuthMethodClientSecretJWT
				have.Clients[0].TokenEndpointAuthSigningAlg = oidc.EndpointToken
				have.Clients[0].Secret = tOpenIDConnectPlainTextClientSecret
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
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
			have := &schema.IdentityProvidersOpenIDConnect{
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					Scopes:                    validOIDCClientScopes,
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
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

			validator := schema.NewStructValidator()

			validateOIDDIssuerSigningAlgsDiscovery(have, validator)
			validateOIDCClient(NewValidateCtx(), 0, have, validator, errDeprecatedFunc)

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
				require.Len(t, validator.Warnings(), len(tc.serrs))

				for i, err := range tc.serrs {
					assert.EqualError(t, validator.Warnings()[i], err)
				}
			})

			t.Run("Errors", func(t *testing.T) {
				require.Len(t, validator.Errors(), len(tc.errs))

				for i, err := range tc.errs {
					t.Run(fmt.Sprintf("Error%d", i+1), func(t *testing.T) {
						assert.EqualError(t, validator.Errors()[i], err)
					})
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
		{
			"ShouldSetDefaultValueConfidential",
			"",
			false,
			oidc.ClientAuthMethodClientSecretBasic,
			nil,
		},
		{
			"ShouldErrorOnInvalidValue",
			"abc",
			false,
			"abc",
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_method' must be one of 'none', 'client_secret_post', 'client_secret_basic', 'private_key_jwt', or 'client_secret_jwt' but it's configured as 'abc'",
			},
		},
		{
			"ShouldErrorOnInvalidValueForPublicClient",
			"client_secret_post",
			true,
			"client_secret_post",
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
			have := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                      "test",
						Public:                  tc.public,
						TokenEndpointAuthMethod: tc.have,
					},
				},
			}

			validator := schema.NewStructValidator()

			have.Clients[0].TokenEndpointAuthMethod, have.Clients[0].TokenEndpointAuthSigningAlg, _, _ = validateOIDCClientEndpointAuth(0, have, attrOIDCTokenAuthMethod, have.Clients[0].TokenEndpointAuthMethod, attrOIDCTokenAuthSigningAlg, have.Clients[0].TokenEndpointAuthSigningAlg, validator)

			assert.Equal(t, tc.expected, have.Clients[0].TokenEndpointAuthMethod)
			assert.Len(t, validator.Warnings(), 0)
			require.Len(t, validator.Errors(), len(tc.errs))

			if tc.errs != nil {
				for i, err := range tc.errs {
					assert.EqualError(t, validator.Errors()[i], err)
				}
			}
		})
	}
}

func TestValidateOIDCClientJWKS(t *testing.T) {
	frankenchain := schema.NewX509CertificateChainFromCerts([]*x509.Certificate{certRSA2048.Leaf(), certRSA1024.Leaf()})
	frankenkey := &rsa.PrivateKey{}

	*frankenkey = *keyRSA2048

	frankenkey.N = nil

	testCases := []struct {
		name     string
		haveURI  *url.URL
		haveJWKS []schema.JWK
		setup    func(config *schema.IdentityProvidersOpenIDConnect)
		expected func(t *testing.T, config *schema.IdentityProvidersOpenIDConnect)
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
				"identity_providers: oidc: clients: client 'test': option 'jwks_uri' must have the 'https' scheme but the scheme is 'http'",
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
				"identity_providers: oidc: clients: client 'test': option 'jwks_uri' must not be defined at the same time as option 'jwks'",
			},
		},
		{
			"ShouldAllowGoodKey",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048Legacy.Public()},
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
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048Legacy},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': jwks: key #1 with key id 'test': option 'key' must be a RSA public key or ECDSA public key but it's type is *rsa.PrivateKey",
			},
		},
		{
			"ShouldErrorOnMissingKID",
			nil,
			[]schema.JWK{
				{Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048Legacy.Public()},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': jwks: key #1: option 'key_id' must be provided",
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
				"identity_providers: oidc: clients: client 'test': jwks: key #1 with key id 'KeyID': option 'key' failed to get key properties: the key type 'string' is unknown or not valid for the configuration",
			},
		},
		{
			"ShouldFailOnBadUseAlg",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: "cne", Algorithm: "bad", Key: keyRSA2048Legacy.Public()},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': jwks: key #1 with key id 'test': option 'use' must be one of 'sig' or 'enc' but it's configured as 'cne'",
				"identity_providers: oidc: clients: client 'test': jwks: key #1 with key id 'test': option 'algorithm' must be one of 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', or 'ES512' but it's configured as 'bad'",
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
				"identity_providers: oidc: clients: client 'test': jwks: key #1 with key id 'test': option 'key' must be provided",
			},
		},
		{
			"ShouldFailOnEmptyKeyNoKeyID",
			nil,
			[]schema.JWK{
				{KeyID: "", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: nil},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': jwks: key #1: option 'key_id' must be provided",
				"identity_providers: oidc: clients: client 'test': jwks: key #1: option 'key' must be provided",
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
				"identity_providers: oidc: clients: client 'test': jwks: key #1: option 'key' option 'key' must be a valid private key but the provided data is malformed as it's missing the public key bits",
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
				"identity_providers: oidc: clients: client 'test': jwks: key #1 with key id 'test': option 'key' is an RSA 1024 bit private key but it must at minimum be a RSA 2048 bit private key",
			},
		},
		{
			"ShouldFailOnMismatchedKeyCert",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048Legacy.Public(), CertificateChain: certRSA1024},
			},
			nil,
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': jwks: key #1 with key id 'test': option 'certificate_chain' does not appear to contain the public key for the public key provided by option 'key'",
			},
		},
		{
			"ShouldNotOnLegacyKey",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: keyRSA2048Legacy.Public(), CertificateChain: certRSA2048},
			},
			nil,
			nil,
			nil,
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
				"identity_providers: oidc: clients: client 'test': jwks: key #1 with key id 'test': option 'certificate_chain' produced an error during validation of the chain: certificate #1 in chain is not signed properly by certificate #2 in chain: x509: invalid signature: parent certificate cannot sign this kind of certificate",
			},
		},
		{
			"ShouldSetDefaultUseAlgRSA",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: "", Algorithm: "", Key: keyRSA2048Legacy.Public()},
			},
			nil,
			func(t *testing.T, config *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].JSONWebKeys[0].Use)
				assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, config.Clients[0].JSONWebKeys[0].Algorithm)
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
			func(t *testing.T, config *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].JSONWebKeys[0].Use)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP256AndSHA256, config.Clients[0].JSONWebKeys[0].Algorithm)
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
			func(t *testing.T, config *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].JSONWebKeys[0].Use)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP384AndSHA384, config.Clients[0].JSONWebKeys[0].Algorithm)
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
			func(t *testing.T, config *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].JSONWebKeys[0].Use)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, config.Clients[0].JSONWebKeys[0].Algorithm)
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
			func(t *testing.T, config *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].JSONWebKeys[0].Use)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, config.Clients[0].JSONWebKeys[0].Algorithm)

				assert.Equal(t, []string{oidc.SigningAlgECDSAUsingP521AndSHA512}, config.Discovery.RequestObjectSigningAlgs)
			},
			nil,
		},
		{
			"ShouldOnlyAllowRequestObjectSigningAlgsThatTheClientHasKeysFor",
			nil,
			[]schema.JWK{
				{KeyID: "test", Use: "", Algorithm: "", Key: keyECDSAP521.Public()},
			},
			func(config *schema.IdentityProvidersOpenIDConnect) {
				config.Clients[0].RequestObjectSigningAlg = oidc.SigningAlgRSAUsingSHA512
			},
			func(t *testing.T, config *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.KeyUseSignature, config.Clients[0].JSONWebKeys[0].Use)
				assert.Equal(t, oidc.SigningAlgECDSAUsingP521AndSHA512, config.Clients[0].JSONWebKeys[0].Algorithm)

				assert.Equal(t, []string{oidc.SigningAlgECDSAUsingP521AndSHA512}, config.Discovery.RequestObjectSigningAlgs)
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'request_object_signing_alg' must be one of 'ES512' configured in the client option 'jwks'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:             "test",
						JSONWebKeysURI: tc.haveURI,
						JSONWebKeys:    tc.haveJWKS,
					},
				},
			}

			if tc.setup != nil {
				tc.setup(config)
			}

			validator := schema.NewStructValidator()

			validateOIDCClientPublicKeys(0, config, validator)

			if tc.expected != nil {
				tc.expected(t, config)
			}

			n := len(tc.errs)

			assert.Len(t, validator.Warnings(), 0)

			theErrors := validator.Errors()
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

	frankenkey.N = nil

	testCases := []struct {
		name     string
		have     *schema.IdentityProvidersOpenIDConnect
		expected schema.IdentityProvidersOpenIDConnect
		errs     []string
	}{
		{
			"ShouldMapLegacyConfiguration",
			&schema.IdentityProvidersOpenIDConnect{
				IssuerPrivateKey: keyRSA2048,
			},
			schema.IdentityProvidersOpenIDConnect{
				IssuerPrivateKey: keyRSA2048,
				JSONWebKeys: []schema.JWK{
					{KeyID: "35db6c-rs256", Key: keyRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs:          map[string]string{oidc.SigningAlgRSAUsingSHA256: "35db6c-rs256"},
					DefaultEncKeyIDs:          map[string]string{},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			nil,
		},
		{
			"ShouldSetDefaultKeyValues",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048},
					{Key: keyECDSAP256, CertificateChain: certECDSAP256},
					{Key: keyECDSAP384, CertificateChain: certECDSAP384},
					{Key: keyECDSAP521, CertificateChain: certECDSAP521},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "35db6c-rs256"},
					{Key: keyECDSAP256, CertificateChain: certECDSAP256, Algorithm: oidc.SigningAlgECDSAUsingP256AndSHA256, Use: oidc.KeyUseSignature, KeyID: "d0fe7d-es256"},
					{Key: keyECDSAP384, CertificateChain: certECDSAP384, Algorithm: oidc.SigningAlgECDSAUsingP384AndSHA384, Use: oidc.KeyUseSignature, KeyID: "45839a-es384"},
					{Key: keyECDSAP521, CertificateChain: certECDSAP521, Algorithm: oidc.SigningAlgECDSAUsingP521AndSHA512, Use: oidc.KeyUseSignature, KeyID: "556238-es512"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{
						oidc.SigningAlgRSAUsingSHA256:          "35db6c-rs256",
						oidc.SigningAlgECDSAUsingP256AndSHA256: "d0fe7d-es256",
						oidc.SigningAlgECDSAUsingP384AndSHA384: "45839a-es384",
						oidc.SigningAlgECDSAUsingP521AndSHA512: "556238-es512",
					},
					DefaultEncKeyIDs:          map[string]string{},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgECDSAUsingP384AndSHA384, oidc.SigningAlgECDSAUsingP521AndSHA512},
				},
			},
			nil,
		},
		{
			"ShouldNotRaiseErrorsMultipleRSA256Keys",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048},
					{Key: keyRSA4096, CertificateChain: certRSA4096},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "35db6c-rs256"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "c4c7ca-rs256"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs:          map[string]string{oidc.SigningAlgRSAUsingSHA256: "35db6c-rs256"},
					DefaultEncKeyIDs:          map[string]string{},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			nil,
		},
		{
			"ShouldRaiseErrorsDuplicateRSA256Keys",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA512},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA512, Use: oidc.KeyUseSignature, KeyID: "c4c7ca-rs512"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs:          map[string]string{oidc.SigningAlgRSAUsingSHA512: "c4c7ca-rs512"},
					DefaultEncKeyIDs:          map[string]string{},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA512},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: keys: must at least have one key supporting the 'RS256' algorithm but only has 'RS512'",
			},
		},
		{
			"ShouldRaiseErrorOnBadCurve",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096},
					{Key: keyECDSAP224, CertificateChain: certECDSAP224},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "c4c7ca-rs256"},
					{Key: keyECDSAP224, CertificateChain: certECDSAP224},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs:          map[string]string{oidc.SigningAlgRSAUsingSHA256: "c4c7ca-rs256"},
					DefaultEncKeyIDs:          map[string]string{},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #2: option 'key' failed to calculate thumbprint to configure key id value: go-jose/go-jose: unsupported/unknown elliptic curve",
			},
		},
		{
			"ShouldRaiseErrorOnBadRSAKey",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA1024, CertificateChain: certRSA1024},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA1024, CertificateChain: certRSA1024, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "09920c-rs256"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs:          map[string]string{oidc.SigningAlgRSAUsingSHA256: "09920c-rs256"},
					DefaultEncKeyIDs:          map[string]string{},
					ResponseObjectSigningAlgs: []string{oidc.SigningAlgRSAUsingSHA256},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id '09920c-rs256': option 'key' is an RSA 1024 bit private key but it must at minimum be a RSA 2048 bit private key",
			},
		},
		{
			"ShouldRaiseErrorOnBadAlg",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: "invalid"},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: "invalid", Use: oidc.KeyUseSignature, KeyID: "c4c7ca-invalid"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs:          map[string]string{},
					DefaultEncKeyIDs:          map[string]string{},
					ResponseObjectSigningAlgs: []string{"invalid"},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id 'c4c7ca-invalid': option 'algorithm' must be one of 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', or 'ES512' but it's configured as 'invalid'",
				"identity_providers: oidc: jwks: keys: must at least have one key supporting the 'RS256' algorithm but only has 'invalid'",
			},
		},
		{
			"ShouldErrorNilKeyWithoutKID",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: nil, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1: option 'key' must be provided",
			},
		},
		{
			"ShouldErrorNilKey",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: nil, KeyID: "example", Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{KeyID: "example", Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id 'example': option 'key' must be provided",
			},
		},
		{
			"ShouldRaiseErrorOnBadUse",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Use: "invalid"},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Use: "invalid", KeyID: "c4c7ca-rs256"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs:          map[string]string{},
					DefaultEncKeyIDs:          map[string]string{},
					ResponseObjectSigningAlgs: []string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id 'c4c7ca-rs256': option 'use' must be one of 'sig' or 'enc' but it's configured as 'invalid'",
			},
		},
		{
			"ShouldRaiseErrorOnBadKeyIDLength",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "thisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolong"},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "thisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolong"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{oidc.SigningAlgRSAUsingSHA256: "thisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolong"},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id 'thisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolongthisistoolong': option `key_id` must be 100 characters or less",
			},
		},
		{
			"ShouldRaiseErrorOnBadKeyIDCharacters",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x@x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "-xx"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "xx."},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x@x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "-xx"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "xx."},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{oidc.SigningAlgRSAUsingSHA256: "x@x"},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id 'x@x': option 'key_id' must only contain RFC3986 unreserved characters and must only start and end with alphanumeric characters",
				"identity_providers: oidc: jwks: key #2 with key id '-xx': option 'key_id' must only contain RFC3986 unreserved characters and must only start and end with alphanumeric characters",
				"identity_providers: oidc: jwks: key #3 with key id 'xx.': option 'key_id' must only contain RFC3986 unreserved characters and must only start and end with alphanumeric characters",
			},
		},
		{
			"ShouldNotRaiseErrorOnGoodKeyIDCharacters",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x-x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "xx"},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x-x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x"},
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "xx"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{oidc.SigningAlgRSAUsingSHA256: "x-x"},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnBadKeyIDDuplicates",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x"},
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAPSSUsingSHA256, KeyID: "x"},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x"},
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.SigningAlgRSAPSSUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{oidc.SigningAlgRSAUsingSHA256: "x", oidc.SigningAlgRSAPSSUsingSHA256: "x"},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #2 with key id 'x': option 'key_id' must be unique",
			},
		},
		{
			"ShouldRaiseErrorOnBadKeyIDDuplicatesEnc",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, KeyID: "x"},
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.EncryptionAlgRSAOAEP256, Use: oidc.KeyUseEncryption, KeyID: "x"},
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.EncryptionAlgRSAOAEP256, Use: oidc.KeyUseEncryption, KeyID: "x"},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA4096, CertificateChain: certRSA4096, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "x"},
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.EncryptionAlgRSAOAEP256, Use: oidc.KeyUseEncryption, KeyID: "x"},
					{Key: keyRSA2048, CertificateChain: certRSA2048, Algorithm: oidc.EncryptionAlgRSAOAEP256, Use: oidc.KeyUseEncryption, KeyID: "x"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{oidc.SigningAlgRSAUsingSHA256: "x"},
					DefaultEncKeyIDs: map[string]string{oidc.EncryptionAlgRSAOAEP256: "x"},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #3 with key id 'x': option 'key_id' must be unique",
			},
		},
		{
			"ShouldRaiseErrorOnEd25519Keys",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyEd2519, CertificateChain: certEd15519},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyEd2519, CertificateChain: certEd15519, KeyID: "ca54bd"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id 'ca54bd': option 'use' must be one of 'sig' or 'enc' but it's configured as ''",
				"identity_providers: oidc: jwks: key #1 with key id 'ca54bd': option 'key' must be a RSA private key or ECDSA private key but it's type is ed25519.PrivateKey",
			},
		},
		{
			"ShouldRaiseErrorOnCertificateAsKey",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: certRSA2048.Certificates()[0].PublicKey},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: certRSA2048.Certificates()[0].PublicKey, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "35db6c-rs256"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{oidc.SigningAlgRSAUsingSHA256: "35db6c-rs256"},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id '35db6c-rs256': option 'key' must be a RSA private key or ECDSA private key but it's type is *rsa.PublicKey",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidChain",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: frankenchain},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, CertificateChain: frankenchain, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature, KeyID: "35db6c-rs256"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{oidc.SigningAlgRSAUsingSHA256: "35db6c-rs256"},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id '35db6c-rs256': option 'certificate_chain' produced an error during validation of the chain: certificate #1 in chain is not signed properly by certificate #2 in chain: x509: invalid signature: parent certificate cannot sign this kind of certificate",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidPrivateKeyN",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: frankenkey},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: frankenkey},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1: option 'key' must be a valid private key but the provided data is malformed as it's missing the public key bits",
			},
		},
		{
			"ShouldRaiseErrorOnInvalidEncPrivateKeyN",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: frankenkey, Use: oidc.KeyUseEncryption},
					{Key: keyRSA2048, Use: oidc.KeyUseSignature},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: frankenkey, Use: oidc.KeyUseEncryption},
					{Key: keyRSA2048, Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, KeyID: "35db6c-rs256"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{"RS256": "35db6c-rs256"},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1: option 'key' must be a valid private key but the provided data is malformed as it's missing the public key bits",
			},
		},
		{
			"ShouldHandleEncKey",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, Use: oidc.KeyUseEncryption},
					{Key: keyRSA2048, Use: oidc.KeyUseSignature},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, Use: oidc.KeyUseEncryption, Algorithm: oidc.EncryptionAlgRSAOAEP256, KeyID: "35db6c-rsa-oaep-256"},
					{Key: keyRSA2048, Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, KeyID: "35db6c-rs256"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{"RS256": "35db6c-rs256"},
					DefaultEncKeyIDs: map[string]string{"RSA-OAEP-256": "35db6c-rsa-oaep-256"},
				},
			},
			nil,
		},
		{
			"ShouldHandleEncKeyBadAlg",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, Use: oidc.KeyUseEncryption, Algorithm: oidc.SigningAlgRSAUsingSHA256},
					{Key: keyRSA2048, Use: oidc.KeyUseSignature},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, Use: oidc.KeyUseEncryption, Algorithm: oidc.SigningAlgRSAUsingSHA256, KeyID: "35db6c-rs256"},
					{Key: keyRSA2048, Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, KeyID: "35db6c-rs256"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{"RS256": "35db6c-rs256"},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id '35db6c-rs256': option 'algorithm' must be one of 'RSA1_5', 'RSA-OAEP', 'RSA-OAEP-256', 'A128KW', 'A192KW', 'A256KW', 'dir', 'ECDH-ES', 'ECDH-ES+A128KW', 'ECDH-ES+A192KW', 'ECDH-ES+A256KW', 'A128GCMKW', 'A192GCMKW', 'A256GCMKW', 'PBES2-HS256+A128KW', 'PBES2-HS384+A192KW', or 'PBES2-HS512+A256KW' but it's configured as 'RS256'",
			},
		},
		{
			"ShouldRaiseErrorOnCertForKey",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: certRSA2048},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: certRSA2048},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id '': option 'key' failed to get key properties: the key type 'schema.X509CertificateChain' is unknown or not valid for the configuration",
			},
		},
	}

	var n int

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			validateOIDCIssuer(tc.have, validator)

			assert.Equal(t, tc.expected.Discovery.DefaultSigKeyIDs, tc.have.Discovery.DefaultSigKeyIDs)
			assert.Equal(t, tc.expected.Discovery.DefaultEncKeyIDs, tc.have.Discovery.DefaultEncKeyIDs)
			assert.Equal(t, tc.expected.IssuerPrivateKey, tc.have.IssuerPrivateKey)
			assert.Equal(t, tc.expected.IssuerCertificateChain, tc.have.IssuerCertificateChain)

			n = len(tc.expected.JSONWebKeys)

			require.Len(t, tc.have.JSONWebKeys, n)

			for i := 0; i < n; i++ {
				t.Run(fmt.Sprintf("Key%d", i), func(t *testing.T) {
					assert.Equal(t, tc.expected.JSONWebKeys[i].Algorithm, tc.have.JSONWebKeys[i].Algorithm)
					assert.Equal(t, tc.expected.JSONWebKeys[i].Use, tc.have.JSONWebKeys[i].Use)
					assert.Equal(t, tc.expected.JSONWebKeys[i].KeyID, tc.have.JSONWebKeys[i].KeyID)
					assert.Equal(t, tc.expected.JSONWebKeys[i].Key, tc.have.JSONWebKeys[i].Key)
					assert.Equal(t, tc.expected.JSONWebKeys[i].CertificateChain, tc.have.JSONWebKeys[i].CertificateChain)
				})
			}

			n = len(tc.errs)

			require.Len(t, validator.Errors(), n)

			for i := 0; i < n; i++ {
				assert.EqualError(t, validator.Errors()[i], tc.errs[i])
			}
		})
	}
}

func TestValidateLifespans(t *testing.T) {
	testCases := []struct {
		name     string
		have     *schema.Configuration
		expected []string
		errors   []string
	}{
		{
			"ShouldHandleNone",
			&schema.Configuration{IdentityProviders: schema.IdentityProviders{OIDC: &schema.IdentityProvidersOpenIDConnect{}}},
			nil,
			nil,
		},
		{
			"ShouldHandleCustom",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Lifespans: schema.IdentityProvidersOpenIDConnectLifespans{
							Custom: map[string]schema.IdentityProvidersOpenIDConnectLifespan{
								"custom": {},
							},
						},
					},
				},
			},
			[]string{"custom"},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			validateOIDCLifespans(tc.have, validator)

			assert.Equal(t, tc.expected, tc.have.IdentityProviders.OIDC.Discovery.Lifespans)
			require.Len(t, validator.Errors(), len(tc.errors))

			for i, err := range tc.errors {
				t.Run(fmt.Sprintf("Error%d", i+1), func(t *testing.T) {
					assert.EqualError(t, validator.Errors()[i], err)
				})
			}
		})
	}
}

func TestValidateOIDCAuthorizationPolicies(t *testing.T) {
	testCases := []struct {
		name     string
		have     *schema.Configuration
		expected []string
		expectf  func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnect)
		errors   []string
	}{
		{
			"ShouldIncludeDefaults",
			&schema.Configuration{IdentityProviders: schema.IdentityProviders{OIDC: &schema.IdentityProvidersOpenIDConnect{}}},
			[]string{"one_factor", "two_factor"},
			nil,
			nil,
		},
		{
			"ShouldErrorOnInvalidPoliciesNoRules",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						AuthorizationPolicies: map[string]schema.IdentityProvidersOpenIDConnectPolicy{
							"example": {
								DefaultPolicy: "two_factor",
							},
						},
					},
				},
			},
			[]string{"one_factor", "two_factor", "example"},
			nil,
			[]string{
				"identity_providers: oidc: authorization_policies: policy 'example': option 'rules' is required",
			},
		},
		{
			"ShouldIncludeValidPolicies",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						AuthorizationPolicies: map[string]schema.IdentityProvidersOpenIDConnectPolicy{
							"example": {
								DefaultPolicy: "two_factor",
								Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
									{
										Policy: "deny",
										Subjects: [][]string{
											{"user:john"},
										},
									},
								},
							},
						},
					},
				},
			},
			[]string{"one_factor", "two_factor", "example"},
			nil,
			nil,
		},
		{
			"ShouldErrorOnInvalidSubject",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						AuthorizationPolicies: map[string]schema.IdentityProvidersOpenIDConnectPolicy{
							"example": {
								DefaultPolicy: "two_factor",
								Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
									{
										Policy: "deny",
										Subjects: [][]string{
											{"[user:john]"},
										},
									},
								},
							},
						},
					},
				},
			},
			[]string{"one_factor", "two_factor", "example"},
			nil,
			[]string{
				"identity_providers: oidc: authorization_policies: policy 'example': rules: rule #1: option 'subject' with value '[user:john]' is invalid: must start with 'user:' or 'group:'",
			},
		},
		{
			"ShouldSetDefaultPoliciesAndErrorOnSubject",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						AuthorizationPolicies: map[string]schema.IdentityProvidersOpenIDConnectPolicy{
							"example": {
								DefaultPolicy: "",
								Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
									{
										Policy: "",
									},
								},
							},
							"": {
								DefaultPolicy: "two_factor",
								Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
									{
										Policy: "two_factor",
										Subjects: [][]string{
											{"user:john"},
										},
									},
								},
							},
							"two_factor": {
								DefaultPolicy: "two_factor",
								Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
									{
										Policy: "two_factor",
										Subjects: [][]string{
											{"user:john"},
										},
									},
								},
							},
						},
					},
				},
			},
			[]string{"one_factor", "two_factor", "example"},
			func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, "two_factor", actual.AuthorizationPolicies["example"].DefaultPolicy)
				assert.Equal(t, "two_factor", actual.AuthorizationPolicies["example"].Rules[0].Policy)
			},
			[]string{
				"identity_providers: oidc: authorization_policies: authorization policies must have a name but one with a blank name exists",
				"identity_providers: oidc: authorization_policies: policy 'example': rules: rule #1: option 'subject' or 'networks' is required",
				"identity_providers: oidc: authorization_policies: policy 'two_factor': option 'name' must not be one of 'one_factor', 'two_factor', and 'deny' but it's configured as 'two_factor'",
			},
		},
		{
			"ShouldErrorBadPolicyValues",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						AuthorizationPolicies: map[string]schema.IdentityProvidersOpenIDConnectPolicy{
							"example": {
								DefaultPolicy: "abc",
								Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
									{
										Policy: "xyz",
										Subjects: [][]string{
											{"user:john"},
										},
									},
								},
							},
						},
					},
				},
			},
			[]string{"one_factor", "two_factor", "example"},
			nil,
			[]string{
				"identity_providers: oidc: authorization_policies: policy 'example': option 'default_policy' must be one of 'one_factor', 'two_factor', and 'deny' but it's configured as 'abc'",
				"identity_providers: oidc: authorization_policies: policy 'example': rules: rule #1: option 'policy' must be one of 'one_factor', 'two_factor', and 'deny' but it's configured as 'xyz'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			validateOIDCAuthorizationPolicies(tc.have, validator)

			assert.Equal(t, tc.expected, tc.have.IdentityProviders.OIDC.Discovery.AuthorizationPolicies)

			errs := validator.Errors()
			sort.Sort(utils.ErrSliceSortAlphabetical(errs))

			require.Len(t, errs, len(tc.errors))

			for i, err := range tc.errors {
				t.Run(fmt.Sprintf("Error%d", i+1), func(t *testing.T) {
					assert.EqualError(t, errs[i], err)
				})
			}

			if tc.expectf != nil {
				tc.expectf(t, tc.have.IdentityProviders.OIDC)
			}
		})
	}
}

func TestShouldValidateOpenIDConnectClaimsPolicies(t *testing.T) {
	testCases := []struct {
		name    string
		have    *schema.Configuration
		scopes  bool
		expectf func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnect)
		errors  []string
	}{
		{
			name: "ShouldValidateWithoutPolicies",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{File: &schema.AuthenticationBackendFile{}},
				IdentityProviders:     schema.IdentityProviders{OIDC: &schema.IdentityProvidersOpenIDConnect{}},
			},
		},
		{
			name: "ShouldSetDefaultIDTokenAudienceMode",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								IDTokenAudienceMode: "",
							},
							"example-spec": {
								IDTokenAudienceMode: "specification",
							},
							"example-merged": {
								IDTokenAudienceMode: "experimental-merged",
							},
						},
					},
				},
			},
			expectf: func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, "specification", actual.ClaimsPolicies["example"].IDTokenAudienceMode)
				assert.Equal(t, "specification", actual.ClaimsPolicies["example-spec"].IDTokenAudienceMode)
				assert.Equal(t, "experimental-merged", actual.ClaimsPolicies["example-merged"].IDTokenAudienceMode)
			},
		},
		{
			name: "ShouldErrorOnInvalidIDTokenAudienceMode",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								IDTokenAudienceMode: "invalid",
							},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: claims_policies: example: option 'id_token_audience_mode' must be one of 'specification' or 'experimental-merged' but it's configured as 'invalid'",
			},
		},
		{
			name: "ShouldNotAllowReservedOrNonMetaClaims",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								IDToken:     []string{"no-id-claim", "sub"},
								AccessToken: []string{"no-at-claim", "sub"},
							},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: claims_policies: example: access_token: claim with name 'no-at-claim' is not known",
				"identity_providers: oidc: claims_policies: example: access_token: claim with name 'sub' can't be used in a claims policy as it's a standard claim",
				"identity_providers: oidc: claims_policies: example: id_token: claim with name 'no-id-claim' is not known",
				"identity_providers: oidc: claims_policies: example: id_token: claim with name 'sub' can't be used in a claims policy as it's a standard claim",
			},
		},
		{
			name: "ShouldAllowNonIDTokenClaims",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								IDToken: []string{"rat", "updated_at"},
							},
						},
					},
				},
			},
		},
		{
			name: "ShouldAllowCustomClaimsManual",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								IDToken:     []string{"id-claim"},
								AccessToken: []string{"at-claim"},
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"id-claim": {Attribute: "email"},
									"at-claim": {Attribute: "email"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ShouldAllowCustomClaimsNameMapping",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
							"example": {
								Claims: []string{"http://example.com/myclaim"},
							},
						},
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								IDToken:     []string{"id-claim"},
								AccessToken: []string{"at-claim"},
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"id-claim": {Attribute: "email"},
									"at-claim": {Attribute: "email"},
									"myclaim":  {Name: "http://example.com/myclaim", Attribute: "email"},
								},
							},
						},
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:           "example",
								Scopes:       []string{"openid", "example"},
								ClaimsPolicy: "example",
							},
						},
					},
				},
			},
			scopes: true,
		},
		{
			name: "ShouldNotAllowCustomClaimsAutoMissingAttribute",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								IDToken:     []string{"id-claim"},
								AccessToken: []string{"at-claim"},
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"id-claim": {},
									"at-claim": {},
								},
							},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: claims_policies: example: claim with name 'at-claim' has an attribute name 'at-claim' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'id-claim' has an attribute name 'id-claim' which is not a known attribute",
			},
		},
		{
			name: "ShouldNotAllowCustomClaimsAutoMissingProvider",
			have: &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								IDToken:     []string{"id-claim"},
								AccessToken: []string{"at-claim"},
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"id-claim": {},
									"at-claim": {},
								},
							},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: claims_policies: example: claim with name 'at-claim' has an attribute name 'at-claim' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'id-claim' has an attribute name 'id-claim' which is not a known attribute",
			},
		},
		{
			name: "ShouldNotAllowCustomClaimsReserved",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"sub": {Attribute: "email"},
								},
							},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: claims_policies: example: custom_claims: claim with name 'sub' can't be used in a claims policy as it's a standard claim",
			},
		},
		{
			name: "ShouldAllowCustomClaimsAuto",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					File: &schema.AuthenticationBackendFile{
						ExtraAttributes: map[string]schema.AuthenticationBackendExtraAttribute{
							"id-claim": {
								ValueType: "string",
							},
							"at-claim": {
								ValueType: "string",
							},
						},
					},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								IDToken:     []string{"id-claim"},
								AccessToken: []string{"at-claim"},
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"id-claim": {},
									"at-claim": {},
								},
							},
						},
					},
				},
			},
			expectf: func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, "at-claim", actual.ClaimsPolicies["example"].CustomClaims["at-claim"].Attribute)
				assert.Equal(t, "id-claim", actual.ClaimsPolicies["example"].CustomClaims["id-claim"].Attribute)
			},
		},
		{
			name: "ShouldNotAllowStandardClaimUnmappedLDAP",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					LDAP: &schema.AuthenticationBackendLDAP{},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								IDToken:     []string{"id-claim"},
								AccessToken: []string{"at-claim"},
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"id-claim":              {Attribute: "given_name"},
									"at-claim":              {Attribute: "middle_name"},
									"claim-family_name":     {Attribute: "family_name"},
									"claim-nickname":        {Attribute: "nickname"},
									"claim-profile":         {Attribute: "profile"},
									"claim-picture":         {Attribute: "picture"},
									"claim-website":         {Attribute: "website"},
									"claim-gender":          {Attribute: "gender"},
									"claim-birthdate":       {Attribute: "birthdate"},
									"claim-zoneinfo":        {Attribute: "zoneinfo"},
									"claim-locale":          {Attribute: "locale"},
									"claim-phone_number":    {Attribute: "phone_number"},
									"claim-phone_extension": {Attribute: "phone_extension"},
									"claim-street_address":  {Attribute: "street_address"},
									"claim-locality":        {Attribute: "locality"},
									"claim-region":          {Attribute: "region"},
									"claim-postal_code":     {Attribute: "postal_code"},
									"claim-country":         {Attribute: "country"},
									"claim-extra1":          {Attribute: "extra1"},
								},
							},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: claims_policies: example: claim with name 'at-claim' has an attribute name 'middle_name' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-birthdate' has an attribute name 'birthdate' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-country' has an attribute name 'country' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-extra1' has an attribute name 'extra1' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-family_name' has an attribute name 'family_name' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-gender' has an attribute name 'gender' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-locale' has an attribute name 'locale' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-locality' has an attribute name 'locality' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-nickname' has an attribute name 'nickname' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-phone_extension' has an attribute name 'phone_extension' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-phone_number' has an attribute name 'phone_number' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-picture' has an attribute name 'picture' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-postal_code' has an attribute name 'postal_code' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-profile' has an attribute name 'profile' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-region' has an attribute name 'region' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-street_address' has an attribute name 'street_address' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-website' has an attribute name 'website' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-zoneinfo' has an attribute name 'zoneinfo' which is not a known attribute",
				"identity_providers: oidc: claims_policies: example: claim with name 'id-claim' has an attribute name 'given_name' which is not a known attribute",
			},
		},
		{
			name: "ShouldNotAllowStandardClaimUnmappedLDAPWithExtra",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					LDAP: &schema.AuthenticationBackendLDAP{
						Attributes: schema.AuthenticationBackendLDAPAttributes{
							Extra: map[string]schema.AuthenticationBackendLDAPAttributesAttribute{},
						},
					},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"claim-extra1": {Attribute: "extra1"},
								},
							},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-extra1' has an attribute name 'extra1' which is not a known attribute",
			},
		},
		{
			name: "ShouldNotAllowStandardClaimMappedLDAPWithExtraMismatchedName",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					LDAP: &schema.AuthenticationBackendLDAP{
						Attributes: schema.AuthenticationBackendLDAPAttributes{
							Extra: map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
								"extra1": {
									Name: "x",
									AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
										ValueType: "string",
									},
								},
							},
						},
					},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"claim-extra1": {Attribute: "extra1"},
								},
							},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: claims_policies: example: claim with name 'claim-extra1' has an attribute name 'extra1' which is not a known attribute",
			},
		},
		{
			name: "ShouldAllowStandardClaimMappedLDAPWithExtra",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					LDAP: &schema.AuthenticationBackendLDAP{
						Attributes: schema.AuthenticationBackendLDAPAttributes{
							Extra: map[string]schema.AuthenticationBackendLDAPAttributesAttribute{
								"extra1": {
									Name: "extra1",
									AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
										ValueType: "string",
									},
								},
								"extra2": {
									AuthenticationBackendExtraAttribute: schema.AuthenticationBackendExtraAttribute{
										ValueType: "string",
									},
								},
							},
						},
					},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"claim-extra1": {Attribute: "extra1"},
									"claim-extra2": {Attribute: "extra2"},
									"claim-extra3": {Attribute: "username"},
									"claim-extra4": {Attribute: "email"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ShouldAllowDefinitionClaim",
			have: &schema.Configuration{
				Definitions: schema.Definitions{
					UserAttributes: map[string]schema.UserAttribute{
						"custom-1": {},
					},
				},
				AuthenticationBackend: schema.AuthenticationBackend{
					LDAP: &schema.AuthenticationBackendLDAP{},
				},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
							"example": {
								CustomClaims: map[string]schema.IdentityProvidersOpenIDConnectCustomClaim{
									"claim-extra1": {Attribute: "custom-1"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val := schema.NewStructValidator()

			validateOIDCClaims(tc.have, val)

			if tc.scopes {
				require.Len(t, tc.have.IdentityProviders.OIDC.Clients, 1)

				validateOIDCClientScopes(0, tc.have.IdentityProviders.OIDC, val, false, func() {

				})
			}

			if tc.expectf != nil {
				tc.expectf(t, tc.have.IdentityProviders.OIDC)
			}

			errs := val.Errors()
			sort.Sort(utils.ErrSliceSortAlphabetical(errs))

			require.Len(t, errs, len(tc.errors))

			for i, err := range tc.errors {
				assert.EqualError(t, errs[i], err)
			}
		})
	}
}

func TestShouldValidateOpenIDConnectScopes(t *testing.T) {
	testCases := []struct {
		name    string
		have    *schema.Configuration
		expectf func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnect)
		errors  []string
	}{
		{
			name: "ShouldValidateWithoutPolicies",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{File: &schema.AuthenticationBackendFile{}},
				IdentityProviders:     schema.IdentityProviders{OIDC: &schema.IdentityProvidersOpenIDConnect{}},
			},
		},
		{
			name: "ShouldNotValidateUnknownClaim",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{File: &schema.AuthenticationBackendFile{}},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
							"scopes-ex": {
								Claims: []string{"example"},
							},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: scopes: scopes-ex: claim with name 'example' is not a known claim",
			},
		},
		{
			name: "ShouldValidateKnownClaim",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{File: &schema.AuthenticationBackendFile{}},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
							"scopes-ex": {
								Claims: []string{"example"},
							},
						},
						Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
							Claims: []string{"example"},
						},
					},
				},
			},
		},
		{
			name: "ShouldNotValidateStandardClaim",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{File: &schema.AuthenticationBackendFile{}},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
							"scopes-ex": {
								Claims: []string{"sub"},
							},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: scopes: scopes-ex: claim with name 'sub' can't be used in a custom scope as it's a standard claim",
			},
		},
		{
			name: "ShouldNotValidateSpecialReservedScopes",
			have: &schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{File: &schema.AuthenticationBackendFile{}},
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Scopes: map[string]schema.IdentityProvidersOpenIDConnectScope{
							"authelia.scopes-ex": {
								Claims: []string{"example"},
							},
							"profile": {},
						},
						Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
							Claims: []string{"example"},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: scopes: scope with name 'authelia.scopes-ex' can't be used as a custom scope because all scopes prefixed with 'authelia.' are reserved",
				"identity_providers: oidc: scopes: scope with name 'profile' can't be used as a custom scope because it's a standard scope",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val := schema.NewStructValidator()

			validateOIDCScopes(tc.have, val)

			if tc.expectf != nil {
				tc.expectf(t, tc.have.IdentityProviders.OIDC)
			}

			errs := val.Errors()
			sort.Sort(utils.ErrSliceSortAlphabetical(errs))

			require.Len(t, errs, len(tc.errors))

			for i, err := range tc.errors {
				assert.EqualError(t, errs[i], err)
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
	fparts := []string{strings.ToLower(alg)}

	if len(mod) != 0 {
		fparts = append(fparts, mod)
	}

	if len(extra) != 0 {
		fparts = append(fparts, extra...)
	}

	var (
		data    []byte
		decoded any
		err     error
	)
	if data, err = os.ReadFile(fmt.Sprintf(pathCrypto, strings.Join(fparts, "."), ext)); err != nil {
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

func MustLoadEd15519PrivateKey(mod string, extra ...string) ed25519.PrivateKey {
	decoded := MustLoadCrypto("ED25519", mod, "pem", extra...)

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

func MustLoadRSAPrivateKey(bits string, extra ...string) *rsa.PrivateKey {
	decoded := MustLoadCrypto("RSA", bits, "pem", extra...)

	key, ok := decoded.(*rsa.PrivateKey)
	if !ok {
		panic(fmt.Errorf("the key was not a *rsa.PrivateKey, it's a %T", key))
	}

	return key
}

func MustLoadRSACertificatePrivateKeyPair(bits string, extra ...string) (chain schema.X509CertificateChain, key *rsa.PrivateKey) {
	return MustLoadCertificateChain("RSA", bits), MustLoadRSAPrivateKey(bits, extra...)
}

func MustLoadECDSACertificatePrivateKeyPair(curve string, extra ...string) (chain schema.X509CertificateChain, key *ecdsa.PrivateKey) {
	return MustLoadCertificateChain("ECDSA", curve), MustLoadECDSAPrivateKey(curve, extra...)
}

const (
	pathCrypto = "../test_resources/crypto/%s.%s"
)

var (
	tOpenIDConnectPBKDF2ClientSecret, tOpenIDConnectPlainTextClientSecret *schema.PasswordDigest

	// Standard RSA key / certificate pairs.
	keyRSA1024, keyRSA2048, keyRSA2048Legacy, keyRSA4096 *rsa.PrivateKey
	certRSA1024, certRSA2048, certRSA4096                schema.X509CertificateChain

	// Standard ECDSA key / certificate pairs.
	keyECDSAP224, keyECDSAP256, keyECDSAP384, keyECDSAP521     *ecdsa.PrivateKey
	certECDSAP224, certECDSAP256, certECDSAP384, certECDSAP521 schema.X509CertificateChain

	// Ed15519 key / certificate pair.
	keyEd2519   ed25519.PrivateKey
	certEd15519 schema.X509CertificateChain
)

func init() {
	tOpenIDConnectPBKDF2ClientSecret = MustDecodeSecret("$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng")
	tOpenIDConnectPlainTextClientSecret = MustDecodeSecret("$plaintext$example")

	certRSA1024, keyRSA1024 = MustLoadRSACertificatePrivateKeyPair("1024")
	certRSA2048, keyRSA2048 = MustLoadRSACertificatePrivateKeyPair("2048")
	certRSA4096, keyRSA4096 = MustLoadRSACertificatePrivateKeyPair("4096")

	certECDSAP224, keyECDSAP224 = MustLoadECDSACertificatePrivateKeyPair("P224")
	certECDSAP256, keyECDSAP256 = MustLoadECDSACertificatePrivateKeyPair("P256")
	certECDSAP384, keyECDSAP384 = MustLoadECDSACertificatePrivateKeyPair("P384")
	certECDSAP521, keyECDSAP521 = MustLoadECDSACertificatePrivateKeyPair("P521")

	certEd15519, keyEd2519 = MustLoadCertificateChain("Ed25519", ""), MustLoadEd15519PrivateKey("")

	keyRSA2048Legacy = MustLoadRSAPrivateKey("2048", "legacy")
}
