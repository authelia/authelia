package validator

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type AccessControl struct {
	suite.Suite
	config    *schema.Configuration
	validator *schema.StructValidator
}

func (suite *AccessControl) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.config = &schema.Configuration{
		AccessControl: schema.AccessControlConfiguration{
			DefaultPolicy: policyDeny,

			Networks: schema.DefaultACLNetwork,
			Rules:    schema.DefaultACLRule,
		},
	}
}

func (suite *AccessControl) TestShouldValidateCompleteConfiguration() {
	ValidateAccessControl(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)
}

func (suite *AccessControl) TestShouldValidateEitherDomainsOrDomainsRegex() {
	domainsRegex := regexp.MustCompile(`^abc.example.com$`)

	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains: []string{"abc.example.com"},
			Policy:  "bypass",
		},
		{
			DomainsRegex: []regexp.Regexp{*domainsRegex},
			Policy:       "bypass",
		},
		{
			Policy: "bypass",
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	assert.EqualError(suite.T(), suite.validator.Errors()[0], "access control: rule #3: rule is invalid: must have the option 'domain' or 'domain_regex' configured")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidDefaultPolicy() {
	suite.config.AccessControl.DefaultPolicy = testInvalid

	ValidateAccessControl(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: option 'default_policy' must be one of 'bypass', 'one_factor', 'two_factor', 'deny' but it is configured as 'invalid'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidNetworkGroupNetwork() {
	suite.config.AccessControl.Networks = []schema.ACLNetwork{
		{
			Name:     "internal",
			Networks: []string{"abc.def.ghi.jkl"},
		},
	}

	ValidateAccessControl(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: networks: network group 'internal' is invalid: the network 'abc.def.ghi.jkl' is not a valid IP or CIDR notation")
}

func (suite *AccessControl) TestShouldRaiseErrorWithNoRulesDefined() {
	suite.config.AccessControl.Rules = []schema.ACLRule{}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: 'default_policy' option 'deny' is invalid: when no rules are specified it must be 'two_factor' or 'one_factor'")
}

func (suite *AccessControl) TestShouldRaiseWarningWithNoRulesDefined() {
	suite.config.AccessControl.Rules = []schema.ACLRule{}

	suite.config.AccessControl.DefaultPolicy = policyTwoFactor

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Errors(), 0)
	suite.Require().Len(suite.validator.Warnings(), 1)

	suite.Assert().EqualError(suite.validator.Warnings()[0], "access control: no rules have been specified so the 'default_policy' of 'two_factor' is going to be applied to all requests")
}

func (suite *AccessControl) TestShouldRaiseErrorsWithEmptyRules() {
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{},
		{
			Policy: "wrong",
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 4)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1: rule is invalid: must have the option 'domain' or 'domain_regex' configured")
	suite.Assert().EqualError(suite.validator.Errors()[1], "access control: rule #1: rule 'policy' option '' is invalid: must be one of 'deny', 'two_factor', 'one_factor' or 'bypass'")
	suite.Assert().EqualError(suite.validator.Errors()[2], "access control: rule #2: rule is invalid: must have the option 'domain' or 'domain_regex' configured")
	suite.Assert().EqualError(suite.validator.Errors()[3], "access control: rule #2: rule 'policy' option 'wrong' is invalid: must be one of 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidPolicy() {
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains: []string{"public.example.com"},
			Policy:  testInvalid,
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): rule 'policy' option 'invalid' is invalid: must be one of 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidNetwork() {
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains:  []string{"public.example.com"},
			Policy:   "bypass",
			Networks: []string{"abc.def.ghi.jkl/32"},
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): the network 'abc.def.ghi.jkl/32' is not a valid Group Name, IP, or CIDR notation")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidMethod() {
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains: []string{"public.example.com"},
			Policy:  "bypass",
			Methods: []string{"GET", "HOP"},
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): 'methods' option 'HOP' is invalid: must be one of 'GET', 'HEAD', 'POST', 'PUT', 'PATCH', 'DELETE', 'TRACE', 'CONNECT', 'OPTIONS', 'COPY', 'LOCK', 'MKCOL', 'MOVE', 'PROPFIND', 'PROPPATCH', 'UNLOCK'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidSubject() {
	domains := []string{"public.example.com"}
	subjects := [][]string{{testInvalid}}
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains:  domains,
			Policy:   "bypass",
			Subjects: subjects,
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): 'subject' option 'invalid' is invalid: must start with 'user:' or 'group:'")
	suite.Assert().EqualError(suite.validator.Errors()[1], fmt.Sprintf(errAccessControlRuleBypassPolicyInvalidWithSubjects, ruleDescriptor(1, suite.config.AccessControl.Rules[0])))
}

func (suite *AccessControl) TestShouldSetQueryDefaults() {
	domains := []string{"public.example.com"}
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{Operator: "", Key: "example"},
				},
				{
					{Operator: "", Key: "example", Value: "test"},
				},
			},
		},
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{Operator: "pattern", Key: "a", Value: "^(x|y|z)$"},
				},
			},
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Assert().Len(suite.validator.Errors(), 0)

	suite.Assert().Equal("present", suite.config.AccessControl.Rules[0].Query[0][0].Operator)
	suite.Assert().Equal("equal", suite.config.AccessControl.Rules[0].Query[1][0].Operator)

	suite.Require().Len(suite.config.AccessControl.Rules, 2)
	suite.Require().Len(suite.config.AccessControl.Rules[1].Query, 1)
	suite.Require().Len(suite.config.AccessControl.Rules[1].Query[0], 1)

	t := &regexp.Regexp{}

	suite.Assert().IsType(t, suite.config.AccessControl.Rules[1].Query[0][0].Value)
}

