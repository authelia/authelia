package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/mocks"
)

type HandlerSignU2FStep1Suite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *HandlerSignU2FStep1Suite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
}

func (s *HandlerSignU2FStep1Suite) TearDownTest() {
	s.mock.Close()
}

func (s *HandlerSignU2FStep1Suite) TestShouldRaiseWhenXForwardedProtoIsMissing() {
	SecondFactorU2FSignGet(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "Missing header X-Forwarded-Proto", s.mock.Hook.LastEntry().Message)
}

func (s *HandlerSignU2FStep1Suite) TestShouldRaiseWhenXForwardedHostIsMissing() {
	s.mock.Ctx.Request.Header.Add("X-Forwarded-Proto", "http")
	SecondFactorU2FSignGet(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "Missing header X-Forwarded-Host", s.mock.Hook.LastEntry().Message)
}

func TestShouldRunHandlerSignU2FStep1Suite(t *testing.T) {
	suite.Run(t, new(HandlerSignU2FStep1Suite))
}
