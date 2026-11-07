// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package conformance

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
)

func TestRelyingPartyBuilder_Build(t *testing.T) {
	suiteURL := &url.URL{Scheme: "https", Host: "conformance.example.com:8443"}
	autheliaURL := &url.URL{Scheme: "https", Host: "login.example.com:8080"}

	basic := &PlanVariant{ClientRegistration: "static_client", RequestType: "plain_http_request"}
	config := &PlanVariant{ClientRegistration: "static_client", ClientAuthType: "client_secret_basic", ResponseMode: "default", RequestType: "plain_http_request"}

	testCases := []struct {
		name     string
		builder  string
		plan     string
		provider string
		mode     string
		friendly string
		module   string
		variant  *PlanVariant
	}{
		{"ShouldHandleBasic", NameRelyingPartyBasic, "oidcc-client-basic-certification-test-plan", "conformance-rp-basic", "query", "Basic RP", "", basic},
		{"ShouldHandleBasicFormPost", NameRelyingPartyBasicFormPost, "oidcc-client-formpost-basic-certification-test-plan", "conformance-rp-basic-form-post", "form_post", "Form Post Basic RP", "", basic},
		{"ShouldHandleConfigDiscovery", "rp-config-discovery", "oidcc-client-config-certification-test-plan", "conformance-rp-config-discovery", "query", "Config RP (discovery)", "oidcc-client-test-discovery-openid-config", config},
		{"ShouldHandleConfigJWKS", "rp-config-jwks", "oidcc-client-config-certification-test-plan", "conformance-rp-config-jwks", "query", "Config RP (jwks)", "oidcc-client-test-discovery-jwks-uri-keys", config},
		{"ShouldHandleConfigIssuer", "rp-config-issuer", "oidcc-client-config-certification-test-plan", "conformance-rp-config-issuer", "query", "Config RP (issuer)", "oidcc-client-test-discovery-issuer-mismatch", config},
		{"ShouldHandleConfigSigNone", "rp-config-sig-none", "oidcc-client-config-certification-test-plan", "conformance-rp-config-sig-none", "query", "Config RP (sig-none)", "oidcc-client-test-idtoken-sig-none", config},
		{"ShouldHandleConfigSigning", "rp-config-signing", "oidcc-client-config-certification-test-plan", "conformance-rp-config-signing", "query", "Config RP (signing)", "oidcc-client-test-signing-key-rotation-just-before-signing", config},
		{"ShouldHandleConfigRotation", "rp-config-rotation", "oidcc-client-config-certification-test-plan", "conformance-rp-config-rotation", "query", "Config RP (rotation)", "oidcc-client-test-signing-key-rotation", config},
	}

	builders := map[string]*RelyingPartyBuilder{}

	for _, builder := range RelyingPartyBuilders("k3j2x9ab", "authelia", suiteURL, autheliaURL) {
		builders[builder.Name] = builder
	}

	require.Len(t, builders, len(testCases))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder, ok := builders[tc.builder]
			require.True(t, ok)

			assert.Equal(t, tc.module, builder.Module)

			suite := builder.Build()

			alias := "conformance-" + tc.builder + "-autheliak3j2x9ab"

			assert.Equal(t, tc.plan, suite.Plan.Name)
			assert.Equal(t, alias, suite.Plan.Alias)
			assert.Equal(t, "Authelia k3j2x9ab "+tc.friendly+" Relying Party Certification Profile", suite.Plan.Description)
			assert.Equal(t, tc.variant, suite.Plan.Variant)
			assert.Empty(t, suite.Clients)

			require.NotNil(t, suite.Plan.Client)
			assert.Equal(t, alias, suite.Plan.Client.ID)
			assert.Len(t, suite.Plan.Client.Secret, 80)
			assert.Equal(t, "https://login.example.com:8080/api/identity/"+tc.provider+"/callback", suite.Plan.Client.RedirectURI)

			require.Len(t, suite.Providers, 1)

			provider := suite.Providers[0]

			assert.Equal(t, tc.provider, provider.ID)
			assert.Equal(t, tc.provider, builder.ProviderID())
			assert.Equal(t, "https://conformance.example.com:8443/test/a/"+alias+"/", provider.Issuer)
			assert.Equal(t, suite.Plan.Client.ID, provider.ClientID)
			assert.Equal(t, suite.Plan.Client.Secret, provider.ClientSecret)
			assert.Equal(t, tc.mode, provider.ResponseMode)

			// The generated provider must be one Authelia accepts as configured, or the suite fails to start rather
			// than failing a test.
			config := &schema.AuthenticationBackendExternalIdentity{Providers: suite.Providers}
			val := schema.NewStructValidator()

			validator.ValidateAuthenticationBackendExternalIdentity(config, val)

			assert.Empty(t, val.Errors())
			assert.Empty(t, val.Warnings())
		})
	}
}

func TestRelyingPartyBuilders_ShouldGenerateProvidersAutheliaAcceptsTogether(t *testing.T) {
	suiteURL := &url.URL{Scheme: "https", Host: "conformance.example.com:8443"}
	autheliaURL := &url.URL{Scheme: "https", Host: "login.example.com:8080"}

	config := &schema.AuthenticationBackendExternalIdentity{}

	for _, builder := range RelyingPartyBuilders("k3j2x9ab", "authelia", suiteURL, autheliaURL) {
		config.Providers = append(config.Providers, builder.Build().Providers...)
	}

	// Every plan's provider is configured in the one Authelia, so no two of them may share an issuer or an id.
	val := schema.NewStructValidator()

	validator.ValidateAuthenticationBackendExternalIdentity(config, val)

	assert.Empty(t, val.Errors())
	assert.Empty(t, val.Warnings())
}

func TestRelyingPartyConfigModules(t *testing.T) {
	names := make([]string, 0, len(RelyingPartyConfigModules))

	for _, module := range RelyingPartyConfigModules {
		names = append(names, module.Name)
	}

	assert.Equal(t, []string{
		"oidcc-client-test-discovery-openid-config",
		"oidcc-client-test-discovery-jwks-uri-keys",
		"oidcc-client-test-discovery-issuer-mismatch",
		"oidcc-client-test-idtoken-sig-none",
		"oidcc-client-test-signing-key-rotation-just-before-signing",
		"oidcc-client-test-signing-key-rotation",
	}, names, "the modules must be those of the Config RP plan, in the plan's order")
}
