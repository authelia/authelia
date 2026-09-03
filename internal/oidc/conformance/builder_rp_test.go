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

	testCases := []struct {
		name     string
		builder  string
		plan     string
		provider string
		mode     string
		friendly string
	}{
		{"ShouldHandleBasic", NameRelyingPartyBasic, "oidcc-client-basic-certification-test-plan", "conformance-rp-basic", "query", "Basic RP"},
		{"ShouldHandleBasicFormPost", NameRelyingPartyBasicFormPost, "oidcc-client-formpost-basic-certification-test-plan", "conformance-rp-basic-form-post", "form_post", "Form Post Basic RP"},
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

			suite := builder.Build()

			alias := "conformance-" + tc.builder + "-autheliak3j2x9ab"

			assert.Equal(t, tc.plan, suite.Plan.Name)
			assert.Equal(t, alias, suite.Plan.Alias)
			assert.Equal(t, "Authelia k3j2x9ab "+tc.friendly+" Relying Party Certification Profile", suite.Plan.Description)
			assert.Equal(t, &PlanVariant{ClientRegistration: "static_client", RequestType: "plain_http_request"}, suite.Plan.Variant)
			assert.Empty(t, suite.Clients)

			require.NotNil(t, suite.Plan.Client)
			assert.Equal(t, alias, suite.Plan.Client.ID)
			assert.Len(t, suite.Plan.Client.Secret, 80)
			assert.Equal(t, "https://login.example.com:8080/api/firstfactor/external-identity/"+tc.provider+"/callback", suite.Plan.Client.RedirectURI)

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
