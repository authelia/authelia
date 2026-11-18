// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
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

//nolint:gosec // Test Credentials.
func TestShouldNotRaiseErrorWhenCORSEndpointsValid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
				IssuerPrivateKey: keyRSA2048,
				CORS: schema.IdentityProvidersOpenIDConnectCORS{
					Endpoints: []string{oidc.EndpointDeviceAuthorization, oidc.EndpointPushedAuthorizationRequest, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo},
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
	assert.Equal(t, []string{oidc.EndpointDeviceAuthorization, oidc.EndpointPushedAuthorizationRequest, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo}, config.IdentityProviders.OIDC.CORS.Endpoints)
}

//nolint:gosec // Test Credentials.
func TestShouldWarnAndRemoveCORSAuthorizationEndpoint(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
				IssuerPrivateKey: keyRSA2048,
				CORS: schema.IdentityProvidersOpenIDConnectCORS{
					Endpoints: []string{oidc.EndpointAuthorization, oidc.EndpointToken},
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
	assert.Equal(t, []string{oidc.EndpointToken}, config.IdentityProviders.OIDC.CORS.Endpoints)

	var found bool

	for _, warning := range validator.Warnings() {
		if warning.Error() == "identity_providers: oidc: cors: option 'endpoints' contains the value 'authorization' which is ignored as Cross-Origin Resource Sharing must not be supported by the Authorization Endpoint as per RFC9700 Section 2.6" {
			found = true
		}
	}

	assert.True(t, found)
}

//nolint:gosec // Test Credentials.
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

	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: cors: option 'endpoints' contains an invalid value 'invalid_endpoint': must be one of 'device-authorization', 'pushed-authorization-request', 'token', 'introspection', 'revocation', or 'userinfo'")
}

//nolint:gosec // Test Credentials.
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

//nolint:gosec // Test Credentials.
func TestShouldValidateOIDCProviderStrategies(t *testing.T) {
	testCases := []struct {
		Name             string
		Audience         string
		Resource         string
		ExpectedAudience string
		ExpectedResource string
		Errors           []string
	}{
		{"ShouldSetDefaults", "", "", "exact", "exact", nil},
		{"ShouldAllowWildcard", "wildcard", "wildcard", "wildcard", "wildcard", nil},
		{"ShouldAllowExact", "exact", "exact", "exact", "exact", nil},
		{"ShouldErrorInvalid", "prefix", "magic", "prefix", "magic", []string{
			"identity_providers: oidc: option 'audience_strategy' must be one of 'exact' or 'wildcard' but it's configured as 'prefix'",
			"identity_providers: oidc: option 'resource_strategy' must be one of 'exact', 'prefix', or 'wildcard' but it's configured as 'magic'",
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
						IssuerPrivateKey: keyRSA2048,
						AudienceStrategy: tc.Audience,
						ResourceStrategy: tc.Resource,
					},
				},
			}

			ValidateIdentityProviders(NewValidateCtx(), config, validator)

			errs := validator.Errors()

			require.Len(t, errs, len(tc.Errors)+1)

			for i, expected := range tc.Errors {
				assert.EqualError(t, errs[i], expected)
			}

			assert.EqualError(t, errs[len(tc.Errors)], "identity_providers: oidc: option 'clients' must have one or more clients configured")
			assert.Equal(t, tc.ExpectedAudience, config.IdentityProviders.OIDC.AudienceStrategy)
			assert.Equal(t, tc.ExpectedResource, config.IdentityProviders.OIDC.ResourceStrategy)
		})
	}
}

//nolint:gosec // Test Credentials.
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

//nolint:gosec // Test Credentials.
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

//nolint:gosec // Test Credentials.
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
		{
			name: "InvalidAudienceStrategy",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-bad-as",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					AudienceStrategy: "prefix",
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-bad-as': option 'audience_strategy' must be one of 'exact' or 'wildcard' but it's configured as 'prefix'",
			},
		},
		{
			name: "InvalidResourceStrategy",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-bad-rs",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					ResourceStrategy: "magic",
				},
			},
			errors: []string{
				"identity_providers: oidc: clients: client 'client-bad-rs': option 'resource_strategy' must be one of 'exact', 'prefix', or 'wildcard' but it's configured as 'magic'",
			},
		},
		{
			name: "ValidAudienceStrategy",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-good-as",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					AudienceStrategy: "wildcard",
				},
			},
			test: func(t *testing.T, actual []schema.IdentityProvidersOpenIDConnectClient) {
				assert.Equal(t, "wildcard", actual[0].AudienceStrategy)
			},
		},
		{
			name: "ValidResourceStrategy",
			clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "client-good-rs",
					Secret:              tOpenIDConnectPlainTextClientSecret,
					AuthorizationPolicy: policyTwoFactor,
					RedirectURIs: []string{
						"https://google.com",
					},
					ResourceStrategy: "exact",
				},
			},
			test: func(t *testing.T, actual []schema.IdentityProvidersOpenIDConnectClient) {
				assert.Equal(t, "exact", actual[0].ResourceStrategy)
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

//nolint:gosec // Test Credentials.
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
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: clients: client 'good_id': option 'grant_types' must only have the values 'authorization_code', 'implicit', 'client_credentials', 'refresh_token', 'urn:ietf:params:oauth:grant-type:device_code', 'urn:ietf:params:oauth:grant-type:token-exchange', 'urn:ietf:params:oauth:grant-type:jwt-bearer', or 'urn:ietf:params:oauth:grant-type:jwt-dpop' but the values 'bad_grant_type' are present")
}

//nolint:gosec // Test Credentials.
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

//nolint:gosec // Test Credentials.
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

func TestValidateIdentityProvidersOpenIDConnectTokenExchangeMaxActorChainDepth(t *testing.T) {
	testCases := []struct {
		name     string
		have     int
		expected int
		errors   []string
	}{
		{
			"ShouldNotOverrideCustomValue",
			3,
			3,
			nil,
		},
		{
			"ShouldSetDefault",
			0,
			8,
			nil,
		},
		{
			"ShouldRaiseErrorOnNegative",
			-1,
			-1,
			[]string{"identity_providers: oidc: token_exchange: option 'max_actor_chain_depth' must be greater than or equal to 1 but it's configured as '-1'"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret:       "abc",
						IssuerPrivateKey: keyRSA2048,
						TokenExchange: schema.IdentityProvidersOpenIDConnectTokenExchange{
							MaxActorChainDepth: tc.have,
						},
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

			assert.Equal(t, tc.expected, config.IdentityProviders.OIDC.TokenExchange.MaxActorChainDepth)
			assert.Len(t, validator.Warnings(), 0)

			require.Len(t, validator.Errors(), len(tc.errors))

			for i, expected := range tc.errors {
				assert.EqualError(t, validator.Errors()[i], expected)
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

		require.Len(t, validator.Warnings(), 1)
		assert.EqualError(t, validator.Warnings()[0], "identity_providers: oidc: clients: client 'owncloud': option 'redirect_uris' has the redirect uri 'http://www.mywebsite.com' which uses the 'http' scheme with a host which is not a loopback address or a 'localhost' name, authorization requests using this redirect uri will be rejected and the 'https' scheme should be used instead")
		assert.Len(t, validator.Errors(), 0)
	})

	t.Run("not public", func(t *testing.T) {
		validator := schema.NewStructValidator()
		have.Clients[0].Public = false
		validateOIDCClientRedirectURIs(0, have, validator, nil)

		require.Len(t, validator.Warnings(), 1)
		assert.EqualError(t, validator.Warnings()[0], "identity_providers: oidc: clients: client 'owncloud': option 'redirect_uris' has the redirect uri 'http://www.mywebsite.com' which uses the 'http' scheme with a host which is not a loopback address or a 'localhost' name, authorization requests using this redirect uri will be rejected and the 'https' scheme should be used instead")
		assert.Len(t, validator.Errors(), 1)
		assert.ElementsMatch(t, validator.Errors(), []error{
			errors.New("identity_providers: oidc: clients: client 'owncloud': option 'redirect_uris' has the redirect uri 'urn:ietf:wg:oauth:2.0:oob' when option 'public' is false but this is invalid as this uri is not valid for the openid connect confidential client type"),
		})
	})
}

func TestValidateOIDCClientRefreshTokenRotation(t *testing.T) {
	testCases := []struct {
		name     string
		public   bool
		disable  bool
		warnings []string
	}{
		{name: "ShouldNotWarnRotatedConfidentialClient"},
		{name: "ShouldNotWarnRotatedPublicClient", public: true},
		{name: "ShouldNotWarnConfidentialClientWithoutRotation", disable: true},
		{
			name:    "ShouldWarnPublicClientWithoutRotation",
			public:  true,
			disable: true,
			warnings: []string{
				"identity_providers: oidc: clients: client 'test': option 'disable_refresh_token_rotation' has no effect when option 'public' is true as the refresh tokens of public clients are always rotated so that the reuse of a refresh token can be detected",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{ID: "test", Public: tc.public, DisableRefreshTokenRotation: tc.disable},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCClientRefreshTokenRotation(0, config, validator)

			require.Len(t, validator.Warnings(), len(tc.warnings))

			for i, err := range tc.warnings {
				assert.EqualError(t, validator.Warnings()[i], err)
			}

			assert.Len(t, validator.Errors(), 0)
		})
	}
}

func TestValidateOIDCClientApplicationType(t *testing.T) {
	testCases := []struct {
		name            string
		applicationType string
		public          bool
		redirectURIs    []string
		grantTypes      []string
		expected        string
		warnings        []string
		errs            []string
	}{
		{
			name:         "ShouldDefaultConfidentialClientToWeb",
			redirectURIs: []string{"http://127.0.0.1/callback"},
			grantTypes:   []string{oidc.GrantTypeAuthorizationCode},
			expected:     oidc.ApplicationTypeWeb,
			errs: []string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has the redirect uri 'http://127.0.0.1/callback' which uses the 'http' scheme with a loopback address or a 'localhost' name but this is only permitted when option 'application_type' is 'native' but it's configured as 'web', the 'https' scheme should be used instead",
			},
		},
		{
			name:            "ShouldRaiseErrorOnWebLocalhost",
			applicationType: oidc.ApplicationTypeWeb,
			public:          true,
			redirectURIs:    []string{"http://localhost:8080/callback"},
			grantTypes:      []string{oidc.GrantTypeAuthorizationCode},
			expected:        oidc.ApplicationTypeWeb,
			errs: []string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has the redirect uri 'http://localhost:8080/callback' which uses the 'http' scheme with a loopback address or a 'localhost' name but this is only permitted when option 'application_type' is 'native' but it's configured as 'web', the 'https' scheme should be used instead",
			},
		},
		{
			name:         "ShouldDefaultPublicClientWithHTTPSToWeb",
			public:       true,
			redirectURIs: []string{"https://app.example.com/callback"},
			grantTypes:   []string{oidc.GrantTypeAuthorizationCode},
			expected:     oidc.ApplicationTypeWeb,
		},
		{
			name:       "ShouldDefaultPublicClientWithoutRedirectURIsToWeb",
			public:     true,
			grantTypes: []string{oidc.GrantTypeDeviceCode},
			expected:   oidc.ApplicationTypeWeb,
		},
		{
			name:         "ShouldDefaultPublicClientWithLoopbackToNative",
			public:       true,
			redirectURIs: []string{"https://app.example.com/callback", "http://127.0.0.1/callback"},
			grantTypes:   []string{oidc.GrantTypeAuthorizationCode},
			expected:     oidc.ApplicationTypeNative,
		},
		{
			name:         "ShouldDefaultPublicClientWithLocalhostToNativeWithoutWarning",
			public:       true,
			redirectURIs: []string{"http://localhost:8080/callback"},
			grantTypes:   []string{oidc.GrantTypeAuthorizationCode},
			expected:     oidc.ApplicationTypeNative,
		},
		{
			name:         "ShouldDefaultPublicClientWithPrivateUseSchemeToNativeWithoutWarning",
			public:       true,
			redirectURIs: []string{"myapp://callback"},
			grantTypes:   []string{oidc.GrantTypeAuthorizationCode},
			expected:     oidc.ApplicationTypeNative,
		},
		{
			name:         "ShouldDefaultPublicClientWithInstalledAppToNative",
			public:       true,
			redirectURIs: []string{oidc.RedirectURISpecialOAuth2InstalledApp},
			grantTypes:   []string{oidc.GrantTypeAuthorizationCode},
			expected:     oidc.ApplicationTypeNative,
		},
		{
			name:         "ShouldDefaultPublicClientWithNonLoopbackHTTPToWeb",
			public:       true,
			redirectURIs: []string{"http://app.example.com/callback", "::invalid"},
			grantTypes:   []string{oidc.GrantTypeAuthorizationCode},
			expected:     oidc.ApplicationTypeWeb,
		},
		{
			name:         "ShouldDefaultPublicClientWithImplicitGrantToWeb",
			public:       true,
			redirectURIs: []string{"http://127.0.0.1/callback"},
			grantTypes:   []string{oidc.GrantTypeImplicit},
			expected:     oidc.ApplicationTypeWeb,
			errs: []string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has the redirect uri 'http://127.0.0.1/callback' which uses the 'http' scheme with a loopback address or a 'localhost' name but this is only permitted when option 'application_type' is 'native' but it's configured as 'web', the 'https' scheme should be used instead",
			},
		},
		{
			name:            "ShouldAllowNativeLoopbackAndReverseDNSScheme",
			applicationType: oidc.ApplicationTypeNative,
			redirectURIs:    []string{"http://127.0.0.1/callback", "http://[::1]/callback", "com.example.app:/callback", "https://app.example.com/callback", "http://app.example.com/callback"},
			grantTypes:      []string{oidc.GrantTypeAuthorizationCode},
			expected:        oidc.ApplicationTypeNative,
		},
		{
			name:            "ShouldRaiseErrorOnNativeImplicitGrant",
			applicationType: oidc.ApplicationTypeNative,
			redirectURIs:    []string{"http://127.0.0.1/callback"},
			grantTypes:      []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit},
			expected:        oidc.ApplicationTypeNative,
			errs: []string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' must not include 'implicit' when option 'application_type' is 'native' as the implicit flow can not be protected by PKCE but it's included and the response types are 'code id_token'",
			},
		},
		{
			name:            "ShouldWarnNativeLocalhost",
			applicationType: oidc.ApplicationTypeNative,
			redirectURIs:    []string{"http://localhost:8080/callback"},
			grantTypes:      []string{oidc.GrantTypeAuthorizationCode},
			expected:        oidc.ApplicationTypeNative,
			warnings: []string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has the redirect uri 'http://localhost:8080/callback' which uses the 'http' scheme with the host 'localhost' but it should use the loopback ip address '127.0.0.1' or '[::1]' when option 'application_type' is 'native' as the name may not resolve to the loopback interface the application listens on",
			},
		},
		{
			name:            "ShouldWarnNativeNonReverseDNSScheme",
			applicationType: oidc.ApplicationTypeNative,
			redirectURIs:    []string{"myapp://callback"},
			grantTypes:      []string{oidc.GrantTypeAuthorizationCode},
			expected:        oidc.ApplicationTypeNative,
			warnings: []string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has the redirect uri 'myapp://callback' which uses the private-use uri scheme 'myapp' but it should be a domain name under the control of the application expressed in reverse order such as 'com.example.app' when option 'application_type' is 'native' so that it does not collide with the scheme of another application",
			},
		},
		{
			name:            "ShouldWarnWebPrivateUseScheme",
			applicationType: oidc.ApplicationTypeWeb,
			redirectURIs:    []string{"com.example.app:/callback"},
			grantTypes:      []string{oidc.GrantTypeAuthorizationCode},
			expected:        oidc.ApplicationTypeWeb,
			warnings: []string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has the redirect uri 'com.example.app:/callback' which uses the private-use uri scheme 'com.example.app' but only native applications receive authorization responses this way so option 'application_type' should be configured as 'native'",
			},
		},
		{
			name:         "ShouldWarnDefaultedConfidentialPrivateUseScheme",
			redirectURIs: []string{"com.example.app:/callback"},
			grantTypes:   []string{oidc.GrantTypeAuthorizationCode},
			expected:     oidc.ApplicationTypeWeb,
			warnings: []string{
				"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has the redirect uri 'com.example.app:/callback' which uses the private-use uri scheme 'com.example.app' but only native applications receive authorization responses this way so option 'application_type' should be configured as 'native'",
			},
		},
		{
			name:            "ShouldRaiseErrorOnInvalidApplicationTypeOnly",
			applicationType: "desktop",
			redirectURIs:    []string{"myapp://callback"},
			grantTypes:      []string{oidc.GrantTypeImplicit},
			expected:        "desktop",
			errs: []string{
				"identity_providers: oidc: clients: client 'test': option 'application_type' must be one of 'web' or 'native' but it's configured as 'desktop'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:              "test",
						Public:          tc.public,
						ApplicationType: tc.applicationType,
						RedirectURIs:    tc.redirectURIs,
						ResponseTypes:   []string{oidc.ResponseTypeHybridFlowIDToken},
						GrantTypes:      tc.grantTypes,
					},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCClientApplicationType(0, config, validator)

			assert.Equal(t, tc.expected, config.Clients[0].ApplicationType)

			require.Len(t, validator.Warnings(), len(tc.warnings))

			for i, err := range tc.warnings {
				assert.EqualError(t, validator.Warnings()[i], err)
			}

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, err := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], err)
			}
		})
	}
}

