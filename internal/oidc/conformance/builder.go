// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package conformance

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-crypt/crypt/algorithm/pbkdf2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
)

// SuiteBuilder builds a Suite.
type SuiteBuilder struct {
	Brand         string
	Name          string
	Friendly      string
	Certification bool
	Version       string
	Consent       string
	Policy        string
	SuiteURL      *url.URL
	AutheliaURL   *url.URL
}

// Description returns the plan description naming release as the version of Authelia under test.
func (b *SuiteBuilder) Description(release string) string {
	if b.Certification {
		return fmt.Sprintf("Authelia %s %s Certification Profile", release, b.Friendly)
	}

	return fmt.Sprintf("Authelia %s %s Test Profile", release, b.Friendly)
}

// Build returns the Suite for this builder.
func (b *SuiteBuilder) Build() Suite {
	var (
		apiname, namePrefix, clientIDPrefix string
		variant                             *PlanVariant
	)

	if b.Certification {
		namePrefix = "conformance-"
		clientIDPrefix = "conformance-certification"
	} else {
		clientIDPrefix = "conformance-test"
	}

	aliasSuffix := fmt.Sprintf("%s-%s", strings.ReplaceAll(strings.ToLower(b.Name), ".", "-"), b.Brand+strings.ReplaceAll(strings.ToLower(b.Version), ".", ""))

	name := fmt.Sprintf("%s%s", namePrefix, b.Name)
	description := b.Description(b.Version)

	switch b.Name {
	case NameBasic, NameBasicFormPost, NameHybrid, NameHybridFormPost, NameImplicit, NameImplicitFormPost:
		variant = &PlanVariant{
			ServerMetadata:     "discovery",
			ClientRegistration: "static_client",
		}
	}

	switch b.Name {
	case NameConfig:
		apiname = "oidcc-config-certification-test-plan"
	case NameBasic:
		apiname = "oidcc-basic-certification-test-plan"
	case NameBasicFormPost:
		apiname = "oidcc-formpost-basic-certification-test-plan"
	case NameHybrid:
		apiname = "oidcc-hybrid-certification-test-plan"
	case NameHybridFormPost:
		apiname = "oidcc-formpost-hybrid-certification-test-plan"
	case NameImplicit:
		apiname = "oidcc-implicit-certification-test-plan"
	case NameImplicitFormPost:
		apiname = "oidcc-formpost-implicit-certification-test-plan"
	}

	suite := Suite{
		Name: name,
		Plan: Plan{
			Name:        apiname,
			Variant:     variant,
			Alias:       fmt.Sprintf("%s%s", namePrefix, aliasSuffix),
			Publish:     "summary",
			Description: description,
			Server: PlanServer{
				DiscoveryURL: b.AutheliaURL.JoinPath(".well-known/openid-configuration").String(),
			},
		},
	}

	if b.SuiteURL == nil {
		return suite
	}

	r := random.New()

	secret := r.StringCustom(80, random.CharSetAlphaNumeric)
	secretAlternate := r.StringCustom(80, random.CharSetAlphaNumeric)
	secretPost := r.StringCustom(80, random.CharSetAlphaNumeric)

	suite.Plan.Client = &PlanClient{
		ID:     fmt.Sprintf("%s-%s", clientIDPrefix, aliasSuffix),
		Secret: secret,
	}

	suite.Plan.ClientAlternate = &PlanClient{
		ID:     fmt.Sprintf("%s-%s-alt", clientIDPrefix, aliasSuffix),
		Secret: secretAlternate,
	}

	suite.Plan.ClientSecretPost = &PlanClient{
		ID:     fmt.Sprintf("%s-%s-post", clientIDPrefix, aliasSuffix),
		Secret: secretPost,
	}

	var (
		grantTypes    []string
		responseTypes []string
		responseModes []string
	)

	switch b.Name {
	case NameImplicit, NameImplicitFormPost:
		grantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit, oidc.GrantTypeRefreshToken}
		responseTypes = []string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowIDToken, oidc.ResponseTypeImplicitFlowToken, oidc.ResponseTypeImplicitFlowBoth}
	case NameHybrid, NameHybridFormPost:
		grantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit, oidc.GrantTypeRefreshToken}
		responseTypes = []string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeHybridFlowBoth}
	default:
		grantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken}
		responseTypes = []string{oidc.ResponseTypeAuthorizationCodeFlow}
	}

	switch b.Name {
	case NameBasicFormPost, NameHybridFormPost, NameImplicitFormPost:
		responseModes = []string{oidc.ResponseModeFormPost, oidc.ResponseModeFormPostJWT}
	default:
		responseModes = []string{oidc.ResponseModeQuery, oidc.ResponseModeQueryJWT}
	}

	suite.Clients = []schema.IdentityProvidersOpenIDConnectClient{
		{
			ID:                      suite.Plan.Client.ID,
			Name:                    description,
			Secret:                  MustHash(suite.Plan.Client.Secret),
			RedirectURIs:            []string{b.SuiteURL.JoinPath("test", "a", suite.Plan.Alias, "callback").String()},
			AuthorizationPolicy:     b.Policy,
			ConsentMode:             b.Consent,
			Public:                  false,
			Scopes:                  []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, oidc.ScopeProfile, oidc.ScopeEmail, oidc.ScopePhone, oidc.ScopeAddress, "all"},
			ResponseTypes:           responseTypes,
			GrantTypes:              grantTypes,
			ResponseModes:           responseModes,
			TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretBasic,
			RequestObjectSigningAlg: oidc.SigningAlgNone,
		},
		{
			ID:                      suite.Plan.ClientAlternate.ID,
			Name:                    fmt.Sprintf("%s (Alternate)", description),
			Secret:                  MustHash(suite.Plan.ClientAlternate.Secret),
			RedirectURIs:            []string{b.SuiteURL.JoinPath("test", "a", suite.Plan.Alias, "callback").String()},
			AuthorizationPolicy:     b.Policy,
			ConsentMode:             b.Consent,
			Public:                  false,
			Scopes:                  []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, oidc.ScopeProfile, oidc.ScopeEmail, oidc.ScopePhone, oidc.ScopeAddress, "all"},
			ResponseTypes:           responseTypes,
			GrantTypes:              grantTypes,
			ResponseModes:           responseModes,
			TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretBasic,
			RequestObjectSigningAlg: oidc.SigningAlgNone,
		},
		{
			ID:                      suite.Plan.ClientSecretPost.ID,
			Name:                    fmt.Sprintf("%s (Secret Post)", description),
			Secret:                  MustHash(suite.Plan.ClientSecretPost.Secret),
			RedirectURIs:            []string{b.SuiteURL.JoinPath("test", "a", suite.Plan.Alias, "callback").String()},
			AuthorizationPolicy:     b.Policy,
			ConsentMode:             b.Consent,
			Public:                  false,
			Scopes:                  []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, oidc.ScopeProfile, oidc.ScopeEmail, oidc.ScopePhone, oidc.ScopeAddress, "all"},
			ResponseTypes:           responseTypes,
			GrantTypes:              grantTypes,
			ResponseModes:           responseModes,
			TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
			RequestObjectSigningAlg: oidc.SigningAlgNone,
		},
	}

	return suite
}

// MustHash returns the digest of the given value and panics if it fails.
func MustHash(value string) *schema.PasswordDigest {
	hash, err := pbkdf2.New()
	if err != nil {
		panic(err)
	}

	var digest algorithm.Digest

	if digest, err = hash.Hash(value); err != nil {
		panic(err)
	}

	return schema.NewPasswordDigest(digest)
}

// Builders returns the conformance suite builders for every profile Authelia is certified for, in a fixed order.
func Builders(version, consent, policy, brand string, suiteURL, autheliaURL *url.URL) []*SuiteBuilder {
	return []*SuiteBuilder{
		{brand, NameConfig, "Config", true, version, consent, policy, nil, autheliaURL},
		{brand, NameBasic, "Basic", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, NameBasicFormPost, "Basic (Form Post)", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, NameHybrid, "Hybrid", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, NameHybridFormPost, "Hybrid (Form Post)", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, NameImplicit, "Implicit", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, NameImplicitFormPost, "Implicit (Form Post)", true, version, consent, policy, suiteURL, autheliaURL},
	}
}
