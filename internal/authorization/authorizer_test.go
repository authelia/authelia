package authorization

import (
	"net"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type AuthorizerSuite struct {
	suite.Suite
}

type AuthorizerTester struct {
	*Authorizer
}

func NewAuthorizerTester(config schema.AccessControl) *AuthorizerTester {
	fullConfig := &schema.Configuration{
		AccessControl: config,
	}

	return &AuthorizerTester{
		NewAuthorizer(fullConfig),
	}
}

func (s *AuthorizerTester) CheckAuthorizations(t *testing.T, subject Subject, requestURI, method string, expectedLevel Level) {
	targetURL, _ := url.ParseRequestURI(requestURI)

	object := NewObject(targetURL, method)

	_, level := s.GetRequiredLevel(subject, object)

	assert.Equal(t, expectedLevel, level)
}

func (s *AuthorizerTester) GetRuleMatchResults(subject Subject, requestURI, method string) (results []RuleMatchResult) {
	targetURL, _ := url.ParseRequestURI(requestURI)

	object := NewObject(targetURL, method)

	return s.Authorizer.GetRuleMatchResults(subject, object)
}

type AuthorizerTesterBuilder struct {
	config schema.AccessControl
}

func NewAuthorizerBuilder() *AuthorizerTesterBuilder {
	return &AuthorizerTesterBuilder{}
}

func (b *AuthorizerTesterBuilder) WithDefaultPolicy(policy string) *AuthorizerTesterBuilder {
	b.config.DefaultPolicy = policy
	return b
}

func (b *AuthorizerTesterBuilder) WithRule(rule schema.AccessControlRule) *AuthorizerTesterBuilder {
	b.config.Rules = append(b.config.Rules, rule)
	return b
}

func (b *AuthorizerTesterBuilder) Build() *AuthorizerTester {
	return NewAuthorizerTester(b.config)
}

var AnonymousUser = Subject{
	Username: "",
	Groups:   []string{},
	IP:       net.ParseIP("127.0.0.1"),
}

var OAuth2UserClientAClient = Subject{
	ClientID: "a_client",
	IP:       net.ParseIP("127.0.0.1"),
}

var UserWithGroups = Subject{
	Username: "john",
	Groups:   []string{"dev", "admins"},
	IP:       net.ParseIP("10.0.0.8"),
}

var John = UserWithGroups

var UserWithoutGroups = Subject{
	Username: "bob",
	Groups:   []string{},
	IP:       net.ParseIP("10.0.0.7"),
}

var Bob = UserWithoutGroups

var UserWithIPv6Address = Subject{
	Username: "sam",
	Groups:   []string{},
	IP:       net.ParseIP("fec0::1"),
}

var Sam = UserWithIPv6Address

var UserWithIPv6AddressAndGroups = Subject{
	Username: "sam",
	Groups:   []string{"dev", "admins"},
	IP:       net.ParseIP("fec0::2"),
}

var Sally = UserWithIPv6AddressAndGroups

func (s *AuthorizerSuite) TestShouldCheckDefaultBypassConfig() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(bypass).Build()

	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/elsewhere", fasthttp.MethodGet, Bypass)
}

func (s *AuthorizerSuite) TestShouldCheckDefaultDeniedConfig() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).Build()

	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/elsewhere", fasthttp.MethodGet, Denied)
}

func (s *AuthorizerSuite) TestShouldCheckMultiDomainRule() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains: []string{"*.example.com"},
			Policy:  bypass,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://private.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/elsewhere", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com.c/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.co/", fasthttp.MethodGet, Denied)
}

func (s *AuthorizerSuite) TestShouldCheckDynamicDomainRules() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains: []string{"{user}.example.com"},
			Policy:  oneFactor,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"{group}.example.com"},
			Policy:  oneFactor,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://john.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://dev.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://admins.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://othergroup.example.com/", fasthttp.MethodGet, Denied)
}

func (s *AuthorizerSuite) TestShouldCheckMultipleDomainRule() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains: []string{"*.example.com", "other.com"},
			Policy:  bypass,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://private.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/elsewhere", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com.c/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.co/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://other.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://other.com/elsewhere", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://private.other.com/", fasthttp.MethodGet, Denied)
}

