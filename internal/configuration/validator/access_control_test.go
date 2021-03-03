package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
)

type AccessControl struct {
	suite.Suite
	configuration schema.AccessControlConfiguration
	validator     *schema.StructValidator
}

func (suite *AccessControl) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration.DefaultPolicy = denyPolicy
	suite.configuration.Networks = schema.DefaultACLNetwork
	suite.configuration.Rules = schema.DefaultACLRule
}

func (suite *AccessControl) TestShouldValidateCompleteConfiguration() {
	ValidateAccessControl(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidDefaultPolicy() {
	suite.configuration.DefaultPolicy = testInvalidPolicy

	ValidateAccessControl(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "'default_policy' must either be 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidNetworkGroupNetwork() {
	suite.configuration.Networks = []schema.ACLNetwork{
		{
			Name:     "internal",
			Networks: []string{"abc.def.ghi.jkl"},
		},
	}

	ValidateAccessControl(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Network [abc.def.ghi.jkl] from network group: internal must be a valid IP or CIDR")
}

func (suite *AccessControl) TestShouldRaiseErrorNoRulesDefined() {
	suite.configuration.Rules = []schema.ACLRule{{}}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.Assert().EqualError(suite.validator.Errors()[0], "No access control rules have been defined")
	suite.Assert().EqualError(suite.validator.Errors()[1], "Policy [] for domain: [] is invalid, a policy must either be 'deny', 'two_factor', 'one_factor' or 'bypass'")
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

	suite.Assert().EqualError(suite.validator.Errors()[0], "Policy [invalid] for domain: [public.example.com] is invalid, a policy must either be 'deny', 'two_factor', 'one_factor' or 'bypass'")
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

	suite.Assert().EqualError(suite.validator.Errors()[0], "Network [abc.def.ghi.jkl/32] for domain: [public.example.com] is not a valid network or network group")
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

	suite.Assert().EqualError(suite.validator.Errors()[0], "Method HOP for domain: [public.example.com] is invalid, must be one of the following methods: GET, HEAD, POST, PUT, PATCH, DELETE, TRACE, CONNECT, OPTIONS")
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

	suite.Assert().EqualError(suite.validator.Errors()[0], "Resource [^/(api.*] for domain: [public.example.com] is invalid, error parsing regexp: missing closing ): `^/(api.*`")
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

	suite.Assert().EqualError(suite.validator.Errors()[0], "Subject [invalid] for domain: [public.example.com] is invalid, must start with 'user:' or 'group:'")
	suite.Assert().EqualError(suite.validator.Errors()[1], fmt.Sprintf(errAccessControlInvalidPolicyWithSubjects, domains, subjects))
}

func TestAccessControl(t *testing.T) {
	suite.Run(t, new(AccessControl))
}
