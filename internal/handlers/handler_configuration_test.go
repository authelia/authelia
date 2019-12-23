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

	expectedBody := ConfigurationBody{
		GoogleAnalyticsTrackingID: GATrackingID,
	}

	ConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}
