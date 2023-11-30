package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/storage"
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
	userSession.AuthenticationLevel = authentication.OneFactor

	s.mock.Clock.Set(time.Unix(1701295903, 0))
	s.mock.Ctx.Clock = &s.mock.Clock
	s.mock.Ctx.Configuration.TOTP = schema.DefaultTOTPConfiguration

	s.Assert().NoError(s.mock.Ctx.SaveSession(userSession))
}

func (s *HandlerSignTOTPSuite) TearDownTest() {
	s.mock.Close()
}

func (s *HandlerSignTOTPSuite) AssertLastLogMessage(message string, err string) {
	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), message, err)
}

func (s *HandlerSignTOTPSuite) TestShouldRedirectUserToDefaultURL() {
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.
			EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(true, getStepTOTP(s.mock.Ctx, -1), nil),
		s.mock.StorageMock.
			EXPECT().
			ExistsTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890), config.HistorySince(s.mock.Clock.Now(), s.mock.Ctx.Configuration.TOTP.Skew)).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			SaveTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890)).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any()).
			Return(nil),
	)

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
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.
			EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(true, getStepTOTP(s.mock.Ctx, -1), nil),
		s.mock.StorageMock.
			EXPECT().
			ExistsTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890), config.HistorySince(s.mock.Clock.Now(), s.mock.Ctx.Configuration.TOTP.Skew)).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			SaveTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890)).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})),
		s.mock.StorageMock.
			EXPECT().
			UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any()).
			Return(errors.New("failed to perform update")),
	)

	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = testRedirectionURL

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")
}

func (s *HandlerSignTOTPSuite) TestShouldNotReturnRedirectURL() {
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.
			EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(true, getStepTOTP(s.mock.Ctx, -1), nil),
		s.mock.StorageMock.
			EXPECT().
			ExistsTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890), config.HistorySince(s.mock.Clock.Now(), s.mock.Ctx.Configuration.TOTP.Skew)).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			SaveTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890)).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any()).
			Return(nil),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), &redirectResponse{Redirect: "https://www.example.com"})
}

func (s *HandlerSignTOTPSuite) TestShouldRedirectUserToSafeTargetURL() {
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}
	s.mock.Ctx.Configuration.Session.Cookies = []schema.SessionCookie{
		{
			Domain: "example.com",
		},
		{
			Domain: "mydomain.local",
		},
	}

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.
			EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(true, getStepTOTP(s.mock.Ctx, -1), nil),
		s.mock.StorageMock.
			EXPECT().
			ExistsTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890), config.HistorySince(s.mock.Clock.Now(), s.mock.Ctx.Configuration.TOTP.Skew)).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			SaveTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890)).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any()).
			Return(nil),
	)

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
	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, testUsername).
			Return(&model.TOTPConfiguration{Secret: []byte("secret"), Period: 30}, nil),
		s.mock.TOTPMock.EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&model.TOTPConfiguration{Secret: []byte("secret"), Period: 30})).
			Return(true, getStepTOTP(s.mock.Ctx, -1), nil),
		s.mock.StorageMock.
			EXPECT().
			ExistsTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890), gomock.Any()).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			SaveTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890)).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any()).
			Return(nil),
	)

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
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.
			EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(true, getStepTOTP(s.mock.Ctx, -1), nil),
		s.mock.StorageMock.
			EXPECT().
			ExistsTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890), config.HistorySince(s.mock.Clock.Now(), s.mock.Ctx.Configuration.TOTP.Skew)).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			SaveTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890)).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateTOTPConfigurationSignIn(s.mock.Ctx, gomock.Any(), gomock.Any()).
			Return(nil),
	)

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

func (s *HandlerSignTOTPSuite) TestShouldHandleErrorSaveHistory() {
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.
			EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(true, getStepTOTP(s.mock.Ctx, -1), nil),
		s.mock.StorageMock.
			EXPECT().
			ExistsTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890), config.HistorySince(s.mock.Clock.Now(), s.mock.Ctx.Configuration.TOTP.Skew)).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			SaveTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890)).
			Return(fmt.Errorf("bad stuff")),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

	s.AssertLastLogMessage("Error occurred validating a TOTP authentication for user 'john': error occurred saving the TOTP history to the storage backend", "bad stuff")
}

func (s *HandlerSignTOTPSuite) TestShouldHandleErrorExistsHistory() {
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.
			EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(true, getStepTOTP(s.mock.Ctx, -1), nil),
		s.mock.StorageMock.
			EXPECT().
			ExistsTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890), config.HistorySince(s.mock.Clock.Now(), s.mock.Ctx.Configuration.TOTP.Skew)).
			Return(false, fmt.Errorf("oh my")),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

	s.AssertLastLogMessage("Error occurred validating a TOTP authentication for user 'john': error occurred checking the TOTP history", "oh my")
}

func (s *HandlerSignTOTPSuite) TestShouldHandleExistsHistory() {
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.
			EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(true, getStepTOTP(s.mock.Ctx, -1), nil),
		s.mock.StorageMock.
			EXPECT().
			ExistsTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890), config.HistorySince(s.mock.Clock.Now(), s.mock.Ctx.Configuration.TOTP.Skew)).
			Return(true, nil),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

	s.AssertLastLogMessage("Error occurred validating a TOTP authentication for user 'john': error occurred satisfying security policies", "the user has already used this code recently and will not be permitted to reuse it")
}