func (s *AuthorizerSuite) TestShouldCheckFactorsPolicy() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains: []string{"single.example.com"},
			Policy:  oneFactor,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"protected.example.com"},
			Policy:  twoFactor,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"public.example.com"},
			Policy:  bypass,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://protected.example.com/", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://single.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", fasthttp.MethodGet, Denied)
}

func (s *AuthorizerSuite) TestShouldCheckQueryPolicy() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains: []string{"one.example.com"},
			Query: [][]schema.AccessControlRuleQuery{
				{
					{
						Operator: operatorEqual,
						Key:      "test",
						Value:    "two",
					},
					{
						Operator: operatorAbsent,
						Key:      "admin",
					},
				},
				{
					{
						Operator: operatorPresent,
						Key:      "public",
					},
				},
			},
			Policy: oneFactor,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"two.example.com"},
			Query: [][]schema.AccessControlRuleQuery{
				{
					{
						Operator: operatorEqual,
						Key:      "test",
						Value:    "one",
					},
				},
				{
					{
						Operator: operatorEqual,
						Key:      "test",
						Value:    "two",
					},
				},
			},
			Policy: twoFactor,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"three.example.com"},
			Query: [][]schema.AccessControlRuleQuery{
				{
					{
						Operator: operatorNotEqual,
						Key:      "test",
						Value:    "one",
					},
					{
						Operator: operatorNotEqual,
						Key:      "test",
						Value:    "two",
					},
				},
			},
			Policy: twoFactor,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"four.example.com"},
			Query: [][]schema.AccessControlRuleQuery{
				{
					{
						Operator: operatorPattern,
						Key:      "test",
						Value:    regexp.MustCompile(`^(one|two|three)$`),
					},
				},
			},
			Policy: twoFactor,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"five.example.com"},
			Query: [][]schema.AccessControlRuleQuery{
				{
					{
						Operator: operatorNotPattern,
						Key:      "test",
						Value:    regexp.MustCompile(`^(one|two|three)$`),
					},
				},
			},
			Policy: twoFactor,
		}).
		Build()

	testCases := []struct {
		name, requestURL string
		expected         Level
	}{
		{"ShouldDenyAbsentRule", "https://one.example.com/?admin=true", Denied},
		{"ShouldAllow1FAPresentRule", "https://one.example.com/?public=true", OneFactor},
		{"ShouldAllow1FAEqualRule", "https://one.example.com/?test=two", OneFactor},
		{"ShouldDenyAbsentRuleWithMatchingPresentRule", "https://one.example.com/?test=two&admin=true", Denied},
		{"ShouldAllow2FARuleWithOneMatchingEqual", "https://two.example.com/?test=one&admin=true", TwoFactor},
		{"ShouldAllow2FARuleWithAnotherMatchingEqual", "https://two.example.com/?test=two&admin=true", TwoFactor},
		{"ShouldDenyRuleWithNotMatchingEqual", "https://two.example.com/?test=three&admin=true", Denied},
		{"ShouldDenyRuleWithNotMatchingNotEqualAND1", "https://three.example.com/?test=one", Denied},
		{"ShouldDenyRuleWithNotMatchingNotEqualAND2", "https://three.example.com/?test=two", Denied},
		{"ShouldAllowRuleWithMatchingNotEqualAND", "https://three.example.com/?test=three", TwoFactor},
		{"ShouldAllowRuleWithMatchingPatternOne", "https://four.example.com/?test=one", TwoFactor},
		{"ShouldAllowRuleWithMatchingPatternTwo", "https://four.example.com/?test=two", TwoFactor},
		{"ShouldAllowRuleWithMatchingPatternThree", "https://four.example.com/?test=three", TwoFactor},
		{"ShouldDenyRuleWithNotMatchingPattern", "https://four.example.com/?test=five", Denied},
		{"ShouldAllowRuleWithMatchingNotPattern", "https://five.example.com/?test=five", TwoFactor},
		{"ShouldDenyRuleWithNotMatchingNotPatternOne", "https://five.example.com/?test=one", Denied},
		{"ShouldDenyRuleWithNotMatchingNotPatternTwo", "https://five.example.com/?test=two", Denied},
		{"ShouldDenyRuleWithNotMatchingNotPatternThree", "https://five.example.com/?test=three", Denied},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tester.CheckAuthorizations(t, UserWithGroups, tc.requestURL, fasthttp.MethodGet, tc.expected)
		})
	}
}

