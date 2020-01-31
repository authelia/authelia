package handlers

import (
	"encoding/json"
	"testing"

	"github.com/authelia/authelia/internal/mocks"
	"github.com/authelia/authelia/internal/session"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/tstranex/u2f"
)

type HandlerSignTOTPSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *HandlerSignTOTPSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = "john"
	userSession.U2FChallenge = &u2f.Challenge{}
	userSession.U2FRegistration = &session.U2FRegistration{}
	s.mock.Ctx.SaveSession(userSession)
}

func (s *HandlerSignTOTPSuite) TearDownTest() {
	s.mock.Close()
}

func (s *HandlerSignTOTPSuite) TestShouldRedirectUserToDefaultURL() {
	verifier := NewMockTOTPVerifier(s.mock.Ctrl)

	s.mock.StorageProviderMock.EXPECT().
		LoadTOTPSecret(gomock.Any()).
		Return("secret", nil)

	verifier.EXPECT().
		Verify(gomock.Eq("abc"), gomock.Eq("secret")).
		Return(true)

	s.mock.Ctx.Configuration.DefaultRedirectionURL = "http://redirection.local"

	bodyBytes, err := json.Marshal(signTOTPRequestBody{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorTOTPPost(verifier)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: "http://redirection.local",
	})
}

func (s *HandlerSignTOTPSuite) TestShouldNotReturnRedirectURL() {
	verifier := NewMockTOTPVerifier(s.mock.Ctrl)

	s.mock.StorageProviderMock.EXPECT().
		LoadTOTPSecret(gomock.Any()).
		Return("secret", nil)

	verifier.EXPECT().
		Verify(gomock.Eq("abc"), gomock.Eq("secret")).
		Return(true)

	bodyBytes, err := json.Marshal(signTOTPRequestBody{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorTOTPPost(verifier)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)
}

func (s *HandlerSignTOTPSuite) TestShouldRedirectUserToSafeTargetURL() {
	verifier := NewMockTOTPVerifier(s.mock.Ctrl)

	s.mock.StorageProviderMock.EXPECT().
		LoadTOTPSecret(gomock.Any()).
		Return("secret", nil)

	verifier.EXPECT().
		Verify(gomock.Eq("abc"), gomock.Eq("secret")).
		Return(true)

	bodyBytes, err := json.Marshal(signTOTPRequestBody{
		Token:     "abc",
		TargetURL: "https://mydomain.local",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorTOTPPost(verifier)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: "https://mydomain.local",
	})
}

func (s *HandlerSignTOTPSuite) TestShouldNotRedirectToUnsafeURL() {
	verifier := NewMockTOTPVerifier(s.mock.Ctrl)

	s.mock.StorageProviderMock.EXPECT().
		LoadTOTPSecret(gomock.Any()).
		Return("secret", nil)

	verifier.EXPECT().
		Verify(gomock.Eq("abc"), gomock.Eq("secret")).
		Return(true)

	bodyBytes, err := json.Marshal(signTOTPRequestBody{
		Token:     "abc",
		TargetURL: "http://mydomain.local",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorTOTPPost(verifier)(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)
}

func TestRunHandlerSignTOTPSuite(t *testing.T) {
	suite.Run(t, new(HandlerSignTOTPSuite))
}
