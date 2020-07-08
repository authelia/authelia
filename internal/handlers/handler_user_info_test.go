package handlers

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/mocks"
	"github.com/authelia/authelia/internal/storage"
)

type FetchSuite struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *FetchSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	// Set the initial user session.
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = 1
	s.mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
}

func (s *FetchSuite) TearDownTest() {
	s.mock.Close()
}

func setPreferencesExpectations(preferences UserInfo, provider *storage.MockProvider) {
	provider.
		EXPECT().
		LoadPreferred2FAMethod(gomock.Eq("john")).
		Return(preferences.Method, nil)

	if preferences.HasU2F {
		u2fData := []byte("abc")
		provider.
			EXPECT().
			LoadU2FDeviceHandle(gomock.Eq("john")).
			Return(u2fData, u2fData, nil)
	} else {
		provider.
			EXPECT().
			LoadU2FDeviceHandle(gomock.Eq("john")).
			Return(nil, nil, storage.ErrNoU2FDeviceHandle)
	}

	if preferences.HasTOTP {
		totpSecret := "secret"
		provider.
			EXPECT().
			LoadTOTPSecret(gomock.Eq("john")).
			Return(totpSecret, nil)
	} else {
		provider.
			EXPECT().
			LoadTOTPSecret(gomock.Eq("john")).
			Return("", storage.ErrNoTOTPSecret)
	}
}

func TestMethodSetToU2F(t *testing.T) {
	table := []UserInfo{
		{
			Method: "totp",
		},
		{
			Method:  "u2f",
			HasU2F:  true,
			HasTOTP: true,
		},
		{
			Method:  "u2f",
			HasU2F:  true,
			HasTOTP: false,
		},
		{
			Method:  "mobile_push",
			HasU2F:  false,
			HasTOTP: false,
		},
	}

	for _, expectedPreferences := range table {
		mock := mocks.NewMockAutheliaCtx(t)
		// Set the initial user session.
		userSession := mock.Ctx.GetSession()
		userSession.Username = testUsername
		userSession.AuthenticationLevel = 1
		mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

		setPreferencesExpectations(expectedPreferences, mock.StorageProviderMock)
		UserInfoGet(mock.Ctx)

		actualPreferences := UserInfo{}
		mock.GetResponseData(t, &actualPreferences)

		t.Run("expected method", func(t *testing.T) {
			assert.Equal(t, expectedPreferences.Method, actualPreferences.Method)
		})

		t.Run("registered u2f", func(t *testing.T) {
			assert.Equal(t, expectedPreferences.HasU2F, actualPreferences.HasU2F)
		})

		t.Run("registered totp", func(t *testing.T) {
			assert.Equal(t, expectedPreferences.HasTOTP, actualPreferences.HasTOTP)
		})
		mock.Close()
	}
}

func (s *FetchSuite) TestShouldGetDefaultPreferenceIfNotInDB() {
	s.mock.StorageProviderMock.
		EXPECT().
		LoadPreferred2FAMethod(gomock.Eq("john")).
		Return("", nil)

	s.mock.StorageProviderMock.
		EXPECT().
		LoadU2FDeviceHandle(gomock.Eq("john")).
		Return(nil, nil, storage.ErrNoU2FDeviceHandle)

	s.mock.StorageProviderMock.
		EXPECT().
		LoadTOTPSecret(gomock.Eq("john")).
		Return("", storage.ErrNoTOTPSecret)

	UserInfoGet(s.mock.Ctx)
	s.mock.Assert200OK(s.T(), UserInfo{Method: "totp"})
}

func (s *FetchSuite) TestShouldReturnError500WhenStorageFailsToLoad() {
	s.mock.StorageProviderMock.EXPECT().
		LoadPreferred2FAMethod(gomock.Eq("john")).
		Return("", fmt.Errorf("Failure"))

	s.mock.StorageProviderMock.
		EXPECT().
		LoadU2FDeviceHandle(gomock.Eq("john"))

	s.mock.StorageProviderMock.
		EXPECT().
		LoadTOTPSecret(gomock.Eq("john"))

	UserInfoGet(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to load user information", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func TestFetchSuite(t *testing.T) {
	suite.Run(t, &FetchSuite{})
}

type SaveSuite struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *SaveSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	// Set the initial user session.
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = 1
	s.mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
}

func (s *SaveSuite) TearDownTest() {
	s.mock.Close()
}

func (s *SaveSuite) TestShouldReturnError500WhenNoBodyProvided() {
	s.mock.Ctx.Request.SetBody(nil)
	MethodPreferencePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenMalformedBodyProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"abc\""))
	MethodPreferencePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenBadBodyProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"weird_key\":\"abc\"}"))
	MethodPreferencePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to validate body: method: non zero value required", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenBadMethodProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"abc\"}"))
	MethodPreferencePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unknown method 'abc', it should be one of totp, u2f, mobile_push", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenDatabaseFailsToSave() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"u2f\"}"))
	s.mock.StorageProviderMock.EXPECT().
		SavePreferred2FAMethod(gomock.Eq("john"), gomock.Eq("u2f")).
		Return(fmt.Errorf("Failure"))

	MethodPreferencePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "Unable to save new preferred 2FA method: Failure", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturn200WhenMethodIsSuccessfullySaved() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"u2f\"}"))
	s.mock.StorageProviderMock.EXPECT().
		SavePreferred2FAMethod(gomock.Eq("john"), gomock.Eq("u2f")).
		Return(nil)

	MethodPreferencePost(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
}

func TestSaveSuite(t *testing.T) {
	suite.Run(t, &SaveSuite{})
}