func (s *HandlerSignTOTPSuite) TestShouldHandleAnonymous() {
	us, err := s.mock.Ctx.GetSession()

	s.Require().NoError(err)

	us.Username = ""
	us.AuthenticationLevel = authentication.NotAuthenticated

	s.Require().NoError(s.mock.Ctx.SaveSession(us))

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

	s.AssertLastLogMessage("Error occurred validating a TOTP authentication", "user is anonymous")
}

func (s *HandlerSignTOTPSuite) TestShouldHandleGETAnonymous() {
	us, err := s.mock.Ctx.GetSession()

	s.Require().NoError(err)

	us.Username = ""
	us.AuthenticationLevel = authentication.NotAuthenticated

	s.Require().NoError(s.mock.Ctx.SaveSession(us))

	TimeBasedOneTimePasswordGET(s.mock.Ctx)
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

	s.AssertLastLogMessage("Error occurred retrieving TOTP configuration", "user is anonymous")
}

func (s *HandlerSignTOTPSuite) TestShouldHandleGETErrorLoadConfiguration() {
	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, testUsername).
			Return(nil, fmt.Errorf("nah")),
	)

	TimeBasedOneTimePasswordGET(s.mock.Ctx)
	s.mock.Assert500KO(s.T(), "Could not find TOTP Configuration for user.")

	s.AssertLastLogMessage("Error occurred retrieving TOTP configuration for user 'john': error occurred retrieving the configuration from the storage backend", "nah")
}

func (s *HandlerSignTOTPSuite) TestShouldHandleGETErrorLoadConfigurationNotFound() {
	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, testUsername).
			Return(nil, storage.ErrNoTOTPConfiguration),
	)

	TimeBasedOneTimePasswordGET(s.mock.Ctx)
	s.mock.Assert404KO(s.T(), "Could not find TOTP Configuration for user.")

	s.AssertLastLogMessage("Error occurred retrieving TOTP configuration for user 'john': error occurred retrieving the configuration from the storage backend", "no TOTP configuration for user")
}

func (s *HandlerSignTOTPSuite) TestShouldReturnErrorOnInvalidValue() {
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(false, uint64(0), fmt.Errorf("invalid")),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Error occurred validating a TOTP authentication for user 'john': error occurred validating the user input", "invalid")
}

func (s *HandlerSignTOTPSuite) TestShouldReturnErrorOnInvalidBoolean() {
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.
			EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(false, uint64(0), nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: false,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})).
			Return(nil),
	)

	bodyBytes, err := json.Marshal(bodySignTOTPRequest{
		Token: "abc",
	})
	s.Require().NoError(err)
	s.mock.Ctx.Request.SetBody(bodyBytes)

	r := regexp.MustCompile("^authelia_session=(.*); path=")
	res := r.FindAllStringSubmatch(string(s.mock.Ctx.Response.Header.PeekCookie("authelia_session")), -1)

	TimeBasedOneTimePasswordPOST(s.mock.Ctx)
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))

	AssertLogEntryMessageAndError(s.T(), MustGetLogLastSeq(s.T(), s.mock.Hook, 1), "Error occurred validating a TOTP authentication for user 'john': error occurred validating the user input", "the user input wasn't valid")
}

func (s *HandlerSignTOTPSuite) TestShouldReturnErrorOnInvalidBooleanMarkErr() {
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(false, uint64(0), nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   testUsername,
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
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

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
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Error occurred validating a TOTP authentication for user 'john': error parsing the request body", "unable to parse body: invalid character '1' after object key")
}

func (s *HandlerSignTOTPSuite) TestShouldReturnErrorOnInvalidBooleanMarkErrSuccess() {
	config := model.TOTPConfiguration{ID: 1, Username: testUsername, Digits: 6, Secret: []byte("secret"), Period: 30, Algorithm: "SHA1"}

	gomock.InOrder(
		s.mock.StorageMock.EXPECT().
			LoadTOTPConfiguration(s.mock.Ctx, gomock.Any()).
			Return(&config, nil),
		s.mock.TOTPMock.EXPECT().
			Validate(gomock.Eq("abc"), gomock.Eq(&config)).
			Return(true, getStepTOTP(s.mock.Ctx, -1), nil),
		s.mock.StorageMock.
			EXPECT().
			ExistsTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890), config.HistorySince(s.mock.Clock.Now(), s.mock.Ctx.Configuration.TOTP.Skew)).
			Return(false, nil),
		s.mock.StorageMock.
			EXPECT().
			SaveTOTPHistory(s.mock.Ctx, testUsername, uint64(1701295890)).
			Return(nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, model.AuthenticationAttempt{
				Username:   testUsername,
				Successful: true,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthTypeTOTP,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			}).
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
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

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
	s.mock.Assert403KO(s.T(), "Authentication failed, please retry later.")

	s.NotEqual(
		res[0][1],
		string(s.mock.Ctx.Request.Header.Cookie("authelia_session")))

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Error occurred validating a TOTP authentication for user 'john': error occurred retreiving the configuration from the storage backend", "not found")
}

func TestRunHandlerSignTOTPSuite(t *testing.T) {
	suite.Run(t, new(HandlerSignTOTPSuite))
}
