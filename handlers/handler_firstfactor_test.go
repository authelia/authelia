package handlers

import (
	"fmt"
	"testing"

	"github.com/clems4ever/authelia/mocks"

	"github.com/clems4ever/authelia/authentication"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FirstFactorSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *FirstFactorSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
}

func (s *FirstFactorSuite) TearDownTest() {
	s.mock.Close()
}

func (s *FirstFactorSuite) assertError500(err string) {
	assert.Equal(s.T(), 500, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte(InternalError), s.mock.Ctx.Response.Body())
	assert.Equal(s.T(), err, s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *FirstFactorSuite) TestShouldFailIfBodyIsNil() {
	FirstFactorPost(s.mock.Ctx)

	// No body
	assert.Equal(s.T(), "Unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfBodyIsInBadFormat() {
	// Missing password
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test"
	}`)
	FirstFactorPost(s.mock.Ctx)

	assert.Equal(s.T(), "Unable to validate body: password: non zero value required", s.mock.Hook.LastEntry().Message)
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfUserProviderCheckPasswordFail() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(false, fmt.Errorf("Failed"))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPost(s.mock.Ctx)

	assert.Equal(s.T(), "Error while checking password for user test: Failed", s.mock.Hook.LastEntry().Message)
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfUserProviderGetDetailsFail() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(nil, fmt.Errorf("Failed"))

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPost(s.mock.Ctx)

	assert.Equal(s.T(), "Error while retrieving details from user test: Failed", s.mock.Hook.LastEntry().Message)
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfAuthenticationLoggingFail() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(nil, nil)

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(fmt.Errorf("failed"))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPost(s.mock.Ctx)

	assert.Equal(s.T(), "Unable to mark authentication: failed", s.mock.Hook.LastEntry().Message)
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldAuthenticateUser() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{
			Emails: []string{"test@example.com"},
			Groups: []string{"dev", "admin"},
		}, nil)

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPost(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	// And store authentication in session.
	session := s.mock.Ctx.GetSession()
	assert.Equal(s.T(), "test", session.Username)
	assert.Equal(s.T(), authentication.OneFactor, session.AuthenticationLevel)
	assert.Equal(s.T(), []string{"test@example.com"}, session.Emails)
	assert.Equal(s.T(), []string{"dev", "admin"}, session.Groups)

}

func TestFirstFactorSuite(t *testing.T) {
	firstFactorSuite := new(FirstFactorSuite)
	suite.Run(t, firstFactorSuite)
}
