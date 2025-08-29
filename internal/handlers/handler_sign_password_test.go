package handlers

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
)

type HandlerSignPasswordSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *HandlerSignPasswordSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	userSession.Username = testUsername
	userSession.AuthenticationMethodRefs.WebAuthn = true
	userSession.AuthenticationMethodRefs.WebAuthnUserPresence = true
	userSession.AuthenticationMethodRefs.WebAuthnUserVerified = true

	s.Assert().NoError(s.mock.Ctx.SaveSession(userSession))

	s.mock.Clock.Set(time.Unix(1701295903, 0))
	s.mock.Ctx.Providers.Clock = &s.mock.Clock
	s.mock.Ctx.Configuration.TOTP = schema.DefaultTOTPConfiguration
}

func (s *HandlerSignPasswordSuite) TearDownTest() {
	s.mock.Close()
}

func (s *HandlerSignPasswordSuite) AssertLastLogMessage(message string, err string) {
	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), message, err)
}

func (s *HandlerSignPasswordSuite) TestShouldRedirectUserToDefaultURL() {
	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq("john"), gomock.Eq("123456")).
			Return(true, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypePassword,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
	)

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySecondFactorPasswordRequest{
		Password: "123456",
	})

	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: testRedirectionURLString,
	})
}

func (s *HandlerSignPasswordSuite) TestShouldHandleOpenIDConnect() {
	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq("john"), gomock.Eq("123456")).
			Return(true, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypePassword,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
	)

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySecondFactorPasswordRequest{
		Password: "123456",
		Flow:     flowNameOpenIDConnect,
		FlowID:   "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
	s.mock.AssertLogEntryAdvanced(s.T(), 0, logrus.ErrorLevel, "Error occurred parsing the consent session flow id", map[string]any{"error": "invalid UUID length: 3", "flow": "openid_connect", "flow_id": "abc", "subflow": ""})
}

func (s *HandlerSignPasswordSuite) TestShouldRedirectUserToDefaultURLDelayFunc() {
	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq("john"), gomock.Eq("123456")).
			Return(true, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypePassword,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
	)

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySecondFactorPasswordRequest{
		Password: "123456",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	delayFunc := func(ctx *middlewares.AutheliaCtx, requestTime time.Time, successful *bool) {}

	SecondFactorPasswordPOST(delayFunc)(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: testRedirectionURLString,
	})
}

func (s *HandlerSignPasswordSuite) TestShouldErrorMarkAttempt() {
	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq("john"), gomock.Eq("123456")).
			Return(true, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypePassword,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(fmt.Errorf("bad socket")),
	)

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySecondFactorPasswordRequest{
		Password: "123456",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), &redirectResponse{Redirect: "https://www.example.com"})
	s.AssertLastLogMessage("Failed to record Password authentication attempt", "bad socket")
}

func (s *HandlerSignPasswordSuite) TestShouldHandleBadPassword() {
	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq("john"), gomock.Eq("123456")).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: false,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypePassword,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
	)

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySecondFactorPasswordRequest{
		Password: "123456",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
	s.AssertLastLogMessage("Unsuccessful Password authentication attempt by user 'john'", "")
}

func (s *HandlerSignPasswordSuite) TestShouldHandleBadPasswordMarkAttemptError() {
	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq("john"), gomock.Eq("123456")).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: false,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypePassword,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(fmt.Errorf("bad sockets")),
	)

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySecondFactorPasswordRequest{
		Password: "123456",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
	s.AssertLastLogMessage("Unsuccessful Password authentication attempt by user 'john'", "")
}

func (s *HandlerSignPasswordSuite) TestShouldHandleBadPasswordWithError() {
	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq("john"), gomock.Eq("123456")).
			Return(false, fmt.Errorf("bad user pass")),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: false,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypePassword,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
	)

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySecondFactorPasswordRequest{
		Password: "123456",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	SecondFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
	s.AssertLastLogMessage("Unsuccessful Password authentication attempt by user 'john'", "bad user pass")
}

func (s *HandlerSignPasswordSuite) TestShouldErrorBadRequestBody() {
	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	s.mock.Ctx.Request.SetBody([]byte(`{"password":`))

	SecondFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
	s.AssertLastLogMessage("Failed to parse 1FA request body", "unable to parse body: unexpected end of JSON input")
}

func TestRunHandlerSignPasswordSuite(t *testing.T) {
	suite.Run(t, new(HandlerSignPasswordSuite))
}