func (s *AuthorizerSuite) TestShouldCheckRulePrecedence() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains: []string{"public.example.com"},
			Policy:  bypass,
		}).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"user:john"}},
		}).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"oauth2:client:a_client"}},
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"*.example.com"},
			Policy:  twoFactor,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://public.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://public.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), OAuth2UserClientAClient, "https://public.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), OAuth2UserClientAClient, "https://protected.example.com/", fasthttp.MethodGet, OneFactor)
}

func (s *AuthorizerSuite) TestShouldCheckDomainMatching() {
	tester := NewAuthorizerBuilder().
		WithRule(schema.AccessControlRule{
			Domains: []string{"public.example.com"},
			Policy:  bypass,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"one-factor.example.com"},
			Policy:  oneFactor,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"two-factor.example.com"},
			Policy:  twoFactor,
		}).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"*.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"group:admins"}},
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"*.example.com"},
			Policy:  twoFactor,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://public.example.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://public.example.com:8080/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://public.example.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://public.example.com:8080", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com", fasthttp.MethodGet, Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://one-factor.example.com", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://one-factor.example.com", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one-factor.example.com", fasthttp.MethodGet, OneFactor)

	tester.CheckAuthorizations(s.T(), John, "https://two-factor.example.com", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://two-factor.example.com", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://two-factor.example.com", fasthttp.MethodGet, TwoFactor)

	tester.CheckAuthorizations(s.T(), John, "https://x.example.com", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://x.example.com", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://x.example.com", fasthttp.MethodGet, OneFactor)

	s.Require().Len(tester.rules, 5)

	s.Require().Len(tester.rules[0].Domains, 1)

	s.Assert().Equal("public.example.com", tester.config.AccessControl.Rules[0].Domains[0])

	ruleMatcher0, ok := tester.rules[0].Domains[0].Matcher.(*AccessControlDomainMatcher)
	s.Require().True(ok)
	s.Assert().Equal("public.example.com", ruleMatcher0.Name)
	s.Assert().False(ruleMatcher0.Wildcard)
	s.Assert().False(ruleMatcher0.UserWildcard)
	s.Assert().False(ruleMatcher0.GroupWildcard)

	s.Require().Len(tester.rules[1].Domains, 1)

	s.Assert().Equal("one-factor.example.com", tester.config.AccessControl.Rules[1].Domains[0])

	ruleMatcher1, ok := tester.rules[1].Domains[0].Matcher.(*AccessControlDomainMatcher)
	s.Require().True(ok)
	s.Assert().Equal("one-factor.example.com", ruleMatcher1.Name)
	s.Assert().False(ruleMatcher1.Wildcard)
	s.Assert().False(ruleMatcher1.UserWildcard)
	s.Assert().False(ruleMatcher1.GroupWildcard)

	s.Require().Len(tester.rules[2].Domains, 1)

	s.Assert().Equal("two-factor.example.com", tester.config.AccessControl.Rules[2].Domains[0])

	ruleMatcher2, ok := tester.rules[2].Domains[0].Matcher.(*AccessControlDomainMatcher)
	s.Require().True(ok)
	s.Assert().Equal("two-factor.example.com", ruleMatcher2.Name)
	s.Assert().False(ruleMatcher2.Wildcard)
	s.Assert().False(ruleMatcher2.UserWildcard)
	s.Assert().False(ruleMatcher2.GroupWildcard)

	s.Require().Len(tester.rules[3].Domains, 1)

	s.Assert().Equal("*.example.com", tester.config.AccessControl.Rules[3].Domains[0])

	ruleMatcher3, ok := tester.rules[3].Domains[0].Matcher.(*AccessControlDomainMatcher)
	s.Require().True(ok)
	s.Assert().Equal(".example.com", ruleMatcher3.Name)
	s.Assert().True(ruleMatcher3.Wildcard)
	s.Assert().False(ruleMatcher3.UserWildcard)
	s.Assert().False(ruleMatcher3.GroupWildcard)

	s.Require().Len(tester.rules[4].Domains, 1)

	s.Assert().Equal("*.example.com", tester.config.AccessControl.Rules[4].Domains[0])

	ruleMatcher4, ok := tester.rules[4].Domains[0].Matcher.(*AccessControlDomainMatcher)
	s.Require().True(ok)
	s.Assert().Equal(".example.com", ruleMatcher4.Name)
	s.Assert().True(ruleMatcher4.Wildcard)
	s.Assert().False(ruleMatcher4.UserWildcard)
	s.Assert().False(ruleMatcher4.GroupWildcard)
}

func (s *AuthorizerSuite) TestShouldCheckDomainRegexMatching() {
	createSliceRegexRule := func(t *testing.T, rules []string) []regexp.Regexp {
		result, err := stringSliceToRegexpSlice(rules)

		require.NoError(t, err)

		return result
	}

	tester := NewAuthorizerBuilder().
		WithRule(schema.AccessControlRule{
			DomainsRegex: createSliceRegexRule(s.T(), []string{`^.*\.example.com$`}),
			Policy:       bypass,
		}).
		WithRule(schema.AccessControlRule{
			DomainsRegex: createSliceRegexRule(s.T(), []string{`^.*\.example2.com$`}),
			Policy:       oneFactor,
		}).
		WithRule(schema.AccessControlRule{
			DomainsRegex: createSliceRegexRule(s.T(), []string{`^(?P<User>[a-zA-Z0-9]+)\.regex.com$`}),
			Policy:       oneFactor,
		}).
		WithRule(schema.AccessControlRule{
			DomainsRegex: createSliceRegexRule(s.T(), []string{`^group-(?P<Group>[a-zA-Z0-9]+)\.regex.com$`}),
			Policy:       twoFactor,
		}).
		WithRule(schema.AccessControlRule{
			DomainsRegex: createSliceRegexRule(s.T(), []string{`^.*\.(one|two).com$`}),
			Policy:       twoFactor,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://john.regex.com", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://john.regex.com", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), Bob, "https://public.example.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example2.com", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://group-dev.regex.com", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://group-dev.regex.com", fasthttp.MethodGet, Denied)

	s.Require().Len(tester.rules, 5)

	s.Require().Len(tester.rules[0].Domains, 1)

	s.Assert().Equal("^.*\\.example.com$", tester.config.AccessControl.Rules[0].DomainsRegex[0].String())

	ruleMatcher0, ok := tester.rules[0].Domains[0].Matcher.(RegexpStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^.*\\.example.com$", ruleMatcher0.String())

	s.Require().Len(tester.rules[1].Domains, 1)

	s.Assert().Equal("^.*\\.example2.com$", tester.config.AccessControl.Rules[1].DomainsRegex[0].String())

	ruleMatcher1, ok := tester.rules[1].Domains[0].Matcher.(RegexpStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^.*\\.example2.com$", ruleMatcher1.String())

	s.Require().Len(tester.rules[2].Domains, 1)

	s.Assert().Equal("^(?P<User>[a-zA-Z0-9]+)\\.regex.com$", tester.config.AccessControl.Rules[2].DomainsRegex[0].String())

	ruleMatcher2, ok := tester.rules[2].Domains[0].Matcher.(RegexpGroupStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^(?P<User>[a-zA-Z0-9]+)\\.regex.com$", ruleMatcher2.String())

	s.Require().Len(tester.rules[3].Domains, 1)

	s.Assert().Equal("^group-(?P<Group>[a-zA-Z0-9]+)\\.regex.com$", tester.config.AccessControl.Rules[3].DomainsRegex[0].String())

	ruleMatcher3, ok := tester.rules[3].Domains[0].Matcher.(RegexpGroupStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^group-(?P<Group>[a-zA-Z0-9]+)\\.regex.com$", ruleMatcher3.String())

	s.Require().Len(tester.rules[4].Domains, 1)

	s.Assert().Equal("^.*\\.(one|two).com$", tester.config.AccessControl.Rules[4].DomainsRegex[0].String())

	ruleMatcher4, ok := tester.rules[4].Domains[0].Matcher.(RegexpStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^.*\\.(one|two).com$", ruleMatcher4.String())
}

func (s *AuthorizerSuite) TestShouldCheckResourceSubjectMatching() {
	createSliceRegexRule := func(t *testing.T, rules []string) []regexp.Regexp {
		result, err := stringSliceToRegexpSlice(rules)

		require.NoError(t, err)

		return result
	}

	tester := NewAuthorizerBuilder().
		WithRule(schema.AccessControlRule{
			Domains:   []string{"id.example.com"},
			Policy:    oneFactor,
			Resources: createSliceRegexRule(s.T(), []string{`^/(?P<User>[a-zA-Z0-9]+)/personal(/|/.*)?$`, `^/(?P<Group>[a-zA-Z0-9]+)/group(/|/.*)?$`}),
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"id.example.com"},
			Policy:    deny,
			Resources: createSliceRegexRule(s.T(), []string{`^/([a-zA-Z0-9]+)/personal(/|/.*)?$`, `^/([a-zA-Z0-9]+)/group(/|/.*)?$`}),
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"id.example.com"},
			Policy:  bypass,
		}).
		Build()

	// Accessing the unprotected root.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com", fasthttp.MethodGet, Bypass)

	// Accessing Personal page.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/john/personal", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/John/personal", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/bob/personal", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/Bob/personal", fasthttp.MethodGet, OneFactor)

	// Accessing an invalid users Personal page.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/invaliduser/personal", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/invaliduser/personal", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/invaliduser/personal", fasthttp.MethodGet, OneFactor)

	// Accessing another users Personal page.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/bob/personal", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/bob/personal", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/Bob/personal", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/Bob/personal", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/john/personal", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/john/personal", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/John/personal", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/John/personal", fasthttp.MethodGet, OneFactor)

	// Accessing a Group page.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/dev/group", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/admins/group", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/dev/group", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/admins/group", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/dev/group", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/admins/group", fasthttp.MethodGet, OneFactor)

	// Accessing an invalid group's Group page.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/invalidgroup/group", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/invalidgroup/group", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/invalidgroup/group", fasthttp.MethodGet, OneFactor)

	s.Require().Len(tester.rules, 3)

	s.Require().Len(tester.rules[0].Resources, 2)

	ruleMatcher00, ok := tester.rules[0].Resources[0].Matcher.(RegexpGroupStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^/(?P<User>[a-zA-Z0-9]+)/personal(/|/.*)?$", ruleMatcher00.String())

	ruleMatcher01, ok := tester.rules[0].Resources[1].Matcher.(RegexpGroupStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^/(?P<Group>[a-zA-Z0-9]+)/group(/|/.*)?$", ruleMatcher01.String())

	s.Require().Len(tester.rules[1].Resources, 2)

	ruleMatcher10, ok := tester.rules[1].Resources[0].Matcher.(RegexpStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^/([a-zA-Z0-9]+)/personal(/|/.*)?$", ruleMatcher10.String())

	ruleMatcher11, ok := tester.rules[1].Resources[1].Matcher.(RegexpStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^/([a-zA-Z0-9]+)/group(/|/.*)?$", ruleMatcher11.String())
}

func (s *AuthorizerSuite) TestShouldCheckUserMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"user:john"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", fasthttp.MethodGet, Denied)
}

func (s *AuthorizerSuite) TestShouldCheckGroupMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"group:admins"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", fasthttp.MethodGet, Denied)
}

func (s *AuthorizerSuite) TestShouldCheckSubjectsMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"group:admins"}, {"user:bob"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Sam, "https://protected.example.com/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", fasthttp.MethodGet, OneFactor)
}

func (s *AuthorizerSuite) TestShouldCheckMultipleSubjectsMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"group:admins", "user:bob"}, {"group:admins", "group:dev"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", fasthttp.MethodGet, OneFactor)
}

func (s *AuthorizerSuite) TestShouldCheckIPMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"protected.example.com"},
			Policy:   bypass,
			Networks: []string{"192.168.1.8", "10.0.0.8"},
		}).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Networks: []string{"10.0.0.7"},
		}).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"net.example.com"},
			Policy:   twoFactor,
			Networks: []string{"10.0.0.0/8"},
		}).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"ipv6.example.com"},
			Policy:   twoFactor,
			Networks: []string{"fec0::1/64"},
		}).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"ipv6-alt.example.com"},
			Policy:   twoFactor,
			Networks: []string{"fec0::1"},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", fasthttp.MethodGet, Denied)

	tester.CheckAuthorizations(s.T(), John, "https://net.example.com/", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://net.example.com/", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://net.example.com/", fasthttp.MethodGet, Denied)

	tester.CheckAuthorizations(s.T(), Sally, "https://ipv6-alt.example.com/", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), Sam, "https://ipv6-alt.example.com/", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), Sally, "https://ipv6.example.com/", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), Sam, "https://ipv6.example.com/", fasthttp.MethodGet, TwoFactor)
}

