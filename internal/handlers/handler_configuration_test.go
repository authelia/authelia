package handlers

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
)

type ConfigurationHandlerFixture struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *ConfigurationHandlerFixture) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "deny",
			Rules:         []schema.AccessControlRule{},
		}})
}

func (s *ConfigurationHandlerFixture) TearDownTest() {
	s.mock.Close()
}

func (s *ConfigurationHandlerFixture) TestShouldHaveAllConfiguredMethods() {
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

func (s *ConfigurationHandlerFixture) TestShouldRemoveTOTPFromAvailableMethodsWhenDisabled() {
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

func (s *ConfigurationHandlerFixture) TestShouldRemoveWebAuthnFromAvailableMethodsWhenDisabled() {
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

func (s *ConfigurationHandlerFixture) TestShouldRemoveDuoFromAvailableMethodsWhenNotConfigured() {
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

func (s *ConfigurationHandlerFixture) TestShouldRemoveAllMethodsWhenNoTwoFactorACLRulesConfigured() {
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

func (s *ConfigurationHandlerFixture) TestShouldRemoveAllMethodsWhenAllDisabledOrNotConfigured() {
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

func (s *ConfigurationHandlerFixture) TestDisablePasswordResetChangeOptions() {
	testCases := []struct {
		name                  string
		passwordChangeDisable bool
		passwordResetDisable  bool
	}{
		{
			name:                  "BothEnabled",
			passwordChangeDisable: false,
			passwordResetDisable:  false,
		},
		{
			name:                  "PasswordChangeDisabled",
			passwordChangeDisable: true,
			passwordResetDisable:  false,
		},
		{
			name:                  "PasswordResetDisabled",
			passwordChangeDisable: false,
			passwordResetDisable:  true,
		},
		{
			name:                  "BothDisabled",
			passwordChangeDisable: true,
			passwordResetDisable:  true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Configuration = schema.Configuration{
				AuthenticationBackend: schema.AuthenticationBackend{
					PasswordChange: schema.AuthenticationBackendPasswordChange{
						Disable: tc.passwordChangeDisable,
					},
					PasswordReset: schema.AuthenticationBackendPasswordReset{
						Disable: tc.passwordResetDisable,
					},
				},
				DuoAPI: schema.DuoAPI{
					Disable: true,
				},
				TOTP: schema.TOTP{
					Disable: true,
				},
				WebAuthn: schema.WebAuthn{
					Disable: true,
				},
			}

			ConfigurationGET(mock.Ctx)

			mock.Assert200OK(s.T(), configurationBody{
				AvailableMethods:       []string{},
				PasswordChangeDisabled: tc.passwordChangeDisable,
				PasswordResetDisabled:  tc.passwordResetDisable,
			})
		})
	}
}

func TestRunSuite(t *testing.T) {
	s := new(ConfigurationHandlerFixture)
	suite.Run(t, s)
}
