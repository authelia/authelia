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
		AccessControl: schema.AccessControl{
			DefaultPolicy: "deny",
			Rules:         []schema.AccessControlRule{},
		}})
}

func (s *SecondFactorAvailableMethodsFixture) TearDownTest() {
	s.mock.Close()
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldHaveAllConfiguredMethods() {
	s.mock.Ctx.Configuration = schema.Configuration{
		DuoAPI: schema.DuoAPI{
			Disable: false,
		},
		TOTP: schema.TOTP{
			Disable: false,
		},
		WebAuthn: schema.WebAuthn{
			Disable: false,
		},
		AccessControl: schema.AccessControl{
			DefaultPolicy: "deny",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"example.com"},
					Policy:  "two_factor",
				},
			},
		}}

	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&s.mock.Ctx.Configuration)

	ConfigurationGET(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), configurationBody{
		AvailableMethods: []string{"totp", "webauthn", "mobile_push"},
	})
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldRemoveTOTPFromAvailableMethodsWhenDisabled() {
	s.mock.Ctx.Configuration = schema.Configuration{
		DuoAPI: schema.DuoAPI{
			Disable: false,
		},
		TOTP: schema.TOTP{
			Disable: true,
		},
		WebAuthn: schema.WebAuthn{
			Disable: false,
		},
		AccessControl: schema.AccessControl{
			DefaultPolicy: "deny",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"example.com"},
					Policy:  "two_factor",
				},
			},
		}}

	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&s.mock.Ctx.Configuration)

	ConfigurationGET(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), configurationBody{
		AvailableMethods: []string{"webauthn", "mobile_push"},
	})
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldRemoveWebAuthnFromAvailableMethodsWhenDisabled() {
	s.mock.Ctx.Configuration = schema.Configuration{
		DuoAPI: schema.DuoAPI{
			Disable: false,
		},
		TOTP: schema.TOTP{
			Disable: false,
		},
		WebAuthn: schema.WebAuthn{
			Disable: true,
		},
		AccessControl: schema.AccessControl{
			DefaultPolicy: "deny",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"example.com"},
					Policy:  "two_factor",
				},
			},
		}}

	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&s.mock.Ctx.Configuration)

	ConfigurationGET(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), configurationBody{
		AvailableMethods: []string{"totp", "mobile_push"},
	})
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldRemoveDuoFromAvailableMethodsWhenNotConfigured() {
	s.mock.Ctx.Configuration = schema.Configuration{
		DuoAPI: schema.DuoAPI{
			Disable: true,
		},
		TOTP: schema.TOTP{
			Disable: false,
		},
		WebAuthn: schema.WebAuthn{
			Disable: false,
		},
		AccessControl: schema.AccessControl{
			DefaultPolicy: "deny",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"example.com"},
					Policy:  "two_factor",
				},
			},
		}}

	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&s.mock.Ctx.Configuration)

	ConfigurationGET(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), configurationBody{
		AvailableMethods: []string{"totp", "webauthn"},
	})
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldRemoveAllMethodsWhenNoTwoFactorACLRulesConfigured() {
	s.mock.Ctx.Configuration = schema.Configuration{
		DuoAPI: schema.DuoAPI{
			Disable: false,
		},
		TOTP: schema.TOTP{
			Disable: false,
		},
		WebAuthn: schema.WebAuthn{
			Disable: false,
		},
		AccessControl: schema.AccessControl{
			DefaultPolicy: "deny",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"example.com"},
					Policy:  "one_factor",
				},
			},
		}}

	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&s.mock.Ctx.Configuration)

	ConfigurationGET(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), configurationBody{
		AvailableMethods: []string{},
	})
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldRemoveAllMethodsWhenAllDisabledOrNotConfigured() {
	s.mock.Ctx.Configuration = schema.Configuration{
		DuoAPI: schema.DuoAPI{
			Disable: true,
		},
		TOTP: schema.TOTP{
			Disable: true,
		},
		WebAuthn: schema.WebAuthn{
			Disable: true,
		},
		AccessControl: schema.AccessControl{
			DefaultPolicy: "deny",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"example.com"},
					Policy:  "two_factor",
				},
			},
		}}

	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&s.mock.Ctx.Configuration)

	ConfigurationGET(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), configurationBody{
		AvailableMethods: []string{},
	})
}

func TestRunSuite(t *testing.T) {
	s := new(SecondFactorAvailableMethodsFixture)
	suite.Run(t, s)
}