func TestValidateOIDCClientRedirectURIsAuthorizationResponseParameters(t *testing.T) {
	testCases := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			"ShouldNotErrorWithoutQuery",
			"https://app.example.com/callback",
			"",
		},
		{
			"ShouldNotErrorWithOtherQuery",
			"https://app.example.com/callback?tenant=abc",
			"",
		},
		{
			"ShouldErrorWithState",
			"https://app.example.com/callback?state=abc",
			"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has an invalid value: redirect uri 'https://app.example.com/callback?state=abc' must not include the authorization response parameter 'state' in the query",
		},
		{
			"ShouldErrorWithIssuer",
			"https://app.example.com/callback?tenant=abc&iss=https%3A%2F%2Fauth.example.com",
			"identity_providers: oidc: clients: client 'test': option 'redirect_uris' has an invalid value: redirect uri 'https://app.example.com/callback?tenant=abc&iss=https%3A%2F%2Fauth.example.com' must not include the authorization response parameter 'iss' in the query",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "test",
						RedirectURIs: []string{tc.uri},
					},
				},
			}

			validateOIDCClientRedirectURIs(0, config, validator, nil)

			assert.Len(t, validator.Warnings(), 0)

			if tc.expected == "" {
				assert.Len(t, validator.Errors(), 0)
			} else {
				require.Len(t, validator.Errors(), 1)
				assert.EqualError(t, validator.Errors()[0], tc.expected)
			}
		})
	}
}

func TestValidateOIDCClientRedirectURIsInsecure(t *testing.T) {
	testCases := []struct {
		Name     string
		URI      string
		Public   bool
		Expected string
	}{
		{"ShouldNotWarnHTTPS", "https://app.example.com/callback", false, ""},
		{"ShouldNotWarnLocalhost", "http://localhost:8080/callback", false, ""},
		{"ShouldNotWarnLocalhostSubdomain", "http://app.localhost/callback", false, ""},
		{"ShouldNotWarnLoopbackIPv4", "http://127.0.0.1:8080/callback", true, ""},
		{"ShouldNotWarnLoopbackIPv6", "http://[::1]:8080/callback", true, ""},
		{"ShouldNotWarnPrivateUseScheme", "com.example.app:/callback", true, ""},
		{"ShouldWarnConfidential", "http://app.example.com/callback", false, "identity_providers: oidc: clients: client 'test': option 'redirect_uris' has the redirect uri 'http://app.example.com/callback' which uses the 'http' scheme with a host which is not a loopback address or a 'localhost' name, authorization requests using this redirect uri will be rejected and the 'https' scheme should be used instead"},
		{"ShouldWarnPublic", "http://192.168.1.10/callback", true, "identity_providers: oidc: clients: client 'test': option 'redirect_uris' has the redirect uri 'http://192.168.1.10/callback' which uses the 'http' scheme with a host which is not a loopback address or a 'localhost' name, authorization requests using this redirect uri will be rejected and the 'https' scheme should be used instead"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:           "test",
						Public:       tc.Public,
						RedirectURIs: []string{tc.URI},
					},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCClientRedirectURIs(0, config, validator, nil)

			assert.Len(t, validator.Errors(), 0)

			if tc.Expected == "" {
				assert.Len(t, validator.Warnings(), 0)
			} else {
				require.Len(t, validator.Warnings(), 1)
				assert.EqualError(t, validator.Warnings()[0], tc.Expected)
			}
		})
	}
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

	testCases := []struct {
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
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowBoth, oidc.ResponseTypeHybridFlowBoth},
				[]string{oidc.ResponseModeFormPost},
				[]string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'response_modes' does not include 'query' which is the default response mode for the configured response types 'code' so authorization requests using these response types which omit the 'response_mode' parameter will be rejected",
				"identity_providers: oidc: clients: client 'test': option 'response_modes' does not include 'fragment' which is the default response mode for the configured response types 'id_token token' and 'code id_token token' so authorization requests using these response types which omit the 'response_mode' parameter will be rejected",
			},
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
				"identity_providers: oidc: clients: client 'test': option 'scopes' only expects the values 'openid', 'email', 'profile', 'address', 'phone', 'groups', 'offline_access', 'offline', 'authelia.bearer.authz', or 'bound_key' but the unknown values 'group' are present and should generally only be used if a particular client requires a scope outside of our standard scopes",
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
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'response_modes' does not include 'query' which is the default response mode for the configured response types 'code' so authorization requests using these response types which omit the 'response_mode' parameter will be rejected",
			},
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
				"identity_providers: oidc: clients: client 'test': option 'grant_types' must only have the values 'authorization_code', 'implicit', 'client_credentials', 'refresh_token', 'urn:ietf:params:oauth:grant-type:device_code', 'urn:ietf:params:oauth:grant-type:token-exchange', 'urn:ietf:params:oauth:grant-type:jwt-bearer', or 'urn:ietf:params:oauth:grant-type:jwt-dpop' but the values 'invalid' are present",
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
			"ShouldRaiseErrorOnTokenExchangeGrantTypeForPublicClient",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].Public = true
				have.Clients[0].Secret = nil
				have.Clients[0].Scopes = []string{oidc.ScopeOpenID}
				have.TokenExchange = tokenExchangeTestDefaults()
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				[]string{oidc.GrantTypeTokenExchange},
			},
			tcv{
				[]string{oidc.ScopeOpenID},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeTokenExchange},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' should only have the 'urn:ietf:params:oauth:grant-type:token-exchange' value if it is of the confidential client type but it's of the public client type",
			},
		},
		{
			"ShouldRaiseErrorOnTokenExchangeGrantTypeForConfidentialClientWithoutAuthentication",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].Public = false
				have.Clients[0].TokenEndpointAuthMethod = oidc.ClientAuthMethodNone
				have.Clients[0].RevocationEndpointAuthMethod = oidc.ClientAuthMethodNone
				have.Clients[0].IntrospectionEndpointAuthMethod = oidc.ClientAuthMethodNone
				have.Clients[0].PushedAuthorizationRequestEndpointAuthMethod = oidc.ClientAuthMethodNone
				have.Clients[0].Secret = nil
				have.Clients[0].Scopes = []string{oidc.ScopeOpenID}
				have.TokenExchange = tokenExchangeTestDefaults()
			},
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeImplicitFlowIDToken},
				nil,
				[]string{oidc.GrantTypeImplicit, oidc.GrantTypeTokenExchange},
			},
			tcv{
				[]string{oidc.ScopeOpenID},
				[]string{oidc.ResponseTypeImplicitFlowIDToken},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeImplicit, oidc.GrantTypeTokenExchange},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' should only have the 'urn:ietf:params:oauth:grant-type:token-exchange' value if the client authenticates at the token endpoint but option 'token_endpoint_auth_method' is configured as 'none'",
			},
		},
		{
			"ShouldNotRaiseErrorOnTokenExchangeGrantTypeForConfidentialClient",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].Public = false
				have.Clients[0].Scopes = []string{oidc.ScopeOpenID}
				have.TokenExchange = tokenExchangeTestDefaults()
			},
			nil,
			tcv{
				nil,
				nil,
				nil,
				[]string{oidc.GrantTypeTokenExchange},
			},
			tcv{
				[]string{oidc.ScopeOpenID},
				[]string{oidc.ResponseTypeAuthorizationCodeFlow},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery},
				[]string{oidc.GrantTypeTokenExchange},
			},
			nil,
			nil,
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
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' must include 'implicit' when option 'response_types' includes 'id_token token'",
			},
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
				have.Clients[0].Audience = []string{"https://app.example.com"}
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
				have.Clients[0].Audience = []string{"https://app.example.com"}
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
				have.Clients[0].Audience = []string{"https://app.example.com"}
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
				have.Clients[0].Audience = []string{"https://app.example.com"}
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
				have.Clients[0].Audience = []string{"https://app.example.com"}
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
					RedirectURIs:                       []string{"https://localhost"},
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
				"identity_providers: oidc: clients: client 'abc': option 'audience' or option 'resource' must be configured when configured with scope 'authelia.bearer.authz' but both are absent",
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
					RedirectURIs:                       []string{"https://localhost"},
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
				"identity_providers: oidc: clients: client 'abc': option 'audience' or option 'resource' must be configured when configured with scope 'authelia.bearer.authz' but both are absent",
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
					RedirectURIs:                       []string{"https://localhost"},
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
				"identity_providers: oidc: clients: client 'abc': option 'audience' or option 'resource' must be configured when configured with scope 'authelia.bearer.authz' but both are absent",
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
				"identity_providers: oidc: clients: client 'abc': option 'audience' or option 'resource' must be configured when configured with scope 'authelia.bearer.authz' but both are absent",
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
					RedirectURIs:                       []string{"https://localhost"},
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
			"ShouldHandleBearerValidWithOnlyResource",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0] = schema.IdentityProvidersOpenIDConnectClient{
					ID:                                 "abc",
					Secret:                             nil,
					Public:                             true,
					RedirectURIs:                       []string{"http://localhost"},
					Resource:                           []string{"https://app.example.com/api"},
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
				"identity_providers: oidc: clients: client 'test': option 'token_endpoint_auth_signing_alg' must be one of 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', 'ES512', 'Ed25519', 'EdDSA', 'ML-DSA-44', 'ML-DSA-65', or 'ML-DSA-87' when option 'token_endpoint_auth_method' is configured to 'private_key_jwt'",
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
				have.Clients[0].Audience = []string{"https://app.example.com"}
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
				have.Clients[0].Audience = []string{"https://app.example.com"}
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
				have.Clients[0].Audience = []string{"https://app.example.com"}

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
			"ShouldRaiseErrorOnRSA15EncryptionAlgWithoutRSA15KeyManagement",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].JSONWebKeysURI = &url.URL{Scheme: "https", Host: "example.com", Path: "/jwks.json"}

				have.Clients[0].IDTokenEncryptedResponseAlg = oidc.EncryptionAlgRSA15
				have.Clients[0].IDTokenEncryptedResponseEnc = ""
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.EncryptionAlgRSA15, have.Clients[0].IDTokenEncryptedResponseAlg)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].IDTokenEncryptedResponseEnc)
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
				"identity_providers: oidc: clients: client 'test': option 'id_token_encrypted_response_alg' is configured as 'RSA1_5' but the 'enable_rsa1_5_key_management' option is not enabled",
			},
		},
		{
			"ShouldNotRaiseErrorOnRSA15EncryptionAlgWithRSA15KeyManagement",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.EnableRSA15KeyManagement = true
				have.Clients[0].JSONWebKeysURI = &url.URL{Scheme: "https", Host: "example.com", Path: "/jwks.json"}

				have.Clients[0].IDTokenEncryptedResponseAlg = oidc.EncryptionAlgRSA15
				have.Clients[0].IDTokenEncryptedResponseEnc = ""
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.EncryptionAlgRSA15, have.Clients[0].IDTokenEncryptedResponseAlg)
				assert.Equal(t, oidc.EncryptionEncA128CBCHS256, have.Clients[0].IDTokenEncryptedResponseEnc)
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
			"ShouldRaiseWarningOnAccessTokenJWTWithoutAudience",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].AccessTokenSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
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
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'audience' should be configured when option 'access_token_signed_response_alg' is configured to 'RS256' as RFC9068 requires the 'aud' claim in a JWT Profile Access Token and it is otherwise absent, which will cause a resource server to reject the token; the audience must also be requested, either by the client via the 'audience' parameter or by configuring option 'requested_audience_mode' to 'implicit'",
			},
			nil,
		},
		{
			"ShouldNotRaiseWarningOnAccessTokenJWTWithAudience",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].AccessTokenSignedResponseAlg = oidc.SigningAlgRSAUsingSHA256
				have.Clients[0].Audience = []string{"https://app.example.com"}
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
			"ShouldNotRaiseWarningOnOpaqueAccessToken",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].AccessTokenSignedResponseAlg = oidc.SigningAlgNone
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
				have.Clients[0].Audience = []string{"https://app.example.com"}

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
				have.Clients[0].Audience = []string{"https://app.example.com"}
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
		{
			"ShouldRaiseErrorOnHybridResponseTypeWithoutImplicitGrantType",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeHybridFlowIDToken},
				nil,
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeHybridFlowIDToken},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeAuthorizationCode},
			},
			nil,
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' must include 'implicit' when option 'response_types' includes 'code id_token'",
			},
		},
		{
			"ShouldRaiseErrorOnHybridResponseTypeWithoutAnyRequiredGrantType",
			nil,
			nil,
			tcv{
				nil,
				[]string{oidc.ResponseTypeHybridFlowToken},
				nil,
				[]string{oidc.GrantTypeRefreshToken},
			},
			tcv{
				[]string{oidc.ScopeOpenID, oidc.ScopeGroups, oidc.ScopeProfile, oidc.ScopeEmail},
				[]string{oidc.ResponseTypeHybridFlowToken},
				[]string{oidc.ResponseModeFormPost, oidc.ResponseModeFragment},
				[]string{oidc.GrantTypeRefreshToken},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' should only have the 'refresh_token' value if the client is also configured with the 'offline_access' scope",
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'grant_types' must include 'authorization_code' and 'implicit' when option 'response_types' includes 'code token'",
			},
		},
		{
			"ShouldSetDefaultApplicationType",
			nil,
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.ApplicationTypeWeb, have.Clients[0].ApplicationType)
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
			"ShouldAllowApplicationTypeNative",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].ApplicationType = oidc.ApplicationTypeNative
			},
			func(t *testing.T, have *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, oidc.ApplicationTypeNative, have.Clients[0].ApplicationType)
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
			"ShouldRaiseErrorOnInvalidApplicationType",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].ApplicationType = "desktop"
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
				"identity_providers: oidc: clients: client 'test': option 'application_type' must be one of 'web' or 'native' but it's configured as 'desktop'",
			},
		},
		{
			"ShouldRaiseErrorOnImplicitAudienceWithWildcardStrategy",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.AudienceStrategy = oidc.StrategyWildcard
				have.Clients[0].RequestedAudienceMode = oidc.ClientRequestedAudienceModeImplicit.String()
				have.Clients[0].Audience = []string{"https://api.example.com/*", "https://app.example.com"}
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
				"identity_providers: oidc: clients: client 'test': option 'audience' must not contain wildcard values when option 'requested_audience_mode' is 'implicit' and option 'audience_strategy' is 'wildcard' as the values are granted literally but it has the wildcard values 'https://api.example.com/*'",
			},
		},
		{
			"ShouldRaiseErrorOnImplicitResourceWithClientWildcardStrategy",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.Clients[0].ResourceStrategy = oidc.StrategyWildcard
				have.Clients[0].RequestedResourceMode = oidc.ClientRequestedResourceModeImplicit.String()
				have.Clients[0].Resource = []string{"https://api.example.com/*"}
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
				"identity_providers: oidc: clients: client 'test': option 'resource' must not contain wildcard values when option 'requested_resource_mode' is 'implicit' and option 'resource_strategy' is 'wildcard' as the values are granted literally but it has the wildcard values 'https://api.example.com/*'",
			},
		},
		{
			"ShouldAllowImplicitAudienceWildcardWithExactStrategy",
			func(have *schema.IdentityProvidersOpenIDConnect) {
				have.AudienceStrategy = oidc.StrategyExact
				have.Clients[0].RequestedAudienceMode = oidc.ClientRequestedAudienceModeImplicit.String()
				have.Clients[0].Audience = []string{"https://api.example.com/*"}
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
	}

	errDeprecatedFunc := func() {}

	for _, tc := range testCases {
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

func TestValidateOIDCClientRequireSignedRequestObject(t *testing.T) {
	testCases := []struct {
		name     string
		provider bool
		client   bool
		alg      string
		expected string
	}{
		{"ShouldNotErrorWhenNotRequired", false, false, oidc.SigningAlgNone, ""},
		{"ShouldNotErrorWhenClientRequiredWithSignedAlg", false, true, oidc.SigningAlgRSAUsingSHA256, ""},
		{"ShouldNotErrorWhenProviderRequiredWithSignedAlg", true, false, oidc.SigningAlgRSAUsingSHA256, ""},
		{"ShouldNotErrorWhenClientRequiredWithEmptyAlg", false, true, "", ""},
		{"ShouldErrorWhenClientRequiredWithAlgNone", false, true, oidc.SigningAlgNone, "identity_providers: oidc: clients: client 'test': option 'request_object_signing_alg' must not be configured as 'none' when the client option 'require_signed_request_object' is 'true' as an unsigned request object can never satisfy this requirement"},
		{"ShouldErrorWhenProviderRequiredWithAlgNone", true, false, oidc.SigningAlgNone, "identity_providers: oidc: clients: client 'test': option 'request_object_signing_alg' must not be configured as 'none' when the provider option 'require_signed_request_object' is 'true' as an unsigned request object can never satisfy this requirement"},
		{"ShouldErrorOnceWhenBothRequiredWithAlgNone", true, true, oidc.SigningAlgNone, "identity_providers: oidc: clients: client 'test': option 'request_object_signing_alg' must not be configured as 'none' when the client option 'require_signed_request_object' is 'true' as an unsigned request object can never satisfy this requirement"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				RequireSignedRequestObject: tc.provider,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                         "test",
						RequireSignedRequestObject: tc.client,
						RequestObjectSigningAlg:    tc.alg,
					},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCClientRequireSignedRequestObject(0, config, validator)

			assert.Len(t, validator.Warnings(), 0)

			if tc.expected == "" {
				assert.Len(t, validator.Errors(), 0)
			} else {
				require.Len(t, validator.Errors(), 1)
				assert.EqualError(t, validator.Errors()[0], tc.expected)
			}
		})
	}
}

