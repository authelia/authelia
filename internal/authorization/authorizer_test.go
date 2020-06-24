package authorization

import (
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
)

type AuthorizerSuite struct {
	suite.Suite
}

type AuthorizerTester struct {
	*Authorizer
}

func NewAuthorizerTester(config schema.AccessControlConfiguration) *AuthorizerTester {
	return &AuthorizerTester{
		NewAuthorizer(config),
	}
}

func (s *AuthorizerTester) CheckAuthorizations(t *testing.T, subject Subject, requestURI string, expectedLevel Level) {
	url, _ := url.ParseRequestURI(requestURI)
	level := s.GetRequiredLevel(Subject{
		Groups:   subject.Groups,
		Username: subject.Username,
		IP:       subject.IP,
	}, *url)

	assert.Equal(t, expectedLevel, level)
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

func (s *AuthorizerSuite) TestShouldCheckDefaultBypassConfig() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("bypass").Build()

	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/elsewhere", Bypass)
}

func (s *AuthorizerSuite) TestShouldCheckDefaultDeniedConfig() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").Build()

	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", Denied)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/", Denied)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/elsewhere", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckMultiDomainRule() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domains: []string{"*.example.com"},
			Policy:  "bypass",
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://private.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/elsewhere", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com.c/", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.co/", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckMultipleDomainRule() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domains: []string{"*.example.com", "other.com"},
			Policy:  "bypass",
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://private.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/elsewhere", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com.c/", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.co/", Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://other.com/", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://other.com/elsewhere", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://private.other.com/", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckFactorsPolicy() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domains: []string{"single.example.com"},
			Policy:  "one_factor",
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"protected.example.com"},
			Policy:  "two_factor",
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"public.example.com"},
			Policy:  "bypass",
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://protected.example.com/", TwoFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://single.example.com/", OneFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckRulePrecedence() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   "bypass",
			Subjects: [][]string{{"user:john"}},
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"protected.example.com"},
			Policy:  "one_factor",
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"*.example.com"},
			Policy:  "two_factor",
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://public.example.com/", TwoFactor)
}

func (s *AuthorizerSuite) TestShouldCheckUserMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   "bypass",
			Subjects: [][]string{{"user:john"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckGroupMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   "bypass",
			Subjects: [][]string{{"group:admins"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckSubjectsMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   "bypass",
			Subjects: [][]string{{"group:admins"}, {"user:bob"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckMultipleSubjectsMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   "bypass",
			Subjects: [][]string{{"group:admins", "user:bob"}, {"group:admins", "group:dev"}},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", Denied)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckIPMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   "bypass",
			Networks: []string{"192.168.1.8", "10.0.0.8"},
		}).
		WithRule(schema.ACLRule{
			Domains:  []string{"protected.example.com"},
			Policy:   "one_factor",
			Networks: []string{"10.0.0.7"},
		}).
		WithRule(schema.ACLRule{
			Domains:  []string{"net.example.com"},
			Policy:   "two_factor",
			Networks: []string{"10.0.0.0/8"},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", Denied)

	tester.CheckAuthorizations(s.T(), John, "https://net.example.com/", TwoFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://net.example.com/", TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://net.example.com/", Denied)
}

func (s *AuthorizerSuite) TestShouldCheckResourceMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domains:   []string{"resource.example.com"},
			Policy:    "bypass",
			Resources: []string{"^/bypass/[a-z]+$", "^/$", "embedded"},
		}).
		WithRule(schema.ACLRule{
			Domains:   []string{"resource.example.com"},
			Policy:    "one_factor",
			Resources: []string{"^/one_factor/[a-z]+$"},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/", Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/abc", Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/", Denied)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/ABC", Denied)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/one_factor/abc", OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/xyz/embedded/abc", Bypass)
}

func (s *AuthorizerSuite) TestPolicyToLevel() {
	s.Assert().Equal(Bypass, PolicyToLevel("bypass"))
	s.Assert().Equal(OneFactor, PolicyToLevel("one_factor"))
	s.Assert().Equal(TwoFactor, PolicyToLevel("two_factor"))
	s.Assert().Equal(Denied, PolicyToLevel("deny"))

	s.Assert().Equal(Denied, PolicyToLevel("whatever"))
}

func TestRunSuite(t *testing.T) {
	s := AuthorizerSuite{}
	suite.Run(t, &s)
}
