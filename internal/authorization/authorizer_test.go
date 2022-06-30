package authorization

import (
	"net"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type AuthorizerSuite struct {
	suite.Suite
}

type AuthorizerTester struct {
	*Authorizer
}

func NewAuthorizerTester(config schema.AccessControlConfiguration) *AuthorizerTester {
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

	level := s.GetRequiredLevel(subject, object)

	assert.Equal(t, expectedLevel, level)
}

func (s *AuthorizerTester) GetRuleMatchResults(subject Subject, requestURI, method string) (results []RuleMatchResult) {
	targetURL, _ := url.ParseRequestURI(requestURI)

	object := NewObject(targetURL, method)

	return s.Authorizer.GetRuleMatchResults(subject, object)
}

type AuthorizerTesterBuilder struct {
	config schema.AccessControlConfiguration
}

func NewAuthorizerBuilder() *AuthorizerTesterBuilder {
	return &AuthorizerTesterBuilder{}
}

func (b *AuthorizerTesterBuilder) WithDefaultPolicy(policy string) *AuthorizerTesterBuilder {
	b.config.DefaultPolicy = policy
	return b
}

func (b *AuthorizerTesterBuilder) WithRule(rule schema.ACLRule) *AuthorizerTesterBuilder {
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

	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/elsewhere", "GET", Bypass)
}

func (s *AuthorizerSuite) TestShouldCheckDefaultDeniedConfig() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).Build()

	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/elsewhere", "GET", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckMultiDomainRule() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains: []string{"*.example.com"},
			Policy:  bypass,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://private.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/elsewhere", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com.c/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.co/", "GET", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckDynamicDomainRules() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains: []string{"{user}.example.com"},
			Policy:  bypass,
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"{group}.example.com"},
			Policy:  bypass,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://john.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://dev.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://admins.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://othergroup.example.com/", "GET", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckMultipleDomainRule() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains: []string{"*.example.com", "other.com"},
			Policy:  bypass,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://private.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/elsewhere", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com.c/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.co/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://other.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://other.com/elsewhere", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://private.other.com/", "GET", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckFactorsPolicy() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains: []string{"single.example.com"},
			Policy:  oneFactor,
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"protected.example.com"},
			Policy:  twoFactor,
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"public.example.com"},
			Policy:  bypass,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://protected.example.com/", "GET", TwoFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://single.example.com/", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", "GET", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckRulePrecedence() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   bypass,
			Subjects: [][]string{{"user:john"}},
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"protected.example.com"},
			Policy:  oneFactor,
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"*.example.com"},
			Policy:  twoFactor,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://public.example.com/", "GET", TwoFactor)
}