func TestValidateOIDCClientTokenEndpointAuthMethod(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
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
				"identity_providers: oidc: clients: client 'test': jwks: key #1 with key id 'test': option 'algorithm' must be one of 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', 'ES512', 'Ed25519', 'EdDSA', 'ML-DSA-44', 'ML-DSA-65', or 'ML-DSA-87' but it's configured as 'bad'",
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
				"identity_providers: oidc: jwks: key #1 with key id 'c4c7ca-invalid': option 'algorithm' must be one of 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', 'ES512', 'Ed25519', 'EdDSA', 'ML-DSA-44', 'ML-DSA-65', or 'ML-DSA-87' but it's configured as 'invalid'",
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
			"ShouldHandleEd25519Keys",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyEd25519, CertificateChain: certEd25519},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyEd25519, CertificateChain: certEd25519, Algorithm: oidc.SigningAlgEd25519, Use: oidc.KeyUseSignature, KeyID: "ca54bd-ed25519"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{oidc.SigningAlgEd25519: "ca54bd-ed25519"},
					DefaultEncKeyIDs: map[string]string{},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: keys: must at least have one key supporting the 'RS256' algorithm but only has 'Ed25519' and 'EdDSA'",
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
				"identity_providers: oidc: jwks: key #1 with key id '35db6c-rs256': option 'key' must be an RSA private key, ECDSA private key, Ed25519 private key, or ML-DSA private key but it's type is *rsa.PublicKey",
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
			"ShouldRaiseErrorOnEncKeyRSA15WithoutRSA15KeyManagement",
			&schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, Use: oidc.KeyUseEncryption, Algorithm: oidc.EncryptionAlgRSA15},
					{Key: keyRSA2048, Use: oidc.KeyUseSignature},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, Use: oidc.KeyUseEncryption, Algorithm: oidc.EncryptionAlgRSA15, KeyID: "35db6c-rsa1_5"},
					{Key: keyRSA2048, Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, KeyID: "35db6c-rs256"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{"RS256": "35db6c-rs256"},
					DefaultEncKeyIDs: map[string]string{"RSA1_5": "35db6c-rsa1_5"},
				},
			},
			[]string{
				"identity_providers: oidc: jwks: key #1 with key id '35db6c-rsa1_5': option 'algorithm' is configured as 'RSA1_5' but the 'enable_rsa1_5_key_management' option is not enabled",
			},
		},
		{
			"ShouldHandleEncKeyRSA15WithRSA15KeyManagement",
			&schema.IdentityProvidersOpenIDConnect{
				EnableRSA15KeyManagement: true,
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, Use: oidc.KeyUseEncryption, Algorithm: oidc.EncryptionAlgRSA15},
					{Key: keyRSA2048, Use: oidc.KeyUseSignature},
				},
			},
			schema.IdentityProvidersOpenIDConnect{
				EnableRSA15KeyManagement: true,
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, Use: oidc.KeyUseEncryption, Algorithm: oidc.EncryptionAlgRSA15, KeyID: "35db6c-rsa1_5"},
					{Key: keyRSA2048, Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, KeyID: "35db6c-rs256"},
				},
				Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
					DefaultSigKeyIDs: map[string]string{"RS256": "35db6c-rs256"},
					DefaultEncKeyIDs: map[string]string{"RSA1_5": "35db6c-rsa1_5"},
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

func TestValidateOIDCIssuerEdwardsAlgorithms(t *testing.T) {
	testCases := []struct {
		name     string
		alg      string
		expected string
		algs     []string
		warns    []string
	}{
		{
			name:     "ShouldAdvertiseBothIdentifiersForAnEd25519Key",
			alg:      oidc.SigningAlgEd25519,
			expected: oidc.SigningAlgEd25519,
			algs:     []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgEd25519, oidc.SigningAlgEdDSA},
			warns:    nil,
		},
		{
			name:     "ShouldAdvertiseBothIdentifiersForAnEdDSAKeyAndWarn",
			alg:      oidc.SigningAlgEdDSA,
			expected: oidc.SigningAlgEdDSA,
			algs:     []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgEdDSA, oidc.SigningAlgEd25519},
			warns: []string{
				"identity_providers: oidc: jwks: key #2 with key id 'ca54bd-eddsa': option 'algorithm' is configured as 'EdDSA' which RFC 9864 has deprecated in the IANA registry as it does not identify the curve in use, it's recommended to configure it as 'Ed25519' instead",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			have := &schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
					{Key: keyEd25519, Algorithm: tc.alg, Use: oidc.KeyUseSignature},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCIssuer(have, validator)

			assert.Len(t, validator.Errors(), 0)
			assert.Equal(t, tc.expected, have.JSONWebKeys[1].Algorithm)
			assert.Equal(t, tc.algs, have.Discovery.ResponseObjectSigningAlgs)

			require.Len(t, validator.Warnings(), len(tc.warns))

			for i := range tc.warns {
				assert.EqualError(t, validator.Warnings()[i], tc.warns[i])
			}
		})
	}
}

func TestValidateOIDCClientEdwardsAlgorithms(t *testing.T) {
	testCases := []struct {
		name  string
		alg   string
		warns []string
	}{
		{
			name:  "ShouldAcceptEd25519",
			alg:   oidc.SigningAlgEd25519,
			warns: nil,
		},
		{
			name: "ShouldAcceptEdDSAAgainstAnEd25519KeyAndWarn",
			alg:  oidc.SigningAlgEdDSA,
			warns: []string{
				"identity_providers: oidc: clients: client 'test': option 'id_token_signed_response_alg' is configured as 'EdDSA' which RFC 9864 has deprecated in the IANA registry as it does not identify the curve in use, it's recommended to configure it as 'Ed25519' instead",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			have := &schema.IdentityProvidersOpenIDConnect{
				JSONWebKeys: []schema.JWK{
					{Key: keyRSA2048, Algorithm: oidc.SigningAlgRSAUsingSHA256, Use: oidc.KeyUseSignature},
					{Key: keyEd25519, Algorithm: oidc.SigningAlgEd25519, Use: oidc.KeyUseSignature},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                       "test",
						Secret:                   tOpenIDConnectPBKDF2ClientSecret,
						AuthorizationPolicy:      policyTwoFactor,
						RedirectURIs:             []string{"https://client.example.com"},
						IDTokenSignedResponseAlg: tc.alg,
					},
				},
			}

			have.Discovery.AuthorizationPolicies = []string{policyOneFactor, policyTwoFactor}

			validator := schema.NewStructValidator()

			validateOIDCIssuer(have, validator)

			require.Len(t, validator.Errors(), 0)

			validateOIDCClients(NewValidateCtx(), have, validator)

			assert.Len(t, validator.Errors(), 0)
			assert.Equal(t, tc.alg, have.Clients[0].IDTokenSignedResponseAlg)

			warns := validator.Warnings()

			require.Len(t, warns, len(tc.warns))

			for i := range tc.warns {
				assert.EqualError(t, warns[i], tc.warns[i])
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
		{
			"ShouldHandleRefreshTokenMaximum",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Lifespans: schema.IdentityProvidersOpenIDConnectLifespans{
							RefreshTokenMaximum: time.Hour * 24,
							Custom: map[string]schema.IdentityProvidersOpenIDConnectLifespan{
								"custom": {RefreshTokenMaximum: time.Hour},
							},
						},
					},
				},
			},
			[]string{"custom"},
			nil,
		},
		{
			"ShouldRaiseErrorOnNegativeRefreshTokenMaximum",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Lifespans: schema.IdentityProvidersOpenIDConnectLifespans{
							RefreshTokenMaximum: -time.Hour,
							Custom: map[string]schema.IdentityProvidersOpenIDConnectLifespan{
								"custom": {RefreshTokenMaximum: -time.Minute},
							},
						},
					},
				},
			},
			[]string{"custom"},
			[]string{
				"identity_providers: oidc: lifespans: option 'refresh_token_maximum' must be greater than or equal to 0 but it's configured as '-1h0m0s'",
				"identity_providers: oidc: lifespans: custom: custom: option 'refresh_token_maximum' must be greater than or equal to 0 but it's configured as '-1m0s'",
			},
		},
		{
			"ShouldHandleRefreshTokenGracePeriod",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Lifespans: schema.IdentityProvidersOpenIDConnectLifespans{
							RefreshTokenGracePeriod: time.Second * 10,
							Custom: map[string]schema.IdentityProvidersOpenIDConnectLifespan{
								"custom": {RefreshTokenGracePeriod: time.Second * 5},
							},
						},
					},
				},
			},
			[]string{"custom"},
			nil,
		},
		{
			"ShouldRaiseErrorOnNegativeRefreshTokenGracePeriod",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Lifespans: schema.IdentityProvidersOpenIDConnectLifespans{
							RefreshTokenGracePeriod: -time.Second,
							Custom: map[string]schema.IdentityProvidersOpenIDConnectLifespan{
								"custom": {RefreshTokenGracePeriod: -time.Minute},
							},
						},
					},
				},
			},
			[]string{"custom"},
			[]string{
				"identity_providers: oidc: lifespans: option 'refresh_token_grace_period' must be between 0s and 10s but it's configured as '-1s'",
				"identity_providers: oidc: lifespans: custom: custom: option 'refresh_token_grace_period' must be between 0s and 10s but it's configured as '-1m0s'",
			},
		},
		{
			"ShouldRaiseErrorOnRefreshTokenGracePeriodAboveMaximum",
			&schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Lifespans: schema.IdentityProvidersOpenIDConnectLifespans{
							RefreshTokenGracePeriod: time.Second*10 + time.Millisecond,
							Custom: map[string]schema.IdentityProvidersOpenIDConnectLifespan{
								"custom": {RefreshTokenGracePeriod: time.Minute},
							},
						},
					},
				},
			},
			[]string{"custom"},
			[]string{
				"identity_providers: oidc: lifespans: option 'refresh_token_grace_period' must be between 0s and 10s but it's configured as '10.001s'",
				"identity_providers: oidc: lifespans: custom: custom: option 'refresh_token_grace_period' must be between 0s and 10s but it's configured as '1m0s'",
			},
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

func TestShouldValidateOpenIDConnectClaimsPoliciesDuplicateCustomClaimName(t *testing.T) {
	config := &schema.Configuration{
		AuthenticationBackend: schema.AuthenticationBackend{
			File: &schema.AuthenticationBackendFile{
				ExtraAttributes: map[string]schema.AuthenticationBackendExtraAttribute{
					"my_claim": {ValueType: "string"},
				},
			},
		},
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{
					"example": {
						CustomClaims: schema.IdentityProvidersOpenIDConnectCustomClaims{
							"my_claim": {Attribute: "my_claim"},
							"alias":    {Name: "my_claim", Attribute: "my_claim"},
						},
					},
				},
			},
		},
	}

	val := schema.NewStructValidator()

	validateOIDCClaims(config, val)

	require.Len(t, val.Errors(), 1)

	errStr := val.Errors()[0].Error()

	assert.Contains(t, errStr, "claim with name 'my_claim' is mapped in both")
	assert.Contains(t, errStr, "alias")
	assert.Contains(t, errStr, "my_claim")
	assert.Contains(t, errStr, "claim configurations")
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
							"profile":   {},
							"bound_key": {},
						},
						Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
							Claims: []string{"example"},
						},
					},
				},
			},
			errors: []string{
				"identity_providers: oidc: scopes: scope with name 'authelia.scopes-ex' can't be used as a custom scope because all scopes prefixed with 'authelia.' are reserved",
				"identity_providers: oidc: scopes: scope with name 'bound_key' can't be used as a custom scope because it's a standard scope",
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

func TestValidateOIDCClientSectorIdentifierURI(t *testing.T) {
	testCases := []struct {
		name          string
		setup         func(t *testing.T) (*url.URL, *http.Client)
		redirectURIs  []string
		expectedErrs  int
		expectedWarns int
		errContains   []string
	}{
		{
			"ShouldSucceedWhenRedirectURIsMatch",
			func(t *testing.T) (*url.URL, *http.Client) {
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode([]string{"https://app.example.com/callback"})
				}))

				t.Cleanup(srv.Close)

				sectorURI, err := url.Parse(srv.URL)
				require.NoError(t, err)

				sectorURI.Scheme = schemeHTTPS

				return sectorURI, srv.Client()
			},
			[]string{"https://app.example.com/callback"},
			0,
			0,
			nil,
		},
		{
			"ShouldErrWhenRedirectURIsMismatch",
			func(t *testing.T) (*url.URL, *http.Client) {
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode([]string{"https://other.example.com/callback"})
				}))

				t.Cleanup(srv.Close)

				sectorURI, err := url.Parse(srv.URL)
				require.NoError(t, err)

				sectorURI.Scheme = schemeHTTPS

				return sectorURI, srv.Client()
			},
			[]string{"https://app.example.com/callback"},
			1,
			0,
			[]string{"sector_identifier_uri", "redirect_uri"},
		},
		{
			"ShouldErrWhenServerReturnsInvalidJSON",
			func(t *testing.T) (*url.URL, *http.Client) {
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte("not json"))
				}))

				t.Cleanup(srv.Close)

				sectorURI, err := url.Parse(srv.URL)
				require.NoError(t, err)

				sectorURI.Scheme = schemeHTTPS

				return sectorURI, srv.Client()
			},
			[]string{"https://app.example.com/callback"},
			1,
			0,
			[]string{"sector_identifier_uri"},
		},
		{
			"ShouldErrWhenServerUnreachable",
			func(t *testing.T) (*url.URL, *http.Client) {
				return &url.URL{Scheme: "https", Host: "127.0.0.1:1"}, &http.Client{
					Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: x509.NewCertPool()}},
					Timeout:   100 * time.Millisecond,
				}
			},
			[]string{"https://app.example.com/callback"},
			1,
			0,
			[]string{"sector_identifier_uri"},
		},
		{
			"ShouldSkipWhenSectorIdentifierURINil",
			func(t *testing.T) (*url.URL, *http.Client) {
				return nil, nil
			},
			nil,
			0,
			0,
			nil,
		},
		{
			"ShouldErrWhenSchemeNotHTTPS",
			func(t *testing.T) (*url.URL, *http.Client) {
				return &url.URL{Scheme: "http", Host: "example.com", Path: "/sector"}, nil
			},
			nil,
			1,
			0,
			[]string{"scheme"},
		},
		{
			"ShouldSucceedWithMultipleRedirectURIs",
			func(t *testing.T) (*url.URL, *http.Client) {
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode([]string{
						"https://app.example.com/callback",
						"https://app.example.com/callback2",
					})
				}))

				t.Cleanup(srv.Close)

				sectorURI, err := url.Parse(srv.URL)
				require.NoError(t, err)

				sectorURI.Scheme = "https"

				return sectorURI, srv.Client()
			},
			[]string{"https://app.example.com/callback", "https://app.example.com/callback2"},
			0,
			0,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sectorURI, client := tc.setup(t)

			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  "test-client",
						SectorIdentifierURI: sectorURI,
						RedirectURIs:        tc.redirectURIs,
					},
				},
			}

			val := schema.NewStructValidator()

			ctx := &ValidateCtx{
				Context:                   context.Background(),
				client:                    client,
				cacheSectorIdentifierURIs: map[string][]string{},
			}

			deprecated := false

			validateOIDCClientSectorIdentifier(ctx, 0, config, val, func() {
				deprecated = true
			})

			assert.Len(t, val.Errors(), tc.expectedErrs)
			assert.Len(t, val.Warnings(), tc.expectedWarns)

			assert.False(t, deprecated)

			if len(tc.errContains) > 0 {
				require.GreaterOrEqual(t, len(val.Errors()), 1)

				for _, s := range tc.errContains {
					assert.Contains(t, val.Errors()[0].Error(), s)
				}
			}
		})
	}
}

func TestValidateIdentityProvidersShouldValidateClientDPoPBoundAccessTokens(t *testing.T) {
	testCases := []struct {
		Name    string
		Have    schema.IdentityProvidersOpenIDConnectDPoP
		Clients []schema.IdentityProvidersOpenIDConnectClient
		Errors  []string
	}{
		{
			"ShouldNotRaiseErrorWhenDisabledAndNotBound",
			schema.IdentityProvidersOpenIDConnectDPoP{},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}},
			},
			nil,
		},
		{
			"ShouldNotRaiseErrorWhenEnabledAndBound",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}, DPoPBoundAccessTokens: true},
			},
			nil,
		},
		{
			"ShouldRaiseErrorWhenDisabledAndBound",
			schema.IdentityProvidersOpenIDConnectDPoP{},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}, DPoPBoundAccessTokens: true},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'dpop_bound_access_tokens' can't be enabled when the 'identity_providers: oidc: dpop: enabled' option is disabled"},
		},
		{
			"ShouldRaiseErrorPerClientWhenDisabledAndBound",
			schema.IdentityProvidersOpenIDConnectDPoP{},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}, DPoPBoundAccessTokens: true},
				{ID: "other", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}},
				{ID: "another", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}, DPoPBoundAccessTokens: true},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'dpop_bound_access_tokens' can't be enabled when the 'identity_providers: oidc: dpop: enabled' option is disabled",
				"identity_providers: oidc: clients: client 'another': option 'dpop_bound_access_tokens' can't be enabled when the 'identity_providers: oidc: dpop: enabled' option is disabled",
			},
		},
		{
			"ShouldRaiseErrorWhenEnforcedAndTokenResponseTypes",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, Enforced: true},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}, ResponseTypes: []string{"code", "token", "code id_token token"}, GrantTypes: []string{"authorization_code", "implicit"}},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'response_types' must not have the values 'token' and 'code id_token token' when either the 'identity_providers: oidc: dpop: enforced' option or the client option 'dpop_bound_access_tokens' is enabled as access tokens issued by the authorization endpoint can't be bound to a DPoP key"},
		},
		{
			"ShouldRaiseErrorWhenBoundAndHybridTokenResponseType",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}, DPoPBoundAccessTokens: true, ResponseTypes: []string{"code token"}, GrantTypes: []string{"authorization_code", "implicit"}},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'response_types' must not have the values 'code token' when either the 'identity_providers: oidc: dpop: enforced' option or the client option 'dpop_bound_access_tokens' is enabled as access tokens issued by the authorization endpoint can't be bound to a DPoP key"},
		},
		{
			"ShouldNotRaiseErrorWhenEnforcedAndNoTokenResponseTypes",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, Enforced: true},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}, ResponseTypes: []string{"code", "id_token", "code id_token"}, GrantTypes: []string{"authorization_code", "implicit"}},
			},
			nil,
		},
		{
			"ShouldNotRaiseErrorWhenNotRequiredAndTokenResponseTypes",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}, ResponseTypes: []string{"code", "token"}, GrantTypes: []string{"authorization_code", "implicit"}},
			},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret:  "abcdefghijklmnopqrstuvwxyz123456",
						JSONWebKeys: []schema.JWK{{Key: keyRSA2048, CertificateChain: certRSA2048}},
						DPoP:        tc.Have,
						Clients:     tc.Clients,
					},
				},
			}

			ValidateIdentityProviders(NewValidateCtx(), config, validator)

			errs := validator.Errors()

			require.Len(t, errs, len(tc.Errors))

			for i, expected := range tc.Errors {
				assert.EqualError(t, errs[i], expected)
			}
		})
	}
}

