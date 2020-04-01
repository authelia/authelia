package handlers

import (
	"github.com/authelia/authelia/internal/mocks"
	"github.com/stretchr/testify/suite"
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
	GATrackingID := "ABC"
	s.mock.Ctx.Configuration.GoogleAnalyticsTrackingID = GATrackingID
	s.mock.Ctx.Configuration.Session.RememberMeDuration = "1y"

	expectedBody := ConfigurationBody{
		GoogleAnalyticsTrackingID: GATrackingID,
		RememberMeEnabled:         true,
	}

	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}

func (s *ConfigurationSuite) TestShouldDisableRememberMe() {
	GATrackingID := "ABC"
	s.mock.Ctx.Configuration.GoogleAnalyticsTrackingID = GATrackingID
	s.mock.Ctx.Configuration.Session.RememberMeDuration = "0y"

	expectedBody := ConfigurationBody{
		GoogleAnalyticsTrackingID: GATrackingID,
		RememberMeEnabled:         false,
	}

	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}
