package handlers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

func TestChangePasswordPOST_ShouldSucceedWithValidCredentials(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Logger.Logger.SetLevel(logrus.DebugLevel)

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

	mock.AssertLogEntryAdvanced(t, 1, logrus.DebugLevel, "User has changed their password", map[string]any{"username": testUsername})

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func TestChangePasswordPOST_ShouldFailWhenPasswordPolicyNotMet(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Logger.Logger.SetLevel(logrus.DebugLevel)

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

	mock.AssertLogEntryAdvanced(t, 0, logrus.DebugLevel, "Unable to change password for user as their new password was weak or empty", map[string]any{"username": testUsername, "error": "the supplied password does not met the security policy"})

	errResponse := mock.GetResponseError(t)

	assert.Equal(t, "KO", errResponse.Status)
	assert.Equal(t, "Your supplied password does not meet the password policy requirements.", errResponse.Message)
}

func TestChangePasswordPOST_ShouldFailWhenRequestBodyIsInvalid(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Logger.Logger.SetLevel(logrus.DebugLevel)

	userSession, err := mock.Ctx.GetSession()
	assert.NoError(t, err)

	userSession.Username = testUsername

	assert.NoError(t, mock.Ctx.SaveSession(userSession))

	mock.Ctx.Request.SetBody([]byte(`{invalid json`))

	ChangePasswordPOST(mock.Ctx)

	mock.AssertLogEntryAdvanced(t, 0, logrus.ErrorLevel, "Unable to change password for user: unable to parse request body", map[string]any{"username": testUsername, "error": regexp.MustCompile(`^(unable to parse body: .+|unable to validate body: .+|Body is not valid)$`)})

	errResponse := mock.GetResponseError(t)
	assert.Equal(t, "KO", errResponse.Status)
	assert.Equal(t, messageUnableToChangePassword, errResponse.Message)
}

func TestChangePasswordPOST_ShouldFailWhenOldPasswordIsIncorrect(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Logger.Logger.SetLevel(logrus.DebugLevel)

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

	mock.AssertLogEntryAdvanced(t, 0, logrus.DebugLevel, "Unable to change password for user as their old password was incorrect", map[string]any{"username": testUsername, "error": "incorrect password"})

	errorField := mock.Hook.LastEntry().Data["error"]
	assert.ErrorIs(t, authentication.ErrIncorrectPassword, errorField.(error))

	errResponse := mock.GetResponseError(t)
	assert.Equal(t, "KO", errResponse.Status)
	assert.Equal(t, messageIncorrectPassword, errResponse.Message)
}

func TestChangePasswordPOST_ShouldFailWhenPasswordReuseIsNotAllowed(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Logger.Logger.SetLevel(logrus.DebugLevel)

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

	mock.AssertLogEntryAdvanced(t, 0, logrus.DebugLevel, "Unable to change password for user as their new password was weak or empty", map[string]any{"username": testUsername, "error": "your supplied password does not meet the password policy requirements"})

	errorField := mock.Hook.LastEntry().Data["error"]
	assert.ErrorIs(t, authentication.ErrPasswordWeak, errorField.(error))

	errResponse := mock.GetResponseError(t)
	assert.Equal(t, "KO", errResponse.Status)
	assert.Equal(t, messagePasswordWeak, errResponse.Message)
}

func TestChangePasswordPOST_ShouldSucceedButLogErrorWhenUserHasNoEmail(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Logger.Logger.SetLevel(logrus.DebugLevel)

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

	mock.AssertLogEntryAdvanced(t, 1, logrus.DebugLevel, "User has changed their password", map[string]any{"username": testUsername})

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func TestChangePasswordPOST_ShouldSucceedButLogErrorWhenNotificationFails(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Logger.Logger.SetLevel(logrus.DebugLevel)

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
		Return(fmt.Errorf("notifier: smtp: failed to send message: connection refused"))

	ChangePasswordPOST(mock.Ctx)

	mock.AssertLogEntryAdvanced(t, 0, logrus.DebugLevel, "Unable to notify user of password change", map[string]any{"username": testUsername, "email": nil, "error": regexp.MustCompile(`^notifier: smtp: failed to .*: .+$`)})

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}
