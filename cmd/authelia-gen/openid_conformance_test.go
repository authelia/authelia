package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestMustHash(t *testing.T) {
	assert.NotPanics(t, func() {
		MustHash("password")
	})
}

func TestOpenIDConnectConformanceSuiteBuilder_Build(t *testing.T) {
	suiteURL := &url.URL{
		Scheme: "https",
		Host:   "conformance.example.com",
	}

	autheliaURL := &url.URL{
		Scheme: "https",
		Host:   "auth.example.com",
	}

	secret := MustHash("password")

	testCases := []struct {
		name     string
		have     *OpenIDConnectConformanceSuiteBuilder
		expected OpenIDConnectConformanceSuite
	}{
		{
			"ShouldHandleConfig",
			&OpenIDConnectConformanceSuiteBuilder{"authelia", "config", "Config", true, "4.40", "implicit", "one_factor", nil, autheliaURL},
			OpenIDConnectConformanceSuite{
				Name: "conformance-config",
				Plan: OpenIDConnectConformanceSuitePlan{
					Name:        "oidcc-config-certification-test-plan",
					Alias:       "conformance-config-authelia440",
					Description: "Authelia 4.40 Config Certification Profile",
					Publish:     "summary",
					Server: OpenIDConnectConformanceSuitePlanServer{
						DiscoveryURL: "https://auth.example.com/.well-known/openid-configuration",
					},
				},
			},
		},
		{
			"ShouldHandleBasic",
			&OpenIDConnectConformanceSuiteBuilder{"authelia", "basic", "Basic", true, "4.40", "implicit", "one_factor", suiteURL, autheliaURL},
			OpenIDConnectConformanceSuite{
				Name: "conformance-basic",
				Plan: OpenIDConnectConformanceSuitePlan{
					Name:        "oidcc-basic-certification-test-plan",
					Alias:       "conformance-basic-authelia440",
					Description: "Authelia 4.40 Basic Certification Profile",
					Publish:     "summary",
					Variant: &OpenIDConnectConformanceSuitePlanVariant{
						ServerMetadata:     "discovery",
						ClientRegistration: "static_client",
					},
					Server: OpenIDConnectConformanceSuitePlanServer{
						DiscoveryURL: "https://auth.example.com/.well-known/openid-configuration",
					},
					Client: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-basic-authelia440",
						Secret: "present",
					},
					ClientAlternate: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-basic-authelia440-alt",
						Secret: "present",
					},
					ClientSecretPost: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-basic-authelia440-post",
						Secret: "present",
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:     "conformance-certification-basic-authelia440",
						Name:   "Authelia 4.40 Basic Certification Profile",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-basic-authelia440/callback",
						},
						ResponseModes:           []string{"query", "query.jwt"},
						ResponseTypes:           []string{"code"},
						GrantTypes:              []string{"authorization_code", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-basic-authelia440-alt",
						Name:   "Authelia 4.40 Basic Certification Profile (Alternate)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-basic-authelia440/callback",
						},
						ResponseModes:           []string{"query", "query.jwt"},
						ResponseTypes:           []string{"code"},
						GrantTypes:              []string{"authorization_code", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-basic-authelia440-post",
						Name:   "Authelia 4.40 Basic Certification Profile (Secret Post)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-basic-authelia440/callback",
						},
						ResponseModes:           []string{"query", "query.jwt"},
						ResponseTypes:           []string{"code"},
						GrantTypes:              []string{"authorization_code", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_post",
					},
				},
			},
		},
		{
			"ShouldHandleBasicFormPost",
			&OpenIDConnectConformanceSuiteBuilder{"authelia", "basic-form-post", "Basic (Form Post)", true, "4.40", "implicit", "one_factor", suiteURL, autheliaURL},
			OpenIDConnectConformanceSuite{
				Name: "conformance-basic-form-post",
				Plan: OpenIDConnectConformanceSuitePlan{
					Name:        "oidcc-basic-form-post-certification-test-plan",
					Alias:       "conformance-basic-form-post-authelia440",
					Description: "Authelia 4.40 Basic (Form Post) Certification Profile",
					Publish:     "summary",
					Variant: &OpenIDConnectConformanceSuitePlanVariant{
						ServerMetadata:     "discovery",
						ClientRegistration: "static_client",
					},
					Server: OpenIDConnectConformanceSuitePlanServer{
						DiscoveryURL: "https://auth.example.com/.well-known/openid-configuration",
					},
					Client: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-basic-form-post-authelia440",
						Secret: "present",
					},
					ClientAlternate: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-basic-form-post-authelia440-alt",
						Secret: "present",
					},
					ClientSecretPost: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-basic-form-post-authelia440-post",
						Secret: "present",
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:     "conformance-certification-basic-form-post-authelia440",
						Name:   "Authelia 4.40 Basic (Form Post) Certification Profile",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-basic-form-post-authelia440/callback",
						},
						ResponseModes:           []string{"form_post", "form_post.jwt"},
						ResponseTypes:           []string{"code"},
						GrantTypes:              []string{"authorization_code", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-basic-form-post-authelia440-alt",
						Name:   "Authelia 4.40 Basic (Form Post) Certification Profile (Alternate)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-basic-form-post-authelia440/callback",
						},
						ResponseModes:           []string{"form_post", "form_post.jwt"},
						ResponseTypes:           []string{"code"},
						GrantTypes:              []string{"authorization_code", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-basic-form-post-authelia440-post",
						Name:   "Authelia 4.40 Basic (Form Post) Certification Profile (Secret Post)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-basic-form-post-authelia440/callback",
						},
						ResponseModes:           []string{"form_post", "form_post.jwt"},
						ResponseTypes:           []string{"code"},
						GrantTypes:              []string{"authorization_code", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_post",
					},
				},
			},
		},
		{
			"ShouldHandleImplicit",
			&OpenIDConnectConformanceSuiteBuilder{"authelia", "implicit", "Implicit", true, "4.40", "implicit", "one_factor", suiteURL, autheliaURL},
			OpenIDConnectConformanceSuite{
				Name: "conformance-implicit",
				Plan: OpenIDConnectConformanceSuitePlan{
					Name:        "oidcc-implicit-certification-test-plan",
					Alias:       "conformance-implicit-authelia440",
					Description: "Authelia 4.40 Implicit Certification Profile",
					Publish:     "summary",
					Variant: &OpenIDConnectConformanceSuitePlanVariant{
						ServerMetadata:     "discovery",
						ClientRegistration: "static_client",
					},
					Server: OpenIDConnectConformanceSuitePlanServer{
						DiscoveryURL: "https://auth.example.com/.well-known/openid-configuration",
					},
					Client: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-implicit-authelia440",
						Secret: "present",
					},
					ClientAlternate: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-implicit-authelia440-alt",
						Secret: "present",
					},
					ClientSecretPost: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-implicit-authelia440-post",
						Secret: "present",
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:     "conformance-certification-implicit-authelia440",
						Name:   "Authelia 4.40 Implicit Certification Profile",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-implicit-authelia440/callback",
						},
						ResponseModes:           []string{"query", "query.jwt"},
						ResponseTypes:           []string{"code", "id_token", "token", "id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-implicit-authelia440-alt",
						Name:   "Authelia 4.40 Implicit Certification Profile (Alternate)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-implicit-authelia440/callback",
						},
						ResponseModes:           []string{"query", "query.jwt"},
						ResponseTypes:           []string{"code", "id_token", "token", "id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-implicit-authelia440-post",
						Name:   "Authelia 4.40 Implicit Certification Profile (Secret Post)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-implicit-authelia440/callback",
						},
						ResponseModes:           []string{"query", "query.jwt"},
						ResponseTypes:           []string{"code", "id_token", "token", "id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_post",
					},
				},
			},
		},
		{
			"ShouldHandleImplicitFormPost",
			&OpenIDConnectConformanceSuiteBuilder{"authelia", "implicit-form-post", "Implicit (Form Post)", true, "4.40", "implicit", "one_factor", suiteURL, autheliaURL},
			OpenIDConnectConformanceSuite{
				Name: "conformance-implicit-form-post",
				Plan: OpenIDConnectConformanceSuitePlan{
					Name:        "oidcc-implicit-form-post-certification-test-plan",
					Alias:       "conformance-implicit-form-post-authelia440",
					Description: "Authelia 4.40 Implicit (Form Post) Certification Profile",
					Publish:     "summary",
					Variant: &OpenIDConnectConformanceSuitePlanVariant{
						ServerMetadata:     "discovery",
						ClientRegistration: "static_client",
					},
					Server: OpenIDConnectConformanceSuitePlanServer{
						DiscoveryURL: "https://auth.example.com/.well-known/openid-configuration",
					},
					Client: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-implicit-form-post-authelia440",
						Secret: "present",
					},
					ClientAlternate: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-implicit-form-post-authelia440-alt",
						Secret: "present",
					},
					ClientSecretPost: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-implicit-form-post-authelia440-post",
						Secret: "present",
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:     "conformance-certification-implicit-form-post-authelia440",
						Name:   "Authelia 4.40 Implicit (Form Post) Certification Profile",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-implicit-form-post-authelia440/callback",
						},
						ResponseModes:           []string{"form_post", "form_post.jwt"},
						ResponseTypes:           []string{"code", "id_token", "token", "id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-implicit-form-post-authelia440-alt",
						Name:   "Authelia 4.40 Implicit (Form Post) Certification Profile (Alternate)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-implicit-form-post-authelia440/callback",
						},
						ResponseModes:           []string{"form_post", "form_post.jwt"},
						ResponseTypes:           []string{"code", "id_token", "token", "id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-implicit-form-post-authelia440-post",
						Name:   "Authelia 4.40 Implicit (Form Post) Certification Profile (Secret Post)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-implicit-form-post-authelia440/callback",
						},
						ResponseModes:           []string{"form_post", "form_post.jwt"},
						ResponseTypes:           []string{"code", "id_token", "token", "id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_post",
					},
				},
			},
		},
		{
			"ShouldHandleHybrid",
			&OpenIDConnectConformanceSuiteBuilder{"authelia", "hybrid", "Hybrid", true, "4.40", "implicit", "one_factor", suiteURL, autheliaURL},
			OpenIDConnectConformanceSuite{
				Name: "conformance-hybrid",
				Plan: OpenIDConnectConformanceSuitePlan{
					Name:        "oidcc-hybrid-certification-test-plan",
					Alias:       "conformance-hybrid-authelia440",
					Description: "Authelia 4.40 Hybrid Certification Profile",
					Publish:     "summary",
					Variant: &OpenIDConnectConformanceSuitePlanVariant{
						ServerMetadata:     "discovery",
						ClientRegistration: "static_client",
					},
					Server: OpenIDConnectConformanceSuitePlanServer{
						DiscoveryURL: "https://auth.example.com/.well-known/openid-configuration",
					},
					Client: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-hybrid-authelia440",
						Secret: "present",
					},
					ClientAlternate: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-hybrid-authelia440-alt",
						Secret: "present",
					},
					ClientSecretPost: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-hybrid-authelia440-post",
						Secret: "present",
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:     "conformance-certification-hybrid-authelia440",
						Name:   "Authelia 4.40 Hybrid Certification Profile",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-hybrid-authelia440/callback",
						},
						ResponseModes:           []string{"query", "query.jwt"},
						ResponseTypes:           []string{"code", "code id_token", "code token", "code id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-hybrid-authelia440-alt",
						Name:   "Authelia 4.40 Hybrid Certification Profile (Alternate)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-hybrid-authelia440/callback",
						},
						ResponseModes:           []string{"query", "query.jwt"},
						ResponseTypes:           []string{"code", "code id_token", "code token", "code id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-hybrid-authelia440-post",
						Name:   "Authelia 4.40 Hybrid Certification Profile (Secret Post)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-hybrid-authelia440/callback",
						},
						ResponseModes:           []string{"query", "query.jwt"},
						ResponseTypes:           []string{"code", "code id_token", "code token", "code id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_post",
					},
				},
			},
		},
		{
			"ShouldHandleHybridFormPost",
			&OpenIDConnectConformanceSuiteBuilder{"authelia", "hybrid-form-post", "Hybrid (Form Post)", true, "4.40", "implicit", "one_factor", suiteURL, autheliaURL},
			OpenIDConnectConformanceSuite{
				Name: "conformance-hybrid-form-post",
				Plan: OpenIDConnectConformanceSuitePlan{
					Name:        "oidcc-hybrid-form-post-certification-test-plan",
					Alias:       "conformance-hybrid-form-post-authelia440",
					Description: "Authelia 4.40 Hybrid (Form Post) Certification Profile",
					Publish:     "summary",
					Variant: &OpenIDConnectConformanceSuitePlanVariant{
						ServerMetadata:     "discovery",
						ClientRegistration: "static_client",
					},
					Server: OpenIDConnectConformanceSuitePlanServer{
						DiscoveryURL: "https://auth.example.com/.well-known/openid-configuration",
					},
					Client: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-hybrid-form-post-authelia440",
						Secret: "present",
					},
					ClientAlternate: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-hybrid-form-post-authelia440-alt",
						Secret: "present",
					},
					ClientSecretPost: &OpenIDConnectConformanceSuitePlanClient{
						ID:     "conformance-certification-hybrid-form-post-authelia440-post",
						Secret: "present",
					},
				},
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:     "conformance-certification-hybrid-form-post-authelia440",
						Name:   "Authelia 4.40 Hybrid (Form Post) Certification Profile",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-hybrid-form-post-authelia440/callback",
						},
						ResponseModes:           []string{"form_post", "form_post.jwt"},
						ResponseTypes:           []string{"code", "code id_token", "code token", "code id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-hybrid-form-post-authelia440-alt",
						Name:   "Authelia 4.40 Hybrid (Form Post) Certification Profile (Alternate)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-hybrid-form-post-authelia440/callback",
						},
						ResponseModes:           []string{"form_post", "form_post.jwt"},
						ResponseTypes:           []string{"code", "code id_token", "code token", "code id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_basic",
					},
					{
						ID:     "conformance-certification-hybrid-form-post-authelia440-post",
						Name:   "Authelia 4.40 Hybrid (Form Post) Certification Profile (Secret Post)",
						Secret: secret,
						RedirectURIs: []string{
							"https://conformance.example.com/test/a/conformance-hybrid-form-post-authelia440/callback",
						},
						ResponseModes:           []string{"form_post", "form_post.jwt"},
						ResponseTypes:           []string{"code", "code id_token", "code token", "code id_token token"},
						GrantTypes:              []string{"authorization_code", "implicit", "refresh_token"},
						TokenEndpointAuthMethod: "client_secret_post",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.have.Build()

			assert.Equal(t, tc.expected.Name, actual.Name)
			require.Equal(t, len(tc.expected.Clients), len(actual.Clients))

			for i, expected := range tc.expected.Clients {
				assert.Equal(t, expected.ID, actual.Clients[i].ID)
				assert.Equal(t, expected.Name, actual.Clients[i].Name)
				assert.Equal(t, expected.Secret.Valid(), actual.Clients[i].Secret.Valid())

				assert.Equal(t, expected.RedirectURIs, actual.Clients[i].RedirectURIs)
				assert.Equal(t, expected.ResponseModes, actual.Clients[i].ResponseModes)
				assert.Equal(t, expected.ResponseTypes, actual.Clients[i].ResponseTypes)
				assert.Equal(t, expected.GrantTypes, actual.Clients[i].GrantTypes)
				assert.Equal(t, expected.TokenEndpointAuthMethod, actual.Clients[i].TokenEndpointAuthMethod)
			}

			if tc.expected.Plan.Client != nil {
				require.NotNil(t, actual.Plan.Client)
				assert.Equal(t, tc.expected.Plan.Client.ID, actual.Plan.Client.ID)

				if len(tc.expected.Plan.Client.Secret) != 0 {
					assert.NotEmpty(t, actual.Plan.Client.Secret)
				} else {
					assert.Empty(t, actual.Plan.Client.Secret)
				}
			} else {
				assert.Nil(t, actual.Plan.Client)
			}

			if tc.expected.Plan.ClientAlternate != nil {
				require.NotNil(t, actual.Plan.ClientAlternate)
				assert.Equal(t, tc.expected.Plan.ClientAlternate.ID, actual.Plan.ClientAlternate.ID)
			} else {
				assert.Nil(t, actual.Plan.ClientAlternate)
			}

			if tc.expected.Plan.ClientSecretPost != nil {
				require.NotNil(t, actual.Plan.ClientSecretPost)
				assert.Equal(t, tc.expected.Plan.ClientSecretPost.ID, actual.Plan.ClientSecretPost.ID)
			} else {
				assert.Nil(t, actual.Plan.ClientSecretPost)
			}
		})
	}
}
