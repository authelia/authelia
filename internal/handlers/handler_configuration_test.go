package handlers

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/mocks"
	"github.com/authelia/authelia/internal/session"
)

type ConfigurationSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *ConfigurationSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
}

func (s *ConfigurationSuite) TearDownTest() {
	s.mock.Close()
}

func (s *ConfigurationSuite) TestShouldReturnConfiguredGATrackingID() {
	s.mock.Ctx.Configuration.Session.RememberMeDuration = schema.DefaultSessionConfiguration.RememberMeDuration

	expectedBody := ConfigurationBody{
		RememberMe:    true,
		ResetPassword: true,
	}

	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}

func (s *ConfigurationSuite) TestShouldDisableRememberMe() {
	s.mock.Ctx.Configuration.Session.RememberMeDuration = "0"
	s.mock.Ctx.Providers.SessionProvider = session.NewProvider(
		s.mock.Ctx.Configuration.Session)
	expectedBody := ConfigurationBody{
		RememberMe:    false,
		ResetPassword: true,
	}

	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}

func (s *ConfigurationSuite) TestShouldDisableResetPassword() {
	s.mock.Ctx.Configuration.AuthenticationBackend.DisableResetPassword = true
	expectedBody := ConfigurationBody{
		RememberMe:    true,
		ResetPassword: false,
	}

	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}

func TestRunHandlerConfigurationSuite(t *testing.T) {
	s := new(ConfigurationSuite)
	suite.Run(t, s)
}
