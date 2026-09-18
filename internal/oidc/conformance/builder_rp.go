// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package conformance

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

// RelyingPartyBuilder builds a Suite which asserts the conformance of Authelia as an OpenID Connect 1.0 Relying Party.
// Its plan's client is the external provider registration Authelia uses, and its providers are what Authelia is
// configured with to use the conformance suite as that external provider.
type RelyingPartyBuilder struct {
	Brand       string
	Name        string
	Friendly    string
	Version     string
	Module      string
	SuiteURL    *url.URL
	AutheliaURL *url.URL
}

// Description returns the plan description naming release as the version of Authelia under test.
func (b *RelyingPartyBuilder) Description(release string) string {
	return fmt.Sprintf("Authelia %s %s Relying Party Certification Profile", release, b.Friendly)
}

// ProviderID returns the id of the provider Authelia is configured with for this builder. It is independent of the
// version so the sign in button the suite drives has a stable identifier.
func (b *RelyingPartyBuilder) ProviderID() string {
	return "conformance-" + b.Name
}

// Build returns the Suite for this builder.
func (b *RelyingPartyBuilder) Build() Suite {
	var apiname, mode string

	variant := &PlanVariant{
		ClientRegistration: "static_client",
		RequestType:        "plain_http_request",
	}

	switch {
	case b.Module != "":
		apiname, mode = "oidcc-client-config-certification-test-plan", "query"
		variant.ClientAuthType, variant.ResponseMode = "client_secret_basic", "default"
	case b.Name == NameRelyingPartyBasic:
		apiname, mode = "oidcc-client-basic-certification-test-plan", "query"
	case b.Name == NameRelyingPartyBasicFormPost:
		apiname, mode = "oidcc-client-formpost-basic-certification-test-plan", "form_post"
	}

	alias := fmt.Sprintf("conformance-%s-%s%s", b.Name, b.Brand, strings.ReplaceAll(strings.ToLower(b.Version), ".", ""))
	id := b.ProviderID()
	secret := random.New().StringCustom(80, random.CharSetAlphaNumeric)

	suite := Suite{
		Name: "conformance-" + b.Name,
		Plan: Plan{
			Name:        apiname,
			Alias:       alias,
			Description: b.Description(b.Version),
			Publish:     "summary",
			Variant:     variant,
			Client: &PlanClient{
				ID:          alias,
				Secret:      secret,
				RedirectURI: b.AutheliaURL.JoinPath("api", "identity", id, "callback").String(),
			},
		},
	}

	issuer := b.SuiteURL.JoinPath("test", "a", alias).String() + "/"

	suite.Providers = []schema.AuthenticationBackendExternalIdentityProvider{
		{
			ID:                      id,
			Name:                    b.Friendly,
			Issuer:                  issuer,
			ClientID:                suite.Plan.Client.ID,
			ClientSecret:            secret,
			Scopes:                  []string{"openid", "profile", "email"},
			ResponseMode:            mode,
			TokenEndpointAuthMethod: "client_secret_basic",
		},
	}

	return suite
}

// RelyingPartyBuilders returns the conformance suite builders for every Relying Party profile Authelia is tested
// against, in a fixed order.
func RelyingPartyBuilders(version, brand string, suiteURL, autheliaURL *url.URL) (builders []*RelyingPartyBuilder) {
	builders = make([]*RelyingPartyBuilder, 0, 2+len(RelyingPartyConfigModules))

	builders = append(builders,
		&RelyingPartyBuilder{Brand: brand, Name: NameRelyingPartyBasic, Friendly: "Basic RP", Version: version, SuiteURL: suiteURL, AutheliaURL: autheliaURL},
		&RelyingPartyBuilder{Brand: brand, Name: NameRelyingPartyBasicFormPost, Friendly: "Form Post Basic RP", Version: version, SuiteURL: suiteURL, AutheliaURL: autheliaURL},
	)

	for _, module := range RelyingPartyConfigModules {
		builders = append(builders, &RelyingPartyBuilder{
			Brand:       brand,
			Name:        module.BuilderName(),
			Friendly:    fmt.Sprintf("Config RP (%s)", module.Suffix),
			Version:     version,
			SuiteURL:    suiteURL,
			AutheliaURL: autheliaURL,
			Module:      module.Name,
		})
	}

	return builders
}
