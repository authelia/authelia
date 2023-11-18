package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
)

type HandlerSignTOTPSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *HandlerSignTOTPSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	userSession.Username = testUsername
	s.Assert().NoError(s.mock.Ctx.SaveSession(userSession))
}

func (s *HandlerSignTOTPSuite) TearDownTest() {
	s.mock.Close()
}

func (s *HandlerSignTOTPSuite) TestShouldRedirectUserToDefaultURL() {
	config := model.TOTPConfiguration{ID: 1, Username: "john", Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	s.mock.StorageMock.EXPECT().
		LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
		Return(&config, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeTOTP,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.TOTPMock.EXPECT().Validate(gomock.Eq("abc"), gomock.Eq(&config)).Return(true, nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any())

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: testRedirectionURLString,
	})
}

func (s *HandlerSignTOTPSuite) TestShouldFailWhenTOTPSignInInfoFailsToUpdate() {
	config := model.TOTPConfiguration{ID: 1, Username: "john", Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	s.mock.StorageMock.EXPECT().
		LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
		Return(&config, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeTOTP,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.TOTPMock.EXPECT().Validate(gomock.Eq("abc"), gomock.Eq(&config)).Return(true, nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any()).Return(errors.New("failed to perform update"))

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")
}

func (s *HandlerSignTOTPSuite) TestShouldNotReturnRedirectURL() {
	config := model.TOTPConfiguration{ID: 1, Username: "john", Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	s.mock.StorageMock.EXPECT().
		LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
		Return(&config, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeTOTP,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.TOTPMock.EXPECT().Validate(gomock.Eq("abc"), gomock.Eq(&config)).Return(true, nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any())

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), &redirectResponse{Redirect: "https://www.example.com"})
}

func (s *HandlerSignTOTPSuite) TestShouldRedirectUserToSafeTargetURL() {
	config := model.TOTPConfiguration{ID: 1, Username: "john", Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}
	s.mock.Ctx.Configuration.Session.Cookies = []schema.SessionCookie{
		{
			Domain: "example.com",
		},
		{
			Domain: "mydomain.local",
		},
	}

	s.mock.StorageMock.EXPECT().
		LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
		Return(&config, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeTOTP,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.TOTPMock.EXPECT().Validate(gomock.Eq("abc"), gomock.Eq(&config)).Return(true, nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any())

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token:     "abc",
		TargetURL: "https://mydomain.example.com",
	})

	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), redirectResponse{
		Redirect: "https://mydomain.example.com",
	})
}

func (s *HandlerSignTOTPSuite) TestShouldNotRedirectToUnsafeURL() {
	s.mock.StorageMock.EXPECT().
		LoadTOTPConfiguration(s.mock.Ctx, "john").
		Return(&model.TOTPConfiguration{Secret: []byte("secret")}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeTOTP,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.StorageMock.
		EXPECT().
		UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any())

	s.mock.TOTPMock.EXPECT().
		Validate(gomock.Eq("abc"), gomock.Eq(&model.TOTPConfiguration{Secret: []byte("secret")})).
		Return(true, nil)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token:     "abc",
		TargetURL: "http://mydomain.example.com",
	})

	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), nil)
}

func (s *HandlerSignTOTPSuite) TestShouldRegenerateSessionForPreventingSessionFixation() {
	config := model.TOTPConfiguration{ID: 1, Username: "john", Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	s.mock.StorageMock.EXPECT().
		LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
		Return(&config, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "john",
			Successful: true,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthTypeTOTP,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.TOTPMock.EXPECT().
		Validate(gomock.Eq("abc"), gomock.Eq(&config)).
		Return(true, nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any())

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), &redirectResponse{Redirect: "https://www.example.com"})

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))
}

func (s *HandlerSignTOTPSuite) TestShouldReturnErrorOnInvalidValue() {
	config := model.TOTPConfiguration{ID: 1, Username: "john", Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(false, fmt.Errorf("invalid")),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Failed to perform TOTP verification", "invalid")
}

func (s *HandlerSignTOTPSuite) TestShouldReturnErrorOnInvalidBoolean() {
	config := model.TOTPConfiguration{ID: 1, Username: "john", Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   "john",
				Successful: false,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Unsuccessful TOTP authentication attempt by user 'john'", "")
}

func (s *HandlerSignTOTPSuite) TestShouldReturnErrorOnInvalidBooleanMarkErr() {
	config := model.TOTPConfiguration{ID: 1, Username: "john", Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   "john",
				Successful: false,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).Return(fmt.Errorf("failed to insert")),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Unable to mark TOTP authentication attempt by user 'john'", "failed to insert")
}

func (s *HandlerSignTOTPSuite) TestShouldReturnErrorOnInvalidJSON() {
	s.mock.Ctx.Request.SetBody([]byte(`{"token:"1234"}`))

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Failed to parse TOTP request body", "unable to parse body: invalid character '1' after object key")
}

func (s *HandlerSignTOTPSuite) TestShouldReturnErrorOnInvalidBooleanMarkErrSuccess() {
	config := model.TOTPConfiguration{ID: 1, Username: "john", Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(true, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   "john",
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(fmt.Errorf("failed to insert")),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Unable to mark TOTP authentication attempt by user 'john'", "failed to insert")
}

func (s *HandlerSignTOTPSuite) TestShouldReturnErrorOnInvalidConfig() {
	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(nil, fmt.Errorf("not found")),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert401KO(s.T(), "Authentication failed, please retry later.")

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Failed to load TOTP configuration", "not found")
}

func TestRunHandlerSignTOTPSuite(t *testing.T) {
	suite.Run(t, new(HandlerSignTOTPSuite))
}