func (suite *AccessControl) TestShouldErrorOnInvalidRulesQuery() {
	domains := []string{"public.example.com"}
	suite.config.AccessControl.Rules = []schema.ACLRule{
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{Operator: "equal", Key: "example"},
				},
			},
		},
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{Operator: "present"},
				},
			},
		},
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{Operator: "present", Key: "a"},
				},
			},
		},
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{Operator: "absent", Key: "a"},
				},
			},
		},
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{},
				},
			},
		},
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{Operator: "not", Key: "a", Value: "a"},
				},
			},
		},
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{Operator: "pattern", Key: "a", Value: "(bad pattern"},
				},
			},
		},
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{Operator: "present", Key: "a", Value: "not good"},
				},
			},
		},
		{
			Domains: domains,
			Policy:  "bypass",
			Query: [][]schema.ACLQueryRule{
				{
					{Operator: "present", Key: "a", Value: 5},
				},
			},
		},
	}

	ValidateRules(suite.config, suite.validator)

	suite.Assert().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 7)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): 'query' option 'value' is invalid: must have a value when the operator is 'equal'")
	suite.Assert().EqualError(suite.validator.Errors()[1], "access control: rule #2 (domain 'public.example.com'): 'query' option 'key' is invalid: must have a value")
	suite.Assert().EqualError(suite.validator.Errors()[2], "access control: rule #5 (domain 'public.example.com'): 'query' option 'key' is invalid: must have a value")
	suite.Assert().EqualError(suite.validator.Errors()[3], "access control: rule #6 (domain 'public.example.com'): 'query' option 'operator' with value 'not' is invalid: must be one of 'present', 'absent', 'equal', 'not equal', 'pattern', 'not pattern'")
	suite.Assert().EqualError(suite.validator.Errors()[4], "access control: rule #7 (domain 'public.example.com'): 'query' option 'value' is invalid: error parsing regexp: missing closing ): `(bad pattern`")
	suite.Assert().EqualError(suite.validator.Errors()[5], "access control: rule #8 (domain 'public.example.com'): 'query' option 'value' is invalid: must not have a value when the operator is 'present'")
	suite.Assert().EqualError(suite.validator.Errors()[6], "access control: rule #9 (domain 'public.example.com'): 'query' option 'value' is invalid: expected type was string but got int")
}

func TestAccessControl(t *testing.T) {
	suite.Run(t, new(AccessControl))
}

func TestShouldReturnCorrectResultsForValidNetworkGroups(t *testing.T) {
	config := schema.AccessControlConfiguration{
		Networks: schema.DefaultACLNetwork,
	}

	validNetwork := IsNetworkGroupValid(config, "internal")
	invalidNetwork := IsNetworkGroupValid(config, loopback)

	assert.True(t, validNetwork)
	assert.False(t, invalidNetwork)
}
