package main

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

type OpenIDConnectConformanceSuiteBuilder struct {
	name          string
	friendly      string
	certification bool
	version       string
	suiteURL      *url.URL
	autheliaURL   *url.URL
}

func (p *OpenIDConnectConformanceSuiteBuilder) Build() OpenIDConnectConformanceSuite {
	var (
		apiname, namePrefix, clientIDPrefix, descriptionSuffix string
		variant                                                *OpenIDConnectConformanceSuitePlanVariant
	)

	if p.certification {
		namePrefix = "conformance-"
		clientIDPrefix = "conformance-certification"
		descriptionSuffix = "Certification Profile"
	} else {
		clientIDPrefix = "conformance-test"
		descriptionSuffix = "Test Profile"
	}

	aliasSuffix := fmt.Sprintf("%s-%s", strings.ReplaceAll(strings.ToLower(p.name), ".", "-"), strings.ReplaceAll(strings.ToLower(p.version), ".", ""))

	name := fmt.Sprintf("%s%s", namePrefix, p.name)

	switch name {
	case suiteConformanceBasic, "conformance-basic-form-post", "conformance-hybrid", "conformance-hybrid-form-post", "conformance-implicit", "conformance-implicit-form-post":
		variant = &OpenIDConnectConformanceSuitePlanVariant{
			ServerMetadata:     "discovery",
			ClientRegistration: "static_client",
		}
	}

	switch name {
	case "conformance-config":
		apiname = "oidcc-config-certification-test-plan"
	case suiteConformanceBasic:
		apiname = "oidcc-basic-certification-test-plan"
	case "conformance-basic-form-post":
		apiname = "oidcc-formpost-basic-certification-test-plan"
	case "conformance-hybrid":
		apiname = "oidcc-hybrid-certification-test-plan"
	case "conformance-hybrid-form-post":
		apiname = "oidcc-formpost-hybrid-certification-test-plan"
	case "conformance-implicit":
		apiname = "oidcc-implicit-certification-test-plan"
	case "conformance-implicit-form-post":
		apiname = "oidcc-formpost-implicit-certification-test-plan"
	}

	suite := OpenIDConnectConformanceSuite{
		Name: name,
		Plan: OpenIDConnectConformanceSuitePlan{
			Name:        apiname,
			Variant:     variant,
			Alias:       fmt.Sprintf("%s%s", namePrefix, aliasSuffix),
			Publish:     "summary",
			Description: fmt.Sprintf("Authelia %s %s %s", p.version, p.friendly, descriptionSuffix),
			Server: OpenIDConnectConformanceSuitePlanServer{
				DiscoveryURL: p.autheliaURL.JoinPath(".well-known/openid-configuration").String(),
			},
		},
	}

	if p.suiteURL == nil {
		return suite
	}

	r := &random.Cryptographical{}

	secret := r.StringCustom(80, random.CharSetAlphaNumeric)
	secretAlternate := r.StringCustom(80, random.CharSetAlphaNumeric)
	secretPost := r.StringCustom(80, random.CharSetAlphaNumeric)

	suite.Plan.Client = &OpenIDConnectConformanceSuitePlanClient{
		ID:     fmt.Sprintf("%s-%s", clientIDPrefix, aliasSuffix),
		Secret: secret,
	}

	suite.Plan.ClientAlternate = &OpenIDConnectConformanceSuitePlanClient{
		ID:     fmt.Sprintf("%s-%s-alt", clientIDPrefix, aliasSuffix),
		Secret: secretAlternate,
	}

	suite.Plan.ClientSecretPost = &OpenIDConnectConformanceSuitePlanClient{
		ID:     fmt.Sprintf("%s-%s-post", clientIDPrefix, aliasSuffix),
		Secret: secretPost,
	}

	var (
		grantTypes    []string
		responseTypes []string
		responseModes []string
	)

	switch p.name {
	case "implicit", suiteNameImplicitFormPost:
		grantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit, oidc.GrantTypeRefreshToken}
		responseTypes = []string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowIDToken, oidc.ResponseTypeImplicitFlowToken, oidc.ResponseTypeImplicitFlowBoth}
	case "hybrid", suiteNameHybridFormPost:
		grantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit, oidc.GrantTypeRefreshToken}
		responseTypes = []string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeHybridFlowBoth}
	default:
		grantTypes = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken}
		responseTypes = []string{oidc.ResponseTypeAuthorizationCodeFlow}
	}

	switch p.name {
	case suiteNameHybridFormPost, suiteNameImplicitFormPost:
		responseModes = []string{oidc.ResponseModeFormPost, oidc.ResponseModeFormPostJWT}
	default:
		responseModes = []string{oidc.ResponseModeQuery, oidc.ResponseModeQueryJWT}
	}

	suite.Clients = []schema.IdentityProvidersOpenIDConnectClient{
		{
			ID:                      suite.Plan.Client.ID,
			Secret:                  MustHash(suite.Plan.Client.Secret),
			RedirectURIs:            []string{p.suiteURL.JoinPath("test", "a", suite.Plan.Alias, "callback").String()},
			AuthorizationPolicy:     "one_factor",
			ConsentMode:             oidc.ClientConsentModeImplicit.String(),
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
			Secret:                  MustHash(suite.Plan.ClientAlternate.Secret),
			RedirectURIs:            []string{p.suiteURL.JoinPath("test", "a", suite.Plan.Alias, "callback").String()},
			AuthorizationPolicy:     "one_factor",
			ConsentMode:             oidc.ClientConsentModeImplicit.String(),
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
			Secret:                  MustHash(suite.Plan.ClientSecretPost.Secret),
			RedirectURIs:            []string{p.suiteURL.JoinPath("test", "a", suite.Plan.Alias, "callback").String()},
			AuthorizationPolicy:     "one_factor",
			ConsentMode:             oidc.ClientConsentModeImplicit.String(),
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