func TestValidateIdentityProvidersShouldValidateClientKeyBindingScope(t *testing.T) {
	testCases := []struct {
		Name     string
		Have     schema.IdentityProvidersOpenIDConnectDPoP
		Clients  []schema.IdentityProvidersOpenIDConnectClient
		Warnings []string
	}{
		{
			"ShouldNotWarnWhenEnabledAndRequested",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, KeyBinding: true},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPBKDF2ClientSecret, RedirectURIs: []string{"https://app.example.com"}, Scopes: []string{oidc.ScopeOpenID, oidc.ScopeBoundKey}},
			},
			nil,
		},
		{
			"ShouldNotWarnWhenDisabledAndNotRequested",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPBKDF2ClientSecret, RedirectURIs: []string{"https://app.example.com"}, Scopes: []string{oidc.ScopeOpenID}},
			},
			nil,
		},
		{
			"ShouldWarnWhenKeyBindingDisabledAndRequested",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPBKDF2ClientSecret, RedirectURIs: []string{"https://app.example.com"}, Scopes: []string{oidc.ScopeOpenID, oidc.ScopeBoundKey}},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'scopes' has the value 'bound_key' which is ineffective when the 'identity_providers: oidc: dpop: key_binding' option is disabled"},
		},
		{
			"ShouldWarnWhenDPoPDisabledAndRequested",
			schema.IdentityProvidersOpenIDConnectDPoP{},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPBKDF2ClientSecret, RedirectURIs: []string{"https://app.example.com"}, Scopes: []string{oidc.ScopeOpenID, oidc.ScopeBoundKey}},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'scopes' has the value 'bound_key' which is ineffective when the 'identity_providers: oidc: dpop: key_binding' option is disabled"},
		},
		{
			"ShouldWarnPerClientWhenDisabledAndRequested",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true},
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", Secret: tOpenIDConnectPBKDF2ClientSecret, RedirectURIs: []string{"https://app.example.com"}, Scopes: []string{oidc.ScopeOpenID, oidc.ScopeBoundKey}},
				{ID: "other", Secret: tOpenIDConnectPBKDF2ClientSecret, RedirectURIs: []string{"https://app.example.com"}, Scopes: []string{oidc.ScopeOpenID}},
				{ID: "another", Secret: tOpenIDConnectPBKDF2ClientSecret, RedirectURIs: []string{"https://app.example.com"}, Scopes: []string{oidc.ScopeOpenID, oidc.ScopeBoundKey}},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'scopes' has the value 'bound_key' which is ineffective when the 'identity_providers: oidc: dpop: key_binding' option is disabled",
				"identity_providers: oidc: clients: client 'another': option 'scopes' has the value 'bound_key' which is ineffective when the 'identity_providers: oidc: dpop: key_binding' option is disabled",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret:  "abcdefghijklmnopqrstuvwxyz123456",
						JSONWebKeys: []schema.JWK{{Key: keyRSA2048, CertificateChain: certRSA2048}},
						DPoP:        tc.Have,
						Clients:     tc.Clients,
					},
				},
			}

			ValidateIdentityProviders(NewValidateCtx(), config, validator)

			assert.Len(t, validator.Errors(), 0)

			warnings := validator.Warnings()

			require.Len(t, warnings, len(tc.Warnings))

			for i, expected := range tc.Warnings {
				assert.EqualError(t, warnings[i], expected)
			}
		})
	}
}

func TestValidateIdentityProvidersShouldSetDefaultDPoPValues(t *testing.T) {
	testCases := []struct {
		Name     string
		Have     schema.IdentityProvidersOpenIDConnectDPoP
		Expected schema.IdentityProvidersOpenIDConnectDPoP
		Errors   []string
	}{
		{
			"ShouldSetDefaults",
			schema.IdentityProvidersOpenIDConnectDPoP{},
			schema.IdentityProvidersOpenIDConnectDPoP{NonceLifespan: time.Minute * 5, ProofLifespan: time.Second * 10, ClockSkew: time.Second * 30},
			nil,
		},
		{
			"ShouldNotOverrideConfigured",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, NonceLifespan: time.Minute, ProofLifespan: time.Second * 30, ClockSkew: time.Second * 10},
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, NonceLifespan: time.Minute, ProofLifespan: time.Second * 30, ClockSkew: time.Second * 10},
			nil,
		},
		{
			"ShouldRaiseErrorOnNegativeNonceLifespan",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, NonceLifespan: -time.Minute},
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, NonceLifespan: -time.Minute, ProofLifespan: time.Second * 10, ClockSkew: time.Second * 30},
			[]string{"identity_providers: oidc: dpop: option 'nonce_lifespan' must be a positive duration but it's configured as '-1m0s'"},
		},
		{
			"ShouldRaiseErrorOnNegativeProofLifespan",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, ProofLifespan: -time.Second},
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, NonceLifespan: time.Minute * 5, ProofLifespan: -time.Second, ClockSkew: time.Second * 30},
			[]string{"identity_providers: oidc: dpop: option 'proof_lifespan' must be a positive duration but it's configured as '-1s'"},
		},
		{
			"ShouldRaiseErrorOnNegativeClockSkew",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, ClockSkew: -time.Second},
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, NonceLifespan: time.Minute * 5, ProofLifespan: time.Second * 10, ClockSkew: -time.Second},
			[]string{"identity_providers: oidc: dpop: option 'clock_skew' must be a positive duration but it's configured as '-1s'"},
		},
		{
			"ShouldRaiseErrorOnEnforcedWithoutEnabled",
			schema.IdentityProvidersOpenIDConnectDPoP{Enforced: true},
			schema.IdentityProvidersOpenIDConnectDPoP{Enforced: true, NonceLifespan: time.Minute * 5, ProofLifespan: time.Second * 10, ClockSkew: time.Second * 30},
			[]string{"identity_providers: oidc: dpop: option 'enforced' can't be enabled when option 'enabled' is disabled"},
		},
		{
			"ShouldRaiseErrorOnNonceEnforcedWithoutEnabled",
			schema.IdentityProvidersOpenIDConnectDPoP{NonceEnforced: true},
			schema.IdentityProvidersOpenIDConnectDPoP{NonceEnforced: true, NonceLifespan: time.Minute * 5, ProofLifespan: time.Second * 10, ClockSkew: time.Second * 30},
			[]string{"identity_providers: oidc: dpop: option 'nonce_enforced' can't be enabled when option 'enabled' is disabled"},
		},
		{
			"ShouldRaiseErrorOnKeyBindingWithoutEnabled",
			schema.IdentityProvidersOpenIDConnectDPoP{KeyBinding: true},
			schema.IdentityProvidersOpenIDConnectDPoP{KeyBinding: true, NonceLifespan: time.Minute * 5, ProofLifespan: time.Second * 10, ClockSkew: time.Second * 30},
			[]string{"identity_providers: oidc: dpop: option 'key_binding' can't be enabled when option 'enabled' is disabled"},
		},
		{
			"ShouldNotRaiseErrorOnKeyBindingWithEnabled",
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, KeyBinding: true},
			schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, KeyBinding: true, NonceLifespan: time.Minute * 5, ProofLifespan: time.Second * 10, ClockSkew: time.Second * 30},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret:  "abcdefghijklmnopqrstuvwxyz123456",
						JSONWebKeys: []schema.JWK{{Key: keyRSA2048, CertificateChain: certRSA2048}},
						DPoP:        tc.Have,
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{ID: "test", Secret: tOpenIDConnectPlainTextClientSecret, RedirectURIs: []string{"https://app.example.com"}},
						},
					},
				},
			}

			ValidateIdentityProviders(NewValidateCtx(), config, validator)

			expected := tc.Expected

			if expected.AllowedAlgorithms == nil {
				expected.AllowedAlgorithms = oidc.DPoPSigningAlgValuesSupported()
			}

			assert.Equal(t, expected, config.IdentityProviders.OIDC.DPoP)

			errs := validator.Errors()

			require.Len(t, errs, len(tc.Errors))

			for i, expected := range tc.Errors {
				assert.EqualError(t, errs[i], expected)
			}
		})
	}
}

func TestValidateIdentityProvidersShouldValidateDPoPPolicyOptions(t *testing.T) {
	testCases := []struct {
		name     string
		have     schema.IdentityProvidersOpenIDConnectDPoP
		clients  []schema.IdentityProvidersOpenIDConnectClient
		expected []string
		errs     []string
		warns    []string
	}{
		{
			name:     "ShouldSetDefaultAllowedAlgorithms",
			have:     schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true},
			expected: []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512", "Ed25519", "EdDSA"},
		},
		{
			name:     "ShouldAllowConfiguredAllowedAlgorithms",
			have:     schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, AllowedAlgorithms: []string{"PS256", "ES256", "Ed25519", "ML-DSA-65"}},
			expected: []string{"PS256", "ES256", "Ed25519", "ML-DSA-65"},
		},
		{
			name:     "ShouldRaiseErrorOnSymmetricAllowedAlgorithm",
			have:     schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, AllowedAlgorithms: []string{"ES256", "HS256", "none"}},
			expected: []string{"ES256", "HS256", "none"},
			errs: []string{
				"identity_providers: oidc: dpop: option 'allowed_algorithms' must only contain values from 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', 'ES512', 'Ed25519', 'EdDSA', 'ML-DSA-44', 'ML-DSA-65', or 'ML-DSA-87' but it contains 'HS256'",
				"identity_providers: oidc: dpop: option 'allowed_algorithms' must only contain values from 'RS256', 'PS256', 'ES256', 'RS384', 'PS384', 'ES384', 'RS512', 'PS512', 'ES512', 'Ed25519', 'EdDSA', 'ML-DSA-44', 'ML-DSA-65', or 'ML-DSA-87' but it contains 'none'",
			},
		},
		{
			name:     "ShouldAllowStrictRefreshTokenBindingWhenEnabled",
			have:     schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, StrictRefreshTokenBinding: true},
			clients:  []schema.IdentityProvidersOpenIDConnectClient{{DPoPStrictRefreshTokenBinding: true}},
			expected: oidc.DPoPSigningAlgValuesSupported(),
		},
		{
			name:     "ShouldRaiseErrorOnStrictRefreshTokenBindingWithoutEnabled",
			have:     schema.IdentityProvidersOpenIDConnectDPoP{StrictRefreshTokenBinding: true},
			expected: oidc.DPoPSigningAlgValuesSupported(),
			errs:     []string{"identity_providers: oidc: dpop: option 'strict_refresh_token_binding' can't be enabled when option 'enabled' is disabled"},
		},
		{
			name:     "ShouldRaiseErrorOnClientStrictRefreshTokenBindingWithoutEnabled",
			clients:  []schema.IdentityProvidersOpenIDConnectClient{{DPoPStrictRefreshTokenBinding: true}},
			expected: oidc.DPoPSigningAlgValuesSupported(),
			errs:     []string{"identity_providers: oidc: clients: client 'test': option 'dpop_strict_refresh_token_binding' can't be enabled when the 'identity_providers: oidc: dpop: enabled' option is disabled"},
		},
		{
			name:     "ShouldAllowNonceStatelessWhenEnabled",
			have:     schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, NonceEnforced: true, NonceStateless: true},
			expected: oidc.DPoPSigningAlgValuesSupported(),
		},
		{
			name:     "ShouldRaiseErrorOnNonceStatelessWithoutEnabled",
			have:     schema.IdentityProvidersOpenIDConnectDPoP{NonceStateless: true},
			expected: oidc.DPoPSigningAlgValuesSupported(),
			errs:     []string{"identity_providers: oidc: dpop: option 'nonce_stateless' can't be enabled when option 'enabled' is disabled"},
		},
		{
			name:     "ShouldNotWarnAtRecommendedMaximums",
			have:     schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, ProofLifespan: time.Minute, ClockSkew: time.Minute, NonceLifespan: time.Hour},
			expected: oidc.DPoPSigningAlgValuesSupported(),
		},
		{
			name:     "ShouldWarnAboveRecommendedMaximums",
			have:     schema.IdentityProvidersOpenIDConnectDPoP{Enabled: true, ProofLifespan: time.Minute * 2, ClockSkew: time.Minute * 5, NonceLifespan: time.Hour * 24},
			expected: oidc.DPoPSigningAlgValuesSupported(),
			warns: []string{
				"identity_providers: oidc: dpop: option 'proof_lifespan' is configured as '2m0s' which is greater than the recommended maximum of '1m0s'",
				"identity_providers: oidc: dpop: option 'clock_skew' is configured as '5m0s' which is greater than the recommended maximum of '1m0s'",
				"identity_providers: oidc: dpop: option 'nonce_lifespan' is configured as '24h0m0s' which is greater than the recommended maximum of '1h0m0s'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			clients := []schema.IdentityProvidersOpenIDConnectClient{{}}

			if tc.clients != nil {
				clients = tc.clients
			}

			for i := range clients {
				clients[i].ID = "test"
				clients[i].Secret = tOpenIDConnectPBKDF2ClientSecret
				clients[i].RedirectURIs = []string{"https://app.example.com"}
			}

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret:  "abcdefghijklmnopqrstuvwxyz123456",
						JSONWebKeys: []schema.JWK{{Key: keyRSA2048, CertificateChain: certRSA2048}},
						DPoP:        tc.have,
						Clients:     clients,
					},
				},
			}

			ValidateIdentityProviders(NewValidateCtx(), config, validator)

			assert.Equal(t, tc.expected, config.IdentityProviders.OIDC.DPoP.AllowedAlgorithms)

			errs := validator.Errors()

			require.Len(t, errs, len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, errs[i], expected)
			}

			warns := validator.Warnings()

			require.Len(t, warns, len(tc.warns))

			for i, expected := range tc.warns {
				assert.EqualError(t, warns[i], expected)
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

func MustLoadEd25519PrivateKey(mod string, extra ...string) ed25519.PrivateKey {
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

	// Ed25519 key / certificate pair.
	keyEd25519  ed25519.PrivateKey
	certEd25519 schema.X509CertificateChain
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

	certEd25519, keyEd25519 = MustLoadCertificateChain("Ed25519", ""), MustLoadEd25519PrivateKey("")

	keyRSA2048Legacy = MustLoadRSAPrivateKey("2048", "legacy")
}

func TestValidateOIDCClientAssertionJWTValidationHeader(t *testing.T) {
	testCases := []struct {
		name      string
		allowNone bool
		types     []string
		warnings  []string
		errors    []string
	}{
		{
			name: "ShouldNotWarnWhenUnconfigured",
		},
		{
			name:      "ShouldWarnWhenEmptyTypeAllowed",
			allowNone: true,
			warnings: []string{
				"identity_providers: oidc: clients: client 'test': option 'client_assertion_jwt_validation_header_allow_empty_type' is enabled which permits client assertions that omit the JWT 'typ' header, this is insecure as explicit typing guards an assertion against being confused with another JWT and should only be used when the client cannot be configured to send the 'typ' header",
			},
		},
		{
			name:  "ShouldNotWarnWhenOnlyExplicitTypeConfigured",
			types: []string{"client-authentication+jwt"},
		},
		{
			name:  "ShouldWarnWhenGenericTypeConfigured",
			types: []string{"client-authentication+jwt", "JWT"},
			warnings: []string{
				"identity_providers: oidc: clients: client 'test': option 'client_assertion_jwt_validation_header_allow_types' includes the type 'JWT' which is insecure as it does not distinguish a client assertion from another JWT, the explicit type 'client-authentication+jwt' should be preferred and another type should only be permitted when the client cannot be configured to send it",
			},
		},
		{
			name:  "ShouldWarnOnDuplicateTypes",
			types: []string{"client-authentication+jwt", "client-authentication+jwt"},
			warnings: []string{
				"identity_providers: oidc: clients: client 'test': option 'client_assertion_jwt_validation_header_allow_types' must have unique values but the values 'client-authentication+jwt' are duplicated",
			},
		},
		{
			name:  "ShouldErrorOnEmptyTypeEntry",
			types: []string{"client-authentication+jwt", ""},
			errors: []string{
				"identity_providers: oidc: clients: client 'test': option 'client_assertion_jwt_validation_header_allow_types' must not have empty values but an empty value was present at index 1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID: "test",
						ClientAssertionJWTValidationHeaderAllowEmptyType: tc.allowNone,
						ClientAssertionJWTValidationHeaderAllowTypes:     tc.types,
					},
				},
			}

			validateOIDCClientAssertionJWTValidationHeader(0, config, validator)

			require.Len(t, validator.Warnings(), len(tc.warnings))

			for i, warning := range tc.warnings {
				assert.EqualError(t, validator.Warnings()[i], warning)
			}

			require.Len(t, validator.Errors(), len(tc.errors))

			for i, err := range tc.errors {
				assert.EqualError(t, validator.Errors()[i], err)
			}
		})
	}
}

func TestValidateOIDCClientIntrospectionAudience(t *testing.T) {
	testCases := []struct {
		name     string
		audience []string
		errors   []string
	}{
		{
			name: "ShouldNotErrorWhenUnconfigured",
		},
		{
			name:     "ShouldNotErrorWhenConfigured",
			audience: []string{"https://api.example.com", "api"},
		},
		{
			name:     "ShouldErrorOnEmptyEntry",
			audience: []string{"api", ""},
			errors: []string{
				"identity_providers: oidc: clients: client 'test': option 'introspection_audience' must not have empty values but an empty value was present at index 1",
			},
		},
		{
			name:     "ShouldErrorOnDuplicateEntries",
			audience: []string{"api", "api"},
			errors: []string{
				"identity_providers: oidc: clients: client 'test': option 'introspection_audience' must have unique values but the values 'api' are duplicated",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                    "test",
						IntrospectionAudience: tc.audience,
					},
				},
			}

			validateOIDCClientIntrospectionAudience(0, config, validator)

			assert.Len(t, validator.Warnings(), 0)
			require.Len(t, validator.Errors(), len(tc.errors))

			for i, err := range tc.errors {
				assert.EqualError(t, validator.Errors()[i], err)
			}
		})
	}
}

