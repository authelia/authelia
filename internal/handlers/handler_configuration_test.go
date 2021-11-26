package handlers

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
)

type SecondFactorAvailableMethodsFixture struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *SecondFactorAvailableMethodsFixture) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControlConfiguration{
			DefaultPolicy: "deny",
			Rules:         []schema.ACLRule{},
		}})
}

func (s *SecondFactorAvailableMethodsFixture) TearDownTest() {
	s.mock.Close()
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldServeDefaultMethods() {
	expectedBody := configurationBody{
		AvailableMethods:    []string{"totp", "u2f"},
		SecondFactorEnabled: false,
	}

	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldServeDefaultMethodsAndMobilePush() {
	s.mock.Ctx.Configuration = schema.Configuration{
		DuoAPI: &schema.DuoAPIConfiguration{},
	}
	expectedBody := configurationBody{
		AvailableMethods:    []string{"totp", "u2f", "mobile_push"},
		SecondFactorEnabled: false,
	}

	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldCheckSecondFactorIsDisabledWhenNoRuleIsSetToTwoFactor() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(
		&schema.Configuration{
			AccessControl: schema.AccessControlConfiguration{
				DefaultPolicy: "bypass",
				Rules: []schema.ACLRule{
					{
						Domains: []string{"example.com"},
						Policy:  "deny",
					},
					{
						Domains: []string{"abc.example.com"},
						Policy:  "single_factor",
					},
					{
						Domains: []string{"def.example.com"},
						Policy:  "bypass",
					},
				},
			}})
	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), configurationBody{
		AvailableMethods:    []string{"totp", "u2f"},
		SecondFactorEnabled: false,
	})
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldCheckSecondFactorIsEnabledWhenDefaultPolicySetToTwoFactor() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControlConfiguration{
			DefaultPolicy: "two_factor",
			Rules: []schema.ACLRule{
				{
					Domains: []string{"example.com"},
					Policy:  "deny",
				},
				{
					Domains: []string{"abc.example.com"},
					Policy:  "single_factor",
				},
				{
					Domains: []string{"def.example.com"},
					Policy:  "bypass",
				},
			},
		}})
	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), configurationBody{
		AvailableMethods:    []string{"totp", "u2f"},
		SecondFactorEnabled: true,
	})
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldCheckSecondFactorIsEnabledWhenSomePolicySetToTwoFactor() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(
		&schema.Configuration{
			AccessControl: schema.AccessControlConfiguration{
				DefaultPolicy: "bypass",
				Rules: []schema.ACLRule{
					{
						Domains: []string{"example.com"},
						Policy:  "deny",
					},
					{
						Domains: []string{"abc.example.com"},
						Policy:  "two_factor",
					},
					{
						Domains: []string{"def.example.com"},
						Policy:  "bypass",
					},
				},
			}})
	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), configurationBody{
		AvailableMethods:    []string{"totp", "u2f"},
		SecondFactorEnabled: true,
	})
}

func TestRunSuite(t *testing.T) {
	s := new(SecondFactorAvailableMethodsFixture)
	suite.Run(t, s)
}
