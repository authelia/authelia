package authorization

import (
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"

	"github.com/stretchr/testify/assert"
)

var NoNet = []string{}
var LocalNet = []string{"127.0.0.1"}
var PrivateNet = []string{"192.168.1.0/24"}
var MultipleNet = []string{"192.168.1.0/24", "10.0.0.0/8"}
var MixedNetIP = []string{"192.168.1.0/24", "192.168.2.4"}

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

func (s *AuthorizerTester) CheckAuthorizations(t *testing.T, subject Subject, requestURI string, method []byte, expectedLevel Level) {
	url, _ := url.ParseRequestURI(requestURI)
	level := s.GetRequiredLevel(Subject{
		Groups:   subject.Groups,
		Username: subject.Username,
		IP:       subject.IP,
	}, *url, method)

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

	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/elsewhere", []byte("GET"), Bypass)
}

func (s *AuthorizerSuite) TestShouldCheckDefaultDeniedConfig() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").Build()

	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://public.example.com/", []byte("GET"), Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", []byte("GET"), Denied)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/", []byte("GET"), Denied)
	tester.CheckAuthorizations(s.T(), UserWithoutGroups, "https://public.example.com/elsewhere", []byte("GET"), Denied)
}

func (s *AuthorizerSuite) TestShouldCheckMultiDomainRule() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domain: "*.example.com",
			Policy: "bypass",
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://private.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/elsewhere", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", []byte("GET"), Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com.c/", []byte("GET"), Denied)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.co/", []byte("GET"), Denied)
}

func (s *AuthorizerSuite) TestShouldCheckFactorsPolicy() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domain: "single.example.com",
			Policy: "one_factor",
		}).
		WithRule(schema.ACLRule{
			Domain: "protected.example.com",
			Policy: "two_factor",
		}).
		WithRule(schema.ACLRule{
			Domain: "public.example.com",
			Policy: "bypass",
		}).
		Build()

	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://public.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://protected.example.com/", []byte("GET"), TwoFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://single.example.com/", []byte("GET"), OneFactor)
	tester.CheckAuthorizations(s.T(), UserWithGroups, "https://example.com/", []byte("GET"), Denied)
}

func (s *AuthorizerSuite) TestShouldCheckRulePrecedence() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domain:  "protected.example.com",
			Policy:  "bypass",
			Subject: "user:john",
		}).
		WithRule(schema.ACLRule{
			Domain: "protected.example.com",
			Policy: "one_factor",
		}).
		WithRule(schema.ACLRule{
			Domain: "*.example.com",
			Policy: "two_factor",
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", []byte("GET"), OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://public.example.com/", []byte("GET"), TwoFactor)
}

func (s *AuthorizerSuite) TestShouldCheckUserMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domain:  "protected.example.com",
			Policy:  "bypass",
			Subject: "user:john",
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", []byte("GET"), Denied)
}

func (s *AuthorizerSuite) TestShouldCheckGroupMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domain:  "protected.example.com",
			Policy:  "bypass",
			Subject: "group:admins",
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", []byte("GET"), Denied)
}

func (s *AuthorizerSuite) TestShouldCheckMethodMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domain:  "api.example.com",
			Policy:  "bypass",
			Methods: []string{"OPTIONS"},
		}).
		WithRule(schema.ACLRule{
			Domain:  "api.example.com",
			Policy:  "one_factor",
			Methods: []string{"GET", "HEAD"},
		}).
		WithRule(schema.ACLRule{
			Domain:  "api.example.com",
			Policy:  "two_factor",
			Methods: []string{"DELETE", "UPDATE"},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://api.example.com", []byte("OPTIONS"), Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://api.example.com", []byte("GET"), OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://api.example.com", []byte("DELETE"), TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://api.example.com", []byte("PATCH"), Denied)
}

func (s *AuthorizerSuite) TestShouldCheckMethodMatchingWithNoHeader() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domain:  "api.example.com",
			Policy:  "bypass",
			Methods: []string{"OPTIONS"},
		}).
		WithRule(schema.ACLRule{
			Domain:  "api.example.com",
			Policy:  "one_factor",
			Methods: []string{"GET", "HEAD", "DELETE", "UPDATE"},
		}).
		WithRule(schema.ACLRule{
			Domain: "api.example.com",
			Policy: "two_factor",
		}).
		WithRule(schema.ACLRule{
			Domain:  "other.example.com",
			Policy:  "two_factor",
			Methods: []string{"GET", "HEAD", "DELETE", "UPDATE", "OPTIONS", "PATCH", "PUT", "TRACE", "POST", "CONNECT"},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://api.example.com", []byte(nil), TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://api.example.com", []byte(nil), TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://api.example.com", []byte(nil), TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://api.example.com", []byte(nil), TwoFactor)
	tester.CheckAuthorizations(s.T(), John, "https://other.example.com", []byte(nil), Denied)
}

func (s *AuthorizerSuite) TestShouldCheckIPMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domain:   "protected.example.com",
			Policy:   "bypass",
			Networks: []string{"192.168.1.8", "10.0.0.8"},
		}).
		WithRule(schema.ACLRule{
			Domain:   "protected.example.com",
			Policy:   "one_factor",
			Networks: []string{"10.0.0.7"},
		}).
		WithRule(schema.ACLRule{
			Domain:   "net.example.com",
			Policy:   "two_factor",
			Networks: []string{"10.0.0.0/8"},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://protected.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), Bob, "https://protected.example.com/", []byte("GET"), OneFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://protected.example.com/", []byte("GET"), Denied)

	tester.CheckAuthorizations(s.T(), John, "https://net.example.com/", []byte("GET"), TwoFactor)
	tester.CheckAuthorizations(s.T(), Bob, "https://net.example.com/", []byte("GET"), TwoFactor)
	tester.CheckAuthorizations(s.T(), AnonymousUser, "https://net.example.com/", []byte("GET"), Denied)
}

func (s *AuthorizerSuite) TestShouldCheckResourceMatching() {
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy("deny").
		WithRule(schema.ACLRule{
			Domain:    "resource.example.com",
			Policy:    "bypass",
			Resources: []string{"^/bypass/[a-z]+$", "^/$", "embedded"},
		}).
		WithRule(schema.ACLRule{
			Domain:    "resource.example.com",
			Policy:    "one_factor",
			Resources: []string{"^/one_factor/[a-z]+$"},
		}).
		Build()

	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/abc", []byte("GET"), Bypass)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/", []byte("GET"), Denied)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/bypass/ABC", []byte("GET"), Denied)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/one_factor/abc", []byte("GET"), OneFactor)
	tester.CheckAuthorizations(s.T(), John, "https://resource.example.com/xyz/embedded/abc", []byte("GET"), Bypass)
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
