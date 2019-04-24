package handlers

import (
	"testing"

	"github.com/clems4ever/authelia/mocks"

	"github.com/clems4ever/authelia/configuration/schema"
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
	SecondFactorAvailableMethodsGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), []string{"totp", "u2f"})
}

func (s *SecondFactorAvailableMethodsFixture) TestShouldServeDefaultMethodsAndDuo() {
	s.mock.Ctx.Configuration = schema.Configuration{
		DuoAPI: &schema.DuoAPIConfiguration{},
	}
	SecondFactorAvailableMethodsGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), []string{"totp", "u2f", "duo_push"})
}

func TestRunSuite(t *testing.T) {
	s := new(SecondFactorAvailableMethodsFixture)
	suite.Run(t, s)
}
