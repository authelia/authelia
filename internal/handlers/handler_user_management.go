package handlers

import (
	"fmt"
	"reflect"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
)

type changeUserRequestBody struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
}

type newUserRequestBody struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Password    string   `json:"password"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
}
type deleteUserRequestBody struct {
	Username string `json:"username"`
}

type AdminConfigRequestBody struct {
	Enabled                bool   `json:"enabled"`
	AdminGroup             string `json:"admin_group"`
	AllowAdminsToAddAdmins bool   `json:"allow_admins_to_add_admins"`
}

// ChangeUserPOST takes a changeUserRequestBody object and saves any changes.
func ChangeUserPOST(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		requestBody changeUserRequestBody
		userDetails *authentication.UserDetails
		adminUser   session.UserSession
	)

	if adminUser, err = ctx.GetSession(); err != nil {
		ctx.Logger.Error("error retrieving admin session")
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.Logger.Error(err, messageUnableToModifyUser)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	if len(requestBody.Username) == 0 {
		ctx.Logger.Debugf("username is blank, user not changed")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)

		return
	}

	if userDetails, err = ctx.Providers.UserProvider.GetDetails(requestBody.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error retrieving details for user '%s'", requestBody.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	}

	if userDetails.DisplayName != requestBody.DisplayName {
		if err = ctx.Providers.UserProvider.ChangeDisplayName(requestBody.Username, requestBody.DisplayName); err != nil {
			ctx.Logger.WithError(err).Errorf("Error changing display name to '%s' for user '%s'", requestBody.DisplayName, requestBody.Username)
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

			return
		}

		ctx.Logger.Debugf("User '%s' display name changed to '%s' by administrator: '%s'", requestBody.Username, requestBody.DisplayName, adminUser.Username)
	}

	if userDetails.Emails[0] != requestBody.Email {
		if err = ctx.Providers.UserProvider.ChangeEmail(requestBody.Username, requestBody.Email); err != nil {
			ctx.Logger.WithError(err).Errorf("Error changing email for user '%s'", requestBody.Username)
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

			return
		}

		ctx.Logger.Debugf("User '%s' email changed to '%s' by administrator: '%s'", requestBody.Username, requestBody.Email, adminUser.Username)
	}

	if !reflect.DeepEqual(userDetails.Groups, requestBody.Groups) {
		if err = ctx.Providers.UserProvider.ChangeGroups(requestBody.Username, requestBody.Groups); err != nil {
			ctx.Logger.WithError(err).Errorf("Error changing groups for user '%s'", requestBody.Username)
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

			return
		}

		ctx.Logger.Debugf("User '%s' groups changed to '%s' by administrator: '%s'", requestBody.Username, requestBody.Groups, adminUser.Username)
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func AdminConfigGET(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		userSession session.UserSession
		adminConfig AdminConfigRequestBody
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred retrieving admin config: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred retrieving admin config")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	adminConfig = AdminConfigRequestBody{
		Enabled:                ctx.Configuration.Administration.Enabled,
		AdminGroup:             ctx.Configuration.Administration.AdminGroup,
		AllowAdminsToAddAdmins: ctx.Configuration.Administration.AllowAdminsToAddAdmins,
	}

	err = ctx.SetJSONBody(adminConfig)
	if err != nil {
		ctx.Logger.Errorf("Unable to set admin config response in body: %+v", err)
	}
}

//nolint:gocyclo
func NewUserPUT(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		userSession session.UserSession
		newUser     newUserRequestBody
		options     []func(*authentication.NewUserDetailsOpts)
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred adding new user: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred adding new user")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !UserIsAdmin(ctx, userSession.Groups) {
		ctx.Logger.Errorf("Error occurred adding new user: %s", fmt.Sprintf(logFmtErrUserNotAdmin, userSession.Username))

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.ParseBody(&newUser); err != nil {
		ctx.Logger.Error(err, messageUnableToAddUser)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	if newUser.Username == "" || newUser.DisplayName == "" || newUser.Password == "" {
		ctx.Logger.Errorf("Username, DisplayName, and Password are required fields.")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageNewUserRequiredFields)

		return
	}

	if err = ValidateUsername(newUser.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Username '%s' is formatted incorrectly.", newUser.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageFmtUsernameWrongFormat)

		return
	}

	if err = ValidatePrintableUnicodeString(newUser.DisplayName); err != nil {
		ctx.Logger.WithError(err).Errorf("Display Name '%s' is formatted incorrectly.", newUser.DisplayName)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageFmtDisplayNameWrongFormat)

		return
	}

	if err = ctx.Providers.PasswordPolicy.Check(newUser.Password); err != nil {
		ctx.Error(err, messagePasswordWeak)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messagePasswordWeak)

		return
	}

	if len(newUser.Groups) > 0 {
		for _, group := range newUser.Groups {
			if err = ValidateGroup(group); err != nil {
				ctx.Logger.WithError(err).Errorf("Group '%s'is formatted incorrectly.", group)
				ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
				ctx.SetJSONError(messageFmtGroupsWrongFormat)

				return
			}
		}

		options = append(options, authentication.WithGroups(newUser.Groups))
	}

	if newUser.Email != "" {
		if ValidateEmailString(newUser.Email) != nil {
			ctx.Logger.WithError(err).Errorf("Email '%s' is not a valid email", newUser.Email)
		}

		options = append(options, authentication.WithEmail(newUser.Email))
	}

	if err = ctx.Providers.UserProvider.AddUser(newUser.Username, newUser.DisplayName, newUser.Password, options...); err != nil {
		ctx.Logger.Error(err, messageUnableToAddUser)
	}
	//TODO: Add user email to notify new user of their new account. Configurable.
	ctx.Logger.Debugf("User '%s' was added.", newUser.Username)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func DeleteUserDELETE(ctx *middlewares.AutheliaCtx) {
	// Delete Opaque User Identifiers, User Preferences, 2FA Devices, Oauth Sessions related to Opaque Ids, Remove User from Backend.
	var (
		err         error
		userSession session.UserSession
		requestBody deleteUserRequestBody
	)

	if userSession, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred deleting specified user: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if userSession.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred deleting specified user")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.Logger.WithError(err).Error(messageUnableToDeleteUser)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	// TODO: Validation for new user's fields: Password Policy, Email Regex, groups etc.

	if err = ctx.Providers.UserProvider.DeleteUser(requestBody.Username); err != nil {
		ctx.Logger.Error(err, messageUnableToDeleteUser)
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}
