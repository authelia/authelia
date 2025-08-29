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
	brand         string
	name          string
	friendly      string
	certification bool
	version       string
	consent       string
	policy        string
	suiteURL      *url.URL
	autheliaURL   *url.URL
}

func (b *OpenIDConnectConformanceSuiteBuilder) Build() OpenIDConnectConformanceSuite {
	var (
		apiname, namePrefix, clientIDPrefix, descriptionSuffix string
		variant                                                *OpenIDConnectConformanceSuitePlanVariant
	)

	if b.certification {
		namePrefix = "conformance-"
		clientIDPrefix = "conformance-certification"
		descriptionSuffix = "Certification Profile"
	} else {
		clientIDPrefix = "conformance-test"
		descriptionSuffix = "Test Profile"
	}

	aliasSuffix := fmt.Sprintf("%s-%s", strings.ReplaceAll(strings.ToLower(b.name), ".", "-"), b.brand+strings.ReplaceAll(strings.ToLower(b.version), ".", ""))

	name := fmt.Sprintf("%s%s", namePrefix, b.name)
	description := fmt.Sprintf("Authelia %s %s %s", b.version, b.friendly, descriptionSuffix)

	switch name {
	case suiteConformanceBasic, suiteConformanceBasicFormPost, suiteConformanceHybrid, suiteConformanceHybridFormPost, suiteConformanceImplicit, suiteConformanceImplicitFormPost:
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
	case suiteConformanceBasicFormPost:
		apiname = "oidcc-formpost-basic-certification-test-plan"
	case suiteConformanceHybrid:
		apiname = "oidcc-hybrid-certification-test-plan"
	case suiteConformanceHybridFormPost:
		apiname = "oidcc-formpost-hybrid-certification-test-plan"
	case suiteConformanceImplicit:
		apiname = "oidcc-implicit-certification-test-plan"
	case suiteConformanceImplicitFormPost:
		apiname = "oidcc-formpost-implicit-certification-test-plan"
	}

	suite := OpenIDConnectConformanceSuite{
		Name: name,
		Plan: OpenIDConnectConformanceSuitePlan{
			Name:        apiname,
			Variant:     variant,
			Alias:       fmt.Sprintf("%s%s", namePrefix, aliasSuffix),
			Publish:     "summary",
			Description: description,
			Server: OpenIDConnectConformanceSuitePlanServer{
				DiscoveryURL: b.autheliaURL.JoinPath(".well-known/openid-configuration").String(),
			},
		},
	}

	if b.suiteURL == nil {
		return suite
	}

	r := random.New()

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

	switch b.name {
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

	switch b.name {
	case suiteNameBasicFormPost, suiteNameHybridFormPost, suiteNameImplicitFormPost:
		responseModes = []string{oidc.ResponseModeFormPost, oidc.ResponseModeFormPostJWT}
	default:
		responseModes = []string{oidc.ResponseModeQuery, oidc.ResponseModeQueryJWT}
	}

	suite.Clients = []schema.IdentityProvidersOpenIDConnectClient{
		{
			ID:                      suite.Plan.Client.ID,
			Name:                    description,
			Secret:                  MustHash(suite.Plan.Client.Secret),
			RedirectURIs:            []string{b.suiteURL.JoinPath("test", "a", suite.Plan.Alias, "callback").String()},
			AuthorizationPolicy:     b.policy,
			ConsentMode:             b.consent,
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
			RedirectURIs:            []string{b.suiteURL.JoinPath("test", "a", suite.Plan.Alias, "callback").String()},
			AuthorizationPolicy:     b.policy,
			ConsentMode:             b.consent,
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
			RedirectURIs:            []string{b.suiteURL.JoinPath("test", "a", suite.Plan.Alias, "callback").String()},
			AuthorizationPolicy:     b.policy,
			ConsentMode:             b.consent,
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