func (s *AuthorizerSuite) TestShouldCheckMethodMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains: []string{"protected.example.com"},
			Policy:  bypass,
			Methods: []string{fasthttp.MethodOptions, fasthttp.MethodHead, fasthttp.MethodGet, fasthttp.MethodConnect, fasthttp.MethodTrace},
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"protected.example.com"},
			Policy:  oneFactor,
			Methods: []string{fasthttp.MethodPut, fasthttp.MethodPatch, fasthttp.MethodPost},
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"protected.example.com"},
			Policy:  twoFactor,
			Methods: []string{fasthttp.MethodDelete},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", fasthttp.MethodOptions, Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", fasthttp.MethodHead, Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", fasthttp.MethodConnect, Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", fasthttp.MethodTrace, Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", fasthttp.MethodPut, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", fasthttp.MethodPatch, OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", fasthttp.MethodPost, OneFactor)

	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", fasthttp.MethodDelete, TwoFactor)
}

func (s *AuthorizerSuite) TestShouldCheckResourceMatching() {
	createSliceRegexRule := func(t *testing.T, rules []string) []regexp.Regexp {
		result, err := stringSliceToRegexpSlice(rules)

		require.NoError(t, err)

		return result
	}

	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"resource.example.com"},
			Policy:    bypass,
			Resources: createSliceRegexRule(s.T(), []string{"^/case/[a-z]+$", "^/$"}),
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"resource.example.com"},
			Policy:    bypass,
			Resources: createSliceRegexRule(s.T(), []string{"^/bypass/.*$", "^/$", "embedded"}),
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"resource.example.com"},
			Policy:    oneFactor,
			Resources: createSliceRegexRule(s.T(), []string{"^/one_factor/.*$"}),
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"resource.example.com"},
			Policy:    twoFactor,
			Resources: createSliceRegexRule(s.T(), []string{"^/a/longer/rule/.*$"}),
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"resource.example.com"},
			Policy:    twoFactor,
			Resources: createSliceRegexRule(s.T(), []string{"^/an/exact/path/$"}),
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/abc", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/one_factor/abc", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/xyz/embedded/abc", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/case/abc", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/case/ABC", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/an/exact/path/", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/../a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/..//a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/..%2f/a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/..%2fa/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/..%2F/a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/..%2Fa/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2e%2e/a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2e%2e//a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2e%2e%2f/a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2e%2e%2fa/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2e%2e%2F/a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2e%2e%2Fa/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2E%2E/a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2E%2E//a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2E%2E%2f/a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2E%2E%2fa/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2E%2E%2F/a/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2E%2E%2Fa/longer/rule/abc", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/%2E%2E%2Fan/exact/path/", fasthttp.MethodGet, TwoFactor)
}

