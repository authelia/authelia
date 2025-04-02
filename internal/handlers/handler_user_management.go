package handlers

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
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
	Disabled    bool     `json:"disabled"`
	Password    string   `json:"password"`
}

type newUserRequestBody struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Password    string   `json:"password"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
	Disabled    *bool    `json:"disabled"`
}
type deleteUserRequestBody struct {
	Username string `json:"username"`
}

type AdminConfigRequestBody struct {
	Enabled                bool   `json:"enabled"`
	AdminGroup             string `json:"admin_group"`
	AllowAdminsToAddAdmins bool   `json:"allow_admins_to_add_admins"`
}

// ChangeUserPUT takes a changeUserRequestBody object and saves any changes.
//
//nolint:gocyclo
func ChangeUserPUT(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		requestBody changeUserRequestBody
		userDetails *authentication.UserDetails
		adminUser   session.UserSession
	)

	if adminUser, err = ctx.GetSession(); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred adding new user: %s", errStrUserSessionData)

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if adminUser.IsAnonymous() {
		ctx.Logger.WithError(errUserAnonymous).Error("Error occurred adding new user")

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if !UserIsAdmin(ctx, adminUser.Groups) {
		ctx.Logger.Errorf("Error occurred adding new user: %s", fmt.Sprintf(logFmtErrUserNotAdmin, adminUser.Username))

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if err = ctx.ParseBody(&requestBody); err != nil {
		ctx.Logger.Error(err, messageUnableToModifyUser)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	if len(requestBody.Username) == 0 {
		ctx.Logger.Debugf("username is required, user not changed")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("Username is required.")

		return
	}

	if userDetails, err = ctx.Providers.UserProvider.GetUser(requestBody.Username); err != nil {
		ctx.Logger.WithError(err).Errorf("Error retrieving details for user '%s'", requestBody.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)

		return
	}

	if userDetails.DisplayName != requestBody.DisplayName && !ValidatePrintableUnicodeString(requestBody.DisplayName) {
		ctx.Logger.WithFields(log.Fields{
			"user":         requestBody.Username,
			"display_name": requestBody.DisplayName,
		}).Debugf("%v: Invalid display name format", messageUnableToModifyUser)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(fmt.Sprintf(
			"Display name '%s' is invalid. Must be 1-100 characters and contain only letters, numbers, symbols, spaces and punctuation. No control characters or invisible unicode allowed. User not modified.",
			requestBody.DisplayName,
		))

		return
	}

	if userDetails.Emails[0] != requestBody.Email && !ValidateEmailString(requestBody.Email) {
		ctx.Logger.WithFields(log.Fields{
			"user":  requestBody.Username,
			"email": requestBody.Email,
		}).Debugf("%v: Email is invalid", messageUnableToModifyUser)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(fmt.Sprintf(
			"Email '%s' is invalid. Must be a valid email. User not modified.",
			requestBody.Email,
		))

		return
	}

	if !reflect.DeepEqual(SortedCopy(userDetails.Groups), SortedCopy(requestBody.Groups)) {
		if valid, badGroup := ValidateGroups(requestBody.Groups); !valid {
			ctx.Logger.WithFields(log.Fields{
				"user":          requestBody.Username,
				"invalid_group": badGroup,
			}).Debugf("%v: Invalid group name format rejected during user modification", messageUnableToModifyUser)

			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetJSONError(fmt.Sprintf(
				"Group '%s' is invalid. Must be 1-100 characters and contain only letters, numbers, and punctuation. User not modified.",
				badGroup,
			))

			return
		}
	}

	if requestBody.Password != "" && ctx.Providers.PasswordPolicy.Check(requestBody.Password) != nil {
		ctx.Logger.WithFields(log.Fields{
			"user": requestBody.Username,
		}).Debugf("%v: Password does not meet the password policy", messageUnableToModifyUser)

		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("New password does not meet the password policy. User not modified.")

		return
	}

	//TODO: Properly handle disabled users and enabling/disabling users.
	if err = ctx.Providers.UserProvider.UpdateUser(requestBody.Username, func(o *authentication.ModifyUserDetailsOpts) {
		o.Password = &requestBody.Password
		o.DisplayName = &requestBody.DisplayName
		o.SetEmail(&requestBody.Email)
		o.SetGroups(requestBody.Groups)
		b := false
		o.SetDisabled(&b)
	}); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred updating user '%s'", requestBody.Username)
	}

	if changes := GenerateUserChangeLog(userDetails, &requestBody); len(changes) > 0 {
		ctx.Logger.Debugf("User '%s' modified by administrator '%s'. Changes: %s",
			requestBody.Username, adminUser.Username, strings.Join(changes, ", "))
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

//nolint:gocyclo
func NewUserPOST(ctx *middlewares.AutheliaCtx) {
	var (
		err         error
		userSession session.UserSession
		newUser     newUserRequestBody
		options     []func(*authentication.NewUserOptionalDetailsOpts)
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
		ctx.SetJSONError(messageOperationFailed)

		return
	}

	if len(newUser.Username) == 0 {
		ctx.Logger.Debugf("User not created, username is required")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("user not created: 'username' is required.")

		return
	}

	if len(newUser.DisplayName) == 0 {
		ctx.Logger.Debugf("user not created: display_name is required")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("user not created, 'display_name' is required")

		return
	}

	if len(newUser.Password) == 0 {
		ctx.Logger.Debugf("user not created, username is required")
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError("user not created: 'password' is required.")

		return
	}

	if !ValidateUsername(newUser.Username) {
		ctx.Logger.WithError(err).Errorf("Username '%s' is formatted incorrectly.", newUser.Username)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetJSONError(messageFmtUsernameWrongFormat)

		return
	}

	if !ValidatePrintableUnicodeString(newUser.DisplayName) {
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
		var errorGroups []string

		for _, group := range newUser.Groups {
			if !ValidateGroup(group) {
				errorGroups = append(errorGroups, group)
			}
		}

		if len(errorGroups) > 0 {
			ctx.Logger.Errorf("user not created: group(s) [%s] are formatted incorrectly", strings.Join(errorGroups, ","))
			ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetJSONError(messageFmtGroupsWrongFormat)

			return
		}

		options = append(options, authentication.WithGroups(newUser.Groups))
	}

	if newUser.Email != "" {
		if !ValidateEmailString(newUser.Email) {
			ctx.Logger.WithError(err).Errorf("Email '%s' is not a valid email", newUser.Email)
		}

		options = append(options, authentication.WithEmail(newUser.Email))
	}

	if err = ctx.Providers.UserProvider.AddUser(newUser.Username, newUser.DisplayName, newUser.Password, options...); err != nil {
		ctx.Logger.WithError(err).Errorf("Error occurred creating user '%s'", newUser.Username)
	}

	if err = ctx.Providers.StorageProvider.CreateNewUserMetadata(ctx, newUser.Username); err != nil {
		ctx.Logger.Error(err)
	}

	//TODO: Add user email to notify new user of their new account. Configurable.
	ctx.Logger.Debugf("User '%s' was added.", newUser.Username)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func DeleteUserDELETE(ctx *middlewares.AutheliaCtx) {
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

	// Delete Opaque User Identifiers, User Preferences, 2FA Devices, Oauth Sessions related to Opaque Ids, Remove User from Backend.

	if err = ctx.Providers.UserProvider.DeleteUser(requestBody.Username); err != nil {
		ctx.Logger.Error(err, messageUnableToDeleteUser)
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
