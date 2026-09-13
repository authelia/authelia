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

	switch b.Name {
	case NameRelyingPartyBasic:
		apiname, mode = "oidcc-client-basic-certification-test-plan", "query"
	case NameRelyingPartyBasicFormPost:
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
			Variant: &PlanVariant{
				ClientRegistration: "static_client",
				RequestType:        "plain_http_request",
			},
			Client: &PlanClient{
				ID:          alias,
				Secret:      secret,
				RedirectURI: b.AutheliaURL.JoinPath("api", "firstfactor", "external-identity", id, "callback").String(),
			},
		},
	}

	// The conformance suite serves an aliased plan's provider under the alias with a trailing slash, and Authelia
	// compares the issuer exactly, so the slash is part of the issuer rather than a formatting detail.
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
func RelyingPartyBuilders(version, brand string, suiteURL, autheliaURL *url.URL) []*RelyingPartyBuilder {
	return []*RelyingPartyBuilder{
		{brand, NameRelyingPartyBasic, "Basic RP", version, suiteURL, autheliaURL},
		{brand, NameRelyingPartyBasicFormPost, "Form Post Basic RP", version, suiteURL, autheliaURL},
	}
}