func TestValidateOIDCClientTokenExchange(t *testing.T) {
	testCases := []struct {
		name     string
		clients  []schema.IdentityProvidersOpenIDConnectClient
		errs     []string
		warns    []string
		expected func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnect)
	}{
		{
			"ShouldRejectUnsupportedSubjectTokenType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                         "test",
					GrantTypes:                 []string{oidc.GrantTypeTokenExchange},
					SubjectTokenTypesSupported: []string{oidc.TokenTypeJWT},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'subject_token_types_supported' must only contain token types enabled by 'token_exchange: token_types' but it has 'urn:ietf:params:oauth:token-type:jwt'"},
			nil,
			nil,
		},
		{
			"ShouldRejectTokenExchangeOptionsWithoutGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                         "test",
					GrantTypes:                 []string{oidc.GrantTypeAuthorizationCode},
					SubjectTokenTypesSupported: []string{oidc.TokenTypeAccessToken},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'subject_token_types_supported' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'"},
			nil,
			nil,
		},
		{
			"ShouldRejectUnknownReferencedClient",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:         "test",
					Audience:   []string{"nonexistent"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "nonexistent"},
					},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'subject_token_clients_supported' must only reference registered clients but the client 'nonexistent' does not exist"},
			nil,
			nil,
		},
		{
			"ShouldRejectDuplicateReferencedClient",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:         "test",
					Audience:   []string{"other"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
						{ClientID: "other"},
					},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'subject_token_clients_supported' must only reference each client once but the client 'other' is referenced multiple times"},
			nil,
			nil,
		},
		{
			"ShouldWarnSubjectTokenClientNotInAudience",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:         "test",
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
					},
				},
			},
			nil,
			[]string{"identity_providers: oidc: clients: client 'test': option 'subject_token_clients_supported' references client 'other' which is not in the 'audience' option, so tokens issued to this client can never be exchanged by it"},
			nil,
		},
		{
			"ShouldWarnActorTokenClientNotInAudience",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:         "test",
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					ActorTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
					},
				},
			},
			nil,
			[]string{"identity_providers: oidc: clients: client 'test': option 'actor_token_clients_supported' references client 'other' which is not in the 'audience' option, so tokens issued to this client can never be exchanged by it"},
			nil,
		},
		{
			"ShouldNotWarnTokenClientInAudience",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:         "test",
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					Audience:   []string{"other"},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
					},
					ActorTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
					},
				},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldNotWarnTokenClientReferencingItself",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:         "test",
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "test"},
					},
					ActorTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "test"},
					},
				},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldRejectMayActWithAllActorTokenTypes",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                             "test",
					GrantTypes:                     []string{oidc.GrantTypeTokenExchange},
					ActorTokenWithoutMayActAllowed: true,
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'actor_token_without_may_act_allowed' can only be enabled when the 'actor_token_types_supported' option is configured, as an empty value permits every supported token type to be presented as an 'actor_token' with the RFC8693 Section 4.4 'may_act' check disabled"},
			nil,
			nil,
		},
		{
			"ShouldNotWarnMayActWithRestrictedActorTokenTypes",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                             "test",
					GrantTypes:                     []string{oidc.GrantTypeTokenExchange},
					ActorTokenTypesSupported:       []string{oidc.TokenTypeAccessToken},
					ActorTokenWithoutMayActAllowed: true,
				},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldRejectMayActWithoutGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                             "test",
					GrantTypes:                     []string{oidc.GrantTypeAuthorizationCode},
					ActorTokenWithoutMayActAllowed: true,
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'actor_token_without_may_act_allowed' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'"},
			nil,
			nil,
		},
		{
			"ShouldRejectSubjectTokenClientsWithoutGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:         "test",
					Audience:   []string{"other"},
					GrantTypes: []string{oidc.GrantTypeAuthorizationCode},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
					},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'subject_token_clients_supported' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'"},
			nil,
			nil,
		},
		{
			"ShouldRejectUnsupportedRequestedTokenTypeInSubjectTokenClients",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:         "test",
					Audience:   []string{"other"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other", RequestedTokenTypes: []string{oidc.TokenTypeJWT}},
					},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'subject_token_clients_supported' must only contain token types enabled by 'token_exchange: token_types' but it has 'urn:ietf:params:oauth:token-type:jwt'"},
			nil,
			nil,
		},
		{
			"ShouldRejectReferencedClientWithoutGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeAuthorizationCode}},
				{
					ID:         "test",
					Audience:   []string{"other"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
					},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'subject_token_clients_supported' must only reference clients which have the 'urn:ietf:params:oauth:grant-type:token-exchange' value in their 'grant_types' option but the client 'other' does not"},
			nil,
			nil,
		},
		{
			"ShouldAllowReferencedClientWithGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:         "test",
					Audience:   []string{"other"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
					},
				},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldAllowSelfReferenceInSubjectTokenClients",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:         "test",
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "test"},
					},
				},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldRejectTokenIssuersWithoutGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                           "test",
					GrantTypes:                   []string{oidc.GrantTypeAuthorizationCode},
					SubjectTokenIssuersSupported: []string{"https://issuer.example.com"},
					ActorTokenIssuersSupported:   []string{"https://actor-issuer.example.com"},
				},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'subject_token_issuers_supported' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'",
				"identity_providers: oidc: clients: client 'test': option 'actor_token_issuers_supported' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'",
			},
			nil,
			nil,
		},
		{
			"ShouldWarnOnDuplicateSubjectTokenTypes",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                         "test",
					GrantTypes:                 []string{oidc.GrantTypeTokenExchange},
					SubjectTokenTypesSupported: []string{oidc.TokenTypeAccessToken, oidc.TokenTypeAccessToken},
				},
			},
			nil,
			[]string{"identity_providers: oidc: clients: client 'test': option 'subject_token_types_supported' must have unique values but the values 'urn:ietf:params:oauth:token-type:access_token' are duplicated"},
			nil,
		},
		{
			"ShouldWarnOnDuplicateRequestedTokenTypesInSubjectTokenClients",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:         "test",
					Audience:   []string{"other"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other", RequestedTokenTypes: []string{oidc.TokenTypeAccessToken, oidc.TokenTypeAccessToken}},
					},
				},
			},
			nil,
			[]string{"identity_providers: oidc: clients: client 'test': option 'subject_token_clients_supported' must have unique values but the values 'urn:ietf:params:oauth:token-type:access_token' are duplicated"},
			nil,
		},
		{
			"ShouldDefaultTokenTypesWhenGrantTypePresent",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
			},
			nil,
			nil,
			func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnect) {
				assert.Equal(t, []string{oidc.TokenTypeAccessToken, oidc.TokenTypeRefreshToken}, actual.Clients[0].SubjectTokenTypesSupported)
				assert.Equal(t, []string{oidc.TokenTypeAccessToken}, actual.Clients[0].RequestTokenTypesSupported)
				require.NotNil(t, actual.Clients[0].OwnAccessTokenExchangeTargetsRestricted)
				assert.True(t, *actual.Clients[0].OwnAccessTokenExchangeTargetsRestricted)
				assert.Empty(t, actual.Clients[0].ActorTokenTypesSupported)
			},
		},
		{
			"ShouldNotDefaultTokenTypesWithoutGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", GrantTypes: []string{oidc.GrantTypeAuthorizationCode}},
			},
			nil,
			nil,
			func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnect) {
				assert.Empty(t, actual.Clients[0].SubjectTokenTypesSupported)
				assert.Empty(t, actual.Clients[0].RequestTokenTypesSupported)
				assert.Nil(t, actual.Clients[0].OwnAccessTokenExchangeTargetsRestricted)
			},
		},
		{
			"ShouldNotOverrideExplicitOwnAccessTokenExchangeTargetsRestricted",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", GrantTypes: []string{oidc.GrantTypeTokenExchange}, OwnAccessTokenExchangeAllowed: true, OwnAccessTokenExchangeTargetsRestricted: new(false)},
			},
			nil,
			nil,
			func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnect) {
				require.NotNil(t, actual.Clients[0].OwnAccessTokenExchangeTargetsRestricted)
				assert.False(t, *actual.Clients[0].OwnAccessTokenExchangeTargetsRestricted)
			},
		},
		{
			"ShouldWarnOwnAccessTokenExchangeTargetsRestrictedWithoutOwnAccessTokenExchange",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", GrantTypes: []string{oidc.GrantTypeTokenExchange}, OwnAccessTokenExchangeTargetsRestricted: new(true)},
			},
			nil,
			[]string{"identity_providers: oidc: clients: client 'test': option 'own_access_token_exchange_targets_restricted' has no effect when option 'own_access_token_exchange_allowed' is not enabled"},
			nil,
		},
		{
			"ShouldRejectRefreshTokenRequestTokenTypeWithoutAllowed",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", GrantTypes: []string{oidc.GrantTypeTokenExchange, oidc.GrantTypeRefreshToken}, RequestTokenTypesSupported: []string{oidc.TokenTypeAccessToken, oidc.TokenTypeRefreshToken}},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'request_token_types_supported' can only contain the 'urn:ietf:params:oauth:token-type:refresh_token' value when the 'token_exchange_refresh_token_allowed' option is enabled"},
			nil,
			nil,
		},
		{
			"ShouldRejectRefreshTokenRequestTokenTypeWithoutRefreshTokenGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", GrantTypes: []string{oidc.GrantTypeTokenExchange}, TokenExchangeRefreshTokenAllowed: true, RequestTokenTypesSupported: []string{oidc.TokenTypeRefreshToken}},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'request_token_types_supported' can only contain the 'urn:ietf:params:oauth:token-type:refresh_token' value when the 'grant_types' option contains 'refresh_token'"},
			nil,
			nil,
		},
		{
			"ShouldAllowRefreshTokenRequestTokenType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", GrantTypes: []string{oidc.GrantTypeTokenExchange, oidc.GrantTypeRefreshToken}, TokenExchangeRefreshTokenAllowed: true, RequestTokenTypesSupported: []string{oidc.TokenTypeRefreshToken}},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldRejectNewOptionsWithoutGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:                                      "test",
					Audience:                                []string{"other"},
					GrantTypes:                              []string{oidc.GrantTypeAuthorizationCode},
					TokenExchangeRefreshTokenAllowed:        true,
					OwnAccessTokenExchangeAllowed:           true,
					OwnAccessTokenExchangeTargetsRestricted: new(true),
					ImpersonationAllowed:                    true,
					ActorTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
					},
				},
			},
			[]string{
				"identity_providers: oidc: clients: client 'test': option 'token_exchange_refresh_token_allowed' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'",
				"identity_providers: oidc: clients: client 'test': option 'own_access_token_exchange_allowed' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'",
				"identity_providers: oidc: clients: client 'test': option 'own_access_token_exchange_targets_restricted' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'",
				"identity_providers: oidc: clients: client 'test': option 'impersonation_allowed' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'",
				"identity_providers: oidc: clients: client 'test': option 'actor_token_clients_supported' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'",
			},
			nil,
			nil,
		},
		{
			"ShouldNotWarnImpersonationDisabledWithoutActorTokenTypes",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "test", GrantTypes: []string{oidc.GrantTypeTokenExchange}, ImpersonationAllowed: false},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldRejectUnknownActorTokenClient",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:         "test",
					Audience:   []string{"nonexistent"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					ActorTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "nonexistent"},
					},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'actor_token_clients_supported' must only reference registered clients but the client 'nonexistent' does not exist"},
			nil,
			nil,
		},
		{
			"ShouldRejectActorTokenClientWithoutGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeAuthorizationCode}},
				{
					ID:         "test",
					Audience:   []string{"other"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					ActorTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
					},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'actor_token_clients_supported' must only reference clients which have the 'urn:ietf:params:oauth:grant-type:token-exchange' value in their 'grant_types' option but the client 'other' does not"},
			nil,
			nil,
		},
		{
			"ShouldRejectDuplicateActorTokenClient",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:         "test",
					Audience:   []string{"other"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					ActorTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other"},
						{ClientID: "other"},
					},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'actor_token_clients_supported' must only reference each client once but the client 'other' is referenced multiple times"},
			nil,
			nil,
		},
		{
			"ShouldRejectUnsupportedRequestedTokenTypeInActorTokenClients",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:         "test",
					Audience:   []string{"other"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					ActorTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other", RequestedTokenTypes: []string{oidc.TokenTypeJWT}},
					},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'actor_token_clients_supported' must only contain token types enabled by 'token_exchange: token_types' but it has 'urn:ietf:params:oauth:token-type:jwt'"},
			nil,
			nil,
		},
		{
			"ShouldAllowActorTokenClients",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "other", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:         "test",
					Audience:   []string{"other"},
					GrantTypes: []string{oidc.GrantTypeTokenExchange},
					ActorTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{
						{ClientID: "other", RequestedTokenTypes: []string{oidc.TokenTypeAccessToken, oidc.TokenTypeRefreshToken}},
					},
				},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldAllowMayActWithoutGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "actor", GrantTypes: []string{oidc.GrantTypeTokenExchange}},
				{
					ID:                  "test",
					Audience:            []string{"actor"},
					GrantTypes:          []string{oidc.GrantTypeAuthorizationCode},
					TokenExchangeMayAct: &schema.IdentityProvidersOpenIDConnectClientMayAct{ClientID: "actor", Subject: "abc"},
				},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldAllowMayActSubjectOnly",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "test",
					GrantTypes:          []string{oidc.GrantTypeAuthorizationCode},
					TokenExchangeMayAct: &schema.IdentityProvidersOpenIDConnectClientMayAct{Subject: "abc"},
				},
			},
			nil,
			nil,
			nil,
		},
		{
			"ShouldRejectMayActEmpty",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "test",
					GrantTypes:          []string{oidc.GrantTypeAuthorizationCode},
					TokenExchangeMayAct: &schema.IdentityProvidersOpenIDConnectClientMayAct{},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'token_exchange_may_act' must have at least one of the 'client_id' or 'subject' options configured"},
			nil,
			nil,
		},
		{
			"ShouldRejectMayActUnknownClient",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "test",
					Audience:            []string{"nonexistent"},
					GrantTypes:          []string{oidc.GrantTypeAuthorizationCode},
					TokenExchangeMayAct: &schema.IdentityProvidersOpenIDConnectClientMayAct{ClientID: "nonexistent"},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'token_exchange_may_act' must only reference registered clients but the client 'nonexistent' does not exist"},
			nil,
			nil,
		},
		{
			"ShouldRejectMayActClientWithoutGrantType",
			[]schema.IdentityProvidersOpenIDConnectClient{
				{ID: "actor", GrantTypes: []string{oidc.GrantTypeAuthorizationCode}},
				{
					ID:                  "test",
					Audience:            []string{"actor"},
					GrantTypes:          []string{oidc.GrantTypeAuthorizationCode},
					TokenExchangeMayAct: &schema.IdentityProvidersOpenIDConnectClientMayAct{ClientID: "actor"},
				},
			},
			[]string{"identity_providers: oidc: clients: client 'test': option 'token_exchange_may_act' must only reference clients which have the 'urn:ietf:params:oauth:grant-type:token-exchange' value in their 'grant_types' option but the client 'actor' does not"},
			nil,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{Clients: tc.clients, TokenExchange: tokenExchangeTestDefaults()}
			validator := schema.NewStructValidator()

			for c := range config.Clients {
				validateOIDCClientTokenExchange(c, config, validator)
			}

			assert.Len(t, validator.Errors(), len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], expected)
			}

			assert.Len(t, validator.Warnings(), len(tc.warns))

			for i, expected := range tc.warns {
				assert.EqualError(t, validator.Warnings()[i], expected)
			}

			if tc.expected != nil {
				tc.expected(t, config)
			}
		})
	}
}

func tokenExchangeTestDefaults() schema.IdentityProvidersOpenIDConnectTokenExchange {
	te := schema.IdentityProvidersOpenIDConnectTokenExchange{
		TokenTypes:                slices.Clone(schema.DefaultOpenIDConnectConfiguration.TokenExchange.TokenTypes),
		DefaultRequestedTokenType: schema.DefaultOpenIDConnectConfiguration.TokenExchange.DefaultRequestedTokenType,
	}

	return te
}

