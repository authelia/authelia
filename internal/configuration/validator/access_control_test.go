package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type AccessControl struct {
	suite.Suite
	configuration schema.AccessControlConfiguration
	validator     *schema.StructValidator
}

func (suite *AccessControl) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration.DefaultPolicy = policyDeny
	suite.configuration.Networks = schema.DefaultACLNetwork
	suite.configuration.Rules = schema.DefaultACLRule
}

func (suite *AccessControl) TestShouldValidateCompleteConfiguration() {
	ValidateAccessControl(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidDefaultPolicy() {
	suite.configuration.DefaultPolicy = testInvalidPolicy

	ValidateAccessControl(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: 'default_policy' option 'invalid' is invalid: must either be 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidNetworkGroupNetwork() {
	suite.configuration.Networks = []schema.ACLNetwork{
		{
			Name:     "internal",
			Networks: []string{"abc.def.ghi.jkl"},
		},
	}

	ValidateAccessControl(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: networks: network group 'internal' is invalid: the network 'abc.def.ghi.jkl' is not a valid IP or CIDR notation")
}

func (suite *AccessControl) TestShouldRaiseErrorWithNoRulesDefined() {
	suite.configuration.Rules = []schema.ACLRule{}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: 'default_policy' option 'deny' is invalid: when no rules are specified it must be 'two_factor' or 'one_factor'")
}

func (suite *AccessControl) TestShouldRaiseWarningWithNoRulesDefined() {
	suite.configuration.Rules = []schema.ACLRule{}

	suite.configuration.DefaultPolicy = policyTwoFactor

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasErrors())
	suite.Require().Len(suite.validator.Warnings(), 1)

	suite.Assert().EqualError(suite.validator.Warnings()[0], "access control: no rules have been specified so the 'default_policy' of 'two_factor' is going to be applied to all requests")
}

func (suite *AccessControl) TestShouldRaiseErrorsWithEmptyRules() {
	suite.configuration.Rules = []schema.ACLRule{
		{},
		{
			Policy: "wrong",
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 4)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1: rule is invalid: must have the option 'domain' configured")
	suite.Assert().EqualError(suite.validator.Errors()[1], "access control: rule #1: rule 'policy' option '' is invalid: must be one of 'deny', 'two_factor', 'one_factor' or 'bypass'")
	suite.Assert().EqualError(suite.validator.Errors()[2], "access control: rule #2: rule is invalid: must have the option 'domain' configured")
	suite.Assert().EqualError(suite.validator.Errors()[3], "access control: rule #2: rule 'policy' option 'wrong' is invalid: must be one of 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidPolicy() {
	suite.configuration.Rules = []schema.ACLRule{
		{
			Domains: []string{"public.example.com"},
			Policy:  testInvalidPolicy,
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): rule 'policy' option 'invalid' is invalid: must be one of 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidNetwork() {
	suite.configuration.Rules = []schema.ACLRule{
		{
			Domains:  []string{"public.example.com"},
			Policy:   "bypass",
			Networks: []string{"abc.def.ghi.jkl/32"},
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): the network 'abc.def.ghi.jkl/32' is not a valid Group Name, IP, or CIDR notation")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidMethod() {
	suite.configuration.Rules = []schema.ACLRule{
		{
			Domains: []string{"public.example.com"},
			Policy:  "bypass",
			Methods: []string{"GET", "HOP"},
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): 'methods' option 'HOP' is invalid: must be one of 'GET', 'HEAD', 'POST', 'PUT', 'PATCH', 'DELETE', 'TRACE', 'CONNECT', 'OPTIONS'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidResource() {
	suite.configuration.Rules = []schema.ACLRule{
		{
			Domains:   []string{"public.example.com"},
			Policy:    "bypass",
			Resources: []string{"^/(api.*"},
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): 'resources' option '^/(api.*' is invalid: error parsing regexp: missing closing ): `^/(api.*`")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidSubject() {
	domains := []string{"public.example.com"}
	subjects := [][]string{{"invalid"}}
	suite.configuration.Rules = []schema.ACLRule{
		{
			Domains:  domains,
			Policy:   "bypass",
			Subjects: subjects,
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.Assert().EqualError(suite.validator.Errors()[0], "access control: rule #1 (domain 'public.example.com'): 'subject' option 'invalid' is invalid: must start with 'user:' or 'group:'")
	suite.Assert().EqualError(suite.validator.Errors()[1], fmt.Sprintf(errAccessControlRuleBypassPolicyInvalidWithSubjects, ruleDescriptor(1, suite.configuration.Rules[0])))
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
