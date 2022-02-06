package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/models"
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
	err := s.mock.Ctx.SaveSession(userSession)
	require.NoError(s.T(), err)
}

func (s *FetchSuite) TearDownTest() {
	s.mock.Close()
}

type expectedResponse struct {
	db  models.UserInfo
	api *models.UserInfo
	err error
}

func TestMethodSetToU2F(t *testing.T) {
	expectedResponses := []expectedResponse{
		{
			db: models.UserInfo{
				Method: "totp",
			},
			err: nil,
		},
		{
			db: models.UserInfo{
				Method:      "webauthn",
				HasWebauthn: true,
				HasTOTP:     true,
			},
			err: nil,
		},
		{
			db: models.UserInfo{
				Method:      "webauthn",
				HasWebauthn: true,
				HasTOTP:     false,
			},
			err: nil,
		},
		{
			db: models.UserInfo{
				Method:      "mobile_push",
				HasWebauthn: false,
				HasTOTP:     false,
			},
			err: nil,
		},
		{
			db:  models.UserInfo{},
			err: sql.ErrNoRows,
		},
		{
			db:  models.UserInfo{},
			err: errors.New("invalid thing"),
		},
	}

	for _, resp := range expectedResponses {
		if resp.api == nil {
			resp.api = &resp.db
		}

		mock := mocks.NewMockAutheliaCtx(t)
		// Set the initial user session.
		userSession := mock.Ctx.GetSession()
		userSession.Username = testUsername
		userSession.AuthenticationLevel = 1
		err := mock.Ctx.SaveSession(userSession)
		require.NoError(t, err)

		mock.StorageMock.
			EXPECT().
			LoadUserInfo(mock.Ctx, gomock.Eq("john")).
			Return(resp.db, resp.err)

		UserInfoGet(mock.Ctx)

		if resp.err == nil {
			t.Run("expected status code", func(t *testing.T) {
				assert.Equal(t, 200, mock.Ctx.Response.StatusCode())
			})

			actualPreferences := models.UserInfo{}

			mock.GetResponseData(t, &actualPreferences)

			t.Run("expected method", func(t *testing.T) {
				assert.Equal(t, resp.api.Method, actualPreferences.Method)
			})

			t.Run("registered webauthn", func(t *testing.T) {
				assert.Equal(t, resp.api.HasWebauthn, actualPreferences.HasWebauthn)
			})

			t.Run("registered totp", func(t *testing.T) {
				assert.Equal(t, resp.api.HasTOTP, actualPreferences.HasTOTP)
			})
		} else {
			t.Run("expected status code", func(t *testing.T) {
				assert.Equal(t, 200, mock.Ctx.Response.StatusCode())
			})

			errResponse := mock.GetResponseError(t)

			assert.Equal(t, "KO", errResponse.Status)
			assert.Equal(t, "Operation failed.", errResponse.Message)
		}

		mock.Close()
	}
}

func (s *FetchSuite) TestShouldReturnError500WhenStorageFailsToLoad() {
	s.mock.StorageMock.EXPECT().
		LoadUserInfo(s.mock.Ctx, gomock.Eq("john")).
		Return(models.UserInfo{}, fmt.Errorf("failure"))

	UserInfoGet(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unable to load user information: failure", s.mock.Hook.LastEntry().Message)
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
	err := s.mock.Ctx.SaveSession(userSession)
	require.NoError(s.T(), err)
}

func (s *SaveSuite) TearDownTest() {
	s.mock.Close()
}

func (s *SaveSuite) TestShouldReturnError500WhenNoBodyProvided() {
	s.mock.Ctx.Request.SetBody(nil)
	MethodPreferencePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenMalformedBodyProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"abc\""))
	MethodPreferencePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenBadBodyProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"weird_key\":\"abc\"}"))
	MethodPreferencePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unable to validate body: method: non zero value required", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenBadMethodProvided() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"abc\"}"))
	MethodPreferencePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unknown method 'abc', it should be one of totp, webauthn, mobile_push", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturnError500WhenDatabaseFailsToSave() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"webauthn\"}"))
	s.mock.StorageMock.EXPECT().
		SavePreferred2FAMethod(s.mock.Ctx, gomock.Eq("john"), gomock.Eq("webauthn")).
		Return(fmt.Errorf("Failure"))

	MethodPreferencePost(s.mock.Ctx)

	s.mock.Assert200KO(s.T(), "Operation failed.")
	assert.Equal(s.T(), "unable to save new preferred 2FA method: Failure", s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *SaveSuite) TestShouldReturn200WhenMethodIsSuccessfullySaved() {
	s.mock.Ctx.Request.SetBody([]byte("{\"method\":\"webauthn\"}"))
	s.mock.StorageMock.EXPECT().
		SavePreferred2FAMethod(s.mock.Ctx, gomock.Eq("john"), gomock.Eq("webauthn")).
		Return(nil)

	MethodPreferencePost(s.mock.Ctx)

	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
}

func TestSaveSuite(t *testing.T) {
	suite.Run(t, &SaveSuite{})
}