// This test assures that rules without domains (not allowed by schema validator at this time) will pass validation correctly.
func (s *AuthorizerSuite) TestShouldMatchAnyDomainIfBlank() {
	tester := NewAuthorizerBuilder().
		WithRule(schema.AccessControlRule{
			Policy:  bypass,
			Methods: []string{fasthttp.MethodOptions, fasthttp.MethodHead, fasthttp.MethodGet, fasthttp.MethodConnect, fasthttp.MethodTrace},
		}).
		WithRule(schema.AccessControlRule{
			Policy:  oneFactor,
			Methods: []string{fasthttp.MethodPut, fasthttp.MethodPatch},
		}).
		WithRule(schema.AccessControlRule{
			Policy:  twoFactor,
			Methods: []string{fasthttp.MethodDelete},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://one.domain-four.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-three.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-two.com", fasthttp.MethodOptions, Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://one.domain-four.com", fasthttp.MethodPut, OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-three.com", fasthttp.MethodPatch, OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-two.com", fasthttp.MethodPut, OneFactor)

	tester.CheckAuthorizations(s.T(), John, "https://one.domain-four.com", fasthttp.MethodDelete, TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-three.com", fasthttp.MethodDelete, TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-two.com", fasthttp.MethodDelete, TwoFactor)

	tester.CheckAuthorizations(s.T(), John, "https://one.domain-four.com", fasthttp.MethodPost, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-three.com", fasthttp.MethodPost, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-two.com", fasthttp.MethodPost, Denied)
}

func (s *AuthorizerSuite) TestShouldMatchResourceWithSubjectRules() {
	createSliceRegexRule := func(t *testing.T, rules []string) []regexp.Regexp {
		result, err := stringSliceToRegexpSlice(rules)

		require.NoError(t, err)

		return result
	}

	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"public.example.com"},
			Resources: createSliceRegexRule(s.T(), []string{"^/admin/.*$"}),
			Subjects:  [][]string{{"group:admins"}},
			Policy:    oneFactor,
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"public.example.com"},
			Resources: createSliceRegexRule(s.T(), []string{"^/admin/.*$"}),
			Policy:    deny,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"public.example.com"},
			Policy:  bypass,
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"public2.example.com"},
			Resources: createSliceRegexRule(s.T(), []string{"^/admin/.*$"}),
			Subjects:  [][]string{{"group:admins"}},
			Policy:    bypass,
		}).
		WithRule(schema.AccessControlRule{
			Domains:   []string{"public2.example.com"},
			Resources: createSliceRegexRule(s.T(), []string{"^/admin/.*$"}),
			Policy:    deny,
		}).
		WithRule(schema.AccessControlRule{
			Domains: []string{"public2.example.com"},
			Policy:  bypass,
		}).
		WithRule(schema.AccessControlRule{
			Domains:  []string{"private.example.com"},
			Subjects: [][]string{{"group:admins"}},
			Policy:   twoFactor,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://public.example.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://public.example.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com", fasthttp.MethodGet, Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://public.example.com/admin/index.html", fasthttp.MethodGet, OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://public.example.com/admin/index.html", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/admin/index.html", fasthttp.MethodGet, OneFactor)

	tester.CheckAuthorizations(s.T(), John, "https://public2.example.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://public2.example.com", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public2.example.com", fasthttp.MethodGet, Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://public2.example.com/admin/index.html", fasthttp.MethodGet, Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://public2.example.com/admin/index.html", fasthttp.MethodGet, Denied)

	// This test returns this result since we validate the schema instead of validating it in code.
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public2.example.com/admin/index.html", fasthttp.MethodGet, Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://private.example.com", fasthttp.MethodGet, TwoFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://private.example.com", fasthttp.MethodGet, Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://private.example.com", fasthttp.MethodGet, TwoFactor)

	results := tester.GetRuleMatchResults(John, "https://private.example.com", fasthttp.MethodGet)

	s.Require().Len(results, 7)

	s.Assert().False(results[0].IsMatch())
	s.Assert().False(results[0].MatchDomain)
	s.Assert().False(results[0].MatchResources)
	s.Assert().True(results[0].MatchSubjects)
	s.Assert().True(results[0].MatchNetworks)
	s.Assert().True(results[0].MatchMethods)

	s.Assert().False(results[1].IsMatch())
	s.Assert().False(results[1].MatchDomain)
	s.Assert().False(results[1].MatchResources)
	s.Assert().True(results[1].MatchSubjects)
	s.Assert().True(results[1].MatchNetworks)
	s.Assert().True(results[1].MatchMethods)

	s.Assert().False(results[2].IsMatch())
	s.Assert().False(results[2].MatchDomain)
	s.Assert().True(results[2].MatchResources)
	s.Assert().True(results[2].MatchSubjects)
	s.Assert().True(results[2].MatchNetworks)
	s.Assert().True(results[2].MatchMethods)

	s.Assert().False(results[3].IsMatch())
	s.Assert().False(results[3].MatchDomain)
	s.Assert().False(results[3].MatchResources)
	s.Assert().True(results[3].MatchSubjects)
	s.Assert().True(results[3].MatchNetworks)
	s.Assert().True(results[3].MatchMethods)

	s.Assert().False(results[4].IsMatch())
	s.Assert().False(results[4].MatchDomain)
	s.Assert().False(results[4].MatchResources)
	s.Assert().True(results[4].MatchSubjects)
	s.Assert().True(results[4].MatchNetworks)
	s.Assert().True(results[4].MatchMethods)

	s.Assert().False(results[5].IsMatch())
	s.Assert().False(results[5].MatchDomain)
	s.Assert().True(results[5].MatchResources)
	s.Assert().True(results[5].MatchSubjects)
	s.Assert().True(results[5].MatchNetworks)
	s.Assert().True(results[5].MatchMethods)

	s.Assert().True(results[6].IsMatch())
	s.Assert().True(results[6].MatchDomain)
	s.Assert().True(results[6].MatchResources)
	s.Assert().True(results[6].MatchSubjects)
	s.Assert().True(results[6].MatchNetworks)
	s.Assert().True(results[6].MatchMethods)
}

func (s *AuthorizerSuite) TestPolicyToLevel() {
	s.Assert().Equal(Bypass, NewLevel(bypass))
	s.Assert().Equal(OneFactor, NewLevel(oneFactor))
	s.Assert().Equal(TwoFactor, NewLevel(twoFactor))
	s.Assert().Equal(Denied, NewLevel(deny))

	s.Assert().Equal(Denied, NewLevel("whatever"))
}

func TestRunSuite(t *testing.T) {
	s := AuthorizerSuite{}
	suite.Run(t, &s)
}

func TestNewAuthorizer(t *testing.T) {
	config := &schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: deny,
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"example.com"},
					Policy:  twoFactor,
					Subjects: [][]string{
						{
							"user:admin",
						},
						{
							"group:admins",
						},
					},
				},
			},
		},
	}

	authorizer := NewAuthorizer(config)

	assert.Equal(t, Denied, authorizer.defaultPolicy)
	assert.Equal(t, TwoFactor, authorizer.rules[0].Policy)

	user, ok := authorizer.rules[0].Subjects[0].Subjects[0].(AccessControlUser)
	require.True(t, ok)
	assert.Equal(t, "admin", user.Name)

	group, ok := authorizer.rules[0].Subjects[1].Subjects[0].(AccessControlGroup)
	require.True(t, ok)
	assert.Equal(t, "admins", group.Name)
}

