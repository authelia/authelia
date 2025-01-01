package handlers

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
)

const (
	testPasswordOld = "old_password123"
	testPasswordNew = "new_password456"
)

type ChangePasswordSuite struct {
	suite.Suite
	mock *mocks.MockAutheliaCtx
}

func (s *ChangePasswordSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	userSession.Username = testUsername
	userSession.DisplayName = testUsername
	userSession.Emails[0] = testEmail
	userSession.AuthenticationLevel = 1
	s.Assert().NoError(s.mock.Ctx.SaveSession(userSession))
}

func (s *ChangePasswordSuite) TearDownTest() {
	s.mock.Close()
}

func TestChangePasswordPOST_ShouldSucceedWithValidCredentials(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	userSession, err := mock.Ctx.GetSession()
	assert.NoError(t, err)

	userSession.Username = testUsername

	assert.NoError(t, mock.Ctx.SaveSession(userSession))

	oldPassword := testPasswordOld
	newPassword := testPasswordNew

	requestBody := changePasswordRequestBody{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	bodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)
	mock.Ctx.Request.SetBody(bodyBytes)

	mock.Ctx.Providers.PasswordPolicy = middlewares.NewPasswordPolicyProvider(schema.PasswordPolicy{})

	mock.NotifierMock.EXPECT().
		Send(mock.Ctx, gomock.Any(), "Password changed successfully", gomock.Any(), gomock.Any()).
		Return(nil)

	mock.UserProviderMock.EXPECT().
		ChangePassword(userSession.Username, oldPassword, newPassword).
		Return(nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(testUsername).
		Return(&authentication.UserDetails{
			Emails: []string{testEmail},
		}, nil)

	ChangePasswordPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func TestChangePasswordPOST_ShouldFailWhenPasswordPolicyNotMet(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	userSession, err := mock.Ctx.GetSession()
	assert.NoError(t, err)

	userSession.Username = testUsername

	assert.NoError(t, mock.Ctx.SaveSession(userSession))

	oldPassword := testPasswordOld
	newPassword := "weak"

	requestBody := changePasswordRequestBody{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	bodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)
	mock.Ctx.Request.SetBody(bodyBytes)

	passwordPolicy := middlewares.NewPasswordPolicyProvider(schema.PasswordPolicy{
		Standard: schema.PasswordPolicyStandard{
			Enabled:          true,
			MinLength:        8,
			MaxLength:        64,
			RequireNumber:    true,
			RequireSpecial:   true,
			RequireUppercase: true,
			RequireLowercase: true,
		},
	})

	mock.Ctx.Providers.PasswordPolicy = passwordPolicy

	ChangePasswordPOST(mock.Ctx)

	errResponse := mock.GetResponseError(t)

	assert.Equal(t, "KO", errResponse.Status)
	assert.Equal(t, "Your supplied password does not meet the password policy requirements.", errResponse.Message)
}

func TestChangePasswordPOST_ShouldFailWhenRequestBodyIsInvalid(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	userSession, err := mock.Ctx.GetSession()
	assert.NoError(t, err)

	userSession.Username = testUsername

	assert.NoError(t, mock.Ctx.SaveSession(userSession))

	mock.Ctx.Request.SetBody([]byte(`{invalid json`))

	ChangePasswordPOST(mock.Ctx)

	errResponse := mock.GetResponseError(t)
	assert.Equal(t, "KO", errResponse.Status)
	assert.Equal(t, messageUnableToChangePassword, errResponse.Message)
}

func TestChangePasswordPOST_ShouldFailWhenOldPasswordIsIncorrect(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	userSession, err := mock.Ctx.GetSession()
	assert.NoError(t, err)

	userSession.Username = testUsername

	assert.NoError(t, mock.Ctx.SaveSession(userSession))

	oldPassword := testPasswordOld
	newPassword := testPasswordNew

	requestBody := changePasswordRequestBody{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	bodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)
	mock.Ctx.Request.SetBody(bodyBytes)

	mock.Ctx.Providers.PasswordPolicy = middlewares.NewPasswordPolicyProvider(schema.PasswordPolicy{})

	mock.UserProviderMock.EXPECT().
		ChangePassword(testUsername, oldPassword, newPassword).
		Return(authentication.ErrIncorrectPassword)

	ChangePasswordPOST(mock.Ctx)

	errResponse := mock.GetResponseError(t)
	assert.Equal(t, "KO", errResponse.Status)
	assert.Equal(t, messageIncorrectPassword, errResponse.Message)
}

func TestChangePasswordPOST_ShouldFailWhenPasswordReuseIsNotAllowed(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	userSession, err := mock.Ctx.GetSession()
	assert.NoError(t, err)

	userSession.Username = testUsername

	assert.NoError(t, mock.Ctx.SaveSession(userSession))

	oldPassword := testPasswordOld
	newPassword := testPasswordOld

	requestBody := changePasswordRequestBody{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	bodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)
	mock.Ctx.Request.SetBody(bodyBytes)

	mock.Ctx.Providers.PasswordPolicy = middlewares.NewPasswordPolicyProvider(schema.PasswordPolicy{})

	mock.UserProviderMock.EXPECT().
		ChangePassword(testUsername, oldPassword, newPassword).
		Return(authentication.ErrPasswordWeak)

	ChangePasswordPOST(mock.Ctx)

	errResponse := mock.GetResponseError(t)
	assert.Equal(t, "KO", errResponse.Status)
	assert.Equal(t, messageCannotReusePassword, errResponse.Message)
}

func TestChangePasswordPOST_ShouldSucceedButLogErrorWhenUserHasNoEmail(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	userSession, err := mock.Ctx.GetSession()
	assert.NoError(t, err)

	userSession.Username = testUsername

	assert.NoError(t, mock.Ctx.SaveSession(userSession))

	oldPassword := testPasswordOld
	newPassword := testPasswordNew

	requestBody := changePasswordRequestBody{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	bodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)
	mock.Ctx.Request.SetBody(bodyBytes)

	mock.Ctx.Providers.PasswordPolicy = middlewares.NewPasswordPolicyProvider(schema.PasswordPolicy{})

	mock.UserProviderMock.EXPECT().
		ChangePassword(testUsername, oldPassword, newPassword).
		Return(nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(testUsername).
		Return(&authentication.UserDetails{
			Emails: []string{},
		}, nil)

	ChangePasswordPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func TestChangePasswordPOST_ShouldSucceedButLogErrorWhenNotificationFails(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	userSession, err := mock.Ctx.GetSession()
	assert.NoError(t, err)

	userSession.Username = testUsername

	assert.NoError(t, mock.Ctx.SaveSession(userSession))

	oldPassword := testPasswordOld
	newPassword := testPasswordNew

	requestBody := changePasswordRequestBody{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	bodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)
	mock.Ctx.Request.SetBody(bodyBytes)

	mock.Ctx.Providers.PasswordPolicy = middlewares.NewPasswordPolicyProvider(schema.PasswordPolicy{})

	mock.UserProviderMock.EXPECT().
		ChangePassword(testUsername, testPasswordOld, newPassword).
		Return(nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(testUsername).
		Return(&authentication.UserDetails{
			Emails: []string{testEmail},
		}, nil)

	mock.NotifierMock.EXPECT().
		Send(mock.Ctx, gomock.Any(), "Password changed successfully", gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("failed to send notification"))

	ChangePasswordPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}