func TestValidateOIDCClientTokenExchangeTokenTypesAgainstGlobal(t *testing.T) {
	const (
		access  = oidc.TokenTypeAccessToken
		refresh = oidc.TokenTypeRefreshToken
		idToken = oidc.TokenTypeIDToken
		txn     = oidc.TokenTypeTxnToken
		prefix  = "identity_providers: oidc: clients: client 'test': option "
		notEn   = " must only contain token types enabled by 'token_exchange: token_types' but it has "
	)

	grants := []string{oidc.GrantTypeTokenExchange}

	withTxn := func(replacement bool) schema.IdentityProvidersOpenIDConnectTokenExchange {
		te := tokenExchangeTestDefaults()
		te.TokenTypes = []string{access, refresh, idToken, txn}
		te.TransactionTokens.TrustDomain = "example.com"
		te.TransactionTokens.Replacement.Enabled = replacement

		return te
	}

	accessOnly := tokenExchangeTestDefaults()
	accessOnly.TokenTypes = []string{access}

	testCases := []struct {
		name     string
		te       schema.IdentityProvidersOpenIDConnectTokenExchange
		client   schema.IdentityProvidersOpenIDConnectClient
		errs     []string
		expected func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnectClient)
	}{
		{"ShouldRejectSubjectTypeNotEnabled", accessOnly, schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: grants, SubjectTokenTypesSupported: []string{refresh}}, []string{prefix + "'subject_token_types_supported'" + notEn + "'" + refresh + "'"}, nil},
		{"ShouldRejectActorTypeNotEnabled", accessOnly, schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: grants, ActorTokenTypesSupported: []string{idToken}}, []string{prefix + "'actor_token_types_supported'" + notEn + "'" + idToken + "'"}, nil},
		{"ShouldRejectRequestTypeNotEnabled", accessOnly, schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: grants, RequestTokenTypesSupported: []string{idToken}}, []string{prefix + "'request_token_types_supported'" + notEn + "'" + idToken + "'"}, nil},
		{
			"ShouldRejectPolicyRequestedTypeNotEnabled", accessOnly,
			schema.IdentityProvidersOpenIDConnectClient{ID: "test", Audience: []string{"test"}, GrantTypes: grants, SubjectTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{{ClientID: "test", RequestedTokenTypes: []string{refresh}}}},
			[]string{prefix + "'subject_token_clients_supported'" + notEn + "'" + refresh + "'"}, nil,
		},
		{
			"ShouldRejectActorPolicyRequestedTypeNotEnabled", accessOnly,
			schema.IdentityProvidersOpenIDConnectClient{ID: "test", Audience: []string{"test"}, GrantTypes: grants, ActorTokenClientsSupported: []schema.IdentityProvidersOpenIDConnectClientTokenExchangePolicy{{ClientID: "test", RequestedTokenTypes: []string{refresh}}}},
			[]string{prefix + "'actor_token_clients_supported'" + notEn + "'" + refresh + "'"}, nil,
		},
		{"ShouldRejectTxnTokenAsActorType", withTxn(true), schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: grants, ActorTokenTypesSupported: []string{txn}}, []string{prefix + "'actor_token_types_supported' must not contain the '" + txn + "' value as Transaction Tokens are not accepted as an 'actor_token'"}, nil},
		{"ShouldRejectTxnTokenSubjectWithoutReplacement", withTxn(true), schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: grants, SubjectTokenTypesSupported: []string{txn}}, []string{prefix + "'subject_token_types_supported' can only contain the '" + txn + "' value when the 'token_exchange_txn_token_replacement' option is enabled"}, nil},
		{"ShouldAllowTxnTokenSubjectWithReplacement", withTxn(true), schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: grants, SubjectTokenTypesSupported: []string{txn}, TokenExchangeTxnTokenReplacement: true}, nil, nil},
		{"ShouldRejectReplacementWithoutGrantType", withTxn(true), schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: []string{oidc.GrantTypeAuthorizationCode}, TokenExchangeTxnTokenReplacement: true}, []string{prefix + "'token_exchange_txn_token_replacement' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'"}, nil},
		{"ShouldRejectReplacementWithoutGlobalEnabled", withTxn(false), schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: grants, TokenExchangeTxnTokenReplacement: true}, []string{prefix + "'token_exchange_txn_token_replacement' can only be enabled when the 'token_exchange: transaction_tokens: replacement: enabled' option is enabled"}, nil},
		{"ShouldRejectTxnRequestWithoutTrustDomainAudience", withTxn(false), schema.IdentityProvidersOpenIDConnectClient{ID: "test", Audience: []string{"other"}, GrantTypes: grants, RequestTokenTypesSupported: []string{txn}}, []string{prefix + "'audience' must include the 'transaction_tokens: trust_domain' value 'example.com' when the 'request_token_types_supported' option contains '" + txn + "'"}, nil},
		{"ShouldAllowTxnRequestWithTrustDomainAudience", withTxn(false), schema.IdentityProvidersOpenIDConnectClient{ID: "test", Audience: []string{"example.com"}, GrantTypes: grants, RequestTokenTypesSupported: []string{txn}}, nil, nil},
		{
			"ShouldDefaultRequestTypeToConfiguredDefault", func() schema.IdentityProvidersOpenIDConnectTokenExchange {
				te := tokenExchangeTestDefaults()
				te.DefaultRequestedTokenType = idToken

				return te
			}(),
			schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: grants}, nil,
			func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnectClient) {
				assert.Equal(t, []string{idToken}, actual.RequestTokenTypesSupported)
				assert.Equal(t, []string{access, refresh}, actual.SubjectTokenTypesSupported)
			},
		},
		{
			"ShouldRejectDefaultSubjectTypesWhenNoneEnabled", func() schema.IdentityProvidersOpenIDConnectTokenExchange {
				te := tokenExchangeTestDefaults()
				te.TokenTypes = []string{idToken}
				te.DefaultRequestedTokenType = idToken

				return te
			}(),
			schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: grants},
			[]string{prefix + "'subject_token_types_supported' must be configured explicitly when 'token_exchange: token_types' does not include any of the default subject token types '" + access + "' and '" + refresh + "'"},
			func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnectClient) {
				assert.Empty(t, actual.SubjectTokenTypesSupported)
			},
		},
		{
			"ShouldIntersectDefaultSubjectTypesWithGlobal", accessOnly, schema.IdentityProvidersOpenIDConnectClient{ID: "test", GrantTypes: grants}, nil,
			func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnectClient) {
				assert.Equal(t, []string{access}, actual.SubjectTokenTypesSupported)
				assert.Equal(t, []string{access}, actual.RequestTokenTypesSupported)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{Clients: []schema.IdentityProvidersOpenIDConnectClient{tc.client}, TokenExchange: tc.te}
			validator := schema.NewStructValidator()

			validateOIDCClientTokenExchange(0, config, validator)

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], expected)
			}

			if tc.expected != nil {
				tc.expected(t, &config.Clients[0])
			}
		})
	}
}

func TestValidateOIDCClientTokenExchangeTxnTokenReplacementPolicy(t *testing.T) {
	const warn = "identity_providers: oidc: clients: client 'test': option 'token_exchange_txn_token_replacement' is enabled but the 'authorization_policy' '%s' requires 'two_factor' by default and replaced Transaction Tokens carry no authentication methods references, so replacement requests will be rejected wherever the policy requires 'two_factor'"

	policies := map[string]schema.IdentityProvidersOpenIDConnectPolicy{
		"named-two": {DefaultPolicy: policyTwoFactor},
		"named-one": {DefaultPolicy: policyOneFactor, Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{{Policy: policyTwoFactor}}},
	}

	testCases := []struct {
		name        string
		policy      string
		replacement bool
		warns       []string
	}{
		{"ShouldWarnTwoFactor", policyTwoFactor, true, []string{fmt.Sprintf(warn, policyTwoFactor)}},
		{"ShouldWarnNamedPolicyDefaultTwoFactor", "named-two", true, []string{fmt.Sprintf(warn, "named-two")}},
		{"ShouldNotWarnOneFactor", policyOneFactor, true, nil},
		{"ShouldNotWarnNamedPolicyDefaultOneFactor", "named-one", true, nil},
		{"ShouldNotWarnWithoutReplacement", policyTwoFactor, false, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			te := tokenExchangeTestDefaults()
			te.TokenTypes = []string{oidc.TokenTypeAccessToken, oidc.TokenTypeRefreshToken, oidc.TokenTypeTxnToken}
			te.TransactionTokens.TrustDomain = "example.com"
			te.TransactionTokens.Replacement.Enabled = true

			config := &schema.IdentityProvidersOpenIDConnect{
				AuthorizationPolicies: policies,
				TokenExchange:         te,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                               "test",
						AuthorizationPolicy:              tc.policy,
						GrantTypes:                       []string{oidc.GrantTypeTokenExchange},
						TokenExchangeTxnTokenReplacement: tc.replacement,
					},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCClientTokenExchange(0, config, validator)

			assert.Len(t, validator.Errors(), 0)
			require.Len(t, validator.Warnings(), len(tc.warns))

			for i, expected := range tc.warns {
				assert.EqualError(t, validator.Warnings()[i], expected)
			}
		})
	}
}

// TestValidateOIDCClientTokenExchangeWiring proves validateOIDCClientTokenExchange is actually reached from the
// exported ValidateIdentityProviders entry point (i.e. wired into validateOIDCClient), rather than only being
// exercised by tests that call it directly.
func TestValidateOIDCClientTokenExchangeWiring(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret:       "hmac1",
				IssuerPrivateKey: keyRSA2048,
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  "client-with-token-exchange-option-without-grant-type",
						Public:              true,
						AuthorizationPolicy: "two_factor",
						RedirectURIs: []string{
							"https://localhost",
						},
						GrantTypes:                 []string{oidc.GrantTypeAuthorizationCode},
						SubjectTokenTypesSupported: []string{oidc.TokenTypeAccessToken},
					},
				},
			},
		},
	}

	ValidateIdentityProviders(NewValidateCtx(), config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: clients: client 'client-with-token-exchange-option-without-grant-type': option 'subject_token_types_supported' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'")
}

func TestValidateOIDCTokenExchange(t *testing.T) {
	const (
		prefix = "identity_providers: oidc: token_exchange: option "
		at     = "urn:ietf:params:oauth:token-type:access_token"
		rt     = "urn:ietf:params:oauth:token-type:refresh_token"
		id     = "urn:ietf:params:oauth:token-type:id_token"
		txn    = "urn:ietf:params:oauth:token-type:txn_token"
		idjag  = "urn:ietf:params:oauth:token-type:id-jag"
		valid  = "'" + at + "', '" + rt + "', '" + id + "', '" + txn + "', or '" + idjag + "'"
	)

	testCases := []struct {
		name     string
		have     schema.IdentityProvidersOpenIDConnectTokenExchange
		expected func(t *testing.T, have schema.IdentityProvidersOpenIDConnectTokenExchange)
		warnings []string
		errors   []string
	}{
		{
			"ShouldApplyDefaults",
			schema.IdentityProvidersOpenIDConnectTokenExchange{},
			func(t *testing.T, have schema.IdentityProvidersOpenIDConnectTokenExchange) {
				assert.Equal(t, []string{at, rt, id}, have.TokenTypes)
				assert.Equal(t, at, have.DefaultRequestedTokenType)
				assert.Equal(t, time.Minute*5, have.TransactionTokens.Lifespan)
				assert.Equal(t, "RS256", have.TransactionTokens.SigningAlgorithm)
				assert.Equal(t, 8, have.TransactionTokens.Replacement.MaxChainLength)
			},
			nil,
			nil,
		},
		{
			"ShouldRejectShortName",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, "access_token"}},
			nil,
			nil,
			[]string{prefix + "'token_types' must only contain the values " + valid + " but it has the value 'access_token'"},
		},
		{
			"ShouldWarnOnDuplicate",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, at}},
			nil,
			[]string{prefix + "'token_types' must have unique values but the values '" + at + "' are duplicated"},
			nil,
		},
		{
			"ShouldRejectDefaultNotConfigured",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at}, DefaultRequestedTokenType: id},
			nil,
			nil,
			[]string{prefix + "'default_requested_token_type' must be one of the configured 'token_types' '" + at + "' but it's configured as '" + id + "'"},
		},
		{
			"ShouldAcceptDefaultIDToken",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, id}, DefaultRequestedTokenType: id},
			nil,
			nil,
			nil,
		},
		{
			"ShouldRejectDefaultRefreshToken",
			schema.IdentityProvidersOpenIDConnectTokenExchange{DefaultRequestedTokenType: rt},
			nil,
			nil,
			[]string{prefix + "'default_requested_token_type' must not be '" + rt + "'"},
		},
		{
			"ShouldRejectDefaultTxnToken",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, txn}, DefaultRequestedTokenType: txn, TransactionTokens: schema.IdentityProvidersOpenIDConnectTransactionTokens{TrustDomain: "example.com"}},
			nil,
			nil,
			[]string{prefix + "'default_requested_token_type' must not be '" + txn + "'"},
		},
		{
			"ShouldRejectDefaultIDJAG",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, idjag}, DefaultRequestedTokenType: idjag},
			nil,
			nil,
			[]string{
				prefix + "'default_requested_token_type' must not be '" + idjag + "'",
				prefix + "'token_types' includes '" + idjag + "' which requires the 'identity_assertion_authorization_grant' option 'resource_authorization_servers' to be configured",
			},
		},
		{
			"ShouldRejectIDJAGWithoutResourceAuthorizationServers",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, idjag}},
			nil,
			nil,
			[]string{prefix + "'token_types' includes '" + idjag + "' which requires the 'identity_assertion_authorization_grant' option 'resource_authorization_servers' to be configured"},
		},
		{
			"ShouldRejectTransactionTokensWithoutType",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TransactionTokens: schema.IdentityProvidersOpenIDConnectTransactionTokens{TrustDomain: "example.com"}},
			nil,
			nil,
			[]string{prefix + "'transaction_tokens' must not be configured when 'token_types' does not include '" + txn + "'"},
		},
		{
			"ShouldRejectMissingTrustDomain",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, txn}},
			nil,
			nil,
			[]string{prefix + "'transaction_tokens: trust_domain' is required when 'token_types' includes '" + txn + "'"},
		},
		{
			"ShouldAcceptValidTransactionTokens",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, txn}, TransactionTokens: schema.IdentityProvidersOpenIDConnectTransactionTokens{TrustDomain: "example.com", SigningAlgorithm: "RS256"}},
			func(t *testing.T, have schema.IdentityProvidersOpenIDConnectTokenExchange) {
				assert.Equal(t, time.Minute*5, have.TransactionTokens.Lifespan)
				assert.Equal(t, 8, have.TransactionTokens.Replacement.MaxChainLength)
			},
			nil,
			nil,
		},
		{
			"ShouldRejectNegativeLifespan",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, txn}, TransactionTokens: schema.IdentityProvidersOpenIDConnectTransactionTokens{TrustDomain: "example.com", Lifespan: -time.Minute}},
			nil,
			nil,
			[]string{prefix + "'transaction_tokens: lifespan' must be a positive duration but it's configured as '-1m0s'"},
		},
		{
			"ShouldRejectNegativeChainLength",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, txn}, TransactionTokens: schema.IdentityProvidersOpenIDConnectTransactionTokens{TrustDomain: "example.com", Replacement: schema.IdentityProvidersOpenIDConnectTransactionTokensReplacement{MaxChainLength: -1}}},
			nil,
			nil,
			[]string{prefix + "'transaction_tokens: replacement: max_chain_length' must be greater than or equal to 1 but it's configured as '-1'"},
		},
		{
			"ShouldRejectSigningAlgorithmWithoutMatchingKey",
			schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: []string{at, txn}, TransactionTokens: schema.IdentityProvidersOpenIDConnectTransactionTokens{TrustDomain: "example.com", SigningAlgorithm: "ES256"}},
			nil,
			nil,
			[]string{prefix + "'transaction_tokens: signing_algorithm' must be one of the algorithms of the configured 'jwks' 'RS256' but it's configured as 'ES256'"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret:       "abc",
						IssuerPrivateKey: keyRSA2048,
						TokenExchange:    tc.have,
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:                  "good_id",
								Secret:              tOpenIDConnectPBKDF2ClientSecret,
								AuthorizationPolicy: "two_factor",
								RedirectURIs:        []string{"https://google.com/callback"},
							},
						},
					},
				},
			}

			ValidateIdentityProviders(NewValidateCtx(), config, validator)

			if tc.expected != nil {
				tc.expected(t, config.IdentityProviders.OIDC.TokenExchange)
			}

			require.Len(t, validator.Warnings(), len(tc.warnings))

			for i, expected := range tc.warnings {
				assert.EqualError(t, validator.Warnings()[i], expected)
			}

			require.Len(t, validator.Errors(), len(tc.errors))

			for i, expected := range tc.errors {
				assert.EqualError(t, validator.Errors()[i], expected)
			}
		})
	}
}

func idjagTestURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	u, err := url.Parse(raw)
	require.NoError(t, err)

	return u
}

func TestValidateIDJAGGlobalDefaults(t *testing.T) {
	config := &schema.IdentityProvidersOpenIDConnect{}
	validator := schema.NewStructValidator()

	validateOIDCIDJAG(NewValidateCtx(), config, validator, nil)

	assert.Empty(t, validator.Errors())
	assert.Empty(t, validator.Warnings())
	require.NotNil(t, config.IdentityAssertionAuthorizationGrant.SingleUse)
	assert.True(t, *config.IdentityAssertionAuthorizationGrant.SingleUse)
	assert.Equal(t, 5*time.Minute, config.IdentityAssertionAuthorizationGrant.Lifespan)
	assert.Equal(t, time.Duration(0), config.IdentityAssertionAuthorizationGrant.AccessTokenAllowance)
}

func TestValidateIDJAGGlobal(t *testing.T) {
	testCases := []struct {
		name  string
		have  schema.IdentityProvidersOpenIDConnectIDJAG
		errs  []string
		warns []string
	}{
		{"ShouldRejectNegativeLifespan", schema.IdentityProvidersOpenIDConnectIDJAG{Lifespan: -time.Second}, []string{"identity_providers: oidc: identity_assertion_authorization_grant: option 'lifespan' must be greater than 0 but it's configured as '-1s'"}, nil},
		{"ShouldWarnLongLifespan", schema.IdentityProvidersOpenIDConnectIDJAG{Lifespan: 2 * time.Hour}, nil, []string{"identity_providers: oidc: identity_assertion_authorization_grant: option 'lifespan' is configured as '2h0m0s' which is longer than the recommended maximum of '1h' for a grant"}},
		{"ShouldAllowOneHourLifespan", schema.IdentityProvidersOpenIDConnectIDJAG{Lifespan: time.Hour}, nil, nil},
		{"ShouldRejectNegativeAccessTokenAllowance", schema.IdentityProvidersOpenIDConnectIDJAG{AccessTokenAllowance: -time.Second}, []string{"identity_providers: oidc: identity_assertion_authorization_grant: option 'access_token_allowance' must be greater than or equal to 0 but it's configured as '-1s'"}, nil},
		{"ShouldKeepExplicitSingleUseFalse", schema.IdentityProvidersOpenIDConnectIDJAG{SingleUse: new(false)}, nil, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{IdentityAssertionAuthorizationGrant: tc.have}
			validator := schema.NewStructValidator()

			validateOIDCIDJAG(NewValidateCtx(), config, validator, nil)

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], expected)
			}

			require.Len(t, validator.Warnings(), len(tc.warns))

			for i, expected := range tc.warns {
				assert.EqualError(t, validator.Warnings()[i], expected)
			}

			if tc.have.SingleUse != nil {
				assert.Equal(t, *tc.have.SingleUse, *config.IdentityAssertionAuthorizationGrant.SingleUse)
			} else {
				assert.True(t, *config.IdentityAssertionAuthorizationGrant.SingleUse)
			}
		})
	}
}