func TestAuthorizerIsSecondFactorEnabledRuleWithNoOIDC(t *testing.T) {
	config := &schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: deny,
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"example.com"},
					Policy:  oneFactor,
				},
			},
		},
	}

	authorizer := NewAuthorizer(config)
	assert.False(t, authorizer.IsSecondFactorEnabled())

	config.AccessControl.Rules[0].Policy = twoFactor
	authorizer = NewAuthorizer(config)
	assert.True(t, authorizer.IsSecondFactorEnabled())
}

func TestAuthorizerIsSecondFactorEnabledRuleWithOIDC(t *testing.T) {
	config := &schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: deny,
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"example.com"},
					Policy:  oneFactor,
				},
			},
		},
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						AuthorizationPolicy: oneFactor,
					},
				},
			},
		},
	}

	authorizer := NewAuthorizer(config)
	assert.False(t, authorizer.IsSecondFactorEnabled())

	config.AccessControl.Rules[0].Policy = twoFactor
	authorizer = NewAuthorizer(config)
	assert.True(t, authorizer.IsSecondFactorEnabled())

	config.AccessControl.Rules[0].Policy = oneFactor
	authorizer = NewAuthorizer(config)
	assert.False(t, authorizer.IsSecondFactorEnabled())

	config.IdentityProviders.OIDC.Clients[0].AuthorizationPolicy = twoFactor
	authorizer = NewAuthorizer(config)
	assert.True(t, authorizer.IsSecondFactorEnabled())

	config.AccessControl.Rules[0].Policy = oneFactor
	config.IdentityProviders.OIDC.Clients[0].AuthorizationPolicy = oneFactor
	authorizer = NewAuthorizer(config)
	assert.False(t, authorizer.IsSecondFactorEnabled())

	config.AccessControl.DefaultPolicy = twoFactor
	authorizer = NewAuthorizer(config)
	assert.True(t, authorizer.IsSecondFactorEnabled())
}
