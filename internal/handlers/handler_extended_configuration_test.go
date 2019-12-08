package handlers

import (
	"testing"

	"github.com/clems4ever/authelia/internal/mocks"

	"github.com/clems4ever/authelia/internal/configuration/schema"
	"github.com/stretchr/testify/suite"
)

type SecondFactorAvailableMethodsFixture struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *SecondFactorAvailableMethodsFixture) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
}

func (s *SecondFactorAvailableMethodsFixture) TearDownTest() {
	s.mock.Close()
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldServeDefaultMethods() {
	expectedBody := ExtendedConfigurationBody{
		AvailableMethods: []string{"totp", "u2f"},
	}
	ExtendedConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldServeDefaultMethodsAndMobilePush() {
	s.mock.Ctx.Configuration = schema.Configuration{
		DuoAPI: &schema.DuoAPIConfiguration{},
	}
	expectedBody := ExtendedConfigurationBody{
		AvailableMethods: []string{"totp", "u2f", "mobile_push"},
	}
	ExtendedConfigurationGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), expectedBody)
}

func TestRunSuite(t *testing.T) {
	s := new(SecondFactorAvailableMethodsFixture)
	suite.Run(t, s)
}