func TestValidateIDJAGResourceAuthorizationServer(t *testing.T) {
	const prefix = "identity_providers: oidc: identity_assertion_authorization_grant: resource_authorization_servers: server #1: option "

	valid := func(mutate func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer)) schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer {
		s := schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{
			Issuer:          idjagTestURL(t, "https://ras.example.com"),
			SubjectClientID: "subject",
		}

		if mutate != nil {
			mutate(&s)
		}

		return s
	}

	testCases := []struct {
		name     string
		have     schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer
		errs     []string
		expected string
	}{
		{"ShouldAcceptValidAndDefaultSigningAlgorithm", valid(nil), nil, "RS256"},
		{"ShouldAcceptConfiguredSigningAlgorithm", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.SigningAlgorithm = "ES256"
		}), nil, "ES256"},
		{"ShouldRejectMissingIssuer", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) { s.Issuer = nil }), []string{prefix + "'issuer' is required"}, "RS256"},
		{"ShouldRejectHTTPIssuer", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.Issuer = idjagTestURL(t, "http://ras.example.com")
		}), []string{prefix + "'issuer' must be an absolute URL with the 'https' scheme but it's configured as 'http://ras.example.com'"}, "RS256"},
		{"ShouldRejectRelativeIssuer", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.Issuer = idjagTestURL(t, "/ras")
		}), []string{prefix + "'issuer' must be an absolute URL with the 'https' scheme but it's configured as '/ras'"}, "RS256"},
		{"ShouldRejectAutheliaIssuer", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.Issuer = idjagTestURL(t, "https://auth.example.com")
		}), []string{prefix + "'issuer' must not be the same as the issuer of this provider but the value 'https://auth.example.com' is"}, "RS256"},
		{"ShouldRejectUnsupportedSigningAlgorithm", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.SigningAlgorithm = "ES512"
		}), []string{prefix + "'signing_algorithm' must be one of 'RS256' or 'ES256' but it's configured as 'ES512'"}, "ES512"},
		{"ShouldRejectNegativeMaxAge", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.MaxAge = -time.Second
		}), []string{prefix + "'max_age' must be greater than or equal to 0 but it's configured as '-1s'"}, "RS256"},
		{"ShouldRejectMissingSubjectClientID", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) { s.SubjectClientID = "" }), []string{prefix + "'subject_client_id' is required"}, "RS256"},
		{"ShouldRejectUnknownSubjectClientID", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.SubjectClientID = "nope"
		}), []string{prefix + "'subject_client_id' must reference a registered client but the client 'nope' does not exist"}, "RS256"},
		{"ShouldRejectUnknownClaimsPolicy", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.ClaimsPolicy = "nope"
		}), []string{prefix + "'claims_policy' must reference a configured claims policy but the claims policy 'nope' does not exist"}, "RS256"},
		{"ShouldAcceptKnownClaimsPolicy", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.ClaimsPolicy = "policy"
		}), nil, "RS256"},
		{"ShouldRejectEmptyAlias", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.Aliases = []string{""}
		}), []string{prefix + "'aliases' must not contain empty values"}, "RS256"},
		{"ShouldRejectAliasEqualToOwnIssuer", valid(func(s *schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer) {
			s.Aliases = []string{"https://ras.example.com"}
		}), []string{prefix + "'aliases' must be unique across the 'issuer' and 'aliases' options of all servers but the value 'https://ras.example.com' is configured multiple times"}, "RS256"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				Clients:        []schema.IdentityProvidersOpenIDConnectClient{{ID: "subject"}},
				ClaimsPolicies: map[string]schema.IdentityProvidersOpenIDConnectClaimsPolicy{"policy": {}},
				Discovery:      schema.IdentityProvidersOpenIDConnectDiscovery{ResponseObjectSigningAlgs: []string{"RS256", "ES256"}},
				IdentityAssertionAuthorizationGrant: schema.IdentityProvidersOpenIDConnectIDJAG{
					ResourceAuthorizationServers: []schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{tc.have},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCIDJAG(NewValidateCtx(), config, validator, []*url.URL{idjagTestURL(t, "https://auth.example.com")})

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], expected)
			}

			assert.Equal(t, tc.expected, config.IdentityAssertionAuthorizationGrant.ResourceAuthorizationServers[0].SigningAlgorithm)
		})
	}
}

func TestValidateIDJAGResourceAuthorizationServerDuplicateIssuer(t *testing.T) {
	config := &schema.IdentityProvidersOpenIDConnect{
		Clients: []schema.IdentityProvidersOpenIDConnectClient{{ID: "subject"}},
		IdentityAssertionAuthorizationGrant: schema.IdentityProvidersOpenIDConnectIDJAG{
			ResourceAuthorizationServers: []schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{
				{Issuer: idjagTestURL(t, "https://ras.example.com"), SubjectClientID: "subject"},
				{Issuer: idjagTestURL(t, "https://ras.example.com"), SubjectClientID: "subject"},
			},
		},
	}

	validator := schema.NewStructValidator()

	validateOIDCIDJAG(NewValidateCtx(), config, validator, nil)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: identity_assertion_authorization_grant: resource_authorization_servers: server #2: option 'issuer' must be unique but the value 'https://ras.example.com' is configured multiple times")
}

func TestValidateIDJAGResourceAuthorizationServerAliasCollision(t *testing.T) {
	config := &schema.IdentityProvidersOpenIDConnect{
		Clients: []schema.IdentityProvidersOpenIDConnectClient{{ID: "subject"}},
		IdentityAssertionAuthorizationGrant: schema.IdentityProvidersOpenIDConnectIDJAG{
			ResourceAuthorizationServers: []schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{
				{Issuer: idjagTestURL(t, "https://ras1.example.com"), SubjectClientID: "subject", Aliases: []string{"alias-one", "https://auth.example.com"}},
				{Issuer: idjagTestURL(t, "https://ras2.example.com"), SubjectClientID: "subject", Aliases: []string{"https://ras1.example.com", "alias-one"}},
			},
		},
	}

	validator := schema.NewStructValidator()

	validateOIDCIDJAG(NewValidateCtx(), config, validator, []*url.URL{idjagTestURL(t, "https://auth.example.com")})

	prefix := "identity_providers: oidc: identity_assertion_authorization_grant: resource_authorization_servers: server #%d: option "

	errs := validator.Errors()
	require.Len(t, errs, 3)
	assert.EqualError(t, errs[0], fmt.Sprintf(prefix, 1)+"'aliases' must not be the same as the issuer of this provider but the value 'https://auth.example.com' is")
	assert.EqualError(t, errs[1], fmt.Sprintf(prefix, 2)+"'aliases' must be unique across the 'issuer' and 'aliases' options of all servers but the value 'https://ras1.example.com' is configured multiple times")
	assert.EqualError(t, errs[2], fmt.Sprintf(prefix, 2)+"'aliases' must be unique across the 'issuer' and 'aliases' options of all servers but the value 'alias-one' is configured multiple times")
}

func TestValidateOIDCClientIDJAGRelationships(t *testing.T) {
	const prefix = "identity_providers: oidc: clients: client 'test': option "

	ras := schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{Issuer: idjagTestURL(t, "https://ras.example.com"), SubjectClientID: "test"}

	relationship := func(mutate func(r *schema.IdentityProvidersOpenIDConnectClientIDJAGRelationship)) schema.IdentityProvidersOpenIDConnectClientIDJAGRelationship {
		r := schema.IdentityProvidersOpenIDConnectClientIDJAGRelationship{Issuer: idjagTestURL(t, "https://ras.example.com"), ClientID: "remote"}

		if mutate != nil {
			mutate(&r)
		}

		return r
	}

	client := func(mutate func(c *schema.IdentityProvidersOpenIDConnectClient)) schema.IdentityProvidersOpenIDConnectClient {
		c := schema.IdentityProvidersOpenIDConnectClient{
			ID:                         "test",
			GrantTypes:                 []string{oidc.GrantTypeTokenExchange},
			SubjectTokenTypesSupported: []string{oidc.TokenTypeIDToken},
			RequestTokenTypesSupported: []string{oidc.TokenTypeIDJAG},
			IDJAGRelationships:         []schema.IdentityProvidersOpenIDConnectClientIDJAGRelationship{relationship(nil)},
		}

		if mutate != nil {
			mutate(&c)
		}

		return c
	}

	testCases := []struct {
		name     string
		have     schema.IdentityProvidersOpenIDConnectClient
		errs     []string
		expected func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnectClient)
	}{
		{"ShouldAcceptValid", client(nil), nil, nil},
		{"ShouldAcceptRefreshTokenSubjectType", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.SubjectTokenTypesSupported = []string{oidc.TokenTypeRefreshToken}
		}), nil, nil},
		{"ShouldRejectMissingGrantType", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.GrantTypes = []string{oidc.GrantTypeAuthorizationCode}
		}), []string{prefix + "'identity_assertion_authorization_grant_relationships' can only be configured when the 'grant_types' option contains 'urn:ietf:params:oauth:grant-type:token-exchange'"}, nil},
		{"ShouldRejectMissingTokenType", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.RequestTokenTypesSupported = []string{oidc.TokenTypeAccessToken}
		}), []string{prefix + "'identity_assertion_authorization_grant_relationships' can only be configured when the 'request_token_types_supported' option contains 'urn:ietf:params:oauth:token-type:id-jag'"}, nil},
		{"ShouldRejectTokenTypeWithoutRelationships", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.IDJAGRelationships = nil
		}), []string{prefix + "'request_token_types_supported' contains 'urn:ietf:params:oauth:token-type:id-jag' which requires at least one 'identity_assertion_authorization_grant_relationships' entry"}, nil},
		{"ShouldRejectBadSubjectTokenTypes", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.SubjectTokenTypesSupported = []string{oidc.TokenTypeAccessToken}
		}), []string{prefix + "'subject_token_types_supported' must contain 'urn:ietf:params:oauth:token-type:id_token' or 'urn:ietf:params:oauth:token-type:refresh_token' when the 'request_token_types_supported' option contains 'urn:ietf:params:oauth:token-type:id-jag'"}, nil},
		{"ShouldRejectUnknownIssuer", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.IDJAGRelationships[0].Issuer = idjagTestURL(t, "https://other.example.com")
		}), []string{prefix + "'identity_assertion_authorization_grant_relationships': relationship #1: option 'issuer' must reference a configured resource authorization server but 'https://other.example.com' does not match one"}, nil},
		{"ShouldRejectMissingIssuer", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.IDJAGRelationships[0].Issuer = nil
		}), []string{prefix + "'identity_assertion_authorization_grant_relationships': relationship #1: option 'issuer' is required"}, nil},
		{"ShouldRejectDuplicateIssuer", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.IDJAGRelationships = append(c.IDJAGRelationships, relationship(nil))
		}), []string{prefix + "'identity_assertion_authorization_grant_relationships': relationship #2: option 'issuer' must be unique but the value 'https://ras.example.com' is configured multiple times"}, nil},
		{"ShouldRejectMissingClientID", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.IDJAGRelationships[0].ClientID = ""
		}), []string{prefix + "'identity_assertion_authorization_grant_relationships': relationship #1: option 'client_id' is required"}, nil},
		{"ShouldRejectFragmentResource", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.IDJAGRelationships[0].Resources = []string{"https://api.example.com/#frag"}
		}), []string{prefix + "'identity_assertion_authorization_grant_relationships': relationship #1: option 'resources' must only contain absolute URIs without a fragment but the value 'https://api.example.com/#frag' is not"}, nil},
		{"ShouldRejectRelativeResource", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.IDJAGRelationships[0].Resources = []string{"/api"}
		}), []string{prefix + "'identity_assertion_authorization_grant_relationships': relationship #1: option 'resources' must only contain absolute URIs without a fragment but the value '/api' is not"}, nil},
		{"ShouldAcceptAbsoluteResource", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.IDJAGRelationships[0].Resources = []string{"https://api.example.com/v1"}
		}), nil, nil},
		{"ShouldKeepConfiguredAuthorizationDetailsTypes", client(func(c *schema.IdentityProvidersOpenIDConnectClient) {
			c.IDJAGRelationships[0].AuthorizationDetailsTypes = []string{"type"}
		}), nil, func(t *testing.T, actual *schema.IdentityProvidersOpenIDConnectClient) {
			assert.Equal(t, []string{"type"}, actual.IDJAGRelationships[0].AuthorizationDetailsTypes)
		}},
		{"ShouldAcceptNoIDJAG", schema.IdentityProvidersOpenIDConnectClient{ID: "test"}, nil, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				Clients:                             []schema.IdentityProvidersOpenIDConnectClient{tc.have},
				IdentityAssertionAuthorizationGrant: schema.IdentityProvidersOpenIDConnectIDJAG{ResourceAuthorizationServers: []schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{ras}},
			}

			validator := schema.NewStructValidator()

			validateOIDCClientIDJAG(0, config, validator)

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], expected)
			}

			if tc.expected != nil {
				tc.expected(t, &config.Clients[0])
			}
		})
	}
}

func TestValidateOIDCClientIDJAGRequestTokenType(t *testing.T) {
	const prefix = "identity_providers: oidc: clients: client 'test': option "

	enabled := []string{oidc.TokenTypeAccessToken, oidc.TokenTypeRefreshToken, oidc.TokenTypeIDToken, oidc.TokenTypeIDJAG}

	testCases := []struct {
		name   string
		types  []string
		client schema.IdentityProvidersOpenIDConnectClient
		errs   []string
	}{
		{
			"ShouldAcceptRequestTokenTypeWhenEnabled",
			enabled,
			schema.IdentityProvidersOpenIDConnectClient{RequestTokenTypesSupported: []string{oidc.TokenTypeIDJAG}},
			nil,
		},
		{
			"ShouldRejectRequestTokenTypeWhenNotEnabled",
			[]string{oidc.TokenTypeAccessToken, oidc.TokenTypeRefreshToken, oidc.TokenTypeIDToken},
			schema.IdentityProvidersOpenIDConnectClient{RequestTokenTypesSupported: []string{oidc.TokenTypeIDJAG}},
			[]string{prefix + "'request_token_types_supported' must only contain token types enabled by 'token_exchange: token_types' but it has 'urn:ietf:params:oauth:token-type:id-jag'"},
		},
		{
			"ShouldRejectSubjectTokenType",
			enabled,
			schema.IdentityProvidersOpenIDConnectClient{SubjectTokenTypesSupported: []string{oidc.TokenTypeIDToken, oidc.TokenTypeIDJAG}},
			[]string{prefix + "'subject_token_types_supported' must not contain the 'urn:ietf:params:oauth:token-type:id-jag' value as Identity Assertion JWT Authorization Grants cannot be exchanged"},
		},
		{
			"ShouldRejectActorTokenType",
			enabled,
			schema.IdentityProvidersOpenIDConnectClient{ActorTokenTypesSupported: []string{oidc.TokenTypeAccessToken, oidc.TokenTypeIDJAG}},
			[]string{prefix + "'actor_token_types_supported' must not contain the 'urn:ietf:params:oauth:token-type:id-jag' value as Identity Assertion JWT Authorization Grants cannot be exchanged"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := tc.client
			client.ID = "test"
			client.GrantTypes = []string{oidc.GrantTypeTokenExchange}

			config := &schema.IdentityProvidersOpenIDConnect{
				TokenExchange: schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: tc.types, DefaultRequestedTokenType: oidc.TokenTypeAccessToken},
				Clients:       []schema.IdentityProvidersOpenIDConnectClient{client},
			}

			validator := schema.NewStructValidator()

			validateOIDCClientTokenExchange(0, config, validator)

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], expected)
			}
		})
	}
}

func TestValidateOIDCTokenExchangeIDJAG(t *testing.T) {
	const (
		at    = "urn:ietf:params:oauth:token-type:access_token"
		rt    = "urn:ietf:params:oauth:token-type:refresh_token"
		id    = "urn:ietf:params:oauth:token-type:id_token"
		idjag = "urn:ietf:params:oauth:token-type:id-jag"
	)

	testCases := []struct {
		name   string
		types  []string
		ras    bool
		errors []string
	}{
		{"ShouldAcceptListedWithResourceAuthorizationServers", []string{at, rt, id, idjag}, true, nil},
		{"ShouldAcceptNeither", []string{at, rt, id}, false, nil},
		{
			"ShouldRejectResourceAuthorizationServersWhenNotListed",
			[]string{at, rt, id},
			true,
			[]string{"identity_providers: oidc: identity_assertion_authorization_grant: option 'resource_authorization_servers' must not be configured when the 'token_exchange' option 'token_types' does not include '" + idjag + "'"},
		},
		{
			"ShouldRejectResourceAuthorizationServersWhenDefaulted",
			nil,
			true,
			[]string{"identity_providers: oidc: identity_assertion_authorization_grant: option 'resource_authorization_servers' must not be configured when the 'token_exchange' option 'token_types' does not include '" + idjag + "'"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						HMACSecret:       "abc",
						IssuerPrivateKey: keyRSA2048,
						TokenExchange:    schema.IdentityProvidersOpenIDConnectTokenExchange{TokenTypes: tc.types},
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID:                  "good_id",
								Secret:              tOpenIDConnectPBKDF2ClientSecret,
								AuthorizationPolicy: "two_factor",
								RedirectURIs:        []string{"https://google.com/callback"},
							},
						},
					},
				},
			}

			if tc.ras {
				config.IdentityProviders.OIDC.IdentityAssertionAuthorizationGrant.ResourceAuthorizationServers = []schema.IdentityProvidersOpenIDConnectIDJAGResourceAuthorizationServer{
					{Issuer: idjagTestURL(t, "https://ras.example.com"), SubjectClientID: "good_id"},
				}
			}

			ValidateIdentityProviders(NewValidateCtx(), config, validator)

			require.Len(t, validator.Errors(), len(tc.errors))

			for i, expected := range tc.errors {
				assert.EqualError(t, validator.Errors()[i], expected)
			}
		})
	}
}

