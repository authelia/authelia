package handlers

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/mocks"
)

type SecondFactorAvailableMethodsFixture struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *SecondFactorAvailableMethodsFixture) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(schema.AccessControlConfiguration{
		DefaultPolicy: "deny",
		Rules:         []schema.ACLRule{},
	})
}

func (s *SecondFactorAvailableMethodsFixture) TearDownTest() {
	s.mock.Close()
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldServeDefaultMethods() {
	s.mock.Ctx.Configuration = schema.Configuration{
		TOTP: &schema.TOTPConfiguration{
			Period: schema.DefaultTOTPConfiguration.Period,
		},
	}
	expectedBody := ConfigurationBody{
		AvailableMethods:    []string{"totp", "u2f"},
		SecondFactorEnabled: false,
		TOTPPeriod:          schema.DefaultTOTPConfiguration.Period,
	}

	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldServeDefaultMethodsAndMobilePush() {
	s.mock.Ctx.Configuration = schema.Configuration{
		DuoAPI: &schema.DuoAPIConfiguration{},
		TOTP: &schema.TOTPConfiguration{
			Period: schema.DefaultTOTPConfiguration.Period,
		},
	}
	expectedBody := ConfigurationBody{
		AvailableMethods:    []string{"totp", "u2f", "mobile_push"},
		SecondFactorEnabled: false,
		TOTPPeriod:          schema.DefaultTOTPConfiguration.Period,
	}

	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldCheckSecondFactorIsDisabledWhenNoRuleIsSetToTwoFactor() {
	s.mock.Ctx.Configuration = schema.Configuration{
		TOTP: &schema.TOTPConfiguration{
			Period: schema.DefaultTOTPConfiguration.Period,
		},
	}
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(schema.AccessControlConfiguration{
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
	})
	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), ConfigurationBody{
		AvailableMethods:    []string{"totp", "u2f"},
		SecondFactorEnabled: false,
		TOTPPeriod:          schema.DefaultTOTPConfiguration.Period,
	})
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldCheckSecondFactorIsEnabledWhenDefaultPolicySetToTwoFactor() {
	s.mock.Ctx.Configuration = schema.Configuration{
		TOTP: &schema.TOTPConfiguration{
			Period: schema.DefaultTOTPConfiguration.Period,
		},
	}
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(schema.AccessControlConfiguration{
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
	})
	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), ConfigurationBody{
		AvailableMethods:    []string{"totp", "u2f"},
		SecondFactorEnabled: true,
		TOTPPeriod:          schema.DefaultTOTPConfiguration.Period,
	})
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldCheckSecondFactorIsEnabledWhenSomePolicySetToTwoFactor() {
	s.mock.Ctx.Configuration = schema.Configuration{
		TOTP: &schema.TOTPConfiguration{
			Period: schema.DefaultTOTPConfiguration.Period,
		},
	}
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(schema.AccessControlConfiguration{
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
	})
	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), ConfigurationBody{
		AvailableMethods:    []string{"totp", "u2f"},
		SecondFactorEnabled: true,
		TOTPPeriod:          schema.DefaultTOTPConfiguration.Period,
	})
}

func TestRunSuite(t *testing.T) {
	s := new(SecondFactorAvailableMethodsFixture)
	suite.Run(t, s)
}
