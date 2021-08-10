package authorization

import (
	"net"
	"net/url"
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
	url, _ := url.ParseRequestURI(requestURI)

	object := Object{
		Scheme: url.Scheme,
		Domain: url.Hostname(),
		Path:   url.Path,
		Method: method,
	}

	level := s.GetRequiredLevel(subject, object)

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
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains:   []string{"resource.example.com"},
			Policy:    bypass,
			Resources: []string{"^/bypass/[a-z]+$", "^/$", "embedded"},
		}).
		WithRule(schema.ACLRule{
			Domains:   []string{"resource.example.com"},
			Policy:    oneFactor,
			Resources: []string{"^/one_factor/[a-z]+$"},
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
	tester := NewAuthorizerBuilder().
		WithDefaultPolicy(deny).
		WithRule(schema.ACLRule{
			Domains:   []string{"public.example.com"},
			Resources: []string{"^/admin/.*$"},
			Subjects:  [][]string{{"group:admins"}},
			Policy:    oneFactor,
		}).
		WithRule(schema.ACLRule{
			Domains:   []string{"public.example.com"},
			Resources: []string{"^/admin/.*$"},
			Policy:    deny,
		}).
		WithRule(schema.ACLRule{
			Domains: []string{"public.example.com"},
			Policy:  bypass,
		}).
		WithRule(schema.ACLRule{
			Domains:   []string{"public2.example.com"},
			Resources: []string{"^/admin/.*$"},
			Subjects:  [][]string{{"group:admins"}},
			Policy:    bypass,
		}).
		WithRule(schema.ACLRule{
			Domains:   []string{"public2.example.com"},
			Resources: []string{"^/admin/.*$"},
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