func (s *AuthorizerSuite) TestShouldCheckDomainMatching() {
	tester := NewAuthorizerBuilder().
		WithRule(schema.ACLRule{
			Domains: []string{"public.example.com"},
			Policy:  bypass,
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"one-factor.example.com"},
			Policy:  oneFactor,
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"two-factor.example.com"},
			Policy:  twoFactor,
		}).
		WithRule(schema.ACLRule{
			Domains:  []string{"*.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"group:admins"}},
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"*.example.com"},
			Policy:  twoFactor,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://public.example.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://public.example.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com", "GET", Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://one-factor.example.com", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://one-factor.example.com", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one-factor.example.com", "GET", OneFactor)

	tester.CheckAuthorizations(s.T(), John, "https://two-factor.example.com", "GET", TwoFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://two-factor.example.com", "GET", TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://two-factor.example.com", "GET", TwoFactor)

	tester.CheckAuthorizations(s.T(), John, "https://x.example.com", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://x.example.com", "GET", TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://x.example.com", "GET", OneFactor)

	s.Require().Len(tester.rules, 5)

	s.Require().Len(tester.rules[0].Domains, 1)

	s.Assert().Equal("public.example.com", tester.configuration.AccessControl.Rules[0].Domains[0])

	ruleMatcher0, ok := tester.rules[0].Domains[0].Matcher.(*AccessControlDomainMatcher)
	s.Require().True(ok)
	s.Assert().Equal("public.example.com", ruleMatcher0.Name)
	s.Assert().False(ruleMatcher0.Wildcard)
	s.Assert().False(ruleMatcher0.UserWildcard)
	s.Assert().False(ruleMatcher0.GroupWildcard)

	s.Require().Len(tester.rules[1].Domains, 1)

	s.Assert().Equal("one-factor.example.com", tester.configuration.AccessControl.Rules[1].Domains[0])

	ruleMatcher1, ok := tester.rules[1].Domains[0].Matcher.(*AccessControlDomainMatcher)
	s.Require().True(ok)
	s.Assert().Equal("one-factor.example.com", ruleMatcher1.Name)
	s.Assert().False(ruleMatcher1.Wildcard)
	s.Assert().False(ruleMatcher1.UserWildcard)
	s.Assert().False(ruleMatcher1.GroupWildcard)

	s.Require().Len(tester.rules[2].Domains, 1)

	s.Assert().Equal("two-factor.example.com", tester.configuration.AccessControl.Rules[2].Domains[0])

	ruleMatcher2, ok := tester.rules[2].Domains[0].Matcher.(*AccessControlDomainMatcher)
	s.Require().True(ok)
	s.Assert().Equal("two-factor.example.com", ruleMatcher2.Name)
	s.Assert().False(ruleMatcher2.Wildcard)
	s.Assert().False(ruleMatcher2.UserWildcard)
	s.Assert().False(ruleMatcher2.GroupWildcard)

	s.Require().Len(tester.rules[3].Domains, 1)

	s.Assert().Equal("*.example.com", tester.configuration.AccessControl.Rules[3].Domains[0])

	ruleMatcher3, ok := tester.rules[3].Domains[0].Matcher.(*AccessControlDomainMatcher)
	s.Require().True(ok)
	s.Assert().Equal(".example.com", ruleMatcher3.Name)
	s.Assert().True(ruleMatcher3.Wildcard)
	s.Assert().False(ruleMatcher3.UserWildcard)
	s.Assert().False(ruleMatcher3.GroupWildcard)

	s.Require().Len(tester.rules[4].Domains, 1)

	s.Assert().Equal("*.example.com", tester.configuration.AccessControl.Rules[4].Domains[0])

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
		WithRule(schema.ACLRule{
			DomainsRegex: createSliceRegexRule(s.T(), []string{`^.*\.example.com$`}),
			Policy:       bypass,
		}).
		WithRule(schema.ACLRule{
			DomainsRegex: createSliceRegexRule(s.T(), []string{`^.*\.example2.com$`}),
			Policy:       oneFactor,
		}).
		WithRule(schema.ACLRule{
			DomainsRegex: createSliceRegexRule(s.T(), []string{`^(?P<User>[a-zA-Z0-9]+)\.regex.com$`}),
			Policy:       oneFactor,
		}).
		WithRule(schema.ACLRule{
			DomainsRegex: createSliceRegexRule(s.T(), []string{`^group-(?P<Group>[a-zA-Z0-9]+)\.regex.com$`}),
			Policy:       twoFactor,
		}).
		WithRule(schema.ACLRule{
			DomainsRegex: createSliceRegexRule(s.T(), []string{`^.*\.(one|two).com$`}),
			Policy:       twoFactor,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://john.regex.com", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://john.regex.com", "GET", Denied)
	tester.CheckAuthorizations(s.T(), Bob, "https://public.example.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example2.com", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://group-dev.regex.com", "GET", TwoFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://group-dev.regex.com", "GET", Denied)

	s.Require().Len(tester.rules, 5)

	s.Require().Len(tester.rules[0].Domains, 1)

	s.Assert().Equal("^.*\\.example.com$", tester.configuration.AccessControl.Rules[0].DomainsRegex[0].String())

	ruleMatcher0, ok := tester.rules[0].Domains[0].Matcher.(RegexpStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^.*\\.example.com$", ruleMatcher0.String())

	s.Require().Len(tester.rules[1].Domains, 1)

	s.Assert().Equal("^.*\\.example2.com$", tester.configuration.AccessControl.Rules[1].DomainsRegex[0].String())

	ruleMatcher1, ok := tester.rules[1].Domains[0].Matcher.(RegexpStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^.*\\.example2.com$", ruleMatcher1.String())

	s.Require().Len(tester.rules[2].Domains, 1)

	s.Assert().Equal("^(?P<User>[a-zA-Z0-9]+)\\.regex.com$", tester.configuration.AccessControl.Rules[2].DomainsRegex[0].String())

	ruleMatcher2, ok := tester.rules[2].Domains[0].Matcher.(RegexpGroupStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^(?P<User>[a-zA-Z0-9]+)\\.regex.com$", ruleMatcher2.String())

	s.Require().Len(tester.rules[3].Domains, 1)

	s.Assert().Equal("^group-(?P<Group>[a-zA-Z0-9]+)\\.regex.com$", tester.configuration.AccessControl.Rules[3].DomainsRegex[0].String())

	ruleMatcher3, ok := tester.rules[3].Domains[0].Matcher.(RegexpGroupStringSubjectMatcher)
	s.Require().True(ok)
	s.Assert().Equal("^group-(?P<Group>[a-zA-Z0-9]+)\\.regex.com$", ruleMatcher3.String())

	s.Require().Len(tester.rules[4].Domains, 1)

	s.Assert().Equal("^.*\\.(one|two).com$", tester.configuration.AccessControl.Rules[4].DomainsRegex[0].String())

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
		WithRule(schema.ACLRule{
			Domains:   []string{"id.example.com"},
			Policy:    oneFactor,
			Resources: createSliceRegexRule(s.T(), []string{`^/(?P<User>[a-zA-Z0-9]+)/personal(/|/.*)?$`, `^/(?P<Group>[a-zA-Z0-9]+)/group(/|/.*)?$`}),
		}).
		WithRule(schema.ACLRule{
			Domains:   []string{"id.example.com"},
			Policy:    deny,
			Resources: createSliceRegexRule(s.T(), []string{`^/([a-zA-Z0-9]+)/personal(/|/.*)?$`, `^/([a-zA-Z0-9]+)/group(/|/.*)?$`}),
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"id.example.com"},
			Policy:  bypass,
		}).
		Build()

	// Accessing the unprotected root.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com", "GET", Bypass)

	// Accessing Personal page.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/john/personal", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/John/personal", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/bob/personal", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/Bob/personal", "GET", OneFactor)

	// Accessing an invalid users Personal page.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/invaliduser/personal", "GET", Denied)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/invaliduser/personal", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/invaliduser/personal", "GET", Denied)

	// Accessing another users Personal page.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/bob/personal", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/bob/personal", "GET", Denied)
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/Bob/personal", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/Bob/personal", "GET", Denied)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/john/personal", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/john/personal", "GET", Denied)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/John/personal", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/John/personal", "GET", Denied)

	// Accessing a Group page.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/dev/group", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/admins/group", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/dev/group", "GET", Denied)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/admins/group", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/dev/group", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/admins/group", "GET", Denied)

	// Accessing an invalid group's Group page.
	tester.CheckAuthorizations(s.T(), John, "https://id.example.com/invalidgroup/group", "GET", Denied)
	tester.CheckAuthorizations(s.T(), Bob, "https://id.example.com/invalidgroup/group", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://id.example.com/invalidgroup/group", "GET", Denied)

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
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"user:john"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", "GET", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckGroupMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"group:admins"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", "GET", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckSubjectsMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"group:admins"}, {"user:bob"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Sam, "https://protected.example.com/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", "GET", OneFactor)
}

func (s *AuthorizerSuite) TestShouldCheckMultipleSubjectsMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Subjects: [][]string{{"group:admins", "user:bob"}, {"group:admins", "group:dev"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", "GET", OneFactor)
}

func (s *AuthorizerSuite) TestShouldCheckIPMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   bypass,
			Networks: []string{"192.168.1.8", "10.0.0.8"},
		}).
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   oneFactor,
			Networks: []string{"10.0.0.7"},
		}).
		WithRule(schema.ACLRule{
			Domains:  []string{"net.example.com"},
			Policy:   twoFactor,
			Networks: []string{"10.0.0.0/8"},
		}).
		WithRule(schema.ACLRule{
			Domains:  []string{"ipv6.example.com"},
			Policy:   twoFactor,
			Networks: []string{"fec0::1/64"},
		}).
		WithRule(schema.ACLRule{
			Domains:  []string{"ipv6-alt.example.com"},
			Policy:   twoFactor,
			Networks: []string{"fec0::1"},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", "GET", Denied)

	tester.CheckAuthorizations(s.T(), John, "https://net.example.com/", "GET", TwoFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://net.example.com/", "GET", TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://net.example.com/", "GET", Denied)

	tester.CheckAuthorizations(s.T(), Sally, "https://ipv6-alt.example.com/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), Sam, "https://ipv6-alt.example.com/", "GET", TwoFactor)
	tester.CheckAuthorizations(s.T(), Sally, "https://ipv6.example.com/", "GET", TwoFactor)
	tester.CheckAuthorizations(s.T(), Sam, "https://ipv6.example.com/", "GET", TwoFactor)
}

func (s *AuthorizerSuite) TestShouldCheckMethodMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains: []string{"protected.example.com"},
			Policy:  bypass,
			Methods: []string{"OPTIONS", "HEAD", "GET", "CONNECT", "TRACE"},
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"protected.example.com"},
			Policy:  oneFactor,
			Methods: []string{"PUT", "PATCH", "POST"},
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"protected.example.com"},
			Policy:  twoFactor,
			Methods: []string{"DELETE"},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", "OPTIONS", Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", "HEAD", Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", "CONNECT", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", "TRACE", Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", "PUT", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", "PATCH", OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", "POST", OneFactor)

	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", "DELETE", TwoFactor)
}

func (s *AuthorizerSuite) TestShouldCheckResourceMatching() {
	createSliceRegexRule := func(t *testing.T, rules []string) []regexp.Regexp {
		result, err := stringSliceToRegexpSlice(rules)

		require.NoError(t, err)

		return result
	}

	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains:   []string{"resource.example.com"},
			Policy:    bypass,
			Resources: createSliceRegexRule(s.T(), []string{"^/bypass/[a-z]+$", "^/$", "embedded"}),
		}).
		WithRule(schema.ACLRule{
			Domains:   []string{"resource.example.com"},
			Policy:    oneFactor,
			Resources: createSliceRegexRule(s.T(), []string{"^/one_factor/[a-z]+$"}),
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/abc", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/", "GET", Denied)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/ABC", "GET", Denied)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/one_factor/abc", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/xyz/embedded/abc", "GET", Bypass)
}

// This test assures that rules without domains (not allowed by schema validator at this time) will pass validation correctly.
func (s *AuthorizerSuite) TestShouldMatchAnyDomainIfBlank() {
	tester := NewAuthorizerBuilder().
		WithRule(schema.ACLRule{
			Policy:  bypass,
			Methods: []string{"OPTIONS", "HEAD", "GET", "CONNECT", "TRACE"},
		}).
		WithRule(schema.ACLRule{
			Policy:  oneFactor,
			Methods: []string{"PUT", "PATCH"},
		}).
		WithRule(schema.ACLRule{
			Policy:  twoFactor,
			Methods: []string{"DELETE"},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://one.domain-four.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-three.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-two.com", "OPTIONS", Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://one.domain-four.com", "PUT", OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-three.com", "PATCH", OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-two.com", "PUT", OneFactor)

	tester.CheckAuthorizations(s.T(), John, "https://one.domain-four.com", "DELETE", TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-three.com", "DELETE", TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-two.com", "DELETE", TwoFactor)

	tester.CheckAuthorizations(s.T(), John, "https://one.domain-four.com", "POST", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-three.com", "POST", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://one.domain-two.com", "POST", Denied)
}

func (s *AuthorizerSuite) TestShouldMatchResourceWithSubjectRules() {
	createSliceRegexRule := func(t *testing.T, rules []string) []regexp.Regexp {
		result, err := stringSliceToRegexpSlice(rules)

		require.NoError(t, err)

		return result
	}

	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains:   []string{"public.example.com"},
			Resources: createSliceRegexRule(s.T(), []string{"^/admin/.*$"}),
			Subjects:  [][]string{{"group:admins"}},
			Policy:    oneFactor,
		}).
		WithRule(schema.ACLRule{
			Domains:   []string{"public.example.com"},
			Resources: createSliceRegexRule(s.T(), []string{"^/admin/.*$"}),
			Policy:    deny,
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"public.example.com"},
			Policy:  bypass,
		}).
		WithRule(schema.ACLRule{
			Domains:   []string{"public2.example.com"},
			Resources: createSliceRegexRule(s.T(), []string{"^/admin/.*$"}),
			Subjects:  [][]string{{"group:admins"}},
			Policy:    bypass,
		}).
		WithRule(schema.ACLRule{
			Domains:   []string{"public2.example.com"},
			Resources: createSliceRegexRule(s.T(), []string{"^/admin/.*$"}),
			Policy:    deny,
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"public2.example.com"},
			Policy:  bypass,
		}).
		WithRule(schema.ACLRule{
			Domains:  []string{"private.example.com"},
			Subjects: [][]string{{"group:admins"}},
			Policy:   twoFactor,
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://public.example.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://public.example.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com", "GET", Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://public.example.com/admin/index.html", "GET", OneFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://public.example.com/admin/index.html", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/admin/index.html", "GET", OneFactor)

	tester.CheckAuthorizations(s.T(), John, "https://public2.example.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://public2.example.com", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public2.example.com", "GET", Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://public2.example.com/admin/index.html", "GET", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://public2.example.com/admin/index.html", "GET", Denied)

	// This test returns this result since we validate the schema instead of validating it in code.
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public2.example.com/admin/index.html", "GET", Bypass)

	tester.CheckAuthorizations(s.T(), John, "https://private.example.com", "GET", TwoFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://private.example.com", "GET", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://private.example.com", "GET", TwoFactor)

	results := tester.GetRuleMatchResults(John, "https://private.example.com", "GET")

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
	s.Assert().Equal(Bypass, PolicyToLevel(bypass))
	s.Assert().Equal(OneFactor, PolicyToLevel(oneFactor))
	s.Assert().Equal(TwoFactor, PolicyToLevel(twoFactor))
	s.Assert().Equal(Denied, PolicyToLevel(deny))

	s.Assert().Equal(Denied, PolicyToLevel("whatever"))
}

func TestRunSuite(t *testing.T) {
	s := AuthorizerSuite{}
	suite.Run(t, &s)
}

func TestNewAuthorizer(t *testing.T) {
	config := &schema.Configuration{
		AccessControl: schema.AccessControlConfiguration{
			DefaultPolicy: deny,
			Rules: []schema.ACLRule{
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
		AccessControl: schema.AccessControlConfiguration{
			DefaultPolicy: deny,
			Rules: []schema.ACLRule{
				{
					Domains: []string{"example.com"},
					Policy:  oneFactor,
				},
			},
		},
	}

	authorizer := NewAuthorizer(config)
	assert.False(t, authorizer.IsSecondFactorEnabled())

	authorizer.rules[0].Policy = TwoFactor
	assert.True(t, authorizer.IsSecondFactorEnabled())
}

func TestAuthorizerIsSecondFactorEnabledRuleWithOIDC(t *testing.T) {
	config := &schema.Configuration{
		AccessControl: schema.AccessControlConfiguration{
			DefaultPolicy: deny,
			Rules: []schema.ACLRule{
				{
					Domains: []string{"example.com"},
					Policy:  oneFactor,
				},
			},
		},
		IdentityProviders: schema.IdentityProvidersConfiguration{
			OIDC: &schema.OpenIDConnectConfiguration{
				Clients: []schema.OpenIDConnectClientConfiguration{
					{
						Policy: oneFactor,
					},
				},
			},
		},
	}

	authorizer := NewAuthorizer(config)
	assert.False(t, authorizer.IsSecondFactorEnabled())

	authorizer.rules[0].Policy = TwoFactor
	assert.True(t, authorizer.IsSecondFactorEnabled())

	authorizer.rules[0].Policy = OneFactor
	assert.False(t, authorizer.IsSecondFactorEnabled())

	config.IdentityProviders.OIDC.Clients[0].Policy = twoFactor

	assert.True(t, authorizer.IsSecondFactorEnabled())

	authorizer.rules[0].Policy = OneFactor
	config.IdentityProviders.OIDC.Clients[0].Policy = oneFactor

	assert.False(t, authorizer.IsSecondFactorEnabled())

	authorizer.defaultPolicy = TwoFactor

	assert.True(t, authorizer.IsSecondFactorEnabled())
}
