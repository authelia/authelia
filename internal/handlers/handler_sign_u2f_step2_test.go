package handlers

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tstranex/u2f"

	"github.com/authelia/authelia/internal/mocks"
	"github.com/authelia/authelia/internal/session"
)

type HandlerSignU2FStep2Suite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *HandlerSignU2FStep2Suite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.U2FChallenge = &u2f.Challenge{}
	userSession.U2FRegistration = &session.U2FRegistration{}
	err := s.mock.Ctx.SaveSession(userSession)
	require.NoError(s.T(), err)
}

func (s *HandlerSignU2FStep2Suite) TearDownTest() {
	s.mock.Close()
}

func (s *HandlerSignU2FStep2Suite) TestShouldRedirectUserToDefaultURL() {
	u2fVerifier := NewMockU2FVerifier(s.mock.Ctrl)

	u2fVerifier.EXPECT().
		Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	s.mock.Ctx.Configuration.DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(signU2FRequestBody{
		SignResponse: u2f.SignResponse{},
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorU2FSignPost(u2fVerifier)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: testRedirectionURL,
	})
}

func (s *HandlerSignU2FStep2Suite) TestShouldNotReturnRedirectURL() {
	u2fVerifier := NewMockU2FVerifier(s.mock.Ctrl)

	u2fVerifier.EXPECT().
		Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	bodyBytes, err := json.Marshal(signU2FRequestBody{
		SignResponse: u2f.SignResponse{},
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorU2FSignPost(u2fVerifier)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)
}

func (s *HandlerSignU2FStep2Suite) TestShouldRedirectUserToSafeTargetURL() {
	u2fVerifier := NewMockU2FVerifier(s.mock.Ctrl)

	u2fVerifier.EXPECT().
		Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	bodyBytes, err := json.Marshal(signU2FRequestBody{
		SignResponse: u2f.SignResponse{},
		TargetURL:    "https://mydomain.local",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorU2FSignPost(u2fVerifier)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: "https://mydomain.local",
	})
}

func (s *HandlerSignU2FStep2Suite) TestShouldNotRedirectToUnsafeURL() {
	u2fVerifier := NewMockU2FVerifier(s.mock.Ctrl)

	u2fVerifier.EXPECT().
		Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	bodyBytes, err := json.Marshal(signU2FRequestBody{
		SignResponse: u2f.SignResponse{},
		TargetURL:    "http://mydomain.local",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorU2FSignPost(u2fVerifier)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)
}

func (s *HandlerSignU2FStep2Suite) TestShouldRegenerateSessionForPreventingSessionFixation() {
	u2fVerifier := NewMockU2FVerifier(s.mock.Ctrl)

	u2fVerifier.EXPECT().
		Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	bodyBytes, err := json.Marshal(signU2FRequestBody{
		SignResponse: u2f.SignResponse{},
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	SecondFactorU2FSignPost(u2fVerifier)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)

	s.Assert().NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))
}

func TestRunHandlerSignU2FStep2Suite(t *testing.T) {
	suite.Run(t, new(HandlerSignU2FStep2Suite))
}