func TestValidateIDJAGTrustedIssuers(t *testing.T) {
	const prefix = "identity_providers: oidc: identity_assertion_authorization_grant: trusted_issuers: issuer #1: option "

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	valid := func(mutate func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer)) schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer {
		i := schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer{
			Issuer:         idjagTestURL(t, "https://idp.example.com"),
			JSONWebKeysURI: idjagTestURL(t, "https://idp.example.com/jwks.json"),
			Clients:        []string{"bearer"},
			UsernameClaim:  "preferred_username",
		}

		if mutate != nil {
			mutate(&i)
		}

		return i
	}

	testCases := []struct {
		name     string
		have     schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer
		errs     []string
		expected []string
	}{
		{"ShouldAcceptValidAndDefaultAlgorithms", valid(nil), nil, []string{"RS256"}},
		{"ShouldAcceptStaticKeys", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = nil
			i.JSONWebKeys = []schema.JWK{{KeyID: "k", Use: "sig", Algorithm: "ES256", Key: &key.PublicKey}}
			i.SigningAlgorithms = []string{"ES256", "PS512"}
		}), nil, []string{"ES256", "PS512"}},
		{"ShouldRejectMissingIssuer", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) { i.Issuer = nil }), []string{prefix + "'issuer' is required"}, []string{"RS256"}},
		{"ShouldRejectAutheliaIssuer", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.Issuer = idjagTestURL(t, "https://auth.example.com")
		}), []string{prefix + "'issuer' must not be the same as the issuer of this provider but the value 'https://auth.example.com' is"}, []string{"RS256"}},
		{"ShouldRejectBothJWKSAndURI", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeys = []schema.JWK{{KeyID: "k", Use: "sig", Algorithm: "ES256", Key: &key.PublicKey}}
		}), []string{prefix + "'jwks' and 'jwks_uri' are mutually exclusive but both are configured"}, []string{"RS256"}},
		{"ShouldRejectNeitherJWKSNorURI", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) { i.JSONWebKeysURI = nil }), []string{prefix + "one of 'jwks' or 'jwks_uri' is required"}, []string{"RS256"}},
		{"ShouldRejectInsecureJWKSURI", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = idjagTestURL(t, "http://idp.example.com/jwks.json")
		}), []string{prefix + "'jwks_uri' must have the 'https' scheme but the scheme is 'http'"}, []string{"RS256"}},
		{"ShouldRejectJWKWithoutKey", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = nil
			i.JSONWebKeys = []schema.JWK{{KeyID: "k"}}
		}), []string{prefix + "'jwks': key #1 with key id 'k': option 'key' must be provided"}, []string{"RS256"}},
		{"ShouldRejectEmptyJWKSURIHost", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = idjagTestURL(t, "https:///jwks.json")
		}), []string{prefix + "'jwks_uri' must have a host"}, []string{"RS256"}},
		{"ShouldRejectUndersizedRSAKey", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = nil
			i.JSONWebKeys = []schema.JWK{{KeyID: "k", Use: "sig", Algorithm: "RS256", Key: keyRSA1024.Public()}}
		}), []string{prefix + "'jwks': key #1 with key id 'k': option 'key' is an RSA 1024 bit private key but it must at minimum be a RSA 2048 bit private key"}, []string{"RS256"}},
		{"ShouldAcceptRSA2048Key", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = nil
			i.JSONWebKeys = []schema.JWK{{KeyID: "k", Key: keyRSA2048.Public(), CertificateChain: certRSA2048}}
		}), nil, []string{"RS256"}},
		{"ShouldRejectEncryptionUseKey", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = nil
			i.JSONWebKeys = []schema.JWK{{KeyID: "k", Use: "enc", Algorithm: "RS256", Key: keyRSA2048.Public()}}
		}), []string{prefix + "'jwks': key #1 with key id 'k': option 'use' must be one of 'sig' but it's configured as 'enc'"}, []string{"RS256"}},
		{"ShouldRejectDisallowedKeyAlgorithm", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = nil
			i.JSONWebKeys = []schema.JWK{{KeyID: "k", Use: "sig", Algorithm: "EdDSA", Key: &key.PublicKey}}
		}), []string{prefix + "'jwks': key #1 with key id 'k': option 'algorithm' must be one of 'RS256', 'RS384', 'RS512', 'PS256', 'PS384', 'PS512', 'ES256', 'ES384', or 'ES512' but it's configured as 'EdDSA'", prefix + "'jwks': key #1 with key id 'k': option 'key' is not usable: the 'EdDSA' algorithm requires an Ed25519 key but the key is a ECDSA P-256 key"}, []string{"RS256"}},
		{"ShouldRejectMissingKeyID", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = nil
			i.JSONWebKeys = []schema.JWK{{Key: &key.PublicKey}}
		}), []string{prefix + "'jwks': key #1: option 'key_id' must be provided"}, []string{"RS256"}},
		{"ShouldRejectDuplicateKeyID", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = nil
			i.JSONWebKeys = []schema.JWK{{KeyID: "k", Key: &key.PublicKey}, {KeyID: "k", Key: &key.PublicKey}}
		}), []string{prefix + "'jwks': key #2: option 'key_id' must be unique but the value 'k' is configured multiple times"}, []string{"RS256"}},
		{"ShouldRejectMismatchedCertificate", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.JSONWebKeysURI = nil
			i.JSONWebKeys = []schema.JWK{{KeyID: "k", Use: "sig", Algorithm: "RS256", Key: keyRSA2048Legacy.Public(), CertificateChain: certRSA1024}}
		}), []string{prefix + "'jwks': key #1 with key id 'k': option 'certificate_chain' does not appear to contain the public key for the public key provided by option 'key'"}, []string{"RS256"}},
		{"ShouldRejectInvalidSigningAlgorithm", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.SigningAlgorithms = []string{"HS256", "EdDSA"}
		}), []string{
			prefix + "'signing_algorithms' must only contain 'RS256', 'RS384', 'RS512', 'PS256', 'PS384', 'PS512', 'ES256', 'ES384', or 'ES512' but it contains 'HS256'",
			prefix + "'signing_algorithms' must only contain 'RS256', 'RS384', 'RS512', 'PS256', 'PS384', 'PS512', 'ES256', 'ES384', or 'ES512' but it contains 'EdDSA'",
		}, []string{"HS256", "EdDSA"}},
		{"ShouldRejectEmptyClients", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) { i.Clients = nil }), []string{prefix + "'clients' must contain at least one client"}, []string{"RS256"}},
		{"ShouldRejectUnknownClient", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) { i.Clients = []string{"nope"} }), []string{prefix + "'clients' must only reference registered clients but the client 'nope' does not exist"}, []string{"RS256"}},
		{"ShouldRejectPublicClient", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) { i.Clients = []string{"public"} }), []string{prefix + "'clients' must only reference confidential clients but the client 'public' is public"}, []string{"RS256"}},
		{"ShouldRejectUnauthenticatedClient", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) { i.Clients = []string{"none"} }), []string{prefix + "'clients' must only reference clients which authenticate at the token endpoint but the client 'none' has a 'token_endpoint_auth_method' of 'none'"}, []string{"RS256"}},
		{"ShouldRejectClientWithoutGrantType", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) { i.Clients = []string{"plain"} }), []string{prefix + "'clients' must only reference clients which have the 'grant_types' option containing 'urn:ietf:params:oauth:grant-type:jwt-bearer' or 'urn:ietf:params:oauth:grant-type:jwt-dpop' but the client 'plain' does not"}, []string{"RS256"}},
		{"ShouldRejectRelativeResource", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) { i.Resources = []string{"/relative"} }), []string{prefix + "'resources' must only contain absolute URIs but the value '/relative' is not"}, []string{"RS256"}},
		{"ShouldAcceptAbsoluteResource", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.Resources = []string{"https://api.example.com"}
		}), nil, []string{"RS256"}},
		{"ShouldRejectMissingUsernameClaim", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) { i.UsernameClaim = "" }), []string{prefix + "'username_claim' is required"}, []string{"RS256"}},
		{"ShouldAcceptAMRExpression", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.AuthenticationMethodsReferences = `'mfa' in claims.amr ? ['pwd', 'otp'] : ['pwd']`
		}), nil, []string{"RS256"}},
		{"ShouldRejectBadAMRExpression", valid(func(i *schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer) {
			i.AuthenticationMethodsReferences = `'two_factor'`
		}), []string{prefix + "'authentication_methods_references' is not a valid expression: expected the output type to be a list of strings but it was 'string'"}, []string{"RS256"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{ID: "bearer", GrantTypes: []string{oidc.GrantTypeJWTBearer}},
					{ID: "public", Public: true, GrantTypes: []string{oidc.GrantTypeAuthorizationCode}},
					{ID: "plain", GrantTypes: []string{oidc.GrantTypeAuthorizationCode}},
					{ID: "none", TokenEndpointAuthMethod: oidc.ClientAuthMethodNone, ResponseTypes: []string{oidc.ResponseTypeImplicitFlowIDToken}, GrantTypes: []string{oidc.GrantTypeImplicit}},
				},
				IdentityAssertionAuthorizationGrant: schema.IdentityProvidersOpenIDConnectIDJAG{
					TrustedIssuers: []schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer{tc.have, {
						Issuer:         idjagTestURL(t, "https://anchor.example.com"),
						JSONWebKeysURI: idjagTestURL(t, "https://anchor.example.com/jwks.json"),
						Clients:        []string{"bearer"},
						UsernameClaim:  "sub",
					}},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCIDJAG(NewValidateCtx(), config, validator, []*url.URL{idjagTestURL(t, "https://auth.example.com")})

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], expected)
			}

			assert.Equal(t, tc.expected, config.IdentityAssertionAuthorizationGrant.TrustedIssuers[0].SigningAlgorithms)
		})
	}
}

func TestValidateIDJAGTrustedIssuersResolvedConsentMode(t *testing.T) {
	const prefix = "identity_providers: oidc: identity_assertion_authorization_grant: trusted_issuers: issuer #1: option "

	testCases := []struct {
		name     string
		mode     string
		duration *time.Duration
		errs     []string
	}{
		{"ShouldRejectExplicit", "explicit", nil, []string{prefix + "'clients' must only reference clients which have a 'consent_mode' of 'implicit' or 'pre-configured' but the client 'bearer' has a 'consent_mode' of 'explicit'"}},
		{"ShouldRejectAutoResolvingToExplicit", "auto", nil, []string{prefix + "'clients' must only reference clients which have a 'consent_mode' of 'implicit' or 'pre-configured' but the client 'bearer' has a 'consent_mode' of 'explicit'"}},
		{"ShouldRejectEmptyResolvingToExplicit", "", nil, []string{prefix + "'clients' must only reference clients which have a 'consent_mode' of 'implicit' or 'pre-configured' but the client 'bearer' has a 'consent_mode' of 'explicit'"}},
		{"ShouldAcceptAutoResolvingToPreConfigured", "auto", new(time.Hour), nil},
		{"ShouldAcceptImplicit", "implicit", nil, nil},
		{"ShouldAcceptPreConfigured", "pre-configured", nil, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{{ID: "bearer", GrantTypes: []string{oidc.GrantTypeJWTBearer}, ConsentMode: tc.mode, ConsentPreConfiguredDuration: tc.duration}},
				IdentityAssertionAuthorizationGrant: schema.IdentityProvidersOpenIDConnectIDJAG{
					TrustedIssuers: []schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer{
						{Issuer: idjagTestURL(t, "https://idp.example.com"), JSONWebKeysURI: idjagTestURL(t, "https://idp.example.com/jwks"), Clients: []string{"bearer"}, UsernameClaim: "sub"},
					},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCClientConsentMode(0, config, validator, true)
			validateOIDCIDJAG(NewValidateCtx(), config, validator, nil)

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], expected)
			}
		})
	}
}

func TestValidateIDJAGTrustedIssuersPolicyUnsatisfiable(t *testing.T) {
	const prefix = "identity_providers: oidc: identity_assertion_authorization_grant: trusted_issuers: issuer #1: option "

	policies := map[string]schema.IdentityProvidersOpenIDConnectPolicy{
		"strict":  {DefaultPolicy: policyTwoFactor, Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{{Policy: policyDeny}}},
		"mixed":   {DefaultPolicy: policyTwoFactor, Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{{Policy: policyOneFactor}}},
		"lenient": {DefaultPolicy: policyOneFactor},
	}

	testCases := []struct {
		name     string
		policy   string
		amr      string
		warnings []string
	}{
		{"ShouldWarnTwoFactor", policyTwoFactor, "", []string{prefix + "'clients' references the client 'bearer' which has an 'authorization_policy' of 'two_factor' that never permits one factor authentication but 'authentication_methods_references' is not configured, so grants from this issuer will always be rejected for this client"}},
		{"ShouldWarnCustomPolicyWithoutOneFactor", "strict", "", []string{prefix + "'clients' references the client 'bearer' which has an 'authorization_policy' of 'strict' that never permits one factor authentication but 'authentication_methods_references' is not configured, so grants from this issuer will always be rejected for this client"}},
		{"ShouldNotWarnTwoFactorWithAMR", policyTwoFactor, "['pwd', 'otp']", nil},
		{"ShouldNotWarnOneFactor", policyOneFactor, "", nil},
		{"ShouldNotWarnCustomPolicyRuleOneFactor", "mixed", "", nil},
		{"ShouldNotWarnCustomPolicyDefaultOneFactor", "lenient", "", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				AuthorizationPolicies: policies,
				Clients:               []schema.IdentityProvidersOpenIDConnectClient{{ID: "bearer", GrantTypes: []string{oidc.GrantTypeJWTBearer}, ConsentMode: "implicit", AuthorizationPolicy: tc.policy}},
				IdentityAssertionAuthorizationGrant: schema.IdentityProvidersOpenIDConnectIDJAG{
					TrustedIssuers: []schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer{
						{Issuer: idjagTestURL(t, "https://idp.example.com"), JSONWebKeysURI: idjagTestURL(t, "https://idp.example.com/jwks"), Clients: []string{"bearer"}, UsernameClaim: "sub", AuthenticationMethodsReferences: tc.amr},
					},
				},
			}

			validator := schema.NewStructValidator()

			validateOIDCIDJAG(NewValidateCtx(), config, validator, nil)

			assert.Len(t, validator.Errors(), 0)
			require.Len(t, validator.Warnings(), len(tc.warnings))

			for i, expected := range tc.warnings {
				assert.EqualError(t, validator.Warnings()[i], expected)
			}
		})
	}
}

func TestValidateIDJAGTrustedIssuersDuplicateIssuer(t *testing.T) {
	config := &schema.IdentityProvidersOpenIDConnect{
		Clients: []schema.IdentityProvidersOpenIDConnectClient{{ID: "bearer", GrantTypes: []string{oidc.GrantTypeJWTBearer}}},
		IdentityAssertionAuthorizationGrant: schema.IdentityProvidersOpenIDConnectIDJAG{
			TrustedIssuers: []schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer{
				{Issuer: idjagTestURL(t, "https://idp.example.com"), JSONWebKeysURI: idjagTestURL(t, "https://idp.example.com/jwks"), Clients: []string{"bearer"}, UsernameClaim: "sub"},
				{Issuer: idjagTestURL(t, "https://idp.example.com"), JSONWebKeysURI: idjagTestURL(t, "https://idp.example.com/jwks"), Clients: []string{"bearer"}, UsernameClaim: "sub"},
			},
		},
	}

	validator := schema.NewStructValidator()

	validateOIDCIDJAG(NewValidateCtx(), config, validator, nil)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "identity_providers: oidc: identity_assertion_authorization_grant: trusted_issuers: issuer #2: option 'issuer' must be unique but the value 'https://idp.example.com' is configured multiple times")
}

func TestValidateOIDCClientIDJAGGrantTypes(t *testing.T) {
	trusted := func(clients ...string) schema.IdentityProvidersOpenIDConnectIDJAG {
		return schema.IdentityProvidersOpenIDConnectIDJAG{
			TrustedIssuers: []schema.IdentityProvidersOpenIDConnectIDJAGTrustedIssuer{
				{Issuer: idjagTestURL(t, "https://idp.example.com"), JSONWebKeysURI: idjagTestURL(t, "https://idp.example.com/jwks"), Clients: append([]string{"listed"}, clients...), UsernameClaim: "sub"},
			},
		}
	}

	testCases := []struct {
		name      string
		grantType string
		dpop      bool
		idjag     schema.IdentityProvidersOpenIDConnectIDJAG
		errs      []string
	}{
		{"ShouldAcceptJWTBearerListed", oidc.GrantTypeJWTBearer, false, trusted("test"), nil},
		{"ShouldAcceptJWTDPoPListedWithDPoP", oidc.GrantTypeJWTDPoP, true, trusted("test"), nil},
		{"ShouldRejectJWTBearerNotListed", oidc.GrantTypeJWTBearer, false, trusted(), []string{"identity_providers: oidc: clients: client 'test': option 'grant_types' contains 'urn:ietf:params:oauth:grant-type:jwt-bearer' which requires the client to be listed in the 'clients' option of at least one trusted issuer"}},
		{"ShouldRejectJWTDPoPWithoutDPoP", oidc.GrantTypeJWTDPoP, false, trusted("test"), []string{"identity_providers: oidc: clients: client 'test': option 'grant_types' contains 'urn:ietf:params:oauth:grant-type:jwt-dpop' which requires the 'dpop' option 'enabled' to be true"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.IdentityProvidersOpenIDConnect{
				DPoP: schema.IdentityProvidersOpenIDConnectDPoP{Enabled: tc.dpop},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{ID: "test", GrantTypes: []string{tc.grantType}},
					{ID: "listed", GrantTypes: []string{oidc.GrantTypeJWTBearer}},
				},
				IdentityAssertionAuthorizationGrant: tc.idjag,
			}

			validator := schema.NewStructValidator()

			validateOIDCIDJAG(NewValidateCtx(), config, validator, nil)

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], expected)
			}
		})
	}

	t.Run("ShouldAcceptGrantTypesAsValid", func(t *testing.T) {
		config := &schema.IdentityProvidersOpenIDConnect{
			Clients: []schema.IdentityProvidersOpenIDConnectClient{{ID: "test", GrantTypes: []string{oidc.GrantTypeJWTBearer, oidc.GrantTypeJWTDPoP}}},
		}

		validator := schema.NewStructValidator()

		validateOIDCClientGrantTypes(0, config, validator, false, func() {})

		assert.Empty(t, validator.Errors())
	})
}
