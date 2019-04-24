package handlers

import (
	"fmt"
	"testing"

	"github.com/clems4ever/authelia/mocks"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SecondFactorPreferencesSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *SecondFactorPreferencesSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	// Set the intial user session.
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = "john"
	userSession.AuthenticationLevel = 1
	s.mock.Ctx.SaveSession(userSession)
}

func (s *SecondFactorPreferencesSuite) TearDownTest() {
	s.mock.Close()
}

// GET

func (s *SecondFactorPreferencesSuite) TestShouldGetPreferenceRetrievedFromStorage() {
	s.mock.StorageProviderMock.EXPECT().
		LoadPrefered2FAMethod(gomock.Eq("john")).
		Return("u2f", nil)
	SecondFactorPreferencesGet(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), preferences{Method: "u2f"})
}

func (s *SecondFactorPreferencesSuite) TestShouldGetDefaultPreferenceIfNotInDB() {
	s.mock.StorageProviderMock.EXPECT().
		LoadPrefered2FAMethod(gomock.Eq("john")).
		Return("", nil)
	SecondFactorPreferencesGet(s.mock.Ctx)

	s.mock.Assert200OK(s.T(), preferences{Method: "totp"})
}

func (s *SecondFactorPreferencesSuite) TestShouldReturnError500WhenStorageFailsToLoad() {
	s.mock.StorageProviderMock.EXPECT().
		LoadPrefered2FAMethod(gomock.Eq("john")).
		Return("", fmt.Errorf("Failure"))
	SecondFactorPreferencesGet(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to load prefered 2FA method: Failure", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

// POST

func (s *SecondFactorPreferencesSuite) TestShouldReturnError500WhenNoBodyProvided() {
	SecondFactorPreferencesPost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SecondFactorPreferencesSuite) TestShouldReturnError500WhenMalformedBodyProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"abc\""))
	SecondFactorPreferencesPost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SecondFactorPreferencesSuite) TestShouldReturnError500WhenBadBodyProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"weird_key\":\"abc\"}"))
	SecondFactorPreferencesPost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to validate body: method: non zero value required", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SecondFactorPreferencesSuite) TestShouldReturnError500WhenBadMethodProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"abc\"}"))
	SecondFactorPreferencesPost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unknown method abc, it should be either u2f, totp or duo_push", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SecondFactorPreferencesSuite) TestShouldReturnError500WhenDatabaseFailsToSave() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"u2f\"}"))
	s.mock.StorageProviderMock.EXPECT().
		SavePrefered2FAMethod(gomock.Eq("john"), gomock.Eq("u2f")).
		Return(fmt.Errorf("Failure"))

	SecondFactorPreferencesPost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to save new prefered 2FA method: Failure", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SecondFactorPreferencesSuite) TestShouldReturn200WhenMethodIsSuccessfullySaved() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"u2f\"}"))
	s.mock.StorageProviderMock.EXPECT().
		SavePrefered2FAMethod(gomock.Eq("john"), gomock.Eq("u2f")).
		Return(nil)

	SecondFactorPreferencesPost(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
}

func TestRunPreferencesSuite(t *testing.T) {
	s := new(SecondFactorPreferencesSuite)
	suite.Run(t, s)
}
